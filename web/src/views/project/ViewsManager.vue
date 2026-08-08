<script setup lang="ts">
/**
 * 视图管理侧边面板 — 保存/加载/编辑/删除命名视图。
 *
 * 功能：
 * 1. 视图列表：显示当前项目下所有可用视图（我的视图/团队视图/默认视图）
 * 2. 保存当前视图：保存当前过滤/排序/分组/列配置为命名视图
 * 3. 加载视图：点击视图名称切换视图配置
 * 4. 编辑/删除：hover 显示编辑/删除按钮（仅 owner）
 * 5. 设为默认：管理员可设置团队默认视图
 */
import { computed, onMounted, ref } from "vue";
import { viewsApi, type SavedView, type ViewScope, type CreateViewInput, type PreferenceViewType } from "@/api/services/preference";
import { toast } from "@/lib/toast";

export interface ViewConfig {
  filters?: unknown;
  sort?: unknown;
  groupBy?: unknown;
  columns?: unknown;
  extra?: unknown;
}

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  /** 当前视图类型（list/kanban/gantt/calendar/spreadsheet） */
  viewType: PreferenceViewType;
  /** 当前活跃的视图配置 */
  currentConfig: ViewConfig;
  /** 当前活跃视图 ID（加载自某个已保存视图时） */
  activeViewId?: number | null;
  /** 当前用户 ID（用于判断 owner 权限） */
  currentUserId?: number | null;
  /** 当前用户是否为管理员 */
  isAdmin?: boolean;
}>();

const emit = defineEmits<{
  (e: "load-view", view: SavedView): void;
  (e: "save-view"): void;
}>();

// ---- 状态 ----
const views = ref<SavedView[]>([]);
const loading = ref(true);
const showCreateModal = ref(false);
const editingView = ref<SavedView | null>(null);

// 创建/编辑表单
const formName = ref("");
const formScope = ref<ViewScope>("personal");
const formShared = ref(false);
const saving = ref(false);

// 分组视图
const personalViews = computed(() => views.value.filter((v) => v.scope === "personal" && v.owner_id === props.currentUserId));
const teamViews = computed(() => views.value.filter((v) => v.is_shared || v.scope === "team"));
const defaultViews = computed(() => views.value.filter((v) => v.scope === "default"));

// ---- 方法 ----
async function loadViews() {
  loading.value = true;
  try {
    views.value = await viewsApi.list(props.workspaceId, props.projectId);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "加载视图列表失败");
  } finally {
    loading.value = false;
  }
}

/** 打开创建视图弹窗 */
function openCreateModal() {
  editingView.value = null;
  formName.value = "";
  formScope.value = "personal";
  formShared.value = false;
  showCreateModal.value = true;
}

/** 打开编辑视图弹窗 */
function openEditModal(view: SavedView) {
  editingView.value = view;
  formName.value = view.name;
  formScope.value = view.scope;
  formShared.value = view.is_shared;
  showCreateModal.value = true;
}

/** 提交创建/编辑 */
async function submitForm() {
  if (!formName.value.trim()) {
    toast.error("视图名称不能为空");
    return;
  }
  saving.value = true;
  try {
    if (editingView.value) {
      // 编辑现有视图
      await viewsApi.update(
        props.workspaceId,
        props.projectId,
        editingView.value.id,
        {
          name: formName.value.trim(),
          scope: formScope.value,
          is_shared: formShared.value,
        },
      );
      toast.success("视图已更新");
    } else {
      // 创建新视图
      const input: CreateViewInput = {
        name: formName.value.trim(),
        type: props.viewType,
        scope: formScope.value,
        config: props.currentConfig as Record<string, unknown>,
        is_shared: formShared.value,
      };
      await viewsApi.create(props.workspaceId, props.projectId, input);
      toast.success("视图已保存");
    }
    showCreateModal.value = false;
    loadViews();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存视图失败");
  } finally {
    saving.value = false;
  }
}

