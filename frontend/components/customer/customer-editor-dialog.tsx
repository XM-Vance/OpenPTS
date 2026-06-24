'use client';

import React, { useEffect, useState } from 'react';
import {
  EditorDialog,
  type EditorDialogProps,
} from '@/components/forms/editor-dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { TagSelector, type Tag } from './tag-selector';
import type { Customer, CustomerInput } from '@/lib/api/customers';
import { createCustomer, updateCustomer } from '@/lib/api/customers';

export interface CustomerEditorDialogProps {
  open: boolean;
  mode: 'create' | 'edit';
  customer?: Customer | null;
  onClose: () => void;
  onSave: () => void;
}

/**
 * 客户编辑弹窗
 * - 新增 / 编辑模式
 * - 基础信息 + 标签
 * - 调用 customers API
 */
export function CustomerEditorDialog({
  open,
  mode,
  customer,
  onClose,
  onSave,
}: CustomerEditorDialogProps) {
  const [formData, setFormData] = useState<CustomerInput>({
    user_name: '',
    short_name: '',
    location: '',
    source: '',
    manager: '',
    tags: [],
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 初始化表单
  useEffect(() => {
    if (open && customer && mode === 'edit') {
      setFormData({
        user_name: customer.user_name,
        short_name: customer.short_name ?? '',
        location: customer.location ?? '',
        source: customer.source ?? '',
        manager: customer.manager ?? '',
        tags: customer.tags ?? [],
      });
    } else if (open && mode === 'create') {
      setFormData({
        user_name: '',
        short_name: '',
        location: '',
        source: '',
        manager: '',
        tags: [],
      });
    }
    setError(null);
  }, [open, customer, mode]);

  const handleChange = (field: keyof CustomerInput, value: string | string[]) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleTagsChange = (tags: Tag[]) => {
    setFormData((prev) => ({ ...prev, tags: tags.map((t) => t.name) }));
  };

  const handleSave = async () => {
    if (!formData.user_name?.trim()) {
      setError('请输入客户全称');
      return;
    }
    if (!formData.short_name?.trim()) {
      setError('请输入客户简称');
      return;
    }

    setSaving(true);
    setError(null);
    try {
      if (mode === 'create') {
        await createCustomer(formData);
      } else if (customer) {
        await updateCustomer(customer.id, formData);
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

  return (
    <EditorDialog
      open={open}
      onClose={onClose}
      onSave={handleSave}
      title={mode === 'create' ? '新增客户' : `编辑客户: ${customer?.user_name ?? ''}`}
      mode={mode}
      width="max-w-md"
      saving={saving}
      error={error}
    >
      <div className="space-y-4">
        {/* 客户全称 */}
        <div className="space-y-1.5">
          <Label htmlFor="user_name">客户全称 *</Label>
          <Input
            id="user_name"
            value={formData.user_name}
            onChange={(e) => handleChange('user_name', e.target.value)}
            placeholder="请输入客户全称"
          />
        </div>

        {/* 客户简称 */}
        <div className="space-y-1.5">
          <Label htmlFor="short_name">客户简称 *</Label>
          <Input
            id="short_name"
            value={formData.short_name ?? ''}
            onChange={(e) => handleChange('short_name', e.target.value)}
            placeholder="请输入客户简称"
          />
        </div>

        {/* 位置 */}
        <div className="space-y-1.5">
          <Label htmlFor="location">位置</Label>
          <Input
            id="location"
            value={formData.location ?? ''}
            onChange={(e) => handleChange('location', e.target.value)}
            placeholder="请输入位置"
          />
        </div>

        {/* 来源 */}
        <div className="space-y-1.5">
          <Label htmlFor="source">客户来源</Label>
          <Input
            id="source"
            value={formData.source ?? ''}
            onChange={(e) => handleChange('source', e.target.value)}
            placeholder="请输入客户来源"
          />
        </div>

        {/* 客户经理 */}
        <div className="space-y-1.5">
          <Label htmlFor="manager">客户经理</Label>
          <Input
            id="manager"
            value={formData.manager ?? ''}
            onChange={(e) => handleChange('manager', e.target.value)}
            placeholder="请输入客户经理"
          />
        </div>

        {/* 标签 */}
        <div className="space-y-1.5">
          <Label>标签</Label>
          <TagSelector
            tags={(formData.tags ?? []).map((name) => ({
              name,
              source: 'MANUAL',
            }))}
            onChange={handleTagsChange}
          />
        </div>
      </div>
    </EditorDialog>
  );
}
