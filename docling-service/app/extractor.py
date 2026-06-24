"""
结构化字段提取 — 文本模型按文档类型定向提取
优先使用 DeepSeek（deepseek-chat），未配置时回退 GLM（glm-4-flash）。
输入：全文 markdown + 文档类型 → 输出：带字段名/单位/置信度的 JSON 字段列表
失败时回退到 glm_parser 的正则实体（散值，置信度低）。
"""

import json, logging, os, re
import httpx

from app.glm_parser import ZHIPU_API_KEY, ZHIPU_BASE_URL, extract_entities

logger = logging.getLogger("docling-service.extractor")

# DeepSeek 配置（结构化字段提取优先使用）
DEEPSEEK_API_KEY = os.getenv("DEEPSEEK_API_KEY", "")
DEEPSEEK_BASE_URL = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1")
DEEPSEEK_TEXT_MODEL = os.getenv("DEEPSEEK_TEXT_MODEL", "deepseek-chat")

# GLM 备用配置
ZHIPU_TEXT_MODEL = os.getenv("ZHIPU_TEXT_MODEL", "glm-4-flash")

# 每类文档要提取的字段 schema（key → 中文标签）。
# 提示词驱动 GLM 只输出这些 key 的 JSON；找不到的字段返回 null。
FIELD_SCHEMAS: dict[str, list[tuple[str, str]]] = {
    "合同": [
        ("contract_no", "合同编号"),
        ("party_a", "甲方"),
        ("party_b", "乙方"),
        ("sign_date", "签订日期"),
        ("start_date", "合同起始日期"),
        ("end_date", "合同截止日期"),
        ("energy", "合同电量"),
        ("price", "合同电价"),
        ("total_amount", "合同金额"),
        # ── 套餐信息 ──
        ("package_name", "套餐名称"),
        ("package_type", "套餐类型（固定/浮动/分成/绿电等）"),
        ("pricing_mode", "计价方式（单一制/两部制等）"),
        ("fixed_price", "固定电价/基准电价"),
        ("floating_ratio", "浮动比例/分成比例"),
        ("green_power_ratio", "绿电比例"),
        ("capacity_fee", "容量电费/基本电费"),
        ("adjustment_rule", "价格调整规则"),
    ],
    "结算单": [
        ("period", "结算月份"),
        ("customer", "客户/单位名称"),
        ("energy", "结算电量"),
        ("amount", "电费金额"),
        ("price", "平均电价"),
        ("capacity_fee", "容量/输配电费"),
        ("ancillary_fee", "辅助服务费"),
        # ── 套餐信息 ──
        ("package_name", "套餐名称"),
        ("package_type", "套餐类型"),
        ("pricing_mode", "计价方式"),
        ("fixed_price", "固定电价"),
        ("floating_ratio", "浮动/分成比例"),
        ("deviation_amount", "偏差电量"),
        ("deviation_penalty", "偏差考核费用"),
    ],
    "账单": [
        ("period", "账单月份"),
        ("customer", "客户/单位名称"),
        ("energy", "用电量"),
        ("amount", "应付金额"),
        ("price", "平均电价"),
    ],
    "政策": [
        ("title", "文件标题"),
        ("doc_no", "文号"),
        ("issuer", "发文单位"),
        ("publish_date", "发布日期"),
        ("effective_date", "生效日期"),
    ],
    "规则": [
        ("title", "文件标题"),
        ("issuer", "发布机构"),
        ("publish_date", "发布日期"),
        ("effective_date", "生效日期"),
    ],
    "资质": [
        ("company", "企业名称"),
        ("cert_name", "证照名称"),
        ("cert_no", "证照编号"),
        ("issue_date", "发证日期"),
        ("expire_date", "有效期至"),
    ],
}

# 通用兜底 schema（其他类型）
GENERIC_SCHEMA = [
    ("title", "标题"),
    ("org", "相关单位"),
    ("date", "关键日期"),
    ("amount", "关键金额"),
]

PROMPT_TMPL = """你是售电公司的文档信息提取助手。从下面的文档内容中提取指定字段。

要求：
1. 只输出一个 JSON 对象，不要任何其他文字、解释或 markdown 代码块标记。
2. JSON 的 key 固定为：{keys}
3. 每个 key 的值是对象：{{"value": "提取到的原文值", "unit": "单位(没有则空字符串)", "confidence": 0到1的小数}}
4. 找不到的字段，value 给 null，confidence 给 0。
5. 数值保留原文写法（含千分位也照抄），日期统一为 YYYY-MM-DD。
6. 金额注意区分元/万元，unit 里写清楚。

字段说明：
{field_desc}

文档内容：
{text}
"""


