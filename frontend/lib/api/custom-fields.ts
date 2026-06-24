import { apiClient, extractErrorMessage } from './client';

export interface CustomField {
  id: string;
  entity_type: string;
  field_label: string;
  field_key: string;
  field_type: string;
  is_required: boolean;
  sort_order: number;
  options?: string[] | null;
  default_value?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CustomFieldInput {
  entity_type: string;
  field_label: string;
  field_key: string;
  field_type: string;
  is_required?: boolean;
  sort_order?: number;
  options?: string[];
  default_value?: string;
}

export async function listCustomFields(entityType?: string): Promise<{ items: CustomField[] }> {
  try {
    const res = await apiClient.get<{ items: CustomField[] }>('/custom-fields', {
      params: entityType ? { entity_type: entityType } : undefined,
    });
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to fetch custom fields'));
  }
}

export async function createCustomField(data: CustomFieldInput): Promise<CustomField> {
  try {
    const res = await apiClient.post<CustomField>('/custom-fields', data);
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to create custom field'));
  }
}

export async function updateCustomField(id: string, data: Partial<CustomFieldInput>): Promise<CustomField> {
  try {
    const res = await apiClient.put<CustomField>(`/custom-fields/${id}`, data);
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to update custom field'));
  }
}

export async function deleteCustomField(id: string): Promise<void> {
  try {
    await apiClient.delete(`/custom-fields/${id}`);
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to delete custom field'));
  }
}
