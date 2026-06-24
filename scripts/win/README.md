# OpenPTS · 开放式电力交易系统 — Windows 11 部署指南

## 系统要求

| 项目 | 要求 |
|------|------|
| 操作系统 | Windows 11 (AMD64/x86_64) |
| 内存 | ≥ 8 GB（推荐 16 GB） |
| 硬盘 | ≥ 10 GB 可用空间 |
| CPU | AMD/Intel x86_64 均可 |
| 网络 | 能访问 Docker Hub（或配了镜像加速器） |

---

## 一、安装 Docker Desktop

> 这一步只需要做一次

1. 打开 https://www.docker.com/products/docker-desktop/
2. 点击 **Download for Windows** 下载安装包
3. 双击 `Docker Desktop Installer.exe` 安装
4. 安装完成后 **重启电脑**
5. 启动 Docker Desktop，等待左下角显示 **绿色 "Engine running"**
6. 确认：打开 CMD，输入 `docker --version`，能看到版本号即可

### 配置镜像加速（国内必须）

Docker Desktop → Settings → Docker Engine → 添加：

```json
{
  "registry-mirrors": [
    "https://docker.1ms.run",
    "https://docker.xuanyuan.me"
  ]
}
```

点击 **Apply & Restart**

---

## 二、获取项目代码

### 方式 A：从 Git 克隆（推荐）

1. 安装 Git：https://git-scm.com/download/win
2. 打开 PowerShell 或 CMD：

```cmd
cd %USERPROFILE%
git clone <你的仓库地址> ptis
cd ptis
```

3. 输入 GitHub 用户名和 Token 登录（私有仓库）

### 方式 B：拷贝压缩包

直接解压到目标目录，如 `D:\ptis`

---

## 三、一键部署

1. 打开 **CMD** 或 **PowerShell**
2. 进入项目目录：

```cmd
cd D:\ptis\scripts\win
```

3. 运行部署脚本：

```cmd
deploy.bat
```

脚本会自动完成：
- ✅ 检查 Docker 环境
- ✅ 生成随机密码并创建 .env 文件
- ✅ 构建 Docker 镜像（首次约 5-10 分钟）
- ✅ 启动所有服务
- ✅ 运行数据库迁移
- ✅ 可选：导入演示数据

4. 看到 `🎉 部署完成！` 后打开浏览器访问：

**http://localhost**

---

## 四、管理命令

在 `scripts\win` 目录下：

| 命令 | 功能 |
|------|------|
| `deploy.bat` | 首次部署 |
| `deploy.bat start` | 启动服务 |
| `deploy.bat stop` | 停止服务 |
| `deploy.bat restart` | 重启服务 |
| `deploy.bat status` | 查看服务状态 |
| `deploy.bat logs` | 实时查看日志（Ctrl+C 退出） |
| `deploy.bat seed` | 重新导入演示数据 |
| `deploy.bat reset` | 完全重置（删数据，慎用） |

---

## 五、端口说明

| 端口 | 服务 | 说明 |
|------|------|------|
| 80 | Nginx | 主入口，访问 http://localhost |
| 3000 | Next.js 前端 | 开发调试用 |
| 8080 | Go 后端 | API 服务 |
| 8300 | docling | 文档解析服务（默认不对外暴露） |
| 5432 | PostgreSQL | 数据库（默认不对外暴露） |
| 9000/9001 | MinIO | 对象存储（默认不对外暴露） |

---

## 六、常见问题

### Q: 构建很慢或报网络错误
配置 Docker 镜像加速器（见第一步），或检查网络代理设置。

### Q: Docker Desktop 报内存不足
Docker Desktop → Settings → Resources → Memory 调到 **4GB+**

### Q: 端口 80 被占用
编辑 `.env` 文件，把 `HTTP_PORT=80` 改为 `HTTP_PORT=8080`，然后访问 http://localhost:8080

### Q: 如何查看日志
```cmd
deploy.bat logs
```
或单独查看某个服务：
```cmd
docker compose logs -f backend
docker compose logs -f frontend
```

### Q: 如何更新代码
```cmd
cd D:\ptis
git pull origin main
deploy.bat restart
```

### Q: 数据存在哪里
数据存储在 Docker 卷中，`deploy.bat stop` 不会丢数据。只有 `deploy.bat reset` 才会删除。

### Q: 如何备份数据库
```cmd
docker compose exec -T postgres pg_dump -U ptis ptis > backup.sql
```

### Q: 如何恢复数据库
```cmd
docker compose exec -T postgres psql -U ptis ptis < backup.sql
```

---

## 七、卸载

1. `deploy.bat stop` 停止服务
2. 删除项目目录
3. 如需彻底清理 Docker 镜像：
```cmd
docker system prune -a
```
4. 卸载 Docker Desktop（控制面板 → 程序与功能）

---

> 技术支持：将此文档和报错截图发给开发人员
