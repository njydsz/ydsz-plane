#!/usr/bin/env python3
import psycopg2

conn = psycopg2.connect(host='127.0.0.1', port=5432, user='postgres', password='Limw1020', dbname='ydsz-plane')
cur = conn.cursor()

print('=== New Tables (task/requirement/defect) ===')
for t in ['task', 'requirement', 'defect', 'task_ext', 'requirement_ext', 'defect_ext']:
    cur.execute("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s", (t,))
    print(f'  {t}: {"EXISTS" if cur.fetchone()[0] else "MISSING"} ')

print()
print('=== Views ===')
for v in ['task_view', 'requirement_view', 'defect_view']:
    cur.execute("SELECT COUNT(*) FROM information_schema.views WHERE table_schema='public' AND table_name=%s", (v,))
    print(f'  {v}: {"EXISTS" if cur.fetchone()[0] else "MISSING"}')

print()
print('=== Old tables that should be DROPPED ===')
old_tables = ['issues', 'issue_comments', 'issue_reactions', 'issue_votes', 
              'issue_activities', 'issue_dependencies', 'issue_relations', 
              'issue_watchers', 'issue_modules', 'issue_labels', 'issue_assignees', 
              'issue_subscriptions', 'issue_sequences', 'project_sequences', 
              'sprint_issues', 'intake_issues']
for t in old_tables:
    cur.execute("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=%s", (t,))
    exists = cur.fetchone()[0]
    status = 'EXISTS (BAD!)' if exists else 'DROPPED (GOOD)'
    print(f'  {t}: {status}')

print()
print('=== Data counts ===')
for t in ['task', 'requirement', 'defect', 'users', 'workspaces', 'states', 'automation_templates', 'projects']:
    try:
        cur.execute(f'SELECT COUNT(*) FROM "{t}"')
        print(f'  {t}: {cur.fetchone()[0]} rows')
    except Exception as e:
        print(f'  {t}: ERROR - {e}')

print()
print('=== Total tables in public schema ===')
cur.execute("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'")
print(f'  Total: {cur.fetchone()[0]} tables')

conn.close()
