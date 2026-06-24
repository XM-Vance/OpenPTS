'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { menuApi, type MenuPageWithRoles } from '@/lib/api/menu';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Loader2, Save, Lock, CheckCircle2, AlertCircle } from 'lucide-react';

const ROLES = [
  { code: 'super_admin', name: '超级管理员', color: 'bg-red-100 text-red-700' },
  { code: 'admin', name: '管理员', color: 'bg-blue-100 text-blue-700' },
  { code: 'analyst', name: '分析师', color: 'bg-green-100 text-green-700' },
  { code: 'viewer', name: '只读用户', color: 'bg-gray-100 text-gray-700' },
];

export default function MenuManagementPage() {
  const qc = useQueryClient();
  const [selectedRole, setSelectedRole] = useState('viewer');
  const [pendingChanges, setPendingChanges] = useState<Record<string, Set<string>>>({});
  const [statusMsg, setStatusMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const { data: pages, isLoading } = useQuery({
    queryKey: ['menu-pages'],
    queryFn: menuApi.getAllPages,
  });

  const saveMutation = useMutation({
    mutationFn: ({ roleCode, pageCodes }: { roleCode: string; pageCodes: string[] }) =>
      menuApi.updateRolePages(roleCode, pageCodes),
    onSuccess: () => {
      setStatusMsg({ type: 'success', text: '菜单权限已保存' });
      qc.invalidateQueries({ queryKey: ['menu-pages'] });
      qc.invalidateQueries({ queryKey: ['menu-visible'] });
      setPendingChanges({});
      setTimeout(() => setStatusMsg(null), 3000);
    },
    onError: () => {
      setStatusMsg({ type: 'error', text: '保存失败，请重试' });
    },
  });

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  // 按模块分组
  const groups = new Map<string, MenuPageWithRoles[]>();
  for (const page of pages || []) {
    if (!groups.has(page.group_name)) {
      groups.set(page.group_name, []);
    }
    groups.get(page.group_name)!.push(page);
  }

  // 获取某角色在某页面的勾选状态（含未提交的改动）
  const isChecked = (roleCode: string, page: MenuPageWithRoles): boolean => {
    const pending = pendingChanges[roleCode];
    if (pending) {
      return pending.has(page.code);
    }
    return page.roles.includes(roleCode);
  };

  const toggle = (roleCode: string, page: MenuPageWithRoles) => {
    if (page.is_required) return; // 必要页面不可取消

    setPendingChanges((prev) => {
      const current = new Set(prev[roleCode] || pages!.filter((p) => p.roles.includes(roleCode)).map((p) => p.code));
      if (current.has(page.code)) {
        current.delete(page.code);
      } else {
        current.add(page.code);
      }
      return { ...prev, [roleCode]: current };
    });
  };

  const handleSave = (roleCode: string) => {
    const codes = Array.from(pendingChanges[roleCode] || []);
    if (codes.length === 0) {
      setStatusMsg({ type: 'error', text: '没有变更需要保存' });
      setTimeout(() => setStatusMsg(null), 3000);
      return;
    }
    saveMutation.mutate({ roleCode, pageCodes: codes });
  };

  const hasChanges = (roleCode: string) => !!pendingChanges[roleCode];

  return (
    <div className="space-y-6 p-6">
      <div>
        <h1 className="text-2xl font-bold">菜单管理</h1>
        <p className="text-sm text-muted-foreground mt-1">
          管理每个角色可以看到的菜单页面。勾选 = 该角色可见，取消 = 隐藏。
          <Lock className="inline h-3 w-3 ml-2" /> 标记的页面为必要页面，不可隐藏。
        </p>
      </div>

      {/* 角色选择 Tab */}
      <div className="flex gap-2 flex-wrap">
        {ROLES.map((role) => (
          <button
            key={role.code}
            onClick={() => setSelectedRole(role.code)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              selectedRole === role.code
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted hover:bg-muted/80'
            }`}
          >
            {role.name}
            {hasChanges(role.code) && (
              <span className="ml-2 inline-flex h-2 w-2 rounded-full bg-amber-400" />
            )}
          </button>
        ))}
      </div>

      {/* 页面列表 */}
      <div className="space-y-4">
        {Array.from(groups.entries()).map(([groupName, groupPages]) => (
          <Card key={groupName}>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">{groupName}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {groupPages.map((page) => (
                  <div
                    key={page.code}
                    className="flex items-center justify-between rounded-lg border p-3 hover:bg-accent/30"
                  >
                    <div className="flex items-center gap-3">
                      <Checkbox
                        checked={isChecked(selectedRole, page)}
                        onCheckedChange={() => toggle(selectedRole, page)}
                        disabled={page.is_required}
                      />
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-sm">{page.label}</span>
                          {page.is_required && (
                            <Badge variant="secondary" className="text-xs">
                              <Lock className="mr-1 h-3 w-3" />
                              必要
                            </Badge>
                          )}
                        </div>
                        <span className="text-xs text-muted-foreground">{page.href}</span>
                      </div>
                    </div>
                    {/* 显示哪些角色有权限 */}
                    <div className="flex gap-1">
                      {ROLES.map((role) => {
                        const visible = page.roles.includes(role.code);
                        return (
                          <Badge
                            key={role.code}
                            variant="outline"
                            className={`text-xs ${
                              visible ? role.color : 'opacity-30'
                            }`}
                          >
                            {role.name}
                          </Badge>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 状态提示 */}
      {statusMsg && (
        <div className={`fixed bottom-20 right-6 flex items-center gap-2 rounded-lg px-4 py-2 text-sm shadow-lg ${
          statusMsg.type === 'success' ? 'bg-green-600 text-white' : 'bg-red-600 text-white'
        }`}>
          {statusMsg.type === 'success' ? (
            <CheckCircle2 className="h-4 w-4" />
          ) : (
            <AlertCircle className="h-4 w-4" />
          )}
          {statusMsg.text}
        </div>
      )}

      {/* 保存按钮 */}
      <div className="sticky bottom-4 flex justify-end gap-2">
        {hasChanges(selectedRole) && (
          <Button
            onClick={() => handleSave(selectedRole)}
            disabled={saveMutation.isPending}
          >
            {saveMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Save className="mr-2 h-4 w-4" />
            )}
            保存 {ROLES.find((r) => r.code === selectedRole)?.name} 的菜单权限
          </Button>
        )}
      </div>
    </div>
  );
}
