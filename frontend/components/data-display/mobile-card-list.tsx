'use client';

import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';

/** 通用数据行类型（与 DataTable 的 DataRow 对齐）。 */
export type MobileRow = Record<string, any>;

/**
 * 移动端卡片字段定义。
 * - primary=true 的字段合并成卡片标题行（加粗），其余作为标签:值 行平铺。
 * - render 优先于 key 取值；用于复用 DataTable 列定义里的自定义渲染。
 */
export interface MobileCardField<T extends MobileRow = MobileRow> {
  /** 取值字段名（render 缺省时用） */
  key?: string;
  /** 标签（primary 字段可省略） */
  label?: string;
  /** 自定义渲染（与 DataTableColumn.render 同源，便于复用列定义） */
  render?: (row: T, index: number) => React.ReactNode;
  /** 作为标题行突出显示 */
  primary?: boolean;
  /** 强调值（如涨跌额/收盘价），右对齐、稍大字号 */
  emphasize?: boolean;
}

export interface MobileCardListProps<T extends MobileRow = MobileRow> {
  /** 数据源 */
  items: T[];
  /** 行唯一标识 */
  itemKey: string | ((row: T) => string);
  /** 字段配置（顺序即展示顺序） */
  fields: MobileCardField<T>[];
  /** 整卡点击 */
  onItemClick?: (row: T, index: number) => void;
  /** 空数据提示 */
  emptyText?: string;
  /**
   * 卡片底部操作区（HIGH 页用：主按钮 + 「⋯」菜单收纳次要动作）。
   * 整卡点击与 actions 互不干扰——actions 区点击不冒泡到 onItemClick。
   * 缺省时不渲染底部，已有卡片零破坏。
   */
  actions?: (row: T, index: number) => React.ReactNode;
}

/**
 * 移动端卡片列表：把表格行渲染为竖排卡片。
 * 替代窄屏下溢出的宽表，每行一张卡，关键字段突出、次要字段平铺。
 *
 * 设计与 StatCard 视觉语言一致（Card + CardContent p-4）。
 * 桌面端不应使用本组件（调用方按 useIsMobile() 切换）。
 */
export function MobileCardList<T extends MobileRow = MobileRow>({
  items,
  itemKey,
  fields,
  onItemClick,
  emptyText = '暂无数据',
  actions,
}: MobileCardListProps<T>) {
  if (!items || items.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">{emptyText}</p>;
  }

  const resolveKey = (row: T, idx: number): string =>
    typeof itemKey === 'function' ? itemKey(row) : String(row[itemKey] ?? idx);

  const primaryFields = fields.filter((f) => f.primary);
  const bodyFields = fields.filter((f) => !f.primary);
  const emphasizeFields = bodyFields.filter((f) => f.emphasize);
  const normalFields = bodyFields.filter((f) => !f.emphasize);

  return (
    <div className="space-y-3">
      {items.map((row, idx) => {
        const renderField = (f: MobileCardField<T>): React.ReactNode =>
          f.render ? f.render(row, idx) : f.key != null ? String(row[f.key] ?? '—') : null;

        return (
          <Card
            key={resolveKey(row, idx)}
            className={cn(onItemClick && 'cursor-pointer transition-shadow hover:shadow-md')}
            onClick={onItemClick ? () => onItemClick(row, idx) : undefined}
          >
            <CardContent className="space-y-2 p-4">
              {/* 标题行：primary 字段用空格拼接 */}
              {primaryFields.length > 0 && (
                <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
                  {primaryFields.map((f, i) => (
                    <span key={i} className="text-sm font-semibold">
                      {renderField(f)}
                    </span>
                  ))}
                </div>
              )}
              {/* 强调值行（涨跌/收盘价等）：右对齐网格 */}
              {emphasizeFields.length > 0 && (
                <div className="grid grid-cols-2 gap-2">
                  {emphasizeFields.map((f, i) => (
                    <div key={i} className="text-right">
                      {f.label && (
                        <p className="text-xs text-muted-foreground">{f.label}</p>
                      )}
                      <p className="text-base font-medium">{renderField(f)}</p>
                    </div>
                  ))}
                </div>
              )}
              {/* 普通字段：label : value 两列 */}
              {normalFields.length > 0 && (
                <dl className="grid grid-cols-2 gap-x-3 gap-y-1.5 text-sm">
                  {normalFields.map((f, i) => (
                    <div key={i} className="flex flex-col">
                      {f.label && (
                        <dt className="text-xs text-muted-foreground">{f.label}</dt>
                      )}
                      <dd className="truncate">{renderField(f)}</dd>
                    </div>
                  ))}
                </dl>
              )}
              {/* 操作区（HIGH 页：主按钮 + 「⋯」菜单）。点击不冒泡到整卡 onItemClick。 */}
              {actions && (
                <div
                  className="flex items-center justify-end gap-1 border-t pt-2"
                  onClick={(e) => e.stopPropagation()}
                >
                  {actions(row, idx)}
                </div>
              )}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
