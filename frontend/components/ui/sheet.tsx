'use client';

import * as React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

/**
 * 轻量侧滑抽屉：不依赖 @radix-ui/react-dialog / vaul，与现有 dialog.tsx 风格一致。
 * 用于移动端导航抽屉（钉钉 H5）。遮罩点击 + ESC 关闭 + 滑入动画（tailwindcss-animate）。
 *
 * 桌面端不会渲染本组件（调用方按 useIsMobile() 决定），故无需做桌面适配。
 */
interface SheetProps {
  open: boolean;
  onClose: () => void;
  side?: 'left' | 'right';
  /** 顶部标题栏（可选） */
  title?: React.ReactNode;
  children: React.ReactNode;
  /** 内容区宽度 */
  widthClass?: string;
}

export function Sheet({
  open,
  onClose,
  side = 'left',
  title,
  children,
  widthClass = 'w-72',
}: SheetProps) {
  React.useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  if (!open) return null;

  const isLeft = side === 'left';

  return (
    <div className="fixed inset-0 z-50">
      {/* 遮罩：fade-in */}
      <div
        className="absolute inset-0 bg-black/50 animate-in fade-in"
        onClick={onClose}
        aria-hidden
      />
      {/* 抽屉面板：从侧边 slide-in */}
      <div
        role="dialog"
        aria-modal="true"
        className={cn(
          'absolute top-0 flex h-full flex-col border-card bg-card shadow-xl',
          widthClass,
          isLeft ? 'left-0 border-r animate-in slide-in-from-left' : 'right-0 border-l animate-in slide-in-from-right',
        )}
      >
        {title !== undefined && (
          <div className="flex h-14 shrink-0 items-center justify-between border-b px-4">
            <span className="text-sm font-bold">{title}</span>
            <button
              type="button"
              onClick={onClose}
              className="text-muted-foreground hover:text-foreground"
              aria-label="关闭"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )}
        <div className="flex-1 overflow-y-auto">{children}</div>
      </div>
    </div>
  );
}
