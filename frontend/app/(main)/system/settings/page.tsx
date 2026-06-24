'use client';

import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Dialog, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { listSettings, updateSetting, type SystemSetting } from '@/lib/api/settings';
import { History, RotateCcw } from 'lucide-react';

const CATEGORY_LABEL: Record<string, string> = {
  general: '通用',
  security: '安全',
  cache: '缓存',
  alert: '告警',
  audit: '审计',
  feature: '功能开关',
  privacy: '隐私',
};

/* ── Mock change history ── */
const CHANGE_HISTORY = [
  { id: '1', key: 'cache_ttl_seconds', oldValue: '300', newValue: '600', operator: 'admin', time: '2026-06-05 14:23' },
  { id: '2', key: 'alert_threshold_load', oldValue: '0.85', newValue: '0.90', operator: 'ops_manager', time: '2026-06-04 10:15' },
  { id: '3', key: 'max_login_attempts', oldValue: '5', newValue: '3', operator: 'admin', time: '2026-06-03 09:00' },
  { id: '4', key: 'data_retention_days', oldValue: '365', newValue: '730', operator: 'admin', time: '2026-06-02 16:45' },
  { id: '5', key: 'feature_ai_forecast', oldValue: 'false', newValue: 'true', operator: 'admin', time: '2026-06-01 11:30' },
];

export default function SystemSettingsPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canEdit = has('system:write');

  const [editing, setEditing] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [rollbackItem, setRollbackItem] = useState<(typeof CHANGE_HISTORY)[number] | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['system-settings'],
    queryFn: () => listSettings(),
  });

  // 按 category 分组
  const grouped: Record<string, SystemSetting[]> = {};
  for (const s of data?.items ?? []) {
    (grouped[s.category] ||= []).push(s);
  }

  const save = async (key: string) => {
    const value = editing[key];
    if (value === undefined) return;
    setError(null);
    setNotice(null);
    try {
      await updateSetting(key, value);
      setNotice(`已更新 ${key}`);
      setEditing((prev) => {
        const cp = { ...prev };
        delete cp[key];
        return cp;
      });
      qc.invalidateQueries({ queryKey: ['system-settings'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    }
  };

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">系统配置</h1>
        <p className="text-sm text-muted-foreground">
          全局参数中心：缓存 TTL / 告警阈值 / 功能开关 / 隐私设置
        </p>
      </div>

      {notice && (
        <Alert>
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* ═══════════ Configuration Change History ═══════════ */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <History className="h-4 w-4" />
            配置变更历史
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>配置键</TableHead>
                  <TableHead>旧值</TableHead>
                  <TableHead>新值</TableHead>
                  <TableHead>操作人</TableHead>
                  <TableHead>时间</TableHead>
                  {canEdit && <TableHead className="text-right">操作</TableHead>}
                </TableRow>
              </TableHeader>
              <TableBody>
                {CHANGE_HISTORY.map((h) => (
                  <TableRow key={h.id}>
                    <TableCell>
                      <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{h.key}</code>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{h.oldValue}</TableCell>
                    <TableCell className="font-mono text-xs text-emerald-600">{h.newValue}</TableCell>
                    <TableCell>{h.operator}</TableCell>
                    <TableCell className="text-xs text-muted-foreground">{h.time}</TableCell>
                    {canEdit && (
                      <TableCell className="text-right">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setRollbackItem(h)}
                        >
                          <RotateCcw className="mr-1 h-3 w-3" />
                          回滚
                        </Button>
                      </TableCell>
                    )}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {isLoading && <p className="text-sm text-muted-foreground">加载中...</p>}

      {Object.entries(grouped).map(([cat, items]) => (
        <Card key={cat}>
          <CardHeader>
            <CardTitle className="text-base">
              {CATEGORY_LABEL[cat] ?? cat}{' '}
              <Badge variant="outline" className="ml-2">
                {items.length}
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3">
              {items.map((s) => {
                const isEditing = editing[s.key] !== undefined;
                return (
                  <li key={s.key} className="flex items-start gap-3 border-b pb-3 last:border-0">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{s.key}</code>
                        <Badge variant="secondary" className="text-xs">
                          {s.value_type}
                        </Badge>
                        {s.is_sensitive && <Badge variant="destructive" className="text-xs">敏感</Badge>}
                        {!s.is_editable && <Badge variant="outline" className="text-xs">只读</Badge>}
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">{s.description ?? '-'}</p>
                    </div>
                    <div className="flex shrink-0 items-center gap-2">
                      {isEditing ? (
                        <Input
                          value={editing[s.key]}
                          onChange={(e) =>
                            setEditing((prev) => ({ ...prev, [s.key]: e.target.value }))
                          }
                          className="w-48"
                        />
                      ) : (
                        <code className="rounded bg-muted px-2 py-1 text-sm">{s.value}</code>
                      )}
                      {canEdit && s.is_editable && (
                        isEditing ? (
                          <>
                            <Button size="sm" onClick={() => save(s.key)}>
                              保存
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() =>
                                setEditing((prev) => {
                                  const cp = { ...prev };
                                  delete cp[s.key];
                                  return cp;
                                })
                              }
                            >
                              取消
                            </Button>
                          </>
                        ) : (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setEditing((prev) => ({ ...prev, [s.key]: s.value }))}
                          >
                            编辑
                          </Button>
                        )
                      )}
                    </div>
                  </li>
                );
              })}
            </ul>
          </CardContent>
        </Card>
      ))}

      {/* ═══════════ Rollback Confirmation Dialog ═══════════ */}
      {rollbackItem && (
        <Dialog open onClose={() => setRollbackItem(null)}>
          <DialogHeader>
            <DialogTitle>确认回滚配置</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 text-sm">
            <p>
              将 <code className="rounded bg-muted px-1 py-0.5 text-xs">{rollbackItem.key}</code> 从
              <code className="mx-1 rounded bg-muted px-1 py-0.5 text-xs text-emerald-600">{rollbackItem.newValue}</code>
              回滚到
              <code className="mx-1 rounded bg-muted px-1 py-0.5 text-xs text-amber-600">{rollbackItem.oldValue}</code>
            </p>
            <p className="text-muted-foreground">此操作会立即生效，请确认无误。</p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRollbackItem(null)}>取消</Button>
            <Button onClick={async () => {
              try {
                await updateSetting(rollbackItem.key, rollbackItem.oldValue);
                setNotice(`已回滚 ${rollbackItem.key} → ${rollbackItem.oldValue}`);
                qc.invalidateQueries({ queryKey: ['system-settings'] });
              } catch (e) {
                setError(extractErrorMessage(e));
              }
              setRollbackItem(null);
            }}>
              <RotateCcw className="mr-1 h-3 w-3" />
              确认回滚
            </Button>
          </DialogFooter>
        </Dialog>
      )}
    </div>
  );
}
