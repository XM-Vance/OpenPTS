'use client';

import { useEffect, useRef, useState, useCallback } from 'react';

interface NumberFlipProps {
  /** 目标数值 */
  value: number;
  /** 小数位数 */
  decimals?: number;
  /** 是否使用千分位 */
  useGrouping?: boolean;
  /** 字体大小 (px) */
  fontSize?: number;
  /** 字体颜色 */
  color?: string;
  /** 额外 CSS */
  style?: React.CSSProperties;
  /** 前缀 */
  prefix?: string;
  /** 后缀 */
  suffix?: string;
}

/**
 * 数字翻牌器：当 value 变化时，通过 CSS transition 实现数字滚动动画。
 * 实现原理：每位数字用独立列，通过 translateY 动画滚动到目标数字。
 */
export function NumberFlip({
  value,
  decimals = 0,
  useGrouping = true,
  fontSize = 24,
  color = '#fff',
  style,
  prefix = '',
  suffix = '',
}: NumberFlipProps) {
  const [displayValue, setDisplayValue] = useState(value);
  const animationRef = useRef<number | null>(null);
  const startValueRef = useRef(value);
  const startTimeRef = useRef(0);
  const duration = 800; // 动画时长 ms

  const animate = useCallback(
    (timestamp: number) => {
      if (!startTimeRef.current) startTimeRef.current = timestamp;
      const progress = Math.min(
        (timestamp - startTimeRef.current) / duration,
        1,
      );
      // easeOutExpo
      const eased = progress === 1 ? 1 : 1 - Math.pow(2, -10 * progress);
      const current =
        startValueRef.current +
        (value - startValueRef.current) * eased;
      setDisplayValue(current);

      if (progress < 1) {
        animationRef.current = requestAnimationFrame(animate);
      }
    },
    [value],
  );

  useEffect(() => {
    startValueRef.current = displayValue;
    startTimeRef.current = 0;
    if (animationRef.current) cancelAnimationFrame(animationRef.current);
    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value, animate]);

  const formatted = displayValue.toFixed(decimals);
  const parts = useGrouping
    ? Number(formatted).toLocaleString('zh-CN', {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals,
      })
    : formatted;

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'baseline',
        fontSize,
        fontWeight: 700,
        color,
        fontFamily: '"DIN Alternate", "Roboto Mono", monospace',
        transition: 'color 0.3s ease',
        ...style,
      }}
    >
      {prefix}
      {parts}
      {suffix}
    </span>
  );
}
