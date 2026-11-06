<script setup lang="ts">
/**
 * MarketplaceView — 应用商店。
 *
 * 功能：
 *   - 顶部分类 Tab（全部 / 应用 / 导入器 / 模板 / 集成）；
 *   - 搜索框（按名称 / 描述过滤）；
 *   - 应用网格：图标 / 名称 / 描述 / 分类徽章 / 评分星级 / 安装数 / 定价徽章；
 *   - 已安装区：应用列表 + 启用 / 停用（卸载）；
 *   - 应用详情弹窗：完整描述、截图、标签、安装按钮、评价列表 + 提交评价表单；
 *   - 发布应用弹窗（名称 / slug / 描述 / 分类 / 标签 / 定价 / 安装方式 / 图标 / 文档）；
 *   - 加载 / 错误 / 空状态。
 */
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";

import {
  marketplaceApi,
  type CategoryInfo,
  type MarketplaceApp,
  type MarketplaceCategory,
  type MarketplaceInstallType,
  type MarketplacePricing,
  type MarketplaceReview,
} from "@/api/services/marketplace";
import { useTranslation } from "@/composables/useTranslation";
import {
  AppBadge,
  AppButton,
  AppEmptyState,
  AppErrorState,
  AppLoadingState,
  AppModal,
  AppSkeleton,
} from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useConfirm } from "@/composables/useConfirm";
import { toast } from "@/lib/toast";

const { t } = useTranslation();
const confirm = useConfirm();
const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

/* ------------------------------------------------------------------ */
/* 常量                                                                */
/* ------------------------------------------------------------------ */

/** 分类选项（含「全部」） */
const CATEGORY_OPTIONS: Array<{ value: MarketplaceCategory | "all"; labelKey: string }> = [
  { value: "all", labelKey: "marketplace.categoryAll" },
  { value: "apps", labelKey: "marketplace.categoryApps" },
  { value: "importers", labelKey: "marketplace.categoryImporters" },
  { value: "templates", labelKey: "marketplace.categoryTemplates" },
  { value: "integrations", labelKey: "marketplace.categoryIntegrations" },
];

const PRICING_META: Record<MarketplacePricing, { labelKey: string; variant: "success" | "warning" | "info" }> = {
  free: { labelKey: "marketplace.pricingFree", variant: "success" },
  paid: { labelKey: "marketplace.pricingPaid", variant: "warning" },
  freemium: { labelKey: "marketplace.pricingFreemium", variant: "info" },
};

const INSTALL_TYPE_META: Record<MarketplaceInstallType, string> = {
  builtin: "marketplace.installTypeBuiltin",
  oauth: "marketplace.installTypeOauth",
  webhook: "marketplace.installTypeWebhook",
  cli: "marketplace.installTypeCli",
};

/* ------------------------------------------------------------------ */
/* 状态                                                                */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const categories = ref<CategoryInfo[]>([]);
const apps = ref<MarketplaceApp[]>([]);
const installed = ref<MarketplaceApp[]>([]);

const activeCategory = ref<MarketplaceCategory | "all">("all");
const searchInput = ref("");
const searchQuery = ref("");
const searchTimer = ref<number | null>(null);

/** 已安装应用 ID 集合（用于网格中的「已安装」标记） */
const installedIds = computed<Set<number>>(() => new Set(installed.value.map((a) => a.id)));

/** 网格中是否正在安装某个应用 */
const installingId = ref<number | null>(null);

/* ------------------------------------------------------------------ */
/* 数据加载                                                            */
/* ------------------------------------------------------------------ */

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const [catRes, installedRes] = await Promise.all([
      marketplaceApi.listCategories(workspaceId.value),
      marketplaceApi.listInstalled(workspaceId.value),
    ]);
    categories.value = catRes.items ?? [];
    installed.value = installedRes.items ?? [];
    await loadApps();
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t("marketplace.loadFailed");
  } finally {
    loading.value = false;
  }
}

async function loadApps() {
  if (!workspaceId.value) return;
  try {
    const res = await marketplaceApi.listApps(workspaceId.value, {
      category: activeCategory.value === "all" ? undefined : activeCategory.value,
      q: searchQuery.value || undefined,
      limit: 100,
      offset: 0,
    });
    apps.value = res.items ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t("marketplace.loadFailed");
  }
}

