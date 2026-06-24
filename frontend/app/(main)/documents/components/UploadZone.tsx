'use client';

import { useState, useCallback, useRef } from 'react';
import { Upload, File as FileIcon, CheckCircle2, XCircle, Loader2, X } from 'lucide-react';
import { uploadDocument } from '@/lib/api/documents';
import { extractErrorMessage } from '@/lib/api/client';
import { Button } from '@/components/ui/button';

const ACCEPTED = '.pdf,.jpg,.jpeg,.png,.bmp,.tiff,.webp,.doc,.docx,.xlsx,.xlsm,.xls,.csv';

type UploadStatus = 'pending' | 'uploading' | 'done' | 'error' | 'duplicate';

interface UploadTask {
  id: string;
  file: File;
  status: UploadStatus;
  progress: number;
  message?: string;
}

function fmtSize(n: number): string {
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB';
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(0) + ' KB';
  return n + ' B';
}

const STATUS_ICON: Record<UploadStatus, React.ReactNode> = {
  pending: <div className="h-4 w-4 rounded-full border-2 border-muted-foreground/30" />,
  uploading: <Loader2 className="h-4 w-4 animate-spin text-blue-500" />,
  done: <CheckCircle2 className="h-4 w-4 text-emerald-500" />,
  error: <XCircle className="h-4 w-4 text-red-500" />,
  duplicate: <CheckCircle2 className="h-4 w-4 text-amber-500" />,
};

const STATUS_LABEL: Record<UploadStatus, string> = {
  pending: '等待中',
  uploading: '上传中',
  done: '完成',
  error: '失败',
  duplicate: '重复跳过',
};

const BAR_COLOR: Record<UploadStatus, string> = {
  pending: 'bg-muted',
  uploading: 'bg-blue-500',
  done: 'bg-emerald-500',
  error: 'bg-red-500',
  duplicate: 'bg-amber-500',
};

interface Props {
  /** 上传全部完成后回调（用于刷新列表） */
  onComplete: () => void;
}

