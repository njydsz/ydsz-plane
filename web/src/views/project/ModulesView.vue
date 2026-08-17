<script setup lang="ts">
/**
 * ModulesView — 模块管理页（独立路由入口，区别于 ProjectSettings 子页）。
 *
 * 功能：
 *  - 列表卡片展示项目模块（名称 / 描述 / 负责人 / 关联工作项数）
 *  - 新建 / 编辑弹窗（ModuleFormModal）
 *  - 删除二次确认（AppModal style overlay）
 *  - 软删除归档（status → archived）
 *
 * 注意：模块 schema 中不含 target_version_id，PRD 提到的"目标版本"列暂缺。
 */
import { computed, onMounted, ref } from "vue";

import { moduleApi, type CreateModuleInput, type Module, type UpdateModuleInput } from "@/api/services/module";
import { workspaceApi, type Member } from "@/api/services/workspace";
import { AppButton, AppCard, AppEmptyState, AppErrorState, AppLoadingState } from "@/components";
import { toast } from "@/lib/toast";

import ModuleFormModal from "./ModuleFormModal.vue";

// ---- route props ----
const props = defineProps<{
  workspaceId: number;
  projectId: number;
}>();

// ---- 数据 ----
const modules = ref<Module[]>([]);
const members = ref<Member[]>([]);
const loading = ref(true);
const error = ref("");

function memberName(id: number | undefined): string {
  if (!id) return "—";
  const m = members.value.find((x) => x.id === id);
  return m?.display_name ?? `ID:${id}`;
}

// ---- 加载 ----
async function loadAll() {
  loading.value = true;
  error.value = "";
  try {
    const [mods, mems] = await Promise.all([
      moduleApi.list(props.workspaceId, props.projectId),
      workspaceApi.listMembers(props.workspaceId),
    ]);
    modules.value = mods;
    members.value = mems;
  } catch (e: any) {
    error.value = e?.message ?? "加载模块列表失败";
  } finally {
    loading.value = false;
  }
}

onMounted(loadAll);

// ---- 弹窗 ----
const showForm = ref(false);
const editingModule = ref<Module | null>(null);

function openCreate() {
  editingModule.value = null;
  showForm.value = true;
}

function openEdit(mod: Module) {
  editingModule.value = mod;
  showForm.value = true;
}

async function handleSubmit(payload: CreateModuleInput | UpdateModuleInput) {
  try {
    if (editingModule.value) {
      await moduleApi.update(props.workspaceId, props.projectId, editingModule.value.id, payload as UpdateModuleInput);
      toast.success("模块已更新");
    } else {
      await moduleApi.create(props.workspaceId, props.projectId, payload as CreateModuleInput);
      toast.success("模块已创建");
    }
    showForm.value = false;
    await loadAll();
  } catch (e: any) {
    toast.error(e?.message ?? "保存失败");
  }
}

// ---- 删除确认 ----
const deleteTarget = ref<Module | null>(null);
const deleting = ref(false);

function askDelete(mod: Module) {
  deleteTarget.value = mod;
}

async function confirmDelete() {
  if (!deleteTarget.value) return;
  deleting.value = true;
  try {
    await moduleApi.remove(props.workspaceId, props.projectId, deleteTarget.value.id);
    toast.success("模块已删除");
    deleteTarget.value = null;
    await loadAll();
  } catch (e: any) {
    toast.error(e?.message ?? "删除失败");
  } finally {
    deleting.value = false;
  }
}

// ---- 归档（软删除） ----
const archivingId = ref<number | null>(null);

async function toggleArchive(mod: Module) {
  archivingId.value = mod.id;
  try {
    const newStatus = mod.status === "archived" ? "active" : "archived";
    await moduleApi.update(props.workspaceId, props.projectId, mod.id, { status: newStatus });
    toast.success(newStatus === "archived" ? "模块已归档" : "模块已恢复");
    await loadAll();
  } catch (e: any) {
    toast.error(e?.message ?? "操作失败");
  } finally {
    archivingId.value = null;
  }
}

// ---- 工作项计数（后端 issue_count 字段；缺失时显示 —） ----
const activeCount = computed(() => modules.value.filter((m) => m.status !== "archived").length);
</script>

