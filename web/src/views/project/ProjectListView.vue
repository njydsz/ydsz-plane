<script setup lang="ts">
/**
 * 项目列表页 — 展示当前工作空间下的全部项目，支持新建项目。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { workspaceApi, type Project, type Workspace } from "@/api/services/workspace";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";

const route = useRoute();
const router = useRouter();
const wsSlug = computed(() => String(route.params.workspaceSlug));

const ws = ref<Workspace | null>(null);
const projects = ref<Project[]>([]);
const loading = ref(true);
const error = ref("");

// 创建项目
const showCreate = ref(false);
const createName = ref("");
const createSending = ref(false);
const createError = ref("");

async function load() {
  loading.value = true;
  error.value = "";
  try {
    ws.value = await workspaceApi.getBySlug(wsSlug.value);
    projects.value = await workspaceApi.listProjects(ws.value.id);
  } catch (e: any) {
    error.value = e.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
}

async function createProject() {
  if (!createName.value.trim() || !ws.value) return;
  createError.value = "";
  createSending.value = true;
  try {
    const p = await workspaceApi.createProject(ws.value.id, {
      name: createName.value.trim(),
    });
    projects.value.push(p);
    showCreate.value = false;
    createName.value = "";
  } catch (e: any) {
    createError.value = e.message ?? "创建失败";
  } finally {
    createSending.value = false;
  }
}

function openSettings() {
  router.push(`/${wsSlug.value}/settings`);
}

onMounted(load);
</script>

<template>
  <div class="projects">
    <header class="projects__header">
      <div>
        <h1 v-if="ws">{{ ws.name }} · 项目</h1>
        <h1 v-else>项目</h1>
        <p class="hint">管理工作空间下的项目</p>
      </div>
      <div class="actions">
        <button class="btn" @click="openSettings">工作空间设置</button>
        <button class="btn btn--primary" @click="showCreate = true">创建项目</button>
      </div>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="projects.length === 0"
      title="暂无项目"
      description="该项目空间下暂无项目"
    >
      <button class="btn btn--primary" @click="showCreate = true">创建项目</button>
    </AppEmptyState>
    <div v-else class="project-grid">
      <div v-for="p in projects" :key="p.id" class="project-card" :style="{ borderTopColor: p.color || 'var(--brand-500)' }">
        <div v-if="p.icon" class="project-card__icon">{{ p.icon }}</div>
        <div class="project-card__body">
          <div class="project-card__name">{{ p.name }}</div>
          <div class="project-card__meta">
            <span class="identifier">{{ p.identifier }}</span>
            <span class="slug">/{{ p.slug }}</span>
          </div>
          <p v-if="p.description" class="project-card__desc">{{ p.description }}</p>
        </div>
      </div>
    </div>

    <!-- 创建项目 Modal -->
    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal">
        <header class="modal__header">
          <h2>创建项目</h2>
          <button class="modal__close" @click="showCreate = false">×</button>
        </header>
        <div class="modal__body">
          <label class="field">
            <span class="field__label">项目名称</span>
            <input v-model="createName" class="field__input" placeholder="例如：用户中心" maxlength="80" autofocus />
          </label>
          <div v-if="createError" class="form-error">{{ createError }}</div>
        </div>
        <footer class="modal__footer">
          <button class="btn" @click="showCreate = false">取消</button>
          <button class="btn btn--primary" :disabled="createSending || !createName.trim()" @click="createProject">
            {{ createSending ? "创建中..." : "创建" }}
          </button>
        </footer>
      </div>
    </div>
  </div>
</template>

<style scoped>
.projects__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.projects__header h1 {
  font-size: 20px;
  margin: 0 0 4px;
}

.hint { color: var(--text-tertiary); font-size: 13px; margin: 0; }

.actions { display: flex; gap: 10px; }

.loading, .error, .empty {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}
.error { color: var(--danger-500); }

.project-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 14px;
}

.project-card {
  padding: 16px;
  border: 1px solid var(--border-default);
  border-top: 3px solid var(--brand-500);
  border-radius: var(--radius-md);
  background: var(--surface-1);
}

.project-card__icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.project-card__name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 15px;
}

.project-card__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.identifier {
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--surface-3);
  color: var(--text-secondary);
  font-weight: 600;
  font-family: var(--font-mono);
}

.project-card__desc {
  color: var(--text-secondary);
  font-size: 13px;
  margin: 8px 0 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  width: 440px;
  max-width: 95vw;
  background: var(--surface-1);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-popover);
}

.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-subtle);
}

.modal__header h2 { font-size: 16px; margin: 0; }
.modal__close { background: none; border: none; font-size: 24px; color: var(--text-tertiary); cursor: pointer; }

.modal__body { padding: 24px; }
.modal__footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 24px; border-top: 1px solid var(--border-subtle); }

.field { display: block; }
.field__label { display: block; font-size: 13px; font-weight: 500; color: var(--text-secondary); margin-bottom: 6px; }

.field__input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 14px;
  font-family: inherit;
}

.field__input:focus { outline: none; border-color: var(--brand-500); box-shadow: 0 0 0 3px var(--brand-50); }

.form-error { color: var(--danger-500); font-size: 13px; padding: 8px 0; }
</style>
