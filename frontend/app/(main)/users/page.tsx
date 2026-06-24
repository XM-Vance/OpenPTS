'use client';

import { useEffect, useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createUser,
  getUser,
  listUsers,
  resetUserPassword,
  setUserRoles,
  updateUser,
  type User,
} from '@/lib/api/users';
import { listRoles, type Role } from '@/lib/api/roles';

/* ── Mock data for analytics panels ── */
const TOP_FEATURES = [
  { name: '负荷预测', count: 1842 },
  { name: '日前交易', count: 1523 },
  { name: '结算管理', count: 1298 },
  { name: '合同查看', count: 1087 },
  { name: '客户管理', count: 945 },
  { name: '报表导出', count: 876 },
  { name: '审批中心', count: 654 },
  { name: '价格查询', count: 532 },
  { name: '保函管理', count: 421 },
  { name: '系统设置', count: 312 },
];

const BAR_COLORS = [
  '#6366f1', '#3b82f6', '#10b981', '#f97316', '#8b5cf6',
  '#ec4899', '#14b8a6', '#f59e0b', '#ef4444', '#64748b',
];

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const FeatureBar = dynamic(
  () => import('./_charts').then((m) => ({ default: m.FeatureBar })),
  { ssr: false, loading: () => <div className="h-[300px] w-full" /> },
);
const OnlinePie = dynamic(
  () => import('./_charts').then((m) => ({ default: m.OnlinePie })),
  { ssr: false, loading: () => <div className="h-[300px] w-full" /> },
);

export default function UsersPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('user_management:write');

  const [keyword, setKeyword] = useState('');
  const [search, setSearch] = useState('');
  const [editing, setEditing] = useState<User | 'new' | null>(null);
  const [resetting, setResetting] = useState<User | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['users', search],
    queryFn: () => listUsers({ keyword: search, limit: 100 }),
  });
  const { data: roles } = useQuery({ queryKey: ['roles'], queryFn: listRoles });

  const items = useMemo(() => data?.items ?? [], [data]);

  // ── Online / Offline Pie data ──
  const onlineStats = useMemo(() => {
    const online = items.filter((u) => u.is_active).length;
    const offline = items.length - online;
    return [
      { name: '在线', value: online, fill: '#10b981' },
      { name: '离线', value: offline, fill: '#94a3b8' },
    ];
  }, [items]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">用户管理</h1>
        {canWrite && <Button onClick={() => setEditing('new')}>新建用户</Button>}
      </div>

      <div className="flex gap-2">
        <Input
          placeholder="搜索用户名 / 显示名"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && setSearch(keyword)}
          className="max-w-xs"
        />
        <Button variant="outline" onClick={() => setSearch(keyword)}>
          搜索
        </Button>
      </div>

      {/* ═══════════ Analytics Panels ═══════════ */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* TOP10 feature usage */}
        <ChartContainer title="功能使用 TOP10（操作热力图）">
          <FeatureBar data={TOP_FEATURES} colors={BAR_COLORS} />
        </ChartContainer>

        {/* Online / Offline Pie */}
        <ChartContainer title="在线用户状态">
          <OnlinePie data={onlineStats} />
        </ChartContainer>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户名</TableHead>
              <TableHead>显示名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>创建时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((u) => (
              <TableRow key={u.id}>
                <TableCell className="font-medium">{u.username}</TableCell>
                <TableCell>{u.display_name || '-'}</TableCell>
                <TableCell>{u.email || '-'}</TableCell>
                <TableCell>
                  {u.is_active ? (
                    <Badge variant="success">启用</Badge>
                  ) : (
                    <Badge variant="secondary">禁用</Badge>
                  )}
                </TableCell>
                <TableCell>{u.created_at?.slice(0, 10)}</TableCell>
                <TableCell className="text-right">
                  {canWrite && (
                    <div className="flex justify-end gap-1">
                      <Button size="sm" variant="ghost" onClick={() => setEditing(u)}>
                        编辑
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => setResetting(u)}>
                        重置密码
                      </Button>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {editing && (
        <UserFormDialog
          user={editing === 'new' ? null : editing}
          roles={roles || []}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            qc.invalidateQueries({ queryKey: ['users'] });
          }}
        />
      )}
      {resetting && (
        <ResetPasswordDialog user={resetting} onClose={() => setResetting(null)} />
      )}
    </div>
  );
}

