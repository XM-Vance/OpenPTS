// Locale 感知的格式化工具：金额（万 ↔ wan/× 10k）、日期、相对时间。
// 用法：const { fmtMoney, fmtDate } = useFormat();
import { useCallback } from 'react';
import { useI18n } from './context';
import type { Locale } from './messages';

function formatMoney(v: number, locale: Locale): string {
  // 中文以「万」为单位；英文以 k/M 为单位。
  if (locale === 'zh-CN') {
    if (Math.abs(v) >= 1e8) {
      return `${(v / 1e8).toLocaleString('zh-CN', { maximumFractionDigits: 2 })} 亿`;
    }
    if (Math.abs(v) >= 1e4) {
      return `${(v / 1e4).toLocaleString('zh-CN', { maximumFractionDigits: 2 })} 万`;
    }
    return v.toLocaleString('zh-CN', { maximumFractionDigits: 2 });
  }
  // en-US
  if (Math.abs(v) >= 1e9) {
    return `${(v / 1e9).toLocaleString('en-US', { maximumFractionDigits: 2 })}B`;
  }
  if (Math.abs(v) >= 1e6) {
    return `${(v / 1e6).toLocaleString('en-US', { maximumFractionDigits: 2 })}M`;
  }
  if (Math.abs(v) >= 1e3) {
    return `${(v / 1e3).toLocaleString('en-US', { maximumFractionDigits: 2 })}k`;
  }
  return v.toLocaleString('en-US', { maximumFractionDigits: 2 });
}

function formatNumber(v: number, locale: Locale, digits = 2): string {
  return v.toLocaleString(locale, { maximumFractionDigits: digits });
}

function formatPercent(v: number, locale: Locale, digits = 1): string {
  return `${v.toLocaleString(locale, { maximumFractionDigits: digits })}%`;
}

function formatDate(s: string, locale: Locale): string {
  if (!s) return '-';
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  if (locale === 'zh-CN') {
    const p = (n: number) => (n < 10 ? `0${n}` : String(n));
    return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
  }
  return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

function formatDateTime(s: string, locale: Locale): string {
  if (!s) return '-';
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  if (locale === 'zh-CN') {
    const p = (n: number) => (n < 10 ? `0${n}` : String(n));
    return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
  }
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatRelative(s: string, locale: Locale): string {
  if (!s) return '-';
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  const diffMs = Date.now() - d.getTime();
  const sec = Math.floor(diffMs / 1000);
  const min = Math.floor(sec / 60);
  const hour = Math.floor(min / 60);
  const day = Math.floor(hour / 24);
  if (locale === 'zh-CN') {
    if (sec < 60) return '刚刚';
    if (min < 60) return `${min} 分钟前`;
    if (hour < 24) return `${hour} 小时前`;
    if (day < 30) return `${day} 天前`;
    return formatDate(s, locale);
  }
  if (sec < 60) return 'just now';
  if (min < 60) return `${min} min ago`;
  if (hour < 24) return `${hour} h ago`;
  if (day < 30) return `${day} d ago`;
  return formatDate(s, locale);
}

export function useFormat() {
  const { locale } = useI18n();
  return {
    fmtMoney: useCallback((v: number) => formatMoney(v, locale), [locale]),
    fmtNumber: useCallback((v: number, digits?: number) => formatNumber(v, locale, digits), [locale]),
    fmtPercent: useCallback((v: number, digits?: number) => formatPercent(v, locale, digits), [locale]),
    fmtDate: useCallback((s: string) => formatDate(s, locale), [locale]),
    fmtDateTime: useCallback((s: string) => formatDateTime(s, locale), [locale]),
    fmtRelative: useCallback((s: string) => formatRelative(s, locale), [locale]),
    locale,
  };
}

// 命令式纯函数版本（非 hook，组件外可用）
export const i18nFormat = {
  money: formatMoney,
  number: formatNumber,
  percent: formatPercent,
  date: formatDate,
  dateTime: formatDateTime,
  relative: formatRelative,
};
