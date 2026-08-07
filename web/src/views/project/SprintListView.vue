<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { sprintApi, type Sprint, type SprintStatus } from "@/api/services/sprint";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));

const sprints = ref<Sprint[]>([]);
const loading = ref(true);
const error = ref("");
const showCreate = ref(false);

// create form
const form = ref({
  name: "",
  description: "",
  goal: "",
  start_date: "",
  end_date: "",
  capacity: undefined as number | undefined,
});
const creating = ref(false);
const createError = ref("");

let wsIdVal = 0;

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.getBySlug(workspaceSlug.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const res = await sprintApi.listSprints(wsId, projectId.value);
    sprints.value = res.results;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载迭代失败";
  } finally {
    loading.value = false;
  }
}

async function createSprint() {
  if (!form.value.name.trim()) {
    createError.value = "迭代名称不能为空";
    return;
  }
  creating.value = true;
  createError.value = "";
  try {
    const wsId = await resolveWsId();
    const payload: Parameters<typeof sprintApi.createSprint>[2] = {
      name: form.value.name,
      description: form.value.description || undefined,
      goal: form.value.goal || undefined,
      start_date: form.value.start_date || undefined,
      end_date: form.value.end_date || undefined,
      capacity: form.value.capacity,
    };
    const sp = await sprintApi.createSprint(wsId, projectId.value, payload);
    showCreate.value = false;
    resetForm();
    sprints.value = [sp, ...sprints.value];
  } catch (e: unknown) {
    createError.value = e instanceof Error ? e.message : "创建失败";
  } finally {
    creating.value = false;
  }
}

function resetForm() {
  form.value = { name: "", description: "", goal: "", start_date: "", end_date: "", capacity: undefined };
}

function openSprint(sp: Sprint) {
  router.push(`/${workspaceSlug.value}/projects/${projectId.value}/sprints/${sp.id}`);
}

function goPlanning() {
  router.push(`/${workspaceSlug.value}/projects/${projectId.value}/sprints/planning`);
}

function statusColor(s: SprintStatus): string {
  const map: Record<SprintStatus, string> = {
    planned: "var(--text-tertiary)",
    active: "var(--success-500)",
    completed: "var(--brand-500)",
  };
  return map[s];
}

function statusLabel(s: SprintStatus): string {
  const map: Record<SprintStatus, string> = {
    planned: "未开始",
    active: "进行中",
    completed: "已完成",
  };
  return map[s];
}

onMounted(load);
</script>

<template>
  <div class="sprint-list">
    <header class="header">
      <div>
        <h1>迭代</h1>
        <p class="hint">管理项目的 Sprint 生命周期</p>
      </div>
      <div class="actions">
        <button class="btn btn-secondary" @click="goPlanning">排期规划</button>
        <button class="btn btn-primary" @click="showCreate = true">新建迭代</button>
      </div>
    </header>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="grid">
      <div
        v-for="sp in sprints"
        :key="sp.id"
        class="sprint-card"
        @click="openSprint(sp)"
      >
        <div class="sprint-card__header">
          <div class="title-row">
            <span class="status-dot" :style="{ background: statusColor(sp.status) }"></span>
            <h3>{{ sp.name }}</h3>
          </div>
          <span class="status-badge" :style="{ color: statusColor(sp.status) }">
            {{ statusLabel(sp.status) }}
          </span>
        </div>
        <p v-if="sp.goal" class="goal">{{ sp.goal }}</p>
        <div class="date-range" v-if="sp.start_date || sp.end_date">
          <span>{{ sp.start_date ?? "?" }}</span>
          <span class="sep">→</span>
          <span>{{ sp.end_date ?? "?" }}</span>
        </div>
        <div v-if="sp.progress" class="progress-row">
          <div class="progress-info">
            <span>{{ sp.progress.done_points }}/{{ sp.progress.total_points }} 故事点</span>
            <span>{{ sp.progress.done_issues }}/{{ sp.progress.total_issues }} 工作项</span>
          </div>
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{
                width: (sp.progress.total_points > 0
                  ? (sp.progress.done_points / sp.progress.total_points) * 100
                  : 0) + '%',
              }"
            ></div>
          </div>
        </div>
      </div>

      <div v-if="sprints.length === 0" class="empty">
        <p>还没有迭代</p>
        <button class="btn btn-primary" @click="showCreate = true">创建第一个迭代</button>
      </div>
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

.loading, .error, .empty {
  text-align: center; padding: 48px 0; color: var(--text-tertiary);
}
.error { color: var(--danger-500); }
.empty .btn { margin-top: 12px; }

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.sprint-card {
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
  margin-bottom: 8px;
}
.title-row { display: flex; align-items: center; gap: 8px; }
.title-row h3 { margin: 0; font-size: 14px; font-weight: 600; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.status-badge { font-size: 11px; font-weight: 500; }

.goal { font-size: 12px; color: var(--text-secondary); margin: 0 0 8px;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.date-range { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); display: flex; gap: 4px; }
.date-range .sep { opacity: 0.5; }

.progress-row { margin-top: 10px; }
.progress-info { display: flex; justify-content: space-between; font-size: 11px;
  color: var(--text-tertiary); font-family: var(--font-mono); margin-bottom: 4px; }
.progress-bar { height: 4px; background: var(--surface-3); border-radius: 2px; overflow: hidden; }
.progress-fill { height: 100%; background: var(--success-500); transition: width 0.3s; }

.btn {
  font-size: 13px; font-weight: 500; padding: 6px 14px; border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle); cursor: pointer; transition: background 0.15s;
}
.btn-primary { background: var(--brand-500); color: #fff; border-color: var(--brand-500); }
.btn-primary:hover:not(:disabled) { background: var(--brand-600); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: var(--surface-2); color: var(--text-primary); }
.btn-secondary:hover { background: var(--surface-3); }

.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
  background: var(--surface-1); padding: 20px; border-radius: var(--radius-md);
  width: 480px; max-width: 90vw; box-shadow: var(--shadow-elevated);
}
.modal header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.modal h2 { margin: 0; font-size: 16px; }
.close { background: none; border: none; font-size: 22px; cursor: pointer; line-height: 1; color: var(--text-tertiary); }

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