function UserFormDialog({
  user,
  roles,
  onClose,
  onSaved,
}: {
  user: User | null;
  roles: Role[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = !user;
  const [username, setUsername] = useState(user?.username ?? '');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState(user?.display_name ?? '');
  const [email, setEmail] = useState(user?.email ?? '');
  const [isActive, setIsActive] = useState(user?.is_active ?? true);
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [loadingRoles, setLoadingRoles] = useState(!isNew);
  // 现有用户的角色加载失败标记:此时 selectedRoles 仍是空的初始值,
  // 若放行保存会用空角色覆盖、清空该用户的全部角色(数据丢失)。故失败时禁存。
  const [roleLoadFailed, setRoleLoadFailed] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isNew && user) {
      setRoleLoadFailed(false);
      getUser(user.id)
        .then((full) => setSelectedRoles(full.roles ?? []))
        .catch((e) => {
          setRoleLoadFailed(true);
          setError(extractErrorMessage(e, '加载用户角色失败，请重试'));
        })
        .finally(() => setLoadingRoles(false));
    }
  }, [isNew, user]);

  const toggleRole = (code: string) => {
    setSelectedRoles((prev) =>
      prev.includes(code) ? prev.filter((c) => c !== code) : [...prev, code],
    );
  };

  const submit = async () => {
    // 双保险:角色未加载成功时绝不提交(否则 setUserRoles 会用空数组清空角色)
    if (!isNew && roleLoadFailed) {
      setError('用户角色未加载成功，已阻止保存以免覆盖现有角色');
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      if (isNew) {
        await createUser({
          username,
          password,
          display_name: displayName,
          email: email || undefined,
          roles: selectedRoles,
        });
      } else if (user) {
        await updateUser(user.id, {
          display_name: displayName,
          email: email || undefined,
          is_active: isActive,
        });
        await setUserRoles(user.id, selectedRoles);
      }
      onSaved();
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open onClose={onClose}>
      <DialogHeader>
        <DialogTitle>{isNew ? '新建用户' : '编辑用户'}</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>用户名</Label>
          <Input value={username} onChange={(e) => setUsername(e.target.value)} disabled={!isNew} />
        </div>
        {isNew && (
          <div className="space-y-2">
            <Label>初始密码</Label>
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
        )}
        <div className="space-y-2">
          <Label>显示名</Label>
          <Input value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>邮箱</Label>
          <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
        {!isNew && (
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={isActive}
              onChange={(e) => setIsActive(e.target.checked)}
            />
            启用账号
          </label>
        )}
        <div className="space-y-2">
          <Label>角色</Label>
          {loadingRoles ? (
            <p className="text-xs text-muted-foreground">加载角色...</p>
          ) : roleLoadFailed ? (
            <p className="text-xs text-destructive">
              当前角色加载失败，为避免误清空已禁止保存，请关闭重试。
            </p>
          ) : (
            <div className="space-y-1">
              {roles.map((r) => (
                <label key={r.code} className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={selectedRoles.includes(r.code)}
                    onChange={() => toggleRole(r.code)}
                  />
                  {r.name}
                  <span className="text-xs text-muted-foreground">({r.code})</span>
                </label>
              ))}
            </div>
          )}
          {isNew && selectedRoles.length === 0 && (
            <p className="text-xs text-amber-600">
              未选择角色时，将默认赋予「只读用户(viewer)」（可登录并只读查看全部模块），可事后提权。
            </p>
          )}
        </div>
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          取消
        </Button>
        <Button onClick={submit} disabled={submitting || roleLoadFailed}>
          {submitting ? '保存中...' : '保存'}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}

function ResetPasswordDialog({ user, onClose }: { user: User; onClose: () => void }) {
  const [pwd, setPwd] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const submit = async () => {
    if (pwd.length < 4) {
      setError('密码至少 4 位');
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      await resetUserPassword(user.id, pwd);
      setDone(true);
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open onClose={onClose}>
      <DialogHeader>
        <DialogTitle>重置密码 · {user.username}</DialogTitle>
      </DialogHeader>
      {done ? (
        <p className="text-sm text-emerald-600">密码已重置</p>
      ) : (
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>新密码</Label>
            <Input type="password" value={pwd} onChange={(e) => setPwd(e.target.value)} />
          </div>
          {error && <p className="text-xs text-destructive">{error}</p>}
        </div>
      )}
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          {done ? '关闭' : '取消'}
        </Button>
        {!done && (
          <Button onClick={submit} disabled={submitting}>
            {submitting ? '提交中...' : '确认重置'}
          </Button>
        )}
      </DialogFooter>
    </Dialog>
  );
}
