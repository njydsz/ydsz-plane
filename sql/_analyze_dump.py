#!/usr/bin/env python3
"""分析 ydsz-plane-init.sql 中 COMMENT ON COLUMN 引用的列是否已在之前定义。

输出需要删除的 COMMENT 语句行号。
"""
import re
import sys

path = sys.argv[1] if len(sys.argv) > 1 else "sql/ydsz-plane-init.sql"

with open(path, encoding="utf-8") as f:
    lines = f.readlines()

# 已定义列集合： (table, column) -> 首次定义行号
defined = {}
# 当前 CREATE TABLE 上下文
create_tbl = None
create_cols = set()

# 收集所有 CREATE TABLE 的列 + ALTER TABLE ADD COLUMN 的列（不关心顺序，先全收）
alter_cols = {}  # table -> set
create_cols_map = {}  # table -> set

for i, line in enumerate(lines, 1):
    m = re.match(r'\s*CREATE TABLE "public"\."([^"]+)"', line)
    if m:
        create_tbl = m.group(1)
        create_cols_map.setdefault(create_tbl, set())
        continue
    if re.match(r'\s*\)\s*;', line):
        create_tbl = None
        continue
    if create_tbl:
        # 列定义: "col" type ...
        cm = re.match(r'\s*"([^"]+)"\s+', line)
        if cm and not line.strip().upper().startswith(("CONSTRAINT", "PRIMARY", "FOREIGN", "UNIQUE", "CHECK")):
            create_cols_map[create_tbl].add(cm.group(1))
        continue
    am = re.match(r'\s*ALTER TABLE public\.(\w+)\s+ADD COLUMN', line)
    if am:
        tbl = am.group(1)
        alter_cols.setdefault(tbl, set())
        # 提取列名
        cm2 = re.match(r'\s*ALTER TABLE public\.\w+\s+ADD COLUMN\s+(?:IF NOT EXISTS\s+)?["`]?(\w+)', line)
        if cm2:
            alter_cols[tbl].add(cm2.group(1))

# 合并某表所有列
def all_cols(tbl):
    return create_cols_map.get(tbl, set()) | alter_cols.get(tbl, set())

# 扫描 COMMENT ON COLUMN 顺序问题
create_tbl = None
seen_cols = {}  # table -> set of cols defined so far (in order)
problems = []
for i, line in enumerate(lines, 1):
    m = re.match(r'\s*CREATE TABLE "public"\."([^"]+)"', line)
    if m:
        create_tbl = m.group(1)
        seen_cols.setdefault(create_tbl, set())
        continue
    if re.match(r'\s*\)\s*;', line):
        create_tbl = None
        continue
    if create_tbl:
        cm = re.match(r'\s*"([^"]+)"\s+', line)
        if cm and not line.strip().upper().startswith(("CONSTRAINT", "PRIMARY", "FOREIGN", "UNIQUE", "CHECK")):
            seen_cols[create_tbl].add(cm.group(1))
        continue
    am = re.match(r'\s*ALTER TABLE public\.(\w+)\s+ADD COLUMN', line)
    if am:
        tbl = am.group(1)
        cm2 = re.match(r'\s*ALTER TABLE public\.\w+\s+ADD COLUMN\s+(?:IF NOT EXISTS\s+)?["`]?(\w+)', line)
        if cm2:
            seen_cols.setdefault(tbl, set()).add(cm2.group(1))
        continue
    cmt = re.match(r'\s*COMMENT ON COLUMN public\.(\w+)\.(\w+)\s+IS', line)
    if cmt:
        tbl, col = cmt.group(1), cmt.group(2)
        if col not in seen_cols.get(tbl, set()):
            problems.append((i, tbl, col))

print(f"需要删除的 COMMENT ON COLUMN (共 {len(problems)} 条):")
for i, tbl, col in problems:
    print(f"  L{i}: public.{tbl}.{col}")
