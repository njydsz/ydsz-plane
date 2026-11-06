<script setup lang="ts">
/**
 * AIAgentChat — Plane AI 风格的会话面板。
 *
 * 能力：
 *  - 左侧会话列表（切换 / 新建）；
 *  - Ask / Build / Autopilot 三种模式卡片切换；
 *  - 消息气泡（user / assistant，含时间戳）；
 *  - Build 模式下的动作卡片：批准 / 拒绝 / 执行已批准动作；
 *  - 加载、空态、错误态与输入态。
 *
 * Props:
 *   workspaceId: 工作空间 ID
 *   agents:      可选 Agent 列表（不传则自行拉取）
 *   initialConversationId: 初始会话 ID（可选）
 *
 * Events:
 *   @agents-change: Agent 列表加载完成后回传
 */
import { nextTick, onMounted, ref, watch } from "vue";

import {
  aiAgentApi,
  parseActionPayload,
  type AIAgent,
  type AIAction,
  type AIConversation,
  type AIMessage,
  type AIMode,
  type AIWorkspaceMemory,
} from "@/api/services/aiagent";
import { workspaceApi, type Project } from "@/api/services/workspace";
import { AppEmptyState, AppLoadingState } from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useI18n } from "@/composables/useI18n";
import { formatRelativeTime } from "@/lib/formatTime";
import { toast } from "@/lib/toast";

const props = withDefaults(
  defineProps<{
    workspaceId: number;
    agents?: AIAgent[];
    initialConversationId?: number | null;
    /** 是否显示会话侧栏（嵌入页时可隐藏） */
    showSidebar?: boolean;
  }>(),
  {
    agents: () => [],
    initialConversationId: null,
    showSidebar: true,
  },
);

const emit = defineEmits<{
  "agents-change": [agents: AIAgent[]];
}>();

const { t } = useI18n();

/* ------------------------------------------------------------------ */
/* 基础数据                                                            */
/* ------------------------------------------------------------------ */

const agentList = ref<AIAgent[]>([...props.agents]);
const projects = ref<Project[]>([]);
const memories = ref<AIWorkspaceMemory[]>([]);

const selectedAgentId = ref<number | null>(null);
const selectedProjectId = ref<number | null>(null);
const mode = ref<AIMode>("ask");
const conversations = ref<AIConversation[]>([]);
const activeConversationId = ref<number | null>(props.initialConversationId);
const messages = ref<AIMessage[]>([]);
const actions = ref<AIAction[]>([]);

const loadingConversations = ref(false);
const loadingMessages = ref(false);
const sending = ref(false);
const inputText = ref("");
const listEl = ref<HTMLElement | null>(null);

/* ------------------------------------------------------------------ */
/* 工具函数                                                            */
/* ------------------------------------------------------------------ */

function errorMessage(err: unknown, fallback: string): string {
  return err instanceof Error && err.message ? err.message : fallback;
}

