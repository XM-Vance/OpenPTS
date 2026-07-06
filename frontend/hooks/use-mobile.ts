'use client';

import { useEffect, useState } from 'react';

/**
 * 判断当前视口是否窄于给定断点（默认 768px = Tailwind md）。
 *
 * SSR 安全：首帧恒返回 false（与桌面服务端渲染一致），挂载后按真实视口修正。
 * 这样可避免水合不匹配警告；CSS 侧的 md: 断点不依赖本 hook，首屏即正确。
 *
 * 用于决定是否渲染移动专属 UI（汉堡菜单、卡片列表、Sheet 抽屉）。
 */
export function useIsMobile(breakpoint = 768): boolean {
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return;
    const mql = window.matchMedia(`(max-width: ${breakpoint - 1}px)`);
    const onChange = () => setIsMobile(mql.matches);
    onChange(); // 挂载即同步真实值
    mql.addEventListener('change', onChange);
    return () => mql.removeEventListener('change', onChange);
  }, [breakpoint]);

  return isMobile;
}
