<script setup lang="ts">
/**
 * 通知偏好设置页 — 管理事件订阅 / 渠道 / 摘要频率 / 免打扰。
 */
import { onMounted, ref } from "vue";
import { notificationApi, type NotificationPreference } from "@/api/services/notification";
import { AppLoadingState, AppErrorState } from "@/components";
import { workspaceApi, type Workspace } from "@/api/services/workspace";

const props = defineProps<{ workspaceSlug: string }>();

const ws = ref<Workspace | null>(null);
const pref = ref<NotificationPreference | null>(null);
const loading = ref(true);
const saving = ref(false);
const error = ref("");
const saved = ref(false);

/** 事件类型选项（与后端 EventType 对齐） */
const EVENT_OPTIONS: Array<{ value: string; label: string }> = [
  { value: "issue.created", label: "工作项创建" },
  { value: "issue.assigned", label: "工作项指派" },
  { value: "issue.status_changed", label: "工作项状态变更" },
  { value: "issue.deleted", label: "工作项删除" },
  { value: "comment.created", label: "评论" },
  { value: "sprint.started", label: "迭代启动" },
  { value: "sprint.completed", label: "迭代完成" },
  { value: "version.released", label: "版本发布" },
  { value: "member.added", label: "成员加入" },
  { value: "member.removed", label: "成员移除" },
  { value: "member.role_changed", label: "成员角色变更" },
  { value: "invitation.sent", label: "邀请发送" },
];

const DIGEST_OPTIONS = [
  { value: "realtime", label: "实时推送" },
  { value: "daily", label: "每日摘要" },
  { value: "weekly", label: "每周摘要" },
  { value: "off", label: "关闭" },
];

