# S14 整体验证清单

> 本文档用于 S14 实施完成后的整体验证，确保桌面客户端与微服务拆分两大能力达到交付标准。
> 执行时机：每个 Phase 完成后逐阶段验证 + S14 最终验证。

---

## 1. 桌面客户端验证（P0 - P1 - P3）

### 1.1 构建验证

| 验证项 | 命令/方法 | 期望结果 | 状态 |
|--------|----------|----------|------|
| Windows 构建 | `wails build -platform windows/amd64` | 产出 `Ydsz Plane.exe`，无编译错误 | ⬜ |
| macOS 构建 | `wails build -platform darwin/universal` | 产出 `Ydsz Plane.app`（Universal Binary） | ⬜ |
| Linux 构建 | `wails build -platform linux/amd64` | 产出 AppImage / 可执行文件 | ⬜ |
| 空包体积 | 可执行文件 ≤ 30MB | `du -h Ydsz Plane.exe` ≤ 30MB | ⬜ |

### 1.2 功能验证（桌面端）

| 场景 | 步骤 | 期望 | 状态 |
|------|------|------|------|
| SSO 认证 | 点击登录按钮 → 系统浏览器 SSO → 回调 → 应用收到 Token | 会话持久化到 OS Keychain，刷新后仍在 | ⬜ |
| Token 刷新 | 等待 Token 过期前 5min | Go 端自动 refresh，前端无感知 | ⬜ |
| 会话持久化 | 关闭应用 → 重新启动 | 自动从 Keychain 读取 Token，免登录 | ⬜ |
| 工作空间选择 | 启动应用 | 列出用户的所有工作空间 | ⬜ |
| 项目看板 | 进入项目 → 看板视图 | WIP 限制正常，可拖动卡片 | ⬜ |
| 工作项详情 | 点击工作项 | 显示状态/评论/活动，可切换状态 | ⬜ |
| 系统托盘 | 点击 × 关闭窗口 | 应用最小化到托盘（不退出） | ⬜ |
| 托盘角标 | 收到新通知 | 托盘图标显示未读计数角标 | ⬜ |
| 全局快捷键 | 应用未聚焦时按 Ctrl+Shift+Y | 应用窗口唤起/隐藏 | ⬜ |
| 原生通知 | 收到新通知时桌面右下角 Toast | Windows Toast / macOS Notification | ⬜ |
| 离线缓存 | 断开网络 → 浏览上次访问的项目 | 仍可查看缓存数据 | ⬜ |
| 离线写草稿 | 断开网络 → 创建工作项草稿 | 草稿存入本地 SQLite | ⬜ |
| 恢复在线 | 重连网络 | 未同步草稿自动提交 | ⬜ |
| 自动更新 | 服务端存在新版本 → 应用启动 | 后台下载更新 → 提醒用户重启 | ⬜ |

### 1.3 性能基线（桌面端）

| 指标 | 基线 | 状态 |
|------|------|------|
| 冷启动时间 | ≤ 1 秒 | ⬜ |
| 空闲内存占用 | ≤ 150 MB | ⬜ |
| 通知列表加载（首次） | ≤ 500 ms | ⬜ |
| 通知列表加载（缓存） | ≤ 100 ms | ⬜ |
| 工作项详情打开 | ≤ 300 ms | ⬜ |
| SSO 端到端耗时 | ≤ 5 秒（含浏览器交互） | ⬜ |

### 1.4 兼容性验证

| 平台 | 版本 | 状态 |
|------|------|------|
| Windows 11 | 23H2+ (WebView2 自带) | ⬜ |
| Windows 10 | 1903+ (需 WebView2 Runtime) | ⬜ |
| macOS | 13+ (Ventura) | ⬜ |
| Ubuntu | 22.04 LTS (webkit2gtk-4.0) | ⬜ |
| 统信 UOS | 20 (内置 GNOME WebView) | ⬜ |
| 麒麟 Kylin | V10 (CEF 兼容层) | ⬜ |

---

## 2. 微服务拆分验证（Phase-0 → Phase-1 → Phase-2）

### 2.1 Phase-0 验证（契约先行）

| 验证项 | 方法 | 期望 | 状态 |
|--------|------|------|------|
| Proto 编译 | `protoc` / `buf generate` | 无语法错误，Go/Java 代码生成成功 | ⬜ |
| gRPC 服务端启动 | 在 core 进程内注册 gRPC service | 进程内 HTTP Handler → gRPC Service 等价调用 | ⬜ |
| 功能等价性 | 对比 REST / gRPC 两种路径的响应 | 字段零差异 | ⬜ |
| 全量单测通过 | `go test ./...` | 通过率 100% | ⬜ |
| 集成测试通过 | E2E 冒烟（核心用户链路） | 通过率 100% | ⬜ |
| 性能基线回归 | K6 负载测试（1000 并发） | P95 延迟 ≤ ±5% | ⬜ |

### 2.2 Phase-1 验证（通知服务独立）