async function scrollToBottom() {
  await nextTick();
  if (listEl.value) listEl.value.scrollTop = listEl.value.scrollHeight;
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

/* ------------------------------------------------------------------ */
/* 数据加载                                                            */
/* ------------------------------------------------------------------ */

async function loadAgents() {
  if (props.agents.length > 0) {
    agentList.value = [...props.agents];
    return;
  }
  try {
    agentList.value = await aiAgentApi.listAgents(props.workspaceId);
    emit("agents-change", agentList.value);
  } catch (err) {
    // 静默：Agent 列表失败时由调用方通过 empty 态处理
    agentList.value = [];
    void errorMessage(err, t("aiAgent.loadFailed"));
  }
}

async function loadProjects() {
  try {
    projects.value = await workspaceApi.listProjects(props.workspaceId);
  } catch {
    projects.value = [];
  }
}

async function loadMemories() {
  try {
    memories.value = await aiAgentApi.listMemory(props.workspaceId);
  } catch {
    memories.value = [];
  }
}

async function loadConversations() {
  if (!props.workspaceId) return;
  loadingConversations.value = true;
  try {
    conversations.value = await aiAgentApi.listConversations(props.workspaceId);
    // 若当前没有选中会话，默认选中最新的一个
    if (!activeConversationId.value && conversations.value.length > 0) {
      await openConversation(conversations.value[0].id);
    }
  } catch (err) {
    conversations.value = [];
    toast.error(errorMessage(err, t("aiAgent.chat.switchConversationFailed")));
  } finally {
    loadingConversations.value = false;
  }
}

async function openConversation(id: number) {
  activeConversationId.value = id;
  loadingMessages.value = true;
  try {
    const detail = await aiAgentApi.getConversation(props.workspaceId, id);
    messages.value = detail.messages ?? [];
    // 会话模式跟随服务端记录
    if (detail.conversation?.mode) mode.value = detail.conversation.mode;
    if (detail.conversation?.agent_id) selectedAgentId.value = detail.conversation.agent_id;
    await scrollToBottom();
    await refreshActions();
  } catch (err) {
    messages.value = [];
    toast.error(errorMessage(err, t("aiAgent.chat.switchConversationFailed")));
  } finally {
    loadingMessages.value = false;
  }
}

async function refreshActions() {
  if (!activeConversationId.value) {
    actions.value = [];
    return;
  }
  try {
    actions.value = await aiAgentApi.listActions(
      props.workspaceId,
      activeConversationId.value,
    );
  } catch (err) {
    actions.value = [];
    toast.error(errorMessage(err, t("aiAgent.chat.loadActionsFailed")));
  }
}

/* ------------------------------------------------------------------ */
/* 会话 / 消息                                                         */
/* ------------------------------------------------------------------ */

function startNewConversation() {
  activeConversationId.value = null;
  messages.value = [];
  actions.value = [];
  inputText.value = "";
}

async function send() {
  const content = inputText.value.trim();
  if (!content || sending.value) return;
  if (!selectedAgentId.value) {
    toast.error(t("aiAgent.chat.noAgents"));
    return;
  }

  sending.value = true;
  const currentMode = mode.value;
  const agentId = selectedAgentId.value;
  const projectId = selectedProjectId.value ?? undefined;

  // 乐观用户消息（在 try 外声明，保证 catch 可回滚）
  const optimisticUser: AIMessage = {
    id: -Date.now(),
    conversation_id: activeConversationId.value ?? 0,
    role: "user",
    content,
    created_at: new Date().toISOString(),
  };

  try {
    messages.value = [...messages.value, optimisticUser];
    inputText.value = "";
    await scrollToBottom();

    if (!activeConversationId.value) {
      // 无会话：先创建会话（含首条消息）
      const conv = await aiAgentApi.createConversation(props.workspaceId, {
        agent_id: agentId,
        mode: currentMode,
        message: content,
        project_id: projectId,
      });
      activeConversationId.value = conv.id;
      await loadConversations();
      await openConversation(conv.id);
      return;
    }

    const reply = await aiAgentApi.sendMessage(
      props.workspaceId,
      activeConversationId.value,
      { content, mode: currentMode, project_id: projectId },
    );
    // 替换乐观消息 + 追加 AI 回复
    messages.value = [
      ...messages.value.filter((m) => m.id !== optimisticUser.id),
      { ...optimisticUser, id: reply.id, conversation_id: reply.conversation_id },
      reply,
    ];
    await scrollToBottom();
    await refreshActions();
    await loadConversations();
  } catch (err) {
    // 回滚乐观消息
    messages.value = messages.value.filter((m) => m.id !== optimisticUser.id);
    toast.error(errorMessage(err, t("aiAgent.chat.sendMessageFailed")));
  } finally {
    sending.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 动作：批准 / 拒绝 / 执行                                             */
/* ------------------------------------------------------------------ */

const pendingActionIds = ref<Set<number>>(new Set());

async function approveAction(action: AIAction) {
  pendingActionIds.value.add(action.id);
  try {
    await aiAgentApi.approveAction(props.workspaceId, action.id);
    toast.success(t("aiAgent.chat.approved"));
    await refreshActions();
  } catch (err) {
    toast.error(errorMessage(err, t("aiAgent.chat.approveFailed")));
  } finally {
    pendingActionIds.value.delete(action.id);
  }
}

async function rejectAction(action: AIAction) {
  pendingActionIds.value.add(action.id);
  try {
    await aiAgentApi.rejectAction(props.workspaceId, action.id);
    toast.success(t("aiAgent.chat.rejected"));
    await refreshActions();
  } catch (err) {
    toast.error(errorMessage(err, t("aiAgent.chat.rejectFailed")));
  } finally {
    pendingActionIds.value.delete(action.id);
  }
}

async function executeAction(action: AIAction) {
  pendingActionIds.value.add(action.id);
  try {
    await aiAgentApi.executeAction(props.workspaceId, action.id);
    toast.success(t("aiAgent.chat.executed"));
    await refreshActions();
    await loadConversations();
  } catch (err) {
    toast.error(errorMessage(err, t("aiAgent.chat.executeFailed")));
  } finally {
    pendingActionIds.value.delete(action.id);
  }
}

async function executeApproved() {
  const approved = actions.value.filter((a) => a.status === "approved");
  if (approved.length === 0) return;
  for (const a of approved) {
    await executeAction(a);
  }
}

/* ------------------------------------------------------------------ */
/* 动作摘要                                                            */
/* ------------------------------------------------------------------ */

const ACTION_TYPE_KEYS: Record<string, string> = {
  create_work_item: "aiAgent.skillCreateWorkItem",
  update_work_item: "aiAgent.skillUpdateWorkItem",
  create_cycle: "aiAgent.skillCreateCycle",
  create_module: "aiAgent.skillCreateModule",
  create_page: "aiAgent.skillCreatePage",
  update_state: "aiAgent.skillUpdateState",
  assign: "aiAgent.skillAssign",
  add_comment: "aiAgent.skillAddComment",
  log_time: "aiAgent.skillLogTime",
  bulk_update: "aiAgent.skillBulkUpdate",
  delete_work_item: "aiAgent.skillDeleteWorkItem",
};

const ACTION_ICONS: Record<string, string> = {
  create_work_item: "CirclePlus",
  update_work_item: "SquarePen",
  create_cycle: "History",
  create_module: "Component",
  create_page: "FileText",
  update_state: "ArrowRight",
  assign: "User",
  add_comment: "MessageSquareText",
  log_time: "Timer",
  bulk_update: "ListChecks",
  delete_work_item: "Trash2",
};

function actionTypeLabel(type: string): string {
  const key = ACTION_TYPE_KEYS[type];
  return key ? t(key) : type;
}

function actionIcon(type: string): string {
  return ACTION_ICONS[type] ?? "Zap";
}

function actionStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: t("aiAgent.chat.pending"),
    approved: t("aiAgent.chat.approved"),
    rejected: t("aiAgent.chat.rejected"),
    executed: t("aiAgent.chat.executed"),
    failed: t("aiAgent.chat.failed"),
  };
  return map[status] ?? status;
}

function actionStatusVariant(status: string): "warning" | "info" | "default" | "success" | "danger" {
  switch (status) {
    case "pending":
      return "warning";
    case "approved":
      return "info";
    case "rejected":
      return "default";
    case "executed":
      return "success";
    case "failed":
      return "danger";
    default:
      return "default";
  }
}

/** 动作 payload 摘要（展示关键字段） */
function actionSummary(action: AIAction): string {
  const payload = parseActionPayload(action.payload);
  if (!payload) return action.payload || "—";
  const parts: string[] = [];
  const title = payload.title ?? payload.name ?? payload.content;
  if (typeof title === "string" && title) parts.push(title);
  if (typeof payload.work_item_type === "string" && payload.work_item_type) {
    parts.push(`[${payload.work_item_type}]`);
  }
  if (typeof payload.priority === "string" && payload.priority) {
    parts.push(`(${payload.priority})`);
  }
  if (Array.isArray(payload.ids) && payload.ids.length) {
    const ids = payload.ids
      .map((id) => (typeof id === "number" || typeof id === "string" ? String(id) : ""))
      .filter(Boolean);
    if (ids.length) parts.push(`#${ids.join(", #")}`);
  }
  return parts.length ? parts.join(" ") : action.payload;
}

/* ------------------------------------------------------------------ */
/* 生命周期                                                            */
/* ------------------------------------------------------------------ */

onMounted(async () => {
  await Promise.all([loadAgents(), loadProjects(), loadMemories()]);
  await loadConversations();
});

watch(
  () => props.workspaceId,
  () => {
    void loadConversations();
  },
);

watch(
  () => props.agents,
  (v) => {
    if (v && v.length) agentList.value = [...v];
  },
);

defineExpose({
  reload: loadConversations,
  openConversation,
  startNewConversation,
});
</script>

<template>
  <div class="ai-chat">
    <!-- 会话侧栏 -->
    <aside v-if="showSidebar" class="ai-chat__side">
      <div class="ai-chat__side-head">
        <span class="ai-chat__side-title">
          <DynamicIcon name="MessagesSquare" :size="16" />
          {{ t("aiAgent.chat.conversations") }}
        </span>
        <button
          class="ai-chat__icon-btn"
          :aria-label="t('aiAgent.chat.newConversation')"
          @click="startNewConversation"
        >
          <DynamicIcon name="CirclePlus" :size="16" />
        </button>
      </div>

      <div class="ai-chat__memory-badge">
        <DynamicIcon name="Brain" :size="14" />
        <span>{{ t("aiMemory.title") }} · {{ memories.length }}</span>
      </div>

      <div v-if="loadingConversations" class="ai-chat__side-loading">
        <AppLoadingState :text="t('common.loading')" />
      </div>
      <div v-else-if="conversations.length === 0" class="ai-chat__side-empty">
        {{ t("aiAgent.chat.noConversation") }}
      </div>
      <ul v-else class="ai-chat__conv-list">
        <li
          v-for="c in conversations"
          :key="c.id"
          class="ai-chat__conv"
          :class="{ 'ai-chat__conv--active': c.id === activeConversationId }"
          role="button"
          tabindex="0"
          @click="openConversation(c.id)"
          @keydown.enter="openConversation(c.id)"
        >
          <span class="ai-chat__conv-mode">{{ c.mode }}</span>
          <span class="ai-chat__conv-title">{{ c.title || t("aiAgent.chat.newConversation") }}</span>
          <span class="ai-chat__conv-time">{{ formatRelativeTime(c.updated_at) }}</span>
        </li>
      </ul>
    </aside>

    <!-- 主区域 -->
    <div class="ai-chat__main">
      <!-- 顶部：Agent / 项目 / 模式 -->
      <div class="ai-chat__topbar">
        <div class="ai-chat__topbar-left">
          <select
            v-model="selectedAgentId"
            class="ai-chat__select"
            :aria-label="t('aiAgent.chat.selectAgent')"
          >
            <option :value="null" disabled>{{ t("aiAgent.chat.selectAgent") }}</option>
            <option v-for="a in agentList" :key="a.id" :value="a.id">
              {{ a.name }}
            </option>
          </select>

          <select
            v-model="selectedProjectId"
            class="ai-chat__select ai-chat__select--sm"
            :aria-label="t('aiAgent.chat.selectProjectOptional')"
          >
            <option :value="null">{{ t("aiAgent.chat.selectProjectOptional") }}</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }}
            </option>
          </select>
        </div>

        <div class="ai-chat__modes">
          <button
            class="ai-chat__mode"
            :class="{ 'ai-chat__mode--active': mode === 'ask' }"
            @click="mode = 'ask'"
          >
            <DynamicIcon name="MessageCircle" :size="16" />
            <span class="ai-chat__mode-name">{{ t("aiAgent.chat.modeAsk") }}</span>
            <span class="ai-chat__mode-desc">{{ t("aiAgent.chat.modeAskDesc") }}</span>
          </button>
          <button
            class="ai-chat__mode"
            :class="{ 'ai-chat__mode--active': mode === 'build' }"
            @click="mode = 'build'"
          >
            <DynamicIcon name="ListChecks" :size="16" />
            <span class="ai-chat__mode-name">{{ t("aiAgent.chat.modeBuild") }}</span>
            <span class="ai-chat__mode-desc">{{ t("aiAgent.chat.modeBuildDesc") }}</span>
          </button>
          <button
            class="ai-chat__mode"
            :class="{ 'ai-chat__mode--active': mode === 'autopilot' }"
            @click="mode = 'autopilot'"
          >
            <DynamicIcon name="Zap" :size="16" />
            <span class="ai-chat__mode-name">{{ t("aiAgent.chat.modeAutopilot") }}</span>
            <span class="ai-chat__mode-desc">{{ t("aiAgent.chat.modeAutopilotDesc") }}</span>
          </button>
        </div>
      </div>

      <!-- 消息列表 -->
      <div ref="listEl" class="ai-chat__messages">
        <div v-if="loadingMessages" class="ai-chat__center">
          <AppLoadingState :text="t('common.loading')" />
        </div>

        <div
          v-else-if="messages.length === 0"
          class="ai-chat__center"
        >
          <AppEmptyState
            icon="🤖"
            :title="t('aiAgent.chat.emptyMessages')"
            :description="t('aiAgent.chat.emptyMessagesDesc')"
          />
        </div>

        <template v-else>
          <div
            v-for="m in messages"
            :key="m.id"
            class="ai-chat__msg"
            :class="m.role === 'user' ? 'ai-chat__msg--user' : 'ai-chat__msg--ai'"
          >
            <div class="ai-chat__bubble">
              <div class="ai-chat__bubble-head">
                <span class="ai-chat__bubble-role">
                  {{ m.role === "user" ? t("aiAgent.chat.you") : t("aiAgent.chat.assistant") }}
                </span>
                <span class="ai-chat__bubble-time">{{ formatTime(m.created_at) }}</span>
              </div>
              <div class="ai-chat__bubble-body">{{ m.content }}</div>
            </div>
          </div>
        </template>

        <!-- 思考中 -->
        <div v-if="sending" class="ai-chat__msg ai-chat__msg--ai">
          <div class="ai-chat__bubble">
            <div class="ai-chat__bubble-head">
              <span class="ai-chat__bubble-role">{{ t("aiAgent.chat.assistant") }}</span>
            </div>
            <div class="ai-chat__thinking">
              <span class="ai-chat__dot" />
              <span class="ai-chat__dot" />
              <span class="ai-chat__dot" />
              <span class="ai-chat__thinking-text">{{ t("aiAgent.chat.thinking") }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 动作卡片（Build 模式） -->
      <div v-if="actions.length > 0" class="ai-chat__actions">
        <div class="ai-chat__actions-head">
          <span class="ai-chat__actions-title">
            <DynamicIcon name="ListChecks" :size="16" />
            {{ t("aiAgent.chat.actionsTitle") }} ({{ actions.length }})
          </span>
          <button
            v-if="actions.some((a) => a.status === 'approved')"
            class="ai-chat__execute-all"
            @click="executeApproved"
          >
            <DynamicIcon name="Play" :size="14" />
            {{ t("aiAgent.chat.executeApproved") }}
          </button>
        </div>

        <div class="ai-chat__action-list">
          <div
            v-for="a in actions"
            :key="a.id"
            class="ai-chat__action"
            :class="`ai-chat__action--${a.status}`"
          >
            <div class="ai-chat__action-icon">
              <DynamicIcon :name="actionIcon(a.type)" :size="16" />
            </div>
            <div class="ai-chat__action-body">
              <div class="ai-chat__action-head">
                <span class="ai-chat__action-type">{{ actionTypeLabel(a.type) }}</span>
                <span
                  class="ai-chat__action-status"
                  :class="`ai-chat__action-status--${actionStatusVariant(a.status)}`"
                >
                  {{ actionStatusLabel(a.status) }}
                </span>
              </div>
              <div class="ai-chat__action-summary">{{ actionSummary(a) }}</div>
              <div v-if="a.error_message" class="ai-chat__action-error">
                {{ a.error_message }}
              </div>
            </div>
            <div class="ai-chat__action-ops">
              <template v-if="a.status === 'pending'">
                <button
                  class="ai-chat__mini ai-chat__mini--ok"
                  :disabled="pendingActionIds.has(a.id)"
                  @click="approveAction(a)"
                >
                  <DynamicIcon name="Check" :size="14" />
                  {{ t("aiAgent.chat.approve") }}
                </button>
                <button
                  class="ai-chat__mini ai-chat__mini--no"
                  :disabled="pendingActionIds.has(a.id)"
                  @click="rejectAction(a)"
                >
                  <DynamicIcon name="X" :size="14" />
                  {{ t("aiAgent.chat.reject") }}
                </button>
              </template>
              <button
                v-else-if="a.status === 'approved'"
                class="ai-chat__mini ai-chat__mini--go"
                :disabled="pendingActionIds.has(a.id)"
                @click="executeAction(a)"
              >
                <DynamicIcon name="Play" :size="14" />
                {{ t("aiAgent.chat.execute") }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="ai-chat__composer">
        <textarea
          v-model="inputText"
          class="ai-chat__input"
          rows="2"
          :placeholder="t('aiAgent.chat.inputPlaceholder')"
          :disabled="sending || !selectedAgentId"
          @keydown.enter.exact.prevent="send"
        />
        <button
          class="ai-chat__send"
          :disabled="sending || !inputText.trim() || !selectedAgentId"
          @click="send"
        >
          <DynamicIcon name="SendHorizontal" :size="16" />
          {{ sending ? t("aiAgent.chat.sending") : t("aiAgent.chat.send") }}
        </button>
      </div>

      <div v-if="!selectedAgentId" class="ai-chat__hint">
        {{ t("aiAgent.chat.noAgents") }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.ai-chat {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  height: 100%;
  min-height: 480px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg, 12px);
  background: var(--surface-1);
  overflow: hidden;
}

.ai-chat__side {
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-subtle);
  background: var(--surface-2);
  min-height: 0;
}

.ai-chat__side-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid var(--border-subtle);
}

