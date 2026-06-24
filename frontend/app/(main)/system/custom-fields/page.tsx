'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
  type EditorDialogProps,
} from '@/components/forms/editor-dialog';
import { ConfirmDialog } from '@/components/feedback/confirm-dialog';
import {
  listCustomFields,
  createCustomField,
  updateCustomField,
  deleteCustomField,
  type CustomField,
  type CustomFieldInput,
} from '@/lib/api/custom-fields';
import { Settings2, Plus, Pencil, Trash2 } from 'lucide-react';

const ENTITY_TYPES = [
  { value: '', label: '全部实体' },
  { value: 'customer', label: '客户' },
  { value: 'contract', label: '合同' },
  { value: 'document', label: '文档' },
  { value: 'intent_customer', label: '意向客户' },
];

const FIELD_TYPES = [
  { value: 'text', label: '文本' },
  { value: 'number', label: '数字' },
  { value: 'date', label: '日期' },
  { value: 'select', label: '下拉选择' },
  { value: 'multi_select', label: '多选' },
  { value: 'boolean', label: '布尔' },
];

const SELECT_CLASS = 'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';

/* ── Editor Dialog ── */
function CustomFieldEditorDialog({
  open,
  mode,
  field,
  onClose,
  onSave,
}: {
  open: boolean;
  mode: 'create' | 'edit';
  field?: CustomField | null;
  onClose: () => void;
  onSave: () => void;
}) {
  const [formData, setFormData] = useState<CustomFieldInput>({
    entity_type: 'customer',
    field_label: '',
    field_key: '',
    field_type: 'text',
    is_required: false,
    sort_order: 0,
    options: [],
    default_value: '',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Reset form when dialog opens
  useMemo(() => {
    if (open) {
      if (field && mode === 'edit') {
        setFormData({
          entity_type: field.entity_type,
          field_label: field.field_label,
          field_key: field.field_key,
          field_type: field.field_type,
          is_required: field.is_required,
          sort_order: field.sort_order,
          options: field.options ?? [],
          default_value: field.default_value ?? '',
        });
      } else {
        setFormData({
          entity_type: 'customer',
          field_label: '',
          field_key: '',
          field_type: 'text',
          is_required: false,
          sort_order: 0,
          options: [],
          default_value: '',
        });
      }
      setError(null);
    }
  }, [open, field, mode]);

  const handleSave = async () => {
    if (!formData.field_label.trim()) {
      setError('请输入字段名称');
      return;
    }
    if (!formData.field_key.trim()) {
      setError('请输入字段标识');
      return;
    }
    setSaving(true);
    setError(null);
    try {
      if (mode === 'create') {
        await createCustomField(formData);
      } else if (field) {
        await updateCustomField(field.id, formData);
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

  const handleChange = (key: keyof CustomFieldInput, value: unknown) => {
    setFormData((prev) => ({ ...prev, [key]: value }));
  };

  return (
    <EditorDialog
      open={open}
      onClose={onClose}
      onSave={handleSave}
      title={mode === 'create' ? '新增字段' : `编辑字段: ${field?.field_label ?? ''}`}
      mode={mode}
      width="max-w-md"
      saving={saving}
      error={error}
    >
      <div className="space-y-4">
        <div className="space-y-1.5">
          <Label htmlFor="entity_type">实体类型 *</Label>
          <select
            id="entity_type"
            value={formData.entity_type}
            onChange={(e) => handleChange('entity_type', e.target.value)}
            className={SELECT_CLASS + ' w-full'}
          >
            {ENTITY_TYPES.filter((t) => t.value).map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="field_label">字段名称 *</Label>
          <Input
            id="field_label"
            value={formData.field_label}
            onChange={(e) => handleChange('field_label', e.target.value)}
            placeholder="如：合同编号"
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="field_key">字段标识 *</Label>
          <Input
            id="field_key"
            value={formData.field_key}
            onChange={(e) => handleChange('field_key', e.target.value)}
            placeholder="如：contract_no"
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="field_type">字段类型</Label>
          <select
            id="field_type"
            value={formData.field_type}
            onChange={(e) => handleChange('field_type', e.target.value)}
            className={SELECT_CLASS + ' w-full'}
          >
            {FIELD_TYPES.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
        </div>
        {(formData.field_type === 'select' || formData.field_type === 'multi_select') && (
          <div className="space-y-1.5">
            <Label htmlFor="options">选项（逗号分隔）</Label>
            <Input
              id="options"
              value={(formData.options ?? []).join(',')}
              onChange={(e) =>
                handleChange(
                  'options',
                  e.target.value ? e.target.value.split(',').map((s) => s.trim()) : [],
                )
              }
              placeholder="选项1,选项2,选项3"
            />
          </div>
        )}
        <div className="flex items-center gap-4">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={formData.is_required ?? false}
              onChange={(e) => handleChange('is_required', e.target.checked)}
              className="rounded border-input"
            />
            必填
          </label>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="sort_order">排序</Label>
          <Input
            id="sort_order"
            type="number"
            value={formData.sort_order ?? 0}
            onChange={(e) => handleChange('sort_order', Number(e.target.value))}
          />
        </div>
      </div>
    </EditorDialog>
  );
}

/* ── Main Page ── */
export default function CustomFieldsPage() {
  const qc = useQueryClient();
  const [entityFilter, setEntityFilter] = useState('');
  const [editing, setEditing] = useState<CustomField | 'new' | null>(null);
  const [deleting, setDeleting] = useState<CustomField | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['custom-fields', entityFilter],
    queryFn: () => listCustomFields(entityFilter || undefined),
  });

  const items: CustomField[] = useMemo(() => data?.items ?? [], [data]);

  const ENTITY_LABELS: Record<string, string> = {
    customer: '客户',
    contract: '合同',
    document: '文档',
    intent_customer: '意向客户',
  };

  const FIELD_TYPE_LABELS: Record<string, string> = Object.fromEntries(
    FIELD_TYPES.map((t) => [t.value, t.label]),
  );

  const handleDelete = async () => {
    if (!deleting) return;
    try {
      await deleteCustomField(deleting.id);
      qc.invalidateQueries({ queryKey: ['custom-fields'] });
      setDeleting(null);
    } catch (err) {
      window.alert(err instanceof Error ? err.message : '删除失败');
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold">
            <Settings2 className="h-6 w-6" />
            自定义字段管理
          </h1>
          <p className="text-sm text-muted-foreground">
            管理各实体的自定义扩展字段，支持文本、数字、日期、下拉等类型
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={entityFilter}
            onChange={(e) => setEntityFilter(e.target.value)}
            className={SELECT_CLASS}
          >
            {ENTITY_TYPES.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
          <Button onClick={() => setEditing('new')}>
            <Plus className="mr-1 h-4 w-4" />
            新增字段
          </Button>
        </div>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>实体类型</TableHead>
              <TableHead>字段名称</TableHead>
              <TableHead>字段标识</TableHead>
              <TableHead>字段类型</TableHead>
              <TableHead className="text-center">必填</TableHead>
              <TableHead className="text-right">排序</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((f) => (
              <TableRow key={f.id}>
                <TableCell>
                  <span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700">
                    {ENTITY_LABELS[f.entity_type] ?? f.entity_type}
                  </span>
                </TableCell>
                <TableCell className="font-medium">{f.field_label}</TableCell>
                <TableCell className="font-mono text-sm text-muted-foreground">{f.field_key}</TableCell>
                <TableCell>{FIELD_TYPE_LABELS[f.field_type] ?? f.field_type}</TableCell>
                <TableCell className="text-center">
                  {f.is_required ? (
                    <span className="text-red-500">是</span>
                  ) : (
                    <span className="text-muted-foreground">否</span>
                  )}
                </TableCell>
                <TableCell className="text-right">{f.sort_order}</TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Button variant="ghost" size="sm" onClick={() => setEditing(f)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => setDeleting(f)}>
                      <Trash2 className="h-4 w-4 text-red-500" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  暂无自定义字段，点击「新增字段」添加
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Editor Dialog */}
      <CustomFieldEditorDialog
        open={editing !== null}
        mode={editing === 'new' ? 'create' : 'edit'}
        field={editing !== null && editing !== 'new' ? editing : null}
        onClose={() => setEditing(null)}
        onSave={() => qc.invalidateQueries({ queryKey: ['custom-fields'] })}
      />

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleting !== null}
        onClose={() => setDeleting(null)}
        onConfirm={handleDelete}
        title="删除字段"
        description={`确认删除字段「${deleting?.field_label ?? ''}」？删除后不可恢复。`}
        confirmText="删除"
        variant="destructive"
      />
    </div>
  );
}
