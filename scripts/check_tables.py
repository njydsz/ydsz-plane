#!/usr/bin/env python3
import psycopg2

conn = psycopg2.connect(host='127.0.0.1', port=5432, user='postgres', password='Limw1020', dbname='ydsz-plane')
cur = conn.cursor()

cur.execute("SELECT table_schema, table_name FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog', 'information_schema') ORDER BY table_schema, table_name")
all_tables = cur.fetchall()
print('All tables in database:')
for schema, table in all_tables:
    if 'task' in table.lower() or 'requirement' in table.lower() or 'defect' in table.lower():
        print(f'  {schema}.{table}')

cur.execute("SELECT tablename FROM pg_tables WHERE schemaname='public'")
tables = [r[0] for r in cur.fetchall()]
print()
print('All public tables:')
for t in sorted(tables):
    print(f'  {t}')
conn.close()
