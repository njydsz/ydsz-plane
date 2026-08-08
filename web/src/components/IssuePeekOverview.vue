<script setup lang="ts">
/**
 * IssuePeekOverview — 右侧抽屉式工作项预览（增强版，支持行内编辑）。
 *
 * v2 增强（参考 Plane inline edit 交互）：
 *  - 标题行内编辑（InlineEdit）
 *  - 状态行内选择（InlineSelectEdit，有颜色指示）
 *  - 优先级行内选择（InlineSelectEdit）
 *  - 描述/富文本可切换编辑态
 *  - 使用 promiseToast 提供操作反馈
 *
 * 原特性保留：
 *  - 滑入/滑出动画（translateX + backdrop fade）
 *  - ESC 关闭、点击遮罩关闭
 *  - 加载完整工作项详情（描述、状态、优先级等）
 *  - 显示评论预览（最多5条）
 *  - 点击「打开详情」跳转 IssueDetailView
 */

import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter } from "vue-router";

import { issueApi, type Issue, type State, type IssuePriority } from "@/api/services/issue";
import { workspaceApi } from "@/api/services/workspace";
import { usePeekStore } from "@/stores/peek";
import { useAuthStore } from "@/stores/auth";
import { AppButton, IssueSocialBar } from "@/components";
import RichTextEditor from "@/components/RichTextEditor.vue";
import CommentList from "@/components/CommentList.vue";
import InlineEdit from "@/components/InlineEdit.vue";
import InlineSelectEdit, { type SelectOption } from "@/components/InlineSelectEdit.vue";
import { promiseToast } from "@/lib/toast";

const router = useRouter();
const peek = usePeekStore();
const auth = useAuthStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

// ---- 数据状态 ----
const loading = ref(true);
const error = ref("");
const issue = ref<Issue | null>(null);
const states = ref<State[]>([]);
const wsId = ref<number | null>(null);

// ---- 行内编辑态 ----
const descEditing = ref(false);

// ---- 派生 ----
const projectId = computed(() => peek.target?.projectId ?? 0);
const issueId = computed(() => peek.target?.issueId ?? 0);

const isOpen = computed(() => peek.visible);

// ---- 选项列表（供 InlineSelectEdit 使用） ----
const stateOptions = computed<SelectOption<number>[]>(() =>
  states.value
    .slice()
    .sort((a, b) => a.sequence - b.sequence)
    .map((s) => ({ value: s.id, label: s.name, color: s.color, title: `${s.group}` }))
);

const priorityOptions: SelectOption<IssuePriority>[] = [
  { value: "urgent", label: "紧急", color: "var(--priority-urgent, #e11d48)" },
  { value: "high", label: "高", color: "var(--priority-high, #f59e0b)" },
  { value: "medium", label: "中", color: "var(--priority-medium, #3b82f6)" },
  { value: "low", label: "低", color: "var(--priority-low, #6366f1)" },
  { value: "none", label: "无", color: "var(--priority-none, #8da2c2)" },
];

