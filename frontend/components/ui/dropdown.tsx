'use client';

import * as React from 'react';
import { MoreHorizontal } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

/**
 * 轻量「⋯」下拉菜单：不依赖 @radix-ui，与现有 dialog.tsx/sheet.tsx 风格一致。
 * 主要用于移动端卡片底部的次要操作收纳（附件/PDF/删除等）。
 * 点触发器展开、点外部或 ESC 关闭。
 */
export interface DropdownItem {
  label: string;
  onClick: () => void;
  icon?: React.ReactNode;
  danger?: boolean;
  disabled?: boolean;
}

interface DropdownProps {
  items: DropdownItem[];
  /** 对齐：right（默认，贴近右边缘） / left */
  align?: 'left' | 'right';
  /** 触发器文案/图标，默认「⋯」图标按钮 */
  trigger?: React.ReactNode;
  triggerLabel?: string;
}

export function Dropdown({ items, align = 'right', trigger, triggerLabel }: DropdownProps) {
  const [open, setOpen] = React.useState(false);
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  const enabled = items.filter((it) => !it.disabled);

  return (
    <div className="relative" ref={ref}>
      {trigger !== undefined ? (
        <span onClick={() => setOpen((v) => !v)}>{trigger}</span>
      ) : (
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0"
          onClick={() => setOpen((v) => !v)}
          aria-label={triggerLabel ?? '更多操作'}
          disabled={enabled.length === 0}
        >
          {triggerLabel ?? <MoreHorizontal className="h-4 w-4" />}
        </Button>
      )}
      {open && (
        <div
          className={cn(
            'absolute z-50 mt-1 min-w-[8rem] overflow-hidden rounded-md border bg-card p-1 shadow-md',
            align === 'right' ? 'right-0' : 'left-0',
          )}
        >
          {enabled.map((it, i) => (
            <button
              key={i}
              type="button"
              onClick={() => {
                setOpen(false);
                it.onClick();
              }}
              className={cn(
                'flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent',
                it.danger && 'text-destructive hover:bg-destructive/10',
              )}
            >
              {it.icon && <span className="shrink-0">{it.icon}</span>}
              {it.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
