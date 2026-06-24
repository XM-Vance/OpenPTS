'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  EditorDialog,
} from '@/components/forms/editor-dialog';
import { ConfirmDialog } from '@/components/feedback/confirm-dialog';
import {
  listTradeRules,
  createTradeRule,
  updateTradeRule,
  deleteTradeRule,
  exportTradeRules,
  type TradeRule,
  type TradeRuleInput,
} from '@/lib/api/trade-rules';
import { BookOpen, Plus, Pencil, Trash2, Download } from 'lucide-react';

const CATEGORIES = [
  { value: '', label: '全部类别' },
  { value: 'settlement', label: '结算规则' },
  { value: 'deviation', label: '偏差考核' },
  { value: 'green', label: '绿电规则' },
  { value: 'registration', label: '注册规则' },
];

const CATEGORY_LABELS: Record<string, string> = {
  settlement: '结算规则',
  deviation: '偏差考核',
  green: '绿电规则',
  registration: '注册规则',
};

const CATEGORY_COLORS: Record<string, string> = {
  settlement: 'bg-blue-100 text-blue-700',
  deviation: 'bg-amber-100 text-amber-700',
  green: 'bg-emerald-100 text-emerald-700',
  registration: 'bg-purple-100 text-purple-700',
};

const SELECT_CLASS = 'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';

