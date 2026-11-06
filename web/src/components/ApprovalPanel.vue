<script setup lang="ts">
/**
 * ApprovalPanel — 工作流审批面板（嵌入工作项详情视图）
 *
 * 能力：
 *  - 展示当前工作项的审批历史（pending / approved / rejected）
 *  - 针对「需要审批」的流转发起审批申请
 *  - 审批人可批准 / 驳回（含意见）
 *  - 会话内实时刷新
 *
 * Props:
 *   workspaceId — 工作空间 ID
 *   projectId   — 项目 ID
 *   issueId     — 工作项 ID
 *   workItemType — 工作项类型（requirement / task / defect）
 */
import { computed, onMounted, ref, watch } from "vue";

import {
  workflowApi,
  type WorkflowApproval,
  type WorkflowTransition,
} from "@/api/services/workflow";
import { workspaceApi } from "@/api/services/workspace";
import { useAuthStore } from "@/stores/auth";
import { toast } from "@/lib/toast";
import AppBadge from "./AppBadge.vue";
import AppLoadingState from "./AppLoadingState.vue";
import AppEmptyState from "./AppEmptyState.vue";

const props = withDefaults(defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
  workItemType: string;
  /** 可选：当前工作项可执行的流转列表（由父组件从工作流配置中解析） */
  transitions?: WorkflowTransition[];
}>(), {
  transitions: () => [],
});

const auth = useAuthStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

/* ---- 状态 ---- */

const approvals = ref<WorkflowApproval[]>([]);
const loading = ref(true);
const error = ref("");

const submitting = ref(false);
const resolvingId = ref<number | null>(null);

// 发起审批表单
const showSubmitForm = ref(false);
const selectedTransitionId = ref<number | null>(null);
const approverId = ref<number>(0);
const submitComment = ref("");

// 处理意见
const resolveComment = ref("");

/* ---- 计算 ---- */

const pendingApprovals = computed(() =>
  approvals.value.filter((a) => a.status === "pending"),
);
const historyApprovals = computed(() =>
  approvals.value.filter((a) => a.status !== "pending"),
);
const pendingForMe = computed(() =>
  pendingApprovals.value.filter((a) => a.approver_id === currentUserId.value),
);
/** 当前工作项涉及的、需要审批的流转 */
const approvalTransitions = computed(() =>
  props.transitions.filter((t) => t.requires_approval),
);

/* ---- 数据加载 ---- */

async function loadApprovals() {
  if (!props.workspaceId) return;
  loading.value = true;
  error.value = "";
  try {
    // 审批列表按工作空间 + 当前用户拉取；这里再按工作项过滤。
    const res = await workflowApi.listApprovals(props.workspaceId, { page: 1 });
    const items = (res as { items?: WorkflowApproval[] }).items ?? [];
    approvals.value = items.filter((a) => a.work_item_id === props.issueId);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载审批记录失败";
  } finally {
    loading.value = false;
  }
}

onMounted(loadApprovals);
watch(() => props.issueId, loadApprovals);

/* ---- 操作 ---- */

async function submitApproval() {
  if (!selectedTransitionId.value || !approverId.value) {
    toast.error("请选择流转与审批人");
    return;
  }
  submitting.value = true;
  try {
    await workflowApi.submitApproval(props.workspaceId, {
      transition_id: selectedTransitionId.value,
      work_item_id: props.issueId,
      work_item_type: props.workItemType,
      approver_id: approverId.value,
      comment: submitComment.value,
    });
    toast.success("审批申请已提交");
    showSubmitForm.value = false;
    selectedTransitionId.value = null;
    approverId.value = 0;
    submitComment.value = "";
    await loadApprovals();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "提交审批失败");
  } finally {
    submitting.value = false;
  }
}

async function resolve(id: number, approve: boolean) {
  resolvingId.value = id;
  try {
    await workflowApi[approve ? "approve" : "reject"](
      props.workspaceId,
      id,
      resolveComment.value,
    );
    toast.success(approve ? "已批准" : "已驳回");
    resolveComment.value = "";
    await loadApprovals();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "处理审批失败");
  } finally {
    resolvingId.value = null;
  }
}

/* ---- 工具 ---- */

function formatTime(s: string | null): string {
  if (!s) return "";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}

const statusMeta: Record<string, { label: string; variant: "success" | "warning" | "danger" | "default" }> = {
  pending: { label: "待审批", variant: "warning" },
  approved: { label: "已批准", variant: "success" },
  rejected: { label: "已驳回", variant: "danger" },
};

/* ---- 成员列表（审批人选择） ---- */

const members = ref<Array<{ id: number; name: string; email: string }>>([]);

async function loadMembers() {
  if (!props.workspaceId) return;
  try {
    const res = await workspaceApi.listMembers(props.workspaceId);
    members.value = (res as unknown as Array<{ id: number; name: string; email: string }>) ?? [];
  } catch {
    members.value = [];
  }
}

onMounted(loadMembers);
watch(() => props.workspaceId, loadMembers);
</script>

