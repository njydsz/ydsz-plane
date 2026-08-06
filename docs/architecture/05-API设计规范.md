# 05 · API 设计规范（S2 起全局生效）

> 参考：Microsoft REST API Guidelines、Google API Design Guide (AIP)、GitHub API v3/v4 实践、OpenAPI 3.0

---

## 1. 总则

- **风格**：REST 资源导向；URL 只含名词，动作用子资源或自定义操作（`POST .../issues/123:transition`）。
- **版本**：路径前缀 `/api/v1`；破坏性变更升级 `/api/v2`，v1 至少并行维护 12 个月。
- **协议**：全 HTTPS；请求/响应 `application/json; charset=utf-8`；时间一律 RFC3339（UTC，前端本地化）。
- **Envelope**：

```json
// 成功（单资源/集合直接返回数据，不套 data 壳，与 GitHub API 一致）
{ "id": 1, "name": "..." }

// 失败（统一错误壳）
{
  "error": {
    "code": "ISSUE.INVALID_STATE_TRANSITION",
    "message": "当前状态不允许流转到「已关闭」",
    "details": [{ "field": "state_id", "reason": "missing transition 5->9" }],
    "request_id": "01J2X8..."
  }
}
```

## 2. 资源命名与层级

```
/api/v1/workspaces/{slug}
/api/v1/workspaces/{slug}/projects/{identifier}
/api/v1/workspaces/{slug}/projects/{identifier}/issues[/{sequence_id}]
/api/v1/workspaces/{slug}/projects/{identifier}/issues/{sequence_id}/comments|relations|dependencies|time-logs|attachments
/api/v1/workspaces/{slug}/projects/{identifier}/sprints[/{id}]        (+ :start / :complete)
/api/v1/workspaces/{slug}/projects/{identifier}/versions[/{id}]       (+ :release / :archive)
/api/v1/workspaces/{slug}/projects/{identifier}/modules|labels|states|members|webhooks|automation-rules|dashboards
/api/v1/workspaces/{slug}/projects/{identifier}/intake/channels|issues (+ :convert)
/api/v1/workspaces/{slug}/kb/spaces[/{slug}/pages...]
/api/v1/search?q=...
/api/v1/me | /api/v1/me/notifications | /api/v1/me/workbench
```

规则：
- 全小写、复数名词、kebab-case 多词；`{slug}/{identifier}` 用可读标识而非数字 ID（对齐 Plane 风格）。
- 层级最深 4 级；更深的用顶层资源 + 查询参数（`/issues?parent=123` 而非无限嵌套）。
- 自定义动作用冒号语法（Google AIP-136）：`POST /sprints/12:complete`。

## 3. HTTP 语义

| 场景 | 方法/状态码 |
|------|------------|
| 查询 | GET → 200 |
| 创建 | POST → 201 + `Location` 头 |
| 全量更新 | PUT → 200（少用） |
| 部分更新 | PATCH → 200，JSON Merge Patch 语义 |
| 删除 | DELETE → 204（软删除） |
| 动作 | POST `:action` → 200 |
| 客户端参数错误 | 400；未认证 401；无权限 403；不存在 404；乐观锁冲突 409；限流 429 + `Retry-After`；校验失败 422 |
| 服务端错误 | 500；依赖不可用 503 |

**404 vs 403**：跨租户资源一律返回 404（不泄露存在性）。

## 4. 查询能力约定

```http
# 过滤：field 精确 / field__op 操作符
GET /issues?state__group=started&priority__in=urgent,high&assignee=me&target_date__lt=2026-09-01
# 操作符：in / ne / lt / lte / gt / gte / contains / isnull
# 排序：- 为降序，多字段逗号分隔
GET /issues?sort=-updated_at,sort_order
# 字段裁剪
GET /issues?fields=id,name,state,assignees
# 展开关联（白名单）
GET /issues/123?expand=state,assignees,labels
```

**分页**（双模式，默认 cursor）：
```http
GET /issues?per_page=50&cursor=eyJpZCI6OTF9
→ 200
{
  "results": [...],
  "next_cursor": "eyJpZCI6NDF9",     // null 表示结束
  "total_count": 1234                // 仅显式 ?include_total=true 时计算（大表 count 昂贵）
}
```
报表/管理类允许 `page/per_page` offset 分页，上限 `per_page ≤ 100`。

