'use client';

// 订阅 backend /api/v1/stream/alerts SSE 流；收到 alert 事件时弹 toast。
// 凭证走 httpOnly 登录 Cookie:同源 EventSource 自动携带,URL 不再拼 token(防泄露进日志/历史)。
import { useEffect, useState } from 'react';
import { useAuth } from '@/lib/auth/context';

interface ToastItem {
  id: number;
  type: string;
  message: string;
  ts: number;
}

let counter = 0;

export function AlertStream() {
  const { user } = useAuth();
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  useEffect(() => {
    if (!user) return;

    const es = new EventSource('/api/v1/stream/alerts');

    const pushToast = (type: string, defaultMsg: string) => (e: Event) => {
      try {
        const data = JSON.parse((e as MessageEvent).data);
        const item: ToastItem = {
          id: ++counter,
          type,
          message: String(data.message ?? defaultMsg),
          ts: Number(data.ts ?? Date.now() / 1000),
        };
        setToasts((prev) => [...prev, item]);
        setTimeout(() => {
          setToasts((prev) => prev.filter((t) => t.id !== item.id));
        }, 6000);
      } catch {
        /* ignore */
      }
    };

    es.addEventListener('alert', pushToast('alert', '收到新告警'));
    es.addEventListener('approval', pushToast('approval', '审批流通知'));
    es.addEventListener('job', pushToast('job', '调度任务完成'));

    es.addEventListener('hello', () => {
      // 连接成功，调试时可在 console 看到
      if (process.env.NODE_ENV !== 'production') {
        console.info('[SSE] connected');
      }
    });

    es.onerror = () => {
      // EventSource 会自动重连，无需手工处理；可在此打日志
    };

    return () => es.close();
  }, [user]);

  if (toasts.length === 0) return null;

  return (
    <div className="pointer-events-none fixed right-4 top-16 z-50 flex w-80 flex-col gap-2">
      {toasts.map((t) => {
        const styles =
          t.type === 'approval'
            ? { colors: 'border-blue-300 bg-blue-50 text-blue-900 dark:border-blue-700 dark:bg-blue-950 dark:text-blue-100', icon: '✅', label: '审批通知' }
            : t.type === 'job'
              ? { colors: 'border-emerald-300 bg-emerald-50 text-emerald-900 dark:border-emerald-700 dark:bg-emerald-950 dark:text-emerald-100', icon: '⏰', label: '调度任务' }
              : { colors: 'border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-100', icon: '🔔', label: '实时告警' };
        const { colors, icon, label } = styles;
        return (
          <div
            key={t.id}
            className={`pointer-events-auto rounded-md border p-3 text-sm shadow-lg ${colors}`}
          >
            <p className="font-medium">{icon} {label}</p>
            <p className="mt-1 text-xs">{t.message}</p>
            <p className="mt-1 text-[10px] opacity-60">
              {new Date(t.ts * 1000).toLocaleTimeString('zh-CN')}
            </p>
          </div>
        );
      })}
    </div>
  );
}
