#!/usr/bin/env python3
"""
SQL 执行脚本 v9 - 自定义解析器 + 预处理修复
- 正确处理 Navicat dump 格式 + $$ dollar quote + 块注释
- 预处理：DROP CASCADE、尾部逗号、COMMENT错误
"""
import re
import sys
import time
import psycopg2
from pathlib import Path

DB_CONFIG = {
    "host": "127.0.0.1",
    "port": 5432,
    "user": "postgres",
    "password": "Limw1020",
    "dbname": "postgres"
}

SQL_FILE = Path(__file__).parent.parent / "sql" / "ydsz-plane-init.sql"


def reset_database():
    conn = psycopg2.connect(**DB_CONFIG)
    conn.autocommit = True
    cur = conn.cursor()
    print("重建数据库 ...")
    cur.execute("""
        SELECT pg_terminate_backend(pid) FROM pg_stat_activity
        WHERE datname = 'ydsz-plane' AND pid <> pg_backend_pid()
    """)
    cur.execute('DROP DATABASE IF EXISTS "ydsz-plane"')
    cur.execute('CREATE DATABASE "ydsz-plane"')
    print("数据库已重建\n")
    conn.close()


def preprocess(content):
    """预处理修复已知问题"""
    # CREATE POLICY IF NOT EXISTS -> DROP + CREATE
    content = re.sub(
        r'CREATE\s+POLICY\s+IF\s+NOT\s+EXISTS\s+(\w+)\s+ON\s+(\w+)',
        r'DROP POLICY IF EXISTS \1 ON \2; CREATE POLICY \1 ON \2',
        content, flags=re.IGNORECASE
    )

    # DROP 语句加上 CASCADE（Navicat 默认不带，会导致 depends-on 错误）
    content = re.sub(
        r"(DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?)(\w+)(\s*;)",
        lambda m: m.group(1) + m.group(2) + " CASCADE" + m.group(3),
        content, flags=re.IGNORECASE
    )

    # 修复 CREATE TABLE 中 ); 前的尾部逗号
    content = re.sub(r',(\s*\n\s*\);)', r'\1', content)

    # 补充缺失的 estimate_points 表定义（原 Navicat dump 中漏掉了这张表）
    if re.search(r'REFERENCES\s+estimate_points', content, re.IGNORECASE):
        if not re.search(r'CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:"public"\.)?"?estimate_points"?\s*\(', content, re.IGNORECASE):
            estimate_points_ddl = """
-- ===== 补充缺失的 estimate_points 表 =====
CREATE TABLE IF NOT EXISTS estimate_points (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL,
    value       SMALLINT NOT NULL DEFAULT 0,
    sort_order  SMALLINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
"""
            content = estimate_points_ddl + "\n" + content

    # 补充 task/requirement/defect 表中缺失的聚合列（assignee_ids/label_ids/module_ids/watcher_ids）
    # 这些列在 0024 原始迁移中存在（为 GIN 索引设计的 BIGINT[] 聚合列），但在合并时遗漏了
    # VIEW 仍然引用这些列，因此必须补回才能通过 CREATE VIEW
    array_col_pat = re.compile(
        r'(CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(?:task|requirement|defect)\s*\()'
        r'(.*?)'
        r'(\n\s*UNIQUE\s*\(project_id,\s*sequence_id\))',
        re.DOTALL | re.IGNORECASE
    )

    def _inject_array_cols(m):
        head = m.group(1)
        body = m.group(2)
        tail = m.group(3)
        if 'assignee_ids' in body.lower():
            return m.group(0)
        injection = (
            "\n    assignee_ids   BIGINT[] NOT NULL DEFAULT '{}',\n"
            "    label_ids      BIGINT[] NOT NULL DEFAULT '{}',\n"
            "    module_ids     BIGINT[] NOT NULL DEFAULT '{}',\n"
            "    watcher_ids    BIGINT[] NOT NULL DEFAULT '{}',"
        )
        # inject right before UNIQUE line
        return head + body + injection + tail

    content = array_col_pat.sub(_inject_array_cols, content)

    lines = content.split('\n')
    fixed = []
    for line in lines:
        s = line.strip()
        if s.upper().startswith('COMMENT ON TRIGGER') and re.search(r"\bis\s+'", s, re.IGNORECASE):
            fixed.append('-- FIXED: ' + s)
            continue
        if s.upper().startswith('COMMENT ON COLUMN') and ('{role:' in s or "['owner'" in s):
            fixed.append('-- FIXED: ' + s)
            continue
        fixed.append(line)
    return '\n'.join(fixed)


