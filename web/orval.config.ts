/**
 * Orval 配置 — 从 OpenAPI YAML 生成类型安全的 Axios client。
 *
 * 安装: pnpm add -D orval
 * 使用: pnpm orval
 *
 * 当前状态：由于后端 swagger 注解覆盖率 <10%（仅 7 个 path），
 * orval 全量生成暂不可用。在注解覆盖率 ≥80% 前，使用 Makefile gen-types 的
 * openapi-typescript fallback 仅生成 TS interface（不生成 service）。
 *
 * 注解覆盖按域分批推进：domain/agile → domain/auth → domain/workflow → ...
 * 每域一个 PR，完成后覆盖率报告自动上升。
 */
import { defineConfig } from "orval";

export default defineConfig({
  // 当 swagger.yaml 路径数 ≥50 时激活完整 orval client 生成
  // plane: {
  //   input: "../docs/swagger/swagger.yaml",
  //   output: {
  //     target: "src/api/generated/issues.ts",
  //     schemas: "src/types/generated",
  //     client: "axios",
  //     mock: true,
  //     override: {
  //       mutator: {
  //         path: "../src/api/client.ts",
  //         name: "http",
  //       },
  //     },
  //   },
  // },
  //
  // 临时占位 — 实际配置在注解覆盖率 ≥80% 后启用
  placeholder: {
    input: "../docs/swagger/swagger.yaml",
    output: {
      target: "src/types/api.generated.ts",
      client: "axios-functions",
    },
  },
});
