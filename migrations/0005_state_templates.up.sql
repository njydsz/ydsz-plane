-- 0005_state_templates: 内置状态模板集（项目创建时复制到 states 表）
-- 模板以 workspace_id=0 的"系统模板"形式存储，应用层在 project.Create 时复制
-- 参考 docs/architecture/07 工作项与状态机设计.md

-- 添加 template_key 列到 states 表，用于标记系统模板
ALTER TABLE states ADD COLUMN template_set TEXT NOT NULL DEFAULT 'custom'
    CHECK (template_set IN ('dev_flow','defect_flow','requirement_flow','custom'));

----------------------------------------------------------------------
-- 研发流（dev_flow） — 用于 requirement / task
----------------------------------------------------------------------
-- Backlog: 'Backlog', group=backlog
-- Todo: '待处理', group=backlog
-- In Progress: '进行中', group=started
-- In Review: '审核中', group=started
-- Done: '已完成', group=completed
-- Cancelled: '已取消', group=cancelled

----------------------------------------------------------------------
-- 缺陷流（defect_flow）
----------------------------------------------------------------------
-- New: '新建', group=backlog
-- Confirmed: '已确认', group=started
-- In Progress: '处理中', group=started
-- Fixed: '已修复', group=started
-- Verifying: '待验证', group=started
-- Closed: '已关闭', group=completed
-- Rejected: '已拒绝', group=cancelled
-- Reopened: '重新打开', group=started

----------------------------------------------------------------------
-- 需求评审流（requirement_flow）
----------------------------------------------------------------------
-- Draft: '草稿', group=backlog
-- Reviewing: '评审中', group=started
-- Accepted: '已采纳', group=completed
-- Rejected: '已拒绝', group=cancelled
-- Verified: '已验证', group=completed

----------------------------------------------------------------------
-- state_transitions 模板数据将在应用层维护（Go 常量），不写入 DB
-- 原因：1. 模板量大 2. 应用层变更比 DB migration 更灵活
-- 代码位置：internal/application/issue/state_templates.go
----------------------------------------------------------------------
