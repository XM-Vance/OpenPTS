"""
GLM-OCR 文档解析器
数字版 PDF → 直接读取文本层（瞬时）
扫描版 PDF / 图片 → 智谱 layout_parsing API（glm-ocr）

API 文档: https://docs.bigmodel.cn/api-reference/模型-api/文档解析
"""

import json, logging, os, re, time, base64
from pathlib import Path
from typing import Optional
import httpx

logger = logging.getLogger("docling-service.parser")

# 智谱API配置
ZHIPU_API_KEY = os.getenv("ZHIPU_API_KEY", "")
ZHIPU_BASE_URL = os.getenv("ZHIPU_BASE_URL", "https://open.bigmodel.cn/api/paas/v4")
ZHIPU_VISION_MODEL = os.getenv("ZHIPU_VISION_MODEL", "glm-ocr")
LAYOUT_PARSING_URL = f"{ZHIPU_BASE_URL}/layout_parsing"

# ── 售电行业分类规则 ──────────────────────────────
CATEGORY_RULES = {
    "合同": ["合同", "协议", "委托", "甲方", "乙方", "签约", "履约"],
    "政策": ["通知", "方案", "办法", "细则", "规定", "意见", "改革", "实施意见"],
    "规则": ["规则", "规程", "交易规则", "结算规则", "运营规则"],
    "账单": ["账单", "结算单", "发票", "电费", "应收", "应付"],
    "资质": ["许可证", "营业执照", "资质", "认证", "备案"],
}

ENTITY_PATTERNS = {
    "金额": [
        r'(\d+(?:\.\d{1,4})?)\s*(?:万元|元|亿元|万)',
        r'(?:人民币|金额|价格|费用|总价|单价)[：:]\s*(\d+(?:\.\d+)?)',
    ],
    "电价": [
        r'(\d+(?:\.\d{2,6}))\s*(?:元[/／每]千瓦时|元[/／每]度|元/kWh|元/MWh)',
        r'(?:电价|电费|上网电价|销售电价)[：:]\s*(\d+(?:\.\d+)?)',
    ],
    "电量": [
        r'(\d+(?:\.\d+)?)\s*(?:万千瓦时|千瓦时|kWh|MWh|GWh|兆瓦时|度|亿千瓦时)',
        r'(?:电量|用电量|发电量|合同电量)[：:]\s*(\d+(?:\.\d+)?)',
    ],
    "企业": [
        r'([\u4e00-\u9fa5]{2,20}(?:有限公司|股份公司|集团公司|公司|电厂|电站|交易中心))',
    ],
    "日期": [
        r'(\d{4})\s*年\s*(\d{1,2})\s*月\s*(\d{1,2})\s*日',
        r'(\d{4}-\d{1,2}-\d{1,2})',
    ],
    # ── 套餐相关实体 ──
    "套餐": [
        r'(?:套餐|产品|方案)\s*[名称编号]*[：:]*\s*([^\n,，。；;]{2,50})',
        r'(固定电价|浮动电价|分成电价|绿电套餐|市场联动|价格联动|峰谷套餐|时段套餐|单一制|两部制)',
    ],
    "浮动比例": [
        r'(\d+(?:\.\d+)?)\s*[%％](?:\s*(?:上下)?浮动)',
        r'(?:浮动比例|分成比例|让利比例|收益分成)[：:]\s*(\d+(?:\.\d+)?)\s*[%％]?',
        r'(?:出让|让利)[：:]\s*(\d+(?:\.\d+)?)\s*(?:元[/／每]千瓦时|元[/／每]度|元/kWh)',
    ],
    "绿电比例": [
        r'(\d+(?:\.\d+)?)\s*[%％](?:\s*(?:绿电|绿色电力|新能源))',
        r'(?:绿电比例|绿色电力比例|新能源占比)[：:]\s*(\d+(?:\.\d+)?)\s*[%％]?',
    ],
    "偏差考核": [
        r'偏差[电量]*[：:]\s*(\d+(?:\.\d+)?)\s*(?:万千瓦时|千瓦时|kWh|MWh|度)?',
        r'(?:偏差考核|偏差费用|考核电费)[：:]\s*(\d+(?:\.\d+)?)\s*(?:万元|元)?',
    ],
}


