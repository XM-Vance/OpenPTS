"""Alertmanager → 飞书机器人 webhook 适配器。

环境变量：
  FEISHU_WEBHOOK_WARNING   普通告警飞书机器人 URL
  FEISHU_WEBHOOK_CRITICAL  严重告警飞书机器人 URL（可与 warning 同一个）
  PORT                     监听端口（默认 8000）

接收 Alertmanager POST 后，按告警 severity 渲染富文本卡片再转发。
"""

from __future__ import annotations

import json
import os
from datetime import datetime
from typing import Any

import httpx
from fastapi import FastAPI, HTTPException, Request

FEISHU_WARNING = os.environ.get("FEISHU_WEBHOOK_WARNING", "")
FEISHU_CRITICAL = os.environ.get("FEISHU_WEBHOOK_CRITICAL", FEISHU_WARNING)

app = FastAPI(title="alert-webhook-proxy")


COLOR_BY_SEVERITY = {
    "critical": "red",
    "warning": "orange",
    "info": "blue",
}


def render_card(payload: dict[str, Any]) -> dict[str, Any]:
    """把 Alertmanager 告警批转成飞书 v2 卡片。"""
    alerts = payload.get("alerts", [])
    status = payload.get("status", "firing")
    common = payload.get("commonLabels", {})
    severity = common.get("severity", "warning")
    color = COLOR_BY_SEVERITY.get(severity, "blue")
    state_label = "已恢复" if status == "resolved" else "告警中"
    title_emoji = "✅" if status == "resolved" else ("🔥" if severity == "critical" else "⚠️")

    lines: list[str] = []
    for a in alerts:
        ann = a.get("annotations", {})
        lab = a.get("labels", {})
        started = a.get("startsAt", "")[:19].replace("T", " ")
        lines.append(
            f"**[{lab.get('alertname','?')}]** ({lab.get('severity','?')})\n"
            f"- 摘要: {ann.get('summary','-')}\n"
            f"- 描述: {ann.get('description','-')}\n"
            f"- 服务: {lab.get('service','-')} · 实例: {lab.get('instance','-')}\n"
            f"- 开始: {started}"
        )

    text_body = "\n\n---\n\n".join(lines) or "（无告警明细）"

    return {
        "msg_type": "interactive",
        "card": {
            "config": {"wide_screen_mode": True},
            "header": {
                "title": {
                    "tag": "plain_text",
                    "content": f"{title_emoji} PTIS 监控 · {state_label}",
                },
                "template": color,
            },
            "elements": [
                {
                    "tag": "markdown",
                    "content": text_body,
                },
                {
                    "tag": "note",
                    "elements": [
                        {
                            "tag": "plain_text",
                            "content": f"alertmanager • {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
                        }
                    ],
                },
            ],
        },
    }


async def forward(target: str, body: dict[str, Any]) -> dict[str, Any]:
    if not target:
        return {"ok": False, "reason": "webhook URL 未配置"}
    async with httpx.AsyncClient(timeout=5.0) as client:
        resp = await client.post(target, json=body)
        return {"ok": resp.is_success, "status": resp.status_code, "body": resp.text[:300]}


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/feishu/warning")
async def warning(req: Request) -> dict[str, Any]:
    body = await req.json()
    card = render_card(body)
    return await forward(FEISHU_WARNING, card)


@app.post("/feishu/critical")
async def critical(req: Request) -> dict[str, Any]:
    body = await req.json()
    card = render_card(body)
    return await forward(FEISHU_CRITICAL, card)


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port)
