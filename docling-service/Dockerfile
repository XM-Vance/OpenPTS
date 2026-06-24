# PTIS 文档OCR服务 — GLM视觉引擎（PyMuPDF + 智谱 API）。
# 两阶段构建：builder 装依赖 → runner 精简镜像，非 root，带健康检查。

# ───── 构建阶段 ─────
FROM python:3.12-slim AS builder

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

WORKDIR /build

COPY requirements.txt .
RUN pip install --user -r requirements.txt

# ───── 运行阶段 ─────
FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PATH=/home/app/.local/bin:$PATH \
    TZ=Asia/Shanghai

# curl 供 HEALTHCHECK 用；tzdata 设时区。无 build-essential，镜像更小。
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -s /bin/bash app

USER app
WORKDIR /app

# 从 builder 复制 user-site 安装的依赖（chown 给非 root 用户）。
COPY --from=builder --chown=app:app /root/.local /home/app/.local
COPY --chown=app:app app/ app/

EXPOSE 8300

HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=3 \
    CMD curl -fs http://localhost:8300/health || exit 1

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8300"]
