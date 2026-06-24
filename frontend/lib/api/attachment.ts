import { apiClient } from './client';

export interface FileAttachment {
  id: string;
  resource: string;
  resource_id: string;
  filename: string;
  object_key: string;
  content_type: string;
  size: number;
  uploaded_by?: string | null;
  note?: string | null;
  created_at: string;
}

export async function listAttachments(resource: string, resourceId: string): Promise<{ items: FileAttachment[] }> {
  const { data } = await apiClient.get('/api/v1/attachments', {
    params: { resource, resource_id: resourceId },
  });
  return data;
}

export async function uploadAttachment(
  resource: string,
  resourceId: string,
  file: File,
  note?: string,
): Promise<FileAttachment> {
  const form = new FormData();
  form.append('file', file);
  if (note) form.append('note', note);
  const { data } = await apiClient.post('/api/v1/attachments', form, {
    params: { resource, resource_id: resourceId },
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
}

export async function downloadAttachment(id: string): Promise<{ url: string; filename: string }> {
  const { data } = await apiClient.get(`/api/v1/attachments/${id}/url`);
  return data;
}

export async function deleteAttachment(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/attachments/${id}`);
}