def _call_text_model(prompt: str) -> str:
    """调用文本模型提取字段。优先 DeepSeek，未配置或失败时回退 GLM。"""
    # ── 优先：DeepSeek ──
    if DEEPSEEK_API_KEY:
        try:
            return _call_deepseek(prompt)
        except Exception as e:
            logger.warning(f"DeepSeek 调用失败，回退 GLM: {e}")

    # ── 回退：GLM 文本模型 ──
    if ZHIPU_API_KEY:
        return _call_glm_text(prompt)

    raise RuntimeError("未配置 DEEPSEEK_API_KEY 或 ZHIPU_API_KEY，无法调用文本模型")


def _call_deepseek(prompt: str) -> str:
    payload = {
        "model": DEEPSEEK_TEXT_MODEL,
        "messages": [{"role": "user", "content": prompt}],
        "temperature": 0.1,
        "max_tokens": 2048,
        "response_format": {"type": "json_object"},
    }
    headers = {"Authorization": f"Bearer {DEEPSEEK_API_KEY}", "Content-Type": "application/json"}
    with httpx.Client(timeout=90) as client:
        resp = client.post(f"{DEEPSEEK_BASE_URL}/chat/completions", json=payload, headers=headers)
        resp.raise_for_status()
        return resp.json()["choices"][0]["message"]["content"]


def _call_glm_text(prompt: str) -> str:
    if not ZHIPU_API_KEY:
        raise RuntimeError("未配置 ZHIPU_API_KEY，无法调用 GLM 文本模型")
    payload = {
        "model": ZHIPU_TEXT_MODEL,
        "messages": [{"role": "user", "content": prompt}],
        "temperature": 0.1,
        "max_tokens": 2048,
    }
    headers = {"Authorization": f"Bearer {ZHIPU_API_KEY}", "Content-Type": "application/json"}
    with httpx.Client(timeout=90) as client:
        resp = client.post(f"{ZHIPU_BASE_URL}/chat/completions", json=payload, headers=headers)
        resp.raise_for_status()
        return resp.json()["choices"][0]["message"]["content"]


def _parse_json_loose(s: str) -> dict:
    """容忍模型偶尔包 ```json 块或带前后缀文字。"""
    m = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", s, re.S)
    if m:
        s = m.group(1)
    else:
        m = re.search(r"\{.*\}", s, re.S)
        if m:
            s = m.group(0)
    return json.loads(s)


def extract_fields(text: str, doc_type: str) -> list[dict]:
    """按文档类型定向提取结构化字段。返回 [{key,label,value,unit,confidence}]。"""
    schema = FIELD_SCHEMAS.get(doc_type, GENERIC_SCHEMA)
    keys = ", ".join(k for k, _ in schema)
    field_desc = "\n".join(f"- {k}: {label}" for k, label in schema)
    # 控制提示词长度：超长文档截前 12000 字（关键信息一般在前部）
    prompt = PROMPT_TMPL.format(keys=keys, field_desc=field_desc, text=text[:12000])

    raw = _call_text_model(prompt)
    data = _parse_json_loose(raw)

    label_map = dict(schema)
    fields = []
    for k, label in schema:
        item = data.get(k)
        if not isinstance(item, dict):
            continue
        value = item.get("value")
        if value is None or str(value).strip() == "":
            continue
        fields.append({
            "key": k,
            "label": label_map.get(k, k),
            "value": str(value).strip(),
            "unit": str(item.get("unit") or "").strip(),
            "confidence": float(item.get("confidence") or 0.5),
        })
    return fields


def extract_fields_with_fallback(text: str, doc_type: str) -> list[dict]:
    """GLM 提取失败时回退正则实体（散值；confidence=0.3 提示人工复核）。"""
    try:
        fields = extract_fields(text, doc_type)
        if fields:
            return fields
    except Exception as e:
        logger.warning(f"GLM 结构化提取失败，回退正则: {e}")

    entities = extract_entities(text)
    fields = []
    for etype, values in entities.items():
        for i, v in enumerate(values[:5]):
            fields.append({
                "key": f"regex_{etype}_{i+1}",
                "label": etype,
                "value": v,
                "unit": "",
                "confidence": 0.3,
            })
    return fields
