-- ===========================================================================
-- V26.09.28 · JSONB 字段结构约束（半结构化字段防 malformed 数据入库）
-- 依据：S18 分析报告 P2-D4
-- ===========================================================================
-- 原则：对已知稳定结构的 JSONB 字段添加 CHECK 约束，防止应用层误写入
--       非结构化字段（如 config / viewport）保持灵活
--
-- 长期进化路径：高频过滤 JSONB 字段逐步提取为独立列（见备注）
-- ===========================================================================

-- ---------- defect：reproduce_steps / fix_steps 必须是 JSON 对象 ----------
ALTER TABLE defect
    ADD CONSTRAINT chk_defect_reproduce_steps_json
    CHECK (reproduce_steps IS NULL OR jsonb_typeof(reproduce_steps) = 'object');

ALTER TABLE defect
    ADD CONSTRAINT chk_defect_fix_steps_json
    CHECK (fix_steps IS NULL OR jsonb_typeof(fix_steps) = 'object');

-- ---------- task / requirement / defect：description_json 必须是对象 ----------
ALTER TABLE task
    ADD CONSTRAINT chk_task_description_json
    CHECK (description_json IS NULL OR jsonb_typeof(description_json) = 'object');

ALTER TABLE requirement
    ADD CONSTRAINT chk_req_description_json
    CHECK (description_json IS NULL OR jsonb_typeof(description_json) = 'object');

ALTER TABLE defect
    ADD CONSTRAINT chk_defect_description_json
    CHECK (description_json IS NULL OR jsonb_typeof(description_json) = 'object');

-- ---------- sprints.viewport 必须是对象 ----------
ALTER TABLE sprints
    ADD CONSTRAINT chk_sprint_viewport_json
    CHECK (viewport IS NULL OR jsonb_typeof(viewport) = 'object');

-- ---------- notification_preferences.channel_settings 必须是对象 ----------
ALTER TABLE notification_preferences
    ADD CONSTRAINT chk_notif_pref_channel_json
    CHECK (channel_settings IS NULL OR jsonb_typeof(channel_settings) = 'object');

-- ⚠️ TODO（后续独立迁移，需同步修改 Go 代码）：
--   1. tenants.brand_config 提取 primary_color / favicon_url 独立列
--   2. workspaces.config 提取 default_view / feature_flags 独立列
--   3. projects.config 提取 default_view / module_nav_mode 独立列
--   4. workbench_configs.widget_states 如结构稳定可提取
