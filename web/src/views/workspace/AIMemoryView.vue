<script setup lang="ts">
/**
 * AIMemoryView — 工作空间级 AI 记忆管理。
 *
 * 功能：
 *  - 按作用域（用户 / 项目 / 工作空间）筛选；
 *  - 记忆列表（key / content / source / 更新时间）；
 *  - 新增 / 编辑 / 删除记忆；
 *  - 空状态。
 */
import { computed, onMounted, reactive, ref } from "vue";

import {
  aiAgentApi,
  type AIWorkspaceMemory,
  type AIWorkspaceMemoryInput,
  type AIMemoryScope,
} from "@/api/services/aiagent";
import { workspaceApi, type Project } from "@/api/services/workspace";
import { AppButton, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer } from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useI18n } from "@/composables/useI18n";
import { formatRelativeTime } from "@/lib/formatTime";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const props = defineProps<{
  workspaceId: number;
}>();

const { t } = useI18n();

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const memories = ref<AIWorkspaceMemory[]>([]);
const projects = ref<Project[]>([]);
const loading = ref(true);
const error = ref("");

const scopeFilter = ref<AIMemoryScope | "">("");

const filteredMemories = computed(() => {
  if (!scopeFilter.value) return memories.value;
  return memories.value.filter((m) => m.scope === scopeFilter.value);
});

async function load() {
  if (!props.workspaceId) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    memories.value = await aiAgentApi.listMemory(props.workspaceId);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t("aiMemory.loadFailed");
  } finally {
    loading.value = false;
  }
}

async function loadProjects() {
  try {
    projects.value = await workspaceApi.listProjects(props.workspaceId);
  } catch {
    projects.value = [];
  }
}

function projectName(id: number): string {
  return projects.value.find((p) => p.id === id)?.name ?? `#${id}`;
}

/* ------------------------------------------------------------------ */
/* 新增 / 编辑                                                         */
/* ------------------------------------------------------------------ */

interface MemoryForm {
  scope: AIMemoryScope | "";
  scope_id: number | null;
  key: string;
  content: string;
  source: string;
}

function emptyForm(): MemoryForm {
  return {
    scope: "",
    scope_id: null,
    key: "",
    content: "",
    source: "manual",
  };
}

const showForm = ref(false);
const editingId = ref<number | null>(null);
const saving = ref(false);
const formError = ref("");
const form = reactive<MemoryForm>(emptyForm());

const isEditing = computed(() => editingId.value !== null);

function openCreate() {
  editingId.value = null;
  Object.assign(form, emptyForm());
  formError.value = "";
  showForm.value = true;
}

function openEdit(memory: AIWorkspaceMemory) {
  editingId.value = memory.id;
  Object.assign(form, {
    scope: memory.scope,
    scope_id: memory.scope_id || null,
    key: memory.key ?? "",
    content: memory.content,
    source: memory.source ?? "manual",
  });
  formError.value = "";
  showForm.value = true;
}

