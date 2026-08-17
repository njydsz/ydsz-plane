<template>
  <div class="tpl-page">
    <!-- 页头 -->
    <div class="tpl-header">
      <div class="tpl-header__left">
        <h2 class="tpl-title">内容模板库</h2>
        <p class="tpl-subtitle">管理需求/任务/缺陷模板，创建时从模板快速填充</p>
      </div>
      <div class="tpl-header__right">
        <button class="tpl-btn tpl-btn--primary" @click="openCreateModal">
          + 新建模板
        </button>
      </div>
    </div>

    <!-- 类型筛选 -->
    <div class="tpl-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="tpl-tab"
        :class="{ 'tpl-tab--active': activeType === tab.key }"
        @click="activeType = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="tpl-loading">
      <span class="tpl-spinner"></span>
      <span>加载中...</span>
    </div>

    <!-- 空状态 -->
    <div v-else-if="filteredTemplates.length === 0" class="tpl-empty">
      <p>暂无{{ typeLabel }}模板</p>
      <p class="tpl-empty__hint">点击"新建模板"按钮，将常用工作项结构保存为模板</p>
    </div>

    <!-- 模板列表 -->
    <div v-else class="tpl-grid">
      <div
        v-for="tpl in filteredTemplates"
        :key="tpl.id"
        class="tpl-card"
      >
        <div class="tpl-card__header">
          <span class="tpl-card__name">{{ tpl.name }}</span>
          <span v-if="tpl.is_default" class="tpl-card__badge">默认</span>
        </div>
        <div class="tpl-card__type">{{ typeLabel }}</div>
        <div v-if="tpl.content_html" class="tpl-card__preview" v-html="truncateHtml(tpl.content_html, 120)"></div>
        <div class="tpl-card__meta">
          创建于 {{ formatDate(tpl.created_at) }}
        </div>
        <div class="tpl-card__actions">
          <button class="tpl-btn tpl-btn--ghost tpl-btn--sm" @click="openEditModal(tpl)">
            编辑
          </button>
          <button class="tpl-btn tpl-btn--ghost tpl-btn--sm tpl-btn--danger" @click="confirmDelete(tpl)">
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- 创建/编辑弹窗 -->
    <div v-if="showModal" class="tpl-modal-overlay" @click.self="closeModal">
      <div class="tpl-modal">
        <div class="tpl-modal__header">
          <h3>{{ editingTpl ? "编辑模板" : "新建模板" }}</h3>
          <button class="tpl-modal__close" @click="closeModal">×</button>
        </div>
        <div class="tpl-modal__body">
          <div class="tpl-form-group">
            <label>模板名称 <span class="tpl-required">*</span></label>
            <input
              v-model="form.name"
              type="text"
              class="tpl-input"
              placeholder="如：标准需求模板、紧急缺陷模板"
              maxlength="255"
            />
          </div>
          <div class="tpl-form-group">
            <label>模板类型 <span class="tpl-required">*</span></label>
            <div class="tpl-type-select">
              <button
                v-for="t in types"
                :key="t.key"
                class="tpl-type-btn"
                :class="{ 'tpl-type-btn--active': form.template_type === t.key }"
                :disabled="editingTpl !== null"
                @click="form.template_type = t.key"
              >
                {{ t.label }}
              </button>
            </div>
          </div>
          <div class="tpl-form-group">
            <label>标题模板</label>
            <input
              v-model="form.content_json.name"
              type="text"
              class="tpl-input"
              placeholder="支持变量：{date} {project} {user}"
            />
          </div>
          <div class="tpl-form-group">
            <label>描述内容 (HTML)</label>
            <textarea
              v-model="form.content_html"
              class="tpl-textarea"
              rows="6"
              placeholder="模板描述内容，支持 HTML 格式"
            ></textarea>
          </div>
          <div class="tpl-form-group">
            <label>优先级</label>
            <select v-model="form.content_json.priority" class="tpl-input">
              <option value="">不设置</option>
              <option value="urgent">紧急</option>
              <option value="high">高</option>
              <option value="medium">中</option>
              <option value="low">低</option>
            </select>
          </div>
          <div class="tpl-form-group tpl-form-group--row">
            <label class="tpl-checkbox-label">
              <input v-model="form.is_default" type="checkbox" />
              <span>设为默认模板</span>
            </label>
          </div>
        </div>
        <div class="tpl-modal__footer">
          <button class="tpl-btn tpl-btn--ghost" @click="closeModal">取消</button>
          <button
            class="tpl-btn tpl-btn--primary"
            :disabled="saving || !form.name"
            @click="saveTpl"
          >
            {{ saving ? "保存中..." : "保存" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { contentTemplateApi, type ContentTemplate, type CreateTemplateRequest } from "@/api/services/contentTemplate";
import { toast } from "@/lib/toast";

const route = useRoute();
const wsId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const loading = ref(false);
const saving = ref(false);
const templates = ref<ContentTemplate[]>([]);
const activeType = ref("requirement");

const showModal = ref(false);
const editingTpl = ref<ContentTemplate | null>(null);

const tabs = [
  { key: "", label: "全部" },
  { key: "requirement", label: "需求" },
  { key: "task", label: "任务" },
  { key: "defect", label: "缺陷" },
];

const types = [
  { key: "requirement", label: "需求" },
  { key: "task", label: "任务" },
  { key: "defect", label: "缺陷" },
];

const form = ref<CreateTemplateRequest & { content_html: string }>({
  name: "",
  template_type: "requirement",
  content_json: {},
  content_html: "",
  is_default: false,
});

const filteredTemplates = computed(() => {
  if (!activeType.value) return templates.value;
  return templates.value.filter((t) => t.template_type === activeType.value);
});

const typeLabel = computed(() => {
  const t = types.find((x) => x.key === activeType.value);
  return t ? t.label : "";
});

async function load() {
  loading.value = true;
  try {
    templates.value = await contentTemplateApi.list(wsId, projectId, activeType.value || undefined);
  } catch (err) {
    toast.error(`加载模板失败: ${err instanceof Error ? err.message : "未知错误"}`);
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  form.value = {
    name: "",
    template_type: activeType.value || "requirement",
    content_json: {},
    content_html: "",
    is_default: false,
  };
}

function openCreateModal() {
  editingTpl.value = null;
  resetForm();
  showModal.value = true;
}

function openEditModal(tpl: ContentTemplate) {
  editingTpl.value = tpl;
  form.value = {
    name: tpl.name,
    template_type: tpl.template_type,
    content_json: { ...tpl.content_json },
    content_html: tpl.content_html || "",
    is_default: tpl.is_default,
  };
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  editingTpl.value = null;
  resetForm();
}

async function saveTpl() {
  if (!form.value.name) return;
  saving.value = true;
  try {
    if (editingTpl.value) {
      await contentTemplateApi.update(wsId, projectId, editingTpl.value.id, {
        name: form.value.name,
        content_json: form.value.content_json,
        content_html: form.value.content_html,
        is_default: form.value.is_default,
      });
      toast.success("模板已更新");
    } else {
      await contentTemplateApi.create(wsId, projectId, {
        name: form.value.name,
        template_type: form.value.template_type,
        content_json: form.value.content_json,
        content_html: form.value.content_html,
        is_default: form.value.is_default,
      });
      toast.success("模板已创建");
    }
    closeModal();
    await load();
  } catch (err) {
    toast.error(`保存失败: ${err instanceof Error ? err.message : "未知错误"}`);
  } finally {
    saving.value = false;
  }
}

async function confirmDelete(tpl: ContentTemplate) {
  if (!confirm(`确定删除模板"${tpl.name}"？`)) return;
  try {
    await contentTemplateApi.delete(wsId, projectId, tpl.id);
    toast.success("模板已删除");
    await load();
  } catch (err) {
    toast.error(`删除失败: ${err instanceof Error ? err.message : "未知错误"}`);
  }
}

function truncateHtml(html: string, maxLen: number): string {
  const text = html.replace(/<[^>]*>/g, "");
  if (text.length <= maxLen) return text;
  return text.slice(0, maxLen) + "...";
}

function formatDate(dateStr: string): string {
  return dateStr.slice(0, 10);
}

watch(activeType, load);

onMounted(load);
</script>

<style scoped>
.tpl-page {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}

/* ---- Header ---- */
.tpl-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}
.tpl-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}
.tpl-subtitle {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
}

/* ---- Tabs ---- */
.tpl-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border-default);
  padding-bottom: 0;
}
.tpl-tab {
  padding: 8px 16px;
  font-size: 13px;
  color: var(--text-secondary);
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
}
.tpl-tab:hover {
  color: var(--text-primary);
}
.tpl-tab--active {
  color: var(--brand-600);
  border-bottom-color: var(--brand-600);
  font-weight: 500;
}

