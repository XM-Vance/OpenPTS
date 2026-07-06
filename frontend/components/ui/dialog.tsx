'use client';

import * as React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DialogProps {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
  className?: string;
  /**
   * 移动端全屏：<768px 时撑满视口（inset-0 / max-h-none / 圆角去除），
   * 适合内容较多的表单/审批 Dialog。桌面端（≥768px）保持 max-w-lg 居中不变。
   */
  fullScreenOnMobile?: boolean;
}

// 轻量 Modal：不依赖 @radix-ui/react-dialog。点遮罩或按 ESC 关闭。
export function Dialog({ open, onClose, children, className, fullScreenOnMobile }: DialogProps) {
  React.useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center p-4',
        // 移动端全屏：容器去 padding、顶端对齐；桌面端恢复居中
        fullScreenOnMobile && 'items-stretch justify-stretch p-0 sm:items-center sm:justify-center sm:p-4',
      )}
    >
      <div className="absolute inset-0 bg-black/50" onClick={onClose} aria-hidden />
      <div
        role="dialog"
        aria-modal="true"
        className={cn(
          'relative z-10 max-h-[85vh] w-full max-w-lg overflow-y-auto rounded-lg border bg-card p-6 shadow-lg',
          // 移动端全屏（mobile-first）：撑满视口、去圆角；≥640px 恢复 max-w-lg 居中卡片
          fullScreenOnMobile &&
            'max-h-none max-w-none rounded-none p-4 sm:max-h-[85vh] sm:max-w-lg sm:rounded-lg sm:p-6',
          className,
        )}
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute right-4 top-4 text-muted-foreground hover:text-foreground"
          aria-label="关闭"
        >
          <X className="h-4 w-4" />
        </button>
        {children}
      </div>
    </div>
  );
}

export function DialogHeader({ children }: { children: React.ReactNode }) {
  return <div className="mb-4 pr-8">{children}</div>;
}

export function DialogTitle({ children }: { children: React.ReactNode }) {
  return <h2 className="text-lg font-semibold">{children}</h2>;
}

export function DialogDescription({ children }: { children: React.ReactNode }) {
  return <p className="mt-1 text-sm text-muted-foreground">{children}</p>;
}

export function DialogFooter({ children }: { children: React.ReactNode }) {
  return <div className="mt-6 flex justify-end gap-2">{children}</div>;
}
