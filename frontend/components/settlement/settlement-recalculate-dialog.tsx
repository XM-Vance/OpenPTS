'use client';

import React, { useState } from 'react';
import {
  Dialog,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { RefreshCw, AlertCircle, CheckCircle2 } from 'lucide-react';
import { generateSettlementDemoData } from '@/lib/api/settlement';

export interface SettlementRecalculateDialogProps {
  /** 是否显示 */
  open: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 重算完成回调 */
  onSuccess?: (days: number) => void;
  /** 自定义重算函数 */
  onRecalculate?: (params: RecalculateParams) => Promise<RecalculateResult>;
  /** 容器 className */
  className?: string;
}

export interface RecalculateParams {
  start_date: string;
  end_date: string;
  version: string;
}

export interface RecalculateResult {
  success: boolean;
  message: string;
  days_processed?: number;
}

/**
 * 结算重算弹窗
 * - 选择日期范围和版本
 * - 触发重算
 * - 展示结果
 */
export function SettlementRecalculateDialog({
  open,
  onClose,
  onSuccess,
  onRecalculate,
  className,
}: SettlementRecalculateDialogProps) {
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [version, setVersion] = useState('PRELIMINARY');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<RecalculateResult | null>(null);

  const handleClose = () => {
    setError(null);
    setResult(null);
    setStartDate('');
    setEndDate('');
    setVersion('PRELIMINARY');
    onClose();
  };

  const handleRecalculate = async () => {
    if (!startDate || !endDate) {
      setError('请选择开始和结束日期');
      return;
    }
    if (startDate > endDate) {
      setError('开始日期不能晚于结束日期');
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      let res: RecalculateResult;
      if (onRecalculate) {
        res = await onRecalculate({ start_date: startDate, end_date: endDate, version });
      } else {
        // 默认行为：调用 demo 数据生成
        const daysDiff = Math.ceil(
          (new Date(endDate).getTime() - new Date(startDate).getTime()) / (1000 * 60 * 60 * 24) + 1,
        );
        await generateSettlementDemoData(daysDiff);
        res = { success: true, message: `成功生成 ${daysDiff} 天结算数据`, days_processed: daysDiff };
      }
      setResult(res);
      if (res.success) {
        onSuccess?.(res.days_processed ?? 0);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '重算失败，请重试';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} className={cn('max-w-md', className)}>
      <DialogHeader>
        <DialogTitle>结算重算</DialogTitle>
        <DialogDescription>
          选择日期范围和版本进行结算重算
        </DialogDescription>
      </DialogHeader>

      {/* 错误提示 */}
      {error && (
        <div className="mb-4 flex items-center gap-2 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          <AlertCircle className="h-4 w-4 shrink-0" />
          {error}
        </div>
      )}

      {/* 结果提示 */}
      {result && (
        <div className={cn(
          'mb-4 flex items-center gap-2 rounded-md border px-3 py-2 text-sm',
          result.success
            ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
            : 'border-red-200 bg-red-50 text-red-700',
        )}>
          {result.success ? <CheckCircle2 className="h-4 w-4 shrink-0" /> : <AlertCircle className="h-4 w-4 shrink-0" />}
          {result.message}
        </div>
      )}

      {!result && (
        <div className="space-y-4">
          {/* 日期范围 */}
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="start_date">开始日期</Label>
              <Input
                id="start_date"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="end_date">结束日期</Label>
              <Input
                id="end_date"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
              />
            </div>
          </div>

          {/* 版本选择 */}
          <div className="space-y-1.5">
            <Label>结算版本</Label>
            <div className="flex gap-2">
              {['PRELIMINARY', 'FINAL'].map((v) => (
                <Badge
                  key={v}
                  variant={version === v ? 'default' : 'outline'}
                  className="cursor-pointer"
                  onClick={() => setVersion(v)}
                >
                  {v === 'PRELIMINARY' ? '初步结算' : '正式结算'}
                </Badge>
              ))}
            </div>
          </div>
        </div>
      )}

      <DialogFooter>
        <Button variant="outline" onClick={handleClose}>
          {result ? '关闭' : '取消'}
        </Button>
        {!result && (
          <Button onClick={handleRecalculate} disabled={loading}>
            {loading ? (
              <RefreshCw className="mr-1 h-3.5 w-3.5 animate-spin" />
            ) : (
              <RefreshCw className="mr-1 h-3.5 w-3.5" />
            )}
            {loading ? '重算中...' : '开始重算'}
          </Button>
        )}
      </DialogFooter>
    </Dialog>
  );
}
