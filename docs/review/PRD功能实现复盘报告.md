# Ydsz Plane PRD V1.0 功能实现复盘报告

> 复盘日期：2026-08-08
> 比对基准：`docs/Ydsz Plane 产品需求文档.docx`（V1.0 最终完整版）
> 代码基线：当前工作区最新前/后端代码（Go 后端 `internal/` + Vue 3 前端 `web/`）

## 一、总体结论

| 层级 | 完成度评估 | 说明 |
|---|---|---|
| 后端（Go, DDD 分层） | **约 85–90%** | 核心研发协同链路高度完整，缺口集中在知识库、项目功能开关、文档版本、回收站 |
| 前端（Vue 3 + Vite + Pinia） | **约 75–80%** | 核心闭环页面齐全，缺口集中在 WBS 树、模块管理、知识库、回收站、SSO、仪表盘高级能力 |

**一句话总结**：PRD 中定义的三大工作项（需求/任务/缺陷）+ 迭代 + 版本日 + 收件箱 + 通知 + 搜索 + 度量 + Webhook/自动化的**主链路已闭环且质量较高**；主要差距在于「知识库模块整体缺失」以及一批 P1/P2 级的增强能力（回收站、导入、保存视图、SSO、拖拽布局等）。

---

## 二、逐模块比对明细

图例：✅ 已实现 ｜ ⚠️ 部分实现 ｜ ❌ 未实现

### 8.1 工作空间管理 — ⚠️ 大部分实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 创建/编辑空间、Slug 唯一校验 | ✅ | `internal/application/workspace/service.go`；workspaces 表唯一索引；前端 `WorkspaceSettingsView.vue` |
| 成员邀请（邮箱链接）、审核 | ✅ | `invitation.go`（邀请 token） |
| 4 级角色（Owner/Admin/Member/Guest） | ✅ | `rbac/store.go`（实际角色集更细：owner/admin/pm/po/techlead/qalead/dev/guest） |
| API Token 管理 | ✅ | `application/apitoken/` + `/me/api-tokens`；前端 `ApiTokensView.vue` |
| 成员移除/列表 | ✅ | 前后端均有 |
| Logo 上传 / 主题色 | ⚠️ | 后端有字段，前端无 Logo 上传 UI |
| 归档/删除空间 | ⚠️ | 后端有归档 service；前端入口不明显 |
| CSV 批量导入成员（P2） | ❌ | 前后端均未实现 |
| SSO/SAML（规划中） | ❌ | 前端 LoginView 无 SSO/OAuth 入口 |

### 8.2 项目管理 — ⚠️ 大部分实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 项目 CRUD、Identifier 自动生成 | ✅ | `workspace/project.go`；identifier 归一化+唯一约束 |
| 网络类型（公开/私有） | ✅ | network=public/private |
| 项目模板 | ✅ | `template.go` / `issue/project_templates.go`；前端创建时可选模板 |
| 归档项目 | ⚠️ | 后端有软删归档；前端 `ProjectSettingsView.vue` 无归档操作入口 |
| **功能模块开关（intake/sprint/version/estimate）** | ❌ | projects 表与代码均不存在该配置，PRD 2.4.2 要求未落地 |
| 项目复制 / 项目分组（规划中） | ❌ | 未实现 |

### 8.3 版本日管理（Version）— ✅ 已实现（PRD 原标注"规划中"，已落地）

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| Version 模型（版本号/发布日期/检查清单/进度快照） | ✅ | `application/version/models.go`；`version_delivery_snapshots` |
| 1:N 关联迭代 | ✅ | `sprints.version_id` |
| 发布/归档流转、Release Notes | ✅ | `POST /versions/:id/activate\|release\|archive`；前端 `VersionReleaseView.vue` 两步发布向导（检查清单 + Release Notes） |
| 进度聚合视图 | ✅ | `VersionListView.vue` 进度聚合 ProgressBar |
| 发布模板 / 语义化版本校验 | ⚠️ | semver 校验有；发布模板（常规/热修复）未见 |

