# 实时协同编辑 — ROI 分析与架构方案评估

> 评估日期：2026-08-08  
> 适用项目：ydsz-plane（研发项目管理平台）  
> 版本：v1.0

---

## 1. 背景与范围

实时协同编辑（Realtime Collaboration）允许多人同时编辑同一文档/字段，常见于：

- **TipTap 文档协同**：多人同时编辑富文本（如工作项描述、Page 文档）
- **看板拖拽同步**：多人操作同一看板时实时看到卡片移动
- **列表内联编辑**：如多人同时更新同一工作项的状态/优先级

本节聚焦**文档类协同编辑**（场景 1），因其技术复杂度最高，也最能体现协同价值。

---

## 2. ROI 分析

### 2.1 收益（Benefits）

| 维度 | 预期收益 | 价值等级 |
|------|---------|---------|
| **协作效率** | 多人并行编辑文档，无需「锁定—等待—合并」 | 高 |
| **用户感知** | 竞品标准功能（Notion、Linear、Plane 均有），缺失将影响产品竞争力 | 中 |
| **减少冲突** | 视觉提示 + 自动合并，避免「覆盖编辑」导致内容丢失 | 中 |
| **团队远程化** | 支撑跨时区协作，编辑时无需频繁同步进度 | 高 |

### 2.2 成本（Costs）

| 维度 | 成本项 | 估算规模 |
|------|--------|---------|
| **服务端** | 运维协同服务（WebSocket 长连接 + 状态同步 + 持久化） | 需新增 1 个微服务，月均基础设施成本 ~¥500-2000 |
| **客户端** | TipTap + Yjs 集成，协同光标/选区渲染 | ~80-120 人时 |
| **后端** | CRDT/OT 服务端、权限校验、房间管理、版本快照 | ~60-100 人时 |
| **测试覆盖** | 冲突场景、断线重连、大数据文档性能 | ~30-50 人时 |
| **持续运维** | 长连接维护、TTTL 快照、断线补偿逻辑 | 每月 ~10-15 人时 |

### 2.3 适用用户规模判断

- **强需求**：团队规模 ≥ 5 人、频繁共同编辑同一文档
- **已有凭据**：Plane 默认全部文档启用协同，协同是其核心卖点之一
- **本项目定位**：对标 Plane，若缺失此功能将形成**产品能力短板**，但非短期核心能力

### 2.4 结论

> **当前阶段（MVP → 1.0）：不建议自建协同编辑服务。**  
> 建议路径：先保证「单人编辑体验 + 实时只读预览」（通过 WebSocket 推送他人状态），待用户规模稳定后（日活用户 > 500、并发编辑场景明确），再评估引入 Yjs + Tiptap Collab 方案。

---

## 3. 技术方案对比

### 3.1 OT（Operational Transformation）

| 项 | 说明 |
|-----|------|
| 代表实现 | ShareDB、Google Docs 早期方案 |
| 核心思路 | 服务端转换操作顺序，保证最终一致 |
| 优势 | 经过成熟产品验证，生态丰富 |
| 劣势 | 服务端单点瓶颈、断线重连逻辑复杂、服务端需理解文档模型 |
| 适用场景 | 服务端主导、文档结构复杂 |
| 集成复杂度 | 高（需实现 OT 算法） |

### 3.2 CRDT（Conflict-free Replicated Data Types）

| 项 | 说明 |
|-----|------|
| 代表实现 | Yjs、Automerge、Diamond Types |
| 核心思路 | 客户端本地 merge，无需服务端参与冲突解决 |
| 优势 | 服务端无状态、离线编辑天然支持、断线重连简单 |
| 劣势 | 内存占用较高、同步带宽较大、大型文档首次加载慢 |
| 适用场景 | 客户端优先、去中心化、离线优先 |
| 集成复杂度 | 中（封装良好，绑定 Y.Text 到 TipTap） |

### 3.3 推荐：Yjs + TipTap Collab

```text
                 ┌───────────────────────────┐
                 │      TipTap Editor        │
                 │  (ProseMirror + Y.Tiptap) │
                 └─────────────┬─────────────┘
                               │ Y.Doc (本地 CRDT)
                               ▼
                 ┌───────────────────────────┐
                 │      Yjs Provider         │
                 │  (y-websocket / y-webrtc) │
                 └─────────────┬─────────────┘
                               │ WebSocket / WebRTC
                               ▼
                 ┌───────────────────────────┐
                 y-websocket Sync Server     │
                 (Node.js / y-redis)         │
                 └─────────────┬─────────────┘
                               │ 持久化
                               ▼
                          PostgreSQL
                        (Y.encodeStateAsUpdate)
```

**为什么推荐 Yjs？**

1. **与 TipTap 集成成熟**：`@tiptap/extension-collaboration` 原生支持 Yjs
2. **服务端轻量**：y-websocket 服务只做消息转发，冲突由客户端自行解决
3. **支持离线**：本地变更优先，恢复后自动合并
4. **生态完善**：y-indexeddb 本地缓存、y-redis 服务端持久化、y-webrtc P2P 方案

---

## 4. 架构设计方案（跟进阶段）

### 4.1 协同 Token 模型

