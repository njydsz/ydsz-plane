<script setup lang="ts">
/**
 * IntakePublicView — 公开门户视图，供外部用户提交工单与匿名追踪。
 *
 * 两种模式（由路由 props 控制）：
 *  - submit: 通过通道提交新工单（需 wsSlug + slug）
 *  - track:  按 tracking_id + email 追踪已有工单状态
 */
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { ApiError } from "@/api/client";
import { intakeApi, type SubmitInput, type TrackResult } from "@/api/services/intake";
import { toast } from "@/lib/toast";

/* ------------------------------------------------------------------ */
/* Props                                                              */
/* ------------------------------------------------------------------ */

const props = defineProps<{
  mode: "submit" | "track";
  wsSlug?: string;
  slug?: string;
}>();

/* ------------------------------------------------------------------ */
/* 路由 / 工具                                                          */
/* ------------------------------------------------------------------ */

const route = useRoute();
const router = useRouter();

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/* ------------------------------------------------------------------ */
/* Submit 模式 — 通道信息加载                                            */
/* ------------------------------------------------------------------ */

interface PublicChannel {
  id: number;
  slug: string;
  name: string;
  description: string;
  default_issue_type: string;
  require_captcha: boolean;
  custom_fields: any;
  branding: any;
}

const channel = ref<PublicChannel | null>(null);
const channelLoading = ref(false);
const channelError = ref("");