### 8.4 迭代管理（Sprint）— ✅ 已实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 迭代 CRUD、目标/容量 | ✅ | `application/sprint/`（Goal、Capacity） |
| Backlog 拖拽规划 | ✅ | 前端 `SprintPlanningView.vue`；后端 `POST/DELETE /sprints/:id/issues` + sort_order（无专用 reorder 端点，略弱） |
| 迭代快照（每日留存） | ✅ | `sprint_snapshots` 表 |
| 燃尽图 / 燃起图 | ✅ | `GET /sprints/:id/burndown`；前端 `BurndownChart.vue` / `BurnupChart.vue` |
| 速率趋势 | ✅ | `VelocityStats`；前端有速率 widget |
| 站会视图 | ✅ | `SprintStandupView.vue` |
| 强制结束 + 未完成任务处理（退回 Backlog/推入下一迭代） | ✅ | `UnfinishedStrategy`（backlog/next_sprint/keep） |
| 复盘数据 | ✅ | SprintDetailView「复盘数据」区块 |
| 中途加项标记 | ⚠️ | 未见"中途加入"标记与速率影响统计 |
| CFD（规划中） | ✅ 超预期 | 已在效能度量模块实现 |

### 8.5 需求管理 — ⚠️ 核心实现，UI 有缺口

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 需求 CRUD、富文本、活动日志 | ✅ | `issue/models.go`、`issue_activities`；前端 `IssueCreateModal.vue` / `IssueDetailView.vue` |
| 三级 WBS（Epic→Feature→Story） | ⚠️ | 后端 `parent_id` + depth≤3 完整；**前端无树形视图/子工作项 UI**（仅 API 层） |
| 需求评审工作流 | ✅ | `RequirementReviewFlowStates` + `ReviewStatus`（超预期，PRD 标注"规划中"） |
| 需求来源（Source） | ✅ | `source` 列（customer/internal/...） |
| 参与人/抄送人 | ✅ | `issue_watchers` |
| 关联缺陷一键创建 | ✅ | RelationPanel + type 关联 |
| 需求进度自动汇总 | ✅ | 后端自动计算 |
| 复制/移动/跨项目 | ✅ | 已实现 |
| 需求模板（规划中） | ⚠️ | 项目模板有，单需求模板未见 |
| 回收站/软删除恢复 | ❌ | 软删字段普遍存在，但**无回收站列表与恢复接口/UI** |

### 8.6 任务管理 — ⚠️ 核心实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 任务 CRUD、三级子任务 | ✅ | 同需求，WBS 后端完整 |
| 工时管理（预估/实际/剩余/日志） | ✅ 超预期 | `time_logs` 表 + `timelog_service.go`；前端 IssueDetailView 记录工时（PRD 标注"规划中"已落地） |
| 任务依赖 FS/SS/FF/SF | ✅ 超预期 | `issue_dependencies` + DFS 防环；但**前端仅暴露 relates_to/blocked_by**，FS/SS/FF/SF 选择 UI 缺失 |
| 甘特图依赖箭头 | ✅ | `GanttChartView.vue` SVG 箭头 |
| 任务分类（Category） | ✅ | `category` 列（后端）；前端表单暴露情况待补 |
| 批量创建/批量操作/Excel 导入（P1） | ❌ | 未见工作项导入端点与 UI |
| 延期原因 / 实际工作量分析（P2） | ⚠️ | 工时有，延期原因枚举未见 |

### 8.7 缺陷管理 — ⚠️ 后端完整，前端表单有缺口

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 缺陷 CRUD | ✅ | 同工作项主链路 |
| 严重程度 5 级（P0 必补项） | ✅ | `severity` 列；前端表单已含 |
| 发现阶段（P0） | ✅ | `found_phase`；前端表单已含 |
| 发现/修复版本（P0） | ⚠️ | 后端 `found_version_id`/`fix_version_id` 完整；前端表单仅有 found_version，**fix_version UI 未见** |
| 复现步骤 + 期望/实际结果（P0） | ⚠️ | 后端 `reproduce_steps{steps,expected,actual}` 完整；前端表单有 reproduce_steps，**期望/实际结果分离字段 UI 未见** |
| 环境信息 / 根因 / 验证人（P2） | ⚠️ | 后端全有；前端 environment 有，root_cause/verifier UI 未见 |
| 缺陷状态机（新建→确认→处理中→修复→待验→关闭） | ✅ | 评审/状态流转流已配置（`project_init.go`） |
| 缺陷分析报表（模块/严重程度/发现阶段/龄期/趋势/根因分布） | ⚠️ | 度量接口与 widget 部分覆盖（缺陷分布、逾期列表）；**缺陷龄分析、根因分布图未见** |
| 缺陷导出 CSV/Excel（P1） | ✅ | `export.go`（WriteCSV/WriteXLSX）；前端 SearchView 可导出 |
| 缺陷模板库（规划中） | ❌ | 未见 |

