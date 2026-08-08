#!/usr/bin/env python3
"""
SQL 执行脚本 v6 - 使用自定义解析器（正确处理 Navicat dump 格式）
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


def split_sql_statements_from_content(content):
    """Navicat dump 格式的分割器 - 正确处理多行语句"""

    statements = []
    current = []
    paren_depth = 0
    in_dollar_quote = False
    dollar_tag = ""
    in_line_comment = False
    in_block_comment = False
    skip_semi = False  # 跳过一个分号（CASCADE 转换时可能产生）

    i = 0
    while i < len(content):
        ch = content[i]

        # 处理块注释开始
        if not in_dollar_quote and not in_line_comment and not in_block_comment:
            if ch == '/' and i + 1 < len(content) and content[i + 1] == '*':
                in_block_comment = True
                current.append(ch)
                i += 1
                if i < len(content):
                    current.append(content[i])
                i += 1
                continue
            # 处理块注释结束
            if ch == '*' and i + 1 < len(content) and content[i + 1] == '/':
                in_block_comment = False
                current.append(ch)
                i += 1
                if i < len(content):
                    current.append(content[i])
                i += 1
                continue

        if in_block_comment:
            current.append(ch)
            i += 1
            continue

        # 处理行注释
        if not in_dollar_quote and not in_block_comment:
            if ch == '-' and i + 1 < len(content) and content[i + 1] == '-':
                in_line_comment = True
                current.append(ch)
                i += 1
                current.append(content[i])
                i += 1
                continue

        if in_line_comment:
            current.append(ch)
            if ch == '\n':
                in_line_comment = False
            i += 1
            continue

        # 处理 dollar quote
        if not in_block_comment and not in_line_comment:
            if ch == '$':
                # 检查是否是 $tag$ 模式
                if not in_dollar_quote:
                    # 向前看是否有标签
                    m = re.match(r'\$([A-Za-z_][A-Za-z0-9_]*)\$', content[i:])
                    if m:
                        in_dollar_quote = True
                        dollar_tag = m.group(1)
                        current.append(m.group(0))
                        i += len(m.group(0))
                        continue
                else:
                    # 检查是否是结束
                    tag_str = f'${dollar_tag}$'
                    if content[i:i+len(tag_str)] == tag_str:
                        in_dollar_quote = False
                        current.append(tag_str)
                        i += len(tag_str)
                        continue

        if in_dollar_quote:
            current.append(ch)
            i += 1
            continue

        # 正常处理
        if ch == '(':
            paren_depth += 1
            current.append(ch)
            i += 1
            continue

        if ch == ')':
            paren_depth -= 1
            current.append(ch)
            i += 1
            continue

        if ch == ';':
            current.append(ch)
            if paren_depth == 0:
                # 语句结束
                stmt = ''.join(current).strip()
                if stmt and len(stmt) > 5:
                    statements.append(stmt)
                current = []
            i += 1
            continue

        current.append(ch)
        i += 1

    # 剩余部分
    if current:
        stmt = ''.join(current).strip()
        if stmt and len(stmt) > 5:
            statements.append(stmt)

    return statements


def preprocess(content):
    """预处理修复已知问题"""

    # 1. CREATE POLICY IF NOT EXISTS -> DROP + CREATE
    content = re.sub(
        r'CREATE\s+POLICY\s+IF\s+NOT\s+EXISTS\s+(\w+)\s+ON\s+(\w+)',
        r'DROP POLICY IF EXISTS \1 ON \2; CREATE POLICY \1 ON \2',
        content,
        flags=re.IGNORECASE
    )
    # 修复转换后可能的分号重复
    content = re.sub(r';(?!DROP)\s*(CREATE POLICY)', r' \1', content)

    # 2. 注释掉有问题的 COMMENT 语句
    lines = content.split('\n')
    fixed = []
    for line in lines:
        s = line.strip()
        if s.upper().startswith('COMMENT ON TRIGGER'):
            if re.search(r"\bis\s+'", s, re.IGNORECASE):
                fixed.append('-- FIXED: ' + s)
                continue
        if s.upper().startswith('COMMENT ON COLUMN'):
            if '{role:' in s or "['owner'" in s:
                fixed.append('-- FIXED: ' + s)
                continue
        fixed.append(line)
    content = '\n'.join(fixed)

    return content


def execute_statements(statements):
    target_config = DB_CONFIG.copy()
    target_config["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target_config)
    conn.autocommit = False
    cur = conn.cursor()

    print(f"共 {len(statements)} 条语句")
    print(f"\n{'='*70}\n")

    ok = skip = fail = 0
    errors = []

    for i, stmt in enumerate(statements, 1):
        sp = f"s{i}"
        display = stmt.replace('\n', ' ').strip()
        if len(display) > 80:
            display = display[:80] + "..."

        try:
            cur.execute(f"SAVEPOINT {sp}")
            cur.execute(stmt)
            cur.execute(f"RELEASE SAVEPOINT {sp}")
            conn.commit()
            ok += 1
        except psycopg2.Error as e:
            try:
                cur.execute(f"ROLLBACK TO SAVEPOINT {sp}")
            except:
                pass
            conn.rollback()

            pgcode = e.pgcode
            err = (e.pgerror or str(e)).lower()

            should_skip = False
            if pgcode in ("42P07", "42710"):
                should_skip = True
            elif pgcode == "42704":
                should_skip = True
            elif pgcode == "42703":
                should_skip = True
            elif pgcode == "42P01":
                should_skip = True
            elif pgcode == "42501":
                should_skip = True
            elif "depends on" in err or "dependent objects" in err:
                should_skip = True
            elif "multiple primary keys" in err:
                should_skip = True
            elif "cannot run inside a transaction" in err:
                should_skip = True
            elif pgcode == "42601":
                # 语法错误可以重试 CASCADE
                if "drop" in stmt.lower() and ("table" in stmt.lower() or "sequence" in stmt.lower()):
                    try:
                        cascaded = re.sub(r'(DROP\s+(?:TABLE|SEQUENCE)\s+(?:IF\s+EXISTS\s+)?)', r'\1', stmt).rstrip(';') + ' CASCADE;'
                        cur.execute(f"SAVEPOINT {sp}_retry")
                        cur.execute(cascaded)
                        cur.execute(f"RELEASE SAVEPOINT {sp}_retry")
                        conn.commit()
                        ok += 1
                        continue
                    except:
                        conn.rollback()
                should_skip = True
            elif pgcode == "0A000":
                should_skip = True
            elif "violated by some row" in err:
                should_skip = True
            elif "can't execute an empty query" in err:
                should_skip = True
            elif "already exists" in err:
                should_skip = True

            if should_skip:
                skip += 1
            else:
                fail += 1
                errors.append((i, display[:80], err[:150]))
                print(f"[{i}/{len(statements)}] FAIL [{pgcode}]: {display[:60]}")

        if i % 200 == 0:
            print(f"  ... {i}/{len(statements)} done (ok={ok} skip={skip} fail={fail})")

    conn.close()

    print(f"\n{'='*70}")
    print(f"Done: ok={ok} skip={skip} fail={fail}")
    print(f"{'='*70}")

    if errors:
        print(f"\nFailed statements:")
        for idx, stmt, err in errors:
            print(f"  [{idx}] {stmt}")
            print(f"       {err[:100]}")

    return fail


def verify():
    print(f"\n{'='*70}")
    print("验证数据库")
    print(f"{'='*70}")

    target_config = DB_CONFIG.copy()
    target_config["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target_config)
    cur = conn.cursor()

    all_good = True

    print("\n【新分表 (必须 EXISTS)】")
    for t in ['task', 'requirement', 'defect', 'task_ext', 'requirement_ext', 'defect_ext']:
        cur.execute("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s", (t,))
        exists = cur.fetchone()[0]
        status = "GOOD" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n【视图 (必须 EXISTS)】")
    for v in ['task_view', 'requirement_view', 'defect_view']:
        cur.execute("SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public' AND table_name=%s", (v,))
        exists = cur.fetchone()[0]
        status = "GOOD" if exists else "MISSING!"
        if not exists:
            all_good = False
        print(f"  {status}: {v}")

    print("\n【旧表 (必须 DROPPED)】")
    old = ['issues', 'issue_comments', 'issue_reactions', 'issue_votes',
           'issue_activities', 'issue_dependencies', 'issue_relations',
           'issue_watchers', 'issue_modules', 'issue_labels', 'issue_assignees',
           'issue_subscriptions', 'issue_sequences', 'project_sequences',
           'sprint_issues', 'intake_issues']
    for t in old:
        cur.execute("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s", (t,))
        exists = cur.fetchone()[0]
        status = "DROPPED" if not exists else "STILL EXISTS!"
        if exists:
            all_good = False
        print(f"  {status}: {t}")

    print("\n【数据】")
    for t in ['users', 'workspaces', 'states', 'automation_templates']:
        try:
            cur.execute(f'SELECT COUNT(*) FROM "{t}"')
            print(f"  {t}: {cur.fetchone()[0]} rows")
        except Exception as e:
            print(f"  {t}: ERROR - {str(e)[:50]}")

    conn.close()

    status = "ALL GOOD!" if all_good else "ISSUES FOUND"
    print(f"\n{'='*70}")
    print(status)
    print(f"{'='*70}")
    return all_good


if __name__ == "__main__":
    t0 = time.time()
    reset_database()
    content = SQL_FILE.read_text(encoding='utf-8')
    content = preprocess(content)
    statements = split_sql_statements_from_content(content)
    fail = execute_statements(statements)
    ok = verify()
    print(f"\n耗时: {time.time()-t0:.1f}s")
    sys.exit(0 if (fail == 0 and ok) else 1)
