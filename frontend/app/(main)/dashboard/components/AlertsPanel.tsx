'use client';

import { Bell, AlertTriangle, Info, XCircle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useQuery } from '@tanstack/react-query';
import { getSettlementSummary } from '@/lib/api/dashboard';

const levelConfig: Record<string, { icon: React.ElementType; color: string; variant: 'destructive' | 'outline' | 'secondary' }> = {
  critical: { icon: XCircle, color: 'text-red-500', variant: 'destructive' },
  warning: { icon: AlertTriangle, color: 'text-amber-500', variant: 'outline' },
  info: { icon: Info, color: 'text-blue-500', variant: 'secondary' },
};

export default function AlertsPanel() {
  const { data, isLoading } = useQuery({
    queryKey: ['settlement-summary'],
    queryFn: () => getSettlementSummary(),
  });

  const alerts = data?.alerts ?? [];

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base flex items-center gap-2">
          <Bell className="h-4 w-4" />
          系统告警
        </CardTitle>
        {alerts.length > 0 && (
          <Badge variant="destructive" className="text-xs">
            {alerts.length}
          </Badge>
        )}
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex h-24 items-center justify-center text-sm text-muted-foreground">
            加载中...
          </div>
        ) : alerts.length > 0 ? (
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {alerts.map((alert) => {
              const cfg = levelConfig[alert.level] ?? levelConfig.info;
              const Icon = cfg.icon;
              return (
                <div
                  key={alert.id}
                  className="flex items-start gap-2 rounded-md border px-3 py-2"
                >
                  <Icon className={`h-4 w-4 mt-0.5 shrink-0 ${cfg.color}`} />
                  <div className="min-w-0 flex-1">
                    <p className="text-sm leading-tight">{alert.message}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {alert.created_at?.slice(0, 16).replace('T', ' ') ?? ''}
                    </p>
                  </div>
                  <Badge variant={cfg.variant} className="text-[10px] shrink-0">
                    {alert.level}
                  </Badge>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="flex h-24 items-center justify-center text-sm text-muted-foreground">
            <Bell className="h-4 w-4 mr-1" />
            暂无告警
          </div>
        )}
      </CardContent>
    </Card>
  );
}