<template>
  <div class="modules-view">
    <!-- 顶部标题 + 操作栏 -->
    <header class="modules-view__header">
      <div>
        <h1 class="modules-view__title">模块管理</h1>
        <p class="modules-view__subtitle">
          管理项目模块，为工作项划分分类维度 · 共 {{ modules.length }} 个模块（{{ activeCount }} 个启用）
        </p>
      </div>
      <AppButton variant="primary" size="sm" @click="openCreate">
        ＋ 新建模块
      </AppButton>
    </header>

    <!-- 加载态 -->
    <AppLoadingState v-if="loading" />

    <!-- 错误态 -->
    <AppErrorState
      v-else-if="error"
      :message="error"
      @retry="loadAll"
    />

    <!-- 表格卡片 -->
    <AppCard v-else padding="none" class="table-card">
      <!-- 空态 -->
      <AppEmptyState
        v-if="modules.length === 0"
        scenario="modules"
        cta-text="新建第一个模块"
        class="table-card__empty"
        @cta-click="openCreate"
      />

      <!-- 表格 -->
      <div v-else class="table-wrap">
        <table class="module-table">
          <thead>
            <tr>
              <th class="col--name">模块名称</th>
              <th class="col--desc">描述</th>
              <th class="col--lead">负责人</th>
              <th class="col--count">关联工作项数</th>
              <th class="col--status">状态</th>
              <th class="col--actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="mod in modules"
              :key="mod.id"
              :class="{ 'row--archived': mod.status === 'archived' }"
            >
              <!-- 名称 -->
              <td class="cell--name">
                <span class="module-name">{{ mod.name }}</span>
              </td>

              <!-- 描述 -->
              <td class="cell--desc">
                <span :title="mod.description || ''">
                  {{ mod.description || "—" }}
                </span>
              </td>

              <!-- 负责人 -->
              <td class="cell--lead">
                <span class="lead-badge">
                  <span class="lead-avatar">{{ memberName(mod.lead_id).charAt(0) }}</span>
                  {{ memberName(mod.lead_id) }}
                </span>
              </td>

              <!-- 关联工作项数 -->
              <td class="cell--count">
                {{ mod.issue_count ?? "—" }}
              </td>

              <!-- 状态 -->
              <td class="cell--status">
                <span
                  class="status-tag"
                  :class="mod.status === 'archived' ? 'status--archived' : 'status--active'"
                >
                  {{ mod.status === "archived" ? "已归档" : "启用中" }}
                </span>
              </td>

              <!-- 操作 -->
              <td class="cell--actions">
                <button
                  class="op-btn"
                  title="编辑"
                  @click="openEdit(mod)"
                >
                  编辑
                </button>
                <button
                  class="op-btn"
                  :disabled="archivingId === mod.id"
                  :title="mod.status === 'archived' ? '恢复' : '归档'"
                  @click="toggleArchive(mod)"
                >
                  {{ archivingId === mod.id ? "处理中…" : mod.status === "archived" ? "恢复" : "归档" }}
                </button>
                <button
                  class="op-btn op-btn--danger"
                  title="删除"
                  @click="askDelete(mod)"
                >
                  删除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AppCard>

    <!-- 新建/编辑弹窗 -->
    <ModuleFormModal
      :visible="showForm"
      :module="editingModule"
      :members="members"
      @close="showForm = false"
      @submit="handleSubmit"
    />

    <!-- 删除确认弹窗 -->
    <Teleport to="body">
      <div
        v-if="deleteTarget"
        class="confirm-overlay"
        @click.self="deleteTarget = null"
      >
        <div
          class="confirm-dialog"
          role="alertdialog"
          aria-modal="true"
          aria-labelledby="delete-confirm-title"
          aria-describedby="delete-confirm-desc"
        >
          <header class="confirm-dialog__header">
            <h2 id="delete-confirm-title" class="confirm-dialog__title">确认删除模块</h2>
            <button
              class="confirm-dialog__close"
              aria-label="关闭"
              @click="deleteTarget = null"
            >
              &times;
            </button>
          </header>
          <div class="confirm-dialog__body">
            <p id="delete-confirm-desc" class="confirm-dialog__text">
              确定要删除模块
              <strong>「{{ deleteTarget.name }}」</strong>
              吗？删除后不可撤销。关联工作项的模块引用会自动清除。
            </p>
          </div>
          <footer class="confirm-dialog__footer">
            <button
              type="button"
              class="btn btn--secondary"
              @click="deleteTarget = null"
            >
              取消
            </button>
            <button
              type="button"
              class="btn btn--danger"
              :disabled="deleting"
              @click="confirmDelete"
            >
              {{ deleting ? "删除中…" : "确认删除" }}
            </button>
          </footer>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.modules-view {
  max-width: 960px;
}

