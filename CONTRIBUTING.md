# 贡献指南

感谢你对 Ydsz Plane 的关注！我们欢迎所有形式的贡献——bug 报告、功能建议和代码提交。

---

## 起步

### 环境要求

- Go 1.26+
- Node.js 22+ / pnpm 10
- Docker + Docker Compose（基础设施）
- Elasticsearch 8（可选，搜索功能需要）

### 本地启动

```shell
# 1. 克隆后启动基础设施
make up

# 2. 数据库迁移 + 种子数据
make migrate && make seed

# 3. 后端（终端 1）
make dev-api

# 4. 前端（终端 2）
make dev-web
```

访问 http://localhost:5173，使用种子账号 `admin@njydsz.com` / `Admin@1020` 登录。

详细配置见 [.env.example](./.env.example)。

---

## 开发流程

1. **Fork 仓库**，在 `feature/xxx` 或 `fix/xxx` 分支开发
2. 提交前确保：
   - `make lint` — 静态检查 0 error
   - `make test` — 全部测试通过（含 race detector）
   - `make coverage` — 查看覆盖率变化
3. PR 描述清楚**变更动机、影响范围与测试结果**
4. 至少 1 人 Review 通过后由维护者 squash merge 到 main

---

## 测试

```shell
# 全量单元测试（race + 覆盖率）
make test
make coverage

# 集成测试（需要真实 PostgreSQL）
export YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test"
make migrate
go test -tags=integration ./internal/application/...

# API 冒烟（启动后校验健康端点）
make smoke
```

集成测试使用 `//go:build integration` tag，未设置数据库 URL 时自动跳过。

---

## 提交规范

建议遵循 [Conventional Commits](https://www.conventionalcommits.org/) 风格：

```
feat(issue): 支持缺陷与任务跨类型联动
fix(auth): 修正 OIDC 回调 state 丢失
docs: 更新 API 设计规范
test(automation): 补充 DSL 边界校验用例
refactor(state): 提取状态流转校验为纯函数
```

简短说明"做了什么 + 为什么"，维护者会在发布时汇总 CHANGELOG。

---

## 分支策略

- `main`：稳定可发布分支，受保护
- `feature/*`、`fix/*`、`docs/*`、`test/*`：短期特性分支

---

## 编码规约

- **Go**：遵循 [Uber Go Style Guide](https://github.com/uber-go/guide)，`golangci-lint` 静态分析
- **前端**：ESLint + Prettier + vue-tsc 强类型检查
- **数据库**：DDL 变更提交至 `sql/`，使用 `make migrate` 验证

---

## 行为准则

对所有贡献者保持尊重与友善。垃圾信息、骚扰或歧视性言行将被移除。

---

## 许可

提交 PR 即表示你同意你的代码基于 [MIT License](./LICENSE) 开源发布。
