<script setup lang="ts">
/**
 * AppEmptyState — 情感化空状态组件。
 *
 * 相比 v1 新增：
 *   - scenario: 预设场景模板 (issues / projects / cycles / modules / search / inbox / notifications / labels / error)
 *   - illustration: 内置 SVG 插画系统 (无需外部资源)
 *   - cta / secondaryCta: 主/次行动按钮
 *   - illustrationSize: sm / md / lg
 *
 * 使用方式：
 *   <AppEmpty-state scenario="issues" />
 *   <AppEmptyState title="暂无数据" description="开始创建第一个工作项">
 *     <template #cta><button>创建工作项</button></template>
 *   </AppEmptyState>
 */

export type EmptyScenario =
  | "default"
  | "issues"
  | "projects"
  | "sprints"
  | "modules"
  | "search"
  | "intake"
  | "notifications"
  | "labels"
  | "members"
  | "analytics"
  | "views"
  | "inbox"
  | "api-token"
  | "webhooks"
  | "error"

export type IllustrationSize = "sm" | "md" | "lg"

const props = withDefaults(
  defineProps<{
    /** 预设场景 — 自动配好插画+标题+描述 */
    scenario?: EmptyScenario
    /** 主标题 (scenario 未提供时使用) */
    title?: string
    /** 副描述 (scenario 未提供时使用) */
    description?: string
    /** 插画尺寸 */
    illustrationSize?: IllustrationSize
    /** 自定义图标/emoji（优先级低于 scenario 内置插画） */
    icon?: string
  }>(),
  {
    scenario: "default",
    title: "",
    description: "",
    illustrationSize: "md",
    icon: "",
  },
)

const emit = defineEmits<{
  "cta-click": []
  "secondary-click": []
}>()

/* ---- 场景模板 ---- */
const scenarioMap: Record<Exclude<EmptyScenario, "default">, { title: string; description: string }> = {
  issues: {
    title: "还没有工作项",
    description: "创建第一个需求、任务或缺陷，开始管理你的项目进度。",
  },
  projects: {
    title: "还没有项目",
    description: "创建一个项目来组织工作项，与团队协作推进。",
  },
  sprints: {
    title: "还没有迭代",
    description: "规划你的第一个冲刺，设定目标和时间范围。",
  },
  modules: {
    title: "还没有模块",
    description: "将复杂项目拆分为更小的模块，便于分工与管理。",
  },
  search: {
    title: "未找到匹配结果",
    description: "尝试调整搜索关键字或筛选条件。",
  },
  intake: {
    title: "收件箱是空的",
    description: "外部提交的工作项会出现在这里，等待你审核与转正。",
  },
  notifications: {
    title: "暂无通知",
    description: "当有新的工作项、评论或提及时，你会收到通知。",
  },
  labels: {
    title: "还没有标签",
    description: "创建标签来分类和标记工作项，方便筛选。",
  },
  members: {
    title: "还没有成员",
    description: "邀请团队成员加入项目，共同协作。",
  },
  analytics: {
    title: "暂无数据可分析",
    description: "当项目有进度变更后，这里会显示效能趋势。",
  },
  views: {
    title: "还没有自定义视图",
    description: "保存当前的筛选和分组条件，创建专属视图。",
  },
  inbox: {
    title: "收件箱是空的",
    description: "所有外部提交的工作项会显示在此。",
  },
  "api-token": {
    title: "还没有 API 令牌",
    description: "创建令牌以通过 API 访问工作空间数据。",
  },
  webhooks: {
    title: "还没有 Webhook",
    description: "配置 Webhook 将事件推送到外部系统。",
  },
  error: {
    title: "出错了",
    description: "加载数据时遇到问题，请稍后重试或联系管理员。",
  },
}

const resolvedTitle = computed(() => {
  if (props.scenario !== "default" && scenarioMap[props.scenario]) {
    return scenarioMap[props.scenario].title
  }
  return props.title || "暂无数据"
})

const resolvedDescription = computed(() => {
  if (props.scenario !== "default" && scenarioMap[props.scenario]) {
    return scenarioMap[props.scenario].description
  }
  return props.description
})

/* ---- 尺寸 ---- */
const sizeMap: Record<IllustrationSize, number> = { sm: 64, md: 96, lg: 128 }
const illustrationPx = computed(() => sizeMap[props.illustrationSize])

import { computed } from "vue"
</script>

