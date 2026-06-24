import { apiClient } from './client';

export interface MenuPage {
  code: string;
  label: string;
  href: string;
  icon: string;
  sort_order: number;
  group_name: string;
  is_required: boolean;
}

export interface MenuPageWithRoles extends MenuPage {
  roles: string[];
}

export const menuApi = {
  /** 获取所有页面（管理员用，含角色分配信息） */
  getAllPages: async (): Promise<MenuPageWithRoles[]> => {
    const { data } = await apiClient.get('/api/v1/menu/pages');
    return data.items;
  },

  /** 获取当前用户可见的页面列表（侧边栏用） */
  getVisiblePages: async (): Promise<MenuPage[]> => {
    const { data } = await apiClient.get('/api/v1/menu/visible');
    return data.items;
  },

  /** 更新某角色的可见页面集 */
  updateRolePages: async (roleCode: string, pageCodes: string[]) => {
    const { data } = await apiClient.put(`/api/v1/menu/roles/${roleCode}`, {
      page_codes: pageCodes,
    });
    return data;
  },
};
