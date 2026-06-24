-- 82: 把新页面注册进动态菜单权限系统（#80 的 menu_pages/role_menu_pages），
-- 否则侧栏走 DB 动态菜单时不会显示。覆盖：客户电量档案(④) + 政策文件(⑤)。
INSERT INTO menu_pages (code, label, href, icon, sort_order, group_name, is_required) VALUES
('page:customer-energy', '客户电量', '/customer-energy', 'Zap',        18, '客户管理', FALSE),
('page:policies',        '政策文件', '/policies',        'ScrollText', 31, '文档中心', FALSE)
ON CONFLICT (code) DO NOTHING;

-- 默认可见角色：super_admin/admin 全见；analyst 两者；viewer 仅政策文件（与其已有「文档解析」一致）
INSERT INTO role_menu_pages (role_code, page_code)
SELECT v.r, v.c FROM (VALUES
  ('super_admin','page:customer-energy'),('super_admin','page:policies'),
  ('admin','page:customer-energy'),('admin','page:policies'),
  ('analyst','page:customer-energy'),('analyst','page:policies'),
  ('viewer','page:policies')
) AS v(r,c)
ON CONFLICT DO NOTHING;