| 验证项 | 方法 | 期望 | 状态 |
|--------|------|------|------|
| 容器构建 | `docker build -f build/Dockerfile.notification` | 镜像构建成功 | ⬜ |
| 独立启动 | `docker compose -f deployments/docker-compose.s14-notification.yml up` | 服务在 :9090 / :8080 正常启动 | ⬜ |
| Schema 迁移 | `psql -f sql/notification_schema_v1.sql` | 表/索引/约束创建成功 | ⬜ |
| gRPC 健康检查 | `curl http://localhost:8080/ready` | `{"ready":true}` | ⬜ |
| gRPC 连通性 | core-service 连接 notification-svc:9090 | 通知 CRUD 功能等价 | ⬜ |
| RabbitMQ 消费 | 触发领域事件（e.g. issue.created） | notification-svc 消费并写入通知 | ⬜ |
| 灰度切流 | feature flag 控制 0%→100% | 流量平滑切换，无 5xx 飙升 | ⬜ |
| 独立扩缩 | `docker compose up -d --scale notification-svc=3` | 3 实例均分流量 | ⬜ |
| 故障隔离 | 模拟 notification-svc down | core-service CRUD 仍可用（通知失败降级为静默） | ⬜ |
| 滚动更新 | `docker compose up -d --build --force-recreate` | 零停机，gRPC draining | ⬜ |

### 2.3 Phase-2 验证（搜索服务独立）

| 验证项 | 方法 | 期望 | 状态 |
|--------|------|------|------|
| 容器构建 | `docker build -f build/Dockerfile.search` | 镜像构建成功 | ⬜ |
| 独立启动 | 部署 search-svc | 服务在 :9091 / :8081 启动 | ⬜ |
| gRPC 健康检查 | `curl http://localhost:8081/ready` | `{"ready":true}` | ⬜ |
| 搜索等价性 | 对比单体搜索 vs 微服务搜索 | 结果集一致（total + 排序等价） | ⬜ |
| ES 索引同步 | 触发 issue.updated 事件 | ES 文档最终一致（≤ 5s） | ⬜ |
| ES 降级路径 | 模拟 ES 不可用 | 自动降级到 PG FTS，response.header X-Search-Backend: pg | ⬜ |
| 全量重建 | 运维触发 Reindex All | 重建完成，搜索数据完整 | ⬜ |

---

## 3. 集群运维验证

### 3.1 监控与告警

| 验证项 | 工具 | 期望 | 状态 |
|--------|------|------|------|
| gRPC 请求成功率 | Prometheus | notification-svc / search-svc success_rate ≥ 99.9% | ⬜ |
| gRPC P95 延迟 | Prometheus Histogram | P95 ≤ 50ms | ⬜ |
| 容器重启次数 | Docker / K8s events | 0 次意外重启 | ⬜ |
| JVM / Go 进程内存 | runtime metrics | 容器 limits 80% 以下 | ⬜ |
| 死信队列（DLQ） | RabbitMQ Queue size | DLQ 堆积 = 0（或稳定趋势） | ⬜ |
| PG WAL 延迟 | replication lag | notification_db ↔ core_db 同步延迟 ≤ 1s | ⬜ |

### 3.2 故障演练（Chaos）

| 场景 | 注入方式 | 期望系统行为 | 状态 |
|------|----------|------------|------|
| notification-svc OOM | 压力测试 | core-service 核心链路不受影响，通知暂时静默 | ⬜ |
| notification_db 宕机 | 停掉容器 | notification-svc 进入重试退避，恢复后自愈 | ⬜ |
| ES 集群不可用 | 停掉 ES | search-svc 自动降级到 PG FTS，搜索不限于 | ⬜ |
| 网络分区（core ↔ notification-svc） | iptables 模拟 | 超时 + circuit breaker + fallback | ⬜ |
| PG 连接池耗尽 | 调低 max_connections | 服务拒绝新连接但现有请求正常，恢复后自愈 | ⬜ |

---

## 4. API 兼容性验证

| 测试 | 方法 | 期望 | 状态 |
|------|------|------|------|
| 现有 REST API 行为 | 对比 S13 基线 Postman collection | 100% 响应字段一致 | ⬜ |
| 现有 WebSocket 实时事件 | 断线重连测试 | 事件顺序不变，补偿机制正常 | ⬜ |
| 现有 Open API 文档 | 比对 /swagger/doc.json | 字段 / 路径无破坏性变化 | ⬜ |
| SDK 客户端（如有） | TypeScript SDK 编译 / 测试 | 现有 TS 客户端不调后端拆分 | ⬜ |

---

## 5. 安全验证

| 验证项 | 方法 | 期望 | 状态 |
|--------|------|------|------|
| gRPC mTLS（可选） | 检查证书双向认证 | 服务间通信加密 | ⬜ |
| 通知服务 RLS | PSQL 直连测试 | 用户无法读写其他 workspace 通知 | ⬜ |
| 桌面端 Token 存储 | 检查 OS Keychain 权限 | DPAPI 绑定用户，其他用户无法读取 | ⬜ |
| 桌面端 SSO 回调 | state 校验 | 篡改 state 后回调被拒绝 | ⬜ |
| Proto 字段敏感度 | 检查通知 body 是否泄露密码/Token | 通知内容不含敏感字段 | ⬜ |

---

## 6. 最终放行条件（Exit Criteria）

**整体通过标准**：
- [ ] 桌面客户端（Wails）三平台构建通过，MVP 功能可用
- [ ] 通知服务（notification-svc）独立部署，功能等价单体
- [ ] 搜索服务（search-svc）独立部署，功能等价单体
- [ ] 全量 API 兼容性回归通过（REST + WebSocket 行为无退化）
- [ ] 性能基线回归：1000 并发 P95 ≤ 200ms（单体入口端）
- [ ] 七项混沌演练通过
- [ ] 监控告警覆盖所有新组件
- [ ] 部署文档 + Runbook + CHANGELOG 更新
