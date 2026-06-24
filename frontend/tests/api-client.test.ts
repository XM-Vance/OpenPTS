import axios, { AxiosError, AxiosHeaders } from 'axios';
import { describe, expect, it } from 'vitest';
import { extractErrorMessage } from '@/lib/api/client';

function makeAxiosError(status: number, data: unknown): AxiosError {
  const err = new AxiosError(
    'Request failed',
    'ERR_BAD_REQUEST',
    { headers: new AxiosHeaders() } as any,
    null,
    {
      data,
      status,
      statusText: '',
      headers: {},
      config: { headers: new AxiosHeaders() } as any,
    },
  );
  // axios.isAxiosError 通过实例判断
  return err;
}

describe('extractErrorMessage', () => {
  it('从 AxiosError data.error 取出后端错误', () => {
    const err = makeAxiosError(400, { error: '客户名称已存在' });
    expect(extractErrorMessage(err)).toBe('客户名称已存在');
  });

  it('无 data.error 时回退到 err.message', () => {
    const err = makeAxiosError(500, {});
    expect(extractErrorMessage(err)).toBe('Request failed');
  });

  it('普通 Error 不是 AxiosError → 返回 fallback', () => {
    const err = new Error('something');
    expect(extractErrorMessage(err)).toBe('请求失败');
  });

  it('支持自定义 fallback', () => {
    expect(extractErrorMessage(undefined, '自定义兜底')).toBe('自定义兜底');
  });

  it('axios.isAxiosError 对非 AxiosError 返回 false', () => {
    expect(axios.isAxiosError(new Error('x'))).toBe(false);
  });
});
