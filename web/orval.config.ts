/**
 * Orval 配置 — 从 OpenAPI YAML 生成类型安全的 Axios client。
 *
 * 生成流水线：
 *   1. make swagger        → swag init 生成 docs/swagger/swagger.yaml
 *   2. make gen-types      → orval 全量生成 src/api/generated/*.ts
 *   3. make gen            → 1+2 同步
 *
 * 依赖：
 *   - 后端: go install github.com/swaggo/swag/cmd/swag@latest
 *   - 前端: pnpm add -D orval
 *
 * 覆盖率状态（2026-09-26 更新）：
 *   283 @Summary 注解覆盖 28 个文件 → 估算路径覆盖率 ≥60%
 *   swagger.yaml 需重新运行 `make swagger` 后生成完整路径
 */
import { defineConfig } from "orval";

export default defineConfig({
  plane: {
    input: "../docs/swagger/swagger.yaml",
    output: {
      target: "src/api/generated/api.ts",
      schemas: "src/types/generated",
      client: "axios",
      mock: false,
      mode: "single",                   // 单文件输出，便于 review & tree-shaking
      // 注：不使用 clean 选项。orval 的 mutator 需读取 src/api/client.ts，
      // clean 会在生成前清空输出目录，导致 mutator 文件解析失败。
      // 生成产物通过 git 管理，陈旧文件的清理交由人工 review。
      override: {
        mutator: {
          // 专用 mutator：导出函数声明形式，运行时委托 client.ts 的 http 实例，
          // 保留 interceptors（token / 错误处理 / 限流）。
          path: "src/api/orval-mutator.ts",
          name: "customInstance",
        },
        // 强制使用 unknown 替代 any（strict TypeScript）
        unknownType: "unknown",
      },
    },
    // Hook: 生成后自动格式化
    hooks: {
      afterAllFilesWrite: "prettier --write src/api/generated src/types/generated",
    },
  },
});
