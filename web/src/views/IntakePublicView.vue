<script setup lang="ts">
/**
 * IntakePublicView — 公开收件箱视图（无需登录）。
 *
 * 模式：
 *  - submit: 通过通道短链 /intake/:wsId/:slug 访问，展示提交表单
 *  - track:  通过 /intake/track 访问，输入 tracking_id + email 查询状态
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { intakeApi, type SubmitResult } from "@/api/services/intake";
import { workspaceApi, type Workspace } from "@/api/services/workspace";
import { ApiError } from "@/api/client";
import { toast } from "@/lib/toast";

import { AppLoadingState, AppErrorState, AppCard, AppButton } from "@/components";

/* ------------------------------------------------------------------ */
/* Props                                                              */
/* ------------------------------------------------------------------ */

const props = defineProps<{
  mode: "submit" | "track";
  wsId?: string;
  slug?: string;
}>();

const route = useRoute();

/* ------------------------------------------------------------------ */
/* 公共状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");

/** submit 模式下解析到的 workspace 及通道信息 */
const workspace = ref<Workspace | null>(null);
const channel = ref<{
  id: number; slug: string; name: string; description: string;
  default_issue_type: string; require_captcha: boolean;
  custom_fields: any; branding: any;
} | null>(null);

/* ------------------------------------------------------------------ */
/* Submit 模式                                                         */
/* ------------------------------------------------------------------ */

const submitting = ref(false);
const submitDone = ref(false);
const submitResult = ref<SubmitResult | null>(null);

const form = ref({
  title: "",
  description: "",
  submitter_name: "",
  submitter_email: "",
  issue_type: "",
});

/** 将 wsId（可能是短字符串或数字字符串）解析为 workspace 数字 ID */
async function resolveWorkspace(slug: string): Promise<number> {
  // 纯数字 → 直接当作 ID
  const asNumber = Number(slug);
  if (Number.isInteger(asNumber) && asNumber > 0) {
    const ws = await workspaceApi.get(asNumber);
    workspace.value = ws;
    return asNumber;
  }
  // 否则当作 slug 查询
  const ws = await workspaceApi.getBySlug(slug);
  workspace.value = ws;
  return ws.id;
}

async function fetchChannel() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWorkspace(props.wsId!);
    channel.value = await intakeApi.getPublicChannel(wsId, props.slug!);
    form.value.issue_type = channel.value.default_issue_type;
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : "无法加载通道信息";
  } finally {
    loading.value = false;
  }
}

