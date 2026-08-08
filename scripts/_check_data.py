import psycopg2
conn = psycopg2.connect(host='127.0.0.1', port=5432, user='postgres', password='Limw1020', dbname='ydsz-plane')
conn.autocommit = True
cur = conn.cursor()

# 列出数据库中所有用户表（按行数降序）
cur.execute("""
    SELECT table_name FROM information_schema.tables 
    WHERE table_schema='public' AND table_type='BASE TABLE'
    ORDER BY table_name
""")
all_tables = [r[0] for r in cur.fetchall()]

print(f"数据库共有 {len(all_tables)} 张表\n")

results = []
for t in all_tables:
    try:
        cur.execute(f'SELECT COUNT(*) FROM "{t}"')
        cnt = cur.fetchone()[0]
        results.append((t, cnt))
    except Exception as e:
        results.append((t, -1))

results.sort(key=lambda x: -x[1])
print(f"{'表名':<40} {'行数':>6}")
print('-' * 50)
for t, cnt in results:
    if cnt > 0:
        print(f'{t:<40} {cnt:>6}')
    elif cnt == 0:
        print(f'{t:<40} {cnt:>6}  (空)')

# schema_migrations 内容
print("\n=== schema_migrations ===")
cur.execute("SELECT * FROM schema_migrations")
for r in cur.fetchall():
    print(f"  {r}")

# users 表结构
print("\n=== users 表字段 ===")
cur.execute("""
    SELECT column_name, data_type, is_nullable 
    FROM information_schema.columns 
    WHERE table_schema='public' AND table_name='users' 
    ORDER BY ordinal_position
""")
for r in cur.fetchall():
    print(f"  {r[0]:<25} {r[1]:<20} {r[2]}")

print("\n--- users 数据 ---")
cur.execute("SELECT * FROM users ORDER BY id")
cols = [d[0] for d in cur.description]
print(f"  列: {cols}")
for r in cur.fetchall():
    print(f"  {r}")

print("\n--- search_documents 结构与数据 ---")
cur.execute("""
    SELECT column_name, data_type 
    FROM information_schema.columns 
    WHERE table_schema='public' AND table_name='search_documents' 
    ORDER BY ordinal_position
""")
for r in cur.fetchall():
    print(f"  {r[0]:<25} {r[1]}")
cur.execute("SELECT * FROM search_documents")
for r in cur.fetchall():
    print(f"  {r}")

conn.close()
