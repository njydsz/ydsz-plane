<script setup lang="ts">
/**
 * 迭代列表页 — 展示全部迭代，支持创建 / 编辑 / 归档 / 状态筛选 / 分页。
 */

import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import dayjs from "dayjs";

import { sprintApi, type Sprint, type SprintStatus } from "@/api/services/sprint";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import SprintStatusBadge from "@/components/sprint/SprintStatusBadge.vue";
import SprintProgressBar from "@/components/sprint/SprintProgressBar.vue";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";

/* ------------------------------------------------------------------ */
/* 路由上下文                                                           */
/* ------------------------------------------------------------------ */

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const { wsId, ready } = useWorkspaceContext();

/* ------------------------------------------------------------------ */
/* 列表状态                                                             */
/* ------------------------------------------------------------------ */

const sprints = ref<Sprint[]>([]);
const total = ref(0);
const loading = ref(true);
const error = ref("");

/** 状态筛选：all / planned / active / completed */
const filterStatus = ref<"all" | SprintStatus>("all");
/** 排序：created_desc（默认）/ start_date */
const sortBy = ref<"created_desc" | "start_date">("created_desc");

const PAGE_SIZE = 9;
const offset = ref(0);
const hasMore = computed(() => offset.value + sprints.value.length < total.value);

const filteredSprints = computed(() => {
  let list = [...sprints.value];
  if (sortBy.value === "start_date") {
    list.sort((a, b) => (a.start_date ?? "").localeCompare(b.start_date ?? ""));
  }
  return list;
});

async function load(reset = true) {
  loading.value = true;
  error.value = "";
  try {
    const params = {
      status: filterStatus.value === "all" ? undefined : filterStatus.value,
      limit: PAGE_SIZE,
      offset: reset ? 0 : offset.value,
    };
    const res = await sprintApi.listSprints(wsId.value, projectId.value, params);
    sprints.value = reset ? res.results : [...sprints.value, ...res.results];
    total.value = res.total;
    offset.value = reset ? res.results.length : offset.value + res.results.length;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载迭代失败";
  } finally {
    loading.value = false;
  }
}

/** 加载更多（分页） */
function loadMore() {
  if (!hasMore.value || loading.value) return;
  void load(false);
}

// 筛选变化时重置并重新加载
watch([filterStatus, ready], ([, r]) => {
  if (r) void load(true);
});

/* ------------------------------------------------------------------ */
/* 创建迭代                                                             */
/* ------------------------------------------------------------------ */

const showCreate = ref(false);
const creating = ref(false);
const createError = ref("");

const emptyForm = () => ({
  name: "",
  description: "",
  goal: "",
  start_date: "",
  end_date: "",
  capacity: undefined as number | undefined,
});
const form = ref(emptyForm());

async function createSprint() {
  if (!form.value.name.trim()) {
    createError.value = "迭代名称不能为空";
    return;
  }
  creating.value = true;
  createError.value = "";
  try {
    const payload = {
      name: form.value.name,
      description: form.value.description || undefined,
      goal: form.value.goal || undefined,
      start_date: form.value.start_date || undefined,
      end_date: form.value.end_date || undefined,
      capacity: form.value.capacity,
    };
    await sprintApi.createSprint(wsId.value, projectId.value, payload);
    showCreate.value = false;
    resetForm();
    await load(true);
  } catch (e: unknown) {
    createError.value = e instanceof Error ? e.message : "创建失败";
  } finally {
    creating.value = false;
  }
}

function resetForm() {
  form.value = emptyForm();
}

/* ------------------------------------------------------------------ */
/* 编辑迭代                                                             */
/* ------------------------------------------------------------------ */

const editTarget = ref<Sprint | null>(null);
const savingEdit = ref(false);
const editError = ref("");

const editForm = ref({
  name: "",
  description: "",
  goal: "",
  start_date: "",
  end_date: "",
  capacity: undefined as number | undefined,
});

function openEdit(sp: Sprint) {
  editTarget.value = sp;
  editError.value = "";
  editForm.value = {
    name: sp.name,
    description: sp.description ?? "",
    goal: sp.goal ?? "",
    start_date: sp.start_date ?? "",
    end_date: sp.end_date ?? "",
    capacity: sp.capacity,
  };
}