### 8.8 项目文档归档 — ⚠️ 部分实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 文档 CRUD、树形目录 | ✅ | `application/pages/`；前端 `PagesView.vue`（Tiptap 富文本，非 Markdown） |
| 文档分类（PRD/设计/接口/测试报告） | ❌ | 无分类字段 |
| 版本管理（自动递增/对比/回滚） | ❌ | 无历史版本表 |
| 关联工作项/迭代/版本日 | ❌ | 无关联模型 |
| 评审流程 / 模板库 / 分享链接 | ❌ | 未见 |

### 8.9 知识库（Knowledge Base）— ❌ 整体未实现

KnowledgeSpace / KnowledgePage / KnowledgePageVersion / KnowledgePageRelation 模型、表、路由、页面**全部缺失**。这是 PRD 中唯一一个零实现的 P0 级模块。注：现有 pages 模块可作为其能力种子（树形+软删+乐观锁），但空间/权限/版本/关联/检索均需新建。

### 8.10 收件箱（Intake）— ✅ 已实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| Channel 配置、公开门户 | ✅ | `intake_channels`；前端 `IntakeSettingsView.vue` / `IntakePublicView.vue` |
| 提交反馈（免登录）+ 跟踪 ID | ✅ | `/api/v1/public/intake/.../submit` |
| 转正为需求/缺陷、拒绝/归档 | ✅ | `ReviewIssue` / `ConvertToIssue` |
| 自动分配规则 | ✅ | `auto_assign_rules` |
| 重复检测（AI，P2） | ⚠️ | `ai.Handler.DetectDuplicates` 存在但未内嵌 intake 提交流程 |
| 邮件自动回复（规划中） | ❌ | 未见 |

### 9.1 项目仪表盘 — ⚠️ 基础实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| Widget 体系（概览/迭代进度/模块分布/质量/风险/速率等） | ✅ | `dashboard/handler.go` + `dashboard_widgets`；前端 `widgetRegistry.ts` 10 种 widget |
| 模板、预警规则 | ✅ | templates、alerts 均有 |
| 拖拽布局自定义 | ❌ | 仅可增删 widget，无拖拽/缩放布局 |
| 全屏驾驶舱 / 卡片联动 / 多项目聚合下钻 | ❌ | 未见 |
| DORA 卡片 | ⚠️ | 在独立 MetricsView，未作为仪表盘 widget |
| 导出报表（PNG/CSV） | ⚠️ | 数据导出有，截图导出未见 |

### 9.2 个人工作台 — ⚠️ 基础实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 我的待办（分桶）、今日任务、迭代概览、最近访问 | ✅ | `workbench/handler.go`；`WorkbenchView.vue` |
| 布局配置、模板 | ✅ | config + templates 接口 |
| 今日日程（会议/日历） | ⚠️ | 无独立日程/会议接口 |
| 关注动态（订阅的工作项变更流） | ❌ | 前端未见 |
| 个人效率报告（P2）/ Focus Mode / 拖拽排序 | ❌ | 未见 |

### 9.3 通知中心 — ✅ 实现度高

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 站内 + 邮件 + 企微/钉钉/飞书推送 | ✅ | `notification/dispatcher.go`（钉钉 HMAC 签名） |
| 订阅配置、免打扰时段 | ✅ | `notification_preferences`；前端 `NotificationPreferencesView.vue` |
| @提及通知 | ✅ | comment 服务内含提及通知 |
| 汇总方式（即时/日报/周报） | ⚠️ | 偏好配置有，digest 聚合发送待确认 |