.ai-chat__side-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-chat__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
}
.ai-chat__icon-btn:hover {
  background: var(--surface-3);
  color: var(--brand-600);
}

.ai-chat__memory-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 6px 10px;
  padding: 5px 8px;
  border-radius: 6px;
  background: rgba(139, 92, 246, 0.08);
  color: var(--brand-500);
  font-size: 11px;
}

.ai-chat__side-loading,
.ai-chat__side-empty {
  padding: 16px 14px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.ai-chat__conv-list {
  list-style: none;
  margin: 0;
  padding: 6px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ai-chat__conv {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  border: 1px solid transparent;
}
.ai-chat__conv:hover {
  background: var(--surface-1);
}
.ai-chat__conv--active {
  background: var(--surface-1);
  border-color: var(--brand-500);
}
.ai-chat__conv-mode {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--brand-600);
  font-weight: 600;
}
.ai-chat__conv-title {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ai-chat__conv-time {
  font-size: 11px;
  color: var(--text-tertiary);
}

/* ---- 主区域 ---- */
.ai-chat__main {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.ai-chat__topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-subtle);
  flex-wrap: wrap;
}

.ai-chat__topbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ai-chat__select {
  height: 32px;
  padding: 0 8px;
  border: 1px solid var(--border-default);
  border-radius: 8px;
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
  max-width: 180px;
}
.ai-chat__select--sm {
  max-width: 140px;
  color: var(--text-secondary);
}

.ai-chat__modes {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.ai-chat__mode {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--border-default);
  border-radius: 8px;
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}
.ai-chat__mode:hover {
  border-color: var(--brand-500);
}
.ai-chat__mode--active {
  background: var(--brand-50);
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.ai-chat__mode-name {
  font-weight: 600;
}
.ai-chat__mode-desc {
  font-size: 11px;
  color: var(--text-tertiary);
}
.ai-chat__mode--active .ai-chat__mode-desc {
  color: var(--brand-600);
  opacity: 0.8;
}

/* ---- 消息 ---- */
.ai-chat__messages {
  flex: 1;
  min-height: 200px;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--surface-1);
}

.ai-chat__center {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ai-chat__msg {
  display: flex;
}
.ai-chat__msg--user {
  justify-content: flex-end;
}
.ai-chat__msg--ai {
  justify-content: flex-start;
}

.ai-chat__bubble {
  max-width: 78%;
  padding: 10px 12px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.6;
  border: 1px solid var(--border-subtle);
  background: var(--surface-2);
  color: var(--text-primary);
}
.ai-chat__msg--user .ai-chat__bubble {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-color);
}

.ai-chat__bubble-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 11px;
  opacity: 0.8;
}
.ai-chat__bubble-role {
  font-weight: 600;
}
.ai-chat__bubble-time {
  font-variant-numeric: tabular-nums;
}

