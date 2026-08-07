# Changelog

本项目遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/) 与 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

发布流程与分支策略见 [docs/architecture/03-工程基座与开发规范.md](docs/architecture/03-工程基座与开发规范.md) 与 `docs/architecture/15-发布管理.md`。

## [Unreleased]

### 工程与质量（进行中）
- 新增 GitHub Actions CI：后端 lint / test(race) / 覆盖率门禁 / govulncheck，前端 lint / typecheck / test / build / audit
- 引入前端组件单测框架（Vitest + @vue/test-utils + happy-dom）并补充原子组件测试
- 引入 Playwright E2E 冒烟链路（鉴权）
- 后端覆盖率渐进式门槛（当前 15%，随补测逐步提升）
- 清理仓库残留临时文件，补充 `.gitignore` 忽略测试产物
- 修复 `web/package.json` 的 `packageManager` 版本与本机工具链不一致问题

## [0.1.0] - 计划中（MVP v0.1）

> 对应路线图 M0–M5，功能底座与核心域 API 已就绪，前端视图骨架完备。

### 新增（已实现，待正式打版）
- **工程基座**：Monorepo（Go module + pnpm workspace）、Docker Compose 全栈、Makefile、DB 迁移（0001–0014）
- **鉴权链路**：注册 / 登录（bcrypt + JWT access/refresh）、Cookie 会话、401 单飞刷新重放、忘记 / 重置密码
- **RBAC**：Owner/Admin/Member/Guest 四角色 × 10 项权限中间件
- **工作空间**：CRUD、Slug 校验、成员邀请（token + 7 天有效 + 可撤销）、审计日志
- **项目**：CRUD、Identifier 生成、RLS 租户隔离
- **工作项**：Issue CRUD、状态机流转、WBS 三级、类型差异化字段、工时记录、关联（6 种关系）、依赖（FS/SS/FF/SF）
- **迭代**：Sprint 生命周期、燃尽图、容量规划、速率建议、复盘快照、站会模式
- **版本日**：Version CRUD + 状态机、进度聚合、交付报告、Release Notes、缺陷面板
- **协同增强**：评论（富文本 + @提及）、附件（MinIO 预签名）、通知（站内信）、WebSocket 实时推送、API Token 认证
- **可观测与安全**：Prometheus RED 指标、8 项安全头、Swagger、SMTP 邮件抽象

### 已知待完善（发布前）
- 前端测试覆盖率仍偏低，需按测试策略文档持续补充
- CI 需接入真实后端的 E2E 环节
- 国密算法、信创数据库方言层为预留接口，尚未完成实测链路

<!--
注意：实际打版时，将 [Unreleased] 的内容按特性拆分为 [0.1.0]，
并补上准确的发布日期与对比链接。
-->