### 9.4 全局搜索 — ✅ 实现度高

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 全文搜索（ES）+ 多对象 + 高亮 | ✅ | `search/es_backend.go`（olivere/elastic），pg search_tsv 兜底 |
| 类 JQL 高级语法 | ✅ 后端 | `pkg/searchql`；**前端未暴露语法输入引导** |
| 快捷键 Cmd+K、搜索历史/收藏 | ✅ | `CommandPalette.vue`、`search_history`/`search_bookmarks` |
| 导出结果 | ✅ | SearchView CSV 导出 |

### 9.5 视图与自定义 — ⚠️ 大部分实现

| PRD 功能点 | 状态 | 证据 / 缺口 |
|---|---|---|
| 看板/列表/甘特/日历/电子表格 五视图 | ✅ | 前端路由齐全 |
| 过滤/分组/排序、列配置持久化 | ✅ | `view_preferences` |
| **保存视图（个人/团队/默认）+ 分享** | ❌ | 仅偏好持久化，无命名视图管理 |
| **CSV/Excel 导入工作项 + 字段映射 + 增量导入** | ❌ | 仅导出；导入未见（`migrate/importer.go` 是数据迁移工具，非用户功能） |

### 9.6 Webhook & API — ✅ 实现度高

Webhook CRUD / HMAC-SHA256 签名 / 指数退避重试 / 推送日志 / 测试推送 全部落地（`webhook/dispatcher.go` + `WebhookSettingsView.vue`）；Swagger（dev 环境 gin-swagger）、Rate Limit 中间件均有。缺口：SDK（P2）未提供；批量操作 API（P1）未见。

### 9.7 自动化引擎 — ✅ 已实现

`automation/`：DSL（trigger/condition/action）校验、cron 定时触发、`rule_executions` 执行日志；前端 `AutomationView.vue`（配置 + 执行历史）。缺口：IF/ELSE 分支、失败自动禁用告警待确认。

### 9.8 研发效能度量 — ✅ 实现度高

`metrics/handler.go`：`/dora`（四指标）、`/cfd`、`/velocity`、`/control-chart`、`/lead-time`、`/quality` + `deployment_events` 表；前端 `MetricsView.vue` 完整覆盖 DORA/CFD/控制图/速率。缺口：数据校准、按项目权限隔离、仪表盘模板预设待确认。

---

## 三、缺口优先级汇总

### P0 —— 建议立即补齐（影响 PRD 主承诺）

| # | 缺口 | 范围 | 工作量预估 |
|---|---|---|---|
| 1 | **知识库模块整体缺失**（8.9） | 后端 4 模型 + 前端全套页面 | 大（2–3 迭代） |
| 2 | **WBS 树形视图与子工作项 UI**（后端已就绪，纯前端工作） | 前端 | 中 |
| 3 | **缺陷表单字段补齐**：期望/实际结果、fix_version、root_cause、verifier（后端已就绪） | 前端 | 小 |
| 4 | **项目功能模块开关**（intake/sprint/version/estimate，PRD 2.4.2） | 前后端 | 中 |
| 5 | **回收站（软删除列表 + 恢复）** | 前后端 | 中 |

### P1 —— 下一迭代规划

| # | 缺口 | 范围 |
|---|---|---|
| 6 | 文档模块增强：分类、版本历史（对比/回滚）、关联工作项 | 前后端 |
| 7 | 工作项 CSV/Excel 导入 + 字段映射 + 增量导入 | 前后端 |
| 8 | 保存视图（个人/团队/默认）+ 视图分享 | 前后端 |
| 9 | 仪表盘拖拽布局 + 多项目聚合 + DORA widget | 前端为主 |
| 10 | 工作台：关注动态、今日日程、个人效率报告 | 前后端 |
| 11 | 任务依赖 FS/SS/FF/SF 前端选择 UI（后端已就绪） | 前端 |
| 12 | 模块管理页（列表/负责人/目标版本，后端已就绪） | 前端 |
| 13 | 成员 CSV 批量导入 | 前后端 |
| 14 | 缺陷龄分析、根因分布报表 | 前后端 |

### P2 —— 后续增强

