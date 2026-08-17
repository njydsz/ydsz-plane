<script setup lang="ts">
/**
 * ReviewPanel — 需求评审工作流面板。
 *
 * 用于 IssueDetailView 右侧栏，提供评审活动展示、提交评审、采纳/驳回决定。
 * 仅需求（requirement）类型需求/任务/缺陷支持评审工作流。
 */
import { computed, onMounted, ref, watch } from "vue";

import { issueApi, type ReviewRecord } from "@/api/services/issue";
import { workspaceApi, type Member } from "@/api/services/workspace";
import { useWorkspaceStore } from "@/stores/workspace";
import { useAuthStore } from "@/stores/auth";
import { toast } from "@/lib/toast";

/* ------------------------------------------------------------------ */
/* Props                                                              */
/* ------------------------------------------------------------------ */

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
  issueType: string; // 'requirement' | 'task' | 'defect' | 'epic'
  reviewStatus?: string | null; // 当前评审状态
}>();

/* ------------------------------------------------------------------ */
/* Stores                                                             */
/* ------------------------------------------------------------------ */

const wsStore = useWorkspaceStore();
const auth = useAuthStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

/* ------------------------------------------------------------------ */
/* State                                                              */
/* ------------------------------------------------------------------ */

const loading = ref(false);
const submitting = ref(false);
const reviews = ref<ReviewRecord[]>([]);
const members = ref<Member[]>([]);

// --- 提交评审表单 ---
const showSubmitForm = ref(false);
const submitName = ref("");
const submitReviewers = ref<number[]>([]);
const submitError = ref("");

/* ------------------------------------------------------------------ */
/* Computed                                                           */
/* ------------------------------------------------------------------ */

/** 仅需求支持评审 */
const isReviewable = computed(() => props.issueType === 'requirement');

/** 当前评审状态 */
const currentStatus = computed(() => props.reviewStatus);

/** 是否处于评审中 */
const isInReview = computed(() => currentStatus.value === 'in_review');

/** 当前用户是否为评审人 */
const isReviewer = computed(() => {
  if (!isInReview.value) return false;
  const activeReview = reviews.value.find(r => r.status === 'active');
  if (!activeReview?.reviewers) return false;
  return activeReview.reviewers.includes(currentUserId.value);
});

/** 当前用户是否已提交过决定 */
const hasDecided = computed(() => {
  // 简化：后端记录评审人决定状态；前端仅通过是否有 active review 判断
  return false;
});

/** 是否允许提交评审（ owner/admin 或创建者） */
const canSubmitReview = computed(() => {
  return wsStore.hasPermission('issue:edit_all') || wsStore.hasPermission('issue:edit_own');
});

/* ------------------------------------------------------------------ */
/* Methods                                                            */
/* ------------------------------------------------------------------ */

async function loadReviews() {
  if (!isReviewable.value) return;
  try {
    const { results } = await issueApi.listReviews(props.workspaceId, props.projectId, props.issueId);
    reviews.value = results;
  } catch {
    // 静默失败
  }
}

async function loadMembers() {
  try {
    const { results } = await workspaceApi.getMembers(props.workspaceId);
    members.value = results;
  } catch {
    // 静默失败
  }
}

function openSubmitForm() {
  submitName.value = '';
  submitReviewers.value = [];
  submitError.value = '';
  showSubmitForm.value = true;
}

async function submitReview() {
  if (!submitName.value.trim()) {
    submitError.value = '请输入评审名称';
    return;
  }
  if (submitReviewers.value.length === 0) {
    submitError.value = '请至少选择一位评审人';
    return;
  }
  submitting.value = true;
  submitError.value = '';
  try {
    await issueApi.submitReview(props.workspaceId, props.projectId, props.issueId, {
      name: submitName.value,
      reviewers: submitReviewers.value,
    });
    toast.success('评审已提交');
    showSubmitForm.value = false;
    await loadReviews();
  } catch (e: unknown) {
    submitError.value = e instanceof Error ? e.message : '提交失败';
  } finally {
    submitting.value = false;
  }
}

async function decideReview(decision: 'approved' | 'rejected') {
  try {
    await issueApi.decideReview(props.workspaceId, props.projectId, props.issueId, decision);
    toast.success(decision === 'approved' ? '已采纳' : '已驳回');
    await loadReviews();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '操作失败';
    toast.error(msg);
  }
}

function getMemberName(id: number): string {
  const m = members.value.find(m => m.id === id);
  return m?.nickname || m?.email || `用户${id}`;
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    active: '进行中',
    completed: '已完成',
    cancelled: '已取消',
  };
  return map[status] || status;
}

function statusClass(status: string): string {
  const map: Record<string, string> = {
    active: 'rp-badge--active',
    completed: 'rp-badge--completed',
    cancelled: 'rp-badge--cancelled',
  };
  return map[status] || '';
}

/* ------------------------------------------------------------------ */
/* Lifecycle                                                          */
/* ------------------------------------------------------------------ */

onMounted(() => {
  if (isReviewable.value) {
    void loadReviews();
    void loadMembers();
  }
});

watch(() => props.issueId, () => {
  if (isReviewable.value) {
    void loadReviews();
  }
});
</script>

