# Changelog

本项目的所有显著变更均记录在此文件中。

格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

---

## [Unreleased] - 开发中

### 新增
- 自动化规则 DSL 校验的单元测试（22 个用例覆盖 trigger/condition/action 边界）
- 核心价值链集成测试：需求 CRUD 状态流转 软删除 恢复闭环
- CI 新增 API 冒烟门禁：`/healthz` + `/readyz` 端到端校验
- `make smoke` 本地冒烟命令

### 修正
- README 对齐代码现实：DDD 宣称改为当前真实架构描述（transaction script + 领域术语命名），未来路线含独立 domain 层
- README 测试描述从金字塔数据改为真实覆盖率基线
- README 开发规范（commitlint/git-cliff 改为建议项，分支策略改为 GitHub Flow）
- CI Go 版本修正为 1.26（与 go.mod 一致）
- CI 测试任务输出覆盖率摘要与 artifact 上传

### 安全
- CI govulncheck 与 CodeQL 持续扫描

---

## [v0.1.0] - 2026-09-24

### 首个 GA 版本

面向中国软件团队的开源项目管理平台首发版本。对标 Jira、Linear、云效、TAPD、ONES。

**核心功能**

- **工作空间与 RBAC**：多租户隔离、四级角色（Owner/Admin/Member/Guest）、邮箱邀请
- **工作项管理**：需求/任务/缺陷独立追踪，WBS 三层级，PDM 四类依赖
- **状态机**：项目级自定义状态与流转规则，6 状态组，内置三组行业模板
- **迭代 Sprint**：生命周期（规划/执行/复盘）、容量规划、燃尽/燃起图
- **版本管理**：产品发版里程碑容器，semver 校验，Release Notes 自动生成
- **看板**：拖拽排序、列约束、组视图、WIP 限制
- **全局搜索**：类 JQL 语法、ES 异步双写、IK 中文分词
- **实时通知**：站内 + 邮件 + IM，订阅矩阵 + Digest 摘要
- **仪表盘**：DORA 指标、燃起/燃尽、累积流图（CFD）、速率趋势
- **自动化**：JSON DSL 规则引擎，事件驱动 + DLQ 保障
- **Webhook**：HMAC 签名 + 重放防护 + 投递日志
- **AI 集成**：OpenAI/Claude 接口，摘要/分类/估算/智能标签
- **知识库**：层级页面、版本历史、评论、@mention

**技术栈**

- 后端：Go 1.26.5 + Gin + pgx
- 前端：Vue 3.5 + TypeScript + Vite 6 + Pinia
- 数据库：PostgreSQL 18（兼容达梦/人大金仓）
- 缓存/队列：Redis 8 + RabbitMQ 4
- 搜索：Elasticsearch 8.14（IK Analyzer）
- 交付：Docker Compose / Helm / 信创（amd64 + arm64）
