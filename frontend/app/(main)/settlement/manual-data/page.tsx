'use client';

import { useState, useMemo, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
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
  createManualItem,
  genManualDemo,
  listManualItems,
  importManualData,
  type ManualImportResult,
} from '@/lib/api/monthly-manual';
import { Download, Upload, Clock, ArrowRight } from 'lucide-react';
import { EmptyState } from '@/components/feedback';
import { Skeleton } from '@/components/ui/skeleton';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const CATEGORIES = ['', '收入', '成本', '偏差', '其他'];
const fmt = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 2 });

function categoryVariant(c: string): 'default' | 'secondary' | 'destructive' | 'success' {
  if (c === '收入') return 'success';
  if (c === '成本') return 'default';
  if (c === '偏差') return 'destructive';
  return 'secondary';
}

// 审计日志类型定义（真实审计 API 接入后填充数据）
interface AuditEntry {
  id: string;
  timestamp: string;
  operator: string;
  action: string;
  field: string;
  oldValue: string;
  newValue: string;
  itemId: string;
  itemName: string;
}

export default function ManualDataPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('settlement_management:write');

  const [category, setCategory] = useState('');
  const [month, setMonth] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [selectedAudit, setSelectedAudit] = useState<AuditEntry | null>(null);
  const [importPreview, setImportPreview] = useState<ManualImportResult | null>(null);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['manual-items', category, month],
    queryFn: () => listManualItems({ category, month }),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genManualDemo();
      qc.invalidateQueries({ queryKey: ['manual-items'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  // 批量导入（WP1.6）：选文件 → dry_run 预览 → 确认入库
  const onImportFile = async (f: File) => {
    setBusy(true);
    setError(null);
    try {
      const res = await importManualData(f, true);
      setImportFile(f);
      setImportPreview(res);
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const confirmImport = async () => {
    if (!importFile) return;
    setImporting(true);
    try {
      await importManualData(importFile, false);
      setImportPreview(null);
      setImportFile(null);
      qc.invalidateQueries({ queryKey: ['manual-items'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setImporting(false);
    }
  };

  const items = useMemo(() => data?.items ?? [], [data]);

  // 审计日志暂无真实后端 API 接入，置空以区分「无数据」与「真实数据」
  const auditLog: AuditEntry[] = [];

  // 导出CSV
  const handleExportCSV = useCallback(() => {
    if (items.length === 0) return;
    const headers = ['月份', '分类', '项目', '数值', '单位', '来源', '录入人', '更新时间'];
    const rows = items.map((i) => [
      i.operating_month,
      i.category,
      i.item_name,
      String(i.value),
      i.unit,
      i.source ?? '',
      i.created_by ?? '',
      i.updated_at.slice(0, 19).replace('T', ' '),
    ]);
    const csv = [headers.join(','), ...rows.map((r) => r.map((c) => `"${c}"`).join(','))].join('\n');
    const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `手工数据_${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }, [items]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">月度手工数据</h1>
          <p className="text-sm text-muted-foreground">
            人工录入的月度运营数据：收入 / 成本 / 偏差 / 其他
          </p>
        </div>
        <div className="flex gap-2">
          {canWrite && (
            <>
              <Button onClick={() => setCreating(true)}>新增条目</Button>
              <Button variant="outline" onClick={onGen} disabled={busy}>
                {busy ? '生成中...' : '生成演示数据'}
              </Button>
            </>
          )}
          {/* 批量导入/导出按钮 */}
          <Button variant="outline" onClick={handleExportCSV} disabled={items.length === 0}>
            <Download className="mr-1 h-4 w-4" />
            导出CSV
          </Button>
          <Button variant="outline" asChild>
            <label className="cursor-pointer">
              <Upload className="mr-1 h-4 w-4" />
              批量导入
              <input
                type="file"
                accept=".csv,.xlsx,.txt"
                className="hidden"
                onChange={(e) => {
                  const f = e.target.files?.[0];
                  e.target.value = '';
                  if (f) onImportFile(f);
                }}
              />
            </label>
          </Button>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">筛选</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-3">
            <div className="space-y-1">
              <Label>分类</Label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className={SELECT_CLASS}
              >
                {CATEGORIES.map((c) => (
                  <option key={c || 'all'} value={c}>
                    {c || '全部分类'}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-1">
              <Label>年月（如 2026-05）</Label>
              <Input value={month} onChange={(e) => setMonth(e.target.value)} placeholder="可选" />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 修改审计日志时间线 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Clock className="h-4 w-4" />
            修改审计日志
          </CardTitle>
        </CardHeader>
        <CardContent>
          {auditLog.length === 0 ? (
            <EmptyState compact className="py-8" title="暂无审计记录" />
          ) : (
            <div className="relative">
              {/* 时间线竖线 */}
              <div className="absolute left-[7px] top-2 bottom-2 w-0.5 bg-border" />
              <div className="space-y-3">
                {auditLog.map((entry) => (
                  <div
                    key={entry.id}
                    className="relative ml-5 cursor-pointer rounded-lg border p-3 hover:bg-muted/50 transition-colors"
                    onClick={() => setSelectedAudit(entry)}
                  >
                    {/* 时间线圆点 */}
                    <div className="absolute -left-[21px] top-4 h-3 w-3 rounded-full border-2 border-primary bg-background" />
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge
                          variant={entry.action === '审核' ? 'success' : 'secondary'}
                          className="text-xs"
                        >
                          {entry.action}
                        </Badge>
                        <span className="text-sm font-medium">{entry.itemName}</span>
                        <span className="text-xs text-muted-foreground">
                          → {entry.field}
                        </span>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-muted-foreground">
                        <span>{entry.operator}</span>
                        <span>{entry.timestamp}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Diff 对比面板 */}
      {/* 批量导入预览弹窗（WP1.6） */}
      <Dialog open={!!importPreview} onClose={() => setImportPreview(null)}>
        <DialogHeader>
          <DialogTitle>批量导入预览</DialogTitle>
        </DialogHeader>
        <div className="max-h-80 overflow-y-auto">
          <p className="mb-2 text-xs text-muted-foreground">
            共 {importPreview?.total ?? 0} 行，有效 {importPreview?.valid ?? 0} 行
            （格式：月份, 类别, 条目, 数值, 单位, 来源?, 备注?）
          </p>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>行</TableHead>
                <TableHead>月份</TableHead>
                <TableHead>类别/条目</TableHead>
                <TableHead className="text-right">数值</TableHead>
                <TableHead>校验</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(importPreview?.rows ?? []).map((r) => (
                <TableRow key={r.line_no}>
                  <TableCell>{r.line_no}</TableCell>
                  <TableCell>{r.operating_month}</TableCell>
                  <TableCell>{r.category} / {r.item_name}</TableCell>
                  <TableCell className="text-right">{r.value}</TableCell>
                  <TableCell className="text-xs">
                    {r.errors?.length ? (
                      <span className="text-destructive">{r.errors.join('；')}</span>
                    ) : (
                      <span className="text-emerald-700">通过</span>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setImportPreview(null)}>
            取消
          </Button>
          <Button onClick={confirmImport} disabled={importing || (importPreview?.valid ?? 0) === 0}>
            {importing ? '导入中...' : `确认导入 ${importPreview?.valid ?? 0} 行`}
          </Button>
        </DialogFooter>
      </Dialog>

      {selectedAudit && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">修改对比 · {selectedAudit.itemName}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              {/* 修改前 */}
              <div className="rounded-lg border border-red-200 dark:border-red-500/30 bg-red-50/50 dark:bg-red-500/10 p-4">
                <p className="text-xs font-semibold text-red-700 mb-2">修改前</p>
                <div className="space-y-1">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">字段</span>
                    <span>{selectedAudit.field}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">原值</span>
                    <span className="font-mono text-red-700 line-through">{selectedAudit.oldValue}</span>
                  </div>
                </div>
              </div>
              {/* 修改后 */}
              <div className="rounded-lg border border-green-200 dark:border-green-500/30 bg-green-50/50 dark:bg-green-500/10 p-4">
                <p className="text-xs font-semibold text-green-700 mb-2">修改后</p>
                <div className="space-y-1">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">字段</span>
                    <span>{selectedAudit.field}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">新值</span>
                    <span className="font-mono text-green-700 font-bold">{selectedAudit.newValue}</span>
                  </div>
                </div>
              </div>
            </div>
            <div className="mt-3 flex items-center justify-center text-sm text-muted-foreground">
              <span className="font-mono text-red-500">{selectedAudit.oldValue}</span>
              <ArrowRight className="mx-3 h-4 w-4" />
              <span className="font-mono text-green-700 font-bold">{selectedAudit.newValue}</span>
            </div>
            <div className="mt-2 flex items-center justify-between text-xs text-muted-foreground">
              <span>操作人：{selectedAudit.operator}</span>
              <span>时间：{selectedAudit.timestamp}</span>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>月份</TableHead>
              <TableHead>分类</TableHead>
              <TableHead>项目</TableHead>
              <TableHead className="text-right">数值</TableHead>
              <TableHead>单位</TableHead>
              <TableHead>来源</TableHead>
              <TableHead>录入人</TableHead>
              <TableHead>更新时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={8}><Skeleton className="h-5 w-full" /></TableCell>
              </TableRow>
            )}
            {items.map((i) => (
              <TableRow key={i.id}>
                <TableCell className="font-medium">{i.operating_month}</TableCell>
                <TableCell>
                  <Badge variant={categoryVariant(i.category)}>{i.category}</Badge>
                </TableCell>
                <TableCell>{i.item_name}</TableCell>
                <TableCell className="text-right">{fmt(i.value)}</TableCell>
                <TableCell>{i.unit}</TableCell>
                <TableCell className="text-muted-foreground">{i.source ?? '-'}</TableCell>
                <TableCell>{i.created_by ?? '-'}</TableCell>
                <TableCell className="text-xs text-muted-foreground">
                  {i.updated_at.slice(0, 19).replace('T', ' ')}
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={8}><EmptyState compact title={<> 暂无手工数据{canWrite && '，可点右上「生成演示数据」或「新增条目」'} </>} /></TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {creating && (
        <ManualCreateDialog
          onClose={() => setCreating(false)}
          onSaved={() => {
            setCreating(false);
            qc.invalidateQueries({ queryKey: ['manual-items'] });
          }}
        />
      )}
    </div>
  );
}

function ManualCreateDialog({
  onClose,
  onSaved,
}: {
  onClose: () => void;
  onSaved: () => void;
}) {
  const [month, setMonth] = useState('');
  const [category, setCategory] = useState('收入');
  const [itemName, setItemName] = useState('');
  const [value, setValue] = useState('');
  const [unit, setUnit] = useState('元');
  const [source, setSource] = useState('');
  const [note, setNote] = useState('');
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    setErr(null);
    if (!month || !category || !itemName) {
      setErr('请填写月份 / 分类 / 项目名');
      return;
    }
    // 月份格式校验（F10）：必须 YYYY-MM，避免 "2026-5"/"20265" 等非法值直传后端
    if (!/^\d{4}-(0[1-9]|1[0-2])$/.test(month)) {
      setErr('月份格式应为 YYYY-MM（如 2026-05）');
      return;
    }
    const v = Number(value);
    if (Number.isNaN(v)) {
      setErr('数值不合法');
      return;
    }
    setBusy(true);
    try {
      await createManualItem({
        operating_month: month,
        category,
        item_name: itemName,
        value: v,
        unit,
        source: source || undefined,
        note: note || undefined,
      });
      onSaved();
    } catch (e) {
      setErr(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog open onClose={onClose}>
      <DialogHeader>
        <DialogTitle>新增手工数据</DialogTitle>
      </DialogHeader>
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <Label>年月 (YYYY-MM)</Label>
            <Input value={month} onChange={(e) => setMonth(e.target.value)} placeholder="2026-05" />
          </div>
          <div className="space-y-1">
            <Label>分类</Label>
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              className={SELECT_CLASS}
            >
              <option value="收入">收入</option>
              <option value="成本">成本</option>
              <option value="偏差">偏差</option>
              <option value="其他">其他</option>
            </select>
          </div>
        </div>
        <div className="space-y-1">
          <Label>项目名</Label>
          <Input value={itemName} onChange={(e) => setItemName(e.target.value)} />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <Label>数值</Label>
            <Input type="number" value={value} onChange={(e) => setValue(e.target.value)} />
          </div>
          <div className="space-y-1">
            <Label>单位</Label>
            <Input value={unit} onChange={(e) => setUnit(e.target.value)} />
          </div>
        </div>
        <div className="space-y-1">
          <Label>来源</Label>
          <Input value={source} onChange={(e) => setSource(e.target.value)} placeholder="可选" />
        </div>
        <div className="space-y-1">
          <Label>备注</Label>
          <Input value={note} onChange={(e) => setNote(e.target.value)} placeholder="可选" />
        </div>
        {err && <p className="text-xs text-destructive">{err}</p>}
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          取消
        </Button>
        <Button onClick={submit} disabled={busy}>
          {busy ? '保存中...' : '保存'}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
