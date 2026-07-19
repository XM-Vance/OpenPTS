'use client';

// 个人设置内的「我的 API Key」卡片：生成 / 列表 / 吊销。
// API Key 用于 MCP server / 外部脚本以本人身份连 ptis。明文只在创建时显示一次。
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Dialog, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { listApiKeys, createApiKey, revokeApiKey, type CreatedApiKey } from '@/lib/api/auth';
import { extractErrorMessage } from '@/lib/api/client';
import { CopyButton } from '@/components/ui/copy-button';
import { KeyRound, Loader2, Plus, Trash2, AlertTriangle } from 'lucide-react';

export function ApiKeySection() {
  const qc = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState('');
  const [newKey, setNewKey] = useState<CreatedApiKey | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { data: keys, isLoading } = useQuery({
    queryKey: ['api-keys'],
    queryFn: listApiKeys,
  });

  const createMut = useMutation({
    mutationFn: () => createApiKey(name || 'default'),
    onSuccess: (created) => {
      setNewKey(created);
      setShowCreate(false);
      setName('');
      setError(null);
      qc.invalidateQueries({ queryKey: ['api-keys'] });
    },
    onError: (e) => setError(extractErrorMessage(e)),
  });

  const revokeMut = useMutation({
    mutationFn: (id: string) => revokeApiKey(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['api-keys'] }),
  });

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2 text-base">
          <KeyRound className="h-4 w-4" />
          我的 API Key
        </CardTitle>
        <Button size="sm" onClick={() => { setShowCreate(true); setError(null); }}>
          <Plus className="mr-1 h-3.5 w-3.5" />
          生成新 Key
        </Button>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-xs text-muted-foreground">
          API Key 用于 MCP server / 外部脚本以本人身份连接 ptis。用户的权限码全程生效（只读账号不能上传）。
          详见「系统设置 → MCP 接入」。
        </p>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* 新生成的 Key 明文提示（只显示一次） */}
        {newKey && (
          <Alert>
            <AlertTriangle className="h-4 w-4 text-amber-500" />
            <AlertDescription>
              <div className="space-y-2">
                <div className="font-medium">Key 已生成，明文仅显示这一次，请立即复制保存：</div>
                <div className="flex items-center gap-2">
                  <code className="flex-1 rounded bg-muted px-2 py-1 text-xs break-all">{newKey.key}</code>
                  <CopyButton value={newKey.key} />
                </div>
                <div className="text-xs text-muted-foreground">关闭后无法再次查看，丢失需重新生成。</div>
                <Button size="sm" variant="outline" onClick={() => setNewKey(null)}>我已保存</Button>
              </div>
            </AlertDescription>
          </Alert>
        )}

        {/* Key 列表（脱敏） */}
        {isLoading ? (
          <div className="flex justify-center py-4"><Loader2 className="h-4 w-4 animate-spin text-muted-foreground" /></div>
        ) : (keys ?? []).length === 0 ? (
          <p className="py-4 text-center text-sm text-muted-foreground">暂无 API Key</p>
        ) : (
          <div className="space-y-2">
            {(keys ?? []).map((k) => (
              <div key={k.id} className="flex items-center justify-between rounded-md border p-2.5">
                <div className="space-y-0.5">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{k.name}</span>
                    {k.revoked_at && <Badge variant="destructive" className="text-xs">已吊销</Badge>}
                    <code className="text-xs text-muted-foreground">{k.key_prefix}…</code>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    创建 {new Date(k.created_at).toLocaleDateString()}
                    {k.last_used_at && ` · 最后使用 ${new Date(k.last_used_at).toLocaleDateString()}`}
                  </div>
                </div>
                {!k.revoked_at && (
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => { if (window.confirm(`吊销 Key「${k.name}」？吊销后用此 Key 的连接立即失效。`)) revokeMut.mutate(k.id); }}
                  >
                    <Trash2 className="h-4 w-4 text-red-500" />
                  </Button>
                )}
              </div>
            ))}
          </div>
        )}

        {/* 生成对话框 */}
        <Dialog open={showCreate} onClose={() => setShowCreate(false)}>
          <DialogHeader>
            <DialogTitle>生成新 API Key</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 py-2">
            <Label>名称（便于辨认，如「我的 Cursor」）</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="我的 MCP" />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>取消</Button>
            <Button onClick={() => createMut.mutate()} disabled={createMut.isPending}>
              {createMut.isPending ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : null}
              生成
            </Button>
          </DialogFooter>
        </Dialog>
      </CardContent>
    </Card>
  );
}
