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

## [0.1.0] - 2026-08-07（MVP v0.1 发布）

> 对应路线图 M0–M5：功能底座、核心域 API 与前端视图骨架全部就绪，Sprint 1–7 完成。
> 特性：附件（MinIO）、评论（富文本 + @提及）、通知（站内信 + 偏好）、WebSocket 实时推送（Redis Pub/Sub 多节点扇出）、CSV/xlsx 导出、性能基线脚本与 CI 压测流水线。

### 新增（已实现）
- **工程基座**：Monorepo（Go module + pnpm workspace）、Docker Compose 全栈、Makefile、DB 迁移（0001–0017）
- **鉴权链路**：注册 / 登录（bcrypt + JWT access/refresh）、Cookie 会话、401 单飞刷新重放、忘记 / 重置密码
- **RBAC**：Owner/Admin/Member/Guest 四角色 × 10 项权限中间件
- **工作空间**：CRUD、Slug 校验、成员邀请（token + 7 天有效 + 可撤销）、审计日志
- **项目**：CRUD、Identifier 生成、RLS 租户隔离
- **工作项**：Issue CRUD、状态机流转、WBS 三级、类型差异化字段、工时记录、关联（6 种关系）、依赖（FS/SS/FF/SF）、CSV/xlsx 导出
- **迭代**：Sprint 生命周期、燃尽图、容量规划、速率建议、复盘快照、站会模式
- **版本日**：Version CRUD + 状态机、进度聚合、交付报告、Release Notes、缺陷面板
- **协同增强**：评论（富文本 + @提及 + 嵌套回复）、附件（MinIO 预签名上传/下载/预览）、通知系统（站内信 + 默认规则矩阵 + 铃铛下拉面板 + 偏好设置）、WebSocket 实时推送（Redis Pub/Sub 多节点扇出）、API Token 认证
- **缺陷分析**：`/analytics/defects` 聚合查询（严重程度/发现阶段/模块/根因/缺陷龄/周趋势）+ `/analytics/defects/export` 明细导出（CSV / xlsx，支持按版本过滤）
- **可观测与安全**：Prometheus RED 指标、8 项安全头、Swagger、SMTP 邮件抽象

### S7 收尾（MVP v0.1）
- 工作项导出：新增 xlsx 格式导出（纯标准库实现，无需第三方依赖），前端增加导出格式下拉菜单（CSV / Excel）
- 缺陷分析导出：抽取通用导出基础设施 `internal/interfaces/http/export`（CSV UTF-8 BOM / 最小合法 OOXML xlsx，含 XML 转义单测），缺陷面板新增导出入口（按当前版本过滤）
- 性能基线：创建 `docs/perf/` 目录归档基线报告，新增 `scripts/seed-scale/` 大规模造数脚本（支持 100 万工作项并发写入、断点续传）；k6 三件套脚本（smoke / load / stress）就绪并验证
- 性能流水线：修复 `perf.yml` CI 脚本路径与 Makefile 压测命令（统一指向 `tests/perf/`），支持冒烟 → 负载 → 压力完整流程
- 通知中心：铃铛下拉面板（未读角标、领域图标、已读/未读高亮）、通知列表页（筛选/分页/全部已读/归档）、通知偏好设置页
- MVP 加固：全链路加载态/空态/错误态组件（AppLoadingState / AppEmptyState / AppErrorState / NotFoundView）覆盖通知、评论、看板、工作项列表、仪表盘、燃尽图、交付报告、发布向导、工作空间设置、版本列表、关联面板；修复 SearchView 既有类型错误

### 已知待完善（后续版本）
- 前端测试覆盖率仍偏低，需按测试策略文档持续补充
- CI 需接入真实后端的 E2E 环节
- 国密算法、信创数据库方言层为预留接口，尚未完成实测链路
- 性能基线实测数据待首轮 staging 全量压测后回填（`docs/perf/README.md`）
