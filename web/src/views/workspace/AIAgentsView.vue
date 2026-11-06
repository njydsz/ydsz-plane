<script setup lang="ts">
/**
 * AIAgentsView — 工作空间级 AI Agent 管理。
 *
 * 功能：
 *  - Agent 卡片网格（名称 / 描述 / 模型 / 状态 / 额度 / 触发器 / 技能）；
 *  - 创建 / 编辑 Agent（指令、作用域、模型、温度、额度）；
 *  - 删除 Agent；
 *  - 点击卡片进入该 Agent 的会话面板。
 */
import { computed, onMounted, reactive, ref } from "vue";

import {
  aiAgentApi,
  type AIAgent,
  type AIAgentInput,
} from "@/api/services/aiagent";
import { workspaceApi, type Project } from "@/api/services/workspace";
import { AIAgentChat, AppBadge, AppButton, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer, ProgressBar } from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useI18n } from "@/composables/useI18n";
import { useConfirm } from "@/composables/useConfirm";
import { toast } from "@/lib/toast";

const confirm = useConfirm();
const props = defineProps<{
  workspaceId: number;
}>();

const { t } = useI18n();

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const agents = ref<AIAgent[]>([]);
const projects = ref<Project[]>([]);
const loading = ref(true);
const error = ref("");

/* 会话面板状态：null 表示未展开 */
const chatAgentId = ref<number | null>(null);
const chatKey = ref(0);

const chatAgent = computed(() => agents.value.find((a) => a.id === chatAgentId.value) ?? null);

