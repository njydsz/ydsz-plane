#!/usr/bin/env python3
"""修复 ydsz-plane-init.sql 中的元数据语句，使其能干净初始化（无 ERROR）。

修复两类问题：
1. 删除引用"在 dump 中从未定义"列的 COMMENT ON COLUMN 语句（Navicat 导出残留）。
2. 修复 ALTER TABLE ... ADD COLUMN ... REFERENCES 未创建表的 FK（顺序问题导致运行失败），
   改为无 FK 的列定义（保持列存在，放弃跨表外键约束，因原 dump 本就未成功建立）。

COMMENT 与 ALTER 均为辅助性 DDL，删除/降级不影响核心表结构（CREATE TABLE）。
"""
import re

path = "sql/ydsz-plane-init.sql"
with open(path, encoding="utf-8") as f:
    lines = f.readlines()

# 阶段1：收集所有 CREATE TABLE 真实定义的列
create_cols = {}
create_tbl = None
for line in lines:
    m = re.match(r'\s*CREATE TABLE "public"\."([^"]+)"', line)
    if m:
        create_tbl = m.group(1)
        create_cols.setdefault(create_tbl, set())
        continue
    if re.match(r'\s*\)\s*;', line):
        create_tbl = None
        continue
    if create_tbl:
        cm = re.match(r'\s*"([^"]+)"\s+', line)
        if cm and not line.strip().upper().startswith(("CONSTRAINT", "PRIMARY", "FOREIGN", "UNIQUE", "CHECK")):
            create_cols[create_tbl].add(cm.group(1))

# 阶段2：逐行处理
out = []
removed_comment = 0
fixed_alter = 0
for line in lines:
    # (1) 删除引用未定义列的 COMMENT ON COLUMN（列在 CREATE TABLE 中不存在即删）
    cmt = re.match(r'\s*COMMENT ON COLUMN public\.(\w+)\.(\w+)\s+IS', line)
    if cmt:
        tbl, col = cmt.group(1), cmt.group(2)
        if col not in create_cols.get(tbl, set()):
            removed_comment += 1
            continue  # 删除整行
        out.append(line)
        continue

    # (2) 修复 ALTER TABLE ... ADD COLUMN ... REFERENCES <未创建表> 的 FK
    am = re.match(r'\s*ALTER TABLE public\.(\w+)\s+ADD COLUMN', line)
    if am:
        # 抽取 REFERENCES 引用的表名
        refs = re.findall(r'REFERENCES\s+"?public"?\.?"?(\w+)', line)
        bad = False
        for r in refs:
            # 引用的表在 CREATE TABLE 集合中不存在 -> FK 会因顺序失败，去掉 REFERENCES 子句
            if r not in create_cols:
                bad = True
                break
        if bad:
            # 去掉 REFERENCES ... (...) 部分（保留列定义）
            new_line = re.sub(r'\s*REFERENCES\s+"?public"?\.?"?\w+"?\s*(\([^)]*\))?', '', line)
            fixed_alter += 1
            out.append(new_line)
            continue
        out.append(line)
        continue

    out.append(line)

with open(path, "w", encoding="utf-8") as f:
    f.writelines(out)

print(f"删除无效 COMMENT ON COLUMN: {removed_comment} 条")
print(f"修复 ALTER ADD COLUMN FK 顺序: {fixed_alter} 条")
