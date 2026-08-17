<script setup lang="ts">
/**
 * 收件箱公开页 — 匿名提报 + 提交跟踪（免登录）。
 * 两个路由模式：
 *  - /intake/:workspaceId/:slug  公开提报表单（按渠道 slug 提交）
 *  - /intake/track               提交者凭 tracking_id + 邮箱跟踪状态
 */
import { computed, ref } from "vue";
import { useRoute } from "vue-router";

import { intakeApi, type IntakeChannel, type IntakeIssue } from "@/api/services/intake";
import { AppErrorState } from "@/components";

const route = useRoute();

// 路由模式判定：/intake/:wsId/:slug → 提交模式；/intake/track → 跟踪模式
const isSubmitMode = computed(() => Boolean(route.params.slug));
const channelSlug = computed(() => String(route.params.slug ?? ""));

const channel = ref<IntakeChannel | null>(null);
const channelError = ref("");
const loadingChannel = ref(true);

const form = ref({
  name: "",
  description: "",
  submitter_name: "",
  submitter_email: "",
  priority: "medium",
});
const submitting = ref(false);
const submitted = ref<IntakeIssue | null>(null);
const submitError = ref("");

// ---- 跟踪模式 ----
const trackForm = ref({ tracking_id: "", submitter_email: "" });
const tracking = ref(false);
const trackResult = ref<IntakeIssue | null>(null);
const trackError = ref("");

const STATUS_TEXT: Record<string, string> = {
  open: "待处理", accepted: "已接受", rejected: "已拒绝", archived: "已归档",
};
const PRIORITY_TEXT: Record<string, string> = {
  urgent: "紧急", high: "高", medium: "中", low: "低", none: "无",
};

async function loadChannel() {
  if (!isSubmitMode.value) return;
  loadingChannel.value = true;
  channelError.value = "";
  try {
    channel.value = await intakeApi.publicGetChannel(channelSlug.value);
  } catch (err: unknown) {
    channelError.value = err instanceof Error ? err.message : "渠道不存在或已停用";
  } finally {
    loadingChannel.value = false;
  }
}

async function submit() {
  if (!form.value.name.trim()) { submitError.value = "请填写标题"; return; }
  if (!form.value.email) { submitError.value = "请填写联系邮箱，用于跟踪处理进度"; return; }
  submitError.value = "";
  submitting.value = true;
  try {
    submitted.value = await intakeApi.publicSubmitIssue({
      channel_slug: channelSlug.value,
      name: form.value.name,
      description: form.value.description,
      submitter_name: form.value.submitter_name,
      submitter_email: form.value.submitter_email,
      priority: form.value.priority,
    });
  } catch (err: unknown) {
    submitError.value = err instanceof Error ? err.message : "提交失败，请稍后再试";
  } finally {
    submitting.value = false;
  }
}

async function track() {
  if (!trackForm.value.tracking_id.trim() || !trackForm.value.submitter_email.trim()) {
    trackError.value = "请填写跟踪号与提交邮箱";
    return;
  }
  trackError.value = "";
  tracking.value = true;
  try {
    trackResult.value = await intakeApi.publicTrackIssue(
      trackForm.value.tracking_id.trim(),
      trackForm.value.submitter_email.trim(),
    );
  } catch (err: unknown) {
    trackResult.value = null;
    trackError.value = err instanceof Error ? err.message : "查询失败";
  } finally {
    tracking.value = false;
  }
}

loadChannel();
</script>