async function handleSubmit() {
  if (!channel.value || !workspace.value) return;
  submitting.value = true;
  error.value = "";
  try {
    const result = await intakeApi.submitIssue(workspace.value.id, channel.value.slug, {
      title: form.value.title,
      description: form.value.description,
      submitter_name: form.value.submitter_name,
      submitter_email: form.value.submitter_email,
      issue_type: form.value.issue_type || undefined,
    });
    submitResult.value = result;
    submitDone.value = true;
    toast.success("提交成功，请保存您的追踪 ID");
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : "提交失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Track 模式                                                          */
/* ------------------------------------------------------------------ */

const trackingId = ref("");
const trackEmail = ref("");
const tracking = ref(false);
const trackError = ref("");
const trackResult = ref<{
  tracking_id: string; status: string; title: string; description?: string;
  status_text?: string; status_reason?: string;
  priority?: number; issue_type?: string;
  submitted_at: string; reviewed_at?: string; converted_issue_id?: number;
} | null>(null);

async function handleTrack() {
  tracking.value = true;
  trackError.value = "";
  trackResult.value = null;
  try {
    const result = await intakeApi.trackIssue(trackingId.value.trim(), trackEmail.value.trim());
    trackResult.value = result;
  } catch (e) {
    trackError.value = e instanceof ApiError ? e.message : "查询失败，请检查追踪 ID 和邮箱";
  } finally {
    tracking.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 路由参数补齐（track 模式从 query 回填）                                */
/* ------------------------------------------------------------------ */

onMounted(() => {
  if (props.mode === "submit" && props.wsId && props.slug) {
    fetchChannel();
  } else if (props.mode === "track") {
    // 从 query 参数回填（方便邮件直链）
    trackingId.value = String(route.query.tracking_id || "");
    trackEmail.value = String(route.query.email || "");
    // 若两个参数都存在则自动查询
    if (trackingId.value && trackEmail.value) {
      handleTrack();
    }
    loading.value = false;
  } else {
    error.value = "缺少必要参数";
    loading.value = false;
  }
});

/* ------------------------------------------------------------------ */
/* 计算属性                                                             */
/* ------------------------------------------------------------------ */

const statusLabels: Record<string, string> = {
  pending: "待审核",
  reviewed: "已审核",
  converted: "已转正",
  rejected: "已拒绝",
  archived: "已归档",
};

const statusColors: Record<string, string> = {
  pending: "var(--warning-500)",
  reviewed: "var(--info-500)",
  converted: "var(--success-500)",
  rejected: "var(--danger-500)",
  archived: "var(--text-tertiary)",
};

const priorityLabels: Record<number, string> = {
  5: "紧急", 4: "高", 3: "中", 2: "低", 1: "最低",
};
const priorityColors: Record<number, string> = {
  5: "var(--danger-500)",
  4: "var(--warning-600)",
  3: "var(--info-500)",
  2: "var(--text-tertiary)",
  1: "var(--text-tertiary)",
};

/** 状态时间线：从提交到当前状态的关键节点 */
interface TimelineStep {
  label: string;
  time?: string;
  active: boolean;
  done: boolean;
}
function buildTimeline(r: {
  status: string; submitted_at: string; reviewed_at?: string;
}): TimelineStep[] {
  return [
    { label: "已提交", time: r.submitted_at, active: true, done: true },
    {
      label: "处理中",
      active: r.status !== "pending",
      done: r.status === "reviewed" || r.status === "converted",
    },
    {
      label: r.status === "rejected" ? "已拒绝" : "已处理",
      time: r.reviewed_at,
      active: !!r.reviewed_at,
      done: r.status === "reviewed" || r.status === "converted",
    },
  ];
}

const pageTitle = computed(() =>
  props.mode === "submit"
    ? (channel.value?.name ?? "提交工单")
    : "追踪提交状态",
);
</script>

<template>
  <div class="intake-public">
    <div class="intake-public__container">
      <!-- 头部 -->
      <header class="intake-public__header">
        <h1 class="intake-public__title">{{ pageTitle }}</h1>
        <p v-if="mode === 'submit' && workspace" class="intake-public__subtitle">
          {{ workspace.name }}
        </p>
      </header>

      <!-- 加载中 -->
      <AppLoadingState v-if="loading" text="加载中..." />

      <!-- 错误 -->
      <AppErrorState
        v-else-if="error"
        :message="error"
        @retry="mode === 'submit' ? fetchChannel() : undefined"
      />

      <!-- ==================== Submit 模式 ==================== -->
      <AppCard v-else-if="mode === 'submit' && channel" padding="lg">
        <!-- 已成功提交 -->
        <div v-if="submitDone && submitResult" class="submit-success">
          <div class="submit-success__icon">✓</div>
          <h2>提交成功</h2>
          <p class="submit-success__msg">{{ submitResult.message }}</p>
          <div class="tracking-box">
            <span class="tracking-box__label">追踪 ID</span>
            <code class="tracking-box__id">{{ submitResult.tracking_id }}</code>
          </div>
          <p class="submit-success__hint">
            请保存此 ID，用于后续查询处理进度。
          </p>
          <div class="submit-success__actions">
            <AppButton @click="submitDone = false">再次提交</AppButton>
            <a
              :href="`/intake/track?tracking_id=${encodeURIComponent(submitResult.tracking_id)}`"
              class="app-btn app-btn--secondary app-btn--md"
            >追踪状态</a>
          </div>
        </div>

        <!-- 提交表单 -->
        <form v-else class="intake-form" @submit.prevent="handleSubmit">
          <p v-if="channel.description" class="intake-form__desc">
            {{ channel.description }}
          </p>

          <!-- 错误提示 -->
          <div v-if="error" class="intake-form__error">{{ error }}</div>

          <label class="field">
            <span class="field-label">标题 <em>*</em></span>
            <input
              v-model="form.title"
              type="text"
              required
              maxlength="200"
              placeholder="简要描述您的问题"
            />
          </label>

          <label class="field">
            <span class="field-label">详细描述</span>
            <textarea
              v-model="form.description"
              rows="6"
              maxlength="5000"
              placeholder="请提供尽可能详细的信息..."
            />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field-label">您的姓名 <em>*</em></span>
              <input
                v-model="form.submitter_name"
                type="text"
                required
                maxlength="100"
              />
            </label>
            <label class="field">
              <span class="field-label">邮箱 <em>*</em></span>
              <input
                v-model="form.submitter_email"
                type="email"
                required
                placeholder="用于接收处理通知"
              />
            </label>
          </div>

          <div class="intake-form__actions">
            <AppButton
              type="submit"
              :loading="submitting"
              :disabled="!form.title || !form.submitter_name || !form.submitter_email"
              block
            >
              提交
            </AppButton>
          </div>
        </form>
      </AppCard>

      <!-- ==================== Track 模式 ==================== -->
      <AppCard v-else-if="mode === 'track'" padding="lg">
        <form class="intake-form" @submit.prevent="handleTrack">
          <h2 class="intake-form__heading">查询提交状态</h2>
          <p class="intake-form__desc">输入您的追踪 ID 与邮箱查看处理进度。</p>

          <div v-if="trackError" class="intake-form__error">{{ trackError }}</div>

          <label class="field">
            <span class="field-label">追踪 ID</span>
            <input
              v-model="trackingId"
              type="text"
              required
              placeholder="例: INK-20250808-0001"
            />
          </label>

          <label class="field">
            <span class="field-label">提交时使用的邮箱</span>
            <input
              v-model="trackEmail"
              type="email"
              required
            />
          </label>

          <div class="intake-form__actions">
            <AppButton type="submit" :loading="tracking" block>查询</AppButton>
          </div>
        </form>

        <!-- 查询结果 -->
        <div v-if="trackResult" class="track-result">
          <h3 class="track-result__title">{{ trackResult.title }}</h3>

          <!-- 状态时间线 -->
          <ol class="timeline">
            <li
              v-for="(step, i) in buildTimeline(trackResult)"
              :key="i"
              class="timeline__step"
              :class="{
                'timeline__step--active': step.active,
                'timeline__step--done': step.done,
              }"
            >
              <span class="timeline__dot" />
              <span class="timeline__label">{{ step.label }}</span>
              <span v-if="step.time" class="timeline__time">{{ step.time }}</span>
            </li>
          </ol>

          <div class="track-result__meta">
            <span
              class="track-result__status"
              :style="{ color: statusColors[trackResult.status] }"
            >
              {{ trackResult.status_text || statusLabels[trackResult.status] || trackResult.status }}
            </span>
            <span
              v-if="trackResult.priority"
              class="track-result__priority"
              :style="{ color: priorityColors[trackResult.priority] }"
            >
              {{ priorityLabels[trackResult.priority] || trackResult.priority }}
            </span>
            <span class="track-result__id">{{ trackResult.tracking_id }}</span>
          </div>

          <!-- 描述 -->
          <p v-if="trackResult.description" class="track-result__desc">
            {{ trackResult.description }}
          </p>

          <!-- 状态说明（拒绝/归档原因等）-->
          <p
            v-if="trackResult.status_reason"
            class="track-result__reason"
          >
            说明：{{ trackResult.status_reason }}
          </p>

          <div class="track-result__times">
            <span>提交时间：{{ trackResult.submitted_at }}</span>
            <span v-if="trackResult.reviewed_at">
              处理时间：{{ trackResult.reviewed_at }}
            </span>
          </div>

          <p v-if="trackResult.converted_issue_id" class="track-result__hint">
            已转正为正式工作项 #{{ trackResult.converted_issue_id }}
          </p>
        </div>
      </AppCard>
    </div>
  </div>
</template>

<style scoped>
.intake-public {
  min-height: 100vh;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 48px 16px;
  background: var(--surface-2);
}

.intake-public__container {
  width: 100%;
  max-width: 560px;
}

.intake-public__header {
  text-align: center;
  margin-bottom: 24px;
}

.intake-public__title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.intake-public__subtitle {
  font-size: 14px;
  color: var(--text-tertiary);
  margin-top: 6px;
}

/* --- 表单 --- */
.intake-form__heading {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 4px;
  color: var(--text-primary);
}

.intake-form__desc {
  font-size: 14px;
  color: var(--text-tertiary);
  margin: 0 0 20px;
  line-height: 1.5;
}

.intake-form__error {
  background: var(--danger-50);
  color: var(--danger-600);
  padding: 10px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
  margin-bottom: 16px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.field-label em {
  color: var(--danger-500);
  font-style: normal;
}

.field input,
.field textarea {
  padding: 10px 14px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-family: inherit;
  background: var(--surface-1);
  color: var(--text-primary);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.field input:focus,
.field textarea:focus {
  outline: none;
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-100);
}

.field textarea {
  resize: vertical;
  min-height: 120px;
}

.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

@media (max-width: 480px) {
  .field-row {
    grid-template-columns: 1fr;
  }
}

.intake-form__actions {
  margin-top: 8px;
}

/* --- 提交成功 --- */
.submit-success {
  text-align: center;
  padding: 16px 0;
}

.submit-success__icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--success-50);
  color: var(--success-600);
  font-size: 28px;
  line-height: 56px;
  margin: 0 auto 12px;
}

.submit-success h2 {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 8px;
  color: var(--text-primary);
}

.submit-success__msg {
  color: var(--text-tertiary);
  font-size: 14px;
  margin-bottom: 20px;
}

.tracking-box {
  background: var(--surface-2);
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  padding: 14px;
  margin-bottom: 12px;
}

.tracking-box__label {
  display: block;
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 6px;
}

.tracking-box__id {
  font-size: 20px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--brand-600);
  letter-spacing: 0.5px;
}

.submit-success__hint {
  font-size: 13px;
  color: var(--text-tertiary);
  margin-bottom: 20px;
}

.submit-success__actions {
  display: flex;
  gap: 8px;
  justify-content: center;
}

/* --- 追踪结果 --- */
.track-result {
  margin-top: 24px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
}

.track-result__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 10px;
}

.track-result__meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.track-result__status {
  font-size: 13px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background: var(--surface-3);
}

.track-result__id {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
}

.track-result__times {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
}

.track-result__hint {
  font-size: 13px;
  color: var(--info-600);
  margin: 0;
}
</style>
