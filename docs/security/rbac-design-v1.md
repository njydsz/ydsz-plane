# Ydsz Plane RBAC 权限体系设计方案

> 版本: v1.0 | 日期: 2026-08-08 | 状态: 待确认

---

## 一、角色定义（8 个标准角色）

| 角色 ID | 角色名称 | 身份描述 | 层级 | 对标业界 |
|---------|----------|----------|------|----------|
| `owner` | **管理员 / 空间所有者** | 空间最高权限，唯一且不可被移除。管理成员、空间设置、审计日志、归档删除空间。 | L5 | Linear Owner / Jira Admin / Plane Owner |
| `admin` | **管理员（联合）** | 被 owner 授予管理权限，除"删除空间/转移 ownership"外拥有全部管理权。可管理成员、项目、工作项、迭代、自动化。 | L4 | Linear Admin / Plane Admin |
| `pm` | **项目经理** | 负责项目全生命周期：创建/归档项目、管理迭代、查看效能报表、管理工作项状态流转、管理版本发布。 | L4 | Jira Project Manager / Asana Admin |
| `po` | **产品经理** | 需求侧负责人：创建/编辑需求、设置优先级与验收标准、管理产品路线图（版本）、查看分析报表。不可删除项目。 | L3 | Jira Product Manager |
| `techlead` | **技术经理 / Tech Lead** | 技术侧负责人：管理 sprint 排期、技术视角的状态流转、代码关联、自动化规则、Webhook、效能度量。 | L3 | Jira Developer Lead |
| `qalead` | **测试经理 / QA Lead** | 质量侧负责人：创建/编辑缺陷、管理缺陷分类/严重度、查看缺陷分析报表、管理质量门禁。 | L3 | Jira QA Lead |
| `dev` | **开发（前端/后端统一）** | 执行者：创建/编辑分配给自己的工作项、更新状态、记录工时、添加评论。不可删除他人工作项。 | L2 | Linear Member / Jira Developer |
| `guest` | **访客 / 只读协作者** | 受邀外部人员：只读浏览指定项目、添加评论。无任何编辑与管理权限。 | L1 | Linear Guest / Jira User |

### 关键设计决策

1. **前端/后端开发统一为 `dev` 角色**：两者的操作权限在工作项层面无差异（都是"编辑自己的工作项"），差异体现在实际工作分配而非工具权限。如有特殊需求可在字段级权限进一步区分。
2. **层级制度**：L5 > L4 > L3 > L2 > L1，高级别自动包含低级别的所有读取权限。
3. **单一角色模型**：每位成员在一个 workspace 只有一个角色（简化 Jira 式多重角色的认知负担）。

---

## 二、菜单权限矩阵（侧栏导航可见性）

> 标记：✅ 可见 | ❌ 不可见

| 菜单项 | 路径 | owner | admin | pm | po | techlead | qalead | dev | guest |
|--------|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| **工作台** | `/:wsId/workbench` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **全局搜索** | `/:wsId/search` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **项目列表** | `/:wsId/projects` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ↳ 仪表盘 | `projects/:pid/dashboard` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ↳ 看板视图 | `projects/:pid/board` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ↳ 列表视图 | `projects/:pid/list` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ↳ 文档/Pages | `projects/:pid/pages` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ↳ 效能度量 | `projects/:pid/metrics` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| ↳ 缺陷分析 | `projects/:pid/analytics` | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| ↳ 甘特图 | `projects/:pid/gantt` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| ↳ 日历视图 | `projects/:pid/calendar` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| ↳ 自动化规则 | `projects/:pid/automation` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **迭代管理** | | | | | | | | | |
| ↳ 迭代列表 | `projects/:pid/sprints` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ✅ |
| ↳ 排期规划 | `projects/:pid/sprints/planning` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **版本管理** | | | | | | | | | |
| ↳ 版本列表 | `projects/:pid/versions` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ |
| ↳ 发布单 | `.../:vid/release` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| ↳ 交付报告 | `.../:vid/delivery-report` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **设置区域** | | | | | | | | | |
| ↳ 工作空间设置 | `/:wsId/settings` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| ↳ 通知偏好 | `/:wsId/settings/notifications` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ↳ Webhook | `/:wsId/settings/webhooks` | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| ↳ 收件箱 | `/:wsId/settings/intake` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| ↳ 审计日志 | `/:wsId/audit-logs` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## 三、按钮/操作权限矩阵

