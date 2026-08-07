<script setup lang="ts">
/**
 * AppEmptyState — 情感化空状态组件。
 *
 * 特性：
 *   - scenario: 预设场景模板 (issues / projects / sprints / modules / search / ...)
 *   - Inline SVG 插画系统（无需外部资源）
 *   - cta / secondaryCta: 主/次行动按钮
 *   - illustrationSize: sm / md / lg
 *
 * 使用方式：
 *   <AppEmptyState scenario="issues" @cta-click="createIssue" />
 *   <AppEmptyState title="暂无数据" description="开始创建第一个工作项" />
 */

import { computed } from "vue"

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
    /** 自定义图标/emoji（优先级高于 scenario 内置插画） */
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
const scenarioMap: Record<Exclude<EmptyScenario, "default" | "error">, { title: string; description: string }> = {
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
}

const resolvedTitle = computed(() => {
  if (props.scenario === "error") return "出错了"
  if (props.scenario !== "default" && scenarioMap[props.scenario]) {
    return scenarioMap[props.scenario].title
  }
  return props.title || "暂无数据"
})

const resolvedDescription = computed(() => {
  if (props.scenario === "error") return "加载数据时遇到问题，请稍后重试或联系管理员。"
  if (props.scenario !== "default" && scenarioMap[props.scenario]) {
    return scenarioMap[props.scenario].description
  }
  return props.description
})

/* ---- 尺寸 ---- */
const sizeMap: Record<IllustrationSize, number> = { sm: 64, md: 96, lg: 128 }
const illustrationPx = computed(() => sizeMap[props.illustrationSize])
</script>

