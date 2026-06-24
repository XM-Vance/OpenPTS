# ptis-v2 前端：Next.js 15 standalone 输出 + 国内 npm 镜像加速。
# 三阶段：deps → builder → runner，最终镜像约 200 MB。

# ───── 依赖装载 ─────
FROM node:20-alpine AS deps
WORKDIR /app
RUN apk add --no-cache libc6-compat

COPY package.json package-lock.json* ./
# CI 模式安装：与 lock 严格一致。
RUN npm config set registry https://registry.npmmirror.com && npm ci --no-audit --no-fund

# ───── 构建 ─────
FROM node:20-alpine AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# 关闭 telemetry，加速首次构建。
ENV NEXT_TELEMETRY_DISABLED=1 NODE_ENV=production

RUN npm run build

# ───── 运行 ─────
FROM node:20-alpine AS runner
WORKDIR /app

RUN apk add --no-cache wget tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && addgroup -S app && adduser -S -G app app

ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 PORT=3000 HOSTNAME=0.0.0.0

# Next standalone 模式：自带最小 node_modules + server.js
COPY --from=builder --chown=app:app /app/.next/standalone ./
COPY --from=builder --chown=app:app /app/.next/static ./.next/static
COPY --from=builder --chown=app:app /app/public ./public

USER app
EXPOSE 3000

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:3000 >/dev/null 2>&1 || exit 1

CMD ["node", "server.js"]
