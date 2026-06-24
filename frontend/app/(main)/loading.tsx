/**
 * 全局 Loading 状态 — Next.js App Router 约定。
 * 在 (main) 路由组下所有页面切换时自动展示。
 * 各子路由也可以覆盖 loading.tsx 提供自定义骨架屏。
 */
export default function Loading() {
  return (
    <div className="flex items-center justify-center py-32">
      <div className="flex flex-col items-center gap-3">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-muted-foreground/20 border-t-primary" />
        <p className="text-sm text-muted-foreground">加载中…</p>
      </div>
    </div>
  );
}
