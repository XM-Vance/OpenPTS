import { describe, expect, it } from 'vitest';
import { i18nFormat } from '@/lib/i18n/format';

describe('i18nFormat.money', () => {
  it('中文 1.5 万', () => {
    expect(i18nFormat.money(15000, 'zh-CN')).toBe('1.5 万');
  });
  it('中文 1.23 亿', () => {
    expect(i18nFormat.money(123_000_000, 'zh-CN')).toBe('1.23 亿');
  });
  it('英文 15k', () => {
    expect(i18nFormat.money(15000, 'en-US')).toBe('15k');
  });
  it('英文 1.5M', () => {
    expect(i18nFormat.money(1_500_000, 'en-US')).toBe('1.5M');
  });
});

describe('i18nFormat.percent', () => {
  it('中文 12.5%', () => {
    expect(i18nFormat.percent(12.5, 'zh-CN')).toBe('12.5%');
  });
});

describe('i18nFormat.date', () => {
  it('中文 yyyy-mm-dd', () => {
    expect(i18nFormat.date('2026-05-25T10:00:00Z', 'zh-CN')).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
  it('空字符串返回 -', () => {
    expect(i18nFormat.date('', 'zh-CN')).toBe('-');
  });
});

describe('i18nFormat.relative', () => {
  it('30 秒前 → 刚刚', () => {
    const ts = new Date(Date.now() - 30_000).toISOString();
    expect(i18nFormat.relative(ts, 'zh-CN')).toBe('刚刚');
  });
  it('5 分钟前 → 5 分钟前', () => {
    const ts = new Date(Date.now() - 5 * 60_000).toISOString();
    expect(i18nFormat.relative(ts, 'zh-CN')).toBe('5 分钟前');
  });
  it('英文 just now', () => {
    const ts = new Date(Date.now() - 5_000).toISOString();
    expect(i18nFormat.relative(ts, 'en-US')).toBe('just now');
  });
});
