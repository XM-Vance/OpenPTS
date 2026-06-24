'use client';

import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { listOrgs, createOrg, updateOrg, setUserOrgs, listOrgMembers } from '@/lib/api/orgs';
import { listUsers } from '@/lib/api/users';
import { extractErrorMessage } from '@/lib/api/client';

export default function OrgsPage() {
  const qc = useQueryClient();
  const { data: orgs = [] } = useQuery({ queryKey: ['orgs'], queryFn: listOrgs });
  const { data: usersResp } = useQuery({
    queryKey: ['users', 'for-org'],
    queryFn: () => listUsers({ limit: 200 }),
  });
  const users = usersResp?.items ?? [];

  // 新建省份
  const [code, setCode] = useState('');
  const [name, setName] = useState('');
  const [err, setErr] = useState('');

  async function handleCreate() {
    setErr('');
    if (!code.trim() || !name.trim()) {
      setErr('编码与名称必填');
      return;
    }
    try {
      await createOrg(code.trim(), name.trim());
      setCode('');
      setName('');
      qc.invalidateQueries({ queryKey: ['orgs'] });
    } catch (e) {
      setErr(extractErrorMessage(e));
    }
  }

  async function toggleActive(id: string, n: string, active: boolean) {
    await updateOrg(id, n, !active);
    qc.invalidateQueries({ queryKey: ['orgs'] });
  }

  // 展开成员
  const [expandedOrg, setExpandedOrg] = useState<string | null>(null);
  const [members, setMembers] = useState<Record<string, any[]>>({});
  const [loadingMembers, setLoadingMembers] = useState(false);

  async function toggleExpand(orgId: string) {
    if (expandedOrg === orgId) {
      setExpandedOrg(null);
      return;
    }
    setExpandedOrg(orgId);
    if (!members[orgId]) {
      setLoadingMembers(true);
      try {
        const list = await listOrgMembers(orgId);
        setMembers(prev => ({ ...prev, [orgId]: list }));
      } catch { /* ignore */ }
      setLoadingMembers(false);
    }
  }

  // 用户授权
  const [userId, setUserId] = useState('');
  const [selOrgs, setSelOrgs] = useState<Set<string>>(new Set());
  const [isHQ, setIsHQ] = useState(false);
  const [primary, setPrimary] = useState('');
  const [assignMsg, setAssignMsg] = useState('');

  function toggleOrg(id: string) {
    setSelOrgs((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }

  async function handleAssign() {
    setAssignMsg('');
    if (!userId) {
      setAssignMsg('请选择用户');
      return;
    }
    try {
      await setUserOrgs(userId, [...selOrgs], isHQ, primary);
      setAssignMsg('✅ 已保存（该用户重新登录后生效）');
      // 刷新所有展开的成员列表
      const newMembers = { ...members };
      for (const key of Object.keys(newMembers)) delete newMembers[key];
      setMembers(newMembers);
      if (expandedOrg) {
        const list = await listOrgMembers(expandedOrg);
        setMembers(prev => ({ ...prev, [expandedOrg]: list }));
      }
    } catch (e) {
      setAssignMsg(extractErrorMessage(e));
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">组织管理（省份）</h1>
        <p className="text-sm text-muted-foreground">
          省份 = 组织：维护各省、把同事分配到所属省份（支持一人管多省、总部看全部）。
        </p>
      </div>

      {/* 省份列表 + 新建 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">省份 / 组织</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap items-end gap-2">
            <div>
              <Label className="text-xs">编码</Label>
              <Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="如 ZJ" className="h-8 w-24" />
            </div>
            <div>
              <Label className="text-xs">名称</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="如 浙江省" className="h-8 w-40" />
            </div>
            <Button size="sm" onClick={handleCreate}>新建省份</Button>
            {err && <span className="text-sm text-red-600">{err}</span>}
          </div>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>编码</TableHead>
                  <TableHead>名称</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>成员</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {orgs.map((o) => (
                  <>
                    <TableRow key={o.id}>
                      <TableCell className="font-mono">{o.code}</TableCell>
                      <TableCell>{o.name}</TableCell>
                      <TableCell>
                        <Badge variant={o.is_active ? 'success' : 'secondary'}>
                          {o.is_active ? '启用' : '停用'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => toggleExpand(o.id)}
                          className="text-xs"
                        >
                          {expandedOrg === o.id ? '收起成员 ▲' : '展开成员 ▼'}
                        </Button>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button variant="outline" size="sm" onClick={() => toggleActive(o.id, o.name, o.is_active)}>
                          {o.is_active ? '停用' : '启用'}
                        </Button>
                      </TableCell>
                    </TableRow>
                    {expandedOrg === o.id && (
                      <TableRow key={o.id + '-members'}>
                        <TableCell colSpan={5} className="bg-muted/30">
                          {loadingMembers && !members[o.id] ? (
                            <p className="text-sm text-muted-foreground py-2">加载中...</p>
                          ) : members[o.id] && members[o.id].length > 0 ? (
                            <div className="flex flex-wrap gap-2 py-1">
                              {members[o.id].map((m: any) => (
                                <Badge key={m.user_id} variant="outline" className="text-xs">
                                  {m.display_name || m.username}
                                  {m.is_hq ? ' 🏢' : ''}
                                  <span className="text-muted-foreground ml-1">@{m.username}</span>
                                </Badge>
                              ))}
                            </div>
                          ) : (
                            <p className="text-sm text-muted-foreground py-2">
                              暂无成员 — 请在下方「用户授权」中分配
                            </p>
                          )}
                        </TableCell>
                      </TableRow>
                    )}
                  </>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* 用户 → 省份 分配 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">用户授权（可访问省份 / 总部）</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap items-end gap-3">
            <div>
              <Label className="text-xs">用户</Label>
              <select
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                className="h-8 w-48 rounded-md border bg-background px-2 text-sm"
              >
                <option value="">选择用户…</option>
                {users.map((u) => (
                  <option key={u.id} value={u.id}>
                    {u.display_name || u.username}（{u.username}）
                  </option>
                ))}
              </select>
            </div>
            <label className="flex items-center gap-1 text-sm">
              <input type="checkbox" checked={isHQ} onChange={(e) => setIsHQ(e.target.checked)} />
              总部（可看全部省）
            </label>
            <div>
              <Label className="text-xs">主省（默认活跃）</Label>
              <select
                value={primary}
                onChange={(e) => setPrimary(e.target.value)}
                className="h-8 w-40 rounded-md border bg-background px-2 text-sm"
              >
                <option value="">（不变/无）</option>
                {orgs.map((o) => (
                  <option key={o.id} value={o.id}>{o.name}</option>
                ))}
              </select>
            </div>
          </div>
          <div>
            <Label className="text-xs">可访问省份</Label>
            <div className="mt-1 flex flex-wrap gap-3">
              {orgs.map((o) => (
                <label key={o.id} className="flex items-center gap-1 text-sm">
                  <input type="checkbox" checked={selOrgs.has(o.id)} onChange={() => toggleOrg(o.id)} />
                  {o.name}
                </label>
              ))}
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Button size="sm" onClick={handleAssign}>保存授权</Button>
            {assignMsg && <span className="text-sm text-muted-foreground">{assignMsg}</span>}
          </div>
          <p className="text-xs text-muted-foreground">
            提示：总部用户勾选「总部」即可看全部省，无需逐省勾选；非总部用户请勾选其负责的省份并设主省。
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
