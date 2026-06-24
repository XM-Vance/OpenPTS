"""
PTIS Docling Service — 文档智能解析微服务
上传文件 → GLM视觉OCR → 提取结构化数据 → 写入PTIS数据库
核心引擎：PyMuPDF + 智谱GLM视觉API（替代Docling）
"""

from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional
import tempfile, os, json, logging, asyncio
from pathlib import Path

from app.glm_parser import parse_file
from app.extractor import extract_fields_with_fallback
from app.db_writer import DBWriter

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("docling-service")

# 上传安全：大小上限（默认 10MB）+ 扩展名白名单（防任意后缀/DoS）
MAX_UPLOAD_BYTES = int(os.getenv("MAX_UPLOAD_BYTES", str(10 * 1024 * 1024)))
ALLOWED_SUFFIXES = {".pdf", ".png", ".jpg", ".jpeg", ".tif", ".tiff"}


def _validate_upload(file: UploadFile, content: bytes) -> None:
    """校验上传文件大小与扩展名，违例抛 413/415。"""
    if len(content) > MAX_UPLOAD_BYTES:
        raise HTTPException(
            status_code=413,
            detail=f"文件过大（{len(content)} 字节），上限 {MAX_UPLOAD_BYTES} 字节",
        )
    suffix = Path(file.filename or "").suffix.lower()
    if suffix and suffix not in ALLOWED_SUFFIXES:
        raise HTTPException(
            status_code=415,
            detail=f"不支持的文件类型 {suffix}，允许：{', '.join(sorted(ALLOWED_SUFFIXES))}",
        )


def _safe_server_error(stage: str, e: Exception) -> HTTPException:
    """记录完整异常到服务端日志，只向客户端返回脱敏的通用消息。

    避免把内部错误细节（如 DB DSN、文件路径、栈信息）透传给调用方。
    """
    logger.error("%s失败: %s", stage, e, exc_info=True)
    return HTTPException(status_code=500, detail=f"{stage}失败，请检查服务端日志")


def _build_package_info(entities: dict, doc_type: str) -> dict | None:
    """从正则实体中提取套餐相关字段，组装成 package_info 结构。"""
    pkg = {}
    # 套餐名称/类型
    if entities.get("套餐"):
        pkg["package_type"] = entities["套餐"][0]
        if len(entities["套餐"]) > 1:
            pkg["package_name"] = entities["套餐"][1]
    # 浮动比例
    if entities.get("浮动比例"):
        pkg["floating_ratio"] = entities["浮动比例"][0]
    # 绿电比例
    if entities.get("绿电比例"):
        pkg["green_power_ratio"] = entities["绿电比例"][0]
    # 偏差考核
    if entities.get("偏差考核"):
        pkg["deviation"] = entities["偏差考核"][0]
    # 电价（从"电价"实体）
    if entities.get("电价"):
        pkg["price"] = entities["电价"][0]
    # 金额
    if entities.get("金额"):
        pkg["amount"] = entities["金额"][0]

    return pkg if pkg else None

