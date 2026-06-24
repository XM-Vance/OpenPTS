'use client';

import React from 'react';
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

export interface ConfirmDialogProps {
  /** 是否显示 */
  open: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 确认回调 */
  onConfirm: () => void;
  /** 标题 */
  title?: string;
  /** 描述 */
  description?: string;
  /** 确认按钮文字 */
  confirmText?: string;
  /** 取消按钮文字 */
  cancelText?: string;
  /** 确认按钮变体 */
  variant?: 'default' | 'destructive';
  /** 是否加载中 */
  loading?: boolean;
}

/**
 * 确认弹窗
 * - 通用确认/取消操作
 * - 支持危险操作变体
 */
export function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title = '确认操作',
  description = '确定要执行此操作吗？',
  confirmText = '确认',
  cancelText = '取消',
  variant = 'default',
  loading = false,
}: ConfirmDialogProps) {
  return (
    <Dialog open={open} onClose={onClose} className="max-w-md">
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>{description}</DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button variant="outline" onClick={onClose} disabled={loading}>
          {cancelText}
        </Button>
        <Button
          variant={variant === 'destructive' ? 'destructive' : 'default'}
          onClick={onConfirm}
          disabled={loading}
        >
          {loading ? '处理中...' : confirmText}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
