'use client';

import { useEffect, useState, useRef, useCallback } from 'react';
import {
  getDisplayOverview,
  getDisplayTrend,
  type DisplayOverview,
  type DisplayTrendItem,
} from '@/lib/api/display-screen';
import {
  BASE_WIDTH,
  BASE_HEIGHT,
  POLL_INTERVAL,
  FALLBACK_OVERVIEW,
  C,
  GridBg,
  HeaderBar,
  LeftColumn,
  CenterColumn,
  RightColumn,
  BottomBar,
} from '@/components/display-screen';

/* ========== 主页面 ========== */
export default function DisplayScreenPage() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(1);
  const [currentTime, setCurrentTime] = useState(new Date());
  const [overview, setOverview] = useState<DisplayOverview>(FALLBACK_OVERVIEW);
  const [trend, setTrend] = useState<DisplayTrendItem[]>([]);
  const [tradeStats] = useState([
    { name: '中长期', value: 58.3, color: C.cyan },
    { name: '现货', value: 22.7, color: C.gold },
    { name: '绿电', value: 10.5, color: C.green },
    { name: '调频', value: 8.5, color: C.pink },
  ]);

  /* 自适应缩放 */
  const updateScale = useCallback(() => {
    if (!containerRef.current) return;
    const parent = containerRef.current.parentElement;
    if (!parent) return;
    const wScale = parent.clientWidth / BASE_WIDTH;
    const hScale = parent.clientHeight / BASE_HEIGHT;
    setScale(Math.min(wScale, hScale, 1));
  }, []);

  useEffect(() => {
    updateScale();
    window.addEventListener('resize', updateScale);
    return () => window.removeEventListener('resize', updateScale);
  }, [updateScale]);

  /* 实时时钟 */
  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  /* 轮询数据 */
  const fetchData = useCallback(async () => {
    try {
      const [ov, tr] = await Promise.all([
        getDisplayOverview().catch(() => null),
        getDisplayTrend().catch(() => null),
      ]);
      if (ov) setOverview(ov);
      if (tr?.items) setTrend(tr.items);
    } catch {
      // 静默失败，保持已有数据
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, POLL_INTERVAL);
    return () => clearInterval(interval);
  }, [fetchData]);

  const timeStr = currentTime.toLocaleTimeString('zh-CN', { hour12: false });
  const dateStr = currentTime.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });

  return (
    <div
      id="display-screen-root"
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100vw',
        height: '100vh',
        zIndex: 99999,
        background: C.bg,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden',
      }}
    >
      <div
        ref={containerRef}
        style={{
          width: BASE_WIDTH,
          height: BASE_HEIGHT,
          transform: `scale(${scale})`,
          transformOrigin: 'top left',
          background: C.bgGradient,
          position: 'relative',
          overflow: 'hidden',
          fontFamily: '"PingFang SC", "Microsoft YaHei", "Helvetica Neue", sans-serif',
        }}
      >
        <GridBg />

        {/* ====== 顶部标题栏 ====== */}
        <HeaderBar timeStr={timeStr} dateStr={dateStr} overview={overview} />

        {/* ====== 主体内容 ====== */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1.2fr 1.8fr 1.2fr',
            gridTemplateRows: 'auto 1fr 230px',
            gap: 12,
            padding: '12px 16px',
            height: BASE_HEIGHT - 70,
          }}
        >
          {/* ──────── 左列 ──────── */}
          <LeftColumn overview={overview} />

          {/* ──────── 中间列 ──────── */}
          <CenterColumn trend={trend} />

          {/* ──────── 右列 ──────── */}
          <RightColumn overview={overview} tradeStats={tradeStats} />

          {/* ──────── 底部：营收柱状图 ──────── */}
          <BottomBar trend={trend} />
        </div>
      </div>
    </div>
  );
}
