import psycopg2
conn = psycopg2.connect(host='127.0.0.1', port=5432, user='postgres', password='Limw1020', dbname='ydsz-plane')
conn.autocommit = True
cur = conn.cursor()

# 核心业务表行数
tables = ['users', 'workspaces', 'workspace_members', 'roles', 'role_permissions',
          'projects', 'project_members', 'states', 'automation_templates',
          'task', 'task_ext', 'requirement', 'requirement_ext', 'defect', 'defect_ext',
          'knowledge_spaces', 'knowledge_pages', 'knowledge_page_versions', 'knowledge_page_relations',
          'notifications', 'notification_deliveries', 'notification_digests', 'notification_preferences',
          'processed_events', 'dlq_events', 'secrets',
          'document_versions', 'document_links', 'page_templates', 'page_shares',
          'tags', 'workflow_definitions', 'biz_entity_relations']

print(f"{'表名':<35} {'行数':>6}")
print('-' * 45)
for t in tables:
    try:
        cur.execute(f'SELECT COUNT(*) FROM "{t}"')
        cnt = cur.fetchone()[0]
        print(f'{t:<35} {cnt:>6}')
    except Exception as e:
        print(f'{t:<35} ERR: {str(e)[:40]}')

conn.close()