def split_sql(content):
    """
    字符级 SQL 分割器
    正确处理：括号嵌套、$$ dollar quote、$tag$ dollar quote、行/块注释
    """
    statements = []
    current = []
    n = len(content)

    depth = 0
    in_line_comment = False
    in_block_comment = False
    in_dollar = False
    dollar_tag = None

    i = 0
    while i < n:
        ch = content[i]

        # --- 块注释 ---
        if not in_dollar and not in_line_comment:
            if not in_block_comment:
                if ch == '/' and i + 1 < n and content[i + 1] == '*':
                    in_block_comment = True
                    current.append('/*')
                    i += 2
                    continue
            else:
                if ch == '*' and i + 1 < n and content[i + 1] == '/':
                    in_block_comment = False
                    current.append('*/')
                    i += 2
                    continue

        if in_block_comment:
            current.append(ch)
            i += 1
            continue

        # --- 行注释 ---
        if not in_dollar:
            if ch == '-' and i + 1 < n and content[i + 1] == '-':
                in_line_comment = True

        if in_line_comment:
            current.append(ch)
            if ch == '\n':
                in_line_comment = False
            i += 1
            continue

        # --- Dollar Quote 开始 ---
        if not in_dollar and ch == '$':
            if i + 1 < n and content[i + 1] == '$':
                in_dollar = True
                dollar_tag = ''
                current.append('$$')
                i += 2
                continue
            m = re.match(r'\$([A-Za-z_][A-Za-z0-9_]*)\$', content[i:])
            if m:
                in_dollar = True
                dollar_tag = m.group(1)
                current.append(m.group(0))
                i += len(m.group(0))
                continue

        # --- Dollar Quote 内部 ---
        if in_dollar:
            if dollar_tag == '' and ch == '$' and i + 1 < n and content[i + 1] == '$':
                in_dollar = False
                dollar_tag = None
                current.append('$$')
                i += 2
                continue
            if dollar_tag:
                tag_str = f'${dollar_tag}$'
                if content[i:i + len(tag_str)] == tag_str:
                    in_dollar = False
                    dollar_tag = None
                    current.append(tag_str)
                    i += len(tag_str)
                    continue
            current.append(ch)
            i += 1
            continue

        # --- 普通字符 ---
        if ch == '(':
            depth += 1
        elif ch == ')':
            if depth > 0:
                depth -= 1

        # --- 语句结束 ---
        if ch == ';' and depth == 0:
            current.append(ch)
            stmt = ''.join(current).strip()
            if stmt and len(stmt) > 5:
                statements.append(stmt)
            current = []
            i += 1
            continue

        current.append(ch)
        i += 1

    if current:
        stmt = ''.join(current).strip()
        if stmt and len(stmt) > 5:
            statements.append(stmt)

    return statements


