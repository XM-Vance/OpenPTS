'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  EditorDialog,
} from '@/components/forms/editor-dialog';
import { ConfirmDialog } from '@/components/feedback/confirm-dialog';
import {
  listTags,
  createTag,
  updateTag,
  deleteTag,
  type Tag,
  type TagInput,
} from '@/lib/api/tags';
import { Tags, Plus, Pencil, Trash2 } from 'lucide-react';

const ENTITY_TYPES = [
  { value: '', label: '全部实体' },
  { value: 'customer', label: '客户' },
  { value: 'contract', label: '合同' },
  { value: 'document', label: '文档' },
  { value: 'intent_customer', label: '意向客户' },
];

const PRESET_COLORS = [
  '#ef4444', '#f97316', '#f59e0b', '#84cc16', '#22c55e',
  '#14b8a6', '#06b6d4', '#3b82f6', '#6366f1', '#8b5cf6',
  '#a855f7', '#d946ef', '#ec4899', '#64748b', '#374151',
];

const SELECT_CLASS = 'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';

/* ── Tag Editor Dialog ── */
function TagEditorDialog({
  open,
  mode,
  tag,
  onClose,
  onSave,
}: {
  open: boolean;
  mode: 'create' | 'edit';
  tag?: Tag | null;
  onClose: () => void;
  onSave: () => void;
}) {
  const [formData, setFormData] = useState<TagInput>({
    name: '',
    color: PRESET_COLORS[0],
    entity_type: 'customer',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useMemo(() => {
    if (open) {
      if (tag && mode === 'edit') {
        setFormData({
          name: tag.name,
          color: tag.color,
          entity_type: tag.entity_type,
        });
      } else {
        setFormData({
          name: '',
          color: PRESET_COLORS[0],
          entity_type: 'customer',
        });
      }
      setError(null);
    }
  }, [open, tag, mode]);

  const handleSave = async () => {
    if (!formData.name.trim()) {
      setError('请输入标签名称');
      return;
    }
    setSaving(true);
    setError(null);
    try {
      if (mode === 'create') {
        await createTag(formData);
      } else if (tag) {
        await updateTag(tag.id, formData);
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

  const handleChange = (key: keyof TagInput, value: string) => {
    setFormData((prev) => ({ ...prev, [key]: value }));
  };

  return (
    <EditorDialog
      open={open}
      onClose={onClose}
      onSave={handleSave}
      title={mode === 'create' ? '新增标签' : `编辑标签: ${tag?.name ?? ''}`}
      mode={mode}
      width="max-w-md"
      saving={saving}
      error={error}
    >
      <div className="space-y-4">
        <div className="space-y-1.5">
          <Label htmlFor="tag_name">标签名称 *</Label>
          <Input
            id="tag_name"
            value={formData.name}
            onChange={(e) => handleChange('name', e.target.value)}
            placeholder="如：VIP客户"
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="tag_entity">实体类型 *</Label>
          <select
            id="tag_entity"
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
          <Label>颜色</Label>
          <div className="flex flex-wrap gap-2">
            {PRESET_COLORS.map((c) => (
              <button
                key={c}
                type="button"
                className={`h-8 w-8 rounded-full border-2 transition-transform hover:scale-110 ${
                  formData.color === c ? 'border-gray-900 scale-110' : 'border-transparent'
                }`}
                style={{ backgroundColor: c }}
                onClick={() => handleChange('color', c)}
              />
            ))}
          </div>
          <div className="flex items-center gap-2 mt-2">
            <Label htmlFor="tag_color_custom" className="text-xs whitespace-nowrap">自定义:</Label>
            <input
              id="tag_color_custom"
              type="color"
              value={formData.color}
              onChange={(e) => handleChange('color', e.target.value)}
              className="h-8 w-8 cursor-pointer rounded border"
            />
            <Input
              value={formData.color}
              onChange={(e) => handleChange('color', e.target.value)}
              className="w-28 font-mono text-sm"
              placeholder="#000000"
            />
          </div>
        </div>
        {/* Preview */}
        <div className="space-y-1.5">
          <Label className="text-xs text-muted-foreground">预览</Label>
          <div className="flex items-center gap-2">
            <span
              className="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium text-white"
              style={{ backgroundColor: formData.color }}
            >
              {formData.name || '标签预览'}
            </span>
          </div>
        </div>
      </div>
    </EditorDialog>
  );
}

/* ── Main Page ── */
export default function TagsPage() {
  const qc = useQueryClient();
  const [entityFilter, setEntityFilter] = useState('');
  const [editing, setEditing] = useState<Tag | 'new' | null>(null);
  const [deleting, setDeleting] = useState<Tag | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['tags', entityFilter],
    queryFn: () => listTags(entityFilter || undefined),
  });

  const items: Tag[] = useMemo(() => data?.items ?? [], [data]);

  const ENTITY_LABELS: Record<string, string> = {
    customer: '客户',
    contract: '合同',
    document: '文档',
    intent_customer: '意向客户',
  };

  const handleDelete = async () => {
    if (!deleting) return;
    try {
      await deleteTag(deleting.id);
      qc.invalidateQueries({ queryKey: ['tags'] });
      setDeleting(null);
    } catch (err) {
      window.alert(err instanceof Error ? err.message : '删除失败');
    }
  };

  // Group tags by entity_type for display
  const grouped = useMemo(() => {
    const map: Record<string, Tag[]> = {};
    for (const t of items) {
      const key = t.entity_type;
      if (!map[key]) map[key] = [];
      map[key].push(t);
    }
    return map;
  }, [items]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold">
            <Tags className="h-6 w-6" />
            标签管理
          </h1>
          <p className="text-sm text-muted-foreground">
            管理各实体的标签，支持自定义颜色与分类
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
            新增标签
          </Button>
        </div>
      </div>

      {isLoading && (
        <div className="py-12 text-center text-sm text-muted-foreground">加载中...</div>
      )}

      {!isLoading && items.length === 0 && (
        <div className="py-12 text-center text-sm text-muted-foreground">
          暂无标签，点击「新增标签」添加
        </div>
      )}

      {!isLoading &&
        Object.entries(grouped).map(([entityType, tags]) => (
          <div key={entityType} className="space-y-2">
            <h3 className="text-sm font-medium text-muted-foreground">
              {ENTITY_LABELS[entityType] ?? entityType}（{tags.length}）
            </h3>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {tags.map((t) => (
                <div
                  key={t.id}
                  className="flex items-center justify-between rounded-lg border bg-card p-3"
                >
                  <div className="flex items-center gap-2">
                    <span
                      className="h-5 w-5 rounded-full shrink-0"
                      style={{ backgroundColor: t.color }}
                    />
                    <span className="font-medium text-sm">{t.name}</span>
                    <span
                      className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium text-white"
                      style={{ backgroundColor: t.color }}
                    >
                      {ENTITY_LABELS[t.entity_type] ?? t.entity_type}
                    </span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Button variant="ghost" size="sm" onClick={() => setEditing(t)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => setDeleting(t)}>
                      <Trash2 className="h-4 w-4 text-red-500" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}

      {/* Editor Dialog */}
      <TagEditorDialog
        open={editing !== null}
        mode={editing === 'new' ? 'create' : 'edit'}
        tag={editing !== null && editing !== 'new' ? editing : null}
        onClose={() => setEditing(null)}
        onSave={() => qc.invalidateQueries({ queryKey: ['tags'] })}
      />

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleting !== null}
        onClose={() => setDeleting(null)}
        onConfirm={handleDelete}
        title="删除标签"
        description={`确认删除标签「${deleting?.name ?? ''}」？删除后不可恢复。`}
        confirmText="删除"
        variant="destructive"
      />
    </div>
  );
}
