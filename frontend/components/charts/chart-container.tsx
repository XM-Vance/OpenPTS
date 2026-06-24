'use client';

import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { Maximize2, Minimize2, Download } from 'lucide-react';

export interface ChartContainerProps {
  /** 图表标题 */
  title: string;
  /** 右上角额外操作按钮 */
  actions?: React.ReactNode;
  /** 容器 className */
  className?: string;
  /** 图表内容（通常是 Recharts 组件） */
  children: React.ReactNode;
  /** 图表最小高度 */
  minHeight?: number;
}

/**
 * 图表容器组件
 * - 带标题栏、全屏切换、导出按钮
 * - 包裹 Recharts 图表
 * - Recharts 焦点外框去除
 */
export function ChartContainer({
  title,
  actions,
  className,
  children,
  minHeight = 300,
}: ChartContainerProps) {
  const [isFullscreen, setIsFullscreen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // 全屏切换
  const toggleFullscreen = useCallback(() => {
    setIsFullscreen((prev) => !prev);
  }, []);

  // ESC 退出全屏
  useEffect(() => {
    if (!isFullscreen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setIsFullscreen(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isFullscreen]);

  // 导出图表为 PNG
  const handleExport = useCallback(() => {
    const svgEl = containerRef.current?.querySelector('.recharts-surface');
    if (!svgEl) return;

    const svgData = new XMLSerializer().serializeToString(svgEl);
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const svgRect = svgEl.getBoundingClientRect();
    const scale = 2;
    canvas.width = svgRect.width * scale;
    canvas.height = svgRect.height * scale;
    ctx.scale(scale, scale);

    const img = new Image();
    img.onload = () => {
      ctx.fillStyle = '#ffffff';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      ctx.drawImage(img, 0, 0, svgRect.width, svgRect.height);
      const link = document.createElement('a');
      link.download = `${title}_${new Date().toISOString().slice(0, 10)}.png`;
      link.href = canvas.toDataURL('image/png');
      link.click();
    };
    img.src = 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(svgData)));
  }, [title]);

  if (isFullscreen) {
    return (
      <div className="fixed inset-0 z-50 flex flex-col bg-background">
        {/* 全屏顶栏 */}
        <div className="flex items-center justify-between border-b px-4 py-3">
          <h2 className="text-lg font-semibold">{title}</h2>
          <div className="flex items-center gap-2">
            {actions}
            <Button variant="ghost" size="icon" onClick={handleExport} title="导出图片">
              <Download className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" onClick={toggleFullscreen} title="退出全屏">
              <Minimize2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
        {/* 全屏图表区域 */}
        <div
          ref={containerRef}
          className="flex-1 p-4 [&_.recharts-surface:focus]:outline-none [&_*:focus]:outline-none"
        >
          {children}
        </div>
      </div>
    );
  }

  return (
    <Card className={cn('relative', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-medium">{title}</CardTitle>
        <div className="flex items-center gap-1">
          {actions}
          <Button variant="ghost" size="icon" onClick={handleExport} title="导出图片">
            <Download className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" onClick={toggleFullscreen} title="全屏">
            <Maximize2 className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent
        ref={containerRef}
        style={{ minHeight }}
        className="[&_.recharts-surface:focus]:outline-none [&_*:focus]:outline-none"
      >
        {children}
      </CardContent>
    </Card>
  );
}
