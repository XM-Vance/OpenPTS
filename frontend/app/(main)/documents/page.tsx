'use client';

// 文档解析中心：上传(原件归档) → 异步解析 → 详情页核对字段 → 确认入库。
import { useState, useRef, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { usePermission } from '@/lib/auth/use-permission';
import { useAuth } from '@/lib/auth/context';
import { listDocuments, type DocumentItem } from '@/lib/api/documents';
import { FileScan, Loader2 } from 'lucide-react';
import UploadZone from './components/UploadZone';

const SELECT_CLASS = 'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const DOC_TYPES = ['合同', '政策', '规则', '账单', '结算单', '资质', '客户清单', '负荷数据', '其他'];

const STATUS_META: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'success' }> = {
  uploaded: { label: '排队中', variant: 'secondary' },
  parsing: { label: '解析中', variant: 'default' },
  parsed: { label: '已解析', variant: 'success' },
  failed: { label: '失败', variant: 'destructive' },
};

const KIND_LABEL: Record<string, string> = {
  pdf: 'PDF', image: '图片', word: 'Word', excel: 'Excel', csv: 'CSV',
};

function fmtTime(s: string): string {
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

function fmtSize(n: number): string {
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB';
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(0) + ' KB';
  return n + ' B';
}

export default function DocumentsPage() {
  const qc = useQueryClient();
  const router = useRouter();
  const { has } = usePermission();
  const { isHQ } = useAuth();
  const canWrite = has('document_management:write');

  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [custFilter, setCustFilter] = useState('');
  const [scopeMine, setScopeMine] = useState(false);

  const { data: docs, isLoading } = useQuery({
    queryKey: ['documents', statusFilter, typeFilter, scopeMine],
    queryFn: () =>
      listDocuments({
        status: statusFilter || undefined,
        doc_type: typeFilter || undefined,
        limit: 200,
        scope: scopeMine ? 'mine' : undefined,
      }),
    // 有文档在解析时自动轮询刷新状态
    refetchInterval: (q) =>
      (q.state.data ?? []).some((d) => d.status === 'parsing' || d.status === 'uploaded')
        ? 4000
        : false,
  });

  const items: DocumentItem[] = useMemo(() => {
    let list = docs ?? [];
    if (custFilter === 'linked') list = list.filter((d) => d.customer_id);
    if (custFilter === 'unlinked') list = list.filter((d) => !d.customer_id);
    return list;
  }, [docs, custFilter]);
  const parsingCount = items.filter((d) => d.status === 'parsing' || d.status === 'uploaded').length;

  const refreshList = () => qc.invalidateQueries({ queryKey: ['documents'] });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold">
            <FileScan className="h-6 w-6" />
            文档解析
          </h1>
          <p className="text-sm text-muted-foreground">
            PDF / 图片 / Word OCR 识别，Excel / CSV 直读 —— 高置信度数据自动入库，低置信度人工确认
          </p>
        </div>
        <div className="flex items-center gap-2">
          {parsingCount > 0 && (
            <Badge variant="default" className="gap-1">
              <Loader2 className="h-3 w-3 animate-spin" />
              {parsingCount} 个解析中
            </Badge>
          )}
          <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className={SELECT_CLASS}>
            <option value="">全部状态</option>
            <option value="parsing">解析中</option>
            <option value="parsed">已解析</option>
            <option value="failed">失败</option>
          </select>
          <select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)} className={SELECT_CLASS}>
            <option value="">全部类型</option>
            {DOC_TYPES.map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
          <select value={custFilter} onChange={(e) => setCustFilter(e.target.value)} className={SELECT_CLASS}>
            <option value="">全部文档</option>
            <option value="linked">已关联客户</option>
            <option value="unlinked">未关联</option>
          </select>
          {isHQ && (
            <label className="flex items-center gap-1.5 text-sm whitespace-nowrap">
              <input
                type="checkbox"
                checked={scopeMine}
                onChange={(e) => setScopeMine(e.target.checked)}
              />
              只看我的
            </label>
          )}
        </div>
      </div>

      {/* 上传区：拖拽 + 多文件 + 逐文件进度 */}
      {canWrite && <UploadZone onComplete={refreshList} />}

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>文件名</TableHead>
              <TableHead>格式</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">页/表</TableHead>
              <TableHead className="text-right">大小</TableHead>
              <TableHead>关联</TableHead>
              <TableHead>上传人</TableHead>
              <TableHead>时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={9} className="text-center text-muted-foreground">加载中...</TableCell>
              </TableRow>
            )}
            {items.map((d) => {
              const st = STATUS_META[d.status] ?? { label: d.status, variant: 'secondary' as const };
              return (
                <TableRow
                  key={d.id}
                  className="cursor-pointer"
                  onClick={() => router.push(`/documents/${d.id}`)}
                >
                  <TableCell className="max-w-[320px] truncate font-medium" title={d.filename}>
                    {d.filename}
                  </TableCell>
                  <TableCell>{KIND_LABEL[d.source_kind] ?? d.source_kind}</TableCell>
                  <TableCell>{d.doc_type ?? '-'}</TableCell>
                  <TableCell>
                    <Badge variant={st.variant} title={d.error ?? undefined}>{st.label}</Badge>
                    {d.auto_applied && d.status === 'parsed' && (
                      <Badge variant="success" className="ml-1" title="系统自动入库（置信度达标）">自动入库</Badge>
                    )}
                    {!d.auto_applied && d.status === 'parsed' && (
                      <Badge variant="secondary" className="ml-1" title="需人工确认入库">待确认</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">{d.page_count || '-'}</TableCell>
                  <TableCell className="text-right">{fmtSize(d.size)}</TableCell>
                  <TableCell>
                    {d.customer_id && (
                      <Badge variant="success" className="cursor-pointer" title={d.customer_id}
                        onClick={(e) => { e.stopPropagation(); router.push(`/customers/${d.customer_id}/360`); }}>
                        客户
                      </Badge>
                    )}
                    {d.contract_id && (
                      <Badge variant="default" className="ml-1 cursor-pointer" title={d.contract_id}
                        onClick={(e) => { e.stopPropagation(); router.push(`/retail/contracts?highlight=${d.contract_id}`); }}>
                        合同
                      </Badge>
                    )}
                    {!d.customer_id && !d.contract_id && <span className="text-muted-foreground">-</span>}
                  </TableCell>
                  <TableCell>{d.uploaded_by ?? '-'}</TableCell>
                  <TableCell>{fmtTime(d.created_at)}</TableCell>
                </TableRow>
              );
            })}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={9} className="text-center text-muted-foreground">
                  暂无文档{canWrite && '，请上传 PDF / 图片 / Word / Excel'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