<template>
  <div v-if="isReviewable" class="rp-panel">
    <!-- 头部 -->
    <div class="rp-panel__header">
      <h3 class="rp-panel__title">需求评审</h3>
      <span v-if="currentStatus && currentStatus !== 'draft'" class="rp-badge" :class="{
        'rp-badge--in-review': currentStatus === 'in_review',
        'rp-badge--approved': currentStatus === 'approved',
        'rp-badge--rejected': currentStatus === 'rejected',
      }">
        {{
          currentStatus === 'in_review' ? '评审中' :
          currentStatus === 'approved' ? '已采纳' :
          currentStatus === 'rejected' ? '已驳回' :
          currentStatus === 'draft' ? '草稿' : currentStatus
        }}
      </span>
    </div>

    <!-- 操作按钮 -->
    <div class="rp-panel__actions">
      <button
        v-if="canSubmitReview && !isInReview"
        class="rp-btn rp-btn--primary"
        @click="openSubmitForm"
      >
        提交评审
      </button>
      <template v-if="isReviewer">
        <button class="rp-btn rp-btn--success" @click="decideReview('approved')">
          采纳
        </button>
        <button class="rp-btn rp-btn--danger" @click="decideReview('rejected')">
          驳回
        </button>
      </template>
    </div>

    <!-- 提交评审表单 -->
    <div v-if="showSubmitForm" class="rp-form">
      <div class="rp-form__field">
        <label class="rp-form__label">评审名称 *</label>
        <input
          v-model="submitName"
          class="rp-input"
          placeholder="如：需求评审-第一轮"
        />
      </div>
      <div class="rp-form__field">
        <label class="rp-form__label">评审人 *</label>
        <select v-model="submitReviewers" class="rp-select" multiple>
          <option v-for="m in members" :key="m.id" :value="m.id">
            {{ m.nickname || m.email }}
          </option>
        </select>
      </div>
      <p v-if="submitError" class="rp-form__error">{{ submitError }}</p>
      <div class="rp-form__actions">
        <button class="rp-btn rp-btn--ghost" @click="showSubmitForm = false">取消</button>
        <button class="rp-btn rp-btn--primary" :disabled="submitting" @click="submitReview">
          {{ submitting ? '提交中...' : '确认提交' }}
        </button>
      </div>
    </div>

    <!-- 评审历史 -->
    <div v-if="reviews.length" class="rp-history">
      <div class="rp-history__title">评审记录</div>
      <div
        v-for="review in reviews"
        :key="review.id"
        class="rp-history__item"
      >
        <div class="rp-history__item-header">
          <span class="rp-history__name">{{ review.name }}</span>
          <span class="rp-badge" :class="statusClass(review.status)">
            {{ statusLabel(review.status) }}
          </span>
        </div>
        <div class="rp-history__meta">
          <span>评审人: {{ review.reviewers?.map(id => getMemberName(id)).join(', ') || '—' }}</span>
        </div>
        <div v-if="review.description" class="rp-history__desc">
          {{ review.description }}
        </div>
        <div class="rp-history__dates">
          <span v-if="review.created_date">开始: {{ review.created_date }}</span>
          <span v-if="review.completed_date">完成: {{ review.completed_date }}</span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <p v-else-if="!loading" class="rp-panel__empty">暂无评审记录</p>
  </div>
</template>

<style scoped>
.rp-panel {
  background: var(--surface-1);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}

.rp-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.rp-panel__title {
  font-size: 15px;
  font-weight: 600;
  margin: 0;
}

.rp-panel__actions {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.rp-panel__empty {
  color: var(--text-tertiary);
  font-size: 13px;
  text-align: center;
  padding: 16px 0;
}

/* Buttons */
.rp-btn {
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--surface-1);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s;
}

.rp-btn--primary {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

.rp-btn--success {
  background: var(--success, #10B981);
  color: white;
  border-color: var(--success, #10B981);
}

.rp-btn--danger {
  color: var(--danger);
  border-color: var(--danger);
  background: transparent;
}

.rp-btn--ghost {
  background: transparent;
}

.rp-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Badge */
.rp-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--surface-2);
  color: var(--text-secondary);
}

.rp-badge--active,
.rp-badge--in-review {
  background: var(--warning-light, #FEF3C7);
  color: var(--warning, #D97706);
}

.rp-badge--completed,
.rp-badge--approved {
  background: var(--success-light, #D1FAE5);
  color: var(--success, #059669);
}

.rp-badge--cancelled,
.rp-badge--rejected {
  background: var(--danger-light, #FEE2E2);
  color: var(--danger, #DC2626);
}

/* Form */
.rp-form {
  background: var(--surface-2);
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 12px;
}

.rp-form__field {
  margin-bottom: 10px;
}

.rp-form__label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.rp-input,
.rp-select {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  background: var(--surface-1);
}

.rp-select[multiple] {
  min-height: 80px;
}

.rp-form__error {
  color: var(--danger);
  font-size: 12px;
  margin: 4px 0;
}

.rp-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

/* History */
.rp-history__title {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.rp-history__item {
  padding: 10px;
  background: var(--surface-2);
  border-radius: 6px;
  margin-bottom: 8px;
}

.rp-history__item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.rp-history__name {
  font-weight: 500;
  font-size: 13px;
}

.rp-history__meta {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.rp-history__desc {
  font-size: 12px;
  color: var(--text-primary);
  margin: 4px 0;
}

.rp-history__dates {
  display: flex;
  gap: 12px;
  font-size: 11px;
  color: var(--text-tertiary);
}
</style>