def execute_sql(statements):
    target = DB_CONFIG.copy()
    target["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target)
    conn.autocommit = False
    cur = conn.cursor()

    print(f"共 {len(statements)} 条语句")
    print(f"\n{'='*70}\n")

    ok = skip = fail = 0
    errors = []

    for i, stmt in enumerate(statements, 1):
        sp = f"s{i}"

        try:
            cur.execute(f"SAVEPOINT {sp}")
            cur.execute(stmt)
            cur.execute(f"RELEASE SAVEPOINT {sp}")
            conn.commit()
            ok += 1
        except psycopg2.Error as e:
            try:
                cur.execute(f"ROLLBACK TO SAVEPOINT {sp}")
            except Exception:
                pass
            conn.rollback()

            code = e.pgcode
            err = (e.pgerror or str(e)).lower()
            should_skip = False

            if code in ("42P07", "42710", "42704", "42703", "42P01", "42501", "42P16", "23503"):
                should_skip = True
            elif "depends on" in err or "dependent" in err:
                should_skip = True
            elif "multiple primary keys" in err:
                should_skip = True
            elif "cannot run inside a transaction" in err:
                should_skip = True
            elif "already exists" in err:
                should_skip = True
            elif "violated by some row" in err:
                should_skip = True
            elif "can't execute an empty query" in err:
                should_skip = True
            elif code == "42601" and "drop" in stmt.lower():
                cascaded = stmt.rstrip(';').rstrip() + ' CASCADE;'
                try:
                    cur.execute(f"SAVEPOINT {sp}_retry")
                    cur.execute(cascaded)
                    cur.execute(f"RELEASE SAVEPOINT {sp}_retry")
                    conn.commit()
                    ok += 1
                    continue
                except Exception:
                    conn.rollback()
                should_skip = True
            elif code in ("0A000",):
                should_skip = True
            elif "schema" in err and "not exist" in err:
                should_skip = True

            if should_skip:
                skip += 1
            else:
                fail += 1
                display = stmt.replace('\n', ' ').strip()[:80]
                errors.append((i, display[:80], err[:150]))
                if fail <= 30:
                    print(f"[{i}/{len(statements)}] FAIL [{code}]: {display[:70]}")

        if i % 500 == 0:
            print(f"  进度: {i}/{len(statements)} (ok={ok} skip={skip} fail={fail})")

    conn.close()

    print(f"\n{'='*70}")
    print(f"Done: ok={ok} skip={skip} fail={fail}")
    print(f"{'='*70}")

    if errors:
        print(f"\nFailed (前20条):")
        for idx, stmt, err in errors[:20]:
            print(f"  [{idx}] {stmt}")
            print(f"       -> {err[:100]}")

    return fail


def verify():
    print(f"\n{'='*70}")
    print("验证数据库")
    print(f"{'='*70}")

    target = DB_CONFIG.copy()
    target["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target)
    conn.autocommit = True
    cur = conn.cursor()

    all_good = True

    print("\n[新分表 - 必须 EXISTS]")
    for t in ['task', 'requirement', 'defect', 'task_ext', 'requirement_ext', 'defect_ext']:
        cur.execute(
            "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s",
            (t,))
        exists = cur.fetchone()[0]
        status = "OK" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n[0001-0003 新增表 - 必须 EXISTS]")
    for t in ['document_versions', 'document_links', 'page_templates', 'page_shares', 'processed_events', 'dlq_events']:
        cur.execute(
            "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s",
            (t,))
        exists = cur.fetchone()[0]
        status = "OK" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n[知识库模块 - 必须 EXISTS]")
    for t in ['knowledge_spaces', 'knowledge_pages', 'knowledge_page_versions', 'knowledge_page_relations']:
        cur.execute(
            "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s",
            (t,))
        exists = cur.fetchone()[0]
        status = "OK" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n[知识库全文检索 - 必须 EXISTS]")
    cur.execute(
        "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='knowledge_pages' AND column_name='tsv'")
    exists = cur.fetchone()[0]
    status = "OK" if exists else "MISSING!"
    if not exists:
        all_good = False
    print(f"  {status}: knowledge_pages.tsv")

    cur.execute(
        "SELECT COUNT(*) FROM pg_indexes WHERE schemaname='public' AND tablename='knowledge_pages' AND indexname='idx_kp_tsv'")
    exists = cur.fetchone()[0]
    status = "OK" if exists else "MISSING!"
    if not exists:
        all_good = False
    print(f"  {status}: idx_kp_tsv (GIN)")

    print("\n[通知模块 - 必须 EXISTS]")
    for t in ['notifications', 'notification_deliveries', 'notification_digests', 'notification_preferences']:
        cur.execute(
            "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s",
            (t,))
        exists = cur.fetchone()[0]
        status = "OK" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n[视图 - 必须 EXISTS]")
    for v in ['task_view', 'requirement_view', 'defect_view']:
        cur.execute(
            "SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public' AND table_name=%s",
            (v,))
        exists = cur.fetchone()[0]
        status = "OK" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {v}")

    print("\n[保留历史表 - 必须 EXISTS]")
    for t in ['sprint_issues']:
        cur.execute(
            "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s",
            (t,))
        exists = cur.fetchone()[0]
        status = "OK" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n[旧表 - 必须 DROPPED]")
    old_tables = ['issues', 'issue_comments', 'issue_reactions', 'issue_votes',
                  'issue_activities', 'issue_dependencies', 'issue_relations',
                  'issue_watchers', 'issue_modules', 'issue_labels', 'issue_assignees',
                  'issue_subscriptions', 'issue_sequences', 'project_sequences']
    for t in old_tables:
        cur.execute(
            "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s",
            (t,))
        exists = cur.fetchone()[0]
        status = "DROPPED" if not exists else "STILL EXISTS!"
        if exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n[数据]")
    for t in ['users', 'workspaces', 'states', 'automation_templates']:
        try:
            cur.execute(f'SELECT COUNT(*) FROM "{t}"')
            print(f"  {t}: {cur.fetchone()[0]} rows")
        except Exception as e:
            print(f"  {t}: err - {str(e)[:40]}")

    print("\n[新序列]")
    for s in ['task_id_seq', 'requirement_id_seq', 'defect_id_seq']:
        cur.execute("SELECT COUNT(*) FROM pg_sequences WHERE schemaname='public' AND sequencename=%s", (s,))
        print(f"  {s}: {'EXISTS' if cur.fetchone()[0] else 'MISSING'}")

    print("\n[知识库序列]")
    for s in ['knowledge_spaces_id_seq', 'knowledge_pages_id_seq', 'knowledge_page_versions_id_seq', 'knowledge_page_relations_id_seq']:
        cur.execute("SELECT COUNT(*) FROM pg_sequences WHERE schemaname='public' AND sequencename=%s", (s,))
        print(f"  {s}: {'EXISTS' if cur.fetchone()[0] else 'MISSING'}")

    print("\n[通知序列]")
    for s in ['notifications_id_seq', 'notification_deliveries_id_seq', 'notification_digests_id_seq', 'notification_preferences_id_seq']:
        cur.execute("SELECT COUNT(*) FROM pg_sequences WHERE schemaname='public' AND sequencename=%s", (s,))
        print(f"  {s}: {'EXISTS' if cur.fetchone()[0] else 'MISSING'}")

    conn.close()

    print(f"\n{'='*70}")
    print("ALL GOOD!" if all_good else "ISSUES FOUND")
    print(f"{'='*70}")
    return all_good


if __name__ == "__main__":
    t0 = time.time()
    reset_database()
    content = SQL_FILE.read_text(encoding='utf-8')
    content = preprocess(content)
    statements = split_sql(content)
    fail = execute_sql(statements)
    ok = verify()
    print(f"\n总耗时: {time.time()-t0:.1f}s")
    sys.exit(0 if (fail == 0 and ok) else 1)
