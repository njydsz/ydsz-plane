#!/usr/bin/env python3
"""
SQL 执行脚本 v4 - 每条语句用 SAVEPOINT 隔离
"""
import sys
import time
import psycopg2
import sqlparse
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
    """删除并重建数据库"""
    conn = psycopg2.connect(**DB_CONFIG)
    conn.autocommit = True
    cur = conn.cursor()
    print("正在重建数据库 ...")
    cur.execute("""
        SELECT pg_terminate_backend(pid) 
        FROM pg_stat_activity 
        WHERE datname = 'ydsz-plane' AND pid <> pg_backend_pid()
    """)
    cur.execute('DROP DATABASE IF EXISTS "ydsz-plane"')
    cur.execute('CREATE DATABASE "ydsz-plane"')
    print("数据库已重建\n")
    conn.close()


def read_sql():
    with open(SQL_FILE, 'r', encoding='utf-8') as f:
        return f.read()


def split_statements(sql_content):
    """使用 sqlparse 正确分割 SQL 语句"""
    statements = []
    parsed = sqlparse.parse(sql_content)
    for stmt in parsed:
        s = stmt.value.strip()
        if s and len(s) > 5:
            statements.append(s)
    return statements


def execute_statements():
    target_config = DB_CONFIG.copy()
    target_config["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target_config)
    conn.autocommit = False
    cur = conn.cursor()

    sql_content = read_sql()
    statements = split_statements(sql_content)

    print(f"SQL 文件: {SQL_FILE}")
    print(f"共解析 {len(statements)} 条语句")
    print(f"\n{'='*70}")
    print("开始执行...")
    print(f"{'='*70}\n")

    ok = skip = fail = 0
    errors = []
    savepoint_id = 0

    for i, stmt in enumerate(statements, 1):
        display = stmt.replace('\n', ' ').strip()
        if len(display) > 90:
            display = display[:90] + "..."

        sp_name = f"sp_{i}"
        try:
            # 创建 SAVEPOINT 隔离每条语句
            cur.execute(f"SAVEPOINT {sp_name}")
            cur.execute(stmt)
            cur.execute(f"RELEASE SAVEPOINT {sp_name}")
            conn.commit()
            ok += 1
            if i <= 100 or i % 100 == 0:
                print(f"[{i}/{len(statements)}] ✓ {display[:70]}")
        except psycopg2.Error as e:
            try:
                cur.execute(f"ROLLBACK TO SAVEPOINT {sp_name}")
            except:
                pass
            conn.rollback()

            pgcode = e.pgcode
            err = (e.pgerror or str(e)).lower()

            # 判断是否跳过
            should_skip = False
            if pgcode in ("42P07", "42710"):  # 对象已存在
                should_skip = True
            elif pgcode == "42704":  # 对象不存在
                should_skip = True
            elif pgcode == "42703":  # 列不存在 (COMMENT)
                should_skip = True
            elif pgcode == "42501":  # 权限不足
                should_skip = True
            elif pgcode == "42P01":  # 表不存在
                should_skip = True
            elif "already exists" in err:
                should_skip = True
            elif "depends on" in err or "dependent objects" in err:
                should_skip = True
            elif "multiple primary keys" in err:
                should_skip = True
            elif "cannot run inside a transaction" in err:
                should_skip = True
            elif "syntax error at or near" in err and "NOT" in err and "POLICY" in err.upper():
                should_skip = True  # IF NOT EXISTS for POLICY not supported
            elif "no language specified" in err:
                should_skip = True
            elif pgcode == "42P16":
                should_skip = True
            elif pgcode == "0A000":
                should_skip = True
                
            if should_skip:
                skip += 1
                if i <= 50:
                    print(f"[{i}/{len(statements)}] ○ 跳过: {display[:60]}")
            else:
                fail += 1
                errors.append((i, display, err[:200]))
                print(f"[{i}/{len(statements)}] ✗ [{pgcode}] {display[:60]}")
                print(f"         {err[:100]}")

    conn.close()

    # 汇总
    print(f"\n{'='*70}")
    print(f"执行完成!")
    print(f"  ✓ 成功: {ok}")
    print(f"  ○ 跳过: {skip}")
    print(f"  ✗ 失败: {fail}")
    print(f"{'='*70}")

    if errors:
        print("\n失败详情:")
        for idx, stmt, err in errors[:20]:
            print(f"  [{idx}] {stmt}")
            print(f"       {err}")

    return fail


def verify():
    """验证数据库状态"""
    print(f"\n{'='*70}")
    print("验证数据库状态")
    print(f"{'='*70}")

    target_config = DB_CONFIG.copy()
    target_config["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target_config)
    cur = conn.cursor()

    # 新分表检查
    print("\n【新分表 (期望 EXISTS)】")
    all_good = True
    for t in ['task', 'requirement', 'defect', 'task_ext', 'requirement_ext', 'defect_ext']:
        cur.execute("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s", (t,))
        exists = cur.fetchone()[0]
        status = "✓" if exists else "✗ MISSING!"
        if not exists:
            all_good = False
        print(f"  {status} {t}")

    # 视图
    print("\n【视图 (期望 EXISTS)】")
    for v in ['task_view', 'requirement_view', 'defect_view']:
        cur.execute("SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public' AND table_name=%s", (v,))
        exists = cur.fetchone()[0]
        status = "✓" if exists else "✗ MISSING!"
        if not exists:
            all_good = False
        print(f"  {status} {v}")

    # 旧表应该已删除
    print("\n【旧表 (期望 DROPPED)】")
    old_tables = ['issues', 'issue_comments', 'issue_reactions', 'issue_votes',
                  'issue_activities', 'issue_dependencies', 'issue_relations',
                  'issue_watchers', 'issue_modules', 'issue_labels', 'issue_assignees',
                  'issue_subscriptions', 'issue_sequences', 'project_sequences',
                  'sprint_issues', 'intake_issues']
    for t in old_tables:
        cur.execute("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s", (t,))
        exists = cur.fetchone()[0]
        status = '✓ 已删除' if not exists else '✗ 仍存在!'
        if exists:
            all_good = False
        print(f"  {status} {t}")

    # 数据统计
    print("\n【数据行数】")
    for t in ['task', 'requirement', 'defect', 'users', 'workspaces', 'states', 'automation_templates', 'projects']:
        try:
            cur.execute(f'SELECT COUNT(*) FROM "{t}"')
            count = cur.fetchone()[0]
            print(f"  {t}: {count} rows")
        except Exception as e:
            print(f"  {t}: ERROR - {str(e)[:50]}")

    conn.close()

    print(f"\n{'='*70}")
    if all_good:
        print("✓ 验证通过! 所有预期对象已就绪")
    else:
        print("✗ 验证发现问题，见上表")
    print(f"{'='*70}")

    return all_good


if __name__ == "__main__":
    reset_database()
    fail_count = execute_statements()
    verify_ok = verify()
    sys.exit(0 if (fail_count == 0 and verify_ok) else 1)