async function submitForm() {
  if (!form.scope) {
    formError.value = t("aiMemory.scopeRequired");
    return;
  }
  if (!form.content.trim()) {
    formError.value = t("aiMemory.contentRequired");
    return;
  }
  saving.value = true;
  formError.value = "";
  const payload: AIWorkspaceMemoryInput = {
    scope: form.scope,
    scope_id: form.scope_id ?? 0,
    key: form.key.trim(),
    content: form.content.trim(),
    source: form.source.trim() || "manual",
  };
  try {
    // 记忆无独立「编辑」接口：后端按 scope+scope_id+key 幂等 upsert，
    // 因此「编辑」通过 setMemory 覆盖实现；若键/作用域变化则新建一条。
    await aiAgentApi.setMemory(props.workspaceId, payload);
    toast.success(t("aiMemory.saveSuccess"));
    showForm.value = false;
    editingId.value = null;
    await load();
  } catch (err: unknown) {
    formError.value = err instanceof Error ? err.message : t("aiMemory.saveFailed");
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 删除                                                                */
/* ------------------------------------------------------------------ */

const deletingId = ref<number | null>(null);

async function removeMemory(memory: AIWorkspaceMemory) {
  if (!await confirm({ title: "删除记忆", message: t("aiMemory.deleteConfirm"), confirmText: "删除", variant: "danger" })) return;
  deletingId.value = memory.id;
  try {
    await aiAgentApi.deleteMemory(props.workspaceId, memory.id);
    toast.success(t("aiMemory.deleteSuccess"));
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("aiMemory.saveFailed"));
  } finally {
    deletingId.value = null;
  }
}

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function scopeLabel(scope: string): string {
  const map: Record<string, string> = {
    user: t("aiMemory.scopeUser"),
    project: t("aiMemory.scopeProject"),
    workspace: t("aiMemory.scopeWorkspace"),
  };
  return map[scope] ?? scope;
}

function scopeVariant(scope: string): "brand" | "success" | "info" {
  switch (scope) {
    case "user":
      return "brand";
    case "project":
      return "success";
    default:
      return "info";
  }
}

onMounted(async () => {
  await Promise.all([load(), loadProjects()]);
});
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="ai-memory">
    <!-- 头部 -->
    <header class="ai-memory__header">
      <div>
        <h1 class="ai-memory__title">
          <DynamicIcon name="Brain" :size="22" />
          {{ t("aiMemory.title") }}
        </h1>
        <p class="ai-memory__subtitle">{{ t("aiMemory.subtitle") }}</p>
      </div>
      <AppButton size="sm" @click="openCreate">
        <DynamicIcon name="CirclePlus" :size="16" />
        {{ t("aiMemory.add") }}
      </AppButton>
    </header>

    <!-- 作用域筛选 -->
    <div class="ai-memory__filters">
      <button
        class="ai-memory__chip"
        :class="{ 'ai-memory__chip--active': scopeFilter === '' }"
        @click="scopeFilter = ''"
      >
        {{ t("aiMemory.allScopes") }}
      </button>
      <button
        class="ai-memory__chip"
        :class="{ 'ai-memory__chip--active': scopeFilter === 'user' }"
        @click="scopeFilter = 'user'"
      >
        {{ t("aiMemory.scopeUser") }}
      </button>
      <button
        class="ai-memory__chip"
        :class="{ 'ai-memory__chip--active': scopeFilter === 'project' }"
        @click="scopeFilter = 'project'"
      >
        {{ t("aiMemory.scopeProject") }}
      </button>
      <button
        class="ai-memory__chip"
        :class="{ 'ai-memory__chip--active': scopeFilter === 'workspace' }"
        @click="scopeFilter = 'workspace'"
      >
        {{ t("aiMemory.scopeWorkspace") }}
      </button>
    </div>

    <AppLoadingState v-if="loading" :text="t('common.loading')" />
    <AppErrorState
      v-else-if="error"
      :message="error"
      :retry="t('common.retry')"
      @retry="load"
    />
    <AppEmptyState
      v-else-if="filteredMemories.length === 0"
      icon="🧠"
      :title="t('aiMemory.empty')"
      :description="t('aiMemory.emptyDesc')"
      :cta-text="t('aiMemory.add')"
      @cta-click="openCreate"
    />

    <div v-else class="ai-memory__list">
      <article v-for="memory in filteredMemories" :key="memory.id" class="ai-memory-card">
        <div class="ai-memory-card__top">
          <div class="ai-memory-card__identity">
            <span class="ai-memory-card__key">{{ memory.key || "—" }}</span>
            <span
              class="ai-memory-card__scope"
              :class="`ai-memory-card__scope--${scopeVariant(memory.scope)}`"
            >
              {{ scopeLabel(memory.scope) }}
            </span>
          </div>
          <div class="ai-memory-card__ops">
            <button class="ai-memory-card__op" @click="openEdit(memory)">
              {{ t("common.edit") }}
            </button>
            <button
              class="ai-memory-card__op ai-memory-card__op--danger"
              :disabled="deletingId === memory.id"
              @click="removeMemory(memory)"
            >
              {{ deletingId === memory.id ? t("common.loading") : t("common.delete") }}
            </button>
          </div>
        </div>

        <p class="ai-memory-card__content">{{ memory.content }}</p>

        <div class="ai-memory-card__meta">
          <span v-if="memory.scope_id" class="ai-memory-card__meta-item">
            {{ t("aiMemory.scopeId") }}: {{ projectName(memory.scope_id) }}
          </span>
          <span v-if="memory.source" class="ai-memory-card__meta-item">
            {{ t("aiMemory.source") }}: {{ memory.source }}
          </span>
          <span class="ai-memory-card__meta-item">
            {{ t("aiMemory.updatedAt") }}: {{ formatRelativeTime(memory.updated_at) }}
          </span>
        </div>
      </article>
    </div>

    <!-- 新增 / 编辑弹窗 -->
    <AppModal
      :visible="showForm"
      :title="isEditing ? t('aiMemory.edit') : t('aiMemory.add')"
      width="560px"
      @close="showForm = false"
    >
      <div class="ai-memory-form">
        <div class="ai-memory-form__row">
          <label class="ai-memory-form__field">
            <span class="ai-memory-form__label">
              {{ t("aiMemory.scope") }} <span class="ai-memory-form__required">*</span>
            </span>
            <select v-model="form.scope" class="ai-memory-form__input">
              <option value="" disabled>{{ t("aiMemory.scope") }}</option>
              <option value="user">{{ t("aiMemory.scopeUser") }}</option>
              <option value="project">{{ t("aiMemory.scopeProject") }}</option>
              <option value="workspace">{{ t("aiMemory.scopeWorkspace") }}</option>
            </select>
          </label>

          <label class="ai-memory-form__field">
            <span class="ai-memory-form__label">{{ t("aiMemory.scopeId") }}</span>
            <select v-model="form.scope_id" class="ai-memory-form__input">
              <option :value="null">—</option>
              <option v-for="p in projects" :key="p.id" :value="p.id">
                {{ p.name }} ({{ p.identifier }})
              </option>
            </select>
          </label>
        </div>

        <label class="ai-memory-form__field">
          <span class="ai-memory-form__label">{{ t("aiMemory.key") }}</span>
          <input
            v-model="form.key"
            type="text"
            class="ai-memory-form__input"
            :placeholder="t('aiMemory.keyPlaceholder')"
          />
        </label>

        <label class="ai-memory-form__field">
          <span class="ai-memory-form__label">
            {{ t("aiMemory.content") }} <span class="ai-memory-form__required">*</span>
          </span>
          <textarea
            v-model="form.content"
            rows="5"
            class="ai-memory-form__input ai-memory-form__input--area"
            :placeholder="t('aiMemory.contentPlaceholder')"
          />
        </label>

        <label class="ai-memory-form__field">
          <span class="ai-memory-form__label">{{ t("aiMemory.source") }}</span>
          <input
            v-model="form.source"
            type="text"
            class="ai-memory-form__input"
            :placeholder="t('aiMemory.sourcePlaceholder')"
          />
        </label>

        <div class="ai-memory-form__hint">{{ t("aiMemory.upsertHint") }}</div>
        <div v-if="formError" class="ai-memory-form__error">{{ formError }}</div>
      </div>

      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showForm = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton size="sm" :loading="saving" @click="submitForm">
          {{ t("common.save") }}
        </AppButton>
      </template>
    </AppModal>
  
    </div>
  </AppPageContainer>
    </AppModal>
  </div>


<style scoped>
.ai-memory {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 24px;
  
  
}

.ai-memory__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.ai-memory__title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-memory__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.ai-memory__filters {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.ai-memory__chip {
  padding: 5px 12px;
  border: 1px solid var(--border-default);
  border-radius: 999px;
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}
.ai-memory__chip:hover {
  border-color: var(--brand-500);
}
.ai-memory__chip--active {
  background: var(--brand-50);
  border-color: var(--brand-500);
  color: var(--brand-600);
  font-weight: 500;
}

/* ---- 列表 ---- */
.ai-memory__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ai-memory-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--surface-1);
  transition: border-color 0.15s, box-shadow 0.15s;
}
.ai-memory-card:hover {
  border-color: var(--border-default);
  box-shadow: var(--shadow-card, 0 2px 8px rgba(0, 0, 0, 0.05));
}