async function load() {
  if (!props.workspaceId) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    agents.value = await aiAgentApi.listAgents(props.workspaceId);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t("aiAgent.loadFailed");
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

function openChat(agent: AIAgent) {
  chatAgentId.value = agent.id;
  chatKey.value += 1;
}

function closeChat() {
  chatAgentId.value = null;
}

function backToAgents() {
  closeChat();
}

/* ------------------------------------------------------------------ */
/* 创建 / 编辑                                                         */
/* ------------------------------------------------------------------ */

interface AgentForm {
  name: string;
  description: string;
  instructions: string;
  triggers: string[];
  skills: string[];
  scope_project_ids: number[];
  scope_work_item_types: string[];
  model: string;
  temperature: number;
  daily_credit_limit: number;
}

function emptyForm(): AgentForm {
  return {
    name: "",
    description: "",
    instructions: "",
    triggers: [],
    skills: [],
    scope_project_ids: [],
    scope_work_item_types: [],
    model: "gpt-4o-mini",
    temperature: 0.3,
    daily_credit_limit: 100,
  };
}

const showForm = ref(false);
const editingId = ref<number | null>(null);
const saving = ref(false);
const formError = ref("");
const form = reactive<AgentForm>(emptyForm());

const isEditing = computed(() => editingId.value !== null);

const ALL_TRIGGERS = ["mention", "assignment", "event", "schedule"];
const ALL_SKILLS = [
  "create_work_item",
  "update_work_item",
  "create_cycle",
  "create_module",
  "create_page",
  "update_state",
  "assign",
  "add_comment",
  "log_time",
  "bulk_update",
  "delete_work_item",
];
const ALL_WORK_ITEM_TYPES = ["requirement", "task", "defect"];
const ALL_MODELS = ["gpt-4o-mini", "gpt-4o", "gpt-4.1", "gpt-4.1-mini", "claude-3-5-sonnet"];

function triggerLabel(trigger: string): string {
  const map: Record<string, string> = {
    mention: t("aiAgent.triggerMention"),
    assignment: t("aiAgent.triggerAssignment"),
    event: t("aiAgent.triggerEvent"),
    schedule: t("aiAgent.triggerSchedule"),
  };
  return map[trigger] ?? trigger;
}

function skillLabel(skill: string): string {
  const map: Record<string, string> = {
    create_work_item: t("aiAgent.skillCreateWorkItem"),
    update_work_item: t("aiAgent.skillUpdateWorkItem"),
    create_cycle: t("aiAgent.skillCreateCycle"),
    create_module: t("aiAgent.skillCreateModule"),
    create_page: t("aiAgent.skillCreatePage"),
    update_state: t("aiAgent.skillUpdateState"),
    assign: t("aiAgent.skillAssign"),
    add_comment: t("aiAgent.skillAddComment"),
    log_time: t("aiAgent.skillLogTime"),
    bulk_update: t("aiAgent.skillBulkUpdate"),
    delete_work_item: t("aiAgent.skillDeleteWorkItem"),
  };
  return map[skill] ?? skill;
}

function openCreate() {
  editingId.value = null;
  Object.assign(form, emptyForm());
  formError.value = "";
  showForm.value = true;
}

function openEdit(agent: AIAgent) {
  editingId.value = agent.id;
  Object.assign(form, {
    name: agent.name,
    description: agent.description ?? "",
    instructions: agent.instructions ?? "",
    triggers: [...(agent.triggers ?? [])],
    skills: [...(agent.skills ?? [])],
    scope_project_ids: [...(agent.scope_project_ids ?? [])],
    scope_work_item_types: [...(agent.scope_work_item_types ?? [])],
    model: agent.model || "gpt-4o-mini",
    temperature: agent.temperature ?? 0.3,
    daily_credit_limit: agent.daily_credit_limit || 100,
  });
  formError.value = "";
  showForm.value = true;
}

async function submitForm() {
  if (!form.name.trim()) {
    formError.value = t("aiAgent.nameRequired");
    return;
  }
  saving.value = true;
  formError.value = "";
  const payload: AIAgentInput = {
    name: form.name.trim(),
    description: form.description.trim(),
    instructions: form.instructions.trim(),
    triggers: form.triggers,
    skills: form.skills,
    scope_project_ids: form.scope_project_ids,
    scope_work_item_types: form.scope_work_item_types,
    model: form.model,
    temperature: form.temperature,
    daily_credit_limit: form.daily_credit_limit,
  };
  try {
    if (isEditing.value) {
      await aiAgentApi.updateAgent(props.workspaceId, editingId.value!, payload);
      toast.success(t("aiAgent.saveSuccess"));
    } else {
      await aiAgentApi.createAgent(props.workspaceId, payload);
      toast.success(t("aiAgent.saveSuccess"));
    }
    showForm.value = false;
    editingId.value = null;
    await load();
  } catch (err: unknown) {
    formError.value = err instanceof Error ? err.message : t("aiAgent.saveFailed");
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 删除                                                                */
/* ------------------------------------------------------------------ */

const deletingId = ref<number | null>(null);

async function removeAgent(agent: AIAgent) {
  if (!await confirm({ title: "删除 Agent", message: t("aiAgent.deleteConfirm", { name: agent.name }), confirmText: "删除", variant: "danger" })) return;
  deletingId.value = agent.id;
  try {
    await aiAgentApi.deleteAgent(props.workspaceId, agent.id);
    if (chatAgentId.value === agent.id) chatAgentId.value = null;
    toast.success(t("aiAgent.deleteSuccess"));
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("aiAgent.saveFailed"));
  } finally {
    deletingId.value = null;
  }
}

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function statusVariant(status: string): "success" | "default" | "warning" {
  if (status === "active") return "success";
  if (status === "archived") return "default";
  return "warning";
}

function statusLabel(status: string): string {
  if (status === "active") return t("aiAgent.statusActive");
  if (status === "inactive") return t("aiAgent.statusInactive");
  return status;
}

function creditPercent(agent: AIAgent): number {
  const limit = agent.daily_credit_limit || 0;
  if (limit <= 0) return 0;
  return Math.min(100, Math.round((agent.used_credits_today / limit) * 100));
}

function creditColor(agent: AIAgent): string {
  const p = creditPercent(agent);
  if (p >= 90) return "var(--danger-500)";
  if (p >= 70) return "var(--warning-500)";
  return "var(--brand-500)";
}

function projectName(id: number): string {
  return projects.value.find((p) => p.id === id)?.name ?? `#${id}`;
}

onMounted(async () => {
  await Promise.all([load(), loadProjects()]);
});
</script>

<template>
  
  <AppPageContainer max="1400px">
    <div class="ai-agents">
    <!-- 头部 -->
    <header class="ai-agents__header">
      <div>
        <h1 class="ai-agents__title">
          <DynamicIcon name="Sparkles" :size="22" />
          {{ t("aiAgent.title") }}
        </h1>
        <p class="ai-agents__subtitle">{{ t("aiAgent.subtitle") }}</p>
      </div>
      <div class="ai-agents__actions">
        <AppButton v-if="chatAgentId" variant="secondary" size="sm" @click="backToAgents">
          ← {{ t("aiAgent.chat.title") }}
        </AppButton>
        <AppButton size="sm" @click="openCreate">
          <DynamicIcon name="CirclePlus" :size="16" />
          {{ t("aiAgent.create") }}
        </AppButton>
      </div>
    </header>

    <!-- 会话面板 -->
    <div v-if="chatAgent" class="ai-agents__chat">
      <div class="ai-agents__chat-head">
        <div class="ai-agents__chat-agent">
          <DynamicIcon name="Bot" :size="18" />
          <span class="ai-agents__chat-name">{{ chatAgent.name }}</span>
          <AppBadge variant="brand">{{ chatAgent.model }}</AppBadge>
        </div>
        <button class="ai-agents__chat-close" :aria-label="t('common.close')" @click="closeChat">
          <DynamicIcon name="X" :size="16" />
        </button>
      </div>
      <AIAgentChat
        :key="`${chatAgent.id}-${chatKey}`"
        :workspace-id="workspaceId"
        :agents="[chatAgent]"
        :initial-conversation-id="null"
      />
    </div>

    <!-- 列表 -->
    <template v-else>
      <AppLoadingState v-if="loading" :text="t('common.loading')" />
      <AppErrorState
        v-else-if="error"
        :message="error"
        :retry="t('common.retry')"
        @retry="load"
      />
      <AppEmptyState
        v-else-if="agents.length === 0"
        icon="🤖"
        :title="t('aiAgent.empty')"
        :description="t('aiAgent.emptyDesc')"
        :cta-text="t('aiAgent.create')"
        @cta-click="openCreate"
      />

      <div v-else class="ai-agents__grid">
        <article
          v-for="agent in agents"
          :key="agent.id"
          class="ai-agent-card"
          role="button"
          tabindex="0"
          @click="openChat(agent)"
          @keydown.enter="openChat(agent)"
        >
          <!-- 顶部：图标 + 名称 + 状态 -->
          <div class="ai-agent-card__top">
            <div class="ai-agent-card__identity">
              <div class="ai-agent-card__avatar">
                <DynamicIcon name="Bot" :size="18" />
              </div>
              <div class="ai-agent-card__meta">
                <h3 class="ai-agent-card__name">{{ agent.name }}</h3>
                <span class="ai-agent-card__model">{{ agent.model }}</span>
              </div>
            </div>
            <AppBadge :variant="statusVariant(agent.status)">
              {{ statusLabel(agent.status) }}
            </AppBadge>
          </div>

          <p v-if="agent.description" class="ai-agent-card__desc">
            {{ agent.description }}
          </p>
          <p v-else class="ai-agent-card__desc ai-agent-card__desc--empty">
            {{ t("aiAgent.noDescription") }}
          </p>

          <!-- 额度 -->
          <div class="ai-agent-card__credit">
            <div class="ai-agent-card__credit-head">
              <span class="ai-agent-card__credit-label">{{ t("aiAgent.creditUsage") }}</span>
              <span class="ai-agent-card__credit-value">
                {{ agent.used_credits_today }} / {{ agent.daily_credit_limit }}
              </span>
            </div>
            <ProgressBar
              :percent="creditPercent(agent)"
              size="sm"
              :color="creditColor(agent)"
              :show-label="false"
            />
          </div>

          <!-- 触发器 -->
          <div class="ai-agent-card__row">
            <span class="ai-agent-card__row-label">{{ t("aiAgent.triggers") }}</span>
            <div class="ai-agent-card__chips">
              <span
                v-for="tr in agent.triggers"
                :key="tr"
                class="ai-agent-chip ai-agent-chip--trigger"
              >
                {{ triggerLabel(tr) }}
              </span>
              <span v-if="!agent.triggers?.length" class="ai-agent-card__empty-chip">
                {{ t("aiAgent.noTriggers") }}
              </span>
            </div>
          </div>

          <!-- 技能 -->
          <div class="ai-agent-card__row">
            <span class="ai-agent-card__row-label">{{ t("aiAgent.skills") }}</span>
            <div class="ai-agent-card__chips">
              <span
                v-for="sk in agent.skills.slice(0, 4)"
                :key="sk"
                class="ai-agent-chip ai-agent-chip--skill"
              >
                {{ skillLabel(sk) }}
              </span>
              <span v-if="agent.skills.length > 4" class="ai-agent-chip ai-agent-chip--more">
                +{{ agent.skills.length - 4 }}
              </span>
              <span v-if="!agent.skills?.length" class="ai-agent-card__empty-chip">
                {{ t("aiAgent.noSkills") }}
              </span>
            </div>
          </div>

          <!-- 作用域 -->
          <div v-if="agent.scope_project_ids?.length" class="ai-agent-card__row">
            <span class="ai-agent-card__row-label">{{ t("aiAgent.scopeProjectIds") }}</span>
            <div class="ai-agent-card__chips">
              <span
                v-for="pid in agent.scope_project_ids"
                :key="pid"
                class="ai-agent-chip ai-agent-chip--scope"
              >
                {{ projectName(pid) }}
              </span>
            </div>
          </div>

          <!-- 操作 -->
          <div class="ai-agent-card__footer" @click.stop>
            <button class="ai-agent-card__action ai-agent-card__action--primary" @click="openChat(agent)">
              <DynamicIcon name="MessageSquareText" :size="14" />
              {{ t("aiAgent.chatAction") }}
            </button>
            <button class="ai-agent-card__action" @click="openEdit(agent)">
              {{ t("common.edit") }}
            </button>
            <button
              class="ai-agent-card__action ai-agent-card__action--danger"
              :disabled="deletingId === agent.id"
              @click="removeAgent(agent)"
            >
              {{ deletingId === agent.id ? t("common.loading") : t("common.delete") }}
            </button>
          </div>
        </article>
      </div>
    </template>

    <!-- 创建 / 编辑弹窗 -->
    <AppModal
      :visible="showForm"
      :title="isEditing ? t('aiAgent.formTitleEdit') : t('aiAgent.formTitleCreate')"
      width="640px"
      @close="showForm = false"
    >
      <div class="ai-agent-form">
        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">
            {{ t("aiAgent.name") }} <span class="ai-agent-form__required">*</span>
          </span>
          <input
            v-model="form.name"
            type="text"
            class="ai-agent-form__input"
            maxlength="120"
          />
        </label>

        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.description") }}</span>
          <input
            v-model="form.description"
            type="text"
            class="ai-agent-form__input"
            maxlength="500"
          />
        </label>

        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.instructions") }}</span>
          <textarea
            v-model="form.instructions"
            rows="4"
            class="ai-agent-form__input ai-agent-form__input--area"
            :placeholder="t('aiAgent.instructionsPlaceholder')"
          />
        </label>

        <!-- 触发器 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.triggers") }}</span>
          <div class="ai-agent-form__chips">
            <label
              v-for="tr in ALL_TRIGGERS"
              :key="tr"
              class="ai-agent-chip ai-agent-chip--trigger ai-agent-chip--check"
            >
              <input
                v-model="form.triggers"
                type="checkbox"
                :value="tr"
                class="ai-agent-chip__input"
              />
              {{ triggerLabel(tr) }}
            </label>
          </div>
        </div>

        <!-- 技能 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.skills") }}</span>
          <div class="ai-agent-form__chips">
            <label
              v-for="sk in ALL_SKILLS"
              :key="sk"
              class="ai-agent-chip ai-agent-chip--skill ai-agent-chip--check"
            >
              <input
                v-model="form.skills"
                type="checkbox"
                :value="sk"
                class="ai-agent-chip__input"
              />
              {{ skillLabel(sk) }}
            </label>
          </div>
        </div>

        <!-- 作用项目 -->
        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.scopeProjectIds") }}</span>
          <select
            v-model="form.scope_project_ids"
            multiple
            class="ai-agent-form__input ai-agent-form__input--multi"
          >
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }} ({{ p.identifier }})
            </option>
          </select>
          <span class="ai-agent-form__hint">{{ t("aiAgent.scopeProjectIdsHint") }}</span>
        </label>

        <!-- 作用工作项类型 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.scopeWorkItemTypes") }}</span>
          <div class="ai-agent-form__chips">
            <label
              v-for="wt in ALL_WORK_ITEM_TYPES"
              :key="wt"
              class="ai-agent-chip ai-agent-chip--scope ai-agent-chip--check"
            >
              <input
                v-model="form.scope_work_item_types"
                type="checkbox"
                :value="wt"
                class="ai-agent-chip__input"
              />
              {{ wt }}
            </label>
          </div>
        </div>

        <!-- 模型 / 额度 -->
        <div class="ai-agent-form__row">
          <label class="ai-agent-form__field">
            <span class="ai-agent-form__label">{{ t("aiAgent.model") }}</span>
            <select v-model="form.model" class="ai-agent-form__input">
              <option v-for="m in ALL_MODELS" :key="m" :value="m">{{ m }}</option>
            </select>
          </label>
          <label class="ai-agent-form__field">
            <span class="ai-agent-form__label">{{ t("aiAgent.dailyCreditLimit") }}</span>
            <input
              v-model.number="form.daily_credit_limit"
              type="number"
              min="1"
              class="ai-agent-form__input"
            />
          </label>
        </div>

        <!-- 温度 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">
            {{ t("aiAgent.temperature") }} — {{ form.temperature.toFixed(2) }}
          </span>
          <input
            v-model.number="form.temperature"
            type="range"
            min="0"
            max="2"
            step="0.05"
            class="ai-agent-form__range"
          />
        </div>

        <div v-if="formError" class="ai-agent-form__error">{{ formError }}</div>
      </div>

      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showForm = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton size="sm" :loading="saving" @click="submitForm">
          {{ t("common.save") }}
        </AppButton>
          </AppModal>
  
    </div>
  </AppPageContainer>


    <!-- 创建 / 编辑弹窗 -->
    <AppModal
      :visible="showForm"
      :title="isEditing ? t('aiAgent.formTitleEdit') : t('aiAgent.formTitleCreate')"
      width="640px"
      @close="showForm = false"
    >
      <div class="ai-agent-form">
        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">
            {{ t("aiAgent.name") }} <span class="ai-agent-form__required">*</span>
          </span>
          <input
            v-model="form.name"
            type="text"
            class="ai-agent-form__input"
            maxlength="120"
          />
        </label>

        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.description") }}</span>
          <input
            v-model="form.description"
            type="text"
            class="ai-agent-form__input"
            maxlength="500"
          />
        </label>

        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.instructions") }}</span>
          <textarea
            v-model="form.instructions"
            rows="4"
            class="ai-agent-form__input ai-agent-form__input--area"
            :placeholder="t('aiAgent.instructionsPlaceholder')"
          />
        </label>

        <!-- 触发器 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.triggers") }}</span>
          <div class="ai-agent-form__chips">
            <label
              v-for="tr in ALL_TRIGGERS"
              :key="tr"
              class="ai-agent-chip ai-agent-chip--trigger ai-agent-chip--check"
            >
              <input
                v-model="form.triggers"
                type="checkbox"
                :value="tr"
                class="ai-agent-chip__input"
              />
              {{ triggerLabel(tr) }}
            </label>
          </div>
        </div>

        <!-- 技能 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.skills") }}</span>
          <div class="ai-agent-form__chips">
            <label
              v-for="sk in ALL_SKILLS"
              :key="sk"
              class="ai-agent-chip ai-agent-chip--skill ai-agent-chip--check"
            >
              <input
                v-model="form.skills"
                type="checkbox"
                :value="sk"
                class="ai-agent-chip__input"
              />
              {{ skillLabel(sk) }}
            </label>
          </div>
        </div>

        <!-- 作用项目 -->
        <label class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.scopeProjectIds") }}</span>
          <select
            v-model="form.scope_project_ids"
            multiple
            class="ai-agent-form__input ai-agent-form__input--multi"
          >
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }} ({{ p.identifier }})
            </option>
          </select>
          <span class="ai-agent-form__hint">{{ t("aiAgent.scopeProjectIdsHint") }}</span>
        </label>

        <!-- 作用工作项类型 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">{{ t("aiAgent.scopeWorkItemTypes") }}</span>
          <div class="ai-agent-form__chips">
            <label
              v-for="wt in ALL_WORK_ITEM_TYPES"
              :key="wt"
              class="ai-agent-chip ai-agent-chip--scope ai-agent-chip--check"
            >
              <input
                v-model="form.scope_work_item_types"
                type="checkbox"
                :value="wt"
                class="ai-agent-chip__input"
              />
              {{ wt }}
            </label>
          </div>
        </div>

        <!-- 模型 / 额度 -->
        <div class="ai-agent-form__row">
          <label class="ai-agent-form__field">
            <span class="ai-agent-form__label">{{ t("aiAgent.model") }}</span>
            <select v-model="form.model" class="ai-agent-form__input">
              <option v-for="m in ALL_MODELS" :key="m" :value="m">{{ m }}</option>
            </select>
          </label>
          <label class="ai-agent-form__field">
            <span class="ai-agent-form__label">{{ t("aiAgent.dailyCreditLimit") }}</span>
            <input
              v-model.number="form.daily_credit_limit"
              type="number"
              min="1"
              class="ai-agent-form__input"
            />
          </label>
        </div>

        <!-- 温度 -->
        <div class="ai-agent-form__field">
          <span class="ai-agent-form__label">
            {{ t("aiAgent.temperature") }} — {{ form.temperature.toFixed(2) }}
          </span>
          <input
            v-model.number="form.temperature"
            type="range"
            min="0"
            max="2"
            step="0.05"
            class="ai-agent-form__range"
          />
        </div>

        <div v-if="formError" class="ai-agent-form__error">{{ formError }}</div>
      </div>

      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showForm = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton size="sm" :loading="saving" @click="submitForm">
          {{ t("common.save") }}
        </AppButton>
      
    </AppModal>
  </div>


