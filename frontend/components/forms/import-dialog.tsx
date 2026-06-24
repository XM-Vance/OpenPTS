'use client';

import React, { useCallback, useRef, useState } from 'react';
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { Upload, FileText, X, AlertCircle, CheckCircle2 } from 'lucide-react';

export interface ImportResult {
  total: number;
  success: number;
  failed: number;
  errors: ImportError[];
}

export interface ImportError {
  row: number;
  field: string;
  message: string;
  suggestion?: string;
}

export interface ImportDialogProps {
  /** 是否显示 */
  open: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 导入回调，接收 File 对象，返回 Promise<ImportResult> */
  onImport: (file: File) => Promise<ImportResult>;
  /** 导入成功回调 */
  onSuccess?: (result: ImportResult) => void;
  /** 标题 */
  title?: string;
  /** 接受的文件类型 */
  accept?: string;
  /** 最大文件大小（字节），默认 10MB */
  maxSize?: number;
  /** 提示文字 */
  hint?: string;
  /** 容器 className */
  className?: string;
}

/**
 * CSV/Excel 导入弹窗
 * - 文件拖拽 + 选择
 * - 文件校验
 * - 导入结果展示
 */
export function ImportDialog({
  open,
  onClose,
  onImport,
  onSuccess,
  title = '导入数据',
  accept = '.csv,.xlsx,.xls',
  maxSize = 10 * 1024 * 1024,
  hint,
  className,
}: ImportDialogProps) {
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ImportResult | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const reset = useCallback(() => {
    setFile(null);
    setLoading(false);
    setError(null);
    setResult(null);
    if (inputRef.current) inputRef.current.value = '';
  }, []);

  const handleClose = useCallback(() => {
    reset();
    onClose();
  }, [reset, onClose]);

  const validateFile = (f: File): boolean => {
    if (f.size > maxSize) {
      setError(`文件大小不能超过 ${(maxSize / 1024 / 1024).toFixed(0)}MB`);
      return false;
    }
    const ext = f.name.split('.').pop()?.toLowerCase();
    const allowedExts = accept.split(',').map((s) => s.trim().replace('.', ''));
    if (ext && !allowedExts.includes(ext)) {
      setError(`不支持的文件格式，请选择 ${accept} 文件`);
      return false;
    }
    return true;
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (!f) return;
    setError(null);
    setResult(null);
    if (validateFile(f)) setFile(f);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    const f = e.dataTransfer.files[0];
    if (!f) return;
    setError(null);
    setResult(null);
    if (validateFile(f)) setFile(f);
  };

  const handleImport = async () => {
    if (!file) return;
    setLoading(true);
    setError(null);
    try {
      const res = await onImport(file);
      setResult(res);
      onSuccess?.(res);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '导入失败，请重试';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} className={cn('max-w-lg', className)}>
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>
          {hint || '选择文件后点击导入'}
        </DialogDescription>
      </DialogHeader>

      {/* 错误提示 */}
      {error && (
        <div className="mb-4 flex items-center gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          <AlertCircle className="h-4 w-4 shrink-0" />
          {error}
        </div>
      )}

      {/* 文件选择区域 */}
      {!result && (
        <div
          className={cn(
            'mb-4 flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 transition-colors',
            isDragOver ? 'border-primary bg-primary/5' : 'border-muted-foreground/25',
          )}
          onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
          onDragLeave={() => setIsDragOver(false)}
          onDrop={handleDrop}
        >
          <Upload className="mb-2 h-8 w-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">
            拖拽文件到此处，或
            <button
              type="button"
              className="mx-1 text-primary underline"
              onClick={() => inputRef.current?.click()}
            >
              点击选择文件
            </button>
          </p>
          <input
            ref={inputRef}
            type="file"
            accept={accept}
            className="hidden"
            onChange={handleFileSelect}
          />
        </div>
      )}

      {/* 已选文件 */}
      {file && !result && (
        <div className="mb-4 flex items-center gap-2 rounded-md border px-3 py-2">
          <FileText className="h-4 w-4 text-muted-foreground" />
          <span className="flex-1 truncate text-sm">{file.name}</span>
          <span className="text-xs text-muted-foreground">
            {(file.size / 1024).toFixed(1)} KB
          </span>
          <button
            type="button"
            onClick={() => { setFile(null); if (inputRef.current) inputRef.current.value = ''; }}
            className="text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* 导入结果 */}
      {result && (
        <div className="mb-4 space-y-3">
          <div className="flex gap-2">
            <Badge variant="secondary">总计: {result.total}</Badge>
            <Badge variant="success">成功: {result.success}</Badge>
            <Badge variant={result.failed > 0 ? 'destructive' : 'secondary'}>
              失败: {result.failed}
            </Badge>
          </div>

          {result.failed === 0 && result.success > 0 && (
            <div className="flex items-center gap-2 text-sm text-emerald-600">
              <CheckCircle2 className="h-4 w-4" />
              成功导入 {result.success} 条数据！
            </div>
          )}

          {result.errors.length > 0 && (
            <div className="max-h-40 overflow-auto rounded-md border text-xs">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="px-2 py-1 text-left">行号</th>
                    <th className="px-2 py-1 text-left">字段</th>
                    <th className="px-2 py-1 text-left">错误</th>
                  </tr>
                </thead>
                <tbody>
                  {result.errors.slice(0, 10).map((err, i) => (
                    <tr key={i} className="border-b last:border-0">
                      <td className="px-2 py-1">{err.row}</td>
                      <td className="px-2 py-1">{err.field}</td>
                      <td className="px-2 py-1 text-red-600">{err.message}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {result.errors.length > 10 && (
                <p className="px-2 py-1 text-center text-muted-foreground">
                  显示前10条，共 {result.errors.length} 条错误
                </p>
              )}
            </div>
          )}
        </div>
      )}

      <DialogFooter>
        <Button variant="outline" onClick={handleClose} disabled={loading}>
          {result ? '关闭' : '取消'}
        </Button>
        {!result && (
          <Button onClick={handleImport} disabled={!file || loading}>
            {loading ? '导入中...' : '导入'}
          </Button>
        )}
      </DialogFooter>
    </Dialog>
  );
}
