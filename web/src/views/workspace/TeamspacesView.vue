<script setup lang="ts">
/**
 * TeamspacesView — 团队空间列表。
 *
 * 团队空间 = 一组人 + 一组关联项目（对标 Plane Teamspaces）。
 * 卡片展示：名称、颜色标识、成员数、项目数、完成率、逾期数。
 * 支持创建 / 编辑 / 归档 / 恢复 / 删除。
 */
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { teamspaceApi, type Teamspace } from "@/api/services/teamspace";
import { useWorkspaceStore } from "@/stores/workspace";
import { useTranslation } from "@/composables/useTranslation";
import { AppBadge, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer, DynamicIcon } from "@/components";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const { t } = useTranslation();
const router = useRouter();
const wsStore = useWorkspaceStore();
const wsId = computed(() => wsStore.current?.id ?? 0);

/* ---- 数据 ---- */

const teamspaces = ref<Teamspace[]>([]);
const progressMap = ref<Record<number, { completion_rate: number; overdue_work_items: number; member_count: number; project_count: number }>>({});
const loading = ref(true);
const error = ref("");

/* ---- 创建/编辑 ---- */

const showForm = ref(false);
const editingId = ref<number | null>(null);
const form = ref({ name: "", description: "", color: "var(--brand-500)", lead_id: 0 });
const saving = ref(false);

const colorPresets = ["var(--brand-500)", "var(--success-500)", "var(--warning-500)", "var(--danger-500)", "var(--brand-500)", "var(--brand-500)", "var(--info-500)", "var(--warning-500)"];

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const res = await teamspaceApi.list(wsId.value);
    teamspaces.value = (res as unknown as Teamspace[]) ?? [];
    // 并行加载进度
    await Promise.all(
      teamspaces.value.map(async (ts) => {
        try {
          const p = await teamspaceApi.progress(wsId.value, ts.id);
          progressMap.value = {
            ...progressMap.value,
            [ts.id]: {
              completion_rate: p.completion_rate,
              overdue_work_items: p.overdue_work_items,
              member_count: p.member_count,
              project_count: p.project_count,
            },
          };
        } catch {
          /* 忽略单个团队的进度失败 */
        }
      }),
    );
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载团队空间失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function openCreate() {
  editingId.value = null;
  form.value = { name: "", description: "", color: colorPresets[0], lead_id: 0 };
  showForm.value = true;
}

function openEdit(ts: Teamspace) {
  editingId.value = ts.id;
  form.value = {
    name: ts.name,
    description: ts.description,
    color: ts.color || colorPresets[0],
    lead_id: ts.lead_id,
  };
  showForm.value = true;
}

async function save() {
  if (!form.value.name.trim()) {
    toast.error("请输入团队名称");
    return;
  }
  saving.value = true;
  try {
    if (editingId.value) {
      await teamspaceApi.update(wsId.value, editingId.value, form.value);
      toast.success("团队空间已更新");
    } else {
      await teamspaceApi.create(wsId.value, form.value);
      toast.success("团队空间已创建");
    }
    showForm.value = false;
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function archive(ts: Teamspace) {
  try {
    await teamspaceApi.archive(wsId.value, ts.id);
    toast.success("已归档");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "归档失败");
  }
}

async function restore(ts: Teamspace) {
  try {
    await teamspaceApi.restore(wsId.value, ts.id);
    toast.success("已恢复");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "恢复失败");
  }
}

