import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

// shadcn 默认 helper：合并 className，处理 Tailwind 冲突。
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