/* ---------- header ---------- */
.modules-view__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.modules-view__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.modules-view__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

/* ---------- card ---------- */
.table-card {
  overflow: hidden;
}

.table-card__empty {
  padding: 60px 24px;
}

/* ---------- table ---------- */
.table-wrap {
  overflow-x: auto;
}

.module-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.module-table th {
  text-align: left;
  padding: 10px 16px;
  color: var(--text-tertiary);
  font-weight: 500;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
  background: var(--surface-1);
}

.module-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
  vertical-align: middle;
}

.module-table tbody tr:hover {
  background: var(--surface-hover, #f9fafb);
}

.module-table tbody tr:last-child td {
  border-bottom: none;
}

.row--archived {
  opacity: 0.55;
}

/* column widths */
.col--name       { width: 18%; }
.col--desc       { width: 28%; max-width: 240px; }
.col--lead       { width: 14%; }
.col--count      { width: 12%; text-align: center; }
.col--status     { width: 10%; }
.col--actions    { width: 18%; }

.cell--count {
  text-align: center;
  font-variant-numeric: tabular-nums;
}

/* ---------- name ---------- */
.module-name {
  font-weight: 500;
}

.cell--desc {
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 240px;
}

/* ---------- lead badge ---------- */
.lead-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.lead-avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--brand-100, #dbe4ff);
  color: var(--brand-600, #3b82f6);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

/* ---------- status tag ---------- */
.status-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
}

.status--active {
  background: var(--success-50, #e6f9f0);
  color: var(--success-600, #0c8c4a);
}

.status--archived {
  background: var(--text-tertiary-bg, #f0f0f0);
  color: var(--text-tertiary);
}

/* ---------- actions ---------- */
.cell--actions {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.op-btn {
  padding: 4px 10px;
  font-size: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  font-family: inherit;
}

.op-btn:hover:not(:disabled) {
  background: var(--surface-hover, #f9fafb);
  color: var(--text-primary);
}

.op-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.op-btn--danger {
  color: var(--danger-500, #ef4444);
  border-color: var(--danger-200, #fecaca);
}

.op-btn--danger:hover:not(:disabled) {
  background: var(--danger-50, #fef2f2);
  color: var(--danger-600, #dc2626);
}

/* ---------- delete confirm ---------- */
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
  animation: fadeIn 0.15s ease;
}

.confirm-dialog {
  width: calc(100% - 32px);
  max-width: 420px;
  background: var(--surface-1);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-popover);
  animation: slideUp 0.2s ease;
}

.confirm-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.confirm-dialog__title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.confirm-dialog__close {
  background: none;
  border: none;
  font-size: 22px;
  color: var(--text-tertiary);
  cursor: pointer;
  line-height: 1;
  padding: 0;
}

.confirm-dialog__close:hover {
  color: var(--text-primary);
}

.confirm-dialog__body {
  padding: 16px 24px;
}

.confirm-dialog__text {
  margin: 0;
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.6;
}

.confirm-dialog__footer {
  padding: 0 24px 20px;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* ---------- buttons ---------- */
.btn {
  padding: 7px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: background 0.15s, border-color 0.15s, opacity 0.15s;
  font-family: inherit;
}

.btn--secondary {
  background: var(--surface-1);
  border-color: var(--border-default);
  color: var(--text-secondary);
}

.btn--secondary:hover {
  background: var(--surface-hover, #f9fafb);
}

.btn--danger {
  background: var(--danger-500, #ef4444);
  border-color: var(--danger-500, #ef4444);
  color: #fff;
}

.btn--danger:hover:not(:disabled) {
  background: var(--danger-600, #dc2626);
}

.btn--danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ---------- animations ---------- */
@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(10px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* ---------- responsive ---------- */
@media (max-width: 768px) {
  .modules-view__header {
    flex-direction: column;
  }

  .col--desc,
  .cell--desc {
    display: none;
  }
}
</style>
