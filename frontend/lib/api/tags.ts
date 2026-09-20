import { apiClient, extractErrorMessage } from './client';

export interface Tag {
  id: string;
  name: string;
  color: string;
  entity_type: string;
  created_at: string;
  updated_at: string;
}

export interface TagInput {
  name: string;
  color: string;
  entity_type: string;
}

export async function listTags(entityType?: string): Promise<{ items: Tag[] }> {
  try {
    const res = await apiClient.get<{ items: Tag[] }>('/api/v1/tags', {
      params: entityType ? { entity_type: entityType } : undefined,
    });
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to fetch tags'));
  }
}

export async function createTag(data: TagInput): Promise<Tag> {
  try {
    const res = await apiClient.post<Tag>('/api/v1/tags', data);
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to create tag'));
  }
}

export async function updateTag(id: string, data: Partial<TagInput>): Promise<Tag> {
  try {
    const res = await apiClient.put<Tag>(`/api/v1/tags/${id}`, data);
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to update tag'));
  }
}

export async function deleteTag(id: string): Promise<void> {
  try {
    await apiClient.delete(`/api/v1/tags/${id}`);
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to delete tag'));
  }
}

// Batch apply tags to entities

// Get tags for a specific entity
