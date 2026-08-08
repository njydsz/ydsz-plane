<script setup lang="ts">
/**
 * ProjectSettingsView — 项目设置页。
 *
 * 展示并编辑项目基本信息（名称、描述、标识、网络类型等）。
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { workspaceApi, type Project, type ProjectModuleToggles } from "@/api/services/workspace";
import { ApiError } from "@/api/client";
import { AppLoadingState, AppErrorState } from "@/components";
import { useWorkspaceStore } from "@/stores/workspace";

const route = useRoute();
const router = useRouter();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const wsStore = useWorkspaceStore();
const myRole = computed(() => wsStore.currentRole);
const canManage = computed(() => ["owner", "admin"].includes(myRole.value));
const isOwner = computed(() => myRole.value === "owner");

// 删除二次确认弹窗
const showDeleteModal = ref(false);
const deleteConfirmInput = ref("");
const archiveLoading = ref(false);
const deleteLoading = ref(false);

const project = ref<Project | null>(null);
const loading = ref(true);
const error = ref("");

const form = reactive({
  name: "",
  description: "",
  network: "public",
  color: "#3f63f1",
  coverImageUrl: "",
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
const coverImgError = ref(false);

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
    form.coverImageUrl = project.value.cover_image_url ?? "";
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
      cover_image_url: form.coverImageUrl.trim() || undefined,
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

// === Danger Zone: 归档 & 删除 ===
async function archiveProject() {
  if (!window.confirm(
    "确定要归档该项目？\n\n" +
    "项目归档后将从项目列表中隐藏，但成员仍可通过链接访问内容，且随时可恢复。"
  )) return;
  archiveLoading.value = true;
  try {
    const ws = await workspaceApi.get(workspaceId);
    await workspaceApi.archiveProject(ws.id, projectId);
    router.push(`/${workspaceId}`);
  } catch (e: any) {
    alert(`归档失败：${e.message ?? "未知错误"}`);
  } finally {
    archiveLoading.value = false;
  }
}

function openDeleteModal() {
  deleteConfirmInput.value = "";
  showDeleteModal.value = true;
}

function closeDeleteModal() {
  if (deleteLoading.value) return;
  showDeleteModal.value = false;
  deleteConfirmInput.value = "";
}

async function confirmDelete() {
  if (deleteConfirmInput.value !== project.value?.name) return;
  deleteLoading.value = true;
  try {
    const ws = await workspaceApi.get(workspaceId);
    await workspaceApi.archiveProject(ws.id, projectId);
    showDeleteModal.value = false;
    router.push(`/${workspaceId}`);
  } catch (e: any) {
    alert(`删除失败：${e.message ?? "未知错误"}`);
  } finally {
    deleteLoading.value = false;
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

    <!-- 封面图 -->
    <section class="panel" style="margin-top: 24px">
      <h2 class="panel__title">封面图</h2>
      <p class="panel__desc">设置项目封面图片 URL，将在项目卡片和项目仪表盘顶部展示。</p>
      <div class="cover-form">
        <label class="form-item form-item--full">
          <span class="form-item__label">封面图片 URL</span>
          <input
            v-model="form.coverImageUrl"
            type="text"
            placeholder="https://example.com/cover.jpg"
            maxlength="500"
          />
        </label>
        <div class="cover-preview">
          <div v-if="form.coverImageUrl.trim()" class="cover-preview__img-wrap">
            <img
              :src="form.coverImageUrl.trim()"
              alt="封面预览"
              class="cover-preview__img"
              @error="coverImgError = true"
              @load="coverImgError = false"
            />
            <span v-if="coverImgError" class="cover-preview__error">图片加载失败，请检查 URL 是否正确</span>
          </div>
          <div v-else class="cover-preview__placeholder">暂无封面图</div>
        </div>
        <div class="cover-actions">
          <button class="btn btn--secondary" type="button" @click="form.coverImageUrl = ''">清除封面</button>
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
            <span class="toggle-name">版本 (Version)</span>
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

    <!-- ========== 危险操作 Danger Zone ========== -->
    <section v-if="canManage" class="settings-section settings-section--danger">
      <h3 class="settings-section__title">危险操作</h3>
      <p class="settings-section__hint">以下操作影响整个项目，请谨慎操作。</p>

      <div class="danger-row">
        <div class="danger-row__info">
          <strong>归档项目</strong>
          <p>项目归档后将从列表中隐藏，但成员仍可通过链接访问。随时可恢复。</p>
        </div>
        <button
          class="btn btn--danger"
          :disabled="!canManage || archiveLoading"
          @click="archiveProject"
        >
          {{ archiveLoading ? "归档中..." : "归档" }}
        </button>
      </div>

      <div class="danger-row">
        <div class="danger-row__info">
          <strong>永久删除</strong>
          <p>项目及其所有工作项、文档等将被永久删除，<strong>不可恢复</strong>。</p>
        </div>
        <button
          class="btn btn--danger"
          :disabled="!canManage || deleteLoading"
          @click="openDeleteModal"
        >
          永久删除
        </button>
      </div>
    </section>
  </div>
  </div>

  <!-- ========== 删除确认弹窗 ========== -->
  <div
    v-if="showDeleteModal"
    class="modal-overlay"
    role="presentation"
    @click.self="closeDeleteModal"
    @keydown.esc="closeDeleteModal"
  >
    <div
      class="modal-dialog"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="delete-modal-title"
      aria-describedby="delete-modal-desc"
    >
      <h2 id="delete-modal-title" class="modal-dialog__title">永久删除项目</h2>
      <div id="delete-modal-desc" class="modal-dialog__body">
        <p class="modal-dialog__warn">
          <strong>此操作不可恢复。</strong>
          项目 <code>{{ project?.name }}</code> 及其所有工作项、文档、迭代等将被永久删除。
        </p>
        <p class="modal-dialog__hint">
          请输入项目名称
          <code class="modal-dialog__code">{{ project?.name }}</code>
          以确认删除：
        </p>
        <input
          v-model="deleteConfirmInput"
          type="text"
          class="modal-dialog__input"
          :placeholder="project?.name"
          :disabled="deleteLoading"
          @keydown.enter="confirmDelete"
        />
      </div>
      <div class="modal-dialog__footer">
        <button class="btn" :disabled="deleteLoading" @click="closeDeleteModal">
          取消
        </button>
        <button
          class="btn btn--danger"
          :disabled="deleteConfirmInput !== project?.name || deleteLoading"
          @click="confirmDelete"
        >
          {{ deleteLoading ? "删除中..." : "确认永久删除" }}
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

/* ===== Cover image ===== */
.cover-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cover-preview {
  width: 100%;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
  background: var(--surface-2);
}