<template>
  <div class="app-empty" :class="[`app-empty--${illustrationSize}`]">
    <!-- 插画区 -->
    <div v-if="icon" class="app-empty__emoji">{{ icon </div>
    <component
      v-else-if="scenario !== 'default' && scenario !== 'error'"
      :is="'svg'"
      class="app-empty__illustration"
      :width="illustrationPx"
      :height="illustrationPx"
      :viewBox="scenario === 'search' ? '0 0 96 96' : '0 0 96 96'"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <!-- 默认空状态插画 -->
      <template v-if="scenario === 'default'">
        <circle cx="48" cy="48" r="46" :fill="`var(--neutral-200, #f0f0f0)`" />
        <rect x="28" y="36" width="40" height="4" rx="2" :fill="`var(--neutral-600, #999)`" />
        <rect x="28" y="46" width="30" height="3" rx="1.5" :fill="`var(--neutral-500, #bbb)`" />
        <rect x="28" y="54" width="35" height="3" rx="1.5" :fill="`var(--neutral-500, #bbb)`" />
      </template>

      <!-- Issues: 卡片 + 勾选 -->
      <template v-else-if="scenario === 'issues'">
        <rect x="14" y="20" width="68" height="56" rx="8" :fill="`var(--neutral-200)`" />
        <rect x="22" y="28" width="28" height="4" rx="2" :fill="`var(--neutral-600)`" opacity="0.7" />
        <rect x="22" y="38" width="40" height="3" rx="1.5" :fill="`var(--neutral-500)`" opacity="0.5" />
        <rect x="22" y="46" width="34" height="3" rx="1.5" :fill="`var(--neutral-500)`" opacity="0.5" />
        <rect x="22" y="58" width="16" height="8" rx="2" :fill="`var(--brand-default)`" opacity="0.15" />
        <circle cx="72" cy="22" r="14" :fill="`var(--brand-default)`" />
        <path d="M66 22l4 4 8-8" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
      </template>

      <!-- Projects: 文件夹 -->
      <template v-else-if="scenario === 'projects'">
        <path d="M16 28a4 4 0 014-4h16l6 6h24a4 4 0 014 4v36a4 4 0 01-4 4H20a4 4 0 01-4-4V28z" :fill="`var(--brand-300)`" opacity="0.3" />
        <path d="M16 36h64a4 4 0 014 4v28a4 4 0 01-4 4H16a4 4 0 01-4-4V40a4 4 0 014-4z" :fill="`var(--neutral-200)`" />
        <rect x="22" y="46" width="18" height="4" rx="2" :fill="`var(--neutral-600)`" opacity="0.6" />
        <rect x="22" y="54" width="30" height="3" rx="1.5" :fill="`var(--neutral-500)`" opacity="0.4" />
      </template>

      <!-- Sprints / Cycles: 循环箭头 -->
      <template v-else-if="scenario === 'sprints'">
        <circle cx="48" cy="48" r="28" :fill="`var(--neutral-200)`" />
        <path d="M48 28a20 20 0 11-14 5.7" :stroke="`var(--brand-default)`" stroke-width="3" fill="none" stroke-linecap="round" />
        <path d="M32 22l16 6-6 16" :fill="`var(--brand-default)`" />
      </template>

      <!-- Search: 放大镜 -->
      <template v-else-if="scenario === 'search'">
        <circle cx="44" cy="44" r="24" :fill="`var(--neutral-200)`" />
        <circle cx="44" cy="44" r="14" :stroke="`var(--neutral-700)`" stroke-width="3" fill="none" />
        <line x1="54" y1="54" x2="72" y2="72" :stroke="`var(--neutral-700)`" stroke-width="3" stroke-linecap="round" />
        <!-- 空搜索结果 -->
        <circle cx="44" cy="44" r="6" :fill="`var(--neutral-500)`" opacity="0.3" />
      </template>

      <!-- Notifications: 铃铛 -->
      <template v-else-if="scenario === 'notifications'">
        <path d="M48 18a20 20 0 0120 20v16l6 8H22l6-8V38a20 20 0 0120-20z" :fill="`var(--neutral-200)`" />
        <rect x="44" y="10" width="8" height="8" rx="4" :fill="`var(--brand-default)`" opacity="0.6" />
        <rect x="22" y="62" width="52" height="8" rx="4" :fill="`var(--neutral-300)`" />
        <circle cx="48" cy="76" r="6" :fill="`var(--neutral-500)`" opacity="0.5" />
      </template>

      <!-- Labels: 标签 -->
      <template v-else-if="scenario === 'labels'">
        <path d="M20 28a8 8 0 018-8h32l16 16v28a8 8 0 01-8 8H28l-16-16V28z" :fill="`var(--extended-color-indigo-50, #eef2fe)`" />
        <circle cx="32" cy="44" r="5" :fill="`var(--extended-color-indigo-700, #4f46e5)`" opacity="0.6" />
      </template>

      <!-- Members: 人群 -->
      <template v-else-if="scenario === 'members'">
        <circle cx="40" cy="32" r="10" :fill="`var(--neutral-300)`" />
        <circle cx="60" cy="36" r="8" :fill="`var(--neutral-300)`" opacity="0.8" />
        <circle cx="32" cy="40" r="7" :fill="`var(--neutral-300)`" opacity="0.6" />
        <path d="M18 68c0-12 10-20 22-20s22 8 22 20" :fill="`var(--neutral-200)`" />
        <path d="M50 68c0-8 6-14 14-14s14 6 14 14" :fill="`var(--neutral-200)`" opacity="0.8" />
      </template>

      <!-- Analytics: 图表 -->
      <template v-else-if="scenario === 'analytics'">
        <rect x="14" y="60" width="12" height="20" rx="2" :fill="`var(--neutral-400)`" opacity="0.4" />
        <rect x="30" y="48" width="12" height="32" rx="2" :fill="`var(--neutral-400)`" opacity="0.5" />
        <rect x="46" y="36" width="12" height="44" rx="2" :fill="`var(--brand-default)`" opacity="0.4" />
        <rect x="62" y="44" width="12" height="36" rx="2" :fill="`var(--neutral-400)`" opacity="0.5" />
        <rect x="74" y="28" width="12" height="52" rx="2" :fill="`var(--brand-default)`" opacity="0.6" />
      </template>

      <!-- Views: 看板 -->
      <template v-else-if="scenario === 'views'">
        <rect x="12" y="22" width="20" height="52" rx="4" :fill="`var(--neutral-200)`" />
        <rect x="38" y="22" width="20" height="52" rx="4" :fill="`var(--neutral-200)`" opacity="0.8" />
        <rect x="64" y="22" width="20" height="52" rx="4" :fill="`var(--neutral-200)`" opacity="0.6" />
        <rect x="16" y="30" width="12" height="3" rx="1.5" :fill="`var(--neutral-600)`" opacity="0.4" />
        <rect x="16" y="38" width="10" height="3" rx="1.5" :fill="`var(--neutral-500)`" opacity="0.3" />
        <rect x="42" y="30" width="12" height="3" rx="1.5" :fill="`var(--neutral-600)`" opacity="0.4" />
      </template>

      <!-- Intake / Inbox: 收件箱 -->
      <template v-else-if="scenario === 'intake' || scenario === 'inbox'">
        <rect x="20" y="32" width="56" height="40" rx="6" :fill="`var(--neutral-200)`" />
        <path d="M20 36l20 16a8 8 0 008 0l20-16" :stroke="`var(--neutral-500)`" stroke-width="2" fill="none" />
        <path d="M10 42l8 4 30-24 30 24 8-4" :stroke="`var(--brand-default)`" stroke-width="1.5" fill="none" opacity="0.4" stroke-dasharray="4 3" />
      </template>

      <!-- API Token: 钥匙 -->
      <template v-else-if="scenario === 'api-token'">
        <circle cx="56" cy="34" r="14" :stroke="`var(--neutral-400)`" stroke-width="4" fill="none" />
        <rect x="30" y="44" width="36" height="8" rx="2" transform="rotate(-45 30 44)" :fill="`var(--neutral-400)`" />
        <circle cx="50" cy="34" r="4" :fill="`var(--neutral-300)`" />
      </template>

      <!-- Webhooks: 连接 -->
      <template v-else-if="scenario === 'webhooks'">
        <circle cx="28" cy="48" r="10" :fill="`var(--brand-300)`" opacity="0.3" />
        <circle cx="68" cy="48" r="10" :fill="`var(--brand-300)`" opacity="0.3" />
        <path d="M38 48h20" :stroke="`var(--brand-default)`" stroke-width="2" stroke-dasharray="4 3" />
        <path d="M35 44l-6 4 6 4" :fill="`var(--brand-default)`" />
        <path d="M61 44l6 4-6 4" :fill="`var(--brand-default)`" />
      </template>

      <!-- Modules: 模块 -->
      <template v-else-if="scenario === 'modules'">
        <rect x="12" y="16" width="32" height="24" rx="4" :fill="`var(--neutral-200)`" />
        <rect x="52" y="16" width="32" height="24" rx="4" :fill="`var(--neutral-200)`" opacity="0.7" />
        <rect x="12" y="56" width="32" height="24" rx="4" :fill="`var(--neutral-200)`" opacity="0.8" />
        <rect x="52" y="56" width="32" height="24" rx="4" :fill="`var(--neutral-200)`" opacity="0.6" />
        <path d="M44 28h8" :stroke="`var(--neutral-400)`" stroke-width="2" stroke-linecap="round" />
        <path d="M28 48v8M68 48v8M44 68h8" :stroke="`var(--neutral-400)`" stroke-width="2" stroke-linecap="round" />
      </template>

      <!-- Error: 警告 -->
      <template v-else-if="scenario === 'error'">
        <circle cx="48" cy="48" r="36" :fill="`var(--red-100)`" />
        <path d="M48 30v22" :stroke="`var(--red-600)`" stroke-width="4" stroke-linecap="round" />
        <circle cx="48" cy="62" r="3" :fill="`var(--red-600)`" />
      </template>
    </component>

    <!-- Error inline SVG fallback -->
    <svg
      v-else-if="scenario === 'error'"
      class="app-empty__illustration"
      :width="illustrationPx"
      :height="illustrationPx"
      viewBox="0 0 96 96"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="48" cy="48" r="36" :fill="`var(--red-100)`" />
      <path d="M48 30v22" :stroke="`var(--red-600)`" stroke-width="4" stroke-linecap="round" />
      <circle cx="48" cy="62" r="3" :fill="`var(--red-600)`" />
    </svg>

    <!-- 文案 -->
    <p class="app-empty__title">{{ resolvedTitle }}</p>
    <p v-if="resolvedDescription" class="app-empty__description">{{ resolvedDescription }}</p>

    <!-- 插槽: 自定义 CTA / 按钮区 -->
    <div class="app-empty__actions">
      <slot>
        <AppButton
          v-if="scenario !== 'default' && scenario !== 'search' && scenario !== 'error'"
          variant="primary"
          size="sm"
          @click="emit('cta-click')"
        >
          {{ ctaLabelMap[scenario] ?? "开始创建" }}
        </AppButton>
        <AppButton
          v-if="scenario === 'error'"
          variant="secondary"
          size="sm"
          @click="emit('secondary-click')"
        >
          重试
        </AppButton>
      </slot>
    </div>
  </div>
</template>

<script lang="ts">
// CTA 标签映射
import AppButton from "./AppButton.vue"

export const ctaLabelMap: Record<EmptyScenario, string> = {
  default: "开始创建",
  issues: "创建工作项",
  projects: "创建项目",
  sprints: "规划迭代",
  modules: "创建模块",
  search: "",
  intake: "",
  notifications: "",
  labels: "创建标签",
  members: "邀请成员",
  analytics: "",
  views: "创建视图",
  inbox: "",
  "api-token": "创建令牌",
  webhooks: "配置 Webhook",
  error: "",
}
</script>

<style scoped>
.app-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
  gap: 8px;
}

.app-empty--sm { padding: 24px 16px; gap: 6px; }
.app-empty--md { padding: 48px 24px; gap: 8px; }
.app-empty--lg { padding: 64px 32px; gap: 12px; }

.app-empty__emoji {
  font-size: 32px;
  margin-bottom: 4px;
  line-height: 1;
}

.app-empty--sm .app-empty__emoji { font-size: 24px; }
.app-empty--md .app-empty__emoji { font-size: 32px; }
.app-empty--lg .app-empty__emoji { font-size: 40px; }

.app-empty__illustration {
  margin-bottom: 8px;
  flex-shrink: 0;
}

.app-empty--sm .app-empty__illustration { margin-bottom: 4px; }
.app-empty--lg .app-empty__illustration { margin-bottom: 12px; }

.app-empty__title {
  margin: 0;
  font-size: 15px;
  color: var(--txt-secondary);
  font-weight: 500;
  line-height: 1.4;
}

.app-empty--sm .app-empty__title { font-size: 13px; }
.app-empty--lg .app-empty__title { font-size: 17px; }

.app-empty__description {
  margin: 0;
  font-size: 13px;
  color: var(--txt-tertiary);
  max-width: 320px;
  line-height: 1.5;
}

.app-empty--sm .app-empty__description { font-size: 12px; max-width: 260px; }
.app-empty--lg .app-empty__description { font-size: 14px; max-width: 380px; }

.app-empty__actions {
  margin-top: 16px;
  display: flex;
  gap: 8px;
  align-items: center;
}

.app-empty--sm .app-empty__actions { margin-top: 10px; }
.app-empty--lg .app-empty__actions { margin-top: 20px; }
</style>