.ai-chat__bubble-body {
  white-space: pre-wrap;
  word-break: break-word;
}

.ai-chat__thinking {
  display: flex;
  align-items: center;
  gap: 5px;
  color: var(--text-tertiary);
}
.ai-chat__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-tertiary);
  animation: aiBlink 1.2s infinite;
}
.ai-chat__dot:nth-child(2) { animation-delay: 0.2s; }
.ai-chat__dot:nth-child(3) { animation-delay: 0.4s; }
.ai-chat__thinking-text {
  font-size: 12px;
  margin-left: 4px;
}
@keyframes aiBlink {
  0%, 80%, 100% { opacity: 0.25; }
  40% { opacity: 1; }
}

/* ---- 动作卡片 ---- */
.ai-chat__actions {
  border-top: 1px solid var(--border-subtle);
  padding: 10px 14px;
  background: var(--surface-2);
  max-height: 220px;
  overflow-y: auto;
}

.ai-chat__actions-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.ai-chat__actions-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-chat__execute-all {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid var(--brand-500);
  border-radius: 6px;
  background: var(--brand-500);
  color: var(--text-on-color);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
}
.ai-chat__execute-all:hover {
  opacity: 0.9;
}

.ai-chat__action-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ai-chat__action {
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr) auto;
  gap: 10px;
  padding: 10px;
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
  background: var(--surface-1);
}
.ai-chat__action--pending {
  border-color: var(--warning-300);
}
.ai-chat__action--approved {
  border-color: var(--brand-300);
}
.ai-chat__action--executed {
  border-color: var(--success-200);
}
.ai-chat__action--failed {
  border-color: var(--danger-200);
}
.ai-chat__action--rejected {
  opacity: 0.7;
}