async function loadChannel() {
  if (!props.wsSlug || !props.slug) {
    channelError.value = "缺少数通参数";
    return;
  }
  channelLoading.value = true;
  channelError.value = "";
  try {
    const wsId = Number(props.wsSlug);
    if (Number.isNaN(wsId)) {
      channelError.value = "无效的工作空间标识";
      return;
    }
    channel.value = await intakeApi.getPublicChannel(wsId, props.slug);
  } catch (e) {
    if (e instanceof ApiError) {
      channelError.value = e.status === 404 ? "该提交通道不存在或已停用" : e.message;
    } else {
      channelError.value = "该提交通道暂不可用";
    }
  } finally {
    channelLoading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Submit 模式 — 表单状态                                                */
/* ------------------------------------------------------------------ */

interface SubmitForm {
  title: string;
  description: string;
  submitter_name: string;
  submitter_email: string;
  issue_type: string;
  custom_fields: Record<string, any>;
}

const form = reactive<SubmitForm>({
  title: "",
  description: "",
  submitter_name: "",
  submitter_email: "",
  issue_type: "",
  custom_fields: {},
});

const submitting = ref(false);
const rateLimitCountdown = ref(0);

/** 提交成功结果 */
interface SubmitSuccess {
  tracking_id: string;
  submitted_at: string;
  message: string;
}
const submitResult = ref<SubmitSuccess | null>(null);

const titleErr = computed(() => {
  if (!form.title.trim()) return "请输入标题";
  if (form.title.length > 200) return "标题不能超过 200 字";
  return "";
});

const emailErr = computed(() => {
  if (!form.submitter_email.trim()) return "请输入邮箱";
  if (!EMAIL_RE.test(form.submitter_email)) return "邮箱格式不正确";
  return "";
});

const nameErr = computed(() => {
  if (!form.submitter_name.trim()) return "请输入姓名";
  return "";
});

const formValid = computed(() => {
  return !titleErr.value && !emailErr.value && !nameErr.value;
});

/* ------------------------------------------------------------------ */
/* Submit 模式 — 提交逻辑                                               */
/* ------------------------------------------------------------------ */

async function handleSubmit() {
  if (!formValid.value || submitting.value) return;
  submitting.value = true;
  try {
    const input: SubmitInput = {
      title: form.title.trim(),
      description: form.description.trim() || undefined,
      submitter_name: form.submitter_name.trim(),
      submitter_email: form.submitter_email.trim(),
    };
    if (form.issue_type) input.issue_type = form.issue_type;
    if (Object.keys(form.custom_fields).length > 0) {
      input.custom_fields = form.custom_fields;
    }
    const result = await intakeApi.submitIssue(Number(props.wsSlug!), props.slug!, input);
    submitResult.value = {
      tracking_id: result.tracking_id,
      submitted_at: result.submitted_at,
      message: result.message,
    };
  } catch (e) {
    handleSubmitError(e);
  } finally {
    submitting.value = false;
  }
}

function handleSubmitError(e: unknown) {
  if (e instanceof ApiError) {
    if (e.isRateLimited) {
      toast.error(e.message || "提交过于频繁，请稍后再试", 4000);
      rateLimitCountdown.value = 5;
      const timer = setInterval(() => {
        rateLimitCountdown.value--;
        if (rateLimitCountdown.value <= 0) clearInterval(timer);
      }, 1000);
    } else {
      toast.error(e.message || "提交失败，请稍后再试");
    }
  } else {
    toast.error("网络异常，请稍后再试");
  }
}

/* ------------------------------------------------------------------ */
/* Submit 模式 — 成功面板                                               */
/* ------------------------------------------------------------------ */

function goToTrack() {
  if (!submitResult.value) return;
  router.push({
    name: "intake-track",
    query: {
      tracking_id: submitResult.value.tracking_id,
      email: form.submitter_email.trim(),
    },
  });
}

function copyTrackingId() {
  if (!submitResult.value) return;
  navigator.clipboard
    .writeText(submitResult.value.tracking_id)
    .then(() => toast.success("Tracking ID 已复制"))
    .catch(() => toast.error("复制失败，请手动复制"));
}

/* ------------------------------------------------------------------ */
/* Track 模式                                                           */
/* ------------------------------------------------------------------ */

const trackForm = reactive({
  tracking_id: "",
  email: "",
});

const tracking = ref(false);
const trackResult = ref<TrackResult | null>(null);
const trackError = ref("");

async function handleTrack() {
  trackError.value = "";
  trackResult.value = null;

  if (!trackForm.tracking_id.trim()) {
    trackError.value = "请输入 Tracking ID";
    return;
  }
  if (!EMAIL_RE.test(trackForm.email.trim())) {
    trackError.value = "请输入有效的邮箱地址";
    return;
  }

  tracking.value = true;
  try {
    trackResult.value = await intakeApi.trackIssue(
      trackForm.tracking_id.trim(),
      trackForm.email.trim(),
    );
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) {
      trackError.value = "未查到对应工单，请检查 tracking_id 和邮箱";
    } else {
      trackError.value = e instanceof ApiError ? e.message : "查询失败，请稍后再试";
    }
  } finally {
    tracking.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Track 模式 — 状态映射                                                */
/* ------------------------------------------------------------------ */

interface StatusMeta {
  label: string;
  color_var: string;
}

const STATUS_MAP: Record<string, StatusMeta> = {
  pending: { label: "待审核", color_var: "--amber-500" },
  reviewed: { label: "已审核", color_var: "--blue-500" },
  converted: { label: "已转正", color_var: "--emerald-500" },
  rejected: { label: "已拒绝", color_var: "--red-500" },
  archived: { label: "已归档", color_var: "--gray-500" },
};

function statusMeta(status: string): StatusMeta {
  return STATUS_MAP[status] ?? { label: status, color_var: "--gray-400" };
}

/* ------------------------------------------------------------------ */
/* 日期格式化                                                           */
/* ------------------------------------------------------------------ */

function fmtDate(s: string): string {
  if (!s) return "—";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  return d.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

/* ------------------------------------------------------------------ */
/* 初始化                                                              */
/* ------------------------------------------------------------------ */

onMounted(() => {
  if (props.mode === "submit") {
    loadChannel();
  } else {
    // 追踪模式：从 URL query 预填并自动查询
    const qTracking = route.query.tracking_id;
    const qEmail = route.query.email;
    if (typeof qTracking === "string") trackForm.tracking_id = qTracking;
    if (typeof qEmail === "string") trackForm.email = qEmail;
    if (trackForm.tracking_id && trackForm.email) {
      handleTrack();
    }
  }
});

// 监听路由 query 变化（同一路由组件复用场景）
watch(
  () => route.query,
  () => {
    if (props.mode !== "track") return;
    const qTracking = route.query.tracking_id;
    const qEmail = route.query.email;
    if (typeof qTracking === "string") trackForm.tracking_id = qTracking;
    if (typeof qEmail === "string") trackForm.email = qEmail;
    if (trackForm.tracking_id && trackForm.email) handleTrack();
  },
);
</script>

<template>
  <!-- 公开门户根容器 -->
  <div class="ip-view">
    <!-- 顶部品牌条 -->
    <header class="ip-header">
      <span class="ip-header__mark">YD</span>
      <span class="ip-header__title">Ydsz Plane</span>
    </header>

    <main class="ip-main">
      <!-- ====================================================== -->
      <!-- Submit 模式                                             -->
      <!-- ====================================================== -->
      <template v-if="mode === 'submit'">
        <!-- 加载态 -->
        <div v-if="channelLoading" class="ip-card ip-card--loading">
          <p class="ip-hint">正在加载通道信息…</p>
        </div>

        <!-- 错误态 -->
        <div v-else-if="channelError" class="ip-card ip-card--error">
          <p class="ip-err ip-err--lg">{{ channelError }}</p>
          <router-link to="/" class="btn btn--sm">返回首页</router-link>
        </div>

        <!-- 提交成功 -->
        <div v-else-if="submitResult" class="ip-card ip-success">
          <div class="ip-success__icon" aria-hidden="true">✓</div>
          <h2 class="ip-success__title">提交成功</h2>
          <p class="ip-success__msg">{{ submitResult.message }}</p>

          <div class="ip-success__tracking">
            <span class="ip-label">Tracking ID</span>
            <div class="ip-success__tracking-id">
              <code>{{ submitResult.tracking_id }}</code>
              <button class="btn btn--sm" type="button" @click="copyTrackingId">复制</button>
            </div>
          </div>

          <button class="btn btn--primary ip-success__track-btn" type="button" @click="goToTrack">
            查看状态
          </button>

          <p class="ip-hint ip-hint--center">建议复制保存 Tracking ID，以便后续追踪。</p>
        </div>

        <!-- 提交表单 -->
        <form v-else class="ip-card" @submit.prevent="handleSubmit">
          <h2 class="ip-card__title">
            {{ channel?.branding?.name ?? channel?.name ?? "提交工单" }}
          </h2>
          <p v-if="channel?.branding?.description ?? channel?.description" class="ip-card__subtitle">
            {{ channel?.branding?.description ?? channel?.description }}
          </p>

          <!-- 标题 -->
          <div class="ip-field">
            <label class="ip-label" for="ip-title">标题 <span class="ip-label__req">*</span></label>
            <input
              id="ip-title"
              v-model="form.title"
              class="ip-input"
              :class="{ 'ip-input--err': titleErr }"
              type="text"
              maxlength="200"
              placeholder="请输入标题"
            />
            <p v-if="titleErr" class="ip-err">{{ titleErr }}</p>
          </div>

          <!-- 描述 -->
          <div class="ip-field">
            <label class="ip-label" for="ip-desc">描述</label>
            <textarea
              id="ip-desc"
              v-model="form.description"
              class="ip-input ip-textarea"
              rows="4"
              placeholder="请详细描述您的问题或需求（可选）"
            ></textarea>
          </div>

          <!-- 提交者姓名 -->
          <div class="ip-field">
            <label class="ip-label" for="ip-name">姓名 <span class="ip-label__req">*</span></label>
            <input
              id="ip-name"
              v-model="form.submitter_name"
              class="ip-input"
              :class="{ 'ip-input--err': nameErr }"
              type="text"
              placeholder="请输入您的姓名"
            />
            <p v-if="nameErr" class="ip-err">{{ nameErr }}</p>
          </div>

          <!-- 提交者邮箱 -->
          <div class="ip-field">
            <label class="ip-label" for="ip-email">
              邮箱 <span class="ip-label__req">*</span>
            </label>
            <input
              id="ip-email"
              v-model="form.submitter_email"
              class="ip-input"
              :class="{ 'ip-input--err': emailErr }"
              type="email"
              placeholder="you@example.com"
            />
            <p v-if="emailErr" class="ip-err">{{ emailErr }}</p>
          </div>

          <!-- 类型下拉 -->
          <div class="ip-field">
            <label class="ip-label" for="ip-type">类型</label>
            <select id="ip-type" v-model="form.issue_type" class="ip-input">
              <option value="">{{ channel?.default_issue_type ?? "默认" }}</option>
              <option value="bug">缺陷</option>
              <option value="requirement">需求</option>
              <option value="task">任务</option>
            </select>
          </div>

          <!-- 自定义字段 -->
          <template v-if="channel?.custom_fields">
            <div
              v-for="(field, key) in (Array.isArray(channel.custom_fields) ? channel.custom_fields : [channel.custom_fields])"
              :key="String(key)"
              class="ip-field"
            >
              <label class="ip-label" :for="`ip-cf-${String(key)}`">
                {{ typeof field === 'object' && field !== null ? (field as any).label ?? String(key) : String(key) }}
              </label>
              <input
                :id="`ip-cf-${String(key)}`"
                class="ip-input"
                :placeholder="typeof field === 'object' && field !== null ? (field as any).placeholder ?? '' : ''"
              />
            </div>
          </template>

          <!-- 提交按钮 -->
          <button
            class="btn btn--primary ip-submit-btn"
            type="submit"
            :disabled="submitting || !formValid || rateLimitCountdown > 0"
          >
            <template v-if="submitting">提交中…</template>
            <template v-else-if="rateLimitCountdown > 0">
              请稍候 {{ rateLimitCountdown }}s
            </template>
            <template v-else>提交</template>
          </button>
        </form>
      </template>

      <!-- ====================================================== -->
      <!-- Track 模式                                              -->
      <!-- ====================================================== -->
      <template v-else>
        <div class="ip-card">
          <h2 class="ip-card__title">追踪工单</h2>
          <p class="ip-card__subtitle">输入 Tracking ID 与邮箱查询工单状态</p>

          <!-- 查询表单 -->
          <form @submit.prevent="handleTrack">
            <div class="ip-field">
              <label class="ip-label" for="ip-tid">Tracking ID</label>
              <input
                id="ip-tid"
                v-model="trackForm.tracking_id"
                class="ip-input"
                type="text"
                placeholder="例如: INK-20250101-XXXXX"
              />
            </div>

            <div class="ip-field">
              <label class="ip-label" for="ip-temail">邮箱</label>
              <input
                id="ip-temail"
                v-model="trackForm.email"
                class="ip-input"
                type="email"
                placeholder="you@example.com"
              />
            </div>

            <button class="btn btn--primary ip-submit-btn" type="submit" :disabled="tracking">
              {{ tracking ? "查询中…" : "查询" }}
            </button>
          </form>

          <!-- 错误提示 -->
          <p v-if="trackError" class="ip-err ip-err--lg">{{ trackError }}</p>

          <!-- 查询结果 -->
          <div v-if="trackResult" class="ip-track-result">
            <!-- 状态徽章 -->
            <div
              class="ip-track-badge"
              :style="{ '--badge-color': statusMeta(trackResult.status).color_var }"
            >
              {{ statusMeta(trackResult.status).label }}
            </div>

            <!-- 工单信息 -->
            <div class="ip-track-info">
              <div class="ip-track-info__row">
                <span class="ip-track-info__label">Tracking ID</span>
                <span class="ip-track-info__value">{{ trackResult.tracking_id }}</span>
              </div>
              <div class="ip-track-info__row">
                <span class="ip-track-info__label">标题</span>
                <span class="ip-track-info__value">{{ trackResult.title }}</span>
              </div>
              <div class="ip-track-info__row">
                <span class="ip-track-info__label">提交时间</span>
                <span class="ip-track-info__value">{{ fmtDate(trackResult.submitted_at) }}</span>
              </div>
              <div v-if="trackResult.reviewed_at" class="ip-track-info__row">
                <span class="ip-track-info__label">审核时间</span>
                <span class="ip-track-info__value">{{ fmtDate(trackResult.reviewed_at) }}</span>
              </div>
              <div v-if="trackResult.converted_issue_id" class="ip-track-info__row">
                <span class="ip-track-info__label">转换状态</span>
                <span class="ip-track-info__value ip-track-info__value--success">
                  已转正式工作项 #{{ trackResult.converted_issue_id }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
/* ------------------------------------------------------------------ */
/* 布局                                                                */
/* ------------------------------------------------------------------ */

.ip-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--surface-2, #f5f5f5);
}

.ip-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px 16px 8px;
}

.ip-header__mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: var(--brand-500, #f97316);
  color: #fff;
  font-weight: 700;
  font-size: 13px;
}

.ip-header__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
}

.ip-main {
  flex: 1;
  display: flex;
  justify-content: center;
  padding: 24px 16px 48px;
}

.ip-card {
  width: 100%;
  max-width: 640px;
  padding: 32px;
  border-radius: var(--radius-lg, 12px);
  border: 1px solid var(--border-subtle, #e5e7eb);
  background: var(--surface-1, #ffffff);
  box-shadow: var(--shadow-card, 0 1px 3px rgba(0, 0, 0, 0.06));
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.ip-card--loading {
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.ip-card--error {
  align-items: center;
  justify-content: center;
  gap: 16px;
  text-align: center;
}

.ip-card__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
}

.ip-card__subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary, #6b7280);
}

/* ------------------------------------------------------------------ */
/* 表单字段                                                             */
/* ------------------------------------------------------------------ */

.ip-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ip-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary, #4b5563);
}

.ip-label__req {
  color: var(--danger-500, #ef4444);
}

.ip-input {
  height: 38px;
  padding: 0 12px;
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border-default, #d1d5db);
  background: var(--surface-1, #fff);
  color: var(--text-primary, #1a1a1a);
  font-size: 14px;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.ip-input:focus {
  border-color: var(--brand-500, #f97316);
  box-shadow: 0 0 0 3px var(--brand-50, #fff7ed);
}

.ip-input--err {
  border-color: var(--danger-500, #ef4444);
}

.ip-textarea {
  height: auto;
  padding: 8px 12px;
  resize: vertical;
  min-height: 80px;
}

.ip-err {
  margin: 0;
  font-size: 12px;
  color: var(--danger-500, #ef4444);
}

.ip-err--lg {
  font-size: 14px;
  text-align: center;
}

.ip-hint {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary, #6b7280);
}

.ip-hint--center {
  text-align: center;
}

/* ------------------------------------------------------------------ */
/* 提交按钮                                                             */
/* ------------------------------------------------------------------ */

.ip-submit-btn {
  width: 100%;
  height: 42px;
  font-size: 14px;
  font-weight: 500;
}

/* ------------------------------------------------------------------ */
/* 成功面板                                                             */
/* ------------------------------------------------------------------ */

.ip-success {
  align-items: center;
  text-align: center;
  gap: 12px;
}

.ip-success__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--emerald-500, #10b981);
  color: #fff;
  font-size: 28px;
  font-weight: 700;
}

.ip-success__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
}

.ip-success__msg {
  margin: 0;
  font-size: 14px;
  color: var(--text-secondary, #4b5563);
}

.ip-success__tracking {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px;
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-2, #f9fafb);
  border: 1px var(--border-subtle, #e5e7eb);
}

.ip-success__tracking .ip-label {
  font-weight: 600;
}

.ip-success__tracking-id {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.ip-success__tracking-id code {
  font-size: 16px;
  font-weight: 600;
  font-family: var(--font-mono, monospace);
  color: var(--text-primary, #1a1a1a);
}

.ip-success__track-btn {
  width: 100%;
  height: 42px;
}

/* ------------------------------------------------------------------ */
/* 追踪结果                                                             */
/* ------------------------------------------------------------------ */

.ip-track-result {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 8px;
}

.ip-track-badge {
  align-self: center;
  display: inline-flex;
  align-items: center;
  padding: 6px 16px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--badge-color, #9ca3af) 15%, transparent);
  color: var(--badge-color, #9ca3af);
  font-size: 13px;
  font-weight: 600;
}

.ip-track-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ip-track-info__row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  font-size: 14px;
}

.ip-track-info__label {
  flex-shrink: 0;
  width: 90px;
  color: var(--text-tertiary, #6b7280);
}

.ip-track-info__value {
  color: var(--text-primary, #1a1a1a);
  word-break: break-all;
}

.ip-track-info__value--success {
  color: var(--emerald-600, #059669);
  font-weight: 500;
}
</style>