# ── 工具函数 ──────────────────────────────────────
def classify_document(text: str) -> str:
    scores: dict[str, int] = {}
    for cat, keywords in CATEGORY_RULES.items():
        scores[cat] = sum(text.count(kw) for kw in keywords)
    best_cat, best_score = "其他", 0
    for cat, score in scores.items():
        if score > best_score:
            best_cat, best_score = cat, score
    return best_cat


def extract_entities(text: str) -> dict:
    entities = {}
    for entity_type, patterns in ENTITY_PATTERNS.items():
        values = set()
        for pattern in patterns:
            matches = re.findall(pattern, text)
            for m in matches:
                if isinstance(m, tuple):
                    m = "".join(m)
                if m:
                    values.add(m)
        if values:
            entities[entity_type] = list(values)[:10]
    return entities


def _file_to_data_url(file_path: str) -> str:
    """将本地文件转为 data URL（小文件直接传 base64）。"""
    path = Path(file_path)
    ext = path.suffix.lower()
    mime_map = {
        ".pdf": "application/pdf",
        ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
        ".png": "image/png",
        ".bmp": "image/bmp",
        ".tiff": "image/tiff", ".tif": "image/tiff",
        ".webp": "image/webp",
    }
    mime = mime_map.get(ext, "application/octet-stream")
    with open(file_path, "rb") as f:
        b64 = base64.b64encode(f.read()).decode()
    return f"data:{mime};base64,{b64}"


def pdf_text_layer(pdf_path: str) -> list[str]:
    """读取 PDF 文本层（数字版 PDF 自带文字，无需 OCR）。"""
    import fitz
    doc = fitz.open(pdf_path)
    texts = [doc[i].get_text("text") or "" for i in range(doc.page_count)]
    doc.close()
    return texts


# ── 核心解析 ──────────────────────────────────────
def _call_layout_parsing(file_input: str, start_page: int = None, end_page: int = None) -> dict:
    """
    调用智谱 layout_parsing API (glm-ocr)。
    file_input: 文件路径 或 URL 或 data URL
    返回: API 原始 JSON 响应
    """
    if not ZHIPU_API_KEY:
        raise RuntimeError("未配置 ZHIPU_API_KEY，无法调用 GLM-OCR API")

    # 如果是本地文件路径，转为 data URL
    if os.path.isfile(file_input):
        file_param = _file_to_data_url(file_input)
    else:
        file_param = file_input  # 已经是 URL 或 data URL

    payload = {
        "model": "glm-ocr",
        "file": file_param,
    }
    if start_page is not None:
        payload["start_page_id"] = start_page
    if end_page is not None:
        payload["end_page_id"] = end_page

    headers = {
        "Authorization": f"Bearer {ZHIPU_API_KEY}",
        "Content-Type": "application/json",
    }

    for attempt in range(3):
        try:
            with httpx.Client(timeout=180) as client:
                resp = client.post(LAYOUT_PARSING_URL, json=payload, headers=headers)
                resp.raise_for_status()
                return resp.json()
        except httpx.HTTPStatusError as e:
            body = e.response.text[:500] if e.response else ""
            logger.warning(f"layout_parsing attempt {attempt+1} failed: {e.response.status_code} {body}")
            if e.response.status_code == 429:
                time.sleep(10 * (attempt + 1))
            elif attempt < 2:
                time.sleep(2 ** attempt)
        except Exception as e:
            logger.warning(f"layout_parsing attempt {attempt+1} error: {e}")
            if attempt < 2:
                time.sleep(2 ** attempt)

    raise RuntimeError("layout_parsing API 调用失败，3次重试后放弃")