<style scoped>
.ai-agents {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 24px;
  
  
}

.ai-agents__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.ai-agents__title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-agents__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.ai-agents__actions {
  display: flex;
  gap: 8px;
}

/* ---- 会话面板 ---- */
.ai-agents__chat {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  overflow: hidden;
  background: var(--surface-1);
}

.ai-agents__chat-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--surface-2);
}

.ai-agents__chat-agent {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ai-agents__chat-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-agents__chat-close {
  border: none;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 14px;
}
.ai-agents__chat-close:hover {
  color: var(--text-primary);
}

/* ---- 卡片网格 ---- */
.ai-agents__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 14px;
}

.ai-agent-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--surface-1);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.ai-agent-card:hover {
  border-color: var(--brand-500);
  box-shadow: var(--shadow-popover, 0 4px 12px rgba(0, 0, 0, 0.08));
}

.ai-agent-card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.ai-agent-card__identity {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.ai-agent-card__avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: var(--brand-50);
  color: var(--brand-600);
  flex-shrink: 0;
}

.ai-agent-card__meta {
  min-width: 0;
}

.ai-agent-card__name {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ai-agent-card__model {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: ui-monospace, monospace;
}

.ai-agent-card__desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.ai-agent-card__desc--empty {
  color: var(--text-tertiary);
  font-style: italic;
}

.ai-agent-card__credit {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.ai-agent-card__credit-head {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-tertiary);
}
.ai-agent-card__credit-value {
  font-variant-numeric: tabular-nums;
}

.ai-agent-card__row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.ai-agent-card__row-label {
  font-size: 11px;
  color: var(--text-tertiary);
}
.ai-agent-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.ai-agent-card__empty-chip {
  font-size: 11px;
  color: var(--text-tertiary);
}

.ai-agent-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 1.5;
}
.ai-agent-chip--trigger {
  background: var(--brand-50);
  color: var(--brand-600);
}
.ai-agent-chip--skill {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success-500);
}
.ai-agent-chip--scope {
  background: var(--surface-3);
  color: var(--text-secondary);
}
.ai-agent-chip--more {
  background: var(--surface-3);
  color: var(--text-tertiary);
}
.ai-agent-chip--check {
  cursor: pointer;
  border: 1px solid transparent;
}
.ai-agent-chip--check:hover {
  border-color: var(--brand-500);
}
.ai-agent-chip__input {
  display: none;
}
.ai-agent-chip--check:has(.ai-agent-chip__input:checked) {
  border-color: var(--brand-500);
  background: var(--brand-50);
}