/* ---- Buttons ---- */
.tpl-btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: all 0.15s;
}
.tpl-btn--primary {
  background: var(--brand-600);
  border-color: var(--brand-600);
  color: #fff;
}
.tpl-btn--primary:hover {
  background: var(--brand-700);
}
.tpl-btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.tpl-btn--ghost {
  background: transparent;
}
.tpl-btn--ghost:hover {
  background: var(--bg-secondary);
}
.tpl-btn--danger {
  color: var(--danger-600, #d32f2f);
}
.tpl-btn--danger:hover {
  background: var(--danger-50, #ffebee);
}
.tpl-btn--sm {
  padding: 4px 10px;
  font-size: 12px;
}

/* ---- Loading & Empty ---- */
.tpl-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 48px;
  color: var(--text-tertiary);
  font-size: 14px;
}
.tpl-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-default);
  border-top-color: var(--brand-500);
  border-radius: 50%;
  animation: tpl-spin 0.8s linear infinite;
}
@keyframes tpl-spin {
  to { transform: rotate(360deg); }
}
.tpl-empty {
  text-align: center;
  padding: 48px;
  color: var(--text-tertiary);
  font-size: 14px;
}
.tpl-empty__hint {
  font-size: 12px;
  margin-top: 8px;
}

/* ---- Grid ---- */
.tpl-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}
.tpl-card {
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 16px;
  background: var(--bg-primary);
  transition: box-shadow 0.15s, border-color 0.15s;
}
.tpl-card:hover {
  border-color: var(--border-hover);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}
