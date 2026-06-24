'use client';

import React, { useCallback, useState } from 'react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { Download, Loader2 } from 'lucide-react';

export interface ExportButtonProps {
  /** 导出回调，返回 Blob 或 void */
  onExport: () => Promise<Blob | void>;
  /** 按钮文字 */
  label?: string;
  /** 按钮变体 */
  variant?: 'default' | 'outline' | 'secondary' | 'ghost';
  /** 按钮大小 */
  size?: 'default' | 'sm' | 'lg' | 'icon';
  /** 默认文件名（不含扩展名） */
  filename?: string;
  /** 文件 MIME 类型 */
  mimeType?: string;
  /** 文件扩展名 */
  extension?: string;
  /** 禁用 */
  disabled?: boolean;
  /** className */
  className?: string;
}

/**
 * 导出按钮
 * - 点击触发导出
 * - 自动下载 Blob
 * - 带加载状态
 */
export function ExportButton({
  onExport,
  label = '导出',
  variant = 'outline',
  size = 'sm',
  filename,
  mimeType = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  extension = 'xlsx',
  disabled = false,
  className,
}: ExportButtonProps) {
  const [loading, setLoading] = useState(false);

  const handleClick = useCallback(async () => {
    setLoading(true);
    try {
      const result = await onExport();
      if (result instanceof Blob) {
        const url = URL.createObjectURL(result);
        const link = document.createElement('a');
        link.href = url;
        const ts = new Date().toISOString().slice(0, 19).replace(/[:-]/g, '');
        link.download = filename
          ? `${filename}_${ts}.${extension}`
          : `export_${ts}.${extension}`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
      }
    } catch (err) {
      console.error('导出失败:', err);
    } finally {
      setLoading(false);
    }
  }, [onExport, filename, extension]);

  return (
    <Button
      variant={variant}
      size={size}
      onClick={handleClick}
      disabled={disabled || loading}
      className={cn(className)}
    >
      {loading ? (
        <Loader2 className="mr-1 h-3.5 w-3.5 animate-spin" />
      ) : (
        <Download className="mr-1 h-3.5 w-3.5" />
      )}
      {loading ? '导出中...' : label}
    </Button>
  );
}