/** 搜索防抖 */
function onSearchInput() {
  if (searchTimer.value) window.clearTimeout(searchTimer.value);
  searchTimer.value = window.setTimeout(() => {
    searchQuery.value = searchInput.value.trim();
    loadApps();
  }, 300);
}

function selectCategory(cat: MarketplaceCategory | "all") {
  activeCategory.value = cat;
  loadApps();
}

/* ------------------------------------------------------------------ */
/* 显示辅助                                                            */
/* ------------------------------------------------------------------ */

/** 应用图标：优先用 icon_url，否则用 lucide 名或首字母 */
function appIcon(app: MarketplaceApp): string {
  return app.icon_url || "Box";
}

function iconIsUrl(icon: string): boolean {
  return icon.startsWith("http://") || icon.startsWith("https://") || icon.startsWith("/");
}

function categoryLabelKey(cat: MarketplaceCategory): string {
  const found = CATEGORY_OPTIONS.find((c) => c.value === cat);
  return found?.labelKey ?? "marketplace.categoryApps";
}

function pricingVariant(pricing: MarketplacePricing): "success" | "warning" | "info" {
  return PRICING_META[pricing]?.variant ?? "info";
}

function pricingLabel(pricing: MarketplacePricing): string {
  return t(PRICING_META[pricing]?.labelKey ?? "marketplace.pricingFree");
}

function formatRating(avg: number): string {
  return avg ? avg.toFixed(1) : "0.0";
}

/* ------------------------------------------------------------------ */
/* 详情弹窗                                                            */
/* ------------------------------------------------------------------ */

const showDetail = ref(false);
const detailLoading = ref(false);
const detailError = ref("");
const detailApp = ref<MarketplaceApp | null>(null);
const reviews = ref<MarketplaceReview[]>([]);
const reviewsLoading = ref(false);

async function openDetail(app: MarketplaceApp) {
  detailApp.value = app;
  detailError.value = "";
  reviews.value = [];
  showDetail.value = true;
  detailLoading.value = true;
  try {
    // 详情（按 slug 拉取，保证字段完整）
    const fresh = await marketplaceApi.getApp(workspaceId.value, app.slug);
    detailApp.value = fresh;
    // 同步网格中的对应项（评分 / 安装数可能变化）
    const idx = apps.value.findIndex((a) => a.id === fresh.id);
    if (idx >= 0) apps.value[idx] = fresh;
  } catch (err: unknown) {
    detailError.value = err instanceof Error ? err.message : t("marketplace.loadFailed");
  } finally {
    detailLoading.value = false;
  }
  await loadReviews(app.id);
}

async function loadReviews(appId: number) {
  reviewsLoading.value = true;
  try {
    const res = await marketplaceApi.listReviews(workspaceId.value, appId, 50);
    reviews.value = res.items ?? [];
  } catch {
    reviews.value = [];
  } finally {
    reviewsLoading.value = false;
  }
}

function closeDetail() {
  showDetail.value = false;
  detailApp.value = null;
  reviews.value = [];
}

/* ------------------------------------------------------------------ */
/* 安装 / 卸载                                                         */
/* ------------------------------------------------------------------ */

async function installApp(app: MarketplaceApp) {
  if (installingId.value) return;
  installingId.value = app.id;
  try {
    await marketplaceApi.install(workspaceId.value, app.id);
    toast.success(t("marketplace.installSuccess", { name: app.name }));
    const res = await marketplaceApi.listInstalled(workspaceId.value);
    installed.value = res.items ?? [];
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("marketplace.installFailed"));
  } finally {
    installingId.value = null;
  }
}