> 标记：✅ 允许 | ❌ 禁止

### 3.1 工作空间级操作

| 操作 | owner | admin | pm | po | techlead | qalead | dev | guest |
|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| 创建工作空间 | 平台级 | 平台级 | — | — | — | — | — | — |
| 编辑空间设置（名称/时区/Logo） | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 归档/删除工作空间 | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 邀请成员 | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 移除成员 | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 修改成员角色 | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 查看所有成员 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 查看审计日志 | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 管理 Webhook | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| 管理收件箱 | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |

### 3.2 项目级操作

| 操作 | owner | admin | pm | po | techlead | qalead | dev | guest |
|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| 创建项目 | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 编辑项目（名称/描述/图㤐） | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 归档/删除项目 | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 管理项目模板 | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 导出项目数据 | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

### 3.3 工作项（Issue）操作

| 操作 | owner | admin | pm | po | techlead | qalead | dev | guest |
|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| 创建需求 | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| 创建任务 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| 创建缺陷 | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ |
| 编辑自己的工作项 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| 编辑他人的工作项 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| 删除/归档工作项 | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| 变更状态（流转） | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| 变更指派人 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| 变更优先级 | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| 变更迭代归属 | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| 记录工时 | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ |
| 添加评论 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 删除自己评论 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 删除他人评论 | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 管理关联关系 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |

### 3.4 迭代（Sprint）操作

| 操作 | owner | admin | pm | po | techlead | qalead | dev | guest |
|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| 创建迭代 | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| 编辑迭代（时间/目标） | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| 开始/结束迭代 | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| 删除迭代 | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| 排期规划（拖拽 Backlog↔Sprint） | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| 查看站会模式 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |

### 3.5 版本（Version）操作

| 操作 | owner | admin | pm | po | techlead | qalead | dev | guest |
|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| 创建版本 | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| 编辑版本 | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| 发布版本 | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| 删除版本 | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

### 3.6 自动化 & 效能

| 操作 | owner | admin | pm | po | techlead | qalead | dev | guest |
|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| 创建/编辑自动化规则 | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| 查看效能度量 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| 查看缺陷分析 | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| 上报部署事件（DORA） | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |

---

## 四、字段编辑权限矩阵（工作项详情页）

> 标记：✅ 可编辑 | 👁 只读可见 | ❌ 不可见

| 字段 | owner | admin | pm | po | techlead | qalead | dev | guest |
|------|:-----:|:-----:|:--:|:--:|:--------:|:------:|:---:|:-----:|
| 标题 (name) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 👁 |
| 描述 (description) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 👁 |
| 状态 (state) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 👁 |
| 优先级 (priority) | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | 👁 |
| 指派给 (assignee) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | 👁 |
| 所属迭代 (sprint) | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | 👁 |
| 所属版本 (version) | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | 👁 |
| 严重度 (severity) | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | 👁 |
| 缺陷分类 (root_cause) | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | 👁 |
| 发现阶段 (found_phase) | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | 👁 |
| 故事点/点数 (point) | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | 👁 |
| 预估工时 (actual_effort) | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ | 👁 |
| 截止日期 (target_end_date) | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | 👁 |
| 标签/标记 (labels) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 👁 |
| 自定义字段 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 👁 |

---

## 五、权限点常量清单（Permission Codes）