/** 应用视图配置 */
function applyView(view: SavedView) {
  emit("load-view", view);
}

/** 删除视图 */
async function deleteView(view: SavedView) {
  if (!confirm(`确定要删除视图「${view.name}」吗？此操作不可撤销。`)) return;
  try {
    await viewsApi.delete(props.workspaceId, props.projectId, view.id);
    toast.success("视图已删除");
    loadViews();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除视图失败");
  }
}

/** 设为默认 */
async function setDefaultView(view: SavedView) {
  try {
    await viewsApi.setDefault(props.workspaceId, props.projectId, view.id);
    toast.success(`已将「${view.name}」设为默认视图`);
    loadViews();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "设置默认视图失败");
  }
}

/** 保存当前视图配置（触发父组件收集当前状态） */
function saveCurrentView() {
  emit("save-view");
  openCreateModal();
}

const scopeLabel = (s: ViewScope) =>
  ({ personal: "个人", team: "团队", default: "默认" } as Record<string, string>)[s] ?? s;

onMounted(loadViews);
</script>

<template>
  <div class="views-manager">
    <!-- 工具栏 -->
    <div class="views-manager__toolbar">
      <button class="btn btn--sm btn--primary" @click="saveCurrentView">
        保存视图
      </button>
    </div>

    <!-- 视图列表 -->
    <div v-if="loading" class="views-manager__loading">加载中...</div>

    <template v-else>
      <!-- 默认视图 -->
      <div v-if="defaultViews.length > 0" class="views-group">
        <div class="views-group__title">默认视图</div>
        <ul class="views-list">
          <li
            v-for="v in defaultViews"
            :key="v.id"
            class="view-item"
            :class="{ 'view-item--active': activeViewId === v.id }"
            @click="applyView(v)"
          >
            <span class="view-item__name">{{ v.name }}</span>
            <span class="view-item__badge">默认</span>
            <div v-if="isAdmin" class="view-item__actions">
              <button class="action-btn" title="取消默认" @click.stop="deleteView(v)">×</button>
            </div>
          </li>
        </ul>
      </div>

      <!-- 团队共享视图 -->
      <div v-if="teamViews.length > 0" class="views-group">
        <div class="views-group__title">团队视图</div>
        <ul class="views-list">
          <li
            v-for="v in teamViews"
            :key="v.id"
            class="view-item"
            :class="{ 'view-item--active': activeViewId === v.id }"
            @click="applyView(v)"
          >
            <span class="view-item__name">{{ v.name }}</span>
            <span class="view-item__scope">{{ scopeLabel(v.scope) }}</span>
            <!-- hover 操作按钮（仅 owner 可见编辑/删除，管理员可见设默认） -->
            <div class="view-item__actions">
              <template v-if="v.owner_id === currentUserId">
                <button class="action-btn" title="编辑" @click.stop="openEditModal(v)">✎</button>
                <button class="action-btn" title="删除" @click.stop="deleteView(v)">×</button>
              </template>
              <button v-if="isAdmin" class="action-btn" title="设为默认" @click.stop="setDefaultView(v)">★</button>
            </div>
          </li>
        </ul>
      </div>

      <!-- 个人视图 -->
      <div class="views-group">
        <div class="views-group__title">我的视图</div>
        <ul class="views-list">
          <li
            v-for="v in personalViews"
            :key="v.id"
            class="view-item"
            :class="{ 'view-item--active': activeViewId === v.id }"
            @click="applyView(v)"
          >
            <span class="view-item__name">{{ v.name }}</span>
            <span v-if="v.is_shared" class="view-item__badge">已共享</span>
            <div class="view-item__actions">
              <button class="action-btn" title="编辑" @click.stop="openEditModal(v)">✎</button>
              <button class="action-btn" title="删除" @click.stop="deleteView(v)">×</button>
            </div>
          </li>
          <li v-if="personalViews.length === 0 && teamViews.length === 0 && defaultViews.length === 0" class="view-item view-item--empty">
            <span class="view-item__name">暂无保存的视图</span>
          </li>
        </ul>
      </div>
    </template>

    <!-- 创建/编辑弹窗 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-box view-modal">
        <h3>{{ editingView ? '编辑视图' : '保存视图' }}</h3>
        <div class="form-group">
          <label class="form-label">视图名称</label>
          <input
            v-model="formName"
            class="form-input"
            type="text"
            placeholder="输入视图名称"
            maxlength="128"
            @keyup.enter="submitForm"
          />
        </div>
        <div class="form-group">
          <label class="form-label">范围</label>
          <select v-model="formScope" class="form-select">
            <option value="personal">个人</option>
            <option value="team">团队</option>
          </select>
        </div>
        <div class="form-group form-check">
          <label class="form-check-label">
            <input v-model="formShared" type="checkbox" />
            分享给团队
          </label>
        </div>
        <p v-if="formScope === 'team' || formShared" class="form-hint">团队视图对项目内所有成员可见，其他成员可通过视图切换面板加载此视图配置。</p>
        <div class="modal-actions">
          <button class="btn btn--ghost" @click="showCreateModal = false">取消</button>
          <button class="btn btn--primary" :disabled="saving || !formName.trim()" @click="submitForm">
            {{ saving ? '保存中...' : (editingView ? '保存' : '创建') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.views-manager {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.views-manager__toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.views-manager__loading {
  font-size: 13px;
  color: var(--text-tertiary);
  padding: 16px 0;
  text-align: center;
}

/* 分组 */
.views-group {
  margin-bottom: 4px;
}

.views-group__title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  padding: 4px 0;
}

.views-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.view-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--text-primary);
  transition: background 0.1s;
}