.ai-memory-card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.ai-memory-card__identity {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.ai-memory-card__key {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: ui-monospace, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ai-memory-card__scope {
  flex-shrink: 0;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
}
.ai-memory-card__scope--brand {
  background: var(--brand-50);
  color: var(--brand-600);
}
.ai-memory-card__scope--success {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success-500);
}
.ai-memory-card__scope--info {
  background: rgba(23, 190, 233, 0.1);
  color: var(--info-500);
}

.ai-memory-card__ops {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.ai-memory-card__op {
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
}
.ai-memory-card__op:hover:not(:disabled) {
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.ai-memory-card__op--danger:hover:not(:disabled) {
  border-color: var(--danger-500);
  color: var(--danger-500);
}
.ai-memory-card__op:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ai-memory-card__content {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.ai-memory-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 11px;
  color: var(--text-tertiary);
}

/* ---- 表单 ---- */
.ai-memory-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.ai-memory-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.ai-memory-form__field {
  display: block;
}
.ai-memory-form__label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--text-secondary);
}
.ai-memory-form__required {
  color: var(--danger-500);
}
.ai-memory-form__input {
  width: 100%;
  height: 34px;
  padding: 0 10px;
  border: 1px solid var(--border-default);
  border-radius: 8px;
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
  box-sizing: border-box;
}
.ai-memory-form__input:focus {
  outline: none;
  border-color: var(--brand-500);
}
.ai-memory-form__input--area {
  height: auto;
  padding: 8px 10px;
  resize: vertical;
  line-height: 1.5;
}
.ai-memory-form__hint {
  font-size: 11px;
  color: var(--text-tertiary);
}
.ai-memory-form__error {
  font-size: 12px;
  color: var(--danger-500);
}

@media (max-width: 640px) {
  .ai-memory {
    padding: 16px;
  }
  .ai-memory-form__row {
    grid-template-columns: 1fr;
  }
}
</style>
