// 鉴权核心流程：登录、查看自身权限、退出。
import { expect, test } from '@playwright/test';

test.describe('鉴权', () => {
  test('admin 登录成功后跳转仪表盘', async ({ page }) => {
    await page.goto('/login');

    await page.getByLabel(/用户名|账号/).fill('admin');
    await page.getByLabel(/密码/).fill(process.env.ADMIN_PASSWORD || '');
    await page.getByRole('button', { name: /登录/ }).click();

    await expect(page).toHaveURL(/\/dashboard/);
    await expect(page.getByRole('heading', { name: '仪表盘' })).toBeVisible();
    // 至少能看到「客户总数」KPI 卡
    await expect(page.getByText('客户总数').first()).toBeVisible();
  });

  test('错误密码登录被拒绝', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel(/用户名|账号/).fill('admin');
    await page.getByLabel(/密码/).fill('wrong_pass');
    await page.getByRole('button', { name: /登录/ }).click();

    // 仍停留在登录页
    await expect(page).toHaveURL(/\/login/);
  });

  test('退出后无法访问仪表盘', async ({ page }) => {
    // 先登录
    await page.goto('/login');
    await page.getByLabel(/用户名|账号/).fill('admin');
    await page.getByLabel(/密码/).fill(process.env.ADMIN_PASSWORD || '');
    await page.getByRole('button', { name: /登录/ }).click();
    await expect(page).toHaveURL(/\/dashboard/);

    // 退出
    await page.getByRole('button', { name: /退出/ }).click();
    await expect(page).toHaveURL(/\/login/);

    // 直接访问 dashboard 应被重定向回登录
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });
});