.ai-chat__action-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--brand-50);
  color: var(--brand-600);
}

.ai-chat__action-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.ai-chat__action-type {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}
.ai-chat__action-status {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 4px;
}
.ai-chat__action-status--warning { background: var(--warning-50); color: var(--warning-500); }
.ai-chat__action-status--info { background: var(--info-50); color: var(--info-500); }
.ai-chat__action-status--success { background: rgba(16,185,129,0.12); color: var(--success-500); }
.ai-chat__action-status--danger { background: rgba(239,68,68,0.12); color: var(--danger-500); }
.ai-chat__action-status--default { background: var(--surface-3); color: var(--text-secondary); }

.ai-chat__action-summary {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ai-chat__action-error {
  margin-top: 4px;
  font-size: 11px;
  color: var(--danger-500);
}

.ai-chat__action-ops {
  display: flex;
  align-items: center;
  gap: 6px;
}

.ai-chat__mini {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 6px;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
}
.ai-chat__mini:hover:not(:disabled) {
  opacity: 0.85;
}
.ai-chat__mini:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.ai-chat__mini--ok {
  color: var(--success-500);
  border-color: var(--success-200);
}
.ai-chat__mini--no {
  color: var(--danger-500);
  border-color: var(--danger-200);
}
.ai-chat__mini--go {
  color: var(--text-on-color);
  background: var(--brand-500);
  border-color: var(--brand-500);
}

/* ---- 输入 ---- */
.ai-chat__composer {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid var(--border-subtle);
  background: var(--surface-1);
}

.ai-chat__input {
  flex: 1;
  resize: none;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: 8px;
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
  line-height: 1.5;
}
.ai-chat__input:focus {
  outline: none;
  border-color: var(--brand-500);
}
.ai-chat__input:disabled {
  background: var(--surface-2);
  cursor: not-allowed;
}

.ai-chat__send {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 14px;
  border: none;
  border-radius: 8px;
  background: var(--brand-500);
  color: var(--text-on-color);
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
  flex-shrink: 0;
}
.ai-chat__send:hover:not(:disabled) {
  background: var(--brand-600);
}
.ai-chat__send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ai-chat__hint {
  padding: 6px 14px 10px;
  font-size: 12px;
  color: var(--warning-500);
}

/* ---- 窄屏 ---- */
@media (max-width: 860px) {
  .ai-chat {
    grid-template-columns: 1fr;
  }
  .ai-chat__side {
    display: none;
  }
  .ai-chat__topbar {
    flex-direction: column;
    align-items: stretch;
  }
  .ai-chat__modes {
    width: 100%;
  }
  .ai-chat__mode {
    flex: 1;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
  }
}
</style>