.cover-preview__img-wrap {
  position: relative;
  width: 100%;
  height: 160px;
}

.cover-preview__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-preview__error {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 8px 12px;
  background: var(--danger-500);
  color: var(--text-on-brand);
  font-size: 12px;
}

.cover-preview__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100px;
  color: var(--text-tertiary);
  font-size: 13px;
}

.cover-actions {
  display: flex;
  gap: 8px;
}

.btn--secondary {
  background: var(--surface-2);
  border-color: var(--border-default);
  color: var(--text-secondary);
}

.btn--secondary:hover {
  background: var(--surface-3);
}

@media (max-width: 600px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}

/* ===== Danger Zone ===== */
.settings-section--danger {
  margin-top: 32px;
  padding: 24px;
  border: 1px solid var(--danger-200, rgba(220, 47, 47, 0.2));
  border-radius: var(--radius-lg);
  background: var(--danger-50, rgba(220, 47, 47, 0.03));
}

.settings-section__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--danger-600, #dc2f2f);
  margin: 0 0 4px;
}

.settings-section__hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0 0 16px;
}

.danger-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 0;
  border-top: 1px solid var(--danger-100, rgba(220, 47, 47, 0.1));
}

.danger-row:first-of-type {
  border-top: none;
}

.danger-row__info {
  flex: 1;
}

.danger-row__info strong {
  display: block;
  font-size: 13px;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.danger-row__info p {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0;
}

.btn--danger {
  background: var(--danger-500, #dc2f2f);
  border-color: var(--danger-500, #dc2f2f);
  color: #fff;
  flex-shrink: 0;
}

.btn--danger:hover:not(:disabled) {
  background: var(--danger-600, #b91c1c);
  border-color: var(--danger-600, #b91c1c);
}

.btn--danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== Delete Confirm Modal ===== */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(2px);
}

.modal-dialog {
  width: 440px;
  max-width: calc(100vw - 32px);
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.18);
  overflow: hidden;
}

.modal-dialog__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--danger-600, #dc2f2f);
  margin: 0;
  padding: 18px 20px 12px;
}

.modal-dialog__body {
  padding: 0 20px 16px;
}

.modal-dialog__warn {
  font-size: 13px;
  color: var(--text-primary);
  margin: 0 0 12px;
  line-height: 1.6;
}

.modal-dialog__warn code {
  padding: 1px 5px;
  background: var(--surface-3);
  border-radius: 3px;
  font-size: 12px;
  font-family: var(--font-mono);
}

.modal-dialog__hint {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0 0 10px;
}

.modal-dialog__code {
  padding: 1px 5px;
  background: var(--danger-50, rgba(220, 47, 47, 0.06));
  border: 1px solid var(--danger-200, rgba(220, 47, 47, 0.15));
  border-radius: 3px;
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--danger-600, #dc2f2f);
}

.modal-dialog__input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  background: var(--surface-1);
  outline: none;
  font-family: inherit;
  box-sizing: border-box;
}

.modal-dialog__input:focus {
  border-color: var(--danger-500, #dc2f2f);
  box-shadow: 0 0 0 3px var(--danger-50, rgba(220, 47, 47, 0.08));
}

.modal-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 20px;
  border-top: 1px solid var(--border-subtle);
  background: var(--surface-2);
}

@media (max-width: 600px) {
  .settings-section--danger {
    padding: 16px;
  }
  .danger-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .btn--danger {
    align-self: flex-end;
  }
}
</style>