def _extract_tables_from_text(pages_text: list[str]) -> list[dict]:
    """从文本中提取 Markdown 表格。"""
    tables = []
    for text in pages_text:
        table_blocks = re.findall(r'(\|.+\|\n\|[-| :]+\|\n(?:\|.+\|\n)*)', text)
        for tb in table_blocks:
            tables.append({"markdown": tb.strip()})
    return tables


def _assemble_result(pages_text: list[str], layout_details: list = None) -> dict:
    """页文本 → 全文/分类/实体/摘要/表格 汇总结果。"""
    full_text = "\n\n---\n\n".join(
        f"## 第{i+1}页\n\n{text}" for i, text in enumerate(pages_text) if text
    )

    # 表格：优先从 layout_details 提取，备从文本提取
    tables = []
    if layout_details:
        for page_layouts in layout_details:
            if isinstance(page_layouts, list):
                for item in page_layouts:
                    if isinstance(item, dict) and item.get("label") == "table":
                        content = item.get("content", "")
                        if content:
                            tables.append({"markdown": content, "bbox": item.get("bbox_2d")})

    if not tables:
        tables = _extract_tables_from_text(pages_text)

    doc_type = classify_document(full_text)
    entities = extract_entities(full_text)
    summary = "\n\n".join(pages_text[:3])[:2000]

    return {
        "text_content": full_text,
        "doc_type": doc_type,
        "entities": entities,
        "summary": summary,
        "tables": tables,
        "page_count": len(pages_text),
    }


def parse_file(file_path: str) -> dict:
    """
    解析文件：
    1. 数字版 PDF → 直接读取文本层（瞬时，免 API）
    2. 扫描版 PDF / 图片 → glm-ocr layout_parsing API
    """
    path = Path(file_path)
    ext = path.suffix.lower()

    if ext not in (".pdf", ".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".tif", ".webp"):
        raise ValueError(f"不支持的文件格式: {ext}")

    file_size_mb = path.stat().st_size / (1024 * 1024)
    logger.info(f"开始解析: {path.name} ({file_size_mb:.1f}MB)")

    # ── PDF：快路径（数字版直读文本层）──
    if ext == ".pdf":
        layer = pdf_text_layer(file_path)
        avg_chars = sum(len(t.strip()) for t in layer) / max(len(layer), 1)
        if layer and avg_chars > 80:
            logger.info(f"数字版 PDF，使用文本层: {path.name}（{len(layer)} 页，均 {avg_chars:.0f} 字/页）")
            pages_text = [t.strip() for t in layer]
            return _assemble_result(pages_text)

        # 扫描版 PDF → glm-ocr
        logger.info(f"扫描版 PDF，走 glm-ocr: {path.name}")
        import fitz
        doc = fitz.open(file_path)
        total_pages = doc.page_count
        doc.close()

        all_md_results = []
        all_layout_details = []

        if total_pages <= 100:
            result = _call_layout_parsing(file_path)
            all_md_results.append(result.get("md_results", ""))
            all_layout_details.extend(result.get("layout_details", []))
        else:
            for batch_start in range(1, total_pages + 1, 100):
                batch_end = min(batch_start + 99, total_pages)
                logger.info(f"解析第 {batch_start}-{batch_end} 页 / 共 {total_pages} 页")
                result = _call_layout_parsing(file_path, start_page=batch_start, end_page=batch_end)
                all_md_results.append(result.get("md_results", ""))
                all_layout_details.extend(result.get("layout_details", []))

        pages_text = [md for md in all_md_results if md]
        return _assemble_result(pages_text, all_layout_details)

    # ── 图片 → glm-ocr ──
    if file_size_mb > 10:
        raise ValueError(f"图片文件过大 ({file_size_mb:.1f}MB)，限制 10MB")

    result = _call_layout_parsing(file_path)
    md_text = result.get("md_results", "")
    layout_details = result.get("layout_details", [])
    return _assemble_result([md_text] if md_text else ["[图片解析结果为空]"], layout_details)