function typeLabel(type: string): string {
  return ({ requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[type] ?? type;
}

function stateName(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.name ?? "未知";
}

function stateColor(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.color ?? "#8DA2C2";
}

function priorityLabel(p: string): string {
  return ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[p] ?? p;
}

function severityText(s?: number | null): string {
  if (s == null) return "-";
  return `S${s}`;
}

// ---- 数据加载 ----
async function loadIssue() {
  if (!wsId.value || !projectId.value || !issueId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const ws = await workspaceApi.get(wsId.value);
    wsId.value = ws.id;
    const [iss, st] = await Promise.all([
      issueApi.getIssue(ws.id, projectId.value, issueId.value),
      issueApi.listStates(ws.id, projectId.value),
    ]);
    issue.value = iss;
    states.value = st;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

// ---- 行内编辑：提交 handler ----
async function onTitleSubmit(newTitle: string) {
  if (!issue.value) return;
  await promiseToast(
    issueApi.updateIssue(wsId.value!, projectId.value, issueId.value, {
      name: newTitle,
      version: issue.value.version,
    }),
    {
      loading: "保存标题中...",
      success: () => { issue.value = { ...issue.value!, name: newTitle }; return "标题已更新"; },
      error: () => "保存失败，请重试",
    }
  );
}

async function onStateSubmit(stateId: unknown) {
  if (!issue.value || typeof stateId !== "number") return;
  const targetState = states.value.find((s) => s.id === stateId);
  await promiseToast(
    issueApi.updateIssue(wsId.value!, projectId.value, issueId.value, {
      state_id: stateId,
      version: issue.value.version,
    }),
    {
      loading: "切换状态中...",
      success: () => { issue.value = { ...issue.value!, state_id: stateId }; return `已移至「${targetState?.name ?? ""}」`; },
      error: () => "状态切换失败",
    }
  );
}

async function onPrioritySubmit(priority: unknown) {
  if (!issue.value) return;
  await promiseToast(
    issueApi.updateIssue(wsId.value!, projectId.value, issueId.value, {
      priority: priority as IssuePriority,
      version: issue.value.version,
    }),
    {
      loading: "更新优先级中...",
      success: () => { issue.value = { ...issue.value!, priority: priority as IssuePriority }; return "优先级已更新"; },
      error: () => "优先级更新失败",
    }
  );
}

async function onDescriptionSubmit(html: string) {
  if (!issue.value) return;
  await promiseToast(
    issueApi.updateIssue(wsId.value!, projectId.value, issueId.value, {
      description_html: html,
      version: issue.value.version,
    }),
    {
      loading: "保存描述中...",
      success: () => { issue.value = { ...issue.value!, description_html: html }; return "描述已保存"; },
      error: () => "描述保存失败",
    }
  );
  descEditing.value = false;
}

function cancelDescriptionEdit() {
  descEditing.value = false;
}

// ---- 导航 ----
function openDetail() {
  router.push(`/${wsId.value}/projects/${projectId.value}/issues/${issueId.value}`);
  peek.close();
}

function close() {
  peek.close();
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && isOpen.value) {
    close();
  }
}

// ---- 副作用 ----
watch(
  () => peek.target,
  (t) => {
    if (t) {
      // 重置编辑态
      descEditing.value = false;
      loadIssue();
    } else {
      issue.value = null;
      states.value = [];
      error.value = "";
    }
  },
  { immediate: true },
);

onMounted(() => {
  document.addEventListener("keydown", onKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", onKeydown);
});
</script>

<template>
  <Teleport to="body">
    <Transition name="peek-backdrop">
      <div v-if="isOpen" class="peek-backdrop" @click="close" />
    </Transition>

    <Transition name="peek-slide">
      <aside v-if="isOpen" class="peek" role="dialog" aria-label="工作项预览">
        <!-- 关闭按钮 -->
        <button class="peek__close" title="关闭 (Escape)" @click="close">
          ✕
        </button>

        <!-- 加载态 -->
        <div v-if="loading" class="peek__loading">
          <div class="peek__spinner" />
          <p>加载中...</p>
        </div>

        <!-- 错误态 -->
        <div v-else-if="error" class="peek__error">
          <p>{{ error }}</p>
          <AppButton size="sm" variant="secondary" @click="loadIssue">重试</AppButton>
        </div>

        <!-- 内容 -->
        <template v-else-if="issue">
          <!-- 头信息 -->
          <header class="peek__header">
            <span class="peek__identifier">{{ issue.identifier }}</span>
            <div class="peek__badges">
              <span class="peek__badge" :class="`badge-${issue.type_code}`">
                {{ typeLabel(issue.type_code) }}
              </span>
              <!-- 状态行内选择 -->
              <InlineSelectEdit
                :model-value="issue.state_id"
                :options="stateOptions"
                placeholder="选择状态"
                @submit="onStateSubmit"
              >
                <template #trigger>
                  <span
                    class="peek__badge peek__badge--state"
                    :style="{ backgroundColor: stateColor(issue.state_id) }"
                  >
                    {{ stateName(issue.state_id) }}
                  </span>
                </template>
              </InlineSelectEdit>
            </div>
          </header>

          <!-- 标题（行内编辑） -->
          <InlineEdit
            :model-value="issue.name"
            placeholder="输入工作项标题"
            :max-length="200"
            class="peek__title"
            @submit="onTitleSubmit"
          />

          <!-- 字段摘要 -->
          <div class="peek__fields">
            <!-- 优先级（行内选择） -->
            <div class="peek__field">
              <span class="peek__field-label">优先级</span>
              <InlineSelectEdit
                :model-value="issue.priority"
                :options="priorityOptions"
                placeholder="选择优先级"
                @submit="onPrioritySubmit"
              >
                <template #trigger>
                  <span class="peek__field-value pri-badge" :class="`pri-${issue.priority}`">
                    {{ priorityLabel(issue.priority) }}
                  </span>
                </template>
              </InlineSelectEdit>
            </div>

            <div v-if="issue.severity" class="peek__field">
              <span class="peek__field-label">严重度</span>
              <span class="peek__field-value">{{ severityText(issue.severity) }}</span>
            </div>
            <div v-if="issue.point != null" class="peek__field">
              <span class="peek__field-label">点数</span>
              <span class="peek__field-value">{{ issue.point }}pt</span>
            </div>
            <div v-if="issue.sprint_id" class="peek__field">
              <span class="peek__field-label">迭代</span>
              <span class="peek__field-value">
                <router-link
                  :to="`/${wsId}/projects/${projectId}/sprints/${issue.sprint_id}`"
                  class="link"
                  @click.stop
                >
                  #{{ issue.sprint_id }}
                </router-link>
              </span>
            </div>
            <div class="peek__field">
              <span class="peek__field-label">指派人</span>
              <span class="peek__field-value">
                <span v-if="issue.assignees?.length">
                  <span v-for="uid in issue.assignees.slice(0, 4)" :key="uid" class="peek__avatar">
                    {{ uid }}
                  </span>
                  <span v-if="issue.assignees.length > 4" class="peek__avatar peek__avatar--more">
                    +{{ issue.assignees.length - 4 }}
                  </span>
                </span>
                <span v-else class="text-muted">未指派</span>
              </span>
            </div>
            <div class="peek__field">
              <span class="peek__field-label">更新时间</span>
              <span class="peek__field-value">{{ new Date(issue.updated_at).toLocaleDateString("zh-CN") }}</span>
            </div>
          </div>

          <!-- 社交反馈：投票 / 表情反应 / 关注 -->
          <div v-if="wsId" class="peek__section peek__section--social">
            <IssueSocialBar
              :workspace-id="wsId"
              :project-id="projectId"
              :issue-id="issueId"
              :initial-watching="issue.watchers.includes(currentUserId)"
            />
          </div>

          <!-- 描述（可切换编辑态） -->
          <div class="peek__section">
            <div class="peek__section-header">
              <h4 class="peek__section-title">描述</h4>
              <button
                v-if="!descEditing && issue.description_html"
                class="peek__edit-btn"
                title="编辑描述"
                @click="descEditing = true"
              >
                编辑
              </button>
            </div>
            <template v-if="descEditing">
              <RichTextEditor
                :content-html="issue.description_html ?? ''"
                :editable="true"
                :min-height="'160px'"
                @update:content-html="onDescriptionSubmit"
              />
              <div class="peek__edit-actions">
                <AppButton size="sm" variant="primary" @click="onDescriptionSubmit(issue.description_html ?? '')">
                  保存
                </AppButton>
                <AppButton size="sm" variant="secondary" @click="cancelDescriptionEdit">
                  取消
                </AppButton>
              </div>
            </template>
            <template v-else>
              <div v-if="issue.description_html" class="peek__desc">
                <RichTextEditor :content-html="issue.description_html" :editable="false" />
              </div>
              <p v-else class="text-muted">
                暂无描述
                <button class="peek__add-btn" @click="descEditing = true">+ 添加描述</button>
              </p>
            </template>
          </div>

          <!-- 缺陷信息 -->
          <div v-if="issue.type_code === 'defect'" class="peek__section">
            <h4 class="peek__section-title">缺陷信息</h4>
            <div class="peek__fields">
              <div v-if="issue.found_phase" class="peek__field">
                <span class="peek__field-label">发现阶段</span>
                <span class="peek__field-value">{{ issue.found_phase }}</span>
              </div>
              <div v-if="issue.root_cause_category" class="peek__field">
                <span class="peek__field-label">根因分类</span>
                <span class="peek__field-value">{{ issue.root_cause_category }}</span>
              </div>
            </div>
          </div>

          <!-- 评论预览 -->
          <div v-if="wsId" class="peek__section">
            <h4 class="peek__section-title">评论</h4>
            <CommentList
              :workspace-id="wsId"
              :project-id="projectId"
              :issue-id="issueId"
              :limit="5"
              compact
            />
          </div>

          <!-- 打开详情 -->
          <footer class="peek__footer">
            <AppButton variant="primary" size="sm" block @click="openDetail">
              打开详情 →
            </AppButton>
          </footer>
        </template>
      </aside>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* ===== Backdrop ===== */
.peek-backdrop {
  position: fixed;
  inset: 0;
  background: var(--bg-backdrop, oklch(0 0 0 / 30%));
  z-index: 900;
}

.peek-backdrop-enter-active,
.peek-backdrop-leave-active {
  transition: opacity 0.25s ease;
}
.peek-backdrop-enter-from,
.peek-backdrop-leave-to {
  opacity: 0;
}

/* ===== Drawer ===== */
.peek {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 480px;
  max-width: 90vw;
  background: var(--bg-surface-1, var(--surface-1, #fff));
  border-left: 1px solid var(--border-subtle, #e5e7eb);
  box-shadow: var(--shadow-overlay-100);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.peek-slide-enter-active {
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.peek-slide-leave-active {
  transition: transform 0.2s ease-in;
}
.peek-slide-enter-from,
.peek-slide-leave-to {
  transform: translateX(100%);
}

/* ===== Close button ===== */
.peek__close {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: var(--bg-surface-2, var(--surface-2, #f3f4f6));
  color: var(--txt-secondary, var(--text-secondary, #6b7280));
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
  z-index: 1;
}

.peek__close:hover {
  background: var(--bg-layer-3, var(--surface-3, #e5e7eb));
  color: var(--txt-primary, var(--text-primary, #1f2937));
}

/* ===== Loading / Error ===== */
.peek__loading,
.peek__error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 24px;
  gap: 12px;
  color: var(--txt-tertiary, var(--text-tertiary, #9ca3af));
}

.peek__spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-default, #e5e7eb);
  border-top-color: var(--brand-default, var(--brand-500, #3b82f6));
  border-radius: 50%;
  animation: peek-spin 0.8s linear infinite;
}

@keyframes peek-spin {
  to { transform: rotate(360deg); }
}

/* ===== Header ===== */
.peek__header {
  padding: 16px 20px 8px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.peek__identifier {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-tertiary, var(--text-tertiary, #9ca3af));
}

.peek__badges {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.peek__badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm, 4px);
  font-weight: 500;
}

.peek__badge.badge-requirement {
  background: var(--brand-50, #eff6ff);
  color: var(--brand-600, #2563eb);
}
.peek__badge.badge-task {
  background: var(--success-50, #ecfdf5);
  color: var(--success-600, #059669);
}
.peek__badge.badge-defect {
  background: var(--danger-50, #fef2f2);
  color: var(--danger-600, #dc2626);
}
.peek__badge--state {
  color: #fff;
  cursor: pointer;
}

/* ===== Title (inline edit) ===== */
.peek__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--txt-primary, var(--text-primary, #1f2937));
  margin: 0 20px 16px;
  padding-right: 40px;
  line-height: 1.35;
  display: block;
}

/* ===== Fields ===== */
.peek__fields {
  padding: 0 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  border-bottom: 1px solid var(--border-subtle, #f3f4f6);
}

.peek__field {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}

.peek__field-label {
  color: var(--txt-tertiary, var(--text-tertiary, #9ca3af));
  min-width: 60px;
  flex-shrink: 0;
}

.peek__field-value {
  color: var(--txt-primary, var(--text-primary, #1f2937));
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.pri-badge.pri-urgent { color: var(--priority-urgent, var(--danger-500, #ef4444)); font-weight: 600; }
.pri-badge.pri-high { color: var(--priority-high, var(--warning-500, #f59e0b)); }
.pri-badge.pri-medium { color: var(--priority-medium, var(--brand-500, #3b82f6)); }
.pri-badge.pri-low, .pri-badge.pri-none { color: var(--txt-tertiary, var(--priority-none, #8da2c2)); }

.peek__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--brand-100, #dbeafe);
  color: var(--brand-600, #2563eb);
  font-size: 10px;
  font-weight: 600;
  border: 2px solid var(--surface-1, #fff);
  margin-left: -4px;
}
.peek__avatar:first-child { margin-left: 0; }
.peek__avatar--more {
  background: var(--surface-3, #e5e7eb);
  color: var(--txt-secondary);
}

/* ===== Section (description) ===== */
.peek__section {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle, #f3f4f6);
}

/* 社交栏区块：无标题、紧凑 */
.peek__section--social {
  padding: 12px 20px;
}

.peek__section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.peek__section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--txt-primary, var(--text-primary));
  margin: 0;
}

.peek__edit-btn {
  font-size: 11px;
  padding: 2px 8px;
  border: 1px solid var(--border-subtle-1, #e5e7eb);
  border-radius: var(--radius-sm, 4px);
  background: transparent;
  color: var(--txt-secondary, #6b7280);
  cursor: pointer;
  transition: background 0.1s, border-color 0.1s;
}
.peek__edit-btn:hover {
  background: var(--bg-layer-1-hover, #f3f4f6);
  border-color: var(--border-strong);
}

.peek__desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--txt-secondary, var(--text-secondary));
}

.peek__edit-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  justify-content: flex-end;
}

.peek__add-btn {
  border: none;
  background: none;
  color: var(--txt-accent-primary, var(--brand-500, #3b82f6));
  font-size: 13px;
  cursor: pointer;
  padding: 0;
  margin-left: 4px;
}
.peek__add-btn:hover {
  text-decoration: underline;
}

.text-muted {
  color: var(--txt-placeholder, var(--txt-tertiary, #9ca3af));
  font-size: 13px;
  margin: 0;
}

.link {
  color: var(--brand-500, #3b82f6);
  text-decoration: none;
}
.link:hover {
  text-decoration: underline;
}

/* ===== Footer ===== */
.peek__footer {
  padding: 16px 20px;
  margin-top: auto;
}

/* ===== Inline select 覆写 ===== */
:deep(.inline-select__trigger--editable:hover) {
  background: var(--bg-layer-1-hover, rgba(0, 0, 0, 0.04));
}
:deep(.inline-select__pop) {
  z-index: 1100; /* above peek backdrop */
}
</style>
