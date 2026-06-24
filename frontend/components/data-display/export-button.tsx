'use client';

import { FileSpreadsheet, FileText, FileDown } from 'lucide-react';
import { Button, type ButtonProps } from '@/components/ui/button';

/* ------------------------------------------------------------------ */
/*  CSV 转义                                                           */
/* ------------------------------------------------------------------ */

function escapeCsv(val: unknown): string {
  const s = val === null || val === undefined ? '' : String(val);
  if (s.includes(',') || s.includes('"') || s.includes('\n')) {
    return `"${s.replace(/"/g, '""')}"`;
  }
  return s;
}

/* ------------------------------------------------------------------ */
/*  导出逻辑                                                           */
/* ------------------------------------------------------------------ */

export type ExportFormat = 'csv' | 'excel' | 'pdf';

export interface ExportButtonProps extends Omit<ButtonProps, 'onClick'> {
  /** 文件名前缀（不含扩展名） */
  filename?: string;
  /** 列头 */
  headers: string[];
  /** 行数据（每行是原始值的数组） */
  rows: (unknown[])[];
  /** 导出格式 */
  format?: ExportFormat;
}

function downloadCsv(headers: string[], rows: unknown[][], filename: string) {
  const csv = [
    headers.map(escapeCsv).join(','),
    ...rows.map((r) => r.map(escapeCsv).join(',')),
  ].join('\n');

  const BOM = '\uFEFF';
  const blob = new Blob([BOM + csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${filename}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

function downloadExcel(headers: string[], rows: unknown[][], filename: string) {
  // 简易 Excel 兼容：用 tab 分隔的 CSV（.xls），Excel 可直接打开
  const tsv = [
    headers.join('\t'),
    ...rows.map((r) => r.map((v) => (v === null || v === undefined ? '' : String(v))).join('\t')),
  ].join('\n');

  const BOM = '\uFEFF';
  const blob = new Blob([BOM + tsv], { type: 'application/vnd.ms-excel;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${filename}.xls`;
  a.click();
  URL.revokeObjectURL(url);
}

function downloadPdf(headers: string[], rows: unknown[][], filename: string) {
  // 简易 PDF 替代：生成可打印 HTML 并触发 window.print()
  // 对于完整 PDF 导出需要额外库，这里做降级处理
  const html = `
    <html><head><title>${filename}</title>
    <style>
      body { font-family: Arial, sans-serif; font-size: 12px; }
      table { border-collapse: collapse; width: 100%; }
      th, td { border: 1px solid #ccc; padding: 4px 8px; text-align: left; }
      th { background: #f5f5f5; font-weight: bold; }
    </style></head><body>
    <h2>${filename}</h2>
    <table>
      <thead><tr>${headers.map((h) => `<th>${h}</th>`).join('')}</tr></thead>
      <tbody>${rows.map((r) => `<tr>${r.map((c) => `<td>${c ?? ''}</td>`).join('')}</tr>`).join('')}</tbody>
    </table>
    <script>window.onload=function(){window.print();}<\/script>
    </body></html>`;

  const blob = new Blob([html], { type: 'text/html' });
  const url = URL.createObjectURL(blob);
  const w = window.open(url, '_blank');
  if (!w) {
    URL.revokeObjectURL(url);
  }
}

export function ExportButton({
  filename = 'export',
  headers,
  rows,
  format = 'csv',
  variant = 'outline',
  size = 'sm',
  children,
  ...rest
}: ExportButtonProps) {
  const handleClick = () => {
    if (rows.length === 0) return;
    switch (format) {
      case 'csv':
        downloadCsv(headers, rows, filename);
        break;
      case 'excel':
        downloadExcel(headers, rows, filename);
        break;
      case 'pdf':
        downloadPdf(headers, rows, filename);
        break;
    }
  };

  const icon =
    format === 'csv' ? (
      <FileText className="h-3.5 w-3.5" />
    ) : format === 'excel' ? (
      <FileSpreadsheet className="h-3.5 w-3.5" />
    ) : (
      <FileDown className="h-3.5 w-3.5" />
    );

  const label =
    format === 'csv' ? '导出 CSV' : format === 'excel' ? '导出 Excel' : '导出 PDF';

  return (
    <Button variant={variant} size={size} onClick={handleClick} {...rest}>
      {icon}
      {children ?? label}
    </Button>
  );
}
