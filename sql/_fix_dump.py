#!/usr/bin/env python3
"""删除 ydsz-plane-init.sql 中引用未定义列的 COMMENT ON COLUMN 语句。

这些 COMMENT 因为 Navicat 导出顺序问题（COMMENT 在列定义/ALTER 之前，或列根本
不存在）会导致 psql/migrate 执行失败。COMMENT 是纯元数据，删除不影响 schema 功能。
"""
import re
import sys

path = sys.argv[1] if len(sys.argv) > 1 else "sql/ydsz-plane-init.sql"

with open(path, encoding="utf-8") as f:
    lines = f.readlines()

# 阶段1：收集所有真正定义的列（CREATE TABLE 列 + ALTER ADD COLUMN 列），不关心顺序
create_cols_map = {}
alter_cols = {}
create_tbl = None
for line in lines:
    m = re.match(r'\s*CREATE TABLE "public"\."([^"]+)"', line)
    if m:
        create_tbl = m.group(1)
        create_cols_map.setdefault(create_tbl, set())
        continue
    if re.match(r'\s*\)\s*;', line):
        create_tbl = None
        continue
    if create_tbl:
        cm = re.match(r'\s*"([^"]+)"\s+', line)
        if cm and not line.strip().upper().startswith(("CONSTRAINT", "PRIMARY", "FOREIGN", "UNIQUE", "CHECK")):
            create_cols_map[create_tbl].add(cm.group(1))
        continue
    am = re.match(r'\s*ALTER TABLE public\.(\w+)\s+ADD COLUMN', line)
    if am:
        tbl = am.group(1)
        alter_cols.setdefault(tbl, set())
        cm2 = re.match(r'\s*ALTER TABLE public\.\w+\s+ADD COLUMN\s+(?:IF NOT EXISTS\s+)?["`]?(\w+)', line)
        if cm2:
            alter_cols[tbl].add(cm2.group(1))

def truly_defined(tbl, col):
    return col in create_cols_map.get(tbl, set()) or col in alter_cols.get(tbl, set())

# 阶段2：按文件顺序，维护已见列；COMMENT 引用的列若到此刻仍未定义 -> 删除
create_tbl = None
seen = {}
remove = set()
for i, line in enumerate(lines):
    m = re.match(r'\s*CREATE TABLE "public"\."([^"]+)"', line)
    if m:
        create_tbl = m.group(1)
        seen.setdefault(create_tbl, set())
        continue
    if re.match(r'\s*\)\s*;', line):
        create_tbl = None
        continue
    if create_tbl:
        cm = re.match(r'\s*"([^"]+)"\s+', line)
        if cm and not line.strip().upper().startswith(("CONSTRAINT", "PRIMARY", "FOREIGN", "UNIQUE", "CHECK")):
            seen[create_tbl].add(cm.group(1))
        continue
    am = re.match(r'\s*ALTER TABLE public\.(\w+)\s+ADD COLUMN', line)
    if am:
        tbl = am.group(1)
        cm2 = re.match(r'\s*ALTER TABLE public\.\w+\s+ADD COLUMN\s+(?:IF NOT EXISTS\s+)?["`]?(\w+)', line)
        if cm2:
            seen.setdefault(tbl, set()).add(cm2.group(1))
        continue
    cmt = re.match(r'\s*COMMENT ON COLUMN public\.(\w+)\.(\w+)\s+IS', line)
    if cmt:
        tbl, col = cmt.group(1), cmt.group(2)
        if not truly_defined(tbl, col) or col not in seen.get(tbl, set()):
            # 列根本不存在，或到此刻尚未定义（顺序问题）-> 删除 COMMENT
            remove.add(i)

kept = [l for i, l in enumerate(lines) if i not in remove]
with open(path, "w", encoding="utf-8") as f:
    f.writelines(kept)

print(f"已删除 {len(remove)} 条有问题的 COMMENT ON COLUMN 语句")
