# docling-service — 文档智能解析微服务

上传文档 → 转图 → 智谱 GLM 视觉 OCR → 结构化抽取(分类/实体/表格)→ 写入 OpenPTS Postgres。
独立的旁路微服务,FastAPI,监听 **8300**。

## 引擎

`PyMuPDF`(PDF/Word 转图)+ 智谱开放平台 **GLM 视觉 API**(逐页 OCR,并发 4 页)。
取代早期的 Docling 引擎(旧 `parser.py` 已废弃,未纳入本服务)。

## 接口

| 方法 | 路径 | 说明 |
|---|---|---|
| GET  | `/health` | 探活(返回引擎/模型) |
| POST | `/api/v1/parse` | 单文件:`file` + `doc_category`(auto/合同/政策/规则/账单)+ `save_to_db` |
| POST | `/api/v1/batch-parse` | 批量 `files` |
| GET  | `/api/v1/documents` | 已解析文档列表(`doc_type`/`limit`) |
| GET  | `/api/v1/documents/{id}` | 单个文档详情 |

支持格式:PDF / 图片(jpg/png/bmp/tiff/webp)/ Word(需镜像内含 LibreOffice,当前 Dockerfile 未装,Word 走 OCR 前需自行补 `libreoffice`)。

## 环境变量

| 变量 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `ZHIPU_API_KEY` | ✅ | (空) | 智谱开放平台密钥。**仅经环境变量注入,代码/仓库不留真实值。** |
| `ZHIPU_BASE_URL` | | `https://open.bigmodel.cn/api/paas/v4` | |
| `ZHIPU_VISION_MODEL` | | `glm-4v-flash` | 视觉模型 |
| `PG_HOST`/`PG_PORT`/`PG_DB`/`PG_USER`/`PG_PASSWORD` | ✅(密码) | localhost/5432/ptis/ptis/(空) | 复用 OpenPTS 主库;写入自有表 `docling_documents`(启动时 `CREATE TABLE IF NOT EXISTS` 自动建) |

## 数据表

`docling_documents` 的 schema 唯一来源是 `db/migrations/0063_docling_documents.up.sql`(纳入
项目统一 golang-migrate 流);服务侧 `db_writer._ensure_table` 仅作 standalone 兜底,
`CREATE TABLE IF NOT EXISTS` 与迁移同构。

`org_id` 为 `uuid`,与主库多租户口径一致;不设 FK(松耦合,避免跨服务外键的启动期脆弱),
`NULL` 表示未归属省份(standalone/历史)。消费方(经 Go 网关)写入时传入调用者的省份
uuid、查询时按 org 过滤,实现租户隔离。

## 本地运行

```bash
cd docling-service
pip install -r requirements.txt
export ZHIPU_API_KEY=... PG_PASSWORD=...
uvicorn app.main:app --host 0.0.0.0 --port 8300
```

Docker / Compose 见仓库根 `docker-compose.yml` 的 `docling` 服务。