app = FastAPI(title="PTIS Document OCR Service", version="2.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

db = DBWriter()


# ── 数据模型 ──────────────────────────────────────────────

class ParseResult(BaseModel):
    filename: str
    doc_type: str          # 合同/政策/规则/账单/其他
    text_content: str      # 全文markdown
    tables: list[dict]     # 提取的表格
    entities: dict         # 关键实体（金额/日期/单位等）
    summary: str           # 摘要
    page_count: int        # 页数

class HealthResponse(BaseModel):
    status: str
    engine: str
    model: str


# ── 接口 ──────────────────────────────────────────────────

@app.get("/health", response_model=HealthResponse)
async def health():
    from app.glm_parser import ZHIPU_VISION_MODEL
    return HealthResponse(
        status="ok",
        engine="GLM Vision OCR (PyMuPDF + 智谱API)",
        model=ZHIPU_VISION_MODEL,
    )


@app.post("/api/v1/parse", response_model=ParseResult)
async def parse_document(
    file: UploadFile = File(...),
    doc_category: str = Form("auto"),    # auto/合同/政策/规则/账单
    save_to_db: bool = Form(True),
    org_id: Optional[str] = Form(None),  # 归属省份(uuid);经 Go 网关传入,standalone 可空
):
    """上传文件 → GLM视觉OCR → 返回结构化数据（可选写入数据库）"""
    suffix = Path(file.filename or "doc.pdf").suffix
    with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
        content = await file.read()
        _validate_upload(file, content)
        tmp.write(content)
        tmp_path = tmp.name

    try:
        logger.info(f"解析文件: {file.filename} ({len(content)} bytes)")
        # parse_file 内部用同步 httpx.Client + time.sleep 重试，会阻塞事件循环；
        # 用 to_thread 卸载到线程池，避免并发解析时饿死其他请求。
        result = await asyncio.to_thread(parse_file, tmp_path)

        # 如果用户指定了分类，覆盖自动分类
        if doc_category != "auto":
            result["doc_type"] = doc_category

        result["filename"] = file.filename

        if save_to_db:
            # 从 entities 中提取套餐信息
            package_info = _build_package_info(result.get("entities", {}), result["doc_type"])
            db.upsert(
                filename=result["filename"],
                doc_type=result["doc_type"],
                text_content=result["text_content"],
                tables=result["tables"],
                entities=result["entities"],
                summary=result["summary"],
                org_id=org_id,
                package_info=package_info,
            )
            logger.info(f"已写入数据库: {result['doc_type']}")

        return ParseResult(**result)
    except HTTPException:
        raise  # 业务校验类（如 400）原样抛出
    except Exception as e:
        raise _safe_server_error("文档解析", e)
    finally:
        os.unlink(tmp_path)


@app.post("/api/v1/batch-parse")
async def batch_parse(
    files: list[UploadFile] = File(...),
    org_id: Optional[str] = Form(None),
):
    """批量上传解析"""
    results = []
    for f in files:
        suffix = Path(f.filename or "doc.pdf").suffix
        with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
            content = await f.read()
            _validate_upload(f, content)
            tmp.write(content)
            tmp_path = tmp.name
        try:
            # 同步 parse_file 卸载到线程池（见 parse_document 注释）。
            r = await asyncio.to_thread(parse_file, tmp_path)
            r["filename"] = f.filename
            db.upsert(
                filename=r["filename"],
                doc_type=r["doc_type"],
                text_content=r["text_content"],
                tables=r["tables"],
                entities=r["entities"],
                summary=r["summary"],
                org_id=org_id,
            )
            results.append(r)
        except Exception as e:
            logger.error("批量解析单文件失败: %s", e, exc_info=True)
            # 只回脱敏标记，不把内部错误细节写入响应体
            results.append({"filename": f.filename, "error": "解析失败"})
        finally:
            os.unlink(tmp_path)
    return {"total": len(results), "results": results}


class ExtractRequest(BaseModel):
    text: str
    doc_type: str = "其他"


@app.post("/api/v1/extract")
async def extract(req: ExtractRequest):
    """从解析后的全文按文档类型定向提取结构化字段（GLM 文本模型，失败回退正则）"""
    if not req.text.strip():
        raise HTTPException(status_code=400, detail="text 为空")
    try:
        # extract_fields_with_fallback 内部用同步 httpx.Client 调 LLM，卸载到线程池。
        fields = await asyncio.to_thread(extract_fields_with_fallback, req.text, req.doc_type)
        return {"doc_type": req.doc_type, "fields": fields}
    except Exception as e:
        raise _safe_server_error("结构化提取", e)


@app.get("/api/v1/documents")
async def list_documents(doc_type: Optional[str] = None, limit: int = 50,
                         org_id: Optional[str] = None):
    """查询已解析的文档列表（org_id 给定时按省份隔离）"""
    return db.list_docs(doc_type=doc_type, limit=limit, org_id=org_id)


@app.get("/api/v1/documents/{doc_id}")
async def get_document(doc_id: int, org_id: Optional[str] = None):
    """查询单个文档详情（org_id 给定时非本省返回 404）"""
    doc = db.get_doc(doc_id, org_id=org_id)
    if not doc:
        raise HTTPException(status_code=404, detail="文档不存在")
    return doc
