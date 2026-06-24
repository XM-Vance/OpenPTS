'use client';

// 文档详情：基础信息 + 提取字段核对(可编辑) + 确认入库 + 原件/解析件下载 + 全文预览。
import { useState, useMemo } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  getDocument,
  downloadDocumentFile,
  updateExtraction,
  addExtraction,
  deleteExtraction,
  reparseDocument,
  applyDocument,
  deleteDocument,
  type DocumentExtraction,
} from '@/lib/api/documents';
import { setContractPrefill } from '@/lib/contract-prefill';
import { listOrgs } from '@/lib/api/orgs';
import { ArrowLeft, Download, FileText, RefreshCw, Trash2, CheckCircle2, Plus, X, FileSignature, MapPin } from 'lucide-react';

// 从日期类文本取 YYYY-MM（容忍 2026-01-01 / 2026/1/1 / 2026年1月1日）
function monthOf(s: string): string {
  const m = s.match(/(\d{4})\D+(\d{1,2})/);
  return m ? `${m[1]}-${m[2].padStart(2, '0')}` : '';
}
// 取文本中的数值（去千分位/单位）
function numText(s: string): string {
  const m = s.replace(/[,，\s]/g, '').match(/\d+(\.\d+)?/);
  return m ? m[0] : '';
}

const SELECT_CLASS = 'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';

const TARGET_LABEL: Record<string, string> = {
  customers: '客户档案',
  intent_customers: '意向客户',
  load: '负荷数据（96点曲线）',
  monthly_settlement: '月度结算',
  customer_energy: '客户电量档案（月度电量）',
  policy: '政策文件',
  contracts: '零售合同（草稿）',
};

// 分类 → 建议去向：解析完按文档类型自动归入对应模块。
// target 为入库目标（走「确认入库」）；contract=true 表示去零售合同（预填表单）。
const DOC_TYPE_ROUTE: Record<string, { label: string; target?: string; contract?: boolean }> = {
  客户清单: { label: '客户档案', target: 'customers' },
  资质: { label: '客户档案', target: 'customers' },
  合同: { label: '零售合同', contract: true },
  政策: { label: '政策文件库', target: 'policy' },
  规则: { label: '政策文件库', target: 'policy' },
  账单: { label: '客户历史电量', target: 'customer_energy' },
  结算单: { label: '月度结算', target: 'monthly_settlement' },
  负荷数据: { label: '负荷数据', target: 'load' },
};

// 手动补充字段时的常用业务字段（key 与入库映射器对齐，选对了才能参与入库）
const COMMON_FIELDS: { key: string; label: string }[] = [
  { key: 'customer_name', label: '客户名称' },
  { key: 'period', label: '结算月份' },
  { key: 'energy', label: '结算电量' },
  { key: 'energy_fee', label: '电能量费' },
  { key: 'capacity_fee', label: '容量费' },
  { key: 'total_fee', label: '总费用' },
  { key: 'date', label: '日期' },
  { key: 'curve', label: '96点曲线' },
  { key: 'party_a', label: '甲方' },
  { key: 'party_b', label: '乙方' },
  { key: 'price', label: '电价' },
  { key: 'total_amount', label: '合同金额' },
];

const STATUS_META: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'success' }> = {
  uploaded: { label: '排队中', variant: 'secondary' },
  parsing: { label: '解析中', variant: 'default' },
  parsed: { label: '已解析', variant: 'success' },
  failed: { label: '失败', variant: 'destructive' },
};

