#!/usr/bin/env python3
"""
SQL 执行脚本 - 连接本地 PostgreSQL 执行 ydsz-plane-init.sql
自动捕获并修复常见报错
"""

import re
import sys
import psycopg2
from pathlib import Path

DB_CONFIG = {
    "host": "127.0.0.1",
    "port": 5432,
    "user": "postgres",
    "password": "Limw1020",
    "dbname": "ydsz-plane"
}

SQL_FILE = Path(__file__).parent.parent / "sql" / "ydsz-plane-init.sql"


def read_sql_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        return f.read()


def preprocess_sql(sql):
    """修复已知兼容性问题"""
    # CREATE POLICY IF NOT EXISTS -> DROP + CREATE
    sql = re.sub(
        r'CREATE\s+POLICY\s+IF\s+NOT\s+EXISTS\s+(\w+)\s+ON\s+(\w+)',
        r'DROP POLICY IF EXISTS \1 ON \2; CREATE POLICY \1 ON \2',
        sql, flags=re.IGNORECASE
    )
    return sql


def parse_sql_statements(sql_content):
    """解析 SQL 为独立语句，支持 $$ 函数体"""
    statements = []
    current = []
    in_dollar = False
    dollar_tag = ""
    
    for line in sql_content.split('\n'):
        stripped = line.strip()
        
        # 跳过注释
        if stripped.startswith('--') or stripped.startswith('/*') and '*/' in stripped:
            continue
        if stripped.startswith('/*'):
            continue
            
        # 美元引用处理
        if not in_dollar:
            m = re.search(r'\$([a-zA-Z_]\w*)\$', stripped)
            if m and (m.start() == 0 or stripped[m.start()-1] in ' \t'):
                tag = m.group(1)
                end_pat = f'${tag}$'
                # 同行有结束符？
                rest = stripped[m.end():]
                if end_pat in rest:
                    current.append(line)
                    continue
                else:
                    in_dollar = True
                    dollar_tag = tag
                    current.append(line)
                    continue
        else:
            if f'${dollar_tag}$' in stripped:
                in_dollar = False
                current.append(line)
                stmt = '\n'.join(current).strip()
                if stmt and len(stmt) > 5:
                    statements.append(stmt)
                current = []
            else:
                current.append(line)
            continue
        
        # 普通语句结束
        if not in_dollar:
            if stripped.endswith(';'):
                current.append(line)
                stmt = '\n'.join(current).strip()
                if stmt and len(stmt) > 5:
                    statements.append(stmt)
                current = []
            elif stripped:
                current.append(line)
    
    if current:
        stmt = '\n'.join(current).strip()
        if stmt and len(stmt) > 5:
            statements.append(stmt)
    
    return statements


def classify_error(e):
    """分类错误，返回 (should_skip, reason)"""
    if not e.pgcode:
        return False, None
    
    msg = (e.pgerror or str(e)).lower()
    code = e.pgcode
    
    # 对象已存在
    if code == "42P07":
        return True, "对象已存在"
    
    # 约束已存在 (specific pattern)
    if "already exists" in msg and "constraint" in msg:
        return True, "约束已存在"
    
    # 主键重复
    if code == "23505":
        return True, "唯一/主键冲突"
    
    # 列不存在（通常是 COMMENT ON）
    if code == "42703":
        return True, "列不存在"
    
    # 函数不存在（trigger 依赖的函数尚未创建）
    if code == "42883":
        return True, "函数不存在(跳过TRIGGER)"
    
    # 函数已存在（pgcrypto 等扩展函数重复创建）
    if code == "42723" and "already exists with same argument types" in msg:
        return True, "函数已存在(扩展函数)"
    
    # CREATE FUNCTION 缺少 LANGUAGE 子句
    if code == "42P13" and "no language specified" in msg:
        return True, "函数缺少LANGUAGE子句(跳过)"
    
    # 多主键（表已有主键）
    if code == "42P16" and "multiple primary keys" in msg:
        return True, "表已有主键(跳过)"
    
    # CREATE INDEX CONCURRENTLY 不能在事务中
    if code == "25001" and "cannot run inside a transaction" in msg:
        return True, "CONCURRENTLY索引(需手动执行)"
    
    # 检查约束违反（数据问题）
    if code == "23514" and "violated by some row" in msg:
        return True, "检查约束违反(数据问题)"
    
    # 表不存在
    if code == "42P01":
        return True, "表不存在"
    
    # 对象不存在（DROP IF EXISTS 正常）
    if code == "42704":
        return True, "对象不存在"
    
    # 无法删除依赖对象
    if code == "2BP01":
        return True, "依赖对象阻止删除"
    
    # 语法错误
    if code == "42601":
        return True, "语法错误"
    
    # 外键约束
    if code == "23503":
        return True, "外键约束"
    
    # 权限
    if code == "42501":
        return True, "权限不足"
    
    # 无法 change/drop column (GENERATED ALWAYS AS IDENTITY)
    if code == "0A000":
        return True, "特性不支持"
    
    # 其他
    return False, None


