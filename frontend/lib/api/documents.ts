import { apiClient } from './client';

// 文档解析管线：上传(存原件) → 异步解析(存解析件) → 结构化提取 → 人工确认入库。

export interface DocumentItem {
  id: string;
  org_id: string | null;
  filename: string;
  content_type: string;
  size: number;
  sha256: string | null;
  source_kind: string; // pdf/image/word/excel/csv
  original_object_key: string | null;
  parsed_object_key: string | null;
  doc_type: string | null;
  status: 'uploaded' | 'parsing' | 'parsed' | 'failed';
  page_count: number;
  summary: string | null;
  error: string | null;
  uploaded_by: string | null;
  auto_applied: boolean;
  customer_id: string | null;
  contract_id: string | null;
  intent_customer_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface DocumentDetail extends DocumentItem {
  text_content?: string | null;
  tables?: unknown;
  entities?: Record<string, string[]> | null;
}

export interface DocumentExtraction {
  id: number;
  document_id: string;
  group_no: number;
  field_key: string;
  field_label: string;
  value_text: string | null;
  value_num: number | null;
  value_date: string | null;
  unit: string | null;
  confidence: number | null;
  source: string; // glm/regex/excel
}

export interface DocumentApply {
  id: number;
  document_id: string;
  target: string;
  applied_rows: number;
  detail?: unknown;
  applied_by: string | null;
  applied_at: string;
}

export interface DocumentGetResponse {
  document: DocumentDetail;
  extractions: DocumentExtraction[];
  applies: DocumentApply[];
  suggest_target: string;
  storage_enabled: boolean;
}

export async function listDocuments(params: {
  status?: string;
  doc_type?: string;
  limit?: number;
  scope?: string;
} = {}): Promise<DocumentItem[]> {
  const { data } = await apiClient.get<{ items: DocumentItem[] }>('/api/v1/documents', { params });
  return data.items;
}

export async function getDocument(id: string): Promise<DocumentGetResponse> {
  const { data } = await apiClient.get<DocumentGetResponse>(`/api/v1/documents/${id}`);
  return data;
}

// 上传即返回（异步解析），不再长等 OCR。
// onProgress 回调接收 0–100 的进度百分比。
export async function uploadDocument(
  file: File,
  onProgress?: (percent: number) => void,
): Promise<{ document: DocumentItem; duplicated?: boolean; message: string }> {
  const form = new FormData();
  form.append('file', file);
  const { data } = await apiClient.post('/api/v1/documents', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120_000,
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded / e.total) * 100));
      }
    },
  });
  return data;
}

// 经网关流式下载（不走 MinIO presigned：内网地址浏览器不可达）。
export async function downloadDocumentFile(
  id: string,
  kind: 'original' | 'parsed',
  fallbackName: string,
): Promise<void> {
  const resp = await apiClient.get(`/api/v1/documents/${id}/${kind}`, {
    responseType: 'blob',
    timeout: 120_000,
  });
  // 文件名优先取响应头（RFC 5987），取不到用调用方给的
  let filename = fallbackName;
  const cd: string = resp.headers['content-disposition'] ?? '';
  const m = cd.match(/filename\*=UTF-8''([^;]+)/);
  if (m) filename = decodeURIComponent(m[1]);
  const url = URL.createObjectURL(resp.data as Blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export async function addExtraction(
  docId: string,
  input: { group_no?: number; field_key?: string; field_label: string; value_text: string; unit?: string },
): Promise<DocumentExtraction> {
  const { data } = await apiClient.post(`/api/v1/documents/${docId}/extractions`, input);
  return data.extraction;
}

export async function deleteExtraction(docId: string, extId: number): Promise<void> {
  await apiClient.delete(`/api/v1/documents/${docId}/extractions/${extId}`);
}

export async function updateExtraction(
  docId: string,
  extId: number,
  valueText: string,
): Promise<void> {
  await apiClient.put(`/api/v1/documents/${docId}/extractions/${extId}`, { value_text: valueText });
}

export async function reparseDocument(id: string): Promise<void> {
  await apiClient.post(`/api/v1/documents/${id}/reparse`);
}

export async function applyDocument(
  id: string,
  target: string,
): Promise<{ applied_rows: number; message: string; detail?: unknown }> {
  const { data } = await apiClient.post(`/api/v1/documents/${id}/apply`, { target });
  return data;
}

export async function deleteDocument(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/documents/${id}`);
}
