# Ydsz Plane 项目记忆

## 项目概况
- 开源项目管理平台（类 Jira/云效/ONES），Go + Vue 3 + PostgreSQL + Redis + RabbitMQ
- 仓库: github.com/njydsz/ydsz-plane, MIT 许可
- S1-S12 全部完工，v0.2 已发布

## Phase 3+ 实施记录 (2026-08-07/08)

### P1: ES 全局搜索升级 ✅
- `pkg/searchql/parser.go` — 类 JQL 语法解析器（递归下降，~400行）
- `internal/infrastructure/es/client.go` — ES 8.x 客户端（连接池/别名切换/批量索引）
- `internal/application/search/es_backend.go` — ES 搜索后端（JQL→ES bool query）
- `sql/0025_es_search.up.sql` — ES 同步状态追踪表
- 更新 `service.go` 支持双后端（ES + PG FTS）+ 降级策略

### P2: SSO/OIDC 企业认证 ✅
- `internal/application/auth/oidc.go` — OIDC 完整流程（PKCE/state/nonce/JWKS验证）
- `sql/0026_sso_oidc.up.sql` — SSO 表（providers/links/sessions）
- users 表扩展 sso_provider/sso_subject

### P3: 国际化 i18n ✅
- `web/src/locales/zh-CN.ts` + `en-US.ts` — 中英文完整翻译（~500 key）
- `web/src/locales/index.ts` — i18n 配置（vue-i18n v10, 组合式API）
- `web/src/components/LocaleSwitcher.vue` — 语言切换组件
- `web/src/composables/useI18n.ts` — 便捷 composable

### P4: 电子表格视图 ✅
- `web/src/views/project/SpreadsheetView.vue` — 类 Airtable 表格（行内编辑/列拖拽/键盘导航）
- 路由已注册 `/projects/:id/spreadsheet`

### P5: PWA 支持 ✅
- `web/vite.config.ts` — vite-plugin-pwa 配置（Workbox/缓存策略/Manifest）
- `web/src/pwa.ts` — SW 注册/更新提示/Web Push 接口

### P6: 数据迁移工具 ✅
- `internal/application/migrate/importer.go` — 导入器框架（Jira CSV/通用 CSV）
- 支持 Jira→Ydsz 类型/优先级/状态映射

### P7: 信创适配层 ✅
- `pkg/crypto/sm.go` — 国密算法接口（SM2/SM3/SM4）+ Build Tag 隔离
- `internal/infrastructure/persistence/dialect.go` — 数据库方言（PostgreSQL/达梦/金仓）

### P8: AI 智能功能 ✅
- `internal/application/ai/service.go` — 智能指派/重复检测/摘要/分类
- LLM Provider 接口抽象，规则引擎兜底