```go
// 读取权限（层级递进）
permWorkspaceRead      = "workspace:read"       // 读取工作空间基本信息
permProjectRead        = "project:read"         // 读取项目列表与详情
permIssueRead          = "issue:read"           // 读取工作项
permSprintRead         = "sprint:read"          // 读取迭代
permVersionRead        = "version:read"         // 读取版本

// 工作空间管理
permWorkspaceUpdate    = "workspace:update"     // 修改空间设置
permWorkspaceDelete    = "workspace:delete"     // 归档/删除空间

// 成员管理
permMemberInvite       = "member:invite"        // 邀请成员
permMemberRemove       = "member:remove"        // 移除成员
permMemberChangeRole   = "member:change_role"   // 修改成员角色

// 项目管理
permProjectCreate      = "project:create"       // 创建项目
permProjectUpdate      = "project:update"       // 编辑项目
permProjectDelete      = "project:delete"       // 归档/删除项目

// 工作项管理
permIssueCreate        = "issue:create"         // 创建工作项
permIssueEditOwn       = "issue:edit_own"       // 编辑自己的工作项
permIssueEditAll       = "issue:edit_all"       // 编辑所有工作项
permIssueDelete        = "issue:delete"         // 删除/归档工作项
permIssueTransition    = "issue:transition"     // 状态流转
permIssueReassign      = "issue:reassign"       // 变更指派人
permIssueChangePriority = "issue:change_priority" // 变更优先级
permIssueManageSprint  = "issue:manage_sprint"  // 变更迭代归属
permCommentModerate    = "comment:moderate"     // 删除他人评论
permRelationManage     = "relation:manage"      // 管理关联关系

// 迭代管理
permSprintCreate       = "sprint:create"        // 创建迭代
permSprintUpdate       = "sprint:update"        // 编辑迭代
permSprintDelete       = "sprint:delete"        // 删除迭代
permSprintLifecycle    = "sprint:lifecycle"     // 开始/结束迭代
permSprintPlan         = "sprint:plan"          // 排期规划

// 版本管理
permVersionCreate      = "version:create"       // 创建版本
permVersionUpdate      = "version:update"       // 编辑版本
permVersionRelease     = "version:release"      // 发布版本
permVersionDelete      = "version:delete"       // 删除版本

// 质量
permDefectCreate       = "defect:create"        // 创建缺陷
permQAReport           = "qa:report"            // 查看缺陷分析报表

// 效能 & 自动化
permProjectAnalytics   = "analytics:read"       // 查看效能度量
permAnalyticsExport    = "analytics:export"     // 导出分析数据
permAutomationManage   = "automation:manage"    // 管理自动化规则
permDeployReport       = "deploy:report"        // 上报部署事件

// 审计 & 集成
permAuditRead          = "audit:read"           // 查看审计日志
permWebhookManage      = "webhook:manage"       // 管理 Webhook
permIntakeManage       = "intake:manage"        // 管理收件箱
permPagesManage        = "pages:manage"         // 管理知识库

// 字段级权限
permFieldEditSeverity  = "field:edit_severity"  // 编辑严重度
permFieldEditEffort    = "field:edit_effort"    // 编辑工时
permFieldEditDeadline  = "field:edit_deadline"  // 编辑截止日期
```

---

## 六、角色-权限映射矩阵

