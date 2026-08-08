#!/usr/bin/env python3
"""
SQL 执行脚本 v3 - 使用 sqlparse 正确解析 SQL
"""
import sys
import psycopg2
import sqlparse
from pathlib import Path

DB_CONFIG = {
    "host": "127.0.0.1",
    "port": 5432,
    "user": "postgres",
    "password": "Limw1020",
    "dbname": "postgres"  # 连接到 postgres 数据库来管理 ydsz-plane
}

SQL_FILE = Path(__file__).parent.parent / "sql" / "ydsz-plane-init.sql"


def reset_database(conn):
    """删除并重建数据库"""
    conn.autocommit = True
    cursor = conn.cursor()
    print("正在重建数据库 ydsz-plane ...")
    # 先终止其他连接
    cursor.execute("""
        SELECT pg_terminate_backend(pid) 
        FROM pg_stat_activity 
        WHERE datname = 'ydsz-plane' AND pid <> pg_backend_pid()
    """)
    cursor.execute("DROP DATABASE IF EXISTS \"ydsz-plane\"")
    cursor.execute("CREATE DATABASE \"ydsz-plane\"")
    print("数据库已重建\n")


def read_sql_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        return f.read()


def split_statements(sql_content):
    """使用 sqlparse 正确分割 SQL 语句"""
    statements = []
    parsed = sqlparse.parse(sql_content)
    for stmt in parsed:
        sql = stmt.value.strip()
        if sql:
            statements.append(sql)
    return statements


def execute_all():
    print("=" * 70)
    print("  Ydsz Plane SQL 执行器 v3 (使用 sqlparse)")
    print("=" * 70)

    # 读取 SQL
    sql_content = read_sql_file(SQL_FILE)
    print(f"\nSQL 文件大小: {len(sql_content):,} 字节")

    # 解析语句
    print("解析 SQL 语句 (sqlparse) ...")
    statements = split_statements(sql_content)
    print(f"共解析 {len(statements)} 条语句")

    # 先连接到 postgres 数据库重建目标库
    conn = psycopg2.connect(**DB_CONFIG)
    reset_database(conn)
    conn.close()

    # 连接到目标数据库
    target_config = DB_CONFIG.copy()
    target_config["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target_config)
    conn.autocommit = True  # DDL 需要 autocommit

    cursor = conn.cursor()

    ok = skip = fail = 0
    errors = []

    print(f"\n{'='*70}")
    print("开始执行...")
    print(f"{'='*70}\n")

    for i, stmt in enumerate(statements, 1):
        # 显示
        display = stmt.replace('\n', ' ').strip()
        if len(display) > 90:
            display = display[:90] + "..."

        try:
            cursor.execute(stmt)
            ok += 1
            print(f"[{i}/{len(statements)}] ✓ {display}")
        except psycopg2.Error as e:
            err = e.pgerror or str(e)
            # 某些错误直接跳过
            if e.pgcode in ("42P07", "42704", "42703", "42501", "42P01", "42P16"):
                skip += 1
                print(f"[{i}/{len(statements)}] ○ 跳过: {display[:60]}...")
            elif "already exists" in err.lower():
                skip += 1
                print(f"[{i}/{len(statements)}] ○ 已存在: {display[:60]}...")
            elif "depends on" in err.lower() or "dependent" in err.lower():
                # 尝试 CASCADE 后重试
                try:
                    cascaded = stmt.rstrip(';') + ' CASCADE;'
                    if cascaded.upper().startswith('DROP'):
                        cursor.execute(cascaded)
                        ok += 1
                        print(f"[{i}/{len(statements)}] ✓ CASCADE重试成功")
                    else:
                        skip += 1
                        print(f"[{i}/{len(statements)}] ○ 依赖跳过: {display[:60]}...")
                except:
                    skip += 1
                    print(f"[{i}/{len(statements)}] ○ CASCADE也失败: {display[:60]}...")
            elif "cannot run inside a transaction" in err.lower():
                skip += 1
                print(f"[{i}/{len(statements)}] ○ 事务限制: {display[:60]}...")
            else:
                fail += 1
                errors.append((i, display, err[:200]))
                print(f"[{i}/{len(statements)}] ✗ [{e.pgcode}] {err[:100]}")

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
        for idx, stmt, err in errors:
            print(f"  [{idx}] {stmt}")
            print(f"       {err}")

    return fail


def verify_database():
    """验证数据库状态"""
    print(f"\n{'='*70}")
    print("验证数据库状态")
    print(f"{'='*70}")

    target_config = DB_CONFIG.copy()
    target_config["dbname"] = "ydsz-plane"
    conn = psycopg2.connect(**target_config)
    cursor = conn.cursor()

    # 新分表检查
    required_tables = ["task", "requirement", "defect", "users", "workspaces", "projects", "states"]
    print("\n【新分表】")
    for t in required_tables:
        cursor.execute(f"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='{t}'")
        exists = cursor.fetchone()[0]
        status = "✓" if exists else "✗"
        print(f"  {status} {t}")

    # 旧表应该被删除
    old_tables = ["issues", "issue_comments", "issue_reactions", "issue_votes", "issue_activities"]
    print("\n【旧表（应为空）】")
    for t in old_tables:
        cursor.execute(f"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='{t}'")
        exists = cursor.fetchone()[0]
        status = "✗ 仍存在!" if exists else "✓ 已删除"
        print(f"  {status} {t}")

    # 视图检查
    required_views = ["task_view", "requirement_view", "defect_view"]
    print("\n【视图】")
    for v in required_views:
        cursor.execute(f"SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public' AND table_name='{v}'")
        exists = cursor.fetchone()[0]
        status = "✓" if exists else "✗"
        print(f"  {status} {v}")

    # 数据统计
    print("\n【数据行数】")
    data_tables = ["task", "requirement", "defect", "users", "workspaces", "states", "automation_templates"]
    for t in data_tables:
        try:
            cursor.execute(f"SELECT COUNT(*) FROM {t}")
            count = cursor.fetchone()[0]
            print(f"  {t}: {count} 行")
        except:
            print(f"  {t}: (查询失败)")

    conn.close()


if __name__ == "__main__":
    fail_count = execute_all()
    verify_database()
    sys.exit(0 if fail_count == 0 else 1)
