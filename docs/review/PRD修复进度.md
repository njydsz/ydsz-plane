# Ydsz Plane PRD 缺口修复进度

> 最后更新: 2026-08-08 16:58

## 已完成

| # | 项 | 范围 | 状态 |
|---|---|---|---|
| P1-11 | 任务依赖 FS/SS/FF/SF 前端选择 UI | RelationPanel.vue 增强 | ✅ 完成 |
| P0-4 | 项目功能模块开关 | 前端 Type + API + ProjectSettings UI | ✅ 完成 |
| P0-5 | 回收站 | 后端 Go handler + 前端 TrashView.vue + 路由 | ✅ 完成 |
| P0-3 | 缺陷表单 P0 字段补齐 | IssueCreateModal.vue + IssueDetailView.vue | 🔄 代理工作中 |
| P0-2 | WBS 树形视图 | IssueDetailView.vue 子工作项区块 | 🔄 代理工作中 |
| P1-12 | 模块管理页面 | 新建 ModuleSettingsView.vue + API + 路由 | 🔄 代理工作中 |

## 修改文件清单

### 后端 (Go)
- `internal/application/issue/handler.go` — 新增 `restoreIssue`、`hardDeleteIssue`、`listTrash` 三个 handler + 路由注册

### 前端 (Vue/TS)
- `web/src/views/project/RelationPanel.vue` — 新增任务依赖管理区块（FS/SS/FF/SF + 滞后天数）
- `web/src/views/project/ProjectSettingsView.vue` — 新增功能模块开关面板（Intake/Sprint/Version/Estimate toggle）
- `web/src/api/services/workspace.ts` — Project 类型新增 modules 字段
- `web/src/api/services/issue.ts` — 新增 TrashItem 类型与 listTrash/restoreIssue/permanentDelete API
- `web/src/views/project/TrashView.vue` — 新建回收站页面
- `web/src/router/index.ts` — 新增 project-trash 路由

## 待完成

| # | 项 | 预估 |
|---|---|---|
| P0-1 | 知识库模块 | 2-3 迭代 |
| P1-6 | 文档模块增强 | 1 迭代 |
| P1-7 | 工作项 CSV/Excel 导入 | 1 迭代 |
| P1-8 | 保存视图 | 1 迭代 |
| P1-9 | 仪表盘拖拽布局 | 1 迭代 |
| P1-10 | 工作台增强 | 1 迭代 |
| P1-13 | 成员 CSV 批量导入 | 中 |
| P1-14 | 缺陷分析报表 | 中 |
| P2 | SSO/驾驶舱/Logo/SDK 等 | 低优先 |