```
┌──────────────────────┬───────┬───────┬────┬────┬────────┬───────┬─────┬───────┐
│ Permission           │ owner │admin │ pm │ po │techlead│qalead │ dev │ guest │
├──────────────────────┼───────┼───────┼────┼────┼────────┼───────┼─────┼───────┤
│ workspace:read       │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ✅ │  ✅   │
│ workspace:update     │  ✅   │  ✅   │ ❌ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ workspace:delete     │  ✅   │  ❌   │ ❌ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ member:invite        │  ✅   │  ✅   │ ❌ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ member:remove        │  ✅   │  ✅   │ ❌ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ member:change_role   │  ✅   │  ✅   │ ❌ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ project:create       │  ✅   │  ✅   │ ✅ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ project:update       │  ✅   │  ✅   │ ✅ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ project:delete       │  ✅   │  ✅   │ ✅ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ issue:create         │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ✅ │  ❌   │
│ issue:edit_own       │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ✅ │  ❌   │
│ issue:edit_all       │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ❌ │  ❌   │
│ issue:delete         │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ✅   │  ❌ │  ❌   │
│ issue:transition     │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ✅ │  ❌   │
│ issue:reassign       │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ❌ │  ❌   │
│ issue:change_priority│  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ❌   │  ❌ │  ❌   │
│ issue:manage_sprint  │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ comment:moderate     │  ✅   │  ✅   │ ❌ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ relation:manage      │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ❌ │  ❌   │
│ sprint:create        │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ sprint:update        │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ sprint:delete        │  ✅   │  ✅   │ ✅ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ sprint:lifecycle     │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ sprint:plan          │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ version:create       │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ❌   │  ❌ │  ❌   │
│ version:update       │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ❌   │  ❌ │  ❌   │
│ version:release      │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ❌   │  ❌ │  ❌   │
│ version:delete       │  ✅   │  ✅   │ ✅ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ defect:create        │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ✅   │  ✅ │  ❌   │
│ qa:report            │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ✅   │  ❌ │  ❌   │
│ analytics:read       │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ✅   │  ❌ │  ❌   │
│ analytics:export     │  ✅   │  ✅   │ ✅ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ automation:manage    │  ✅   │  ✅   │ ✅ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ deploy:report        │  ✅   │  ✅   │ ❌ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ audit:read           │  ✅   │  ✅   │ ❌ │ ❌ │   ❌   │  ❌   │  ❌ │  ❌   │
│ webhook:manage       │  ✅   │  ✅   │ ❌ │ ❌ │   ✅   │  ❌   │  ❌ │  ❌   │
│ intake:manage        │  ✅   │  ✅   │ ✅ │ ✅ │   ❌   │  ❌   │  ❌ │  ❌   │
│ pages:manage         │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ❌   │  ❌ │  ❌   │
│ field:edit_severity  │  ✅   │  ✅   │ ❌ │ ❌ │   ✅   │  ✅   │  ✅ │  ❌   │
│ field:edit_effort    │  ✅   │  ✅   │ ❌ │ ❌ │   ✅   │  ❌   │  ✅ │  ❌   │
│ field:edit_deadline  │  ✅   │  ✅   │ ✅ │ ✅ │   ✅   │  ❌   │  ❌ │  ❌   │
└──────────────────────┴───────┴───────┴────┴────┴────────┴───────┴─────┴───────┘
```

---

## 七、数据库变更计划

### 7.1 workspace_members 表变更

```sql
-- 1. 增强 workspace_members 表结构
ALTER TABLE workspace_members 
  ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'member';

-- 2. 确保 role 为合法枚举（通过 CHECK 约束）
ALTER TABLE workspace_members 
  DROP CONSTRAINT IF EXISTS chk_workspace_member_role;
ALTER TABLE workspace_members 
  ADD CONSTRAINT chk_workspace_member_role 
  CHECK (role IN ('owner','admin','pm','po','techlead','qalead','dev','guest'));

-- 3. 添加 is_active 字段（已有则跳过）
ALTER TABLE workspace_members 
  ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

-- 4. 添加审计字段
ALTER TABLE workspace_members 
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS created_by INT8 REFERENCES users(id);
```

### 7.2 权限不存 DB，固化在代码层

> 决策依据：当前 RBAC 模型量级 8 角色 × ~40 权限 ≈ 320 条映射。这种规模直接硬编码在 Go map 中是最佳实践 —— 每次请求查询 DB 查权限是多此一举，Plane/Linear 全部采用代码固化。未来如需自定义角色再考虑 DB 化。

---

## 八、代码改造清单

### 8.1 后端改造

