<script setup lang="ts">
/**
 * ProjectSettingsView — 项目设置页。
 *
 * 展示并编辑项目基本信息（名称、描述、标识、网络类型等）。
 */
import { onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import { workspaceApi, type Project, type ProjectModuleToggles } from "@/api/services/workspace";
import { ApiError } from "@/api/client";
import { AppLoadingState, AppErrorState } from "@/components";

const route = useRoute();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const project = ref<Project | null>(null);
const loading = ref(true);
const error = ref("");

const form = reactive({
  name: "",
  description: "",
  network: "public",
  color: "#3f63f1",
  modules: {
    intake: true,
    sprint: true,
    version: true,
    estimate: true,
  } as ProjectModuleToggles,
});

const moduleToggles = reactive({
  intake: true,
  sprint: true,
  version: true,
  estimate: true,
});
const saving = ref(false);
const saveError = ref("");
const saveSuccess = ref("");

const networkOptions = [
  { value: "public", label: "公开 — 空间内所有成员可见" },
  { value: "internal", label: "内部 — 仅项目成员可见（对外隐藏）" },
  { value: "private", label: "私有 — 仅 Owner/Admin 可见" },
];

const colorPresets = [
  "#3f63f1", "#0fc27b", "#f59e0b", "#dc2f2f",
  "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16",
];

async function loadProject() {
  loading.value = true;
  error.value = "";
  try {
    // 通过 slug 拿到 wsId，再拿 project
    const ws = await workspaceApi.get(workspaceId);
    project.value = await workspaceApi.getProject(ws.id, projectId);
    form.name = project.value.name;
    form.description = project.value.description ?? "";
    form.network = project.value.network ?? "public";
    form.color = project.value.color ?? "#3f63f1";
    if (project.value.modules) {
      moduleToggles.intake = project.value.modules.intake;
      moduleToggles.sprint = project.value.modules.sprint;
      moduleToggles.version = project.value.modules.version;
      moduleToggles.estimate = project.value.modules.estimate;
    }
  } catch (e: any) {
    error.value = e.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
}

async function save() {
  saveError.value = "";
  saveSuccess.value = "";
  if (!form.name.trim()) {
    saveError.value = "项目名称不能为空";
    return;
  }

  saving.value = true;
  try {
    const ws = await workspaceApi.get(workspaceId);
    const updated = await workspaceApi.updateProject(ws.id, projectId, {
      name: form.name.trim(),
      description: form.description.trim() || undefined,
      network: form.network,
      color: form.color,
      modules: {
        intake: moduleToggles.intake,
        sprint: moduleToggles.sprint,
        version: moduleToggles.version,
        estimate: moduleToggles.estimate,
      },
    });
    project.value = updated;
    saveSuccess.value = "保存成功";
    setTimeout(() => { saveSuccess.value = ""; }, 2000);
  } catch (e) {
    saveError.value = e instanceof ApiError ? e.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

onMounted(loadProject);
</script>

<template>
  <div class="project-settings">
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="loadProject" />
    <div v-else-if="project" class="settings">
    <header class="settings__header">
      <h1>{{ project.name }} 设置</h1>
      <p class="meta">#{{ project.identifier }} · 项目配置与管理</p>
    </header>

    <nav class="settings-tabs">
      <router-link
        :to="`/${route.params.workspaceId}/projects/${route.params.projectId}/settings`"
        class="tab"
        exact-active-class="tab--active"
      >基本信息</router-link>
      <router-link
        :to="`/${route.params.workspaceId}/projects/${route.params.projectId}/settings/members`"
        class="tab"
        active-class="tab--active"
      >成员管理</router-link>
      <router-link
        :to="`/${route.params.workspaceId}/projects/${route.params.projectId}/settings/modules`"
        class="tab"
        active-class="tab--active"
      >模块管理</router-link>
    </nav>

    <section class="panel">
      <h2 class="panel__title">基本信息</h2>

      <div class="form-grid">
        <!-- 名称 -->
        <label class="form-item">
          <span class="form-item__label">项目名称</span>
          <input v-model="form.name" type="text" maxlength="128" />
        </label>

        <!-- 标识符（只读） -->
        <div class="form-item">
          <span class="form-item__label">标识符</span>
          <span class="form-item__value">{{ project.identifier }}</span>
        </div>

        <!-- 描述 -->
        <label class="form-item form-item--full">
          <span class="form-item__label">描述</span>
          <textarea v-model="form.description" rows="3" maxlength="500" placeholder="项目的简要描述" />
        </label>

        <!-- 网络类型 -->
        <label class="form-item">
          <span class="form-item__label">可见范围</span>
          <select v-model="form.network">
            <option v-for="n in networkOptions" :key="n.value" :value="n.value">
              {{ n.label }}
            </option>
          </select>
        </label>

        <!-- 颜色 -->
        <div class="form-item">
          <span class="form-item__label">主题色</span>
          <div class="color-picker">
            <button
              v-for="c in colorPresets"
              :key="c"
              type="button"
              class="color-swatch"
              :class="{ active: form.color === c }"
              :style="{ background: c }"
              :aria-label="`选择颜色 ${c}`"
              @click="form.color = c"
            />
            <label class="color-custom">
              <input v-model="form.color" type="color" />
              <span>自定义</span>
            </label>
          </div>
        </div>

        <!-- Slug（只读） -->
        <div class="form-item">
          <span class="form-item__label">链接标识</span>
          <span class="form-item__value mono">{{ project.slug }}</span>
        </div>

        <!-- 创建时间（只读） -->
        <div class="form-item">
          <span class="form-item__label">创建时间</span>
          <span class="form-item__value">{{ project.created_at.slice(0, 10) }}</span>
        </div>
      </div>

    </section>

    <!-- 功能模块开关 -->
    <section class="panel" style="margin-top: 24px">
      <h2 class="panel__title">功能模块开关</h2>
      <p class="panel__desc">启用或禁用项目中的功能模块，关闭后在导航中隐藏对应入口。</p>
      <div class="toggle-grid">
        <label class="toggle-item">
          <div class="toggle-info">
            <span class="toggle-name">收件箱 (Intake)</span>
            <span class="toggle-desc">外部反馈收集与审核通道</span>
          </div>
          <input type="checkbox" v-model="moduleToggles.intake" class="toggle-switch" />
        </label>
        <label class="toggle-item">
          <div class="toggle-info">
            <span class="toggle-name">迭代 (Sprint)</span>
            <span class="toggle-desc">敏捷迭代规划与进度管理</span>
          </div>
          <input type="checkbox" v-model="moduleToggles.sprint" class="toggle-switch" />
        </label>
        <label class="toggle-item">
          <div class="toggle-info">
            <span class="toggle-name">版本日 (Version)</span>
            <span class="toggle-desc">发布里程碑与迭代聚合管理</span>
          </div>
          <input type="checkbox" v-model="moduleToggles.version" class="toggle-switch" />
        </label>
        <label class="toggle-item">
          <div class="toggle-info">
            <span class="toggle-name">估算 (Estimate)</span>
            <span class="toggle-desc">故事点与工时估算体系</span>
          </div>
          <input type="checkbox" v-model="moduleToggles.estimate" class="toggle-switch" />
        </label>
      </div>
    </section>

    <!-- 保存 -->
    <div class="actions" style="margin-top: 20px">
      <p v-if="saveError" class="msg error">{{ saveError }}</p>
      <p v-if="saveSuccess" class="msg success">{{ saveSuccess }}</p>
      <button class="btn btn--primary" :disabled="saving" @click="save">
        {{ saving ? "保存中..." : "保存所有变更" }}
      </button>
    </div>
  </div>
  </div>
</template>

<style scoped>
.project-settings { max-width: 700px; }
.loading,
.error-state {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}
.error-state { color: var(--danger-500); }

.settings { max-width: 700px; }

.settings__header {
  margin-bottom: 16px;
}

/* ---------- Tab 导航 ---------- */
.settings-tabs {
  display: flex;
  gap: 0;
  border-bottom: 2px solid var(--border-subtle);
  margin-bottom: 20px;
}

.tab {
  padding: 8px 18px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-tertiary);
  text-decoration: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: color 0.15s, border-color 0.15s;
}

.tab:hover {
  color: var(--text-primary);
}

.tab--active {
  color: var(--brand-500);
  border-bottom-color: var(--brand-500);
}

.settings__header h1 {
  font-size: 20px;
  margin: 0;
}

.meta {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 4px 0 0;
}

.panel {
  padding: 24px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--surface-1);
}

.panel__title {
  font-size: 15px;
  color: var(--text-primary);
  margin: 0 0 8px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.panel__desc {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 0 0 16px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-item--full {
  grid-column: 1 / -1;
}

.form-item__label {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.form-item input[type="text"],
.form-item textarea,
.form-item select {
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  background: var(--surface-1);
  outline: none;
  font-family: inherit;
}

.form-item input:focus,
.form-item textarea:focus,
.form-item select:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50);
}

.form-item textarea {
  resize: vertical;
  min-height: 64px;
}

.form-item__value {
  font-size: 13px;
  color: var(--text-primary);
  padding: 8px 0;
}

.mono {
  font-family: var(--font-mono);
}

.color-picker {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.color-swatch {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: transform 0.15s;
}

.color-swatch:hover {
  transform: scale(1.15);
}

.color-swatch.active {
  border-color: var(--text-primary);
  box-shadow: 0 0 0 2px var(--surface-1), 0 0 0 4px var(--text-primary);
}

.color-custom {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-tertiary);
  cursor: pointer;
}

.color-custom input[type="color"] {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 50%;
  padding: 0;
  cursor: pointer;
  appearance: none;
}

.color-custom input[type="color"]::-webkit-color-swatch-wrapper {
  padding: 0;
}

.color-custom input[type="color"]::-webkit-color-swatch {
  border-radius: 50%;
  border: 1px solid var(--border-default);
}

.actions {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.msg {
  font-size: 13px;
  margin: 0;
}

.msg.error { color: var(--danger-500); }
.msg.success { color: var(--success-500); }

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
  align-self: flex-start;
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.btn--primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ===== Toggle Grid ===== */
.toggle-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.toggle-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
  cursor: pointer;
  transition: border-color 0.15s;
}

.toggle-item:hover {
  border-color: var(--border-default);
}

.toggle-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.toggle-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.toggle-desc {
  font-size: 11px;
  color: var(--text-tertiary);
}

/* Toggle switch */
.toggle-switch {
  position: relative;
  width: 44px;
  height: 24px;
  appearance: none;
  background: var(--surface-3);
  border: 1px solid var(--border-default);
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
  flex-shrink: 0;
}

.toggle-switch:checked {
  background: var(--brand-500);
  border-color: var(--brand-500);
}

.toggle-switch::before {
  content: "";
  position: absolute;
  top: 2px;
  left: 2px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--text-on-brand);
  transition: transform 0.2s;
}

.toggle-switch:checked::before {
  transform: translateX(20px);
}

@media (max-width: 600px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
