# ydsz-plane vs Plane (makeplane) — 1:1 复刻度对比分析

> 对比日期：2026-11-03  
> 对比对象：ydsz-plane（本地项目） vs [makeplane/plane](https://github.com/makeplane/plane)（GitHub 41.7k+ Stars）  
> 分析维度：功能模块覆盖、交互体验、技术架构

---

## 一、总体评估

| 维度 | ydsz-plane | Plane | 复刻度 |
|------|-----------|-------|--------|
| 核心工作项管理 | Requirement/Task/Defect + Epic | Work Items (Task/Epic + 自定义类型) | **~85%** |
| 敏捷迭代 (Sprint/Cycle) | Sprint 全生命周期 | Cycle 全生命周期 + 并行周期 | **~80%** |
| 模块管理 | Module CRUD + 进度 | Module + 状态 + 归档 + 导出 | **~75%** |
| 视图系统 | 9 种视图 | Board/List/Spreadsheet/Gantt/Calendar | **~90%** |
| 文档/知识库 | Pages + Knowledge Base | Pages/Wiki + Block Editor | **~60%** |
| 分析/仪表盘 | Dashboard + Metrics + Defect Analytics | Analytics + Custom Dashboards | **~70%** |
| 搜索 | ES 全局搜索 + Cmd+K | PQL + Cmd+K + AI 搜索 | **~65%** |
| 通知 | WebSocket 实时 + 偏好 | Inbox + Subscribers + Comments | **~75%** |
| 自动化 | TCA 规则引擎 | Automations + Plane Runner | **~60%** |
| 版本管理 | Version 全生命周期 | Releases (项目/工作空间) | **~85%** |
| Intake | 公开渠道 + 审核 | In-app/Email/Forms + Triage | **~75%** |
| 权限/RBAC | 动态权限矩阵 | 自定义角色 + 条件授权 | **~70%** |
| AI 能力 | 智能分配/查重/分类/摘要 | Plane AI + Agents + MCP + Memory | **~25%** |
| 交互体验 | 快捷键/命令面板/Peek/Favorites | Power K + 全键盘流 + Peek + Propel | **~70%** |
| 集成/开放 | Webhook + API Token + Swagger | REST API + OAuth + SDK + Marketplace | **~50%** |
| 自托管 | Docker Compose + Helm | Docker/K8s/Prime CLI/Air-gapped | **~80%** |

**综合复刻度：约 70%**

---

## 二、核心模块逐一对比

### 2.1 工作项 (Work Items / Issues)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 工作项类型 | Task(0级) + Epic(1级) + 自定义类型 | Requirement/Task/Defect 独立聚合根 + Epic 类型码 | ⚠️ Plane 支持自定义工作项类型，ydsz 为固定三类 |
| 富文本编辑器 | Markdown + Block 编辑器 (slash 菜单) | RichTextEditor (基础) | ❌ Plane 的 block editor 远超 ydsz 的简单编辑器 |
| 子工作项 | 父子层级 (Epic > Task) | WBS parent-child | ✅ 基本对等 |
| 关联关系 | relates/duplicate/blocked/blocking + 自定义 | IssueRelation + BizEntityRelation | ⚠️ ydsz 缺少 duplicate 和 blocked/blocking 语义 |
| 属性 | state/priority/assignees/labels/dates/estimate/cycle/module/release/milestone/customer/type/parent/link/attachment | 类似但分散在不同类型 | ⚠️ ydsz 缺少 milestone、customer、release 直接关联 |
| 评论 | 评论 + @mentions + reactions + 审核隐藏 | CommentList/CommentForm | ⚠️ ydsz 缺少 reactions 和审核功能 |
| 活动日志 | 完整活动流 | IssueActivity | ✅ 基本对等 |
| 附件 | 文件上传 | AttachmentUploader | ✅ 基本对等 |
| 时间追踪 | Time log (Plane 原生) | TimeLog (Task 子域) | ✅ 有实现，交互可能不同 |
| 批量操作 | 批量编辑/移动/删除 | 批量操作 | ✅ 有实现 |
| 重复工作项 | Recurring work items | 无 | ❌ 缺失 |
| 草稿 | Draft work items | 无 | ❌ 缺失 |
| 工作项模板 | Work item templates | ContentTemplate (内容模板) | ⚠️ ydsz 有内容模板但非工作项模板 |
| URL 创建 | Create-via-URL | 无 | ❌ 缺失 |
| 安全/受众 | Work item security/audiences | 无 | ❌ 缺失 |
| 自定义属性 | Custom properties per type | WorkitemExtension | ⚠️ ydsz 有扩展机制但无 UI |

### 2.2 迭代/Sprint (Cycles)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 生命周期 | Draft/Upcoming/Active/Completed | Planned/Active/Completed | ⚠️ ydsz 缺少 Draft/Upcoming 状态 |
| 燃尽图 | Burn-down + Build-up | BurndownChart + BurnupChart | ✅ 基本对等 |
| 容量规划 | Capacity tracking | Sprint capacity validation | ✅ 有实现 |
| Backlog 拖拽 | Drag-and-drop planning | SprintPlanningView (drag) | ✅ 基本对等 |
| 中期变更跟踪 | Mid-sprint changes | added_midway 字段 | ✅ 有实现 |
| Review 快照 | Review snapshot | Sprint review snapshot | ✅ 基本对等 |
| Velocity | Velocity stats | Velocity avg/P50 | ✅ 基本对等 |
| 并行周期 | Parallel cycles (Enterprise) | 无 | ❌ 缺失（仅一个 active sprint/project） |
| 自动排期 | Auto-schedule (Business) | 无 | ❌ 缺失 |
| 提前开始/完成 | Start/complete early (Pro) | 无 | ❌ 缺失 |
| 导出 | CSV/Excel/JSON export | DataJob (异步任务) | ⚠️ ydsz 通过 DataJob 间接实现 |
| Standup 模式 | 无原生 standup | SprintStandupView | ✅ ydsz 额外有 standup |
| Teamspace 过滤 | Cycle teamspace filters | 无 | ❌ 缺失 |

### 2.3 模块 (Modules)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 创建/管理 | CRUD + lead + members | Module CRUD + 成员 | ✅ 基本对等 |
| 状态 | Backlog/Planned/InProgress/Paused/Completed/Cancelled | 无明确状态机 | ❌ ydsz 模块缺少状态流转 |
| 进度跟踪 | Progress tracking | Module progress | ✅ 有实现 |
| 布局 | Gallery/List/Timeline | ModulesView (列表) | ⚠️ ydsz 缺少 Gallery/Timeline 布局 |
| 归档/恢复 | Archive/restore | 无 | ❌ 缺失 |
| 导出 | Export | 无 | ❌ 缺失 |
| 分解视图 | Breakdowns | 无 | ❌ 缺失 |

### 2.4 视图 (Views)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 看板视图 | Board | KanbanBoardView | ✅ 有实现 |
| 列表视图 | List | IssueListView | ✅ 有实现 |
| 表格视图 | Spreadsheet | SpreadsheetView | ✅ 有实现 |
| 甘特图 | Gantt | GanttChartView | ✅ 有实现 |
| 日历视图 | Calendar | CalendarView | ✅ 有实现 |
| 时间线 | Timeline | TimelineView (ECharts) | ✅ 有实现 |
| WBS 树 | 无 | WbsTreeView | ✅ ydsz 独有 |
| 保存视图 | Saved views with filters | ViewsManager | ✅ 有实现 |
| 工作空间视图 | Workspace-level views | 无 | ❌ 缺失 |
| 发布到 Web | Publish to web (public link) | PublicPageView (Pages only) | ⚠️ ydsz 仅 Pages 可公开分享 |
| 锁定视图 | Lock view (Pro) | 无 | ❌ 缺失 |
| 收藏/复制 | Favorite/duplicate | Favorites store | ⚠️ ydsz 有 favorites 但无 duplicate view |

### 2.5 文档/知识库 (Pages / Wiki)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 编辑器类型 | Block editor (slash menu, DnD blocks) | RichTextEditor (基础) | ❌ **重大差距** — Plane 的 block editor 是核心差异化 |
| 块类型 | Lists/Layout/Media/Callouts/Dividers/Image/Link/Embedded work items/Link-pages/draw.io/Video/MS Office | 基础富文本 | ❌ 大量块类型缺失 |
| 实时协作 | Real-time co-editing (WebSocket) | 无 | ❌ 缺失 |
| 页面评论 | Page-level + inline comments (threads/reactions/mentions) | 无 | ❌ 缺失 |
| 版本历史 | Version history | KnowledgePageDetailView (有版本) | ✅ 有实现 |
| 页面分析 | Page analytics | 无 | ❌ 缺失 |
| 发布到 Web | Publish to web | PublicPageView | ✅ 有实现 |
| 导出 | Export | 无 | ❌ 缺失 |
| 页面标签 | Page labels | 无 | ❌ 缺失 |
| 嵌套页面 | Nested pages | KnowledgePageTreeNode | ✅ 有实现 |
| 页面模板 | Page templates | ContentTemplate | ⚠️ 有内容模板 |
| 数据库 (Database) | Spreadsheet-style grids in pages (Enterprise) | 无 | ❌ 缺失 |
| draw.io 图表 | Embedded diagrams | 无 | ❌ 缺失 |
| MS Office 编辑 | In-place Word/Excel/PPT editing | 无 | ❌ 缺失 |
| Wiki 层级 | Workspace-level wiki (Collections) | KnowledgeBase (spaces) | ✅ 基本对等 |
| AI 写作 | AI-assisted writing | 无 | ❌ 缺失 |

### 2.6 分析/仪表盘 (Analytics / Dashboards)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 工作空间分析 | Workspace analytics | WorkspaceAnalyticsView | ✅ 有实现 |
| 自定义仪表盘 | Custom dashboards (canvas + widgets) | DashboardView + widgetRegistry | ⚠️ ydsz 有部件注册但自由度较低 |
| 图表类型 | Bar/Line/Area/Donut/Pie/Number/Smart Counter/Gauge/Table | Bar/Line/Pie/StackedBar/Donut | ⚠️ 缺少 Area/Smart Counter/Gauge/Pivot |
| 工作项统计 | Work items statistics widget | 有 | ✅ |
| 周期分析 | Cycle analytics | SprintAnalyticsView | ✅ |
| 模块分析 | Module analytics | 无独立视图 | ⚠️ |
| 缺陷分析 | 无原生 | DefectAnalyticsView + DefectPanel | ✅ ydsz 独有优势 |
| DORA 指标 | 无原生 | MetricsView (DORA) | ✅ ydsz 独有优势 |
| 工作量热力图 | 无原生 | WorkloadHeatmapView | ✅ ydsz 独有优势 |
| 项目对比 | 无原生 | ProjectCompareWidget | ✅ ydsz 独有优势 |
| 效率指标 | 无原生 | velocity/lead time/quality/WIP | ✅ ydsz 独有优势 |
| 自定义属性过滤 | Custom property dashboard filters | 无 | ❌ 缺失 |

### 2.7 搜索 (Search)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 全局搜索 | ✅ (PQL + 自然语言) | ES-backed search | ✅ 基本对等 |
| 命令面板 | Power K (Cmd+K) | CommandPalette (Cmd+K) | ✅ 有实现 |
| PQL 查询语言 | Plane Query Language | 无 | ❌ 缺失 |
| AI 搜索 | Natural language → PQL | 无 | ❌ 缺失 |
| 搜索历史 | Search history | Search store | ✅ |
| 书签 | Bookmarks | 无 | ❌ 缺失 |
| 跨项目搜索 | Across projects | Workspace-scoped | ✅ |

### 2.8 通知 (Notifications)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 应用内通知 | Inbox/notifications | NotificationsView + NotificationBell | ✅ |
| 未读徽章 | Unread badge | ✅ | ✅ |
| 实时推送 | WebSocket | WebSocket hub + Redis Pub/Sub | ✅ |
| 通知偏好 | Notification preferences | NotificationPreferencesView | ✅ |
| 订阅者 | Subscribers | Watchers (M2M) | ⚠️ 概念接近但不完全相同 |
| 评论审核 | Comment moderation (hide with reason) | 无 | ❌ 缺失 |
| 摘要模式 | Digest modes | 有 | ✅ |
| Slack/Teams | Integration-level | 无 | ❌ 缺失 |

### 2.9 自动化 (Automation)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 事件规则 | Event-based custom automations | TCA rule engine (trigger/conditions/actions) | ✅ 有实现 |
| Plane Runner | Sandboxed JS/TS scripts | 无 | ❌ 缺失 |
| 执行日志 | Execution logs | Automation execution logs | ✅ |
| 失败处理 | Failure handling | Failure handling | ✅ |
| 脚本函数库 | Plane SDK + Functions library | 无 | ❌ 缺失 |
| Cron 调度 | Scheduled automations | 无 | ❌ 缺失 |

### 2.10 版本/发布 (Releases)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 版本生命周期 | Planning/Active/Released/Archived | planning/active/released/archived | ✅ 基本对等 |
| Semver 验证 | ✅ | semver.go | ✅ |
| 质量门禁 | 无原生 | Quality gate (critical bugs = 0) | ✅ ydsz 独有优势 |
| 发布说明 | Changelog editor | Release Notes generation | ✅ 有实现 |
| 交付报告 | 无原生 | DeliveryReportView | ✅ ydsz 独有优势 |
| 跨项目版本 | Workspace releases (Business) | Workspace versions hub | ✅ 有实现 |
| Sprint 聚合 | 无原生 | Sprint aggregation to version | ✅ ydsz 独有 |
| 检查清单 | Release checklist | Checklist all-required-checked | ✅ ydsz 独有 |
| 缺陷面板 | 无原生 | Defect panel per version | ✅ ydsz 独有 |

### 2.11 Intake (需求受理)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 来源渠道 | In-app/Email/Forms | Slug-based public channels | ⚠️ ydsz 缺少 Email 和 Forms |
| 审核流程 | Accept/decline/snooze/mark-duplicate | Accept/reject/convert | ✅ 基本对等 |
| 匿名追踪 | 无 | Anonymous tracking by ID | ✅ ydsz 独有 |
| 富文本 | ✅ | ✅ | ✅ |
| 附件 | ✅ | ✅ | ✅ |
| 导出 | CSV export | 无 | ❌ 缺失 |
| 容量统计 | Intake volume widget | 无 | ❌ 缺失 |

### 2.12 权限/RBAC

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 工作空间角色 | Owner/Admin/Member/Guest | Workspace roles | ✅ |
| 项目角色 | Admin/Contributor/Commenter/Guest | Project roles | ✅ |
| 自定义角色 | Custom roles | RBAC matrix (dynamic) | ⚠️ ydsz 有权限矩阵但无自定义角色名 |
| 条件授权 | Conditional grants | 无 | ❌ 缺失 |
| 权限方案 | Permission schemes | Permission matrix UI | ✅ 基本对等 |
| 审计日志 | Workspace + project audit | AuditReportView | ✅ |
| 敏感数据扫描 | Argus | 无 | ❌ 缺失 |

### 2.13 AI 能力 — **最大差距领域**

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| Plane AI | Ask/Build/Autopilot 三模式 | 无 | ❌ **完全缺失** |
| AI 记忆 | Persistent context (user/project/workspace) | 无 | ❌ 缺失 |
| AI Skills | Configurable skills | 无 | ❌ 缺失 |
| Applets | Mini apps | 无 | ❌ 缺失 |
| Agents | Configured AI assistants | 无 | ❌ 缺失 |
| Agent Runs | Full lifecycle tracking | 无 | ❌ 缺失 |
| MCP Server | 30 tools, 207 actions | 无 | ❌ 缺失 |
| AI 辅助写作 | Page editor AI | 无 | ❌ 缺失 |
| 智能分配 | ✅ | smart assign recommendation | ✅ ydsz 有基础版 |
| 查重 | ✅ | duplicate detection | ✅ ydsz 有基础版 |
| 分类 | ✅ | auto-classification | ✅ ydsz 有基础版 |
| 摘要 | ✅ | text summarization | ✅ ydsz 有基础版 |
| 语音输入 | Voice input | 无 | ❌ 缺失 |
| Web 搜索 | OpenAI/Anthropic web search | 无 | ❌ 缺失 |

### 2.14 交互体验 (UX)

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| 命令面板 | Power K (context-aware actions) | CommandPalette | ✅ 有实现 |
| 快捷键 | Full keyboard flow (S/P/A/I/L...) | useKeyboardShortcuts (?/c//) | ⚠️ ydsz 快捷键较少 |
| Peek 视图 | Side panel peek | IssuePeekOverview drawer | ✅ 有实现 |
| 侧边栏 AI | Sidecar for Plane AI | 无 | ❌ 缺失 |
| 拖放编辑器 | DnD block editor | 无 | ❌ 缺失 |
| Propel 设计系统 | Unified design system | Element Plus + 自研组件 | ⚠️ 设计风格差异大 |
| 暗色模式 | ✅ | ThemeToggle | ✅ |
| 品牌色 | Workspace icons | useBrandColor | ✅ |
| 收藏 | Favorites | Favorites store | ✅ |
| 最近访问 | Recent visited | RecentVisited store | ✅ |
| 全局搜索触发 | Cmd+K | Cmd+K | ✅ |
| 侧边栏折叠 | Cmd+B | Sidebar collapse | ✅ |
| 移动端 | iOS/Android apps | Mobile responsive drawer | ⚠️ ydsz 仅响应式，无原生 App |
| 桌面端 | Mac/Windows apps | 无 | ❌ 缺失 |
| 快速创建 | N then letter | Quick-create button | ⚠️ 交互方式不同 |
| 导航键 | G then letter | 无 | ❌ 缺失 |

### 2.15 集成与开放

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| REST API | 180+ endpoints | Gin REST API | ✅ 有实现 |
| API 鉴权 | X-API-Key/OAuth Bearer/IdP | JWT + API Tokens | ⚠️ 缺少 OAuth |
| Webhook | HMAC-signed | Webhook outbound | ✅ |
| OAuth Apps | Build Plane apps | 无 | ❌ 缺失 |
| SDK | Node.js + Python | 无 | ❌ 缺失 |
| Marketplace | Apps/Agents/Importers/Templates | 无 | ❌ 缺失 |
| GitHub 集成 | ✅ | 无 | ❌ 缺失 |
| GitLab 集成 | ✅ | 无 | ❌ 缺失 |
| Slack 集成 | @Plane bot | 无 | ❌ 缺失 |
| Sentry 集成 | ✅ | 无 | ❌ 缺失 |
| 导入器 | Jira/Asana/Linear/ClickUp/Monday/Notion/Confluence/CSV | ImportDialog (基础) | ⚠️ ydsz 有导入但非专用连接器 |
| MCP Server | Native MCP server | 无 | ❌ 缺失 |
| Projects-as-Code | Plane Compose (YAML) | 无 | ❌ 缺失 |

### 2.16 企业/治理特性

| 功能点 | Plane | ydsz-plane | 差距 |
|--------|-------|------------|------|
| Initiatives | ✅ (workspace-level) | 无 | ❌ 缺失 |
| Epics as type | ✅ (level 1) | Epic type code | ⚠️ 概念接近 |
| Teamspaces | ✅ (team + projects) | 无 | ❌ 缺失 |
| Customer Requests | ✅ (customer records) | 无 | ❌ 缺失 |
| Project States + Updates | ✅ | 无 | ❌ 缺失 |
| Milestones | ✅ | 无独立实体 | ❌ 缺失 |
| Workspace Governance | ✅ | 无 | ❌ 缺失 |
| God Mode | Instance admin panel | DLQMonitoring + AuditReport | ⚠️ 部分对应 |
| Workflows & Approvals | ✅ (state transitions + approval gates) | StateMachineConfigView | ⚠️ ydsz 有状态机但无审批流 |
| SSO/SAML/LDAP | ✅ | OIDC + SAML | ✅ 基本对等 |
| SMTP 配置 | God Mode | 无显式 | ❌ 缺失 |

---

## 三、差距汇总（按优先级）

### P0 — 核心体验差距（影响 1:1 复刻目标）

1. **Block Editor（块编辑器）**：Plane 的核心差异化，支持 slash 菜单、拖放块、多种块类型。ydsz 仅有基础 RichTextEditor。
2. **AI 能力整体缺失**：Plane AI（Ask/Build/Autopilot）、Agents、MCP Server 是 Plane 近年的核心卖点，ydsz 仅有智能分配/查重/分类/摘要四个辅助功能。
3. **Initiatives（倡议）**：Plane 的 Epics + Initiatives 层级体系是产品规划的核心，ydsz 缺少 Initiative 实体。
4. **Workflows & Approvals（审批流）**：Plane Business 以上的核心功能，ydsz 仅有状态机配置但无审批门控。
5. **Teamspaces（团队空间）**：Plane Pro 的核心组织单元，ydsz 完全没有。

### P1 — 重要功能差距

6. **自定义工作项类型**：Plane 允许项目/工作空间级别定义类型，ydsz 固定为 Requirement/Task/Defect。
7. **实时协作编辑**：Plane Pages 支持多人实时共编，ydsz 无。
8. **页面评论系统**：页面级 + 行内评论（线程/reactions/mentions），ydsz 无。
9. **Projects-as-Code (Plane Compose)**：YAML 定义项目，Git 版本化。
10. **Marketplace + Apps/Integrations**：GitHub/GitLab/Slack/Sentry 等集成。
11. **PQL 查询语言**：结构化查询 + AI 辅助生成。
12. **Customer Requests（客户需求）**：客户记录 + 请求关联。
13. **Project States + Updates（项目状态与更新）**：On Track/At Risk/Off Track。
14. **OAuth Apps + SDK**：第三方开发生态。
15. **移动端 App + 桌面端 App**。

### P2 — 增强体验差距

16. **并行周期 (Parallel Cycles)**
17. **模块状态机 + 归档/导出**
18. **工作空间级视图 (Workspace Views)**
19. **视图发布到 Web + 锁定**
20. **评论审核 (Comment Moderation)**
21. **重复工作项 (Recurring Work Items)**
22. **草稿工作项 (Drafts)**
23. **URL 创建工作项**
24. **Milestones（里程碑）独立实体**
25. **页面分析 (Page Analytics)**
26. **数据库块 (Database Blocks)**
27. **draw.io 嵌入**
28. **MS Office 在线编辑**
29. **Prime CLI**
30. **条件授权 (Conditional Grants)**
31. **敏感数据扫描 (Argus)**

---

## 四、ydsz-plane 独有优势（Plane 不具备或较弱的部分）

| 功能 | 说明 |
|------|------|
| **WBS 树视图** | 工作分解结构树形展示 |
| **缺陷分析** | 独立缺陷分析报表 + 面板 |
| **DORA 指标** | 部署频率/变更前置时间/MTTR/变更失败率 |
| **工作量热力图** | 人员工作负载可视化 |
| **项目对比** | 多项目横向对比 |
| **效率指标** | 流速/前置时间/质量/WIP |
| **版本质量门禁** | Critical bugs = 0 才能发布 |
| **版本检查清单** | Release checklist |
| **交付报告** | 版本交付报告 |
| **Sprint Standup 模式** | 站会视图 |
| **异步任务监控 (DataJob)** | 导入/导出/同步/报告任务 |
| **DLQ 死信队列监控** | 消息队列治理 |
| **审计报表** | 完整审计报告视图 |
| **Intake 匿名追踪** | 通过追踪 ID 匿名查询进度 |
| **状态机配置** | 项目级状态流转规则 + 必填字段 |
| **SM 国密抽象** | 信创合规支持 |
| **PostgreSQL RLS** | 行级安全多租户 |

---

## 五、结论与建议

### 复刻度总结

ydsz-plane 在**核心工作流**上已经实现了 Plane 约 70% 的功能覆盖，特别是在工作项管理、Sprint/Cycle、版本管理、视图系统、知识库、搜索、通知、自动化等方面有较好的对应。同时，ydsz-plane 在**研发效能度量**（DORA/缺陷分析/工作量热力图/项目对比）方面甚至超越了 Plane。

但距离 **1:1 复刻**还有显著差距，主要集中在三大领域：

1. **Block Editor 与文档协作**（P0）——Plane 的文档编辑器是其核心体验差异化，需要从架构层面重建。
2. **AI 原生能力**（P0）——Plane AI/Agents/MCP 是 Plane 近年来的核心战略方向，ydsz 的四个辅助功能仅能覆盖冰山一角。
3. **企业级组织单元**（P0）——Initiatives/Teamspaces/Workflows & Approvals 是 Plane 向上突破企业市场的关键。

### 建议优先补齐顺序

若以"达到 1:1 复刻"为目标，建议按以下优先级推进：

**第一阶段（核心交互体验对齐）：**
- 引入 Block Editor（可基于 TipTap/ProseMirror 二次开发）
- 补全键盘快捷键体系（对齐 Plane 的 S/P/A/I/L 等）
- 添加 Initiatives 实体 + 进度 Rollup
- 实现 Workflows & Approvals 审批流

**第二阶段（AI 能力建设）：**
- 接入 LLM 实现 Ask/Build 模式（可基于现有 ai 服务层扩展）
- 实现 MCP Server（让外部 Agent 可操作 ydsz-plane）
- AI 辅助写作（Block Editor 内）

**第三阶段（生态与集成）：**
- GitHub/GitLab 集成
- Slack/Teams 集成
- OAuth Apps + SDK
- 移动端适配（PWA 或 Capacitor）

**第四阶段（企业级特性）：**
- Teamspaces
- Customer Requests
- Projects-as-Code
- Marketplace
