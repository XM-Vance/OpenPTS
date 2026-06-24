'use client';

// 通用附件面板：拖拽上传 + 列表 + 下载 + 删除。可嵌入任何业务详情页。
import { useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Download, Paperclip, Trash2, Upload } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { extractErrorMessage } from '@/lib/api/client';
import {
  deleteAttachment,
  downloadAttachment,
  listAttachments,
  uploadAttachment,
  type FileAttachment,
} from '@/lib/api/attachment';

interface Props {
  resource: string;
  resourceId: string;
  canWrite?: boolean;
  canDelete?: boolean;
  title?: string;
}

function fmtSize(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  if (n < 1024 * 1024 * 1024) return `${(n / 1024 / 1024).toFixed(2)} MB`;
  return `${(n / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

function fmtTime(s: string): string {
  const d = new Date(s);
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

function fileEmoji(ct: string, name: string): string {
  if (ct.startsWith('image/')) return '🖼️';
  if (ct === 'application/pdf' || name.endsWith('.pdf')) return '📕';
  if (name.endsWith('.xlsx') || name.endsWith('.xls')) return '📗';
  if (name.endsWith('.docx') || name.endsWith('.doc')) return '📘';
  if (name.endsWith('.csv') || ct === 'text/csv') return '📄';
  if (name.endsWith('.zip') || name.endsWith('.tar') || name.endsWith('.gz')) return '🗜️';
  return '📎';
}

export function AttachmentPanel({
  resource,
  resourceId,
  canWrite = true,
  canDelete = true,
  title = '附件',
}: Props) {
  const qc = useQueryClient();
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['attachments', resource, resourceId],
    queryFn: () => listAttachments(resource, resourceId),
    enabled: !!resourceId,
  });

  const items = data?.items ?? [];

  const doUpload = async (files: FileList | File[]) => {
    setError(null);
    setBusy(true);
    try {
      for (const f of Array.from(files)) {
        await uploadAttachment(resource, resourceId, f);
      }
      qc.invalidateQueries({ queryKey: ['attachments', resource, resourceId] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragging(false);
    if (e.dataTransfer.files?.length) {
      doUpload(e.dataTransfer.files);
    }
  };

  const onDownload = async (a: FileAttachment) => {
    try {
      const { url } = await downloadAttachment(a.id);
      window.open(url, '_blank');
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  const onDelete = async (a: FileAttachment) => {
    if (!window.confirm(`确认删除「${a.filename}」？`)) return;
    try {
      await deleteAttachment(a.id);
      qc.invalidateQueries({ queryKey: ['attachments', resource, resourceId] });
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Paperclip className="h-4 w-4" />
          {title} {items.length > 0 && <Badge variant="secondary">{items.length}</Badge>}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {canWrite && (
          <div
            className={`flex flex-col items-center justify-center gap-2 rounded-md border-2 border-dashed p-6 text-sm transition-colors ${
              dragging
                ? 'border-blue-400 bg-blue-50 dark:bg-blue-950'
                : 'border-border bg-muted/30'
            }`}
            onDragEnter={(e) => {
              e.preventDefault();
              setDragging(true);
            }}
            onDragLeave={() => setDragging(false)}
            onDragOver={(e) => e.preventDefault()}
            onDrop={onDrop}
          >
            <Upload className="h-6 w-6 text-muted-foreground" />
            <p className="text-muted-foreground">
              拖拽文件到此处，或{' '}
              <button
                type="button"
                className="text-blue-600 hover:underline"
                onClick={() => inputRef.current?.click()}
              >
                点击选择
              </button>
              {busy && <span className="ml-2">（上传中...）</span>}
            </p>
            <input
              ref={inputRef}
              type="file"
              multiple
              className="hidden"
              onChange={(e) => e.target.files && doUpload(e.target.files)}
            />
          </div>
        )}

        {error && (
          <p className="rounded-md border border-destructive/50 bg-destructive/10 p-2 text-xs text-destructive">
            {error}
          </p>
        )}

        {isLoading ? (
          <p className="text-sm text-muted-foreground">加载中...</p>
        ) : items.length === 0 ? (
          <p className="text-center text-sm text-muted-foreground">暂无附件</p>
        ) : (
          <ul className="space-y-1">
            {items.map((a) => (
              <li
                key={a.id}
                className="flex items-center justify-between rounded-md border bg-card p-2 text-sm"
              >
                <div className="flex min-w-0 items-center gap-2">
                  <span className="text-lg">{fileEmoji(a.content_type, a.filename)}</span>
                  <div className="min-w-0">
                    <p className="truncate font-medium">{a.filename}</p>
                    <p className="text-xs text-muted-foreground">
                      {fmtSize(a.size)} · {a.uploaded_by ?? '-'} · {fmtTime(a.created_at)}
                    </p>
                  </div>
                </div>
                <div className="flex shrink-0 gap-1">
                  <Button size="sm" variant="ghost" onClick={() => onDownload(a)}>
                    <Download className="h-4 w-4" />
                  </Button>
                  {canDelete && (
                    <Button size="sm" variant="ghost" onClick={() => onDelete(a)}>
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
