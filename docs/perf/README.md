# 性能基线报告

> 性能测试结果归档目录，每次里程碑发布前更新。
> 压测工具：k6（Grafana），脚本位于 `scripts/k6/`。

## 基线记录

| 日期 | 场景 | 并发 VU | 数据规模 | API P95 | 错误率 | 全局 P95 | 结论 | 报告 |
|------|------|---------|----------|---------|--------|----------|------|------|
| 待首测 | smoke | 100 | 种子数据 | - | - | - | 待运行 | - |

## 压测场景说明

### smoke — 冒烟 + 基线压测

- **阶段**: 预热(10 VU/30s) → 爬坡(50 VU/1m) → 峰值(100 VU/2m) → 冷却(0 VU/30s)
- **覆盖**: Health / Me / Issue 列表 / Issue 创建+详情 / Sprint 列表 / Version 列表 / States
- **阈值**: API P95 < 200ms, P99 < 500ms, 错误率 < 1%, 全局 P95 < 500ms

### 运行方式

```bash
# 本地冒烟
k6 run scripts/k6/smoke.js

# 指定环境 + 并发
k6 run scripts/k6/smoke.js -e API_URL=https://staging.example.com/api/v1 -e VUS=1000

# 输出 JSON 报告
k6 run scripts/k6/smoke.js --out json=perf-result.json
```

## 回归流程

1. 每次性能敏感改动（SQL、索引、缓存、中间件）后运行 smoke 压测
2. 对比当前基线：若 P95 回退 > 10%，触发 Review 定位原因
3. 关键查询需 `EXPLAIN ANALYZE` 验证索引命中
4. S7 和 S12 各执行一次全量压测，报告归档至本目录

## 性能目标

| 指标 | 目标值 |
|------|--------|
| API 响应 P95 | ≤ 200ms |
| API 响应 P99 | ≤ 500ms |
| 错误率 | < 1% |
| 全局 HTTP P95 | ≤ 500ms |
| 并发用户数 | ≥ 1000 |
| 单项目工作项 | ≥ 100 万 |