async function saveEdit() {
  if (!editTarget.value) return;
  if (!editForm.value.name.trim()) {
    editError.value = "迭代名称不能为空";
    return;
  }
  savingEdit.value = true;
  editError.value = "";
  try {
    const sp = editTarget.value;
    await sprintApi.updateSprint(wsId.value, projectId.value, sp.id, {
      name: editForm.value.name,
      description: editForm.value.description || undefined,
      goal: editForm.value.goal || undefined,
      start_date: editForm.value.start_date || undefined,
      end_date: editForm.value.end_date || undefined,
      capacity: editForm.value.capacity,
      version: 0,
    });
    editTarget.value = null;
    await load(true);
  } catch (e: unknown) {
    editError.value = e instanceof Error ? e.message : "保存失败";
  } finally {
    savingEdit.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 删除迭代（仅 planned / completed 可归档，active 后端拒绝）            */
/* ------------------------------------------------------------------ */

const deleteTarget = ref<Sprint | null>(null);
const deleting = ref(false);
const deleteError = ref("");

function askDelete(sp: Sprint) {
  deleteTarget.value = sp;
  deleteError.value = "";
}

async function confirmDelete() {
  if (!deleteTarget.value) return;
  deleting.value = true;
  deleteError.value = "";
  try {
    await sprintApi.deleteSprint(wsId.value, projectId.value, deleteTarget.value.id);
    deleteTarget.value = null;
    await load(true);
  } catch (e: unknown) {
    deleteError.value = e instanceof Error ? e.message : "删除失败";
  } finally {
    deleting.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 工具                                                                 */
/* ------------------------------------------------------------------ */

function openSprint(sp: Sprint) {
  router.push(`/${workspaceSlug.value}/projects/${projectId.value}/sprints/${sp.id}`);
}

function goPlanning() {
  router.push(`/${workspaceSlug.value}/projects/${projectId.value}/sprints/planning`);
}

function fmtDate(d?: string): string {
  return d ? dayjs(d).format("YYYY-MM-DD") : "—";
}

function stopProp(e: MouseEvent) {
  e.stopPropagation();
}

onMounted(() => {
  if (ready.value) void load(true);
});
watch(ready, (r) => {
  if (r) void load(true);
});
</script>

<template>
  <div class="sprint-list">
    <header class="header">
      <div>
        <h1>迭代</h1>
        <p class="hint">管理项目的 Sprint 生命周期 · 共 {{ total }} 个迭代</p>
      </div>
      <div class="actions">
        <button class="btn btn-secondary" @click="goPlanning">排期规划</button>
        <button class="btn btn-primary" @click="showCreate = true">新建迭代</button>
      </div>
    </header>

    <!-- 筛选与排序工具条 -->
    <div class="toolbar">
      <div class="filter-tabs" role="tablist" aria-label="按状态筛选迭代">
        <button
          v-for="f in ([{ key: 'all', label: '全部' }, { key: 'planned', label: '未开始' }, { key: 'active', label: '进行中' }, { key: 'completed', label: '已完成' }] as const)"
          :key="f.key"
          class="filter-tab"
          :class="{ active: filterStatus === f.key }"
          role="tab"
          :aria-selected="filterStatus === f.key"
          @click="filterStatus = f.key"
        >
          {{ f.label }}
        </button>
      </div>
      <label class="sort-select">
        排序
        <select v-model="sortBy">
          <option value="created_desc">最新创建</option>
          <option value="start_date">开始日期</option>
        </select>
      </label>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load(true)" />

    <AppEmptyState
      v-else-if="!loading && !error && sprints.length === 0"
      title="暂无迭代"
      description="当前筛选条件下没有迭代"
    >
      <button class="btn btn-primary" @click="showCreate = true">创建第一个迭代</button>
    </AppEmptyState>

    <div v-else class="grid">
      <div
        v-for="sp in filteredSprints"
        :key="sp.id"
        class="sprint-card"
        @click="openSprint(sp)"
      >
        <div class="sprint-card__header">
          <div class="title-row">
            <SprintStatusBadge :status="sp.status" dot />
            <h3>{{ sp.name }}</h3>
          </div>
          <SprintStatusBadge :status="sp.status" />
        </div>
        <p v-if="sp.goal" class="goal">{{ sp.goal }}</p>
        <div class="date-range" v-if="sp.start_date || sp.end_date">
          <span>{{ fmtDate(sp.start_date) }}</span>
          <span class="sep">→</span>
          <span>{{ fmtDate(sp.end_date) }}</span>
          <span v-if="sp.capacity" class="capacity">容量 {{ sp.capacity }}pt</span>
        </div>
        <div v-if="sp.progress" class="progress-row">
          <SprintProgressBar
            :done-points="sp.progress.done_points"
            :total-points="sp.progress.total_points"
          />
        </div>

        <!-- 卡片操作（不触发跳转） -->
        <div class="card-actions" @click="stopProp">
          <button
            class="icon-action"
            title="编辑迭代"
            :aria-label="`编辑 ${sp.name}`"
            @click="openEdit(sp)"
          >
            ✏️
          </button>
          <button
            class="icon-action danger"
            title="归档迭代"
            :aria-label="`归档 ${sp.name}`"
            :disabled="sp.status === 'active'"
            @click="askDelete(sp)"
          >
            🗑️
          </button>
        </div>
      </div>

    </div>

    <!-- 加载更多 -->
    <div v-if="hasMore" class="load-more">
      <button class="btn btn-secondary" :disabled="loading" @click="loadMore">
        {{ loading ? "加载中..." : "加载更多" }}
      </button>
    </div>

    <!-- Create modal -->
    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal">
        <header>
          <h2>新建迭代</h2>
          <button class="close" @click="showCreate = false">×</button>
        </header>
        <form @submit.prevent="createSprint">
          <label>名称 <span class="req">*</span>
            <input v-model="form.name" placeholder="比如 Sprint 5" maxlength="80" />
          </label>
          <label>描述
            <textarea v-model="form.description" rows="2" maxlength="500" />
          </label>
          <label>目标
            <input v-model="form.goal" placeholder="本轮迭代要达成的目标" maxlength="500" />
          </label>
          <div class="row">
            <label>开始日期 <input v-model="form.start_date" type="date" /></label>
            <label>结束日期 <input v-model="form.end_date" type="date" /></label>
          </div>
          <label>容量（故事点）
            <input v-model.number="form.capacity" type="number" min="0" step="1" placeholder="推荐基于速率统计设定" />
          </label>
          <div v-if="createError" class="error">{{ createError }}</div>
          <footer>
            <button type="button" class="btn btn-secondary" @click="showCreate = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="creating">
              {{ creating ? "创建中..." : "创建" }}
            </button>
          </footer>
        </form>
      </div>
    </div>

    <!-- Edit modal -->
    <div v-if="editTarget" class="modal-overlay" @click.self="editTarget = null">
      <div class="modal">
        <header>
          <h2>编辑迭代：{{ editTarget.name }}</h2>
          <button class="close" @click="editTarget = null">×</button>
        </header>
        <form @submit.prevent="saveEdit">
          <label>名称 <span class="req">*</span>
            <input v-model="editForm.name" maxlength="80" />
          </label>
          <label>描述
            <textarea v-model="editForm.description" rows="2" maxlength="500" />
          </label>
          <label>目标
            <input v-model="editForm.goal" maxlength="500" />
          </label>
          <div class="row">
            <label>开始日期 <input v-model="editForm.start_date" type="date" /></label>
            <label>结束日期 <input v-model="editForm.end_date" type="date" /></label>
          </div>
          <label>容量（故事点）
            <input v-model.number="editForm.capacity" type="number" min="0" step="1" />
          </label>
          <p class="hint-inline">仅「未开始」状态的迭代可编辑核心字段</p>
          <div v-if="editError" class="error">{{ editError }}</div>
          <footer>
            <button type="button" class="btn btn-secondary" @click="editTarget = null">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="savingEdit">
              {{ savingEdit ? "保存中..." : "保存" }}
            </button>
          </footer>
        </form>
      </div>
    </div>

    <!-- Delete confirm modal -->
    <div v-if="deleteTarget" class="modal-overlay" @click.self="deleteTarget = null">
      <div class="modal modal-sm">
        <header>
          <h2>归档迭代</h2>
          <button class="close" @click="deleteTarget = null">×</button>
        </header>
        <p class="confirm-text">
          确定要归档迭代 <b>{{ deleteTarget.name }}</b> 吗？
          其中的工作项将退回 Backlog。此操作不可撤销。
        </p>
        <div v-if="deleteError" class="error">{{ deleteError }}</div>
        <footer>
          <button type="button" class="btn btn-secondary" @click="deleteTarget = null">取消</button>
          <button type="button" class="btn btn-danger" :disabled="deleting" @click="confirmDelete">
            {{ deleting ? "归档中..." : "确认归档" }}
          </button>
        </footer>
      </div>
    </div>
  </div>
</template>

<style scoped>
.header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
  flex-wrap: wrap;
}
.header h1 { margin: 0; font-size: 20px; }
.hint { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }
.actions { display: flex; gap: 8px; }

/* ---- 工具条 ---- */
.toolbar {
  display: flex; align-items: center; justify-content: space-between;
  gap: 12px; margin-bottom: 16px; flex-wrap: wrap;
}
.filter-tabs {
  display: flex; gap: 4px; padding: 3px;
  background: var(--surface-2); border-radius: var(--radius-md);
}
.filter-tab {
  border: none; background: none; padding: 5px 14px; font-size: 12px;
  color: var(--text-secondary); cursor: pointer; border-radius: var(--radius-sm);
  transition: all 0.15s;
}
.filter-tab.active { background: var(--surface-1); color: var(--text-primary); font-weight: 500; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
.sort-select { font-size: 12px; color: var(--text-secondary); display: flex; align-items: center; gap: 6px; }
.sort-select select {
  font-size: 12px; padding: 4px 8px; border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm); background: var(--surface-1); color: var(--text-primary);
}

.loading, .error, .empty {
  text-align: center; padding: 48px 0; color: var(--text-tertiary);
}
.error { color: var(--danger-500); }
.empty .btn { margin-top: 12px; }

.center-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
  gap: 10px;
  color: var(--text-tertiary);
}
.center-message.error { color: var(--danger-500); }
.center-message p { margin: 0; font-size: 14px; }
.center-message p.detail { font-size: 12px; opacity: 0.8; max-width: 400px; word-break: break-word; }

.skeleton-row {
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
}
.skeleton-line {
  height: 12px;
  background: var(--surface-2);
  border-radius: 4px;
  animation: pulse 1.5s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.sprint-card {
  position: relative;
  padding: 14px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: box-shadow 0.15s, transform 0.1s;
}
.sprint-card:hover {
  box-shadow: var(--shadow-card);
  transform: translateY(-1px);
}

.sprint-card__header {
  display: flex; align-items: flex-start; justify-content: space-between;
  margin-bottom: 8px; padding-right: 54px;
}
.title-row { display: flex; align-items: center; gap: 8px; }
.title-row h3 { margin: 0; font-size: 14px; font-weight: 600; }

/* 卡片悬浮操作 */
.card-actions {
  position: absolute; top: 10px; right: 10px;
  display: flex; gap: 4px; opacity: 0;
  transition: opacity 0.15s;
}
.sprint-card:hover .card-actions { opacity: 1; }
.icon-action {
  border: none; background: var(--surface-2); font-size: 12px;
  padding: 4px 6px; border-radius: 4px; cursor: pointer;
  transition: background 0.15s;
}
.icon-action:hover { background: var(--surface-3); }
.icon-action.danger:hover { background: var(--danger-50); }
.icon-action:disabled { opacity: 0.3; cursor: not-allowed; }

.goal { font-size: 12px; color: var(--text-secondary); margin: 0 0 8px;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.date-range { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); display: flex; gap: 4px; flex-wrap: wrap; }
.date-range .sep { opacity: 0.5; }
.date-range .capacity { margin-left: 4px; padding: 1px 5px; background: var(--surface-2); border-radius: 3px; color: var(--text-secondary); }

.progress-row { margin-top: 10px; }

.load-more { display: flex; justify-content: center; margin-top: 20px; }

.btn {
  font-size: 13px; font-weight: 500; padding: 6px 14px; border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle); cursor: pointer; transition: background 0.15s;
}
.btn-primary { background: var(--brand-500); color: #fff; border-color: var(--brand-500); }
.btn-primary:hover:not(:disabled) { background: var(--brand-600); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: var(--surface-2); color: var(--text-primary); }
.btn-secondary:hover { background: var(--surface-3); }
.btn-danger { background: var(--danger-500); color: #fff; border-color: var(--danger-500); }
.btn-danger:hover:not(:disabled) { background: var(--danger-600); }
.btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }

.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
  background: var(--surface-1); padding: 20px; border-radius: var(--radius-md);
  width: 480px; max-width: 90vw; box-shadow: var(--shadow-elevated);
}
.modal-sm { width: 400px; }
.modal header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.modal h2 { margin: 0; font-size: 16px; }
.close { background: none; border: none; font-size: 22px; cursor: pointer; line-height: 1; color: var(--text-tertiary); }
.confirm-text { font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.hint-inline { font-size: 11px; color: var(--text-tertiary); margin: 0; }

form { display: flex; flex-direction: column; gap: 10px; }
label { font-size: 12px; color: var(--text-secondary); display: flex; flex-direction: column; gap: 4px; }
.req { color: var(--danger-500); }
input, textarea {
  font-size: 13px; padding: 6px 10px; border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm); background: var(--surface-2); color: var(--text-primary);
  font-family: inherit; resize: vertical;
}
input:focus, textarea:focus { outline: none; border-color: var(--brand-500); }

.row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }

form footer { display: flex; justify-content: flex-end; gap: 8px; margin-top: 8px; }
</style>
