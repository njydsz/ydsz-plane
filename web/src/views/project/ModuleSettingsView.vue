<script setup lang="ts">
/**
 * ModuleSettingsView — 模块管理页。
 *
 * 展示项目下全部模块的列表，支持新建、行内编辑、删除、状态切换等操作。
 * 模块负责人可从工作空间成员中搜索选择。
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import { moduleApi, type CreateModuleInput, type Module, type UpdateModuleInput } from "@/api/services/module";
import { workspaceApi, type Member } from "@/api/services/workspace";
import { AppErrorState, AppLoadingState, AppModal } from "@/components";
import { toast } from "@/lib/toast";

/** 模板中使用 setTimeout 时需显式暴露，避免组件实例类型上找不到该全局函数。 */
const setTimeout = (fn: () => void, ms: number) => window.setTimeout(fn, ms);

const route = useRoute();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

// ---------- 数据加载 ----------
const modules = ref<Module[]>([]);
const loading = ref(true);
const error = ref("");

async function loadModules() {
  loading.value = true;
  error.value = "";
  try {
    modules.value = await moduleApi.list(workspaceId, projectId);
  } catch (e: any) {
    error.value = e.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
}

// ---------- 成员列表（负责人搜索） ----------
const members = ref<Member[]>([]);
const memberSearch = ref("");
const showMemberDropdown = ref<number | null>(null);

const filteredMembers = computed(() => {
  const q = memberSearch.value.toLowerCase();
  if (!q) return members.value.slice(0, 20);
  return members.value.filter(
    (m) => m.display_name.toLowerCase().includes(q) || m.email.toLowerCase().includes(q),
  ).slice(0, 20);
});

function getMemberName(id: number | undefined): string {
  if (!id) return "—";
  const m = members.value.find((x) => x.id === id);
  return m?.display_name ?? `ID:${id}`;
}

async function loadMembers() {
  try {
    members.value = await workspaceApi.listMembers(workspaceId);
  } catch {
    // 静默失败，负责人选择器为空
  }
}

// ---------- 新建模块 ----------
const showCreate = ref(false);
const createForm = reactive({ name: "", description: "", lead_id: undefined as number | undefined });
const creating = ref(false);

function openCreate() {
  showCreate.value = true;
  createForm.name = "";
  createForm.description = "";
  createForm.lead_id = undefined;
  memberSearch.value = "";
  showMemberDropdown.value = null;
  if (members.value.length === 0) loadMembers();
}

function cancelCreate() {
  showCreate.value = false;
}

async function doCreate() {
  if (!createForm.name.trim()) return;
  creating.value = true;
  try {
    const input: CreateModuleInput = {
      name: createForm.name.trim(),
    };
    if (createForm.description.trim()) input.description = createForm.description.trim();
    if (createForm.lead_id) input.lead_id = createForm.lead_id;
    await moduleApi.create(workspaceId, projectId, input);
    toast.success("模块创建成功");
    showCreate.value = false;
    await loadModules();
  } catch (e: any) {
    toast.error(e.message ?? "创建失败");
  } finally {
    creating.value = false;
  }
}

// ---------- 行内编辑 ----------
const editingId = ref<number | null>(null);
const editForm = reactive({ name: "", description: "", lead_id: undefined as number | undefined });
const saving = ref(false);

function startEdit(mod: Module) {
  editingId.value = mod.id;
  editForm.name = mod.name;
  editForm.description = mod.description ?? "";
  editForm.lead_id = mod.lead_id;
  memberSearch.value = "";
  showMemberDropdown.value = null;
  if (members.value.length === 0) loadMembers();
}

function cancelEdit() {
  editingId.value = null;
}

async function saveEdit(modId: number) {
  if (!editForm.name.trim()) return;
  saving.value = true;
  try {
    const input: UpdateModuleInput = { name: editForm.name.trim() };
    if (editForm.description.trim()) {
      input.description = editForm.description.trim();
    } else {
      input.description = "";
    }
    input.lead_id = editForm.lead_id;
    await moduleApi.update(workspaceId, projectId, modId, input);
    toast.success("模块更新成功");
    editingId.value = null;
    await loadModules();
  } catch (e: any) {
    toast.error(e.message ?? "更新失败");
  } finally {
    saving.value = false;
  }
}

// ---------- 状态切换 ----------
async function toggleStatus(mod: Module) {
  try {
    const newStatus = mod.status === "active" ? "archived" : "active";
    await moduleApi.update(workspaceId, projectId, mod.id, { status: newStatus });
    mod.status = newStatus;
    toast.success(newStatus === "active" ? "模块已启用" : "模块已禁用");
  } catch (e: any) {
    toast.error(e.message ?? "操作失败");
  }
}

// ---------- 删除 ----------
const deleteTarget = ref<Module | null>(null);
const deleting = ref(false);

async function confirmDelete() {
  if (!deleteTarget.value) return;
  deleting.value = true;
  try {
    await moduleApi.remove(workspaceId, projectId, deleteTarget.value.id);
    toast.success("模块已删除");
    deleteTarget.value = null;
    await loadModules();
  } catch (e: any) {
    toast.error(e.message ?? "删除失败");
  } finally {
    deleting.value = false;
  }
}

// ---------- 成员选择辅助 ----------
function selectMember(target: "create" | "edit", member: Member) {
  if (target === "create") {
    createForm.lead_id = member.id;
    memberSearch.value = member.display_name;
    showMemberDropdown.value = null;
  } else {
    editForm.lead_id = member.id;
    memberSearch.value = member.display_name;
    showMemberDropdown.value = null;
  }
}

function clearLead(target: "create" | "edit") {
  if (target === "create") {
    createForm.lead_id = undefined;
    memberSearch.value = "";
  } else {
    editForm.lead_id = undefined;
    memberSearch.value = "";
  }
}

onMounted(() => {
  loadModules();
  loadMembers();
});
</script>

<template>
  <div class="module-settings">
    <div class="settings__header">
      <h1>模块管理</h1>
      <p class="meta">管理项目模块，为工作项划分分类维度</p>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="loadModules" />

    <div v-else class="content">
      <!-- 操作栏 -->
      <div class="toolbar">
        <button class="btn btn--primary" @click="openCreate">添加模块</button>
      </div>

      <!-- 内联创建表单 -->
      <div v-if="showCreate" class="panel create-panel">
        <h3 class="panel__title">新建模块</h3>
        <div class="form-grid">
          <label class="form-item">
            <span class="form-item__label">模块名称 <span class="required">*</span></span>
            <input v-model="createForm.name" type="text" maxlength="128" placeholder="输入模块名称" />
          </label>
          <label class="form-item">
            <span class="form-item__label">描述</span>
            <input v-model="createForm.description" type="text" maxlength="500" placeholder="模块用途简介" />
          </label>
          <div class="form-item">
            <span class="form-item__label">负责人</span>
            <div class="member-picker">
              <div
class="picker-input" tabindex="0" @focus="showMemberDropdown = 1"
                @blur="setTimeout(() => showMemberDropdown = null, 200)"
>
                <span v-if="createForm.lead_id" class="selected-member">{{ getMemberName(createForm.lead_id) }}</span>
                <input v-else v-model="memberSearch" type="text" placeholder="搜索成员..." @focus="showMemberDropdown = 1" />
                <button v-if="createForm.lead_id" type="button" class="clear-btn" @click="clearLead('create')">&times;</button>
              </div>
              <div v-if="showMemberDropdown === 1" class="picker-dropdown">
                <div v-if="filteredMembers.length === 0" class="dropdown-empty">无匹配成员</div>
                <button
v-for="m in filteredMembers" :key="m.id" type="button" class="dropdown-item"
                  :class="{ active: createForm.lead_id === m.id }" @mousedown.prevent="selectMember('create', m)"
>
                  <span class="member-avatar">{{ m.display_name.charAt(0) }}</span>
                  <span class="member-name">{{ m.display_name }}</span>
                  <span class="member-email">{{ m.email }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
        <div class="actions">
          <button class="btn btn--primary" :disabled="creating || !createForm.name.trim()" @click="doCreate">
            {{ creating ? "创建中..." : "确认创建" }}
          </button>
          <button class="btn" @click="cancelCreate">取消</button>
        </div>
      </div>

      <!-- 模块列表 -->
      <div class="panel">
        <h3 class="panel__title">模块列表</h3>
        <div v-if="modules.length === 0" class="empty-state">
          暂无模块，点击"添加模块"创建第一个
        </div>
        <div v-else class="table-wrap">
          <table class="module-table">
            <thead>
              <tr>
                <th class="col--name">模块名称</th>
                <th class="col--desc">描述</th>
                <th class="col--lead">负责人</th>
                <th class="col--status">状态</th>
                <th class="col--count">工作项</th>
                <th class="col--sort">排序</th>
                <th class="col--actions">操作</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="mod in modules" :key="mod.id">
                <!-- 正常行 -->
                <tr v-if="editingId !== mod.id">
                  <td class="cell--name">{{ mod.name }}</td>
                  <td class="cell--desc" :title="mod.description">{{ mod.description || "—" }}</td>
                  <td>{{ getMemberName(mod.lead_id) }}</td>
                  <td>
                    <span class="status-tag" :class="mod.status === 'active' ? 'status--active' : 'status--archived'">
                      {{ mod.status === "active" ? "启用" : "禁用" }}
                    </span>
                  </td>
                  <td>{{ mod.issue_count ?? 0 }}</td>
                  <td>{{ mod.sort_order }}</td>
                  <td class="cell--actions">
                    <button class="btn btn--sm" @click="startEdit(mod)">编辑</button>
                    <button class="btn btn--sm" @click="toggleStatus(mod)">
                      {{ mod.status === "active" ? "禁用" : "启用" }}
                    </button>
                    <button class="btn btn--sm btn--danger" @click="deleteTarget = mod">删除</button>
                  </td>
                </tr>
                <!-- 编辑行 -->
                <tr v-else class="edit-row">
                  <td>
                    <input v-model="editForm.name" type="text" maxlength="128" class="inline-input" />
                  </td>
                  <td>
                    <input v-model="editForm.description" type="text" maxlength="500" class="inline-input" placeholder="可选" />
                  </td>
                  <td>
                    <div class="member-picker">
                      <div
class="picker-input" tabindex="0" @focus="showMemberDropdown = 2"
                        @blur="setTimeout(() => showMemberDropdown = null, 200)"
>
                        <span v-if="editForm.lead_id" class="selected-member">{{ getMemberName(editForm.lead_id) }}</span>
                        <input v-else v-model="memberSearch" type="text" placeholder="搜索..." @focus="showMemberDropdown = 2" />
                        <button v-if="editForm.lead_id" type="button" class="clear-btn" @click="clearLead('edit')">&times;</button>
                      </div>
                      <div v-if="showMemberDropdown === 2" class="picker-dropdown">
                        <div v-if="filteredMembers.length === 0" class="dropdown-empty">无匹配成员</div>
                        <button
v-for="m in filteredMembers" :key="m.id" type="button" class="dropdown-item"
                          :class="{ active: editForm.lead_id === m.id }" @mousedown.prevent="selectMember('edit', m)"
>
                          <span class="member-avatar">{{ m.display_name.charAt(0) }}</span>
                          <span class="member-name">{{ m.display_name }}</span>
                          <span class="member-email">{{ m.email }}</span>
                        </button>
                      </div>
                    </div>
                  </td>
                  <td colspan="3"></td>
                  <td class="cell--actions">
                    <button class="btn btn--sm btn--primary" :disabled="saving || !editForm.name.trim()" @click="saveEdit(mod.id)">
                      {{ saving ? "保存中..." : "保存" }}
                    </button>
                    <button class="btn btn--sm" @click="cancelEdit">取消</button>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <AppModal :visible="deleteTarget !== null" title="确认删除" @close="deleteTarget = null">
      <p class="delete-confirm-text">
        确定要删除模块「{{ deleteTarget?.name }}」吗？此操作不可撤销。
      </p>
      <template #footer>
        <button class="btn" @click="deleteTarget = null">取消</button>
        <button class="btn btn--danger" :disabled="deleting" @click="confirmDelete">
          {{ deleting ? "删除中..." : "确认删除" }}
        </button>
      </template>
    </AppModal>
  </div>
</template>

<style scoped>
.module-settings { max-width: 900px; }

.settings__header {
  margin-bottom: 24px;
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

.content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  justify-content: flex-end;
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
  margin: 0 0 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.create-panel {
  border-color: var(--brand-300);
  background: var(--brand-50, #f0f4ff);
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 16px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-item__label {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.required {
  color: var(--danger-500);
}

.form-item input[type="text"] {
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  background: var(--surface-1);
  outline: none;
  font-family: inherit;
}

.form-item input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50);
}

.actions {
  margin-top: 16px;
  display: flex;
  gap: 8px;
}

/* ---------- 表格 ---------- */
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
  padding: 8px 12px;
  color: var(--text-tertiary);
  font-weight: 500;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.module-table td {
  padding: 12px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
}

.module-table tbody tr:hover {
  background: var(--surface-hover);
}

.col--name { width: 16%; }
.col--desc { width: 24%; max-width: 200px; }
.col--lead { width: 12%; }
.col--status { width: 8%; }
.col--count { width: 7%; text-align: center; }
.col--sort { width: 7%; text-align: center; }
.col--actions { width: 20%; }

.module-table th.col--count,
.module-table th.col--sort,
.module-table td:nth-child(5),
.module-table td:nth-child(6) {
  text-align: center;
}

.cell--name {
  font-weight: 500;
}

.cell--desc {
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.cell--actions {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.status-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.status--active {
  background: var(--success-50, #e6f9f0);
  color: var(--success-600, #0c8c4a);
}

.status--archived {
  background: var(--text-tertiary-bg, #f0f0f0);
  color: var(--text-tertiary);
}

/* ---------- 编辑行 ---------- */
.edit-row td {
  padding: 8px 12px;
}

.inline-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  background: var(--surface-1);
  outline: none;
  font-family: inherit;
}

.inline-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

/* ---------- 成员选择器 ---------- */
.member-picker {
  position: relative;
}

.picker-input {
  display: flex;
  align-items: center;
  padding: 7px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  min-height: 34px;
  cursor: text;
}

.picker-input:focus-within {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50);
}

.picker-input input {
  border: none;
  outline: none;
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  background: transparent;
  font-family: inherit;
}

.selected-member {
  font-size: 13px;
  color: var(--text-primary);
  flex: 1;
}

.clear-btn {
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 16px;
  padding: 0 2px;
  line-height: 1;
}

.clear-btn:hover {
  color: var(--text-primary);
}

.picker-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  z-index: 50;
  margin-top: 4px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  box-shadow: var(--shadow-popover);
  max-height: 200px;
  overflow-y: auto;
}

.dropdown-empty {
  padding: 12px;
  font-size: 13px;
  color: var(--text-tertiary);
  text-align: center;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-primary);
  text-align: left;
}

.dropdown-item:hover {
  background: var(--surface-hover);
}

.dropdown-item.active {
  background: var(--brand-50, #f0f4ff);
  color: var(--brand-600);
}

.member-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--brand-100, #dbe4ff);
  color: var(--brand-600);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.member-name {
  flex: 1;
}

.member-email {
  color: var(--text-tertiary);
  font-size: 12px;
}

/* ---------- 空白状态 ---------- */
.empty-state {
  text-align: center;
  padding: 40px 0;
  color: var(--text-tertiary);
  font-size: 14px;
}

/* ---------- 通用按钮 ---------- */
.btn {
  padding: 7px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn:hover {
  background: var(--surface-hover);
}

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
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

.btn--danger {
  background: var(--danger-500);
  border-color: var(--danger-500);
  color: var(--text-on-brand);
}

.btn--danger:hover:not(:disabled) {
  background: var(--danger-600);
}

.btn--danger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ---------- 删除确认 ---------- */
.delete-confirm-text {
  font-size: 14px;
  color: var(--text-primary);
  margin: 0;
}

@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }

  .col--desc,
  .cell--desc {
    display: none;
  }
}
</style>