async function uninstallApp(app: MarketplaceApp) {
  if (!await confirm({
    title: t("marketplace.uninstallConfirmTitle"),
    message: t("marketplace.uninstallConfirmMessage", { name: app.name }),
    confirmText: t("marketplace.uninstall"),
    variant: "danger",
  })) return;
  try {
    await marketplaceApi.uninstall(workspaceId.value, app.id);
    toast.success(t("marketplace.uninstallSuccess", { name: app.name }));
    const res = await marketplaceApi.listInstalled(workspaceId.value);
    installed.value = res.items ?? [];
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("marketplace.uninstallFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 评价表单                                                            */
/* ------------------------------------------------------------------ */

const reviewForm = reactive({ rating: 0, comment: "" });
const reviewSubmitting = ref(false);

function setRating(n: number) {
  reviewForm.rating = n;
}

async function submitReview() {
  const app = detailApp.value;
  if (!app || reviewSubmitting.value) return;
  if (reviewForm.rating < 1) {
    toast.warning(t("marketplace.ratingRequired"));
    return;
  }
  reviewSubmitting.value = true;
  try {
    await marketplaceApi.createReview(workspaceId.value, app.id, {
      rating: reviewForm.rating,
      comment: reviewForm.comment.trim(),
    });
    toast.success(t("marketplace.reviewSuccess"));
    reviewForm.rating = 0;
    reviewForm.comment = "";
    await loadReviews(app.id);
    // 刷新详情评分
    const fresh = await marketplaceApi.getApp(workspaceId.value, app.slug);
    detailApp.value = fresh;
    const idx = apps.value.findIndex((a) => a.id === fresh.id);
    if (idx >= 0) apps.value[idx] = fresh;
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("marketplace.reviewFailed"));
  } finally {
    reviewSubmitting.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 发布应用弹窗                                                        */
/* ------------------------------------------------------------------ */

const showPublish = ref(false);
const publishSaving = ref(false);

interface PublishForm {
  name: string;
  slug: string;
  description: string;
  category: MarketplaceCategory | "";
  tags: string;
  pricing: MarketplacePricing;
  install_type: MarketplaceInstallType;
  icon_url: string;
  docs_url: string;
  publisher: string;
}

const publishForm = reactive<PublishForm>({
  name: "",
  slug: "",
  description: "",
  category: "",
  tags: "",
  pricing: "free",
  install_type: "builtin",
  icon_url: "",
  docs_url: "",
  publisher: "",
});

function openPublish() {
  publishForm.name = "";
  publishForm.slug = "";
  publishForm.description = "";
  publishForm.category = "";
  publishForm.tags = "";
  publishForm.pricing = "free";
  publishForm.install_type = "builtin";
  publishForm.icon_url = "";
  publishForm.docs_url = "";
  publishForm.publisher = "";
  showPublish.value = true;
}

function closePublish() {
  showPublish.value = false;
}

async function submitPublish() {
  if (publishSaving.value) return;
  if (!publishForm.name.trim()) {
    toast.warning(t("marketplace.nameRequired"));
    return;
  }
  if (!publishForm.category) {
    toast.warning(t("marketplace.categoryRequired"));
    return;
  }
  publishSaving.value = true;
  try {
    const tags = publishForm.tags
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    await marketplaceApi.createApp(workspaceId.value, {
      slug: publishForm.slug.trim(),
      name: publishForm.name.trim(),
      description: publishForm.description.trim(),
      category: publishForm.category,
      publisher: publishForm.publisher.trim() || undefined,
      icon_url: publishForm.icon_url.trim() || undefined,
      docs_url: publishForm.docs_url.trim() || undefined,
      tags,
      pricing: publishForm.pricing,
      install_type: publishForm.install_type,
    });
    toast.success(t("marketplace.publishSuccess"));
    closePublish();
    await loadApps();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("marketplace.publishFailed"));
  } finally {
    publishSaving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 生命周期                                                            */
/* ------------------------------------------------------------------ */

watch(workspaceId, (id) => {
  if (id > 0) load();
});

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <!-- 头部 -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">{{ t("marketplace.title") }}</h1>
        <p class="mt-1 text-sm text-[var(--text-tertiary)]">{{ t("marketplace.subtitle") }}</p>
      </div>
      <div class="flex items-center gap-2">
        <div class="relative">
          <span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-tertiary)]">
            <DynamicIcon name="Search" :size="16" />
          </span>
          <input
            v-model="searchInput"
            type="text"
            class="h-9 w-64 rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] pl-9 pr-3 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
            :placeholder="t('marketplace.searchPlaceholder')"
            @input="onSearchInput"
          />
        </div>
        <AppButton variant="primary" @click="openPublish">
          {{ t("marketplace.publishApp") }}
        </AppButton>
      </div>
    </div>

    <!-- 分类 Tab -->
    <div class="flex flex-wrap items-center gap-2">
      <button
        v-for="opt in CATEGORY_OPTIONS"
        :key="opt.value"
        type="button"
        class="rounded-full border px-3 py-1.5 text-sm transition"
        :class="
          activeCategory === opt.value
            ? 'border-[var(--brand-500)] bg-[var(--brand-50)] text-[var(--brand-600)]'
            : 'border-[var(--border-default)] bg-[var(--surface-1)] text-[var(--text-secondary)] hover:bg-[var(--surface-2)]'
        "
        @click="selectCategory(opt.value)"
      >
        {{ t(opt.labelKey) }}
      </button>
    </div>

    <!-- 加载 -->
    <div v-if="loading" class="space-y-3">
      <AppSkeleton class="h-32 w-full" />
      <AppSkeleton class="h-32 w-full" />
      <AppSkeleton class="h-32 w-full" />
    </div>

    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 已安装 -->
      <section>
        <h2 class="mb-3 text-base font-semibold text-[var(--text-primary)]">
          {{ t("marketplace.installedTitle") }}
        </h2>
        <AppEmptyState
          v-if="installed.length === 0"
          :title="t('marketplace.installedEmptyTitle')"
          :description="t('marketplace.installedEmptyDescription')"
        />
        <div v-else class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="app in installed"
            :key="app.id"
            class="flex items-start justify-between gap-3 rounded-lg border border-[var(--border-subtle)] bg-[var(--surface-1)] p-4"
          >
            <div class="flex min-w-0 items-start gap-3">
              <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded-md bg-[var(--surface-2)]">
                <img v-if="iconIsUrl(appIcon(app))" :src="appIcon(app)" :alt="app.name" class="h-10 w-10 object-cover" />
                <DynamicIcon v-else :name="appIcon(app)" :size="20" />
              </div>
              <div class="min-w-0">
                <div class="truncate text-sm font-medium text-[var(--text-primary)]">{{ app.name }}</div>
                <div class="mt-1 flex items-center gap-1.5">
                  <AppBadge variant="success">{{ t("marketplace.enabled") }}</AppBadge>
                  <span class="text-xs text-[var(--text-tertiary)]">{{ t(categoryLabelKey(app.category)) }}</span>
                </div>
              </div>
            </div>
            <div class="flex flex-shrink-0 items-center gap-2">
              <AppButton variant="ghost" size="sm" @click="openDetail(app)">
                {{ t("common.more") }}
              </AppButton>
              <AppButton variant="ghost" size="sm" @click="uninstallApp(app)">
                {{ t("marketplace.uninstall") }}
              </AppButton>
            </div>
          </div>
        </div>
      </section>

      <!-- 应用网格 -->
      <section>
        <div v-if="apps.length === 0" class="py-8">
          <AppEmptyState
            v-if="searchQuery"
            :title="t('marketplace.searchEmptyTitle')"
            :description="t('marketplace.searchEmptyDescription')"
          />
          <AppEmptyState
            v-else
            :title="t('marketplace.emptyTitle')"
            :description="t('marketplace.emptyDescription')"
          />
        </div>
        <div v-else class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="app in apps"
            :key="app.id"
            class="group flex flex-col rounded-lg border border-[var(--border-subtle)] bg-[var(--surface-1)] p-4 transition hover:border-[var(--border-default)]"
          >
            <div class="flex items-start gap-3">
              <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center overflow-hidden rounded-lg bg-[var(--surface-2)]">
                <img v-if="iconIsUrl(appIcon(app))" :src="appIcon(app)" :alt="app.name" class="h-12 w-12 object-cover" />
                <DynamicIcon v-else :name="appIcon(app)" :size="22" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span class="truncate text-sm font-semibold text-[var(--text-primary)]">{{ app.name }}</span>
                  <AppBadge v-if="installedIds.has(app.id)" variant="success">
                    {{ t("marketplace.installedTitle") }}
                  </AppBadge>
                </div>
                <div class="mt-0.5 truncate text-xs text-[var(--text-tertiary)]">
                  {{ app.publisher || "official" }}
                </div>
              </div>
            </div>

            <p class="mt-3 line-clamp-2 flex-1 text-xs leading-relaxed text-[var(--text-secondary)]">
              {{ app.description }}
            </p>

            <div class="mt-3 flex flex-wrap items-center gap-1.5">
              <AppBadge variant="brand">{{ t(categoryLabelKey(app.category)) }}</AppBadge>
              <AppBadge :variant="pricingVariant(app.pricing)">{{ pricingLabel(app.pricing) }}</AppBadge>
              <AppBadge variant="default">{{ t(INSTALL_TYPE_META[app.install_type] ?? "marketplace.installTypeBuiltin") }}</AppBadge>
            </div>

            <div class="mt-3 flex items-center justify-between border-t border-[var(--border-subtle)] pt-3">
              <div class="flex items-center gap-3 text-xs text-[var(--text-tertiary)]">
                <span class="flex items-center gap-1">
                  <DynamicIcon name="Star" :size="13" class="text-[var(--warning-500)]" />
                  {{ formatRating(app.rating_avg) }}
                  <span class="text-[var(--text-tertiary)]">({{ app.rating_count }})</span>
                </span>
                <span class="flex items-center gap-1">
                  <DynamicIcon name="Download" :size="13" />
                  {{ app.install_count }}
                </span>
              </div>
              <div class="flex items-center gap-2">
                <AppButton variant="ghost" size="sm" @click="openDetail(app)">
                  {{ t("common.more") }}
                </AppButton>
                <AppButton
                  v-if="!installedIds.has(app.id)"
                  variant="primary"
                  size="sm"
                  :loading="installingId === app.id"
                  @click="installApp(app)"
                >
                  {{ t("marketplace.install") }}
                </AppButton>
                <AppButton v-else variant="secondary" size="sm" @click="uninstallApp(app)">
                  {{ t("marketplace.uninstall") }}
                </AppButton>
              </div>
            </div>
          </div>
        </div>
      </section>
    </template>

    <!-- 应用详情弹窗 -->
    <AppModal
      :visible="showDetail"
      :title="t('marketplace.detailTitle')"
      width="720px"
      @close="closeDetail"
    >
      <div class="space-y-5">
        <AppLoadingState v-if="detailLoading" :text="t('common.loading')" />
        <AppErrorState v-else-if="detailError" :message="detailError" @retry="openDetail(detailApp!)" />
        <template v-else-if="detailApp">
          <!-- 头部信息 -->
          <div class="flex items-start gap-4">
            <div class="flex h-16 w-16 flex-shrink-0 items-center justify-center overflow-hidden rounded-xl bg-[var(--surface-2)]">
              <img v-if="iconIsUrl(appIcon(detailApp))" :src="appIcon(detailApp)" :alt="detailApp.name" class="h-16 w-16 object-cover" />
              <DynamicIcon v-else :name="appIcon(detailApp)" :size="30" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <h3 class="text-lg font-semibold text-[var(--text-primary)]">{{ detailApp.name }}</h3>
                <AppBadge v-if="installedIds.has(detailApp.id)" variant="success">
                  {{ t("marketplace.installedTitle") }}
                </AppBadge>
              </div>
              <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-[var(--text-tertiary)]">
                <span>{{ t("marketplace.publisher") }}: {{ detailApp.publisher || "official" }}</span>
                <span>·</span>
                <span class="flex items-center gap-1">
                  <DynamicIcon name="Star" :size="13" class="text-[var(--warning-500)]" />
                  {{ formatRating(detailApp.rating_avg) }}
                  ({{ t("marketplace.ratingCount", { count: detailApp.rating_count }) }})
                </span>
                <span>·</span>
                <span>{{ t("marketplace.installCount", { count: detailApp.install_count }) }}</span>
              </div>
              <div class="mt-2 flex flex-wrap items-center gap-1.5">
                <AppBadge variant="brand">{{ t(categoryLabelKey(detailApp.category)) }}</AppBadge>
                <AppBadge :variant="pricingVariant(detailApp.pricing)">{{ pricingLabel(detailApp.pricing) }}</AppBadge>
                <AppBadge variant="default">{{ t(INSTALL_TYPE_META[detailApp.install_type] ?? "marketplace.installTypeBuiltin") }}</AppBadge>
                <a
                  v-if="detailApp.docs_url"
                  :href="detailApp.docs_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-xs text-[var(--brand-600)] hover:underline"
                >
                  {{ t("marketplace.docs") }} ↗
                </a>
              </div>
            </div>
          </div>

          <!-- 描述 -->
          <div>
            <h4 class="mb-1 text-xs font-medium text-[var(--text-tertiary)]">{{ t("marketplace.description") }}</h4>
            <p class="whitespace-pre-wrap text-sm leading-relaxed text-[var(--text-secondary)]">
              {{ detailApp.description }}
            </p>
          </div>

          <!-- 截图 -->
          <div v-if="detailApp.screenshots && detailApp.screenshots.length > 0">
            <h4 class="mb-2 text-xs font-medium text-[var(--text-tertiary)]">{{ t("marketplace.screenshots") }}</h4>
            <div class="flex flex-wrap gap-2">
              <img
                v-for="(shot, i) in detailApp.screenshots"
                :key="i"
                :src="shot"
                :alt="`${detailApp.name} screenshot ${i + 1}`"
                class="h-28 w-44 rounded-md border border-[var(--border-subtle)] object-cover"
              />
            </div>
          </div>

          <!-- 标签 -->
          <div v-if="detailApp.tags && detailApp.tags.length > 0">
            <h4 class="mb-2 text-xs font-medium text-[var(--text-tertiary)]">{{ t("marketplace.tags") }}</h4>
            <div class="flex flex-wrap gap-1.5">
              <AppBadge v-for="tag in detailApp.tags" :key="tag" variant="default">{{ tag }}</AppBadge>
            </div>
          </div>

          <!-- 操作 -->
          <div class="flex items-center justify-end gap-2 border-t border-[var(--border-subtle)] pt-4">
            <AppButton v-if="!installedIds.has(detailApp.id)" variant="primary" @click="installApp(detailApp)">
              {{ t("marketplace.install") }}
            </AppButton>
            <AppButton v-else variant="secondary" @click="uninstallApp(detailApp)">
              {{ t("marketplace.uninstall") }}
            </AppButton>
          </div>

          <!-- 评价 -->
          <div class="border-t border-[var(--border-subtle)] pt-4">
            <h4 class="mb-3 text-sm font-semibold text-[var(--text-primary)]">
              {{ t("marketplace.reviews") }}
            </h4>

            <!-- 提交表单 -->
            <div class="mb-4 rounded-md border border-[var(--border-subtle)] bg-[var(--surface-2)] p-3">
              <div class="mb-2 flex items-center gap-2">
                <span class="text-xs text-[var(--text-tertiary)]">{{ t("marketplace.ratingLabel") }}</span>
                <div class="flex items-center gap-0.5">
                  <button
                    v-for="n in 5"
                    :key="n"
                    type="button"
                    class="p-0.5"
                    :aria-label="`${n} star`"
                    @click="setRating(n)"
                  >
                    <DynamicIcon
                      name="Star"
                      :size="18"
                      :class="n <= reviewForm.rating ? 'text-[var(--warning-500)]' : 'text-[var(--text-tertiary)]'"
                    />
                  </button>
                </div>
              </div>
              <textarea
                v-model="reviewForm.comment"
                rows="2"
                class="w-full resize-none rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
                :placeholder="t('marketplace.commentPlaceholder')"
              />
              <div class="mt-2 flex justify-end">
                <AppButton
                  variant="primary"
                  size="sm"
                  :loading="reviewSubmitting"
                  @click="submitReview"
                >
                  {{ t("marketplace.submitReview") }}
                </AppButton>
              </div>
            </div>

            <!-- 列表 -->
            <AppLoadingState v-if="reviewsLoading" :text="t('common.loading')" />
            <AppEmptyState
              v-else-if="reviews.length === 0"
              :title="t('marketplace.reviewsEmpty')"
              compact
            />
            <div v-else class="space-y-3">
              <div
                v-for="review in reviews"
                :key="review.id"
                class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] p-3"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-1">
                    <DynamicIcon
                      v-for="n in 5"
                      :key="n"
                      name="Star"
                      :size="14"
                      :class="n <= review.rating ? 'text-[var(--warning-500)]' : 'text-[var(--text-tertiary)]'"
                    />
                  </div>
                  <span class="text-xs text-[var(--text-tertiary)]">
                    {{ new Date(review.created_at).toLocaleDateString() }}
                  </span>
                </div>
                <p v-if="review.comment" class="mt-1.5 text-sm text-[var(--text-secondary)]">
                  {{ review.comment }}
                </p>
              </div>
            </div>
          </div>
        </template>
      </div>
    </AppModal>

    <!-- 发布应用弹窗 -->
    <AppModal
      :visible="showPublish"
      :title="t('marketplace.publishTitle')"
      width="560px"
      @close="closePublish"
    >
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("marketplace.formName") }}
            <span class="text-[var(--danger-500)]">*</span>
          </label>
          <input
            v-model="publishForm.name"
            type="text"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
        </div>

        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("marketplace.formSlug") }}
          </label>
          <input
            v-model="publishForm.slug"
            type="text"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
          <p class="mt-1 text-xs text-[var(--text-tertiary)]">{{ t("marketplace.formSlugHint") }}</p>
        </div>

        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("marketplace.formDescription") }}
          </label>
          <textarea
            v-model="publishForm.description"
            rows="3"
            class="w-full resize-none rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("marketplace.formCategory") }}
              <span class="text-[var(--danger-500)]">*</span>
            </label>
            <select
              v-model="publishForm.category"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)]"
            >
              <option value="" disabled>{{ t("common.filter") }}</option>
              <option
                v-for="opt in CATEGORY_OPTIONS.filter((c) => c.value !== 'all')"
                :key="opt.value"
                :value="opt.value"
              >
                {{ t(opt.labelKey) }}
              </option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("marketplace.formPricing") }}
            </label>
            <select
              v-model="publishForm.pricing"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)]"
            >
              <option value="free">{{ t("marketplace.pricingFree") }}</option>
              <option value="paid">{{ t("marketplace.pricingPaid") }}</option>
              <option value="freemium">{{ t("marketplace.pricingFreemium") }}</option>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("marketplace.formInstallType") }}
            </label>
            <select
              v-model="publishForm.install_type"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)]"
            >
              <option value="builtin">{{ t("marketplace.installTypeBuiltin") }}</option>
              <option value="oauth">{{ t("marketplace.installTypeOauth") }}</option>
              <option value="webhook">{{ t("marketplace.installTypeWebhook") }}</option>
              <option value="cli">{{ t("marketplace.installTypeCli") }}</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("marketplace.formPublisher") }}
            </label>
            <input
              v-model="publishForm.publisher"
              type="text"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
            />
          </div>
        </div>

        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("marketplace.formTags") }}
          </label>
          <input
            v-model="publishForm.tags"
            type="text"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
          <p class="mt-1 text-xs text-[var(--text-tertiary)]">{{ t("marketplace.formTagsHint") }}</p>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("marketplace.formIconUrl") }}
            </label>
            <input
              v-model="publishForm.icon_url"
              type="text"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("marketplace.formDocsUrl") }}
            </label>
            <input
              v-model="publishForm.docs_url"
              type="text"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
            />
          </div>
        </div>
      </div>

      <template #footer>
        <AppButton variant="secondary" @click="closePublish">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton variant="primary" :loading="publishSaving" @click="submitPublish">
          {{ t("marketplace.publishApp") }}
        </AppButton>
      </template>
    </AppModal>
  </div>
</template>

<style scoped>
/* 星级按钮无默认样式 */
button {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  color: inherit;
}
</style>
