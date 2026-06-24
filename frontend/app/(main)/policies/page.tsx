'use client';

// 政策文件库：政策文件归纳为结构化条目（来源：文档解析「确认入库 → 政策文件」或手动新增）。
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { listPolicies, createPolicy, deletePolicy, type PolicyInput } from '@/lib/api/policy';
import { extractErrorMessage } from '@/lib/api/client';
import { ScrollText, Plus, Trash2, Loader2, FileScan } from 'lucide-react';

const SELECT_CLASS = 'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const CATEGORIES = ['市场规则', '补贴政策', '准入注册', '价格机制', '其他'];

const emptyForm: PolicyInput = { title: '', doc_no: '', category: '市场规则', effective_date: '', summary: '' };

export default function PoliciesPage() {
  const qc = useQueryClient();
  const [categoryFilter, setCategoryFilter] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<PolicyInput>(emptyForm);
  const [error, setError] = useState<string | null>(null);

  const { data: rows, isLoading } = useQuery({
    queryKey: ['policies', categoryFilter],
    queryFn: () => listPolicies({ category: categoryFilter || undefined }),
  });

  const createMut = useMutation({
    mutationFn: () => createPolicy(form),
    onSuccess: () => {
      setForm(emptyForm);
      setShowForm(false);
      setError(null);
      qc.invalidateQueries({ queryKey: ['policies'] });
    },
    onError: (e) => setError(extractErrorMessage(e)),
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => deletePolicy(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['policies'] }),
  });

  const set = (k: keyof PolicyInput, v: string) => setForm((p) => ({ ...p, [k]: v }));

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">政策文件库</h1>
          <p className="text-sm text-muted-foreground">
            政策文件结构化归档 —— 文档解析时把政策文件「确认入库 → 政策文件」即自动归纳到此，也可手动新增
          </p>
        </div>
        <Button onClick={() => { setShowForm((v) => !v); setError(null); }}>
          <Plus className="mr-1 h-4 w-4" />
          新增政策
        </Button>
      </div>

      {/* 新增表单 */}
      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">新增政策条目</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            <div className="grid gap-3 md:grid-cols-2">
              <div className="space-y-1.5 md:col-span-2">
                <Label htmlFor="p_title">政策标题 *</Label>
                <Input id="p_title" value={form.title} onChange={(e) => set('title', e.target.value)} placeholder="如：福建电力市场化交易实施细则" />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="p_no">文号</Label>
                <Input id="p_no" value={form.doc_no} onChange={(e) => set('doc_no', e.target.value)} placeholder="如：闽电交〔2026〕12号" />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="p_cat">分类</Label>
                <select id="p_cat" value={form.category} onChange={(e) => set('category', e.target.value)} className={SELECT_CLASS + ' w-full'}>
                  {CATEGORIES.map((c) => (
                    <option key={c} value={c}>{c}</option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="p_date">生效日期</Label>
                <Input id="p_date" type="date" value={form.effective_date} onChange={(e) => set('effective_date', e.target.value)} />
              </div>
              <div className="space-y-1.5 md:col-span-2">
                <Label htmlFor="p_sum">要点摘要</Label>
                <Input id="p_sum" value={form.summary} onChange={(e) => set('summary', e.target.value)} placeholder="政策要点（可选）" />
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => { setShowForm(false); setError(null); }}>取消</Button>
              <Button onClick={() => createMut.mutate()} disabled={createMut.isPending || !form.title.trim()}>
                {createMut.isPending ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : null}
                保存
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 筛选 + 列表 */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            <ScrollText className="h-4 w-4" />
            政策文件
          </CardTitle>
          <select value={categoryFilter} onChange={(e) => setCategoryFilter(e.target.value)} className={SELECT_CLASS}>
            <option value="">全部分类</option>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>标题</TableHead>
                <TableHead>文号</TableHead>
                <TableHead>分类</TableHead>
                <TableHead>生效日期</TableHead>
                <TableHead>来源</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    <Loader2 className="mx-auto h-4 w-4 animate-spin" />
                  </TableCell>
                </TableRow>
              ) : (rows ?? []).length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    暂无政策文件，可手动新增或在文档解析中把政策文件「确认入库 → 政策文件」
                  </TableCell>
                </TableRow>
              ) : (
                (rows ?? []).map((p) => (
                  <TableRow key={p.id}>
                    <TableCell>
                      <div className="font-medium">{p.title}</div>
                      {p.summary && <div className="text-xs text-muted-foreground line-clamp-1">{p.summary}</div>}
                    </TableCell>
                    <TableCell>{p.doc_no ?? '-'}</TableCell>
                    <TableCell>{p.category ? <Badge variant="secondary">{p.category}</Badge> : '-'}</TableCell>
                    <TableCell>{p.effective_date ?? '-'}</TableCell>
                    <TableCell>
                      {p.source === 'document' ? (
                        <span className="flex items-center gap-1 text-xs text-muted-foreground" title={p.document_name ?? ''}>
                          <FileScan className="h-3 w-3" />文档解析
                        </span>
                      ) : (
                        <span className="text-xs text-muted-foreground">手动</span>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => { if (window.confirm(`删除政策「${p.title}」？`)) deleteMut.mutate(p.id); }}
                      >
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