<template>
  <div class="mx-auto flex min-h-screen max-w-lg flex-col justify-center px-4 py-10">
    <div class="mb-6 text-center">
      <h1 class="text-2xl font-bold tracking-tight">{{ isSubmitMode ? (channel?.name ?? "提报") : "提报跟踪" }}</h1>
      <p class="mt-1 text-sm text-[var(--text-tertiary)]">
        {{ isSubmitMode ? (channel?.description || "提交问题或需求，团队会尽快处理") : "输入跟踪号与提交邮箱，查询处理进度" }}
      </p>
    </div>

    <!-- 提交模式 -->
    <template v-if="isSubmitMode">
      <div v-if="loadingChannel" class="space-y-3">
        <div v-for="i in 3" :key="i" class="h-11 animate-pulse rounded-md bg-[var(--bg-secondary)]" />
      </div>
      <AppErrorState v-else-if="channelError" :message="channelError" />

      <div v-else-if="submitted" class="space-y-4 rounded-xl border border-[var(--border-subtle)] p-6 text-center">
        <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-[var(--success-bg,#e6f7ee)] text-2xl">✓</div>
        <h2 class="text-lg font-semibold">提交成功</h2>
        <p class="text-sm text-[var(--text-tertiary)]">请保存您的跟踪号，用于查询处理进度：</p>
        <p class="rounded-md bg-[var(--bg-secondary)] py-2 font-mono text-lg font-semibold text-[var(--brand-600)]">{{ submitted.tracking_id }}</p>
        <div class="text-left text-sm">
          <div class="flex justify-between py-1"><span class="text-[var(--text-tertiary)]">标题</span><span>{{ submitted.name }}</span></div>
          <div class="flex justify-between py-1"><span class="text-[var(--text-tertiary)]">当前状态</span><span>{{ STATUS_TEXT[submitted.status] }}</span></div>
          <div class="flex justify-between py-1"><span class="text-[var(--text-tertiary)]">优先级</span><span>{{ PRIORITY_TEXT[submitted.priority] ?? submitted.priority }}</span></div>
        </div>
        <RouterLink to="/intake/track" class="inline-block text-sm text-[var(--brand-600)]">去跟踪我的提报 →</RouterLink>
      </div>

      <form v-else class="space-y-4 rounded-xl border border-[var(--border-subtle)] p-6" @submit.prevent="submit">
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">标题 *</label>
          <input v-model="form.name" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="一句话描述问题或需求" />
        </div>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">详细描述</label>
          <textarea v-model="form.description" rows="5" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="背景、复现步骤、期望结果…"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="text-xs text-[var(--text-tertiary)]">称呼</label>
            <input v-model="form.submitter_name" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="如何称呼您" />
          </div>
          <div>
            <label class="text-xs text-[var(--text-tertiary)]">优先级</label>
            <select v-model="form.priority" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-2 py-2 text-sm">
              <option value="low">低</option>
              <option value="medium">中</option>
              <option value="high">高</option>
              <option value="urgent">紧急</option>
            </select>
          </div>
        </div>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">联系邮箱 *</label>
          <input v-model="form.submitter_email" type="email" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="用于查询处理进度" />
        </div>
        <p v-if="submitError" class="text-sm text-[var(--danger,#ef4444)]">{{ submitError }}</p>
        <button
          :disabled="submitting"
          class="w-full rounded-md bg-[var(--brand-600)] px-4 py-2.5 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
        >
          {{ submitting ? "提交中…" : "提交" }}
        </button>
        <RouterLink to="/intake/track" class="block text-center text-xs text-[var(--text-tertiary)] hover:text-[var(--brand-600)]">已有跟踪号？查询进度 →</RouterLink>
      </form>
    </template>

    <!-- 跟踪模式 -->
    <template v-else>
      <form class="space-y-4 rounded-xl border border-[var(--border-subtle)] p-6" @submit.prevent="track">
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">跟踪号</label>
          <input v-model="trackForm.tracking_id" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="如 INC-1A2B3C4D" />
        </div>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">提交邮箱</label>
          <input v-model="trackForm.submitter_email" type="email" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="提交时填写的邮箱" />
        </div>
        <p v-if="trackError" class="text-sm text-[var(--danger,#ef4444)]">{{ trackError }}</p>
        <button :disabled="tracking" class="w-full rounded-md bg-[var(--brand-600)] px-4 py-2.5 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50">
          {{ tracking ? "查询中…" : "查询进度" }}
        </button>
      </form>

      <div v-if="trackResult" class="mt-4 space-y-3 rounded-xl border border-[var(--border-subtle)] p-6">
        <div class="flex items-center justify-between">
          <span class="font-medium">{{ trackResult.name }}</span>
          <span
            class="rounded px-1.5 py-0.5 text-xs"
            :class="{
              'bg-[var(--warning-bg,#fef3c7)] text-[var(--warning,#d97706)]': trackResult.status === 'open',
              'bg-[var(--success-bg,#e6f7ee)] text-[var(--success,#16a34a)]': trackResult.status === 'accepted',
              'bg-[var(--bg-secondary)] text-[var(--text-tertiary)]': trackResult.status === 'rejected' || trackResult.status === 'archived',
            }"
          >{{ STATUS_TEXT[trackResult.status] }}</span>
        </div>
        <div class="text-sm text-[var(--text-tertiary)]">
          <p>渠道：{{ trackResult.channel_name || "—" }}</p>
          <p>提交时间：{{ new Date(trackResult.created_at).toLocaleString() }}</p>
          <p v-if="trackResult.linked_entity_identifier">已转正：{{ trackResult.linked_entity_identifier }}</p>
        </div>
      </div>

      <RouterLink to="/login" class="mt-4 block text-center text-xs text-[var(--text-tertiary)] hover:text-[var(--brand-600)]">返回登录 →</RouterLink>
    </template>
  </div>
</template>
