'use client';

import React, { useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import { Search, RotateCcw } from 'lucide-react';

// ───── 筛选字段配置 ─────
export interface FilterField {
  /** 字段类型 */
  type: 'text' | 'select' | 'date' | 'date-range';
  /** 字段唯一标识 */
  key: string;
  /** 字段标签 */
  label: string;
  /** 下拉选项（select 类型） */
  options?: { value: string; label: string }[];
  /** 占位文本 */
  placeholder?: string;
  /** 默认值 */
  defaultValue?: unknown;
}

export interface FilterBarProps {
  /** 筛选字段配置 */
  fields: FilterField[];
  /** 当前值 */
  values: Record<string, any>;
  /** 值变更回调 */
  onChange: (values: Record<string, any>) => void;
  /** 搜索按钮回调 */
  onSearch?: () => void;
  /** 重置按钮回调 */
  onReset?: () => void;
  /** 右侧操作区（如"新增"按钮） */
  action?: React.ReactNode;
  /** 容器 className */
  className?: string;
}

/**
 * 通用筛选栏
 * - 水平排列筛选字段
 * - 内置搜索和重置按钮
 * - 右侧 action 区域
 */
export function FilterBar({
  fields,
  values,
  onChange,
  onSearch,
  onReset,
  action,
  className,
}: FilterBarProps) {
  const handleChange = useCallback(
    (key: string, value: string) => {
      onChange({ ...values, [key]: value });
    },
    [values, onChange],
  );

  const renderField = (field: FilterField) => {
    switch (field.type) {
      case 'text':
        return (
          <Input
            key={field.key}
            placeholder={field.placeholder || field.label}
            value={values[field.key] ?? ''}
            onChange={(e) => handleChange(field.key, e.target.value)}
            className="h-8 w-[200px]"
          />
        );

      case 'select':
        return (
          <select
            key={field.key}
            value={values[field.key] ?? ''}
            onChange={(e) => handleChange(field.key, e.target.value)}
            className="h-8 w-[200px] rounded-md border border-input bg-background px-3 text-sm"
          >
            <option value="">全部</option>
            {field.options?.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        );

      case 'date':
        return (
          <Input
            key={field.key}
            type="date"
            value={values[field.key] ?? ''}
            onChange={(e) => handleChange(field.key, e.target.value)}
            className="h-8 w-[200px]"
          />
        );

      case 'date-range':
        return (
          <div key={field.key} className="flex items-center gap-1">
            <Input
              type="date"
              value={values[`${field.key}_start`] ?? ''}
              onChange={(e) => handleChange(`${field.key}_start`, e.target.value)}
              className="h-8 w-[170px]"
            />
            <span className="text-muted-foreground">~</span>
            <Input
              type="date"
              value={values[`${field.key}_end`] ?? ''}
              onChange={(e) => handleChange(`${field.key}_end`, e.target.value)}
              className="h-8 w-[170px]"
            />
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div
      className={cn(
        'min-h-12 border-b bg-slate-50 px-4 py-2 flex items-center gap-3 flex-wrap',
        className,
      )}
    >
      {/* 筛选字段 */}
      <div className="flex flex-1 flex-wrap items-center gap-3">
        {fields.map(renderField)}

        {onSearch && (
          <Button variant="default" size="sm" onClick={onSearch}>
            <Search className="mr-1 h-3.5 w-3.5" />
            搜索
          </Button>
        )}

        {onReset && (
          <Button variant="outline" size="sm" onClick={onReset}>
            <RotateCcw className="mr-1 h-3.5 w-3.5" />
            重置
          </Button>
        )}
      </div>

      {/* 右侧操作 */}
      {action && <div className="shrink-0">{action}</div>}
    </div>
  );
}