.tpl-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.tpl-card__name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.tpl-card__badge {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  background: var(--brand-100);
  color: var(--brand-700);
  border-radius: 10px;
}
.tpl-card__type {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
}
.tpl-card__preview {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 12px;
  max-height: 60px;
  overflow: hidden;
  line-height: 1.5;
}
.tpl-card__meta {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-bottom: 12px;
}
.tpl-card__actions {
  display: flex;
  gap: 8px;
}

/* ---- Modal ---- */
.tpl-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.tpl-modal {
  background: var(--bg-primary);
  border-radius: 12px;
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
}
.tpl-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-default);
}
.tpl-modal__header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}
.tpl-modal__close {
  font-size: 20px;
  color: var(--text-tertiary);
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0 4px;
}
.tpl-modal__close:hover {
  color: var(--text-primary);
}
.tpl-modal__body {
  padding: 20px;
}
.tpl-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-default);
}

/* ---- Form ---- */
.tpl-form-group {
  margin-bottom: 16px;
}
.tpl-form-group label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 6px;
}
.tpl-form-group--row {
  display: flex;
  align-items: center;
}
.tpl-form-group--row label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}
.tpl-required {
  color: var(--danger-600, #d32f2f);
}
.tpl-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
  background: var(--bg-primary);
  box-sizing: border-box;
}
.tpl-input:focus {
  outline: none;
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
}
.tpl-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
  background: var(--bg-primary);
  resize: vertical;
  box-sizing: border-box;
  font-family: inherit;
}
.tpl-textarea:focus {
  outline: none;
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
}
.tpl-type-select {
  display: flex;
  gap: 8px;
}
.tpl-type-btn {
  padding: 6px 14px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
  background: var(--bg-primary);
  cursor: pointer;
  transition: all 0.15s;
}
.tpl-type-btn--active {
  background: var(--brand-600);
  border-color: var(--brand-600);
  color: #fff;
}
.tpl-type-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.tpl-checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
}
</style>
