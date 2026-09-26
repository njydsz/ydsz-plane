/**
 * 这是一个迁移样板文件。运行 `pnpm gen:types` 之后，
 * orval 会在 src/api/generated/api.ts 中自动生成完整的类型安全 API client。
 *
 * 步骤：
 *   1. 打开终端，确保当前目录在 web/
 *   2. 运行 `pnpm gen:types`（依赖 swagger.yaml 由 make swagger 生成）
 *   3. orval 自动扫描 swagger 注解并生成类型 + service 函数
 *   4. 逐个域迁移现有手工 service 文件到自动生成的版本
 *
 * 迁移优先级：
 *   1. workspace.service.ts → 依赖 workspace 域 API 服务
 *   2. project.service.ts → 依赖 project 域 API 服务
 *   3. issue.service.ts → 依赖 issue 域 API 服务（最复杂，留至最后）
 */

// =====================================================
// 迁移前（手工 service 示例）
// =====================================================
import { http } from "../client";
import type { CreateWorkspaceRequest } from "../../types/workspace";

export async function createWorkspace(req: CreateWorkspaceRequest) {
  const { data } = await http.post("/api/v1/workspaces", req);
  return data;
}

// =====================================================
// 迁移后（orval 生成的 service — 示意结构）
// =====================================================
// orval 会生成如下结构：
//   import { useCreateWorkspace } from '@/api/generated';
// 或：
//   import { createWorkspace } from '@/api/generated';
//
// 调用方直接引用，无需手动维护接口 URL 或类型：
//
//   // React/Vue 风格（取决于 orval 配置）
//   const { data, error } = await createWorkspace(createWorkspaceRequest);
//   if (error) throw error;
//   return data;
//
// orval 自动：
//   ✅ 类型安全的请求/响应
//   ✅ 自动使用 http client（mutator 已配置）
//   ✅ 基于 swagger 注解的 JSDoc
//   ✅ TypeScript 编译期检测 breaking change
// =====================================================

export {};
