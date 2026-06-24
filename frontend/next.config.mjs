/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Docker 部署使用 standalone 模式：node server.js 直接跑（不需 next start）。
  output: 'standalone',
  // Next.js 15 默认启用 SWC 压缩（swcMinify 已移除，无需显式配置）
  // 远程图片域名白名单（R7：按需添加，避免 next/image 403）
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**.amazonaws.com',
      },
      {
        protocol: 'http',
        hostname: 'minio',
        port: '9000',
      },
    ],
  },
  // 开发期将 /api/* 反向代理到 Go 网关；生产环境由 nginx 反代，无需此 rewrites。
  async rewrites() {
    const apiBase = process.env.NEXT_PUBLIC_API_BASE || 'http://backend:8080';
    return [
      {
        source: '/api/:path*',
        destination: `${apiBase}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
