// Playwright 配置：本地 npx playwright test 跑前先把 backend + frontend 拉起来。
// CI 中由 GitHub Actions 启 compose 后跑。
// 默认地址按「本地 next dev :3000」；对 compose 全栈跑时前端不再对外暴露 3000，
// 需 E2E_BASE_URL=http://localhost（经 nginx）覆盖。
// admin 密码由种子程序随机生成，运行前用 ADMIN_PASSWORD 环境变量传入。
import { defineConfig, devices } from '@playwright/test';

const BASE_URL = process.env.E2E_BASE_URL || 'http://localhost:3000';

export default defineConfig({
  testDir: './specs',
  fullyParallel: false,   // 业务有强状态，串行更稳
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  timeout: 30_000,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: BASE_URL,
    trace: 'on-first-retry',
    video: 'retain-on-failure',
    screenshot: 'only-on-failure',
    viewport: { width: 1440, height: 900 },
    locale: 'zh-CN',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
