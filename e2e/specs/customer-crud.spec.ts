// 客户档案 CRUD：新建 → 列表可见 → 删除。
// 文案与 app/(main)/customers/_view.tsx 实际渲染对齐（R2 测试对齐）。
// 注意：customer 表单的 <Label> 无 htmlFor、<Input> 无 id，getByLabel 定位不到，
// 改用「按 Label 文本定位其后的 Input」模式。
//
// 稳健性要点（曾导致 flaky）：
// 1. 多租户铁律：HQ「全部省」(`X-Org-Id: *`) 下写业务数据返回 400（ErrOrgRequired）。
//    activeOrg 存于 localStorage + React state，且不在 list 的 useQuery key 里——
//    切换省份后列表不会自动 refetch。所以这里先选具体省，再 reload，
//    确保页面整条链路（列表查询、后续创建）都以新省发起。
// 2. 不用 waitForTimeout 等 React Query 网络回包——改为显式 wait 后端 POST 响应，
//    再用搜索框（服务端 keyword 过滤）定位新客户，与排序/分页无关。
import { expect, test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/login');
  await page.getByLabel(/用户名|账号/).fill('admin');
  await page.getByLabel(/密码/).fill(process.env.ADMIN_PASSWORD || '');
  await page.getByRole('button', { name: /登录/ }).click();
  // 关键：等登录跳转完成、鉴权 cookie 落定，否则后续 goto 业务页会被重定向回 /login。
  await expect(page).toHaveURL(/\/dashboard/);
});

// 按 Label 文本定位紧邻的输入框（Label 与 Input 在同一 div.space-y-2 内）。
// 限定在 dialog 内，避免与表格表头（同名）冲突。
const fieldAfterLabel = (dialog: import('@playwright/test').Locator, labelText: string) =>
  dialog.locator('div.space-y-2', { hasText: labelText }).getByRole('textbox').first();

test('新建客户 → 列表可见 → 删除', async ({ page }) => {
  await page.goto('/customers');
  await expect(page).not.toHaveURL(/\/login/);
  await expect(page.getByRole('heading', { name: '客户档案管理' })).toBeVisible({ timeout: 15000 });

  // 多租户铁律：HQ「全部省」下写业务数据返回 400（ErrOrgRequired）。
  // 切到第一个具体省，并 reload 让列表查询与新创建都以该省作用域发起。
  const orgSelect = page.locator('select[title="切换省份"]');
  await expect(orgSelect).toBeVisible();
  const firstOrgOption = await orgSelect
    .locator('option')
    .filter({ hasNotText: '全部省' })
    .first()
    .getAttribute('value');
  expect(firstOrgOption).toBeTruthy();
  await orgSelect.selectOption(firstOrgOption as string);
  await page.reload();
  await expect(page.getByRole('heading', { name: '客户档案管理' })).toBeVisible();
  // reload 后选中值应保持（来自 localStorage）
  await expect(orgSelect).toHaveValue(firstOrgOption as string);

  const uniqueName = `e2e测试客户_${Date.now()}`;

  await page.getByRole('button', { name: '新建客户' }).click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible({ timeout: 5000 });

  await fieldAfterLabel(dialog, '客户名称').fill(uniqueName);
  await fieldAfterLabel(dialog, '简称').fill('e2e');
  await fieldAfterLabel(dialog, '所在地').fill('广州');

  // 监听创建请求，等真正成功（2xx）后再继续，避免与 React Query refetch 竞态。
  const createRespPromise = page.waitForResponse(
    (res) => res.url().endsWith('/api/v1/customers') && res.request().method() === 'POST' && res.ok(),
    { timeout: 15000 },
  );

  await dialog.getByRole('button', { name: /^保存$/ }).click();

  // 等创建请求 2xx 返回（onSaved → invalidateQueries → 列表 refetch）
  const resp = await createRespPromise;
  expect(resp.ok(), '创建客户请求应成功').toBeTruthy();

  // 用搜索框做服务端 keyword 过滤，定位新客户，与排序/分页无关。
  const searchInput = page.getByPlaceholder('搜索客户名 / 简称');
  // 先注册响应监听，再触发搜索，避免竞态漏掉请求。
  const searchRespPromise = page
    .waitForResponse(
      (res) =>
        res.url().includes('/api/v1/customers') &&
        res.url().includes(encodeURIComponent(uniqueName)) &&
        res.ok(),
      { timeout: 15000 },
    )
    .catch(() => null);
  await searchInput.fill(uniqueName);
  await searchInput.press('Enter');
  await searchRespPromise;

  await expect(page.getByText(uniqueName).first()).toBeVisible({ timeout: 10000 });

  // 找到该行的「删除」按钮（同行内），window.confirm 自动确认。
  const row = page.locator('tr', { hasText: uniqueName });
  await expect(row).toBeVisible();
  page.once('dialog', (d) => d.accept());
  await row.getByRole('button', { name: /删除/ }).click();

  // 删除后用同样的搜索确认该客户已不在。
  const afterDeleteSearchPromise = page
    .waitForResponse(
      (res) =>
        res.url().includes('/api/v1/customers') &&
        res.url().includes(encodeURIComponent(uniqueName)) &&
        res.ok(),
      { timeout: 15000 },
    )
    .catch(() => null);
  await searchInput.fill(uniqueName);
  await searchInput.press('Enter');
  await afterDeleteSearchPromise;

  await expect(page.getByText(uniqueName)).toHaveCount(0, { timeout: 10000 });
});
