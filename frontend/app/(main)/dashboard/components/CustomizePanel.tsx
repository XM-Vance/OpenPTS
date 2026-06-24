'use client';

import { useState, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Settings2,
  ChevronUp,
  ChevronDown,
  Eye,
  EyeOff,
  RotateCcw,
  Check,
} from 'lucide-react';
import {
  type DashboardConfig,
  type WidgetConfig,
  WIDGET_REGISTRY,
  getDefaultConfig,
  loadConfig,
  saveConfig,
} from '@/lib/api/dashboard-config';

interface Props {
  config: DashboardConfig;
  onChange: (config: DashboardConfig) => void;
}

/** 自定义模式面板：开关 + 上下移动 + 恢复默认 */
export default function CustomizePanel({ config, onChange }: Props) {
  const [dirty, setDirty] = useState(false);
  const [saving, setSaving] = useState(false);

  const visibleCount = config.widgets.filter((w) => w.visible).length;
  const hiddenCount = config.widgets.length - visibleCount;

  const toggle = (id: string) => {
    onChange({
      widgets: config.widgets.map((w) =>
        w.id === id ? { ...w, visible: !w.visible } : w,
      ),
    });
    setDirty(true);
  };

  const move = (index: number, dir: -1 | 1) => {
    const newIndex = index + dir;
    if (newIndex < 0 || newIndex >= config.widgets.length) return;
    const widgets = [...config.widgets];
    [widgets[index], widgets[newIndex]] = [widgets[newIndex], widgets[index]];
    onChange({ widgets });
    setDirty(true);
  };

  const reset = () => {
    onChange(getDefaultConfig());
    setDirty(true);
  };

  const save = async () => {
    setSaving(true);
    try {
      await saveConfig(config);
      setDirty(false);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card className="mb-4 border-primary/30">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-1.5 text-sm font-medium">
            <Settings2 className="h-4 w-4 text-primary" />
            自定义仪表盘
            <span className="text-xs font-normal text-muted-foreground">
              ({visibleCount} 显示 · {hiddenCount} 隐藏)
            </span>
          </CardTitle>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="ghost"
              className="h-7 gap-1 text-xs"
              onClick={reset}
              disabled={saving}
            >
              <RotateCcw className="h-3 w-3" />
              恢复默认
            </Button>
            <Button
              size="sm"
              variant={dirty ? 'default' : 'outline'}
              className="h-7 gap-1 text-xs"
              onClick={save}
              disabled={!dirty || saving}
            >
              <Check className="h-3 w-3" />
              {saving ? '保存中…' : '保存'}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          {config.widgets.map((widget: WidgetConfig, index: number) => {
            const def = WIDGET_REGISTRY.find((w) => w.id === widget.id);
            if (!def) return null;
            const groupLabel =
              def.group === 'kpi' ? '卡片' : def.group === 'spark' ? '趋势' : '面板';
            return (
              <div
                key={widget.id}
                className={`flex items-center gap-3 rounded-lg border px-3 py-2 transition-colors ${
                  widget.visible ? 'bg-white' : 'bg-muted/30 opacity-60'
                }`}
              >
                {/* 显示/隐藏开关 */}
                <button
                  onClick={() => toggle(widget.id)}
                  className="shrink-0 rounded p-1 hover:bg-muted"
                  title={widget.visible ? '点击隐藏' : '点击显示'}
                >
                  {widget.visible ? (
                    <Eye className="h-4 w-4 text-primary" />
                  ) : (
                    <EyeOff className="h-4 w-4 text-muted-foreground" />
                  )}
                </button>

                {/* 信息 */}
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{def.label}</span>
                    <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                      {groupLabel}
                    </span>
                  </div>
                  <p className="truncate text-xs text-muted-foreground">
                    {def.description}
                  </p>
                </div>

                {/* 上下移动 */}
                <div className="flex shrink-0 flex-col">
                  <button
                    onClick={() => move(index, -1)}
                    disabled={index === 0}
                    className="rounded p-0.5 text-muted-foreground hover:bg-muted disabled:opacity-20"
                    title="上移"
                  >
                    <ChevronUp className="h-3.5 w-3.5" />
                  </button>
                  <button
                    onClick={() => move(index, 1)}
                    disabled={index === config.widgets.length - 1}
                    className="rounded p-0.5 text-muted-foreground hover:bg-muted disabled:opacity-20"
                    title="下移"
                  >
                    <ChevronDown className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