function fmtTime(s?: string | null): string {
  if (!s) return '-';
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

export default function DocumentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('document_management:write');
  const canDelete = has('document_management:delete');
  const canContract = has('retail_management:write');

  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [editing, setEditing] = useState<Record<number, string>>({});
  const [target, setTarget] = useState('');
  const [busy, setBusy] = useState(false);
  // 手动添加字段表单
  const [newFieldKey, setNewFieldKey] = useState('customer_name');
  const [newFieldLabel, setNewFieldLabel] = useState('');
  const [newFieldValue, setNewFieldValue] = useState('');
  const [newFieldGroup, setNewFieldGroup] = useState('0');

  const { data, isLoading } = useQuery({
    queryKey: ['document', id],
    queryFn: () => getDocument(id),
    refetchInterval: (q) => {
      const st = q.state.data?.document.status;
      return st === 'parsing' || st === 'uploaded' ? 4000 : false;
    },
  });

  const doc = data?.document;
  const exts = useMemo(() => data?.extractions ?? [], [data]);
  const applies = data?.applies ?? [];
  // 按分类得到的建议去向（解析完自动路由到对应模块）
  const route = doc?.doc_type ? DOC_TYPE_ROUTE[doc.doc_type] : undefined;

  // 组织列表：用于跨省提醒（org_id → 省名 + 省名清单）
  const { data: orgs } = useQuery({ queryKey: ['orgs'], queryFn: listOrgs });
  // 跨省提醒（非阻断）：提取信息里明确出现了与归属省不同的省名时提示人工确认归属。
  // 只在明确命中省名时报（高精度、不误拦）；不自动改省。
  const provinceWarning = useMemo(() => {
    if (!doc?.org_id || !orgs) return null;
    const docOrg = orgs.find((o) => o.id === doc.org_id)?.name;
    if (!docOrg) return null;
    const text = exts.map((e) => e.value_text || '').join(' ');
    for (const o of orgs) {
      if (!o.name || o.name === docOrg || o.name.includes('默认')) continue;
      const short = o.name.replace(/[省市]$/, '');
      if (text.includes(o.name) || (short.length >= 2 && text.includes(short))) {
        return { detected: o.name, docOrg };
      }
    }
    return null;
  }, [doc, orgs, exts]);
  const effectiveTarget = target || data?.suggest_target || '';

  // 提取字段按行组分桶（0=文档级）
  const groups = useMemo(() => {
    const m = new Map<number, DocumentExtraction[]>();
    for (const e of exts) {
      if (!m.has(e.group_no)) m.set(e.group_no, []);
      m.get(e.group_no)!.push(e);
    }
    return Array.from(m.entries()).sort((a, b) => a[0] - b[0]);
  }, [exts]);

  const onDownload = async (kind: 'original' | 'parsed') => {
    setError(null);
    try {
      const fallback = kind === 'parsed' ? `${doc?.filename ?? 'document'}.md` : doc?.filename ?? 'document';
      await downloadDocumentFile(id, kind, fallback);
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  const onAddField = async () => {
    const isCustom = newFieldKey === '__custom__';
    const label = isCustom ? newFieldLabel.trim() : (COMMON_FIELDS.find((f) => f.key === newFieldKey)?.label ?? '');
    if (!label) {
      setError('请填写字段名');
      return;
    }
    try {
      await addExtraction(id, {
        group_no: Number(newFieldGroup) || 0,
        field_key: isCustom ? undefined : newFieldKey,
        field_label: label,
        value_text: newFieldValue,
      });
      setNewFieldValue('');
      setNewFieldLabel('');
      qc.invalidateQueries({ queryKey: ['document', id] });
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  const onDeleteField = async (extId: number) => {
    try {
      await deleteExtraction(id, extId);
      qc.invalidateQueries({ queryKey: ['document', id] });
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  // 合同文档 → 预填零售合同（所有可识别字段自动映射，跳合同页核对修改后保存）
  const onCreateContract = () => {
    const val = (...keys: string[]) => {
      for (const k of keys) {
        const hit = exts.find((e) => e.field_key === k);
        if (hit?.value_text) return hit.value_text;
      }
      return '';
    };
    const valLabel = (label: string) => exts.find((e) => e.field_label?.includes(label))?.value_text ?? '';

    // 客户名称候选
    const partyA = val('party_a') || valLabel('甲方');
    const partyB = val('party_b') || valLabel('乙方');
    const customerName = val('customer_name', 'customer', 'company') || partyA || partyB;

    // 电量（支持多种字段名）
    const energyRaw = val('energy', 'electricity', 'total_energy', 'purchase_energy') || valLabel('电量') || valLabel('购电量');
    // 电价
    const priceRaw = val('price', 'unit_price', 'electricity_price') || valLabel('电价') || valLabel('单价') || valLabel('价格');
    // 合同金额
    const amountRaw = val('total_amount', 'contract_amount', 'amount') || valLabel('金额') || valLabel('总金额') || valLabel('合同金额');
    // 合同编号
    const contractNoRaw = val('contract_no', 'contract_number') || valLabel('合同编号') || valLabel('编号');
    // 签订日期
    const signDateRaw = val('sign_date', 'contract_date', 'signed_date') || valLabel('签订日期') || valLabel('签约日期');
    // 日期范围
    const startDateRaw = val('start_date', 'effective_date', 'valid_from') || valLabel('起始') || valLabel('开始');
    const endDateRaw = val('end_date', 'expire_date', 'valid_to', 'expiry_date') || valLabel('截止') || valLabel('到期') || valLabel('结束');
    // 绿电比例
    const greenRaw = val('green_ratio', 'green_power_ratio') || valLabel('绿电') || valLabel('绿电比例');
    // 套餐名称提示
    const packageNameHint = val('package_name', 'product_name', 'plan_name') || valLabel('套餐') || valLabel('产品');
    // 电压等级
    const voltageRaw = val('voltage_level', 'voltage') || valLabel('电压');
    // 结算方式
    const settlementRaw = val('settlement_method', 'settlement') || valLabel('结算方式') || valLabel('结算');

    // 收集所有提取字段原文（供参考）
    const allFields = exts.map((e) => ({
      key: e.field_key ?? '',
      label: e.field_label ?? e.field_key ?? '',
      value: e.value_text ?? '',
    }));

    setContractPrefill({
      fromDoc: doc?.filename,
      customerCandidates: [customerName, partyA, partyB].filter(Boolean),
      energyMwh: numText(energyRaw),
      greenRatio: greenRaw ? numText(greenRaw) : '',
      startMonth: monthOf(startDateRaw),
      endMonth: monthOf(endDateRaw),
      price: numText(priceRaw),
      totalAmount: numText(amountRaw),
      contractNo: contractNoRaw,
      signDate: signDateRaw,
      partyA,
      partyB,
      packageNameHint,
      voltageLevel: voltageRaw,
      settlementMethod: settlementRaw,
      allFields,
    });
    router.push('/retail/contracts');
  };

  const onSaveField = async (extId: number) => {
    const v = editing[extId];
    if (v === undefined) return;
    try {
      await updateExtraction(id, extId, v);
      setEditing((prev) => {
        const next = { ...prev };
        delete next[extId];
        return next;
      });
      qc.invalidateQueries({ queryKey: ['document', id] });
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  const applyTo = async (targetKey: string) => {
    if (!targetKey) {
      setError('请选择入库目标');
      return;
    }
    if (!window.confirm(`确认把提取结果归入「${TARGET_LABEL[targetKey] ?? targetKey}」？`)) return;
    setBusy(true);
    setError(null);
    setNotice(null);
    try {
      const r = await applyDocument(id, targetKey);
      setNotice(r.message);
      qc.invalidateQueries({ queryKey: ['document', id] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const onApply = () => applyTo(effectiveTarget);

  const onReparse = async () => {
    setError(null);
    try {
      await reparseDocument(id);
      setNotice('已加入重新解析队列');
      qc.invalidateQueries({ queryKey: ['document', id] });
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  const onDelete = async () => {
    if (!window.confirm('确认删除该文档及其归档文件？此操作不可恢复。')) return;
    try {
      await deleteDocument(id);
      router.push('/documents');
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  if (isLoading || !doc) {
    return <p className="text-sm text-muted-foreground">{isLoading ? '加载中...' : '文档不存在'}</p>;
  }

  const st = STATUS_META[doc.status] ?? { label: doc.status, variant: 'secondary' as const };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => router.push('/documents')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="max-w-[480px] truncate text-xl font-bold" title={doc.filename}>
              {doc.filename}
            </h1>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Badge variant={st.variant}>{st.label}</Badge>
              <span>{doc.doc_type ?? '未分类'} · {doc.page_count || '-'} 页/表 · 上传人 {doc.uploaded_by ?? '-'} · {fmtTime(doc.created_at)}</span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {data?.storage_enabled && doc.original_object_key && (
            <Button variant="outline" size="sm" onClick={() => onDownload('original')}>
              <Download className="mr-1 h-4 w-4" />原件
            </Button>
          )}
          {doc.status === 'parsed' && (
            <Button variant="outline" size="sm" onClick={() => onDownload('parsed')}>
              <FileText className="mr-1 h-4 w-4" />解析件
            </Button>
          )}
          {canWrite && doc.original_object_key && (
            <Button variant="outline" size="sm" onClick={onReparse}>
              <RefreshCw className="mr-1 h-4 w-4" />重新解析
            </Button>
          )}
          {canDelete && (
            <Button variant="destructive" size="sm" onClick={onDelete}>
              <Trash2 className="mr-1 h-4 w-4" />删除
            </Button>
          )}
        </div>
      </div>

      {doc.status === 'failed' && doc.error && (
        <Alert variant="destructive">
          <AlertDescription>解析失败：{doc.error}</AlertDescription>
        </Alert>
      )}
      {notice && (
        <Alert>
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* ── 跨省提醒（非阻断）：提取信息含其它省名时提示核对归属 ── */}
      {doc.status === 'parsed' && provinceWarning && (
        <Alert className="border-amber-300 bg-amber-50">
          <AlertDescription className="flex items-center gap-2 text-sm text-amber-800">
            <MapPin className="h-4 w-4 shrink-0" />
            提取信息疑似涉及「{provinceWarning.detected}」，但本文档归属「{provinceWarning.docOrg}」——
            请确认省份归属是否正确（不会自动改省）。
          </AlertDescription>
        </Alert>
      )}

      {/* ── 建议归入（按分类自动路由到对应模块）── */}
      {doc.status === 'parsed' && route && (
        <Alert className="border-primary/30 bg-primary/5">
          <AlertDescription className="flex flex-wrap items-center justify-between gap-2">
            {doc.auto_applied ? (
              <span className="text-sm">
                <CheckCircle2 className="mr-1 inline h-4 w-4 text-emerald-600" />
                已自动识别为「{doc.doc_type}」并归入「<span className="font-semibold">{route.label}</span>」
              </span>
            ) : (
              <>
                <span className="text-sm">
                  识别为「{doc.doc_type ?? '未分类'}」，建议归入{' '}
                  <span className="font-semibold text-primary">{route.label}</span>
                </span>
                {route.contract
                  ? canContract && (
                      <Button size="sm" onClick={onCreateContract}>
                        <FileSignature className="mr-1 h-4 w-4" />一键归入{route.label}
                      </Button>
                    )
                  : canWrite && (
                      <Button size="sm" onClick={() => applyTo(route.target!)} disabled={busy}>
                        <CheckCircle2 className="mr-1 h-4 w-4" />
                        {busy ? '归入中...' : `一键归入${route.label}`}
                      </Button>
                    )}
              </>
            )}
          </AlertDescription>
        </Alert>
      )}

      {/* ── 提取字段核对 + 确认入库 ── */}
      {doc.status === 'parsed' && (
        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">
                提取字段{doc.auto_applied ? '（已自动入库）' : '（核对后入库）'}
              </CardTitle>
              {canWrite && exts.length > 0 && (
                <div className="flex items-center gap-2">
                  <select value={effectiveTarget} onChange={(e) => setTarget(e.target.value)} className={SELECT_CLASS}>
                    <option value="">选择入库目标...</option>
                    {Object.entries(TARGET_LABEL).map(([k, v]) => (
                      <option key={k} value={k}>{v}</option>
                    ))}
                  </select>
                  <Button onClick={onApply} disabled={busy || !effectiveTarget}>
                    <CheckCircle2 className="mr-1 h-4 w-4" />
                    {busy ? '入库中...' : '确认入库'}
                  </Button>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent>
            {exts.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                未提取到结构化字段{doc.entities && Object.keys(doc.entities).length > 0 && '（下方有正则实体可参考）'}
              </p>
            ) : (
              <div className="max-h-[420px] overflow-auto rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-14">行</TableHead>
                      <TableHead>字段</TableHead>
                      <TableHead>值（可修正）</TableHead>
                      <TableHead className="w-20">单位</TableHead>
                      <TableHead className="w-20">置信度</TableHead>
                      <TableHead className="w-16">来源</TableHead>
                      {canWrite && <TableHead className="w-10" />}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {groups.map(([g, fields]) =>
                      fields.map((e) => (
                        <TableRow key={e.id}>
                          <TableCell className="text-muted-foreground">{g === 0 ? '-' : g}</TableCell>
                          <TableCell className="font-medium">{e.field_label}</TableCell>
                          <TableCell>
                            {canWrite ? (
                              <div className="flex items-center gap-1">
                                <input
                                  className="w-full rounded border bg-transparent px-2 py-1 text-sm"
                                  value={editing[e.id] ?? e.value_text ?? ''}
                                  onChange={(ev) =>
                                    setEditing((prev) => ({ ...prev, [e.id]: ev.target.value }))
                                  }
                                />
                                {editing[e.id] !== undefined && editing[e.id] !== (e.value_text ?? '') && (
                                  <Button size="sm" variant="outline" onClick={() => onSaveField(e.id)}>
                                    存
                                  </Button>
                                )}
                              </div>
                            ) : (
                              <span className="break-all">{e.value_text ?? '-'}</span>
                            )}
                          </TableCell>
                          <TableCell>{e.unit ?? '-'}</TableCell>
                          <TableCell>
                            {e.confidence != null ? (
                              <span className={e.confidence < 0.6 ? 'text-amber-600' : ''}>
                                {(e.confidence * 100).toFixed(0)}%
                              </span>
                            ) : '-'}
                          </TableCell>
                          <TableCell className="text-xs text-muted-foreground">{e.source}</TableCell>
                          {canWrite && (
                            <TableCell>
                              <button
                                title="删除该字段"
                                className="text-muted-foreground hover:text-destructive"
                                onClick={() => onDeleteField(e.id)}
                              >
                                <X className="h-4 w-4" />
                              </button>
                            </TableCell>
                          )}
                        </TableRow>
                      )),
                    )}
                  </TableBody>
                </Table>
              </div>
            )}

            {/* 手动补充字段：选常用业务字段（key 对齐入库映射）或自定义 */}
            {canWrite && (
              <div className="mt-3 flex flex-wrap items-center gap-2 rounded-lg border border-dashed p-3">
                <span className="text-sm text-muted-foreground">补充字段</span>
                <select
                  value={newFieldKey}
                  onChange={(e) => setNewFieldKey(e.target.value)}
                  className={SELECT_CLASS}
                >
                  {COMMON_FIELDS.map((f) => (
                    <option key={f.key} value={f.key}>{f.label}</option>
                  ))}
                  <option value="__custom__">自定义...</option>
                </select>
                {newFieldKey === '__custom__' && (
                  <input
                    className="h-9 w-32 rounded-md border bg-transparent px-2 text-sm"
                    placeholder="字段名"
                    value={newFieldLabel}
                    onChange={(e) => setNewFieldLabel(e.target.value)}
                  />
                )}
                <input
                  className="h-9 w-56 rounded-md border bg-transparent px-2 text-sm"
                  placeholder="值"
                  value={newFieldValue}
                  onChange={(e) => setNewFieldValue(e.target.value)}
                />
                <input
                  className="h-9 w-20 rounded-md border bg-transparent px-2 text-sm"
                  placeholder="行号"
                  title="行号：0=文档级；与某数据行同号则归入该行"
                  value={newFieldGroup}
                  onChange={(e) => setNewFieldGroup(e.target.value)}
                />
                <Button size="sm" variant="outline" onClick={onAddField}>
                  <Plus className="mr-1 h-4 w-4" />添加
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* ── 入库历史 ── */}
      {applies.length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">入库历史</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-1 text-sm">
              {applies.map((a) => (
                <li key={a.id} className="text-muted-foreground">
                  {fmtTime(a.applied_at)} — {a.applied_by ?? '?'} 写入「{TARGET_LABEL[a.target] ?? a.target}」{a.applied_rows} 行
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {/* ── 正则实体（参考） ── */}
      {doc.entities && Object.keys(doc.entities).length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">识别到的实体（参考）</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
            {Object.entries(doc.entities).map(([k, vals]) =>
              (vals ?? []).slice(0, 8).map((v, i) => (
                <Badge key={`${k}-${i}`} variant="outline">
                  {k}: {v}
                </Badge>
              )),
            )}
          </CardContent>
        </Card>
      )}

      {/* ── 解析全文 ── */}
      {doc.text_content && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">解析全文（Markdown）</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="max-h-[480px] overflow-auto whitespace-pre-wrap rounded-lg border bg-muted/30 p-4 text-xs leading-relaxed">
              {doc.text_content}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
