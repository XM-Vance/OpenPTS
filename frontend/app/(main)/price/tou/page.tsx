'use client';

import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { genTOUDemo, listTOURules, type TOURule } from '@/lib/api/tou';

const TAG_COLOR: Record<string, string> = {
  peak: '#f59e0b',
  sharp: '#ef4444',
  shoulder: '#3b82f6',
  valley: '#10b981',
};
const TAG_LABEL: Record<string, string> = {
  peak: '峰',
  sharp: '尖峰',
  shoulder: '平',
  valley: '谷',
};

// 24-hour TOU visualization data (one bar per hour)
function generate24HourTOUData(tags: string[]) {
  // tags is 96 entries (15-min each), collapse to 24 hours by majority
  const hourData: { hour: number; label: string; tag: string; color: string; tagLabel: string }[] = [];
  for (let h = 0; h < 24; h++) {
    const start = h * 4;
    const end = start + 4;
    const hourTags = tags.slice(start, end);
    const counts: Record<string, number> = {};
    for (const t of hourTags) {
      counts[t] = (counts[t] || 0) + 1;
    }
    const majorTag = Object.entries(counts).sort((a, b) => b[1] - a[1])[0]?.[0] ?? 'shoulder';
    hourData.push({
      hour: h,
      label: `${String(h).padStart(2, '0')}:00`,
      tag: majorTag,
      color: TAG_COLOR[majorTag] ?? '#94a3b8',
      tagLabel: TAG_LABEL[majorTag] ?? majorTag,
    });
  }
  return hourData;
}

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const TouBar = dynamic(
  () => import('./_tou-bar').then((m) => ({ default: m.TouBar })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);

export default function TOURulesPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('price_management:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading } = useQuery({ queryKey: ['tou-rules'], queryFn: listTOURules });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genTOUDemo();
      qc.invalidateQueries({ queryKey: ['tou-rules'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const rules = data?.items ?? [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">TOU 时段规则</h1>
          <p className="text-sm text-muted-foreground">
            分时电价规则定义 · 24 小时时段可视化 · 规则变更时间线
          </p>
        </div>
        {canWrite && (
          <Button variant="outline" onClick={onGen} disabled={busy}>
            {busy ? '生成中...' : '生成演示规则'}
          </Button>
        )}
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 24 小时分时电价时段可视化 */}
      {rules.length > 0 && (
        <ChartContainer title="24 小时分时电价时段可视化">
          {rules.map((r) => {
            const tags = r.periods?.tags ?? [];
            const hourData = generate24HourTOUData(tags);
            const barData = hourData.map((d) => ({
              ...d,
              value: 1,
            }));
            return (
              <div key={r.id} className="mb-4">
                <p className="mb-2 text-sm font-medium">{r.rule_name}（{r.effective_from.slice(0, 10)} ~ {r.effective_to?.slice(0, 10) ?? '至今'}）</p>
                <div
                  className="[&_.recharts-surface:focus]:outline-none"
                  style={{ width: '100%', height: 80 }}
                >
                  <TouBar data={barData} />
                </div>
                <div className="mt-2 flex gap-3 text-xs">
                  <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.peak }} />峰</span>
                  <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.shoulder }} />平</span>
                  <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.valley }} />谷</span>
                  <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.sharp }} />尖峰</span>
                </div>
              </div>
            );
          })}
        </ChartContainer>
      )}

      {/* 96 段时段图谱 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">96 段时段图谱</CardTitle>
        </CardHeader>
        <CardContent>
          {rules.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              暂无 TOU 规则{canWrite && '，可点右上「生成演示规则」'}
            </p>
          ) : (
            rules.map((r) => <RuleBar key={r.id} rule={r} />)
          )}
        </CardContent>
      </Card>

      {/* TOU 规则变更时间线 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">规则变更时间线</CardTitle>
        </CardHeader>
        <CardContent>
          {rules.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无规则变更记录</p>
          ) : (
            <div className="space-y-4">
              {rules.map((r) => {
                const tags = r.periods?.tags ?? [];
                const counts = tags.reduce(
                  (acc, t) => {
                    if (t === 'peak') acc.peak += 1;
                    else if (t === 'valley') acc.valley += 1;
                    else acc.shoulder += 1;
                    return acc;
                  },
                  { peak: 0, valley: 0, shoulder: 0 },
                );
                return (
                  <div key={r.id} className="flex items-start gap-4">
                    <div className="flex flex-col items-center">
                      <div className="h-3 w-3 rounded-full bg-primary" />
                      <div className="w-px flex-1 bg-border" />
                    </div>
                    <div className="flex-1 pb-4">
                      <p className="font-medium text-sm">{r.rule_name}</p>
                      <p className="text-xs text-muted-foreground">
                        生效：{r.effective_from.slice(0, 10)} ~ {r.effective_to?.slice(0, 10) ?? '至今'}
                      </p>
                      <div className="mt-1 flex gap-2">
                        <Badge variant="default">峰 {(counts.peak / 4).toFixed(1)}h</Badge>
                        <Badge variant="secondary">平 {(counts.shoulder / 4).toFixed(1)}h</Badge>
                        <Badge variant="success">谷 {(counts.valley / 4).toFixed(1)}h</Badge>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 规则明细表格 */}
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>规则名</TableHead>
              <TableHead>生效起</TableHead>
              <TableHead>生效止</TableHead>
              <TableHead>峰段时长</TableHead>
              <TableHead>谷段时长</TableHead>
              <TableHead>平段时长</TableHead>
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
            {rules.map((r) => {
              const tags = r.periods?.tags ?? [];
              const counts = tags.reduce(
                (acc, t) => {
                  if (t === 'peak') acc.peak += 1;
                  else if (t === 'valley') acc.valley += 1;
                  else acc.shoulder += 1;
                  return acc;
                },
                { peak: 0, valley: 0, shoulder: 0 },
              );
              const h = (n: number) => `${(n / 4).toFixed(1)} h`;
              return (
                <TableRow key={r.id}>
                  <TableCell className="font-medium">{r.rule_name}</TableCell>
                  <TableCell>{r.effective_from.slice(0, 10)}</TableCell>
                  <TableCell>{r.effective_to?.slice(0, 10) ?? '-'}</TableCell>
                  <TableCell>
                    <Badge variant="default">{h(counts.peak)}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant="success">{h(counts.valley)}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{h(counts.shoulder)}</Badge>
                  </TableCell>
                </TableRow>
              );
            })}
            {rules.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  暂无规则
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

function RuleBar({ rule }: { rule: TOURule }) {
  const tags = rule.periods?.tags ?? [];
  return (
    <div className="mb-4">
      <p className="mb-2 text-sm font-medium">{rule.rule_name}</p>
      <div className="flex h-6 overflow-hidden rounded border border-border">
        {tags.map((tag, i) => (
          <div
            key={i}
            style={{ width: `${100 / 96}%`, backgroundColor: TAG_COLOR[tag] ?? '#94a3b8' }}
            title={`${String(Math.floor(i / 4)).padStart(2, '0')}:${(i % 4) * 15 || '00'} ${TAG_LABEL[tag] ?? tag}`}
          />
        ))}
      </div>
      <div className="mt-1 grid grid-cols-24 text-[10px] text-muted-foreground" style={{ gridTemplateColumns: 'repeat(24, 1fr)' }}>
        {Array.from({ length: 24 }).map((_, h) => (
          <div key={h} className="text-center">
            {h}
          </div>
        ))}
      </div>
      <div className="mt-2 flex gap-3 text-xs">
        <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.peak }} />峰</span>
        <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.shoulder }} />平</span>
        <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.valley }} />谷</span>
        <span><span className="mr-1 inline-block h-3 w-3 rounded-sm" style={{ background: TAG_COLOR.sharp }} />尖峰</span>
      </div>
    </div>
  );
}
