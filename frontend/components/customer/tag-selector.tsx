'use client';

import React, { useMemo, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { Plus, X, Settings, Search, Loader2 } from 'lucide-react';

export interface Tag {
  name: string;
  source?: string;
  expire?: string | null;
  reason?: string | null;
}

export interface TagOption {
  name: string;
  category?: string;
}

export interface TagSelectorProps {
  /** 已选标签 */
  tags: Tag[];
  /** 变更回调 */
  onChange: (tags: Tag[]) => void;
  /** 所有可选标签（外部提供或异步加载） */
  allTags?: TagOption[];
  /** 是否加载中 */
  loading?: boolean;
  /** 是否只读 */
  readonly?: boolean;
  /** 创建新标签回调（输入不存在标签时） */
  onCreateTag?: (name: string) => Promise<void>;
  /** 右上角管理按钮回调 */
  onManage?: () => void;
  /** className */
  className?: string;
}

/**
 * 标签选择器
 * - 展示已选标签（Badge）
 * - Popover 弹窗搜索选择
 * - 支持新建标签
 */
export function TagSelector({
  tags,
  onChange,
  allTags = [],
  loading = false,
  readonly = false,
  onCreateTag,
  onManage,
  className,
}: TagSelectorProps) {
  const [popoverOpen, setPopoverOpen] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [creating, setCreating] = useState(false);

  // 按分类分组
  const groupedTags = useMemo(() => {
    const groups: Record<string, TagOption[]> = {};
    allTags.forEach((tag) => {
      const category = tag.category || '其他';
      if (!groups[category]) groups[category] = [];
      groups[category].push(tag);
    });
    return Object.entries(groups).map(([category, items]) => ({ category, items }));
  }, [allTags]);

  // 搜索过滤
  const filteredGroups = useMemo(() => {
    if (!searchText.trim()) return groupedTags;
    const lower = searchText.toLowerCase();
    return groupedTags
      .map((g) => ({
        ...g,
        items: g.items.filter((t) => t.name.toLowerCase().includes(lower)),
      }))
      .filter((g) => g.items.length > 0);
  }, [groupedTags, searchText]);

  // 搜索文本是否已存在于标签库
  const tagExists = useMemo(() => {
    if (!searchText.trim()) return true;
    return allTags.some((t) => t.name.toLowerCase() === searchText.toLowerCase());
  }, [allTags, searchText]);

  const selectedNames = useMemo(() => new Set(tags.map((t) => t.name)), [tags]);

  const handleRemove = (name: string) => {
    onChange(tags.filter((t) => t.name !== name));
  };

  const handleAdd = (name: string) => {
    if (selectedNames.has(name)) return;
    onChange([...tags, { name, source: 'MANUAL' }]);
  };

  const handleCreateAndAdd = async () => {
    if (!searchText.trim() || tagExists) return;
    setCreating(true);
    try {
      if (onCreateTag) await onCreateTag(searchText.trim());
      handleAdd(searchText.trim());
      setSearchText('');
    } catch (err) {
      console.error('创建标签失败:', err);
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className={cn('space-y-2', className)}>
      {/* 已选标签 */}
      <div className="flex flex-wrap items-center gap-1.5">
        {tags.map((tag) => (
          <Badge
            key={tag.name}
            variant={tag.source === 'AUTO' ? 'secondary' : 'default'}
            className="gap-1 pr-1"
          >
            {tag.name}
            {!readonly && (
              <button
                type="button"
                onClick={() => handleRemove(tag.name)}
                className="ml-0.5 rounded-full p-0.5 hover:bg-black/10"
              >
                <X className="h-3 w-3" />
              </button>
            )}
          </Badge>
        ))}

        {!readonly && (
          <button
            type="button"
            onClick={() => setPopoverOpen(true)}
            className="inline-flex items-center gap-1 rounded-md border border-dashed border-muted-foreground/30 px-2 py-0.5 text-xs text-muted-foreground hover:border-primary hover:text-primary"
          >
            <Plus className="h-3 w-3" />
            添加
          </button>
        )}

        {!readonly && onManage && (
          <button
            type="button"
            onClick={onManage}
            className="inline-flex items-center gap-1 rounded-md border border-dashed border-muted-foreground/30 px-2 py-0.5 text-xs text-muted-foreground hover:border-primary hover:text-primary"
            title="管理标签"
          >
            <Settings className="h-3 w-3" />
          </button>
        )}
      </div>

      {/* 选择弹窗 */}
      {popoverOpen && (
        <div className="absolute z-50 mt-1 w-[300px] rounded-lg border bg-popover p-0 shadow-lg">
          {/* 搜索框 */}
          <div className="border-b p-2">
            <div className="relative">
              <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                placeholder="搜索或新建标签..."
                className="h-8 pl-7 text-sm"
                autoFocus
              />
            </div>
          </div>

          {/* 标签列表 */}
          <div className="max-h-[250px] overflow-y-auto p-1">
            {loading ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            ) : filteredGroups.length === 0 ? (
              <p className="py-3 text-center text-sm text-muted-foreground">
                未找到匹配的标签
              </p>
            ) : (
              filteredGroups.map((group) => (
                <div key={group.category}>
                  <p className="px-2 py-1 text-xs font-semibold text-muted-foreground bg-muted/50">
                    {group.category}
                  </p>
                  {group.items.map((tag) => {
                    const isSelected = selectedNames.has(tag.name);
                    return (
                      <button
                        key={tag.name}
                        type="button"
                        disabled={isSelected}
                        onClick={() => {
                          handleAdd(tag.name);
                          setPopoverOpen(false);
                        }}
                        className={cn(
                          'flex w-full items-center justify-between rounded px-2 py-1.5 text-sm',
                          isSelected
                            ? 'text-muted-foreground'
                            : 'hover:bg-accent',
                        )}
                      >
                        <span>{tag.name}</span>
                        {isSelected && (
                          <span className="text-xs text-muted-foreground">已添加</span>
                        )}
                      </button>
                    );
                  })}
                </div>
              ))
            )}
          </div>

          {/* 新建标签 */}
          {searchText.trim() && !tagExists && (
            <div className="border-t p-2">
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={handleCreateAndAdd}
                disabled={creating}
              >
                {creating ? (
                  <Loader2 className="mr-1 h-3 w-3 animate-spin" />
                ) : (
                  <Plus className="mr-1 h-3 w-3" />
                )}
                新建标签: &quot;{searchText.trim()}&quot;
              </Button>
            </div>
          )}

          {/* 关闭 */}
          <div className="border-t p-1">
            <Button
              variant="ghost"
              size="sm"
              className="w-full"
              onClick={() => { setPopoverOpen(false); setSearchText(''); }}
            >
              关闭
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
