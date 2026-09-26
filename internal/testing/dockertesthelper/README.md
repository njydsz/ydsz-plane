# dockertesthelper

基于 testcontainers-go 的集成测试容器管理。

## 依赖状态

`github.com/testcontainers/testcontainers-go` 尚未安装于 go.mod 中。

本目录下的所有 `*_test.go` 和 `dockertest.go` 文件均带有 `//go:build dockertest` 构建标签，
默认构建不引用 testcontainers，不影响已有 CI 流程。

## 安装依赖

在所有 `-tags dockertest` 的测试可用之前，需要：

```bash
go get github.com/testcontainers/testcontainers-go@v0.33.0
go get github.com/testcontainers/testcontainers-go/modules/postgres@v0.33.0
go mod tidy
```

版本选择：
- v0.33.0 支持 Go 1.21+（与本项目的 go 1.26.5 兼容）
- 生产 Docker 环境需要支持 `POSTGRES_HOST_AUTH_METHOD=trust` 或直接使用 WithPassword 设置密码

## 使用方式

### 并行容器隔离（每个子测试独占）

```go
func TestMyFeatureParallel(t *testing.T) {
    t.Parallel()
    pool, _ := dockertesthelper.ParallelPool(ctx, t)
    // ...
}
```

### 共享容器 + 事务回滚（高性能，推荐）

```go
func TestMyFeatureInTx(t *testing.T) {
    pool, _ := dockertesthelper.NewPool(ctx, t)
    fx, _ := txfixture.NewTxFixture(ctx, pool, t)
    tx := fx.PgTx()
    // ... 测试结束时 fx 的 t.Cleanup 自动回滚
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DOCKERTEST_POSTGRES_IMAGE` | `postgres:16-alpine` | 自定义镜像 |
| `DOCKERTEST_REUSE` | `false` | `true` 时复用已存在的同名容器 |

## 运行测试

```bash
# 运行 smoke test 验证容器启动链路
go test -tags dockertest -run TestSmoke_PostgresContainer ./internal/testing/dockertesthelper/...

# 并行子测试的事务隔离验证
go test -tags dockertest -run TestSmoke_ParallelContainersWithTxRollback ./internal/testing/dockertesthelper/...

# 完整集成测试（需要 Docker daemon）
go test -tags dockertest -v ./internal/testing/...
```