## 5. 写操作规范

- **幂等**：客户端对创建类请求生成 `Idempotency-Key: <uuid>`；服务端 24h 内重放直接返回首次响应（配合 `idempotency_keys` 表）。
- **乐观锁**：更新请求体带 `version`；不匹配返回 409 + 最新版本号，前端引导刷新合并。
- **批量**：`POST /issues:batch-update`，单请求 ≤100 条；响应逐条给出成功/失败明细；默认非原子，需原子时传 `"atomic": true`（单事务）。
- **部分失败语义**：`207 Multi-Status` 不用；统一 200 + `results[].status` 明细（前端处理更简单）。

## 6. 认证与限流

| 方式 | 场景 | 头 |
|------|------|----|
| Session Cookie（HttpOnly + SameSite=Lax） | Web SPA | 自动携带 + CSRF Token（`X-CSRF-Token`） |
| API Token | 脚本/集成 | `X-Api-Key: ydz_xxx`（仅存 hash，scopes 收敛） |
| OAuth2/OIDC | Phase 3 SSO | `Authorization: Bearer` |

**限流**（Redis 令牌桶，响应头携带 `X-RateLimit-Limit/Remaining/Reset`）：

| 端点 | 额度 |
|------|------|
| 默认 | 100 req/min/用户 |
| 登录/找回密码 | 10 req/min/IP |
| 全局搜索 | 30 req/min/用户 |
| Webhook 测试推送 | 10 req/hour/项目 |
| Intake 公开提交 | 20 req/min/IP + 人机校验 |

## 7. 错误码注册表（节选，全量见 `api/openapi/errors.md` 生成物）

| code | HTTP | 说明 |
|------|------|------|
| `AUTH.INVALID_CREDENTIALS` | 401 | 凭证错误（模糊化，不区分账号/密码） |
| `AUTH.TOKEN_EXPIRED` | 401 | access token 过期，走刷新 |
| `RBAC.FORBIDDEN` | 403 | 权限不足 |
| `TENANT.ARCHIVED` | 403 | 空间已归档只读 |
| `ISSUE.NOT_FOUND` | 404 | — |
| `ISSUE.VERSION_CONFLICT` | 409 | 乐观锁冲突 |
| `ISSUE.INVALID_STATE_TRANSITION` | 422 | 流转非法 |
| `ISSUE.WBS_DEPTH_EXCEEDED` | 422 | 超过三级 |
| `ISSUE.CIRCULAR_PARENT` | 422 | 循环父链 |
| `ISSUE.DUPLICATE_IN_ACTIVE_SPRINT` | 422 | 已在活跃迭代 |
| `VALIDATION.FAILED` | 422 | 通用字段校验（details 带 field） |
| `RATE_LIMIT.EXCEEDED` | 429 | — |

规则：`DOMAIN.SNAKE_CASE` 两段式；message 面向用户（中文），code 面向程序；新错误码在 PR 中登记注册表。

## 8. OpenAPI 治理

- swaggo 注解即事实源；`make openapi` 生成 `api/openapi/openapi.yaml`。
- CI 检查：①注解与代码一致（生成物 diff 为空）；②`oasdiff` breaking-change 检测，破坏性变更需 `BREAKING:` 提交标记 + 版本委员会（架构师+PM）批准。
- 前端类型 `openapi-typescript` 生成，CI 校验无漂移。
- Swagger UI 挂 `/api/docs`（生产可关）；每个端点必填 summary、权限点标注（`x-permission` 扩展字段）、错误码示例。

## 9. WebSocket 协议（实时通道）

```
wss://host/ws?workspace={slug}          （Session 鉴权）
服务端 → 客户端：{ "channel": "issue", "event": "issue.updated", "data": { "id": 1, ... }, "ts": ... }
客户端 → 服务端：{ "type": "subscribe", "channels": ["project:YD"] }
心跳：30s ping/pong；断线指数退避重连（1s→30s 上限）
```
前端收到事件后只做 **Query 失效**，不直接改本地数据（单一事实源在服务端）。
