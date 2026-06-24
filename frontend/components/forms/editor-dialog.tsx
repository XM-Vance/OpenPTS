'use client';

import React, { useCallback } from 'react';
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export interface EditorDialogProps {
  /** 是否显示 */
  open: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 保存回调 */
  onSave: () => void;
  /** 标题 */
  title: string;
  /** 模式 */
  mode?: 'create' | 'edit';
  /** 宽度 */
  width?: string;
  /** 是否正在保存 */
  saving?: boolean;
  /** 错误信息 */
  error?: string | null;
  /** 禁用保存 */
  disableSave?: boolean;
  /** 表单内容 */
  children: React.ReactNode;
  /** 额外底部按钮 */
  footerExtra?: React.ReactNode;
  /** 容器 className */
  className?: string;
}

/**
 * 通用编辑弹窗
 * - create / edit 模式
 * - 保存 / 取消 按钮
 * - 错误提示
 */
export function EditorDialog({
  open,
  onClose,
  onSave,
  title,
  mode = 'create',
  width = 'max-w-lg',
  saving = false,
  error,
  disableSave = false,
  children,
  footerExtra,
  className,
}: EditorDialogProps) {
  const handleBackdropClick = useCallback(
    (e: React.MouseEvent) => {
      // 阻止遮罩层点击关闭（编辑中防误操作）
      e.stopPropagation();
    },
    [],
  );

  return (
    <Dialog
      open={open}
      onClose={onClose}
      className={cn(width, 'max-h-[85vh]', className)}
    >
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        {mode === 'edit' && (
          <DialogDescription>修改后点击保存生效</DialogDescription>
        )}
      </DialogHeader>

      {/* 错误提示 */}
      {error && (
        <div className="mb-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* 表单内容 */}
      <div className="space-y-4 py-2">{children}</div>

      <DialogFooter>
        {footerExtra}
        <Button variant="outline" onClick={onClose} disabled={saving}>
          取消
        </Button>
        <Button onClick={onSave} disabled={saving || disableSave}>
          {saving ? '保存中...' : '保存'}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
