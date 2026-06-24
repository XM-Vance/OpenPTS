'use client';

import { useState, useMemo, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { DemoBadge } from '@/components/feedback';
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
} from '@/lib/api/monthly-manual';
import { Download, Upload, Clock, ArrowRight } from 'lucide-react';

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

// 模拟审计日志数据（基于手工数据项生成）
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

function generateAuditLog(items: { id: string; item_name: string; value: number; updated_at: string; created_by?: string | null }[]): AuditEntry[] {
  return items.slice(0, 10).map((item, idx) => {
    const actions = ['修改', '修改', '修改', '审核', '修改'];
    const fields = ['数值', '数值', '来源', '备注', '数值'];
    const action = actions[idx % actions.length];
    const field = fields[idx % fields.length];
    const oldVal = field === '数值' ? fmt(item.value * (0.9 + Math.random() * 0.1)) : '旧值';
    const newVal = field === '数值' ? fmt(item.value) : '新值';
    return {
      id: `audit-${idx}`,
      timestamp: item.updated_at.slice(0, 19).replace('T', ' '),
      operator: item.created_by ?? 'admin',
      action,
      field,
      oldValue: oldVal,
      newValue: newVal,
      itemId: item.id,
      itemName: item.item_name,
    };
  });
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

  const items = useMemo(() => data?.items ?? [], [data]);

  // 生成审计日志
  const auditLog = useMemo(() => generateAuditLog(items), [items]);

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
                accept=".csv"
                className="hidden"
                onChange={(e) => {
                  // 文件选择后的处理逻辑占位
                  e.target.value = '';
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
      {auditLog.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Clock className="h-4 w-4" />
              修改审计日志
              <DemoBadge tooltip="审计日志的旧值为随机生成，非真实变更记录" />
            </CardTitle>
          </CardHeader>
          <CardContent>
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
          </CardContent>
        </Card>
      )}

      {/* Diff 对比面板 */}
      {selectedAudit && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">修改对比 · {selectedAudit.itemName}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              {/* 修改前 */}
              <div className="rounded-lg border border-red-200 bg-red-50/50 p-4">
                <p className="text-xs font-semibold text-red-600 mb-2">修改前</p>
                <div className="space-y-1">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">字段</span>
                    <span>{selectedAudit.field}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">原值</span>
                    <span className="font-mono text-red-600 line-through">{selectedAudit.oldValue}</span>
                  </div>
                </div>
              </div>
              {/* 修改后 */}
              <div className="rounded-lg border border-green-200 bg-green-50/50 p-4">
                <p className="text-xs font-semibold text-green-600 mb-2">修改后</p>
                <div className="space-y-1">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">字段</span>
                    <span>{selectedAudit.field}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">新值</span>
                    <span className="font-mono text-green-600 font-bold">{selectedAudit.newValue}</span>
                  </div>
                </div>
              </div>
            </div>
            <div className="mt-3 flex items-center justify-center text-sm text-muted-foreground">
              <span className="font-mono text-red-500">{selectedAudit.oldValue}</span>
              <ArrowRight className="mx-3 h-4 w-4" />
              <span className="font-mono text-green-600 font-bold">{selectedAudit.newValue}</span>
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
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
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
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  暂无手工数据{canWrite && '，可点右上「生成演示数据」或「新增条目」'}
                </TableCell>
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