| # | 缺口 |
|---|---|
| 15 | SSO/OAuth 登录 |
| 16 | 全屏驾驶舱、卡片联动、报表截图导出 |
| 17 | 工作空间 Logo 上传 UI、品牌定制 |
| 18 | 邮件自动回复（Intake）、AI 重复检测内嵌提交流 |
| 19 | 缺陷/需求模板库 |
| 20 | 中途加项标记与速率影响统计 |
| 21 | SDK（Python/JS）、批量操作 API |

---

## 四、下一步优化完善建议

### 建议一：优先做「后端已就绪、前端缺 UI」的快赢项（投入小、见效快）

以下能力后端模型和接口已完整，只差前端表单/页面，建议打包为一个「UI 补齐迭代」：

1. **WBS 树形视图**（Epic→Feature→Story 折叠树 + 添加子工作项入口）——这是 PRD 3.5 的核心架构承诺，当前用户完全无法感知三级层级。
2. **缺陷表单 P0 字段**：期望结果/实际结果分离输入、修复版本（fix_version）、验证人、根因分类——PRD 第 7 章标注的 P0 必补字段后端已全部落库，补 UI 即可关闭。
3. **任务依赖类型选择**（FS/SS/FF/SF 下拉）+ 循环依赖错误提示展示。
4. **模块管理页**（列表、负责人、目标版本日）——后端 `module_service.go` 已就绪。
5. **项目设置页补归档操作 + 功能模块开关**。

### 建议二：知识库模块分两期落地

- **一期（P0）**：基于现有 pages 模块扩展——加 KnowledgeSpace 容器、空间级权限（Owner/Admin/Editor/Viewer）、文档版本快照表、全文搜索接入已有 ES 索引。
- **二期（P1）**：文档-工作项双向关联、评审流、模板库、导入导出、公开分享链接。
- 不建议另起炉灶重写编辑器，Tiptap 已可用，Markdown 导出可在序列化层补。

### 建议三：数据安全与用户信任项

1. **回收站**：软删除已普遍存在（`deleted_at`），补统一的回收站列表/恢复/彻底删除接口 + 前端入口，覆盖需求/任务/缺陷/文档，兑现 PRD 多处「删除后进入回收站可恢复」的验收标准。
2. **导入能力**：CSV/Excel 导入 + 字段映射 + external_id 增量识别，是团队从 TAPD/Jira 迁移的关键路径，直接影响获客。
3. **通知 digest**：即时/日报/周报聚合发送，避免通知轰炸（PRD 9.3 明确要求）。

### 建议四：度量与驾驶舱增强（差异化卖点）

1. 仪表盘升级：拖拽布局（gridstack/vue-grid-layout）、多项目 PMO 聚合视图、DORA widget 并入仪表盘。
2. 缺陷分析报表补齐：缺陷龄分析（超阈值标红）、根因分布、逃逸率趋势——后端质量接口已有数据基础。
3. 全屏驾驶舱模式可作为开源版的展示亮点，成本低传播价值高。

### 建议五：一致性修正

1. PRD 中角色定义为 4 级（Owner/Admin/Member/Guest），后端实际为 8 角色——建议 PRD 更新或后端做角色映射层，避免文档与实现漂移。
2. PRD 竞品对标表中多处「△ 规划中/增强中」已实际落地（版本日、评审流、工时、依赖、CFD），建议回写 PRD 状态，保持文档可信度。
3. 迭代拖拽规划建议补专用 reorder 端点，替代当前 sort_order 全量提交。

---

## 五、结论

Ydsz Plane 当前代码实现与 PRD V1.0 的吻合度**超出文档自身标注的预期**——多处 PRD 标注「规划中」的能力（版本日、需求评审流、工时管理、任务依赖、CFD、效能度量）已在代码中落地。剩余差距集中在：

- **一个整体缺失的模块**：知识库（8.9）；
- **一批"后端有、前端没有"的 UI 缺口**：WBS 树、缺陷字段、依赖类型、模块管理页；
- **一组用户信任与迁移能力**：回收站、导入、保存视图、SSO。

按第四节建议的快赢项优先执行，可在 1–2 个迭代内将 PRD 吻合度提升到 95% 左右。
