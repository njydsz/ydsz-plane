<script setup lang="ts">
/**
 * EstimatePointsView — 估算点数配置页。
 * 项目管理员可在此创建多套估算点集（斐波那契、T恤码等），
 * 工作项发布时从中选择规模评估。
 */
import { onMounted, ref, computed } from "vue";
import { useRoute } from "vue-router";

import {
  estimatePointApi,
  type EstimatePoint,
  type EstimatePointDef,
} from "@/api/services/estimatePoints";
import { AppEmptyState, AppErrorState, AppLoadingState } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const points = ref<EstimatePoint[]>([]);
const loading = ref(true);
const error = ref("");

// 创建
const showCreate = ref(false);
const createForm = ref({ name: "", description: "", is_default: false });
const createPointDefs = ref<EstimatePointDef[]>([{ label: "1", value: 1, color: "#22c55e" }]);
const creating = ref(false);

// 编辑
const editingId = ref<number | null>(null);
const editForm = ref({ name: "", description: "", is_default: false });
const editPointDefs = ref<EstimatePointDef[]>([]);
const saving = ref(false);

// 默认点集
const defaultPoints = computed(() => points.value.filter((p) => p.is_default));

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const r = await estimatePointApi.list(workspaceId, projectId);
    points.value = r.results ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function addPointDef(target: { value: EstimatePointDef[] }) {
  target.value.push({ label: "", value: target.value.length + 1, color: "#3b82f6" });
}

function removePointDef(target: { value: EstimatePointDef[] }, idx: number) {
  if (target.value.length <= 1) return;
  target.value.splice(idx, 1);
}

function resetCreateForm() {
  createForm.value = { name: "", description: "", is_default: false };
  createPointDefs.value = [{ label: "1", value: 1, color: "#22c55e" }];
}

async function doCreate() {
  if (!createForm.value.name.trim()) {
    toast.warning("请输入名称");
    return;
  }
  if (createPointDefs.value.some((p) => !p.label.trim())) {
    toast.warning("每个点数都需要标签名");
    return;
  }
  creating.value = true;
  try {
    await estimatePointApi.create(workspaceId, projectId, {
      name: createForm.value.name.trim(),
      description: createForm.value.description.trim() || undefined,
      points: createPointDefs.value,
      is_default: createForm.value.is_default,
    });
    toast.success("创建成功");
    showCreate.value = false;
    resetCreateForm();
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "创建失败");
  } finally {
    creating.value = false;
  }
}

function startEdit(p: EstimatePoint) {
  editingId.value = p.id;
  editForm.value = {
    name: p.name,
    description: p.description ?? "",
    is_default: p.is_default,
  };
  editPointDefs.value = p.points?.length
    ? JSON.parse(JSON.stringify(p.points))
    : [{ label: "1", value: 1, color: "#22c55e" }];
}

