'use client';

import React, { useMemo, useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { ArrowUpDown, ArrowUp, ArrowDown, ChevronLeft, ChevronRight } from 'lucide-react';

// ───── 泛型行类型 ─────
export type DataRow = Record<string, any>;

// ───── 列定义 ─────
export interface DataTableColumn<T extends DataRow = DataRow> {
  /** 列标识 */
  key: string;
  /** 表头标题 */
  header: string;
  /** 是否可排序 */
  sortable?: boolean;
  /** 自定义渲染 */
  render?: (row: T, index: number) => React.ReactNode;
  /** 列宽 */
  width?: number | string;
  /** 对齐 */
  align?: 'left' | 'center' | 'right';
}

export interface DataTableProps<T extends DataRow = DataRow> {
  /** 数据源 */
  data: T[];
  /** 列配置 */
  columns: DataTableColumn<T>[];
  /** 行唯一标识字段或函数 */
  rowKey: string | ((row: T) => string);
  /** 是否显示序号列 */
  showIndex?: boolean;
  /** 是否支持行选择 */
  selectable?: boolean;
  /** 已选中的行 key 集合 */
  selectedKeys?: Set<string>;
  /** 行选择变更回调 */
  onSelectionChange?: (keys: Set<string>) => void;
  /** 分页大小（0 或 undefined 不分页） */
  pageSize?: number;
  /** 是否显示分页控件 */
  showPagination?: boolean;
  /** 空数据提示 */
  emptyText?: string;
  /** 行点击回调 */
  onRowClick?: (row: T, index: number) => void;
  /** 容器 className */
  className?: string;
  /** 加载中 */
  loading?: boolean;
}

type SortDir = 'asc' | 'desc' | null;

/**
 * 数据表格组件
 * - 基于 shadcn Table
 * - 支持排序、分页、行选择
 */
export function DataTable<T extends DataRow = DataRow>({
  data,
  columns,
  rowKey,
  showIndex = false,
  selectable = false,
  selectedKeys,
  onSelectionChange,
  pageSize = 0,
  showPagination = true,
  emptyText = '暂无数据',
  onRowClick,
  className,
  loading = false,
}: DataTableProps<T>) {
  // ──── 排序状态 ────
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDir>(null);

  // ──── 分页状态 ────
  const [page, setPage] = useState(1);
  const effectivePageSize = pageSize > 0 ? pageSize : 10;

  // 获取行 key
  const getKey = (row: T): string =>
    typeof rowKey === 'function' ? rowKey(row) : String(row[rowKey]);

  // 排序后数据
  const sorted = useMemo(() => {
    if (!sortKey || !sortDir) return data;
    return [...data].sort((a, b) => {
      const va = a[sortKey];
      const vb = b[sortKey];
      if (va == null) return 1;
      if (vb == null) return -1;
      const cmp = typeof va === 'number' && typeof vb === 'number'
        ? va - vb
        : String(va).localeCompare(String(vb));
      return sortDir === 'asc' ? cmp : -cmp;
    });
  }, [data, sortKey, sortDir]);

  // 分页后数据
  const totalPages = Math.max(1, Math.ceil(sorted.length / effectivePageSize));
  const paged = showPagination && effectivePageSize > 0
    ? sorted.slice((page - 1) * effectivePageSize, page * effectivePageSize)
    : sorted;

  // 排序切换
  const handleSort = (key: string) => {
    setSortKey((prev) => {
      if (prev !== key) { setSortDir('asc'); return key; }
      setSortDir((d) => (d === 'asc' ? 'desc' : d === 'desc' ? null : 'asc'));
      return prev;
    });
    if (sortKey === key && sortDir === 'desc') setSortKey(null);
  };

  // 全选
  const allKeys = new Set(data.map(getKey));
  const isAllSelected = selectedKeys != null && selectedKeys.size > 0 && selectedKeys.size === data.length;
  const isIndeterminate = selectedKeys != null && selectedKeys.size > 0 && selectedKeys.size < data.length;

  const toggleAll = () => {
    if (!onSelectionChange) return;
    onSelectionChange(isAllSelected ? new Set() : allKeys);
  };

  const toggleRow = (key: string) => {
    if (!onSelectionChange || !selectedKeys) return;
    const next = new Set(selectedKeys);
    if (next.has(key)) next.delete(key); else next.add(key);
    onSelectionChange(next);
  };

  // 排序图标
  const SortIcon = ({ colKey }: { colKey: string }) => {
    if (sortKey !== colKey || !sortDir) return <ArrowUpDown className="ml-1 h-3 w-3" />;
    return sortDir === 'asc'
      ? <ArrowUp className="ml-1 h-3 w-3" />
      : <ArrowDown className="ml-1 h-3 w-3" />;
  };

  if (loading) {
    return (
      <div className={cn('flex items-center justify-center py-12 text-muted-foreground', className)}>
        <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        <span className="ml-2">加载中...</span>
      </div>
    );
  }

  return (
    <div className={cn('space-y-3', className)}>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {selectable && (
                <TableHead className="w-10">
                  <input
                    type="checkbox"
                    checked={isAllSelected}
                    ref={(el) => {
                      if (el) el.indeterminate = isIndeterminate;
                    }}
                    onChange={toggleAll}
                    className="h-4 w-4 rounded border-gray-300"
                  />
                </TableHead>
              )}
              {showIndex && <TableHead className="w-12">#</TableHead>}
              {columns.map((col) => (
                <TableHead
                  key={col.key}
                  style={{ width: col.width, textAlign: col.align }}
                  className={cn(col.sortable && 'cursor-pointer select-none')}
                  onClick={col.sortable ? () => handleSort(col.key) : undefined}
                >
                  <span className="inline-flex items-center">
                    {col.header}
                    {col.sortable && <SortIcon colKey={col.key} />}
                  </span>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {paged.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={(selectable ? 1 : 0) + (showIndex ? 1 : 0) + columns.length}
                  className="h-24 text-center text-muted-foreground"
                >
                  {emptyText}
                </TableCell>
              </TableRow>
            ) : (
              paged.map((row, idx) => {
                const key = getKey(row);
                const globalIdx = (page - 1) * effectivePageSize + idx;
                return (
                  <TableRow
                    key={key}
                    className={cn(onRowClick && 'cursor-pointer')}
                    onClick={() => onRowClick?.(row, globalIdx)}
                  >
                    {selectable && (
                      <TableCell onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={selectedKeys?.has(key) ?? false}
                          onChange={() => toggleRow(key)}
                          className="h-4 w-4 rounded border-gray-300"
                        />
                      </TableCell>
                    )}
                    {showIndex && (
                      <TableCell className="text-muted-foreground">
                        {globalIdx + 1}
                      </TableCell>
                    )}
                    {columns.map((col) => (
                      <TableCell key={col.key} style={{ textAlign: col.align }}>
                        {col.render ? col.render(row, globalIdx) : String(row[col.key] ?? '')}
                      </TableCell>
                    ))}
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>

      {/* 分页控件 */}
      {showPagination && totalPages > 1 && (
        <div className="flex items-center justify-between px-1">
          <p className="text-sm text-muted-foreground">
            共 {sorted.length} 条，第 {page}/{totalPages} 页
          </p>
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              disabled={page <= 1}
              onClick={() => setPage(1)}
            >
              <ChevronLeft className="h-4 w-4" />
              <ChevronLeft className="h-4 w-4 -ml-3" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              disabled={page >= totalPages}
              onClick={() => setPage(totalPages)}
            >
              <ChevronRight className="h-4 w-4" />
              <ChevronRight className="h-4 w-4 -ml-3" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
