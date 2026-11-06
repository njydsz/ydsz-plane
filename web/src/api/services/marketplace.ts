/**
 * 应用商店（Marketplace）域 API — 对接后端应用目录 REST 接口。
 *
 * 路由前缀: /api/v1/workspaces/:workspace_id/marketplace
 *
 * 能力：
 *   - 分类目录 / 应用目录（按分类、关键字过滤，分页）；
 *   - 应用详情（按 slug）；
 *   - 安装 / 卸载 / 已安装列表；
 *   - 评分与评论（列表 / 提交）；
 *   - 发布 / 更新应用（实例管理员或官方）。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 应用分类 slug */
export type MarketplaceCategory = "apps" | "importers" | "templates" | "integrations";

/** 定价模式 */
export type MarketplacePricing = "free" | "paid" | "freemium";

/** 安装方式 */
export type MarketplaceInstallType = "builtin" | "oauth" | "webhook" | "cli";

/** 发布状态 */
export type MarketplaceStatus = "published" | "draft" | "deprecated";

/** 分类信息 */
export interface CategoryInfo {
  slug: string;
  name: string;
  description?: string;
  icon?: string;
  sort_order: number;
}

/** 应用目录条目 */
export interface MarketplaceApp {
  id: number;
  slug: string;
  name: string;
  description: string;
  category: MarketplaceCategory;
  publisher: string;
  publisher_url?: string;
  icon_url?: string;
  screenshots?: string[];
  tags?: string[];
  pricing: MarketplacePricing;
  install_type: MarketplaceInstallType;
  install_config?: Record<string, unknown>;
  docs_url?: string;
  rating_avg: number;
  rating_count: number;
  install_count: number;
  status: MarketplaceStatus;
  created_by?: number;
  created_at: string;
  updated_at: string;
}

/** 安装记录 */
export interface MarketplaceInstall {
  id: number;
  workspace_id: number;
  app_id: number;
  installed_by: number;
  config?: Record<string, unknown>;
  enabled: boolean;
  installed_at: string;
  updated_at: string;
}

/** 评分与评论 */
export interface MarketplaceReview {
  id: number;
  app_id: number;
  workspace_id: number;
  user_id: number;
  rating: number;
  comment?: string;
  created_at: string;
  updated_at: string;
}

/** 发布应用入参 */
export interface CreateMarketplaceAppInput {
  slug: string;
  name: string;
  description: string;
  category: MarketplaceCategory;
  publisher?: string;
  publisher_url?: string;
  icon_url?: string;
  screenshots?: string[];
  tags?: string[];
  pricing?: MarketplacePricing;
  install_type?: MarketplaceInstallType;
  install_config?: Record<string, unknown>;
  docs_url?: string;
}

/** 更新应用入参（字段可选，仅传需要修改的字段） */
export interface UpdateMarketplaceAppInput {
  name?: string;
  description?: string;
  publisher_url?: string;
  icon_url?: string;
  screenshots?: string[];
  tags?: string[];
  pricing?: MarketplacePricing;
  install_config?: Record<string, unknown>;
  docs_url?: string;
  status?: MarketplaceStatus;
}

/** 列表查询参数 */
export interface ListMarketplaceAppsParams {
  category?: MarketplaceCategory;
  q?: string;
  limit?: number;
  offset?: number;
}

/** 评分提交入参 */
export interface CreateMarketplaceReviewInput {
  rating: number;
  comment?: string;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 应用商店域 API */
export const marketplaceApi = {
  /* --- 目录 --- */

  /** 分类列表 */
  listCategories: (workspaceId: number | string) =>
    wrap<{ items: CategoryInfo[] }>(
      http.get(`/workspaces/${workspaceId}/marketplace/categories`),
    ),

  /** 应用目录（可按分类 / 关键字过滤，分页） */
  listApps: (
    workspaceId: number | string,
    params?: ListMarketplaceAppsParams,
  ) =>
    wrap<{ items: MarketplaceApp[] }>(
      http.get(`/workspaces/${workspaceId}/marketplace/apps`, { params }),
    ),

  /** 应用详情（按 slug） */
  getApp: (workspaceId: number | string, slug: string) =>
    wrap<MarketplaceApp>(
      http.get(`/workspaces/${workspaceId}/marketplace/apps/${slug}`),
    ),

  /** 评分列表 */
  listReviews: (
    workspaceId: number | string,
    appId: number | string,
    limit = 20,
  ) =>
    wrap<{ items: MarketplaceReview[] }>(
      http.get(`/workspaces/${workspaceId}/marketplace/apps/${appId}/reviews`, {
        params: { limit },
      }),
    ),

  /* --- 安装 --- */

  /** 已安装应用（仅 enabled） */
  listInstalled: (workspaceId: number | string) =>
    wrap<{ items: MarketplaceApp[] }>(
      http.get(`/workspaces/${workspaceId}/marketplace/installed`),
    ),

  /** 安装应用（幂等） */
  install: (
    workspaceId: number | string,
    appId: number | string,
    config?: Record<string, unknown>,
  ) =>
    wrap<MarketplaceInstall>(
      http.post(
        `/workspaces/${workspaceId}/marketplace/apps/${appId}/install`,
        { config },
      ),
    ),

  /** 卸载应用（软删除：enabled = false） */
  uninstall: (workspaceId: number | string, appId: number | string) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/marketplace/apps/${appId}/install`,
      ),
    ),

  /* --- 评分 --- */

  /** 提交 / 更新评分 */
  createReview: (
    workspaceId: number | string,
    appId: number | string,
    input: CreateMarketplaceReviewInput,
  ) =>
    wrap<MarketplaceReview>(
      http.post(
        `/workspaces/${workspaceId}/marketplace/apps/${appId}/reviews`,
        input,
      ),
    ),

  /* --- 发布 --- */

  /** 发布应用 */
  createApp: (
    workspaceId: number | string,
    input: CreateMarketplaceAppInput,
  ) =>
    wrap<MarketplaceApp>(
      http.post(`/workspaces/${workspaceId}/marketplace/apps`, input),
    ),

  /** 更新应用 */
  updateApp: (
    workspaceId: number | string,
    slug: string,
    input: UpdateMarketplaceAppInput,
  ) =>
    wrap<MarketplaceApp>(
      http.patch(
        `/workspaces/${workspaceId}/marketplace/apps/${slug}`,
        input,
      ),
    ),
};