.ai-agent-card__footer {
  display: flex;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px solid var(--border-subtle);
}

.ai-agent-card__action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
}
.ai-agent-card__action:hover:not(:disabled) {
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.ai-agent-card__action--primary {
  color: var(--brand-600);
}
.ai-agent-card__action--danger {
  margin-left: auto;
}
.ai-agent-card__action--danger:hover:not(:disabled) {
  border-color: var(--danger-500);
  color: var(--danger-500);
}
.ai-agent-card__action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ---- 表单 ---- */
.ai-agent-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.ai-agent-form__field {
  display: block;
}
.ai-agent-form__label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--text-secondary);
}
.ai-agent-form__required {
  color: var(--danger-500);
}
.ai-agent-form__input {
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
.ai-agent-form__input:focus {
  outline: none;
  border-color: var(--brand-500);
}
.ai-agent-form__input--area {
  height: auto;
  padding: 8px 10px;
  resize: vertical;
  line-height: 1.5;
}
.ai-agent-form__input--multi {
  height: auto;
  min-height: 80px;
  padding: 6px;
}
.ai-agent-form__hint {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-tertiary);
}
.ai-agent-form__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.ai-agent-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.ai-agent-form__range {
  width: 100%;
  accent-color: var(--brand-500);
}
.ai-agent-form__error {
  font-size: 12px;
  color: var(--danger-500);
}

@media (max-width: 640px) {
  .ai-agents {
    padding: 16px;
  }
  .ai-agents__grid {
    grid-template-columns: 1fr;
  }
  .ai-agent-form__row {
    grid-template-columns: 1fr;
  }
}
</style>