<template>
  <div class="approval-panel">
    <!-- 头部 -->
    <div class="approval-panel__header">
      <h3 class="approval-panel__title">审批</h3>
      <button
        v-if="approvalTransitions.length"
        class="approval-panel__action"
        @click="showSubmitForm = !showSubmitForm"
      >
        {{ showSubmitForm ? "取消" : "发起审批" }}
      </button>
    </div>

    <!-- 发起审批表单 -->
    <div v-if="showSubmitForm" class="approval-panel__form">
      <label class="approval-panel__label">目标流转</label>
      <select v-model="selectedTransitionId" class="approval-panel__select">
        <option :value="null" disabled>选择需要审批的流转</option>
        <option v-for="t in approvalTransitions" :key="t.id" :value="t.id">
          {{ t.name || `状态 ${t.from_state_id} → ${t.to_state_id}` }}
        </option>
      </select>

      <label class="approval-panel__label">审批人</label>
      <select v-model="approverId" class="approval-panel__select">
        <option :value="0" disabled>选择审批人</option>
        <option v-for="m in members" :key="m.id" :value="m.id">
          {{ m.name || m.email }}
        </option>
      </select>

      <label class="approval-panel__label">审批意见（可选）</label>
      <textarea
        v-model="submitComment"
        class="approval-panel__textarea"
        rows="2"
        placeholder="说明本次流转的原因"
      />

      <div class="approval-panel__form-actions">
        <button
          class="approval-panel__submit"
          :disabled="submitting || !selectedTransitionId || !approverId"
          @click="submitApproval"
        >
          {{ submitting ? "提交中..." : "提交审批" }}
        </button>
      </div>
    </div>

    <!-- 加载 / 错误 -->
    <AppLoadingState v-if="loading" />
    <div v-else-if="error" class="approval-panel__error">{{ error }}</div>

    <!-- 待我处理 -->
    <div v-else-if="pendingForMe.length" class="approval-panel__section">
      <div class="approval-panel__section-title">待我处理</div>
      <div
        v-for="a in pendingForMe"
        :key="a.id"
        class="approval-panel__item approval-panel__item--pending"
      >
        <div class="approval-panel__item-head">
          <AppBadge :variant="statusMeta[a.status].variant">
            {{ statusMeta[a.status].label }}
          </AppBadge>
          <span class="approval-panel__time">{{ formatTime(a.created_at) }}</span>
        </div>
        <div class="approval-panel__item-body">
          <p class="approval-panel__meta">
            流转 #{{ a.transition_id }} · 工作项 #{{ a.work_item_id }}
          </p>
          <p v-if="a.comment" class="approval-panel__comment">{{ a.comment }}</p>
          <textarea
            v-model="resolveComment"
            class="approval-panel__textarea"
            rows="1"
            placeholder="审批意见"
          />
          <div class="approval-panel__item-actions">
            <button
              class="approval-panel__btn approval-panel__btn--approve"
              :disabled="resolvingId === a.id"
              @click="resolve(a.id, true)"
            >
              批准
            </button>
            <button
              class="approval-panel__btn approval-panel__btn--reject"
              :disabled="resolvingId === a.id"
              @click="resolve(a.id, false)"
            >
              驳回
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 空态 -->
    <AppEmptyState
      v-else-if="!approvals.length"
      title="暂无审批记录"
      description="当工作项发生需要审批的状态流转时，记录会显示在这里"
    />

    <!-- 历史记录 -->
    <div v-else class="approval-panel__section">
      <div class="approval-panel__section-title">审批历史</div>
      <div v-for="a in historyApprovals" :key="a.id" class="approval-panel__item">
        <div class="approval-panel__item-head">
          <AppBadge :variant="statusMeta[a.status].variant">
            {{ statusMeta[a.status].label }}
          </AppBadge>
          <span class="approval-panel__time">{{ formatTime(a.created_at) }}</span>
        </div>
        <p v-if="a.comment" class="approval-panel__comment">{{ a.comment }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.approval-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.approval-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.approval-panel__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--txt-primary);
  margin: 0;
}

.approval-panel__action {
  font-size: 13px;
  color: var(--accent);
  background: none;
  border: none;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
}

.approval-panel__action:hover {
  background: var(--bg-subtle);
}

.approval-panel__form {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface-2);
}

.approval-panel__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-secondary);
}

.approval-panel__select,
.approval-panel__textarea {
  width: 100%;
  padding: 6px 8px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.approval-panel__form-actions {
  display: flex;
  justify-content: flex-end;
}

.approval-panel__submit {
  padding: 6px 14px;
  font-size: 13px;
  color: var(--text-on-color);
  background: var(--accent);
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.approval-panel__submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.approval-panel__section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.approval-panel__section-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.approval-panel__item {
  padding: 10px 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
}

.approval-panel__item--pending {
  border-color: var(--warning);
  background: var(--warning-bg);
}

.approval-panel__item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.approval-panel__time {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.approval-panel__item-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.approval-panel__meta {
  margin: 0;
  font-size: 12px;
  color: var(--txt-secondary);
}

.approval-panel__comment {
  margin: 0;
  font-size: 13px;
  color: var(--txt-primary);
  white-space: pre-wrap;
}

.approval-panel__item-actions {
  display: flex;
  gap: 8px;
}

.approval-panel__btn {
  padding: 4px 12px;
  font-size: 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  color: var(--text-on-color);
}

.approval-panel__btn--approve {
  background: var(--success);
}

.approval-panel__btn--reject {
  background: var(--danger);
}

.approval-panel__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.approval-panel__error {
  font-size: 13px;
  color: var(--danger);
}
</style>
