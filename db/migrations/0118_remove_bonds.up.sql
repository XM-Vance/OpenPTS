-- 0118: 删除「保函管理」模块（下线）。
-- 背景：保函管理功能下线，移除 bonds 表及其菜单种子。数据不可恢复。
-- 历史迁移（0040 建表、0046 性能索引、0093 货币精度、0102 加 org_id）均为
-- 共享迁移，不改原文，仅以此迁移正向 DROP。
-- 注意：role_menu_pages 对 menu_pages(code) 有 ON DELETE CASCADE，删菜单种子会
-- 自动清理各角色的 page:bonds 关联行。

-- 1) 删菜单种子（role_menu_pages 经外键级联自动清理）
DELETE FROM menu_pages WHERE code = 'page:bonds';

-- 2) 删表（CASCADE 连带删除 0040/0046/0102 建立的全部索引）
DROP TABLE IF EXISTS bonds CASCADE;
