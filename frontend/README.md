# ptis-frontend (Next.js)

电力交易信息系统 v2 前端。

## 技术栈
- Next.js 15（App Router + RSC）
- React 19
- TypeScript 5.6+
- shadcn/ui + Tailwind CSS
- @tanstack/react-query
- react-hook-form + zod
- recharts

## 本地启动

```bash
cp .env.local.example .env.local
npm install
npm run dev
```

打开 http://localhost:3000

## 目录

```
frontend/
├── app/                     App Router 页面
│   ├── layout.tsx          根布局
│   ├── page.tsx            首页
│   └── globals.css         Tailwind + 主题变量
├── components/
│   └── ui/                 shadcn 组件（按需 add）
├── lib/
│   └── utils.ts            shadcn cn helper
├── public/                 静态资源
├── components.json         shadcn 配置
├── tailwind.config.ts
├── next.config.mjs
└── tsconfig.json
```

## 添加 shadcn 组件

```bash
# 阶段 1 开始按需添加：
npx shadcn@latest add button
npx shadcn@latest add input
npx shadcn@latest add table
npx shadcn@latest add dialog
# ……
```

## 与后端联调

`next.config.mjs` 配置了 rewrites，把 `/api/*` 代理到 Go 网关（默认 `http://localhost:8080`）。
所有业务请求通过 `/api/v1/...` 调用即可，无需关心跨域。
