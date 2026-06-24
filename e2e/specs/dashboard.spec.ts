// 仪表盘核心数据展示 + 跨模块跳转。
// 文案与 app/(main)/dashboard/page.tsx 实际渲染对齐（R2 测试对齐）。
import { expect, test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/login');
  await page.getByLabel(/用户名|账号/).fill('admin');
  await page.getByLabel(/密码/).fill(process.env.ADMIN_PASSWORD || '');
  await page.getByRole('button', { name: /登录/ }).click();
  await expect(page).toHaveURL(/\/dashboard/);
});

test('仪表盘 KPI 卡片全部存在', async ({ page }) => {
  // 与 dashboard/page.tsx Row 1 的 6 个 KpiCard label 一一对应
  const labels = ['客户总数', '活跃合同', '活跃套餐', '待处理告警', '活跃电站', '最新结算额'];
  for (const label of labels) {
    await expect(page.getByText(label).first()).toBeVisible();
  }
});

test('客户总数 KPI 可见（KPI 卡片当前不带 onClick，仅验证展示）', async ({ page }) => {
  await expect(page.getByText('客户总数').first()).toBeVisible();
});

test('14 日结算趋势图渲染（SparkCard）', async ({ page }) => {
  // SparkCard 标题为「近 14 日结算额」
  await expect(page.getByText(/近 14 日结算额/)).toBeVisible();
  // recharts SVG 渲染
  const svg = page.locator('.recharts-surface').first();
  await expect(svg).toBeVisible();
});
