'use client';

import { useEffect, useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
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
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createRole,
  deleteRole,
  getRole,
  listPermissions,
  listRoles,
  setRolePermissions,
  updateRole,
  type Permission,
  type Role,
} from '@/lib/api/roles';

export default function RolesPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('user_management:write');
  const canDelete = has('user_management:delete');

  const { data: roles, isLoading } = useQuery({ queryKey: ['roles'], queryFn: listRoles });
  const { data: allPerms } = useQuery({ queryKey: ['permissions'], queryFn: listPermissions });

  const [editing, setEditing] = useState<Role | 'new' | null>(null);
  const [permFor, setPermFor] = useState<Role | null>(null);
  const [compareA, setCompareA] = useState<string>('');
  const [compareB, setCompareB] = useState<string>('');

  const onDelete = async (code: string) => {
    if (!window.confirm(`确认删除角色 ${code}？`)) return;
    try {
      await deleteRole(code);
      qc.invalidateQueries({ queryKey: ['roles'] });
    } catch (e) {
      window.alert(extractErrorMessage(e));
    }
  };

  // ── Permission tree grouped by module ──
  const permTree = useMemo(() => {
    const grouped: Record<string, Permission[]> = {};
    (allPerms ?? []).forEach((p) => {
      (grouped[p.module_code] ||= []).push(p);
    });
    return grouped;
  }, [allPerms]);

  // 提取模块中文名（从 name 字段 "中文模块 - 中文操作" 的前半段）
  const moduleNameCN = useMemo(() => {
    const map: Record<string, string> = {};
    for (const p of allPerms ?? []) {
      if (!map[p.module_code] && p.name?.includes(' - ')) {
        map[p.module_code] = p.name.split(' - ')[0].trim();
      }
    }
    return map;
  }, [allPerms]);

  // 提取操作中文名（从 name 字段后半段）
  function actionCN(p: Permission): string {
    if (p.name?.includes(' - ')) return p.name.split(' - ').slice(1).join(' - ').trim();
    return p.name || '';
  }

  // ── Role perm lookup (from role.permissions field) ──
  const rolePermSet = useMemo(() => {
    const map = new Map<string, Set<string>>();
    for (const r of roles ?? []) {
      map.set(r.code, new Set(r.permissions ?? []));
    }
    return map;
  }, [roles]);

  // ── Comparison data ──
  const roleA = roles?.find((r) => r.code === compareA);
  const roleB = roles?.find((r) => r.code === compareB);
  const permSetA = roleA ? rolePermSet.get(roleA.code) : undefined;
  const permSetB = roleB ? rolePermSet.get(roleB.code) : undefined;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">角色管理</h1>
        {canWrite && <Button onClick={() => setEditing('new')}>新建角色</Button>}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {/* ═══════════ Permission Tree ═══════════ */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">权限树可视化</CardTitle>
          </CardHeader>
          <CardContent>
            {Object.keys(permTree).length === 0 ? (
              <p className="text-sm text-muted-foreground">暂无权限数据</p>
            ) : (
              <div className="space-y-3 max-h-[60vh] overflow-y-auto">
                {Object.entries(permTree).map(([mod, perms]) => (
                  <div key={mod} className="rounded-md border">
                    <div className="flex items-center gap-2 bg-muted/40 px-3 py-2">
                      <span className="font-mono text-xs font-semibold">{mod}</span>
                      {moduleNameCN[mod] && (
                        <span className="text-xs text-muted-foreground">{moduleNameCN[mod]}</span>
                      )}
                      <Badge variant="outline" className="text-xs">{perms.length}</Badge>
                    </div>
                    <ul className="px-3 py-2 space-y-1.5">
                      {perms.map((p) => {
                        // find which roles have this perm
                        const holders = (roles ?? []).filter(
                          (r) => rolePermSet.get(r.code)?.has(p.code),
                        );
                        const cn = actionCN(p);
                        return (
                          <li key={p.code} className="flex items-start gap-2 text-sm">
                            <span className="shrink-0 mt-0.5 h-2 w-2 rounded-full bg-primary" />
                            <div className="min-w-0">
                              <span className="font-medium">{p.action}</span>
                              {cn && <span className="ml-1 text-xs">{cn}</span>}
                              <span className="text-xs text-muted-foreground ml-1">({p.code})</span>
                              {holders.length > 0 && (
                                <div className="flex flex-wrap gap-1 mt-0.5">
                                  {holders.map((r) => (
                                    <Badge key={r.code} variant="secondary" className="text-[10px] px-1 py-0">
                                      {r.name}
                                    </Badge>
                                  ))}
                                </div>
                              )}
                            </div>
                          </li>
                        );
                      })}
                    </ul>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* ═══════════ Role Comparison Panel ═══════════ */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">角色权限对比</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-muted-foreground">角色 A</label>
                <select
                  value={compareA}
                  onChange={(e) => setCompareA(e.target.value)}
                  className="mt-1 flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
                >
                  <option value="">选择角色...</option>
                  {roles?.map((r) => (
                    <option key={r.code} value={r.code}>{r.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs text-muted-foreground">角色 B</label>
                <select
                  value={compareB}
                  onChange={(e) => setCompareB(e.target.value)}
                  className="mt-1 flex h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
                >
                  <option value="">选择角色...</option>
                  {roles?.map((r) => (
                    <option key={r.code} value={r.code}>{r.name}</option>
                  ))}
                </select>
              </div>
            </div>

            {compareA && compareB && permSetA && permSetB ? (
              <div className="space-y-2">
                {Object.entries(permTree).map(([mod, perms]) => (
                  <div key={mod} className="rounded-md border">
                    <div className="bg-muted/40 px-3 py-1.5">
                      <span className="font-mono text-xs font-semibold">{mod}</span>
                      {moduleNameCN[mod] && (
                        <span className="ml-1.5 text-xs text-muted-foreground">{moduleNameCN[mod]}</span>
                      )}
                    </div>
                    <div className="px-3 py-2 space-y-1">
                      {perms.map((p) => {
                        const aHas = permSetA.has(p.code);
                        const bHas = permSetB.has(p.code);
                        const cn = actionCN(p);
                        return (
                          <div key={p.code} className="grid grid-cols-[1fr_40px_40px] items-center gap-2 text-xs">
                            <span>{p.action}{cn && <span className="ml-1 text-muted-foreground">{cn}</span>}</span>
                            <span className={aHas ? 'text-emerald-600 font-bold' : 'text-muted-foreground'}>
                              {aHas ? '✓' : '✗'}
                            </span>
                            <span className={bHas ? 'text-emerald-600 font-bold' : 'text-muted-foreground'}>
                              {bHas ? '✓' : '✗'}
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                ))}
                <div className="grid grid-cols-[1fr_40px_40px] text-[10px] text-muted-foreground px-3">
                  <span />
                  <span className="text-center">{roleA?.name}</span>
                  <span className="text-center">{roleB?.name}</span>
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">请选择两个角色进行对比</p>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>编码</TableHead>
              <TableHead>名称</TableHead>
              <TableHead>描述</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>状态</TableHead>
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
            {roles?.map((r) => (
              <TableRow key={r.code}>
                <TableCell className="font-mono text-xs">{r.code}</TableCell>
                <TableCell className="font-medium">{r.name}</TableCell>
                <TableCell className="text-muted-foreground">{r.description || '-'}</TableCell>
                <TableCell>
                  {r.is_system ? (
                    <Badge variant="secondary">系统</Badge>
                  ) : (
                    <Badge variant="outline">自定义</Badge>
                  )}
                </TableCell>
                <TableCell>
                  {r.is_active ? (
                    <Badge variant="success">启用</Badge>
                  ) : (
                    <Badge variant="secondary">禁用</Badge>
                  )}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-1">
                    {canWrite && (
                      <Button size="sm" variant="ghost" onClick={() => setPermFor(r)}>
                        权限
                      </Button>
                    )}
                    {canWrite && !r.is_system && (
                      <Button size="sm" variant="ghost" onClick={() => setEditing(r)}>
                        编辑
                      </Button>
                    )}
                    {canDelete && !r.is_system && (
                      <Button size="sm" variant="ghost" onClick={() => onDelete(r.code)}>
                        删除
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {editing && (
        <RoleFormDialog
          role={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            qc.invalidateQueries({ queryKey: ['roles'] });
          }}
        />
      )}
      {permFor && (
        <RolePermissionsDialog
          role={permFor}
          onClose={() => setPermFor(null)}
          onSaved={() => {
            setPermFor(null);
            qc.invalidateQueries({ queryKey: ['roles'] });
          }}
        />
      )}
    </div>
  );
}

function RoleFormDialog({
  role,
  onClose,
  onSaved,
}: {
  role: Role | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = !role;
  const [code, setCode] = useState(role?.code ?? '');
  const [name, setName] = useState(role?.name ?? '');
  const [description, setDescription] = useState(role?.description ?? '');
  const [isActive, setIsActive] = useState(role?.is_active ?? true);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const submit = async () => {
    setError(null);
    setSubmitting(true);
    try {
      if (isNew) {
        await createRole({ code, name, description: description || undefined });
      } else if (role) {
        await updateRole(role.code, {
          name,
          description: description || undefined,
          is_active: isActive,
        });
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
        <DialogTitle>{isNew ? '新建角色' : '编辑角色'}</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>编码</Label>
          <Input
            value={code}
            onChange={(e) => setCode(e.target.value)}
            disabled={!isNew}
            placeholder="如 ops_manager"
          />
        </div>
        <div className="space-y-2">
          <Label>名称</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>描述</Label>
          <Input value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>
        {!isNew && (
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={isActive}
              onChange={(e) => setIsActive(e.target.checked)}
            />
            启用
          </label>
        )}
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          取消
        </Button>
        <Button onClick={submit} disabled={submitting}>
          {submitting ? '保存中...' : '保存'}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}

function RolePermissionsDialog({
  role,
  onClose,
  onSaved,
}: {
  role: Role;
  onClose: () => void;
  onSaved: () => void;
}) {
  const { data: allPerms } = useQuery({ queryKey: ['permissions'], queryFn: listPermissions });
  const [selected, setSelected] = useState<string[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    getRole(role.code)
      .then((full) => {
        setSelected(full.permissions ?? []);
        setLoaded(true);
      })
      .catch((e) => setError(extractErrorMessage(e, '加载角色权限失败，请重试')));
  }, [role.code]);

  const toggle = (permCode: string) => {
    setSelected((prev) =>
      prev.includes(permCode) ? prev.filter((c) => c !== permCode) : [...prev, permCode],
    );
  };

  const submit = async () => {
    setSubmitting(true);
    setError(null);
    try {
      await setRolePermissions(role.code, selected);
      onSaved();
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setSubmitting(false);
    }
  };

  // 按 module_code 分组
  const grouped: Record<string, Permission[]> = {};
  (allPerms ?? []).forEach((p) => {
    (grouped[p.module_code] ||= []).push(p);
  });

  return (
    <Dialog open onClose={onClose} className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>权限分配 · {role.name}</DialogTitle>
      </DialogHeader>
      {!loaded ? (
        error ? (
          <p className="text-sm text-destructive">{error}</p>
        ) : (
          <p className="text-sm text-muted-foreground">加载中...</p>
        )
      ) : (
        <div className="max-h-[50vh] space-y-3 overflow-y-auto">
          {Object.entries(grouped).map(([mod, perms]) => {
            const modCN = perms[0]?.name?.includes(' - ') ? perms[0].name.split(' - ')[0].trim() : '';
            return (
            <div key={mod} className="rounded-md border p-3">
              <p className="mb-2 font-mono text-xs font-medium">
                {mod}
                {modCN && <span className="ml-2 text-muted-foreground font-normal">{modCN}</span>}
              </p>
              <div className="flex flex-wrap gap-3">
                {perms.map((p) => {
                  const cn = p.name?.includes(' - ') ? p.name.split(' - ').slice(1).join(' - ').trim() : '';
                  return (
                  <label key={p.code} className="flex items-center gap-1.5 text-sm">
                    <input
                      type="checkbox"
                      checked={selected.includes(p.code)}
                      onChange={() => toggle(p.code)}
                    />
                    {p.action}
                    {cn && <span className="text-muted-foreground">({cn})</span>}
                    {p.permission_type === 'critical' && (
                      <Badge variant="destructive">critical</Badge>
                    )}
                  </label>
                  );
                })}
              </div>
            </div>
            );
          })}
        </div>
      )}
      {error && <p className="mt-2 text-xs text-destructive">{error}</p>}
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          取消
        </Button>
        <Button onClick={submit} disabled={submitting || !loaded}>
          {submitting ? '保存中...' : '保存'}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
