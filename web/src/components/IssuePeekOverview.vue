<script setup lang="ts">
/**
 * IssuePeekOverview — 右侧抽屉式工作项预览。
 *
 * 特性：
 *  - 滑入/滑出动画（translateX + backdrop fade）
 *  - ESC 关闭、点击遮罩关闭
 *  - 加载完整工作项详情（描述、状态、优先级等）
 *  - 显示评论预览（最多5条）
 *  - 点击「打开详情」跳转 IssueDetailView
 *  - 使用设计令牌（tokens.css）保持视觉一致
 */

import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter } from "vue-router";

import { issueApi, type Issue, type State } from "@/api/services/issue";
import { workspaceApi } from "@/api/services/workspace";
import { usePeekStore } from "@/stores/peek";
import { AppButton } from "@/components";
import RichTextEditor from "@/components/RichTextEditor.vue";
import CommentList from "@/components/CommentList.vue";

const router = useRouter();
const peek = usePeekStore();

// ---- 数据状态 ----
const loading = ref(true);
const error = ref("");
const issue = ref<Issue | null>(null);
const states = ref<State[]>([]);
const wsId = ref<number | null>(null);

// ---- 派生 ----
const projectId = computed(() => peek.target?.projectId ?? 0);
const issueId = computed(() => peek.target?.issueId ?? 0);

const isOpen = computed(() => peek.visible);

function stateName(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.name ?? "未知";
}

function stateColor(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.color ?? "#8DA2C2";
}

function typeLabel(type: string): string {
  return ({ requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[type] ?? type;
}

function priorityLabel(p: string): string {
  return ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[p] ?? p;
}

function severityText(s?: number | null): string {
  if (s == null) return "-";
  return `S${s}`;
}

// ---- 方法 ----
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
watch(() => peek.target, (t) => {
  if (t) {
    loadIssue();
  } else {
    issue.value = null;
    states.value = [];
    error.value = "";
  }
}, { immediate: true });

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
              <span
                class="peek__badge peek__badge--state"
                :style="{ backgroundColor: stateColor(issue.state_id) }"
              >
                {{ stateName(issue.state_id) }}
              </span>
            </div>
          </header>

          <h2 class="peek__title">{{ issue.name }}</h2>

          <!-- 字段摘要 -->
          <div class="peek__fields">
            <div class="peek__field">
              <span class="peek__field-label">优先级</span>
              <span class="peek__field-value pri-badge" :class="`pri-${issue.priority}`">
                {{ priorityLabel(issue.priority) }}
              </span>
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
                <router-link :to="`/${wsId}/projects/${projectId}/sprints/${issue.sprint_id}`" class="link" @click.stop>
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

          <!-- 描述 -->
          <div class="peek__section">
            <h4 class="peek__section-title">描述</h4>
            <div v-if="issue.description_html" class="peek__desc">
              <RichTextEditor
                :content-html="issue.description_html"
                :editable="false"
              />
            </div>
            <p v-else class="text-muted">暂无描述</p>
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
  background: rgba(0, 0, 0, 0.2);
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
  border-left: 1px solid var(--border-default, #e5e7eb);
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.08);
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
  background: var(--bg-surface-3, var(--surface-3, #e5e7eb));
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
  border-top-color: var(--brand-500, #3b82f6);
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
}

/* ===== Title ===== */
.peek__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--txt-primary, var(--text-primary, #1f2937));
  margin: 0 20px 16px;
  line-height: 1.35;
  padding-right: 40px;
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

.pri-badge.pri-urgent { color: var(--danger-500, #ef4444); font-weight: 600; }
.pri-badge.pri-high { color: var(--warning-500, #f59e0b); }
.pri-badge.pri-medium { color: var(--brand-500, #3b82f6); }
.pri-badge.pri-low, .pri-badge.pri-none { color: var(--txt-tertiary); }

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

/* ===== Section ===== */
.peek__section {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle, #f3f4f6);
}

.peek__section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--txt-primary, var(--text-primary));
  margin: 0 0 10px;
}

.peek__desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--txt-secondary, var(--text-secondary));
}

.text-muted {
  color: var(--txt-tertiary, var(--text-tertiary, #9ca3af));
  font-size: 13px;
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
</style>
