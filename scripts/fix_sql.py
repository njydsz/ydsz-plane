#!/usr/bin/env python3
"""
直接修复 SQL 文件中的语法问题
"""
from pathlib import Path
import re

SQL_FILE = Path(__file__).parent.parent / "sql" / "ydsz-plane-init.sql"

with open(SQL_FILE, 'r', encoding='utf-8') as f:
    content = f.read()

original_size = len(content)
print(f"Original size: {original_size:,} bytes\n")

# ---------------------------
# Fix 1: CREATE TABLE 中的 UNIQUE (...) WHERE ...
# ---------------------------
lines = content.split('\n')
result = []
i = 0
fix_count = 0

while i < len(lines):
    line = lines[i]
    
    # 检测问题行: UNIQUE (col) WHERE condition 在 CREATE TABLE 内部
    m = re.match(r'^(\s+UNIQUE\s*\([^)]+)\)\s*WHERE\s+(.+)$', line)
    if m:
        cols_part = m.group(1).strip()  # "UNIQUE (col1, col2)"
        cols = cols_part.replace('UNIQUE', '').strip().strip('()').strip()
        where = m.group(2).strip()
        
        # 向前查找表名
        table = None
        for j in range(len(result)-1, max(len(result)-80, -1), -1):
            tm = re.search(r'CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:public\.)?"?(\w+)"?', result[j], re.IGNORECASE)
            if tm:
                table = tm.group(1).strip('"')
                break
        
        if table:
            # 生成索引名
            cols_clean = cols.replace(',', '_').replace(' ', '').replace('"', '')
            idx_name = f"uidx_{table}_{cols_clean}"
            
            # 向前找到 ); 位置，在其后插入 CREATE UNIQUE INDEX
            for k in range(len(result)-1, max(len(result)-30, -1), -1):
                if result[k].strip() == ';':
                    idx_stmt = f"CREATE UNIQUE INDEX IF NOT EXISTS {idx_name} ON {table}({cols}) WHERE {where};"
                    result.insert(k + 1, idx_stmt)
                    fix_count += 1
                    print(f"  [{fix_count}] Fixed {table}: moved UNIQUE WHERE to separate CREATE INDEX")
                    break
        
        # 跳过当前行（不添加到 result）
        i += 1
        continue
    
    result.append(line)
    i += 1

content = '\n'.join(result)

# ---------------------------
# Fix 2: ALTER TABLE ... ADD CONSTRAINT ... UNIQUE ... WHERE
# ---------------------------
def fix_alter_unique(m):
    table = m.group(2)
    name = m.group(3)
    cols = m.group(4)
    where = m.group(5)
    return f'CREATE UNIQUE INDEX IF NOT EXISTS {name} ON {table}({cols}) WHERE {where};'

content = re.sub(
    r'(ALTER\s+TABLE\s+(\w+)\s+ADD\s+CONSTRAINT\s+(\w+)\s+UNIQUE\s*\()([^)]+)\)\s*WHERE\s+([^;]+);',
    fix_alter_unique,
    content,
    flags=re.IGNORECASE
)

# ---------------------------
# Fix 3: CREATE POLICY IF NOT EXISTS
# ---------------------------
content = re.sub(
    r'CREATE\s+POLICY\s+IF\s+NOT\s+EXISTS\s+(\w+)\s+ON\s+(\w+)',
    r'DROP POLICY IF EXISTS \1 ON \2; CREATE POLICY \1 ON \2',
    content,
    flags=re.IGNORECASE
)

# 修复转换后的重复分号
content = re.sub(r';(\s*CREATE POLICY)', r'\1', content)

# ---------------------------
# Fix 4: COMMENT ON 语句中的未转义单引号
# ---------------------------
lines = content.split('\n')
fixed_lines = []
for line in lines:
    stripped = line.strip()
    if stripped.upper().startswith('COMMENT ON TRIGGER'):
        # 检测触发器 COMMENT 中是否有未转义单引号破坏语法
        if re.search(r"(?:^|\s)is\s+'", stripped, re.IGNORECASE):
            fixed_lines.append('-- FIXED: ' + stripped)
            continue
    if stripped.upper().startswith('COMMENT ON COLUMN'):
        if '{role:' in stripped or "['owner'" in stripped:
            fixed_lines.append('-- FIXED: ' + stripped)
            continue
    fixed_lines.append(line)
content = '\n'.join(fixed_lines)

# ---------------------------
# 保存
# ---------------------------
with open(SQL_FILE, 'w', encoding='utf-8') as f:
    f.write(content)

print(f"\nFixed file size: {len(content):,} bytes")
print(f"Total fixes applied: {fix_count}")
print("\nDone!")
