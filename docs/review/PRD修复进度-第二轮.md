# Ydsz Plane PRD 缺口修复进度 — 第二轮

> 更新日期：2026-08-08 21:45
> 说明：本轮与用户本人并行开发同步进行，部分功能已由用户提交（见 git log）

## 状态总览

| # | 优先级 | 功能 | 状态 |
|---|---|---|---|
| P0-1 | P0 | 知识库（后端） | ✅ 已提交（9507c9b） |
| P0-1 | P0 | 知识库（前端页面） | 🔄 代理实现中 |
| P0-2~P0-5 | P0 | WBS/缺陷字段/模块开关/回收站 | ✅ 第一轮已完成 |
| P1-6 | P1 | 文档模块增强 | ✅ 第一轮已完成 |
| P1-7 | P1 | 工作项 CSV 导入 | ✅ 已提交（13b8625） |
| P1-8 | P1 | 保存视图 | ✅ 已提交（13b8625） |
| P1-9 | P1 | 仪表盘拖拽 + DORA + 多项目 | ✅ 已提交（aeb26ab） |
| P1-10 | P1 | 工作台增强 | ✅ 已提交（13b8625） |
| P1-11 | P1 | 任务依赖 UI | ✅ 第一轮已完成 |
| P1-12 | P1 | 模块管理页 | ✅ 第一轮已完成 |
| P1-13 | P1 | 成员 CSV 导入 | ✅ 已提交（13b8625） |
| P1-14 | P1 | 缺陷分析报表 | ✅ 第一轮已完成 |
| P2 | P2 | SSO/OIDC 登录 | ✅ 已提交（aac107d） |
| P2 | P2 | 全屏驾驶舱 | ⬜ 待定 |

## 用户本人并行提交（git log）

- `13b8625` feat(api): 项目成员管理、工作项导入、关注动态与效率报告、命名视图管理
- `9507c9b` feat(知识库): 知识库权限定义 + 路由注册 + seed 数据（后端 2102 行）
- `aeb26ab` feat(dashboard): 多项目对比、DORA 小部件、仪表盘编辑模式
- `aac107d` feat(auth): OIDC SSO 单点登录 + HttpOnly Cookie

## 当前在途

- fe-knowledge：知识库前端（空间列表 + 文档树 + Markdown 编辑 + 版本历史 + 关联工作项）
- fe-workbench：工作台增强收尾（部分已提交，确认剩余）
- fe-verify：集成验证（部分完成，因轮次超限终止，改动已由用户提交消化）

## 未提交改动（用户并行工作，勿动）

- `internal/application/search/grpc_service.go` — 搜索 gRPC 结果拍平
- `web/src/views/project/IssueDetailView.vue` — 空值防御
- `cmd/search-service/`、`build/Dockerfile.search`、`sql/seed-test-data.sql` — 新文件
