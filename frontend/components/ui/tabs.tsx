'use client';

// 简易 Tabs 组件（纯 React state + Tailwind，不依赖 radix）。
// 用于 MCP 接入说明页按客户端切换配置示例。API 与 shadcn tabs 一致，方便未来替换。
import { createContext, useContext, useState, type ReactNode } from 'react';

interface TabsCtx {
  value: string;
  setValue: (v: string) => void;
}
const Ctx = createContext<TabsCtx | null>(null);

export function Tabs({ defaultValue, children, className }: {
  defaultValue: string;
  children: ReactNode;
  className?: string;
}) {
  const [value, setValue] = useState(defaultValue);
  return (
    <Ctx.Provider value={{ value, setValue }}>
      <div className={className}>{children}</div>
    </Ctx.Provider>
  );
}

export function TabsList({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={`inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground ${className ?? ''}`}>
      {children}
    </div>
  );
}

export function TabsTrigger({ value, children, className }: {
  value: string;
  children: ReactNode;
  className?: string;
}) {
  const ctx = useContext(Ctx);
  if (!ctx) throw new Error('TabsTrigger must be inside <Tabs>');
  const active = ctx.value === value;
  return (
    <button
      type="button"
      onClick={() => ctx.setValue(value)}
      className={`inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all ${
        active ? 'bg-background text-foreground shadow' : 'hover:text-foreground'
      } ${className ?? ''}`}
    >
      {children}
    </button>
  );
}

export function TabsContent({ value, children, className }: {
  value: string;
  children: ReactNode;
  className?: string;
}) {
  const ctx = useContext(Ctx);
  if (!ctx) throw new Error('TabsContent must be inside <Tabs>');
  if (ctx.value !== value) return null;
  return <div className={`mt-2 ${className ?? ''}`}>{children}</div>;
}
