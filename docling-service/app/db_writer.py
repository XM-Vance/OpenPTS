"""
数据库写入器
解析结果写入 PTIS PostgreSQL 数据库
"""

import json, logging
from datetime import datetime
from typing import Optional

import psycopg2
import psycopg2.extras
from psycopg2 import pool as pgpool

logger = logging.getLogger("docling-service.db_writer")

# PTIS PG 连接配置（密码仅从环境变量读取，仓库不留真实凭证）
import os
PG_HOST = os.getenv("PG_HOST", "localhost")
PG_PORT = os.getenv("PG_PORT", "5432")
PG_DB = os.getenv("PG_DB", "ptis")
PG_USER = os.getenv("PG_USER", "ptis")
PG_PASSWORD = os.getenv("PG_PASSWORD", "")
PG_DSN = f"host={PG_HOST} port={PG_PORT} dbname={PG_DB} user={PG_USER} password={PG_PASSWORD}"

# 建表SQL（standalone 兜底；schema 唯一来源是 db/migrations/0063_docling_documents.up.sql，
# 二者同构。org_id 为 uuid，与主库多租户口径一致）
CREATE_TABLE_SQL = """
CREATE TABLE IF NOT EXISTS docling_documents (
    id          SERIAL PRIMARY KEY,
    org_id      uuid,
    filename    VARCHAR(512) NOT NULL,
    doc_type    VARCHAR(32),
    text_content TEXT,
    tables      JSONB,
    entities    JSONB,
    summary     TEXT,
    file_size   INTEGER,
    package_info JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_docling_doc_type ON docling_documents(doc_type);
CREATE INDEX IF NOT EXISTS idx_docling_org_id ON docling_documents(org_id);
"""


class DBWriter:
    def __init__(self):
        self._pool = None
        self._ensure_table()

    def _conn(self):
        """从连接池借一个连接；池延迟创建（PG 在启动期可能未就绪）。

        此前每次 _conn() 都 psycopg2.connect() 新建连接，高频写入时连接开销大。
        改用 ThreadedConnectionPool（minconn=1/maxconn=5）复用连接。
        """
        if self._pool is None:
            self._pool = pgpool.ThreadedConnectionPool(1, 5, PG_DSN)
        return self._pool.getconn()

    def _put_conn(self, conn):
        """归还连接到池（出错时回滚再归还，避免脏状态复用）。"""
        try:
            conn.rollback()
        except Exception:
            pass
        if self._pool is not None:
            self._pool.putconn(conn)

    def _ensure_table(self):
        conn = None
        try:
            conn = self._conn()
            with conn.cursor() as cur:
                cur.execute(CREATE_TABLE_SQL)
            conn.commit()
            logger.info("docling_documents 表就绪")
        except Exception as e:
            logger.warning(f"建表跳过（PG未就绪）: {e}")
            # PG 未就绪时重置池，下次 _conn 重新建池（避免持有坏池）
            self._reset_pool()
        finally:
            if conn is not None:
                self._put_conn(conn)

    def _reset_pool(self):
        """关闭并清空连接池（PG 重启/未就绪后调用，下次 _conn 重建）。"""
        if self._pool is not None:
            try:
                self._pool.closeall()
            except Exception:
                pass
            self._pool = None

    def upsert(self, filename: str, doc_type: str, text_content: str,
               tables: list, entities: dict, summary: str,
               org_id: Optional[str] = None, package_info: Optional[dict] = None):
        conn = None
        try:
            conn = self._conn()
            with conn.cursor() as cur:
                cur.execute("""
                    INSERT INTO docling_documents
                        (org_id, filename, doc_type, text_content, tables, entities,
                         summary, file_size, package_info)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                    RETURNING id
                """, (
                    org_id, filename, doc_type, text_content,
                    json.dumps(tables, ensure_ascii=False),
                    json.dumps(entities, ensure_ascii=False),
                    summary,
                    len(text_content.encode()),
                    json.dumps(package_info, ensure_ascii=False) if package_info else None,
                ))
                doc_id = cur.fetchone()[0]
            conn.commit()
            logger.info(f"写入文档 id={doc_id}: {filename} [{doc_type}] package={bool(package_info)}")
            return doc_id
        finally:
            if conn is not None:
                self._put_conn(conn)

    def list_docs(self, doc_type: Optional[str] = None, limit: int = 50,
                  org_id: Optional[str] = None):
        conn = None
        try:
            conn = self._conn()
            # RealDictCursor:返回 dict 而非 tuple,前端/网关拿到的是规整 JSON 对象
            with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
                sql = """
                    SELECT id, org_id, filename, doc_type, summary, created_at
                    FROM docling_documents
                    WHERE 1=1
                """
                params = []
                # org_id 给定时按省份隔离;为空表示总部「全部省」,不过滤
                if org_id:
                    sql += " AND org_id = %s"
                    params.append(org_id)
                if doc_type:
                    sql += " AND doc_type = %s"
                    params.append(doc_type)
                sql += " ORDER BY created_at DESC LIMIT %s"
                params.append(limit)
                cur.execute(sql, params)
                rows = cur.fetchall()
            conn.commit()
            return rows
        finally:
            if conn is not None:
                self._put_conn(conn)

    def get_doc(self, doc_id: int, org_id: Optional[str] = None):
        conn = None
        try:
            conn = self._conn()
            with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
                if org_id:
                    cur.execute(
                        "SELECT * FROM docling_documents WHERE id = %s AND org_id = %s",
                        (doc_id, org_id),
                    )
                else:
                    cur.execute("SELECT * FROM docling_documents WHERE id = %s", (doc_id,))
                row = cur.fetchone()
            conn.commit()
            return row
        finally:
            if conn is not None:
                self._put_conn(conn)