export default function UploadZone({ onComplete }: Props) {
  const [tasks, setTasks] = useState<UploadTask[]>([]);
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const updateTask = useCallback((id: string, patch: Partial<UploadTask>) => {
    setTasks((prev) => prev.map((t) => (t.id === id ? { ...t, ...patch } : t)));
  }, []);

  const processQueue = useCallback(
    async (initialTasks: UploadTask[]) => {
      setUploading(true);
      let hasSuccess = false;

      for (const task of initialTasks) {
        updateTask(task.id, { status: 'uploading', progress: 0 });
        try {
          const r = await uploadDocument(task.file, (pct) => {
            updateTask(task.id, { progress: pct });
          });
          if (r.duplicated) {
            updateTask(task.id, { status: 'duplicate', progress: 100, message: r.message });
          } else {
            updateTask(task.id, { status: 'done', progress: 100 });
            hasSuccess = true;
          }
        } catch (e) {
          updateTask(task.id, { status: 'error', progress: 0, message: extractErrorMessage(e) });
        }
      }

      setUploading(false);
      if (hasSuccess) {
        // 延迟刷新列表，等后端写入
        setTimeout(onComplete, 300);
      }
    },
    [updateTask, onComplete],
  );

  const handleFiles = useCallback(
    (fileList: FileList | null) => {
      if (!fileList || fileList.length === 0) return;
      const newTasks: UploadTask[] = Array.from(fileList).map((file, i) => ({
        id: `${Date.now()}-${i}-${file.name}`,
        file,
        status: 'pending',
        progress: 0,
      }));
      setTasks((prev) => [...prev, ...newTasks]);
      processQueue(newTasks);
    },
    [processQueue],
  );

  // 拖拽事件
  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      handleFiles(e.dataTransfer.files);
    },
    [handleFiles],
  );

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  }, []);

  const onDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
  }, []);

  // 清除已完成/失败的
  const clearFinished = useCallback(() => {
    setTasks((prev) => prev.filter((t) => t.status === 'uploading' || t.status === 'pending'));
  }, []);

  // 全部清除
  const clearAll = useCallback(() => setTasks([]), []);

  const doneCount = tasks.filter((t) => t.status === 'done').length;
  const errCount = tasks.filter((t) => t.status === 'error').length;
  const dupCount = tasks.filter((t) => t.status === 'duplicate').length;
  const totalProgress =
    tasks.length > 0
      ? Math.round(tasks.reduce((sum, t) => sum + t.progress, 0) / tasks.length)
      : 0;

  return (
    <div className="space-y-3">
      {/* Drop zone */}
      <div
        onDrop={onDrop}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
        onClick={() => fileInputRef.current?.click()}
        className={`flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-6 transition-colors ${
          dragOver
            ? 'border-primary bg-primary/5'
            : 'border-muted-foreground/25 hover:border-primary/50 hover:bg-muted/30'
        }`}
      >
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
          <Upload className="h-6 w-6 text-primary" />
        </div>
        <p className="text-sm font-medium">
          点击选择文件或拖拽文件到此处
        </p>
        <p className="text-xs text-muted-foreground">
          支持 PDF / 图片 / Word / Excel / CSV，可同时选择多个文件
        </p>
        <input
          ref={fileInputRef}
          type="file"
          multiple
          accept={ACCEPTED}
          className="hidden"
          onChange={(e) => {
            handleFiles(e.target.files);
            e.target.value = ''; // 重置以便重复选同一文件
          }}
        />
      </div>

      {/* Upload progress */}
      {tasks.length > 0 && (
        <div className="rounded-lg border bg-white">
          {/* Header */}
          <div className="flex items-center justify-between border-b px-4 py-2">
            <div className="flex items-center gap-3 text-sm">
              <span className="font-medium">上传队列</span>
              <span className="text-muted-foreground">
                共 {tasks.length} 个
                {doneCount > 0 && <span className="text-emerald-600"> · 成功 {doneCount}</span>}
                {dupCount > 0 && <span className="text-amber-600"> · 重复 {dupCount}</span>}
                {errCount > 0 && <span className="text-red-600"> · 失败 {errCount}</span>}
              </span>
              {uploading && (
                <span className="flex items-center gap-1 text-blue-600">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  总进度 {totalProgress}%
                </span>
              )}
            </div>
            {!uploading && (
              <div className="flex items-center gap-2">
                {tasks.some((t) => t.status !== 'uploading' && t.status !== 'pending') && (
                  <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={clearFinished}>
                    清除已完成
                  </Button>
                )}
                <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={clearAll}>
                  <X className="mr-0.5 h-3 w-3" />
                  全部清除
                </Button>
              </div>
            )}
          </div>

          {/* Task list */}
          <div className="max-h-64 overflow-y-auto divide-y">
            {tasks.map((task) => (
              <div key={task.id} className="flex items-center gap-3 px-4 py-2.5">
                <FileIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between gap-2">
                    <span className="truncate text-sm font-medium" title={task.file.name}>
                      {task.file.name}
                    </span>
                    <span className="shrink-0 text-xs text-muted-foreground">
                      {fmtSize(task.file.size)}
                    </span>
                  </div>
                  {/* Progress bar */}
                  <div className="mt-1 flex items-center gap-2">
                    <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                      <div
                        className={`h-full rounded-full transition-all duration-300 ${BAR_COLOR[task.status]}`}
                        style={{ width: `${task.progress}%` }}
                      />
                    </div>
                    <div className="flex w-16 shrink-0 items-center justify-end gap-1 text-xs">
                      {STATUS_ICON[task.status]}
                      <span
                        className={
                          task.status === 'done'
                            ? 'text-emerald-600'
                            : task.status === 'error'
                              ? 'text-red-600'
                              : task.status === 'duplicate'
                                ? 'text-amber-600'
                                : 'text-muted-foreground'
                        }
                      >
                        {task.status === 'uploading' ? `${task.progress}%` : STATUS_LABEL[task.status]}
                      </span>
                    </div>
                  </div>
                  {/* Error message */}
                  {task.message && task.status === 'error' && (
                    <p className="mt-0.5 truncate text-xs text-red-500" title={task.message}>
                      {task.message}
                    </p>
                  )}
                  {task.message && task.status === 'duplicate' && (
                    <p className="mt-0.5 truncate text-xs text-amber-500" title={task.message}>
                      {task.message}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