async function remove(ts: Teamspace) {
  if (!await confirm({ title: "删除团队空间", message: `确认删除团队空间「${ts.name}」？此操作不可撤销。`, confirmText: "删除", variant: "danger" })) return;
  try {
    await teamspaceApi.remove(wsId.value, ts.id);
    toast.success("已删除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

function openDetail(ts: Teamspace) {
  router.push(`/${wsId.value}/teamspaces/${ts.id}`);
}

const activeList = computed(() => teamspaces.value.filter((t) => !t.archived_at));
const archivedList = computed(() => teamspaces.value.filter((t) => t.archived_at));
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="teamspaces">
    <!-- 头部 -->
    <div class="teamspaces__header">
      <div>
        <h1 class="teamspaces__title">{{ t("teamspace.title", "团队空间") }}</h1>
        <p class="teamspaces__subtitle">
          {{ t("teamspace.subtitle", "将一组人与一组关联项目组织在一起，聚合查看团队进度") }}
        </p>
      </div>
      <button class="teamspaces__primary" @click="openCreate">
        <DynamicIcon name="Plus" :size="14" /> {{ t("teamspace.create", "新建团队空间") }}
      </button>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!teamspaces.length"
      title="暂无团队空间"
      description="创建团队空间，把相关人员和项目组织到一起"
    >
      <button class="teamspaces__primary" @click="openCreate">
        <DynamicIcon name="Plus" :size="14" /> {{ t("teamspace.create", "新建团队空间") }}
      </button>
    </AppEmptyState>

    <template v-else>
      <!-- 活跃列表 -->
      <div class="teamspaces__grid">
        <div
          v-for="ts in activeList"
          :key="ts.id"
          class="teamspaces__card"
          @click="openDetail(ts)"
        >
          <div class="teamspaces__card-top">
            <span class="teamspaces__dot" :style="{ background: ts.color || 'var(--brand-500)' }" />
            <span class="teamspaces__name">{{ ts.name }}</span>
          </div>
          <p v-if="ts.description" class="teamspaces__desc">{{ ts.description }}</p>

          <div class="teamspaces__stats">
            <div class="teamspaces__stat">
              <span class="teamspaces__stat-value">
                {{ progressMap[ts.id]?.member_count ?? 0 }}
              </span>
              <span class="teamspaces__stat-label">成员</span>
            </div>
            <div class="teamspaces__stat">
              <span class="teamspaces__stat-value">
                {{ progressMap[ts.id]?.project_count ?? 0 }}
              </span>
              <span class="teamspaces__stat-label">项目</span>
            </div>
            <div class="teamspaces__stat">
              <span class="teamspaces__stat-value">
                {{ Math.round(progressMap[ts.id]?.completion_rate ?? 0) }}%
              </span>
              <span class="teamspaces__stat-label">完成率</span>
            </div>
            <div class="teamspaces__stat">
              <span
                class="teamspaces__stat-value"
                :class="{ 'teamspaces__stat-value--overdue': (progressMap[ts.id]?.overdue_work_items ?? 0) > 0 }"
              >
                {{ progressMap[ts.id]?.overdue_work_items ?? 0 }}
              </span>
              <span class="teamspaces__stat-label">逾期</span>
            </div>
          </div>

          <div class="teamspaces__card-actions" @click.stop>
            <button class="teamspaces__link" @click="openEdit(ts)">编辑</button>
            <button class="teamspaces__link" @click="archive(ts)">归档</button>
            <button class="teamspaces__link teamspaces__link--danger" @click="remove(ts)">
              删除
            </button>
          </div>
        </div>
      </div>

      <!-- 已归档 -->
      <div v-if="archivedList.length" class="teamspaces__archived">
        <h2 class="teamspaces__archived-title">已归档</h2>
        <div
          v-for="ts in archivedList"
          :key="ts.id"
          class="teamspaces__archived-item"
        >
          <span class="teamspaces__dot" :style="{ background: ts.color || 'var(--text-tertiary)' }" />
          <span class="teamspaces__archived-name">{{ ts.name }}</span>
          <AppBadge variant="default">已归档</AppBadge>
          <button class="teamspaces__link" @click="restore(ts)">恢复</button>
          <button class="teamspaces__link teamspaces__link--danger" @click="remove(ts)">
            删除
          </button>
        </div>
      </div>
    </template>

    <!-- 创建/编辑弹窗 -->
    <AppModal
      :visible="showForm"
      :title="editingId ? '编辑团队空间' : '新建团队空间'"
      @close="showForm = false"
    >
      <div class="teamspaces__form">
        <label class="teamspaces__label">名称</label>
        <input
          v-model="form.name"
          class="teamspaces__input"
          placeholder="例如：平台研发组"
        />

        <label class="teamspaces__label">描述</label>
        <textarea
          v-model="form.description"
          class="teamspaces__textarea"
          rows="2"
          placeholder="这个团队负责什么"
        />

        <label class="teamspaces__label">标识色</label>
        <div class="teamspaces__colors">
          <button
            v-for="c in colorPresets"
            :key="c"
            class="teamspaces__color"
            :class="{ 'teamspaces__color--active': form.color === c }"
            :style="{ background: c }"
            @click="form.color = c"
          />
        </div>
      </div>

      <template #footer>
        <button class="teamspaces__ghost" @click="showForm = false">取消</button>
        <button class="teamspaces__primary" :disabled="saving" @click="save">
          {{ saving ? "保存中..." : "保存" }}
        </button>
          </AppModal>
  
    </div>
  </AppPageContainer>


    <!-- 创建/编辑弹窗 -->
    <AppModal
      :visible="showForm"
      :title="editingId ? '编辑团队空间' : '新建团队空间'"
      @close="showForm = false"
    >
      <div class="teamspaces__form">
        <label class="teamspaces__label">名称</label>
        <input
          v-model="form.name"
          class="teamspaces__input"
          placeholder="例如：平台研发组"
        />

        <label class="teamspaces__label">描述</label>
        <textarea
          v-model="form.description"
          class="teamspaces__textarea"
          rows="2"
          placeholder="这个团队负责什么"
        />

        <label class="teamspaces__label">标识色</label>
        <div class="teamspaces__colors">
          <button
            v-for="c in colorPresets"
            :key="c"
            class="teamspaces__color"
            :class="{ 'teamspaces__color--active': form.color === c }"
            :style="{ background: c }"
            @click="form.color = c"
          />
        </div>
      </div>

      <template #footer>
        <button class="teamspaces__ghost" @click="showForm = false">取消</button>
        <button class="teamspaces__primary" :disabled="saving" @click="save">
          {{ saving ? "保存中..." : "保存" }}
        </button>
      
    </AppModal>
  </div>


<style scoped>
.teamspaces {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  
  
}

.teamspaces__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.teamspaces__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--txt-primary);
}

.teamspaces__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--txt-secondary);
}