async function load() {
  loading.value = true;
  error.value = "";
  try {
    ws.value = await workspaceApi.getBySlug(props.workspaceSlug);
    pref.value = await notificationApi.getPreference(ws.value.id);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function toggleEvent(value: string) {
  if (!pref.value) return;
  const arr = pref.value.event_types ?? [];
  const idx = arr.indexOf(value);
  if (idx >= 0) {
    arr.splice(idx, 1);
  } else {
    arr.push(value);
  }
  pref.value.event_types = arr;
}

function toggleChannel(value: string) {
  if (!pref.value) return;
  // 始终保留 "in_app"（站内信不可取消）
  const arr = pref.value.channels?.filter((c: string) => c !== "in_app") ?? [];
  const idx = arr.indexOf(value);
  if (idx >= 0) {
    arr.splice(idx, 1);
  } else {
    arr.push(value);
  }
  pref.value.channels = ["in_app", ...arr];
}

function isChannelOn(value: string): boolean {
  if (!pref.value) return value === "in_app";
  return (pref.value.channels ?? []).includes(value);
}

function isEventOn(value: string): boolean {
  // 空数组 = 全部启用
  if (!pref.value) return false;
  if (pref.value.event_types.length === 0) return true;
  return pref.value.event_types.includes(value);
}

async function save() {
  if (!ws.value || !pref.value || saving.value) return;
  saving.value = true;
  saved.value = false;
  error.value = "";
  try {
    pref.value = await notificationApi.updatePreference(ws.value.id, {
      event_types: pref.value.event_types,
      channels: pref.value.channels,
      digest: pref.value.digest,
      dnd_enabled: pref.value.dnd_enabled,
      dnd_start: pref.value.dnd_start,
      dnd_end: pref.value.dnd_end,
      is_enabled: pref.value.is_enabled,
    });
    saved.value = true;
    setTimeout(() => (saved.value = false), 2000);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="notification-prefs">
    <div class="notification-prefs__header">
      <h1 class="notification-prefs__title">通知设置</h1>
      <p class="notification-prefs__subtitle">控制你在此工作空间收到的通知类型与投递方式</p>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <div v-else-if="pref" class="notification-prefs__body">
      <!-- 总开关 -->
      <div class="notification-prefs__card">
        <div class="notification-prefs__row notification-prefs__row--switch">
          <div>
            <div class="notification-prefs__label">接收通知</div>
            <div class="notification-prefs__desc">关闭后不再收到此工作空间的任何通知</div>
          </div>
          <label class="switch">
            <input v-model="pref.is_enabled" type="checkbox" />
            <span class="switch__slider"></span>
          </label>
        </div>
      </div>

      <!-- 事件订阅 -->
      <div class="notification-prefs__card">
        <div class="notification-prefs__card-title">事件订阅</div>
        <p class="notification-prefs__hint">不勾选任何事件 = 接收全部通知</p>
        <div class="notification-prefs__grid">
          <label v-for="opt in EVENT_OPTIONS" :key="opt.value" class="notification-prefs__checkbox">
            <input
              type="checkbox"
              :checked="isEventOn(opt.value)"
              @change="toggleEvent(opt.value)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </div>

      <!-- 投递方式 -->
      <div class="notification-prefs__card">
        <div class="notification-prefs__card-title">投递方式</div>
        <div class="notification-prefs__select-row">
          <label class="notification-prefs__field-label">摘要频率</label>
          <select v-model="pref.digest" class="notification-prefs__select">
            <option v-for="opt in DIGEST_OPTIONS" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
        <p class="notification-prefs__hint">
          站内信（铃铛）始终保留；选择摘要后，即时渠道消息将聚合后按周期送达。
        </p>

        <!-- 通知渠道 -->
        <div class="notification-prefs__channels">
          <div class="notification-prefs__channels-header">
            <span class="notification-prefs__channels-title">通知渠道</span>
            <span class="notification-prefs__channels-subtitle">勾选后渠道将同步推送</span>
          </div>
          <div class="notification-prefs__grid notification-prefs__grid--channels">
            <label class="notification-prefs__checkbox notification-prefs__checkbox--disabled">
              <input type="checkbox" checked disabled />
              <span>站内信</span>
              <span class="notification-prefs__channel-hint">始终开启</span>
            </label>
            <label class="notification-prefs__checkbox">
              <input type="checkbox" :checked="isChannelOn('email')" @change="toggleChannel('email')" />
              <span>邮件</span>
            </label>
            <label class="notification-prefs__checkbox">
              <input type="checkbox" :checked="isChannelOn('wecom')" @change="toggleChannel('wecom')" />
              <span>企业微信</span>
              <span class="notification-prefs__channel-hint">需要配置 webhook</span>
            </label>
            <label class="notification-prefs__checkbox">
              <input type="checkbox" :checked="isChannelOn('dingtalk')" @change="toggleChannel('dingtalk')" />
              <span>钉钉</span>
            </label>
            <label class="notification-prefs__checkbox">
              <input type="checkbox" :checked="isChannelOn('feishu')" @change="toggleChannel('feishu')" />
              <span>飞书</span>
            </label>
          </div>
        </div>
      </div>

      <!-- 免打扰 -->
      <div class="notification-prefs__card">
        <div class="notification-prefs__row notification-prefs__row--switch">
          <div>
            <div class="notification-prefs__label">免打扰时段</div>
            <div class="notification-prefs__desc">该时段内不推送即时渠道通知</div>
          </div>
          <label class="switch">
            <input v-model="pref.dnd_enabled" type="checkbox" />
            <span class="switch__slider"></span>
          </label>
        </div>
        <div v-if="pref.dnd_enabled" class="notification-prefs__dnd">
          <input v-model="pref.dnd_start" type="time" class="notification-prefs__time" />
          <span class="notification-prefs__dnd-sep">至</span>
          <input v-model="pref.dnd_end" type="time" class="notification-prefs__time" />
        </div>
      </div>

      <!-- 操作 -->
      <div class="notification-prefs__actions">
        <span v-if="error" class="notification-prefs__save-error">{{ error }}</span>
        <span v-if="saved" class="notification-prefs__saved">✓ 已保存</span>
        <button class="btn btn--primary" :disabled="saving" @click="save">
          {{ saving ? "保存中..." : "保存设置" }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.notification-prefs {
  max-width: 720px;
  padding: 24px;
}

.notification-prefs__header {
  margin-bottom: 24px;
}

.notification-prefs__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0 0 4px;
}

.notification-prefs__subtitle {
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
  margin: 0;
}

.notification-prefs__state {
  padding: 48px 0;
  text-align: center;
  color: var(--text-tertiary, #9ca3af);
  font-size: 13px;
}

.notification-prefs__state--error {
  color: var(--danger-500, #ef4444);
}

.notification-prefs__body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.notification-prefs__card {
  padding: 20px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 10px);
  background: var(--surface-1, #fff);
}

.notification-prefs__card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin-bottom: 4px;
}

.notification-prefs__hint {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  margin: 4px 0 16px;
}

.notification-prefs__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.notification-prefs__row--switch {
  margin-bottom: 4px;
}

.notification-prefs__label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

.notification-prefs__desc {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  margin-top: 2px;
}

.notification-prefs__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.notification-prefs__checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary, #4b5563);
  cursor: pointer;
}

.notification-prefs__checkbox input {
  accent-color: var(--brand-500, #3b82f6);
}

.notification-prefs__select-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.notification-prefs__field-label {
  font-size: 13px;
  color: var(--text-secondary, #4b5563);
  min-width: 72px;
}

.notification-prefs__select {
  padding: 6px 10px;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary, #1f2937);
  background: var(--surface-2, #f9fafb);
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: var(--radius-sm, 6px);
  outline: none;
}

/* ===== Channel subscription ===== */
.notification-prefs__channels {
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle, #e5e7eb);
}

.notification-prefs__channels-header {
  margin-bottom: 12px;
}

.notification-prefs__channels-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.notification-prefs__channels-subtitle {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  margin-top: 2px;
}

.notification-prefs__grid--channels {
  grid-template-columns: repeat(2, 1fr);
}

.notification-prefs__checkbox--disabled {
  opacity: 0.6;
  cursor: default;
}

.notification-prefs__channel-hint {
  font-size: 10px;
  color: var(--text-tertiary, #9ca3af);
  margin-left: auto;
}

.notification-prefs__dnd {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
}

.notification-prefs__time {
  padding: 6px 10px;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary, #1f2937);
  background: var(--surface-2, #f9fafb);
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: var(--radius-sm, 6px);
  outline: none;
}

.notification-prefs__dnd-sep {
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
}

.notification-prefs__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

.notification-prefs__saved {
  font-size: 13px;
  color: var(--success-500, #16a34a);
}

.notification-prefs__save-error {
  font-size: 13px;
  color: var(--danger-500, #ef4444);
}

/* Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
  flex-shrink: 0;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.switch__slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background: var(--surface-3, #e5e7eb);
  border-radius: 22px;
  transition: background 0.2s;
}

.switch__slider::before {
  content: "";
  position: absolute;
  width: 16px;
  height: 16px;
  left: 3px;
  top: 3px;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.2s;
}

.switch input:checked + .switch__slider {
  background: var(--brand-500, #3b82f6);
}

.switch input:checked + .switch__slider::before {
  transform: translateX(18px);
}

.btn {
  padding: 6px 16px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  cursor: pointer;
  border: 1px solid var(--border-default, #d1d5db);
  background: var(--surface-1, #fff);
  color: var(--text-secondary, #4b5563);
  font-family: inherit;
}

.btn--primary {
  background: var(--brand-500, #3b82f6);
  color: #fff;
  border-color: var(--brand-500, #3b82f6);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600, #2563eb);
}

.btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
}
</style>
