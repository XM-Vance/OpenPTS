'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Search, Loader2, Building2, FileText, Calculator, UserCircle } from 'lucide-react';
import { apiClient } from '@/lib/api/client';

interface SearchResult {
  id: string;
  type: string;
  name: string;
  description?: string;
}

interface GroupedResults {
  [entityType: string]: SearchResult[];
}

const ENTITY_LABELS: Record<string, string> = {
  customer: '客户',
  contract: '合同',
  document: '文档',
  intent_customer: '意向客户',
  settlement: '结算',
  tag: '标签',
  station: '电站',
};

const ENTITY_ICONS: Record<string, React.ElementType> = {
  customer: Building2,
  contract: Calculator,
  document: FileText,
  intent_customer: UserCircle,
};

const ENTITY_ROUTES: Record<string, string> = {
  customer: '/customers',
  contract: '/contracts',
  document: '/documents',
  intent_customer: '/intent-customers',
  settlement: '/settlements',
  tag: '/system/tags',
  station: '/stations',
};

export function GlobalSearch() {
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<GroupedResults>({});
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Debounced search
  const doSearch = useCallback(async (q: string) => {
    if (!q.trim()) {
      setResults({});
      setOpen(false);
      return;
    }
    setLoading(true);
    try {
      const res = await apiClient.get<{ items?: SearchResult[]; results?: SearchResult[] }>('/search', {
        params: { q },
      });
      const data = res.data;
      const items: SearchResult[] = data.items ?? data.results ?? [];

      // Group by entity type
      const grouped: GroupedResults = {};
      for (const item of items) {
        const key = item.type ?? 'other';
        if (!grouped[key]) grouped[key] = [];
        grouped[key].push(item);
      }
      setResults(grouped);
      setOpen(true);
    } catch {
      setResults({});
    } finally {
      setLoading(false);
    }
  }, []);

  const handleInputChange = (value: string) => {
    setQuery(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => doSearch(value), 300);
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Keyboard shortcut: Ctrl/Cmd+K to focus
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        inputRef.current?.focus();
      }
      if (e.key === 'Escape') {
        setOpen(false);
        inputRef.current?.blur();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleSelect = (item: SearchResult) => {
    setOpen(false);
    setQuery('');
    const basePath = ENTITY_ROUTES[item.type] ?? `/${item.type}s`;
    router.push(`${basePath}/${item.id}`);
  };

  const totalResults = Object.values(results).reduce((sum, arr) => sum + arr.length, 0);

  return (
    <div ref={containerRef} className="relative">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => handleInputChange(e.target.value)}
          onFocus={() => {
            if (totalResults > 0) setOpen(true);
          }}
          placeholder="搜索客户、合同、文档… (⌘K)"
          className="h-9 w-[240px] rounded-md border border-input bg-transparent pl-9 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring lg:w-[320px]"
        />
        {loading && (
          <Loader2 className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin text-muted-foreground" />
        )}
      </div>

      {/* Dropdown */}
      {open && totalResults > 0 && (
        <div className="absolute left-0 top-full z-50 mt-1 w-full min-w-[320px] rounded-lg border bg-card shadow-lg lg:min-w-[400px]">
          <div className="max-h-[400px] overflow-y-auto p-2">
            {Object.entries(results).map(([entityType, items]) => {
              const Icon = ENTITY_ICONS[entityType] ?? Search;
              return (
                <div key={entityType}>
                  <div className="flex items-center gap-1.5 px-2 py-1.5 text-xs font-medium text-muted-foreground">
                    <Icon className="h-3.5 w-3.5" />
                    {ENTITY_LABELS[entityType] ?? entityType}
                    <span className="ml-auto rounded-full bg-gray-100 px-1.5 py-0.5 text-xs">
                      {items.length}
                    </span>
                  </div>
                  {items.slice(0, 5).map((item) => (
                    <button
                      key={item.id}
                      onClick={() => handleSelect(item)}
                      className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-accent"
                    >
                      <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
                      <div className="min-w-0 flex-1">
                        <div className="truncate font-medium">{item.name}</div>
                        {item.description && (
                          <div className="truncate text-xs text-muted-foreground">
                            {item.description}
                          </div>
                        )}
                      </div>
                    </button>
                  ))}
                  {items.length > 5 && (
                    <div className="px-2 py-1 text-xs text-muted-foreground">
                      还有 {items.length - 5} 条结果…
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* No results */}
      {open && query.trim() && !loading && totalResults === 0 && (
        <div className="absolute left-0 top-full z-50 mt-1 w-full min-w-[320px] rounded-lg border bg-card p-4 text-center text-sm text-muted-foreground shadow-lg">
          未找到「{query}」相关结果
        </div>
      )}
    </div>
  );
}
