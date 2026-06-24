import { describe, expect, it } from 'vitest';
import { cn } from '@/lib/utils';

describe('cn 工具', () => {
  it('合并多个类名', () => {
    expect(cn('a', 'b')).toBe('a b');
  });

  it('过滤 falsy', () => {
    expect(cn('a', false, null, undefined, 'b')).toBe('a b');
  });

  it('Tailwind 冲突时后者覆盖前者', () => {
    // px-2 与 px-4 都是 padding-x，后者胜出
    expect(cn('px-2', 'px-4')).toBe('px-4');
  });

  it('条件类支持对象语法', () => {
    expect(cn('base', { active: true, disabled: false })).toBe('base active');
  });
});
