'use client';

/**
 * 全局 Error Boundary — Next.js App Router 约定。
 * 捕获 (main) 路由组下所有页面运行时错误。
 * 必须是客户端组件。
 */
export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex items-center justify-center py-32">
      <div className="flex flex-col items-center gap-4 max-w-md text-center">
        <div className="text-4xl">⚠️</div>
        <h2 className="text-lg font-semibold">页面加载出错</h2>
        <p className="text-sm text-muted-foreground">
          {error.message || '发生了未知错误，请稍后重试。'}
        </p>
        {error.digest && (
          <p className="text-xs text-muted-foreground/60">
            错误编号: {error.digest}
          </p>
        )}
        <button
          onClick={reset}
          className="rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          重试
        </button>
      </div>
    </div>
  );
}