.teamspaces__primary {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-on-color);
  background: var(--accent);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  white-space: nowrap;
}

.teamspaces__primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.teamspaces__ghost {
  padding: 8px 16px;
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.teamspaces__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.teamspaces__card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--bg-surface);
  cursor: pointer;
  transition: box-shadow 0.15s ease, border-color 0.15s ease;
}

.teamspaces__card:hover {
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
}

.teamspaces__card-top {
  display: flex;
  align-items: center;
  gap: 8px;
}

.teamspaces__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.teamspaces__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--txt-primary);
}

.teamspaces__desc {
  margin: 0;
  font-size: 13px;
  color: var(--txt-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.teamspaces__stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  padding: 12px 0;
  border-top: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
}

.teamspaces__stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.teamspaces__stat-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--txt-primary);
}

.teamspaces__stat-value--overdue {
  color: var(--danger);
}

.teamspaces__stat-label {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.teamspaces__card-actions {
  display: flex;
  gap: 12px;
}

.teamspaces__link {
  font-size: 12px;
  color: var(--txt-secondary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 2px 0;
}

.teamspaces__link:hover {
  color: var(--accent);
}

.teamspaces__link--danger:hover {
  color: var(--danger);
}

.teamspaces__archived {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.teamspaces__archived-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--txt-tertiary);
}

.teamspaces__archived-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface-2);
}

.teamspaces__archived-name {
  flex: 1;
  font-size: 13px;
  color: var(--txt-secondary);
}

.teamspaces__form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.teamspaces__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-secondary);
}

.teamspaces__input,
.teamspaces__textarea {
  width: 100%;
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.teamspaces__colors {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.teamspaces__color {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
}

.teamspaces__color--active {
  border-color: var(--txt-primary);
}
</style>
