-- S11 性能优化索引（自动化规则 + 效能度量）
-- 对应 Sprint 11 大厂标准效能优化
-- 
-- 设计原则：
-- 1. 覆盖索引消除回表
-- 2. 局部索引排除已删除数据
-- 3. 复合索引等值列在前、排序列在后
--
-- 参考：美团 DBA《高吞吐 OLTP 索引规范 v3.1》

-- ============================================
-- 自动化规则表索引优化
-- ============================================

-- 高频查询：按 project + trigger_type + status 查找活跃规则（事件触发器核心路径）
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_automation_rules_trigger_active" 
ON "automation_rules" (project_id, trigger_type, sort_order) 
WHERE status = 'active' AND project_id IS NOT NULL;

-- ============================================
-- 规则执行审计表索引优化
-- ============================================

-- 高频查询：最近执行历史（按 rule_id + created_at 倒序分页）
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_rule_executions_rule_created" 
ON "rule_executions" (rule_id DESC, created_at DESC) 
WHERE trigger_event_id IS NOT NULL;

-- 防重复投递：按 trigger_event_id + rule_id 幂等去重
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS "idx_rule_executions_idempotent" 
ON "rule_executions" (rule_id, trigger_event_id) 
WHERE trigger_event_id IS NOT NULL;

-- ============================================
-- 指标快照表索引优化
-- ============================================

-- 趋势查询：按 workspace + project + metric + snapshot_date 覆盖索引
-- 消除回表同时加速快照列表的渲染
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_metric_snap_trend" 
ON "metric_snapshots" (workspace_id, project_id, metric, snapshot_date DESC, value);

-- ============================================
-- 部署事件表索引优化
-- ============================================

-- DORA 查询：按 project + status + deployed_at 过滤成功部署
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_deployment_dora" 
ON "deployment_events" (project_id, status, deployed_at DESC) 
WHERE project_id IS NOT NULL;

-- 幂等键：(deployment_id, env, project_id) 部分索引
-- 已在表定义中的 UNIQUE 约束部分索引