async function saveEdit(p: EstimatePoint) {
  if (!editForm.value.name.trim()) {
    toast.warning("名称不能为空");
    return;
  }
  saving.value = true;
  try {
    await estimatePointApi.update(workspaceId, projectId, p.id, {
      name: editForm.value.name.trim(),
      description: editForm.value.description.trim() || undefined,
      points: editPointDefs.value,
      is_default: editForm.value.is_default,
    });
    toast.success("已保存");
    editingId.value = null;
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function remove(p: EstimatePoint) {
  if (p.is_default) {
    toast.warning("默认估算点数无法删除");
    return;
  }
  if (!confirm(`确定删除「${p.name}」？`)) return;
  try {
    await estimatePointApi.remove(workspaceId, projectId, p.id);
    toast.success("已删除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <!-- 头部 -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
          估算点数配置
        </h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          管理工作量规模评估参照系（如斐波那契数列、T恤码等）
        </p>
      </div>
      <button
        class="inline-flex items-center rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        @click="showCreate = true; resetCreateForm()"
      >
        + 新建点集
      </button>
    </div>

    <!-- 加载 / 错误 -->
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!points.length"
      description="还没有配置估算点数，工作项将无法评估规模"
      action-label="立即创建"
      @action="showCreate = true; resetCreateForm()"
    />

    <!-- 点集列表 -->
    <div v-else class="space-y-4">
      <div
        v-for="p in points"
        :key="p.id"
        class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800"
      >
        <!-- 查看 -->
        <div v-if="editingId !== p.id">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100">
                  {{ p.name }}
                </h3>
                <span
                  v-if="p.is_default"
                  class="rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900 dark:text-blue-300"
                >
                  默认
                </span>
              </div>
              <p v-if="p.description" class="mt-1 text-sm text-gray-500">
                {{ p.description }}
              </p>
              <div class="mt-3 flex flex-wrap gap-1.5">
                <span
                  v-for="(pt, idx) in p.points"
                  :key="idx"
                  class="rounded-md px-2 py-0.5 text-xs font-medium"
                  :style="{ backgroundColor: pt.color + '20', color: pt.color }"
                >
                  {{ pt.label }} ({{ pt.value }})
                </span>
              </div>
            </div>
            <div class="flex gap-2">
              <button
                class="text-sm text-blue-600 hover:text-blue-700"
                @click="startEdit(p)"
              >
                编辑
              </button>
              <button
                class="text-sm text-red-600 hover:text-red-700 disabled:opacity-40"
                :disabled="p.is_default"
                @click="remove(p)"
              >
                删除
              </button>
            </div>
          </div>
        </div>

        <!-- 编辑表单 -->
        <div v-else class="space-y-3">
          <input
            v-model="editForm.name"
            class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
            placeholder="名称，如：斐波那契点集"
          />
          <input
            v-model="editForm.description"
            class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
            placeholder="描述（可选）"
          />

          <div>
            <div class="mb-1 flex items-center justify-between">
              <span class="text-xs font-medium text-gray-600 dark:text-gray-400">点值配置</span>
              <button
                class="text-xs text-blue-600"
                @click="addPointDef({ value: editPointDefs })"
              >
                + 添加
              </button>
            </div>
            <div class="space-y-1.5">
              <div
                v-for="(pt, idx) in editPointDefs"
                :key="idx"
                class="flex items-center gap-2"
              >
                <input
                  v-model="pt.label"
                  class="w-20 rounded border border-gray-300 px-2 py-1 text-xs dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
                  placeholder="标签"
                />
                <input
                  v-model.number="pt.value"
                  type="number"
                  class="w-20 rounded border border-gray-300 px-2 py-1 text-xs dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
                  placeholder="值"
                />
                <input
                  v-model="pt.color"
                  type="color"
                  class="h-7 w-8 cursor-pointer rounded border border-gray-300"
                />
                <button
                  class="text-xs text-red-500 hover:text-red-700 disabled:opacity-30"
                  :disabled="editPointDefs.length <= 1"
                  @click="removePointDef({ value: editPointDefs }, idx)"
                >
                  移除
                </button>
              </div>
            </div>
          </div>

          <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input v-model="editForm.is_default" type="checkbox" />
            设为默认点集
          </label>

          <div class="flex gap-2 pt-2">
            <button
              :disabled="saving"
              class="rounded-md bg-blue-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              @click="saveEdit(p)"
            >
              {{ saving ? "保存中..." : "保存" }}
            </button>
            <button
              class="rounded-md border border-gray-300 px-4 py-1.5 text-sm text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300"
              @click="editingId = null"
            >
              取消
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建弹窗 -->
    <Teleport to="body">
      <div
        v-if="showCreate"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      >
        <div class="mx-4 w-full max-w-lg rounded-lg bg-white p-6 shadow-xl dark:bg-gray-800">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            新建估算点集
          </h3>

          <div class="mt-4 space-y-3">
            <input
              v-model="createForm.name"
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
              placeholder="名称，如：斐波那契点集"
            />
            <input
              v-model="createForm.description"
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
              placeholder="描述（可选）"
            />

            <div>
              <div class="mb-1 flex items-center justify-between">
                <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
                  点值配置
                </span>
                <button
                  class="text-xs text-blue-600"
                  @click="addPointDef({ value: createPointDefs })"
                >
                  + 添加
                </button>
              </div>
              <div class="space-y-1.5">
                <div
                  v-for="(pt, idx) in createPointDefs"
                  :key="idx"
                  class="flex items-center gap-2"
                >
                  <input
                    v-model="pt.label"
                    class="w-20 rounded border border-gray-300 px-2 py-1 text-xs dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
                    placeholder="标签"
                  />
                  <input
                    v-model.number="pt.value"
                    type="number"
                    class="w-20 rounded border border-gray-300 px-2 py-1 text-xs dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
                    placeholder="值"
                  />
                  <input
                    v-model="pt.color"
                    type="color"
                    class="h-7 w-8 cursor-pointer rounded border border-gray-300"
                  />
                  <button
                    class="text-xs text-red-500 hover:text-red-700 disabled:opacity-30"
                    :disabled="createPointDefs.length <= 1"
                    @click="removePointDef({ value: createPointDefs }, idx)"
                  >
                    移除
                  </button>
                </div>
              </div>
            </div>

            <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <input v-model="createForm.is_default" type="checkbox" />
              设为默认点集
            </label>
          </div>

          <div class="mt-5 flex justify-end gap-2">
            <button
              :disabled="creating"
              class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              @click="doCreate"
            >
              {{ creating ? "创建中..." : "创建" }}
            </button>
            <button
              class="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300"
              @click="showCreate = false"
            >
              取消
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
