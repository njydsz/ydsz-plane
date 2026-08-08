SELECT id, name, workspace_id FROM projects ORDER BY id LIMIT 20;
SELECT count(*) AS project_count FROM projects;
SELECT id, slug FROM workspaces ORDER BY id LIMIT 10;
SELECT count(*) AS workspace_count FROM workspaces;
SELECT count(*) AS user_count FROM users;