<template>
  <div class="app-empty" :class="[`app-empty--${illustrationSize}`]">
    <!-- 自定义 emoji 图标（优先级最高） -->
    <div v-if="icon" class="app-empty__emoji">{{ icon }}</div>

    <!-- 场景插画（SVG 内联） -->
    <svg
      v-else-if="scenario !== 'default' && scenario !== 'error'"
      class="app-empty__illustration"
      :width="illustrationPx"
      :height="illustrationPx"
      viewBox="0 0 96 96"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <!-- Issues: 卡片 + 勾选 -->
      <g v-if="scenario === 'issues'">
        <rect x="14" y="20" width="68" height="56" rx="8" fill="var(--neutral-200, #e5e7eb)" />
        <rect x="22" y="28" width="28" height="4" rx="2" fill="var(--neutral-600, #9ca3af)" opacity="0.7" />
        <rect x="22" y="38" width="40" height="3" rx="1.5" fill="var(--neutral-500, #d1d5db)" opacity="0.5" />
        <rect x="22" y="46" width="34" height="3" rx="1.5" fill="var(--neutral-500, #d1d5db)" opacity="0.5" />
        <rect x="22" y="58" width="16" height="8" rx="2" fill="var(--brand-default, #3b82f6)" opacity="0.15" />
        <circle cx="72" cy="22" r="14" fill="var(--brand-default, #3b82f6)" />
        <path d="M66 22l4 4 8-8" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
      </g>

      <!-- Projects: 文件夹 -->
      <g v-else-if="scenario === 'projects'">
        <path d="M16 28a4 4 0 014-4h16l6 6h24a4 4 0 014 4v36a4 4 0 01-4 4H20a4 4 0 01-4-4V28z" fill="var(--brand-300, #93c5fd)" opacity="0.3" />
        <path d="M16 36h64a4 4 0 014 4v28a4 4 0 01-4 4H16V36z" fill="var(--neutral-200, #e5e7eb)" />
        <rect x="22" y="46" width="18" height="4" rx="2" fill="var(--neutral-600, #9ca3af)" opacity="0.6" />
        <rect x="22" y="54" width="30" height="3" rx="1.5" fill="var(--neutral-500, #d1d5db)" opacity="0.4" />
      </g>

      <!-- Sprints: 火山/冲刺 -->
      <g v-else-if="scenario === 'sprints'">
        <circle cx="48" cy="56" r="28" fill="var(--neutral-200, #e5e7eb)" />
        <path d="M32 56l8-16 8 12 6-8 6 8 8-12v24H32z" fill="var(--brand-default, #3b82f6)" opacity="0.3" />
        <circle cx="68" cy="28" r="12" fill="var(--warning-300, #fcd34d)" opacity="0.4" />
        <path d="M65 28l3 3 6-6" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
      </g>

      <!-- Modules: 积木/模块 -->
      <g v-else-if="scenario === 'modules'">
        <rect x="16" y="40" width="24" height="24" rx="4" fill="var(--brand-200, #bfdbfe)" />
        <rect x="44" y="40" width="24" height="24" rx="4" fill="var(--success-200, #bbf7d0)" />
        <rect x="30" y="20" width="24" height="24" rx="4" fill="var(--warning-200, #fde68a)" />
        <rect x="44" y="56" width="36" height="24" rx="4" fill="var(--neutral-200, #e5e7eb)" opacity="0.6" />
      </g>

      <!-- Search: 放大镜 -->
      <g v-else-if="scenario === 'search'">
        <circle cx="44" cy="44" r="20" stroke="var(--neutral-400, #9ca3af)" stroke-width="4" fill="none" />
        <path d="M58 58l16 16" stroke="var(--neutral-400, #9ca3af)" stroke-width="4" stroke-linecap="round" />
        <circle cx="44" cy="44" r="8" fill="var(--neutral-200, #e5e7eb)" />
      </g>

      <!-- Intake/Inbox: 收件箱 -->
      <g v-else-if="scenario === 'intake' || scenario === 'inbox'">
        <rect x="16" y="32" width="64" height="40" rx="6" fill="var(--neutral-200, #e5e7eb)" />
        <path d="M16 32l32 20 32-20" stroke="var(--neutral-400, #9ca3af)" stroke-width="2.5" fill="none" />
        <path d="M32 20h12l8 12H24z" fill="var(--brand-200, #bfdbfe)" />
      </g>

      <!-- Notifications: 铃铛 -->
      <g v-else-if="scenario === 'notifications'">
        <path d="M48 16c-12 0-20 8-20 22v8l-6 8h52l-6-8v-8c0-14-8-22-20-22z" fill="var(--neutral-200, #e5e7eb)" />
        <rect x="44" y="10" width="8" height="12" rx="4" fill="var(--brand-default, #3b82f6)" opacity="0.6" />
        <ellipse cx="48" cy="76" rx="10" ry="4" fill="var(--neutral-300, #d1d5db)" />
      </g>

      <!-- Labels: 标签 -->
      <g v-else-if="scenario === 'labels'">
        <path d="M20 28l20-12 32 12v32l-32 16-20-16z" fill="var(--neutral-200, #e5e7eb)" />
        <circle cx="36" cy="44" r="6" fill="var(--brand-default, #3b82f6)" opacity="0.5" />
      </g>

      <!-- Members: 人群 -->
      <g v-else-if="scenario === 'members'">
        <circle cx="36" cy="32" r="10" fill="var(--brand-200, #bfdbfe)" />
        <circle cx="60" cy="32" r="10" fill="var(--success-200, #bbf7d0)" />
        <circle cx="48" cy="48" r="10" fill="var(--warning-200, #fde68a)" />
        <path d="M20 72c0-10 8-18 18-18s18 8 18 18" stroke="var(--neutral-300, #d1d5db)" stroke-width="3" fill="none" />
        <path d="M52 72c0-10 8-18 18-18s18 8 18 18" stroke="var(--neutral-300, #d1d5db)" stroke-width="3" fill="none" />
      </g>

      <!-- Analytics: 图表 -->
      <g v-else-if="scenario === 'analytics'">
        <rect x="20" y="56" width="10" height="20" rx="2" fill="var(--brand-default, #3b82f6)" opacity="0.6" />
        <rect x="36" y="44" width="10" height="32" rx="2" fill="var(--brand-default, #3b82f6)" opacity="0.4" />
        <rect x="52" y="34" width="10" height="42" rx="2" fill="var(--brand-default, #3b82f6)" opacity="0.7" />
        <rect x="68" y="48" width="10" height="28" rx="2" fill="var(--brand-default, #3b82f6)" opacity="0.5" />
      </g>

      <!-- Views: 视图/网格 -->
      <g v-else-if="scenario === 'views'">
        <rect x="16" y="20" width="28" height="24" rx="4" fill="var(--neutral-200, #e5e7eb)" />
        <rect x="52" y="20" width="28" height="24" rx="4" fill="var(--neutral-200, #e5e7eb)" />
        <rect x="16" y="52" width="28" height="24" rx="4" fill="var(--brand-200, #bfdbfe)" />
        <rect x="52" y="52" width="28" height="24" rx="4" fill="var(--neutral-200, #e5e7eb)" />
      </g>

      <!-- API Token: 钥匙 -->
      <g v-else-if="scenario === 'api-token'">
        <circle cx="36" cy="44" r="14" stroke="var(--neutral-400, #9ca3af)" stroke-width="4" fill="none" />
        <path d="M46 54l30 30" stroke="var(--neutral-400, #9ca3af)" stroke-width="4" stroke-linecap="round" />
        <path d="M70 70l-6 6M76 76l-6 6" stroke="var(--neutral-400, #9ca3af)" stroke-width="3" stroke-linecap="round" />
      </g>

      <!-- Webhooks: 钩子/链接 -->
      <g v-else-if="scenario === 'webhooks'">
        <rect x="16" y="36" width="28" height="24" rx="4" fill="var(--neutral-200, #e5e7eb)" />
        <rect x="52" y="36" width="28" height="24" rx="4" fill="var(--neutral-200, #e5e7eb)" />
        <path d="M44 48h8" stroke="var(--brand-default, #3b82f6)" stroke-width="3" stroke-linecap="round" stroke-dasharray="2 3" />
        <circle cx="44" cy="48" r="4" fill="var(--brand-default, #3b82f6)" opacity="0.4" />
        <circle cx="52" cy="48" r="4" fill="var(--brand-default, #3b82f6)" opacity="0.4" />
      </g>

      <!-- Default fallback: simple box -->
      <g v-else>
        <circle cx="48" cy="48" r="32" fill="var(--neutral-200, #e5e7eb)" />
        <rect x="32" y="40" width="32" height="4" rx="2" fill="var(--neutral-500, #9ca3af)" opacity="0.5" />
        <rect x="36" y="48" width="24" height="3" rx="1.5" fill="var(--neutral-400, #d1d5db)" opacity="0.4" />
      </g>
    </svg>

    <!-- Error 插画 -->
    <svg
      v-else-if="scenario === 'error'"
      class="app-empty__illustration"
      :width="illustrationPx"
      :height="illustrationPx"
      viewBox="0 0 96 96"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <circle cx="48" cy="48" r="40" fill="var(--danger-50, #fef2f2)" />
      <path d="M48 28v24" stroke="var(--danger-500, #ef4444)" stroke-width="5" stroke-linecap="round" />
      <circle cx="48" cy="66" r="3" fill="var(--danger-500, #ef4444)" />
    </svg>

    <!-- 标题与描述 -->
    <h3 class="app-empty__title">{{ resolvedTitle }}</h3>
    <p v-if="resolvedDescription" class="app-empty__desc">{{ resolvedDescription }}</p>

    <!-- 默认插槽（自定义内容） -->
    <div v-if="$slots.default" class="app-empty__content">
      <slot />
    </div>

    <!-- CTA 按钮 -->
    <div class="app-empty__actions">
      <button
        v-if="$slots.cta"
        class="app-empty__cta"
        @click="emit('cta-click')"
      >
        <slot name="cta" />
      </button>
      <button
        v-if="$slots.secondary"
        class="app-empty__cta app-empty__cta--secondary"
        @click="emit('secondary-click')"
      >
        <slot name="secondary" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.app-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
}

