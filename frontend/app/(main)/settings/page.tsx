'use client';

import { useState } from 'react';
import { Sun, Moon, Monitor, Languages } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { useAuth } from '@/lib/auth/context';
import { usePermission } from '@/lib/auth/use-permission';
import { useTheme } from '@/components/theme/theme-provider';
import { useI18n } from '@/lib/i18n/context';
import { LOCALES } from '@/lib/i18n/messages';
import { extractErrorMessage } from '@/lib/api/client';
import { changePassword } from '@/lib/api/auth';

export default function SettingsPage() {
  const { user } = useAuth();
  const { permissions } = usePermission();
  const { theme, setTheme } = useTheme();
  const { locale, setLocale } = useI18n();

  const [oldPwd, setOldPwd] = useState('');
  const [newPwd, setNewPwd] = useState('');
  const [confirmPwd, setConfirmPwd] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submitPwd = async () => {
    setError(null);
    setNotice(null);
    if (newPwd.length < 8) {
      setError('新密码至少 8 位');
      return;
    }
    if (newPwd !== confirmPwd) {
      setError('两次输入的新密码不一致');
      return;
    }
    setBusy(true);
    try {
      await changePassword(oldPwd, newPwd);
      setNotice('密码已更新');
      setOldPwd('');
      setNewPwd('');
      setConfirmPwd('');
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">个人设置</h1>
        <p className="text-sm text-muted-foreground">账户信息 / 修改密码 / 显示偏好</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">账户信息</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-3">
          <div>
            <p className="text-xs text-muted-foreground">用户名</p>
            <p className="font-medium">{user?.username}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">显示名</p>
            <p className="font-medium">{user?.display_name ?? '-'}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">权限数</p>
            <p className="font-medium">{permissions.length}</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">显示偏好</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <p className="mb-2 text-xs text-muted-foreground">主题</p>
            <div className="flex gap-2">
              {([
                { code: 'system', label: '跟随系统', Icon: Monitor },
                { code: 'light', label: '浅色', Icon: Sun },
                { code: 'dark', label: '深色', Icon: Moon },
              ] as const).map(({ code, label, Icon }) => (
                <Button
                  key={code}
                  size="sm"
                  variant={theme === code ? 'default' : 'outline'}
                  onClick={() => setTheme(code)}
                >
                  <Icon className="mr-1 h-4 w-4" />
                  {label}
                </Button>
              ))}
            </div>
          </div>
          <div>
            <p className="mb-2 text-xs text-muted-foreground">语言</p>
            <div className="flex gap-2">
              {LOCALES.map((l) => (
                <Button
                  key={l.code}
                  size="sm"
                  variant={locale === l.code ? 'default' : 'outline'}
                  onClick={() => setLocale(l.code)}
                >
                  <Languages className="mr-1 h-4 w-4" />
                  {l.flag} {l.label}
                </Button>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">修改密码</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
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
          <div className="grid gap-3 md:grid-cols-3">
            <div className="space-y-1">
              <Label>原密码</Label>
              <Input
                type="password"
                value={oldPwd}
                onChange={(e) => setOldPwd(e.target.value)}
                autoComplete="current-password"
              />
            </div>
            <div className="space-y-1">
              <Label>新密码（≥ 8 位）</Label>
              <Input
                type="password"
                value={newPwd}
                onChange={(e) => setNewPwd(e.target.value)}
                autoComplete="new-password"
              />
            </div>
            <div className="space-y-1">
              <Label>确认新密码</Label>
              <Input
                type="password"
                value={confirmPwd}
                onChange={(e) => setConfirmPwd(e.target.value)}
                autoComplete="new-password"
              />
            </div>
          </div>
          <Button onClick={submitPwd} disabled={busy || !oldPwd || !newPwd}>
            {busy ? '提交中...' : '提交修改'}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">我的权限码（{permissions.length}）</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex max-h-60 flex-wrap gap-1 overflow-y-auto">
            {permissions.map((p) => (
              <Badge key={p} variant="outline" className="text-xs">
                {p}
              </Badge>
            ))}
            {permissions.length === 0 && (
              <p className="text-sm text-muted-foreground">无</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