| 改造项 | 文件 | 改造内容 |
|--------|------|----------|
| 权限常量扩展 | `internal/auth/rbac.go` | 从 18 个扩展到 ~40 个 Perm 常量 |
| 角色枚举扩展 | `internal/auth/rbac.go` | 新增 RolePM, RolePO, RoleTechLead, RoleQALead, RoleDev |
| 权限矩阵扩展 | `internal/auth/rbac.go` | Roles map 扩写 8 角色 × 权限映射 |
| 中间件改造 | `internal/middleware/rbac.go` | 支持新角色解析；新增"字段级权限检查"API handler |
| RoleMembershipStore | `internal/auth/rbac.go` | HasPermission 保持不变；可选新增 HasAnyPermission |
| API: GET /role | `internal/interfaces/http/router.go` | 新增 `ws.GET("/workspaces/:id/role")` 返回当前角色 |
| Handler 分流 | `internal/interfaces/http/handlers.go` | 按新权限常量对现有 endpoint 重新标注 |
| API Token scope | `internal/application/apitoken` | PermissionScope 映射补齐新权限 |

### 8.2 前端改造

| 改造项 | 文件 | 改造内容 |
|--------|------|----------|
| 类型定义 | `web/src/types/permission.ts`（新建） | 导出 Perm 类型与 roles/permissions 映射 |
| Workspace Store | `web/src/stores/workspace.ts` | 新增 `role` 状态、`hasPermission(perm)` 方法、`canReadIssue/canEditIssue/getters` |
| 路由守卫 | `web/src/router/index.ts` | 增加 ws 加载校验，无成员记录时跳 /forbidden |
| 侧栏菜单 | `web/src/layouts/WorkspaceLayout.vue` | v-if 基于权限条件渲染：`v-if="ws.hasPermission('analytics:read')"` |
| 权限组件（新建）| `web/src/components/Permission.vue` | `<Permission perm="issue:edit_all">...</Permission>` 包裹式组件 |
| 权限指令（新建）| `web/src/directives/permission.ts` | `v-permission="'issue:delete'"` 指令式控制按钮显示 |
| IssueDetailView | `web/src/views/project/IssueDetailView.vue` | 按权限控制：编辑字段、删除按钮、流转按钮 |
| ListView 操作列 | `web/src/views/project/IssueListView.vue` | 按权限控制行内操作按钮 |
| ForbiddenView（新建）| `web/src/views/ForbiddenView.vue` | 403 专用提示页："你没有权限访问此空间 / 项目" |
| API 封装 | `web/src/api/services/auth.ts` | 新增 `getWorkspaceRole(wsId)` API 方法 |
| 错误处理增强 | `web/src/api/client.ts` | 403 响应统一 toast + 可选自动跳转 |

---

## 九、实施里程碑

| 阶段 | 内容 | 影响范围 |
|------|------|----------|
| **Phase 1: 数据准备** | ALTER TABLE + 存量角色迁移（owner/admin/member/guest 映射到新增的 PM/TechLead 等） | DB schema |
| **Phase 2: 后端 RBAC 扩展** | 写新 Roles map + Perm 常量 + 中间件适配 | 后端编译通过 |
| **Phase 3: 前端权限基建** | 类型 + Store + 组件/指令 + 路由守卫 | 前端编译通过 |
| **Phase 4: 视图权限贴敷** | 各 vue 文件加 v-if / v-permission | 视觉层 |
| **Phase 5: E2E 测试** | 编写各角色的 smoke test | CI |

---

## 十、关键风险与对策

| 风险 | 影响 | 对策 |
|------|------|------|
| 存量数据 `role` 文本值为 `owner/admin/member/guest` 与新增枚举并存 | 查询/校验失败 | 存量数据只保留已有枚举，新增角色通过后台管理 UI 分配 |
| 硬编码权限 vs 运营可配置矛盾 | 未来需支持"自定义角色" | 架构预留：权限检查抽象为接口，当前实现为 CodePermissionChecker，未来可替换为 DB-backed |
| 前端侧栏 ≠ 后端权限 | 不能仅靠前端隐藏保障安全 | 前端隐藏仅优化体验，安全由后端 RBAC 中间件保障（已有） |
| `dev` 前端/后端区分 | 部分团队可能需要更细粒度 | 如需区分，可在未来新增 `fe_dev` / `be_dev` 扩展角色（schema 预留了 20 字符空间） |