.app-empty--sm { padding: 32px 16px; }
.app-empty--md { padding: 48px 24px; }
.app-empty--lg { padding: 64px 32px; }

.app-empty__emoji {
  font-size: 48px;
  margin-bottom: 12px;
}

.app-empty__illustration {
  margin-bottom: 16px;
}

.app-empty--sm .app-empty__illustration { width: 64px; height: 64px; }
.app-empty--md .app-empty__illustration { width: 96px; height: 96px; }
.app-empty--lg .app-empty__illustration { width: 128px; height: 128px; }

.app-empty__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--txt-primary, var(--text-primary, #1f2937));
  margin: 0 0 8px;
}

.app-empty__desc {
  font-size: 13px;
  color: var(--txt-tertiary, var(--text-tertiary, #9ca3af));
  margin: 0 0 16px;
  line-height: 1.5;
  max-width: 320px;
}

.app-empty__content {
  margin-bottom: 12px;
}

.app-empty__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.app-empty__cta {
  padding: 6px 14px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--bg-surface-1, var(--surface-1, #fff));
  color: var(--txt-primary);
  transition: all 0.15s;
  font-family: inherit;
}

.app-empty__cta:hover {
  background: var(--bg-surface-2, var(--surface-2, #f9fafb));
}

.app-empty__cta--secondary {
  background: transparent;
  border-color: transparent;
  color: var(--txt-secondary, var(--text-secondary, #6b7280));
}

.app-empty__cta--secondary:hover {
  background: var(--bg-surface-2);
}
</style>