```typescript
interface CollabRoom {
  /** roomId: `${entityType}:${entityId}` */
  id: string;
  /** 关联实体 */
  entityType: "issue" | "page" | "sprint_goal";
  entityId: number;
  /** 活跃用户列表 */
  participants: CollabUser[];
  /** 权限：owner 可写，viewer 只读 */
  access: "write" | "read";
}

interface CollabUser {
  userId: number;
  displayName: string;
  /** 光标颜色（8 色循环） */
  color: string;
  /** 选区 / 光标位置 */
  cursor: { anchor?: number; head?: number } | null;
  /** 最后活跃时间 */
  lastSeen: number;
}
```

### 4.2 数据流

1. 用户打开文档 → GET `/api/collab/token?entity=issue:123`
2. 后端校验权限 + 签发 WebRTC/WS 连接凭证
3. 前端通过 y-websocket 连接 room
4. Yjs 同步文档状态 + Awareness Protocol 传递光标/选区信息
5. 编辑完成定时/失焦后 `Y.encodeStateAsUpdate` 写入数据库（通过 REST API 备份）

### 4.3 安全防护

| 措施 | 实现 |
|------|------|
| 读权限 | JWT + room_id 校验，仅 room 成员可连接 |
| 写权限 | TipTap 仅对 hasWritePermission=true 用户启用编辑，服务端忽略未授权 update |
| 单次 update 大小 | gateway 层限制 < 1MB，防止 DoS |
| 文档最大长度 | ProseMirror 配置 `maxSize`（建议 100KB） |
| 版本快照 | 每天凌晨定时快照，支持历史版本恢复 |

### 4.4 部署建议（跟进实施）

```
  y-websocket-server [Docker]
         │
         ├── (可选) Redis 作为横向扩展的消息总线
         └── (可选) PostgreSQL 持久化 Y.Doc 状态

  y-websocket 建议部署为独立 Pod：
    - 副本数 ≥ 2（高可用）
    - 单节点支持 ~5000 并发连接
    - 使用 sticky session（一致性 hash userId）
```

---

## 5. 里程碑与推荐实施路径

| 阶段 | 目标 | 验收标准 |
|------|------|---------|
| **Phase A：只读实时预览** | WebSocket 推送「谁在查看此文档」 | 多人打开同一文档时能看到彼此头像 |
| **Phase B：简易协同编辑** | 引入 y-websocket，支持文本协同 | 多人同时编辑同一文档，实时合并 |
| **Phase C：光标 + 选区** | Awareness Protocol 同步光标 | 看到他人光标位置 + 用户名 |
| **Phase D：结构化协同** | 表格、callout、任务列表等 Node 协同 | 协同编辑富文本 Node 不冲突 |
| **Phase E：版本快照 + 历史** | 编辑历史 + 回滚 | 「版本历史」Tab 可切换查看历史版本 |

**建议从 Phase A 开始**，投入 < 5 天，即可支撑 PM 推动协同需求；后续阶段按用户反馈决定优先级。

---

## 6. 参考竞品对标

| 产品 | 协同方案 | 启用范围 |
|------|---------|---------|
| Notion | 自研 CRDT | 全文档 |
| Google Docs | OT (Journal) | 全文档 |
| Linear | 自建协同（用户状态为主） | 问题详情 |
| Plane (开源) | 暂未启用强协同（仅 Presence） | — |
| Confluence | 自研协同（如同 Google） | 全文档 |
| 飞书文档 | 自研 CRDT | 全文档 |

---

## 7. 结论与行动建议

**短期（1-2 个月）**：不投入完整的实时协同编辑实现；专注打磨单人编辑体验（表格、callout、slash command 等已落地）。

**中期（3-6 个月）**：实现 Phase A（只读 Presence）作为「协同预览」，积累协同场景数据。

**长期（6 个月+）**：基于 Phase A 的数据（协同频率、冲突频次、用户反馈），决定是否推进到 Phase B/C。若决定推进，优先复用 y-websocket 开源方案 + TipTap 官方 Collab 扩展，避免自研 OT 算法。

---

## 附录 A：关键依赖版本参考

```json
{
  "yjs": "^13.6.0",
  "y-websocket": "^2.0.0",
  "y-indexeddb": "^9.0.0",
  "@tiptap/extension-collaboration": "^2.6.0",
  "@tiptap/extension-collaboration-cursor": "^2.6.0",
  "lib0": "^0.2.90"
}
```

## 附录 B：Yjs + TipTap 最小接入示例（伪代码）

```typescript
// provider.ts
import * as Y from "yjs";
import { WebsocketProvider } from "y-websocket";
import { IndexeddbPersistence } from "y-indexeddb";

const ydoc = new Y.Doc();
const wsProvider = new WebsocketProvider(
  "wss://collab.example.com",
  `issue-${issueId}`,
  ydoc,
);
const indexeddbProvider = new IndexeddbPersistence(`issue-${issueId}`, ydoc);

export const yXmlFragment = ydoc.getXmlFragment("prosemirror");
```

```typescript
// editor-setup.ts
import Collaboration from "@tiptap/extension-collaboration";

const editor = useEditor({
  extensions: [
    // ... 其他 extensions
    Collaboration.configure({
      document: ydoc,
    }),
    // 光标协同（Phase C 开启）
    // CollaborationCursor.configure({
    //   provider: wsProvider,
    //   user: { name: currentUser.name, color: userColor },
    // }),
  ],
});
```