.view-item:hover {
  background: var(--surface-2);
}

.view-item--active {
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 500;
}

.view-item--empty {
  color: var(--text-tertiary);
  cursor: default;
}

.view-item--empty:hover {
  background: none;
}

.view-item__name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.view-item__scope {
  font-size: 10px;
  color: var(--text-tertiary);
  background: var(--surface-2);
  padding: 1px 6px;
  border-radius: 3px;
  flex-shrink: 0;
}

.view-item__badge {
  font-size: 10px;
  color: var(--brand-500);
  background: var(--brand-50);
  padding: 1px 6px;
  border-radius: 3px;
  flex-shrink: 0;
}

.view-item__actions {
  display: none;
  gap: 2px;
  flex-shrink: 0;
}

.view-item:hover .view-item__actions {
  display: flex;
}

.action-btn {
  width: 20px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  font-size: 12px;
  font-family: inherit;
  color: var(--text-tertiary);
  background: none;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.1s;
}

.action-btn:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

/* 按钮 */
.btn {
  padding: 5px 12px;
  font-size: 12px;
  font-family: inherit;
  border-radius: var(--radius-sm);
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-primary);
}

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
}

.btn--primary {
  background: var(--brand-500);
  color: var(--text-on-brand);
  border-color: var(--brand-500);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--ghost {
  background: none;
  color: var(--text-secondary);
  border-color: transparent;
}

.btn--ghost:hover {
  background: var(--surface-3);
}

/* 弹窗 */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-backdrop);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-box {
  background: var(--surface-1);
  padding: 24px;
  border-radius: var(--radius-md);
  max-width: 400px;
  width: 90%;
}

.modal-box h3 {
  margin: 0 0 16px;
  font-size: 16px;
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 16px;
}

/* 表单 */
.form-group {
  margin-bottom: 12px;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.form-input,
.form-select {
  width: 100%;
  padding: 6px 10px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
  box-sizing: border-box;
}

.form-input:focus,
.form-select:focus {
  border-color: var(--brand-500);
}

.form-check {
  display: flex;
  align-items: center;
}

.form-check-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}

.form-check-label input {
  accent-color: var(--brand-500);
}

.form-hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 8px 0 0;
}
</style>
