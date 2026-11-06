# Ydsz Plane SDK

官方类型化 SDK，覆盖 Ydsz Plane 全部公开 API。

| 语言 | 包名 | 安装 |
|------|------|------|
| TypeScript / Node.js | `ydsz-plane` | `npm install ydsz-plane` |
| Python | `ydsz-plane` | `pip install ydsz-plane` |

## 快速上手

### TypeScript

```ts
import { PlaneClient, PlaneError } from 'ydsz-plane';

const plane = new PlaneClient({
  baseURL: 'https://your-host',
  token: 'ydz_xxx', // API Token；OAuth 用 oat_ 前缀
});

// 列出任务
const tasks = await plane.workItems.list({
  workspaceId: 1,
  projectId: 10,
  type: 'task',
  state: '进行中',
});

// 创建工作项
const created = await plane.workItems.create({
  workspaceId: 1,
  projectId: 10,
  type: 'task',
  title: '接入支付网关',
  priority: 'high',
});

// 迭代
const cycle = await plane.cycles.create({
  workspaceId: 1,
  projectId: 10,
  name: 'Sprint 21',
  goal: '完成支付模块',
});

// AI Agent 对话（Ask / Build / Autopilot）
const result = await plane.aiAgents.chat(1, 42, '本周有哪些高优先级缺陷？', 'ask');
console.log(result.assistant_message.content);
console.log('token 用量:', result.usage);

// 错误处理
try {
  await plane.workItems.get(1, 10, 'task', 999);
} catch (err) {
  if (err instanceof PlaneError) {
    console.error(err.status, err.code, err.message);
  }
}
```

### Python

```python
from ydsz_plane import PlaneClient, PlaneError

plane = PlaneClient(base_url="https://your-host", token="ydz_xxx")

# 列出任务
tasks = plane.work_items.list(10, workspace_id=1, type="task")

# 创建工作项
created = plane.work_items.create(1, 10, "task", "接入支付网关", priority="high")

# 迭代
cycle = plane.cycles.create(1, 10, "Sprint 21", goal="完成支付模块")

# AI Agent 对话
result = plane.ai_agents.chat(1, 42, "本周有哪些高优先级缺陷？", mode="ask")
print(result["assistant_message"]["content"])

# 错误处理
try:
    plane.work_items.get(1, 10, 999)
except PlaneError as exc:
    print(exc.status, exc.code, exc)
```

## 覆盖的资源

| 模块 | TypeScript | Python |
|------|-----------|--------|
| 工作空间 | `workspaces` | `workspaces` |
| 项目 | `projects` | `projects` |
| 工作项 | `workItems` | `work_items` |
| 迭代 | `cycles` | `cycles` |
| 模块 | `modules` | `modules` |
| 版本 | `versions` | `versions` |
| 文档 | `pages` | `pages` |
| 成员 | `members` | `members` |
| 搜索 | `search` | `search` |
| AI Agent | `aiAgents` | `ai_agents` |
| 视图 | `views` | `views` |
| 倡议 | `initiatives` | `initiatives` |
| 团队空间 | `teamspaces` | `teamspaces` |
| 客户 | `customers` | `customers` |

## 认证

支持两种 Token：

- **API Token**（`ydz_` 前缀）：在「设置 → API Token」中创建，长期有效，可按 scope 授权。
- **OAuth Access Token**（`oat_` 前缀）：通过 OAuth2 授权码流程获取，代表用户身份，可设过期时间。

## MCP 接入

AI 客户端（Claude / Cursor / ChatGPT / Codex 等）可直接通过 MCP 协议接入，无需 SDK：

```json
{
  "mcpServers": {
    "ydsz-plane": {
      "url": "https://your-host/mcp",
      "headers": { "Authorization": "Bearer ydz_xxx" }
    }
  }
}
```

## 许可证

MIT