/* ── Rule Editor Dialog ── */
function TradeRuleEditorDialog({
  open,
  mode,
  rule,
  onClose,
  onSave,
}: {
  open: boolean;
  mode: 'create' | 'edit';
  rule?: TradeRule | null;
  onClose: () => void;
  onSave: () => void;
}) {
  const [formData, setFormData] = useState<TradeRuleInput>({
    category: 'settlement',
    rule_key: '',
    rule_value: '',
    effective_date: '',
    expiry_date: null,
    description: '',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useMemo(() => {
    if (open) {
      if (rule && mode === 'edit') {
        setFormData({
          category: rule.category,
          rule_key: rule.rule_key,
          rule_value: rule.rule_value,
          effective_date: rule.effective_date?.slice(0, 10) ?? '',
          expiry_date: rule.expiry_date?.slice(0, 10) ?? null,
          description: rule.description ?? '',
        });
      } else {
        setFormData({
          category: 'settlement',
          rule_key: '',
          rule_value: '',
          effective_date: '',
          expiry_date: null,
          description: '',
        });
      }
      setError(null);
    }
  }, [open, rule, mode]);

  const handleSave = async () => {
    if (!formData.rule_key.trim()) {
      setError('请输入规则标识');
      return;
    }
    if (!formData.rule_value.trim()) {
      setError('请输入规则值');
      return;
    }
    if (!formData.effective_date) {
      setError('请选择生效日期');
      return;
    }
    setSaving(true);
    setError(null);
    try {
      if (mode === 'create') {
        await createTradeRule(formData);
      } else if (rule) {
        await updateTradeRule(rule.id, formData);
      }
      onSave();
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '保存失败';
      setError(msg);
    } finally {
      setSaving(false);
    }
  };

  const handleChange = (key: keyof TradeRuleInput, value: unknown) => {
    setFormData((prev) => ({ ...prev, [key]: value }));
  };

  return (
    <EditorDialog
      open={open}
      onClose={onClose}
      onSave={handleSave}
      title={mode === 'create' ? '新增规则' : `编辑规则: ${rule?.rule_key ?? ''}`}
      mode={mode}
      width="max-w-md"
      saving={saving}
      error={error}
    >
      <div className="space-y-4">
        <div className="space-y-1.5">
          <Label htmlFor="rule_category">规则类别 *</Label>
          <select
            id="rule_category"
            value={formData.category}
            onChange={(e) => handleChange('category', e.target.value)}
            className={SELECT_CLASS + ' w-full'}
          >
            {CATEGORIES.filter((c) => c.value).map((c) => (
              <option key={c.value} value={c.value}>{c.label}</option>
            ))}
          </select>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="rule_key">规则标识 *</Label>
          <Input
            id="rule_key"
            value={formData.rule_key}
            onChange={(e) => handleChange('rule_key', e.target.value)}
            placeholder="如：settlement.price_spread"
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="rule_value">规则值 *</Label>
          <Input
            id="rule_value"
            value={formData.rule_value}
            onChange={(e) => handleChange('rule_value', e.target.value)}
            placeholder="如：0.05"
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <Label htmlFor="effective_date">生效日期 *</Label>
            <Input
              id="effective_date"
              type="date"
              value={formData.effective_date}
              onChange={(e) => handleChange('effective_date', e.target.value)}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="expiry_date">失效日期</Label>
            <Input
              id="expiry_date"
              type="date"
              value={formData.expiry_date ?? ''}
              onChange={(e) => handleChange('expiry_date', e.target.value || null)}
            />
          </div>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="rule_description">描述</Label>
          <Input
            id="rule_description"
            value={formData.description ?? ''}
            onChange={(e) => handleChange('description', e.target.value)}
            placeholder="规则说明（可选）"
          />
        </div>
      </div>
    </EditorDialog>
  );
}

/* ── Main Page ── */
export default function TradeRulesPage() {
  const qc = useQueryClient();
  const [categoryFilter, setCategoryFilter] = useState('');
  const [editing, setEditing] = useState<TradeRule | 'new' | null>(null);
  const [deleting, setDeleting] = useState<TradeRule | null>(null);
  const [exporting, setExporting] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['trade-rules', categoryFilter],
    queryFn: () => listTradeRules(categoryFilter || undefined),
  });

  const items: TradeRule[] = useMemo(() => data?.items ?? [], [data]);

  // Group by category
  const grouped = useMemo(() => {
    const map: Record<string, TradeRule[]> = {};
    for (const r of items) {
      const key = r.category;
      if (!map[key]) map[key] = [];
      map[key].push(r);
    }
    return map;
  }, [items]);

  const handleDelete = async () => {
    if (!deleting) return;
    try {
      await deleteTradeRule(deleting.id);
      qc.invalidateQueries({ queryKey: ['trade-rules'] });
      setDeleting(null);
    } catch (err) {
      window.alert(err instanceof Error ? err.message : '删除失败');
    }
  };

  const handleExport = async () => {
    setExporting(true);
    try {
      const blob = await exportTradeRules();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `trade-rules-${new Date().toISOString().slice(0, 10)}.xlsx`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (err) {
      window.alert(err instanceof Error ? err.message : '导出失败');
    } finally {
      setExporting(false);
    }
  };

  const isExpired = (rule: TradeRule) => {
    if (!rule.expiry_date) return false;
    return new Date(rule.expiry_date) < new Date();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold">
            <BookOpen className="h-6 w-6" />
            福建交易规则
          </h1>
          <p className="text-sm text-muted-foreground">
            管理福建省电力交易结算、偏差考核、绿电等规则配置
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className={SELECT_CLASS}
          >
            {CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>{c.label}</option>
            ))}
          </select>
          <Button variant="outline" onClick={handleExport} disabled={exporting}>
            <Download className="mr-1 h-4 w-4" />
            {exporting ? '导出中...' : '导出'}
          </Button>
          <Button onClick={() => setEditing('new')}>
            <Plus className="mr-1 h-4 w-4" />
            新增规则
          </Button>
        </div>
      </div>

      {/* Group by category */}
      {isLoading && (
        <div className="py-12 text-center text-sm text-muted-foreground">加载中...</div>
      )}

      {!isLoading && items.length === 0 && (
        <div className="py-12 text-center text-sm text-muted-foreground">
          暂无规则，点击「新增规则」添加
        </div>
      )}

      {!isLoading &&
        Object.entries(grouped).map(([category, rules]) => (
          <div key={category} className="space-y-2">
            <h3 className="flex items-center gap-2 text-sm font-medium">
              <span
                className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                  CATEGORY_COLORS[category] ?? 'bg-gray-100 text-gray-700'
                }`}
              >
                {CATEGORY_LABELS[category] ?? category}
              </span>
              <span className="text-muted-foreground">（{rules.length} 条）</span>
            </h3>
            <div className="rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>规则标识</TableHead>
                    <TableHead>规则值</TableHead>
                    <TableHead>生效日期</TableHead>
                    <TableHead>失效日期</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>描述</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((r) => (
                    <TableRow key={r.id} className={isExpired(r) ? 'opacity-50' : ''}>
                      <TableCell className="font-mono text-sm">{r.rule_key}</TableCell>
                      <TableCell className="font-medium">{r.rule_value}</TableCell>
                      <TableCell>{r.effective_date?.slice(0, 10)}</TableCell>
                      <TableCell>{r.expiry_date?.slice(0, 10) ?? '长期'}</TableCell>
                      <TableCell>
                        {isExpired(r) ? (
                          <Badge variant="secondary">已失效</Badge>
                        ) : (
                          <Badge variant="default">生效中</Badge>
                        )}
                      </TableCell>
                      <TableCell className="max-w-[200px] truncate text-muted-foreground">
                        {r.description ?? '-'}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="sm" onClick={() => setEditing(r)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => setDeleting(r)}>
                            <Trash2 className="h-4 w-4 text-red-500" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        ))}

      {/* Editor Dialog */}
      <TradeRuleEditorDialog
        open={editing !== null}
        mode={editing === 'new' ? 'create' : 'edit'}
        rule={editing !== null && editing !== 'new' ? editing : null}
        onClose={() => setEditing(null)}
        onSave={() => qc.invalidateQueries({ queryKey: ['trade-rules'] })}
      />

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleting !== null}
        onClose={() => setDeleting(null)}
        onConfirm={handleDelete}
        title="删除规则"
        description={`确认删除规则「${deleting?.rule_key ?? ''}」？删除后不可恢复。`}
        confirmText="删除"
        variant="destructive"
      />
    </div>
  );
}
