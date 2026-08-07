# Ydsz Plane API 参考 (v0.2)

> RESTful API，所有端点在 `/api/v1/` 前缀下。
> 认证方式：`Authorization: Bearer <access_token>` 或 Cookie 会话。

---

## 通用约定

### 响应格式
```json
// 成功
{ "data": { ... } }

// 错误（统一）
{ "error": { "code": "NOT_FOUND", "message": "workspace not found" } }
```

### 分页
列表接口统一返回：
```json
{
  "results": [...],
  "total": 1234,
  "next": "?cursor=eyJpZCI6MTAwfQ==",
  "previous": null
}
```

### 状态码
| Code | 含义 |
|------|------|
| 200 | 成功 |
| 201 | 已创建 |
| 204 | 删除成功无返回 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 冲突（如 slug 重复） |
| 429 | 限流（附 Retry-After） |
| 500 | 内部错误 |

---

## 认证 (Public)

| Method | Path | 说明 |
|--------|------|------|
| POST | `/auth/register` | 注册 |
| POST | `/auth/login` | 登录，返回 access_token + refresh_token |
| POST | `/auth/refresh` | 刷新 access_token |
| POST | `/auth/logout` | 登出 |
| POST | `/auth/forgot-password` | 请求密码重置 |
| POST | `/auth/reset-password` | 提交密码重置 |
| GET | `/invitations/:token` | 邀请预览（公开） |
| POST | `/invitations/accept` | 接受邀请 |

---

## 工作空间 (Workspace)

| Method | Path | 权限 | 说明 |
|--------|------|------|------|
| GET | `/workspaces` | - | 列出当前用户工作空间 |
| POST | `/workspaces` | - | 创建工作空间 |
| GET | `/workspaces/:id` | workspace:read | 获取详情 |
| PATCH | `/workspaces/:id` | workspace:update | 更新 |
| DELETE | `/workspaces/:id` | workspace:delete | 归档 |
| GET | `/workspaces/slug/:slug` | - | 按 slug 查找 |
| GET | `/workspaces/:ws/members` | workspace:read | 列出成员 |
| PATCH | `/workspaces/:ws/members/:uid` | member:change_role | 修改角色 |
| DELETE | `/workspaces/:ws/members/:uid` | member:remove | 移除成员 |
| POST | `/workspaces/:ws/invitations` | member:invite | 发送邀请 |
| GET | `/workspaces/:ws/invitations` | - | 列出邀请 |
| DELETE | `/workspaces/:ws/invitations/:iid` | member:invite | 撤销邀请 |
| GET | `/workspaces/:ws/audit-logs` | audit:read | 审计日志 |
| GET | `/workspaces/:ws/templates` | project:create | 列出项目模板 |
| GET | `/workspaces/:ws/projects` | workspace:read | 列出项目 |
| POST | `/workspaces/:ws/projects` | project:create | 创建项目 |
| GET | `/workspaces/:ws/projects/:pid` | workspace:read | 获取项目 |
| PATCH | `/workspaces/:ws/projects/:pid` | project:create | 更新项目 |
| DELETE | `/workspaces/:ws/projects/:pid` | project:delete | 归档项目 |

---

## 项目内资源 (Project-scoped)

### 工作项 (Issue)
| Method | Path | 说明 |
|--------|------|------|
| GET | `.../projects/:pid/issues?page=&limit=&type_code=&state_id=&priority=&q=` | 列表 |
| POST | `.../projects/:pid/issues` | 创建 |
| GET | `.../projects/:pid/issues/:iid` | 详情 |
| PATCH | `.../projects/:pid/issues/:iid` | 更新 |
| DELETE | `.../projects/:pid/issues/:iid` | 删除（软删除） |
| POST | `.../projects/:pid/issues/:iid/transitions/:tsid` | 状态流转 |
| GET | `.../projects/:pid/issues/:iid/comments` | 评论列表 |
| POST | `.../projects/:pid/issues/:iid/comments` | 发表评论 |

### 迭代 (Sprint)
| Method | Path | 说明 |
|--------|------|------|
| GET | `.../projects/:pid/sprints` | 列表 |
| POST | `.../projects/:pid/sprints` | 创建 |
| GET | `.../projects/:pid/sprints/:sid` | 详情 |
| POST | `.../projects/:pid/splits/:sid/start` | 启动 |
| POST | `.../projects/:pid/sprints/:sid/complete` | 完成 |
| POST | `.../projects/:pid/sprints/:sid/issues` | 加入工作项 |

### 版本 (Version)
| Method | Path | 说明 |
|--------|------|------|
| GET | `.../projects/:pid/versions` | 列表 |
| POST | `.../projects/:pid/versions` | 创建 |
| GET | `.../projects/:pid/versions/:vid` | 详情 |
| POST | `.../projects/:pid/versions/:vid/release` | 发布 |

### 模块 / 标签 / 状态
| Method | Path | 说明 |
|--------|------|------|
| GET/POST | `.../projects/:pid/modules` | 模块 CRUD |
| GET/POST | `.../projects/:pid/labels` | 标签 CRUD |
| GET/POST | `.../projects/:pid/states` | 状态定义 |

## 工作台 & 仪表盘 (Read)

| Method | Path | 说明 |
|--------|------|------|
| GET | `.../projects/:pid/workbench/summary` | 我的工作汇总 |
| GET | `.../workspaces/:ws/workbench/summary` | 工作空间级汇总 |
| GET | `.../projects/:pid/dashboard/widgets` | 仪表盘卡片 |
| GET | `.../projects/:pid/search?q=` | 项目内搜索 |
| GET | `.../workspaces/:ws/search?q=` | 全局搜索 |

---

## Webhook (Public Inbound)

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/v1/public/intake/channels/:ws/:slug/submit` | 公开提交工单 |
| GET | `/api/v1/public/intake/track?ref=` | 匿名查询处理状态 |

Webhook 出站事件签名：
```
X-Signature: sha256=<HMAC-SHA256(secret, timestamp+body)>
X-Timestamp: 1722988800
```

---

## 效能度量 (S11)

| Method | Path | 说明 |
|--------|------|------|
| GET | `.../projects/:pid/metrics/velocity` | 速度趋势 |
| GET | `.../projects/:pid/metrics/lead-time` | 前置时间 |
| GET | `.../projects/:pid/metrics/quality` | 质量指标 |
| GET | `.../projects/:pid/metrics/dora` | DORA 四指标 |
| POST | `.../projects/:pid/metrics/deployments` | 上报部署事件 |

---

## WebSocket

```
wss://<host>/ws/:workspace_id
```

消息类型：`issue.updated` / `issue.created` / `notification.new` / ...

---

## API Token

| Method | Path | 说明 |
|--------|------|------|
| POST | `/workspaces/:ws/api-tokens` | 创建 Token |
| GET | `/workspaces/:ws/api-tokens` | 列出 |
| DELETE | `/workspaces/:ws/api-tokens/:tid` | 吊销 |

Scopes：`read:issues` / `write:issues` / `read:projects` / `admin:*`

---

## Swagger UI（开发环境）

浏览器访问 `/swagger/index.html`，可在线调试全部 API。
