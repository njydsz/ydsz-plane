-- 0004_issue_core: rollback
DROP TABLE IF EXISTS issue_dependencies CASCADE;
DROP TABLE IF EXISTS issue_relations CASCADE;
DROP TABLE IF EXISTS time_logs CASCADE;
DROP TABLE IF EXISTS issue_activities CASCADE;
DROP TABLE IF EXISTS issue_watchers CASCADE;
DROP TABLE IF EXISTS issue_modules CASCADE;
DROP TABLE IF EXISTS issue_labels CASCADE;
DROP TABLE IF EXISTS issue_assignees CASCADE;
DROP TABLE IF EXISTS state_transitions CASCADE;
DROP TABLE IF EXISTS project_sequences CASCADE;
DROP TABLE IF EXISTS issues CASCADE;
DROP TABLE IF EXISTS labels CASCADE;
DROP TABLE IF EXISTS modules CASCADE;
DROP TABLE IF EXISTS states CASCADE;

DROP TRIGGER IF EXISTS trg_states_updated_at ON states;
DROP TRIGGER IF EXISTS trg_modules_updated_at ON modules;
DROP TRIGGER IF EXISTS trg_labels_updated_at ON labels;
DROP TRIGGER IF EXISTS trg_issues_updated_at ON issues;
DROP TRIGGER IF EXISTS trg_time_logs_updated_at ON time_logs;