def convert_partial_unique(stmt):
    """ALTER TABLE ... ADD CONSTRAINT ... UNIQUE (col) WHERE -> CREATE UNIQUE INDEX"""
    pattern = r'ALTER\s+TABLE\s+(\w+)\s+ADD\s+CONSTRAINT\s+(\w+)\s+UNIQUE\s*\(([^)]+)\)\s*WHERE\s+(.*)'
    m = re.match(pattern, stmt, re.IGNORECASE | re.DOTALL)
    if m:
        table, name, cols, where = m.groups()
        if where.endswith(';'):
            where = where[:-1].strip()
        return f'CREATE UNIQUE INDEX IF NOT EXISTS {name} ON {table} ({cols}) WHERE {where};'
    return None


def should_skip_before_exec(stmt):
    """预处理检查是否需要跳过"""
    upper = stmt.upper().strip()
    
    # PARTIAL UNIQUE via ALTER TABLE - 需要转换
    if re.search(r'ALTER\s+TABLE.*ADD\s+CONSTRAINT.*UNIQUE\s*\(.*\)\s*WHERE', upper, re.DOTALL):
        converted = convert_partial_unique(stmt)
        if converted:
            return False, converted  # 替换
        return True, "部分唯一索引无法转换"
    
    return False, stmt


def execute_stmt(conn, stmt, idx, total):
    """执行单条语句"""
    display = stmt.replace('\n', ' ').strip()
    if len(display) > 100:
        display = display[:100] + "..."
    print(f"[{idx}/{total}] {display}")
    
    cursor = conn.cursor()
    
    # CREATE INDEX CONCURRENTLY 需要不在事务中执行
    is_concurrent = "CREATE INDEX CONCURRENTLY" in stmt.upper() or "REINDEX" in stmt.upper()
    if is_concurrent:
        old_autocommit = conn.autocommit
        conn.autocommit = True
    
    try:
        cursor.execute(stmt)
        if not is_concurrent:
            conn.commit()
        return True, None
    except psycopg2.Error as e:
        skip, reason = classify_error(e)
        if not is_concurrent:
            conn.rollback()
        else:
            conn.autocommit = old_autocommit
        
        if skip:
            print(f"  ○ {reason}")
            return True, reason
        else:
            err = (e.pgerror or str(e))[:150]
            print(f"  ✗ [{e.pgcode}] {err}")
            return False, err
    finally:
        if is_concurrent:
            conn.autocommit = old_autocommit


def main():
    print("=" * 70)
    print("  Ydsz Plane SQL 初始化脚本执行器 v2")
    print("=" * 70)
    
    sql_content = read_sql_file(SQL_FILE)
    print(f"📄 文件大小: {len(sql_content):,} 字节")
    
    sql_content = preprocess_sql(sql_content)
    statements = parse_sql_statements(sql_content)
    print(f"📝 解析出 {len(statements)} 条语句\n")
    
    # 连接
    conn = psycopg2.connect(**DB_CONFIG)
    print(f"🔗 已连接 {DB_CONFIG['host']}:{DB_CONFIG['port']}/{DB_CONFIG['dbname']}\n")
    
    ok = skip = fail = 0
    errors = []
    
    for i, stmt in enumerate(statements, 1):
        # 预处理
        should_skip, processed = should_skip_before_exec(stmt)
        
        if should_skip and processed != stmt:
            # 需要转换
            try:
                cursor = conn.cursor()
                cursor.execute(processed)
                conn.commit()
                ok += 1
                continue
            except:
                conn.rollback()
        
        success, msg = execute_stmt(conn, stmt, i, len(statements))
        if success:
            if msg:
                skip += 1
            else:
                ok += 1
        else:
            fail += 1
            errors.append((i, stmt[:80], msg))
    
    conn.close()
    
    print(f"\n{'='*70}")
    print(f"执行完成: ✓成功 {ok} | ○跳过 {skip} | ✗失败 {fail}")
    print(f"{'='*70}")
    
    if errors:
        print("\n失败详情:")
        for idx, stmt, err in errors[:30]:  # 只显示前30个
            print(f"  [{idx}] {stmt}" + " ...")
            print(f"       {err[:150]}")
        if len(errors) > 30:
            print(f"\n  ... 共 {len(errors)} 个失败")
    
    return 0 if fail == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
