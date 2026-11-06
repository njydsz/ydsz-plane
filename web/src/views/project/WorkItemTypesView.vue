<script setup lang="ts">
/**
 * WorkItemTypesView — 自定义工作项类型管理页。
 *
 * 能力：
 *  - 类型列表展示（名称 / 标识 / 层级 / 颜色 / 属性数 / 启用与默认徽章）
 *  - 新建 / 编辑类型（名称、标识、颜色、图标、描述、层级、启用、默认）
 *  - 每个类型下挂载自定义属性（名称、标识、字段类型、选项、必填、排序）
 *  - 归档 / 恢复 / 删除
 */
import { computed, onMounted, ref } from "vue";

import {
  workItemTypeApi,
  type CreatePropertyInput,
  type CreateWorkItemTypeInput,
  type FieldType,
  type UpdateWorkItemTypeInput,
  type WorkItemType,
  type WorkItemTypeProperty,
} from "@/api/services/workitemtype";
import {
  AppBadge,
  AppButton,
  AppEmptyState,
  AppErrorState,
  AppLoadingState,
  AppModal,
  DynamicIcon,
} from "@/components";
import { useTranslation } from "@/composables/useTranslation";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const props = defineProps<{
  workspaceId: number;
  projectId: number;
}>();

const { t } = useTranslation();

/* ------------------------------------------------------------------ */
/* 常量                                                                */
/* ------------------------------------------------------------------ */

/** 预设颜色板 */
const COLOR_PRESETS = [
  "var(--danger-500)",
  "var(--warning-500)",
  "var(--warning-500)",
  "var(--success-500)",
  "var(--info-500)",
  "var(--brand-500)",
  "var(--brand-500)",
  "var(--brand-500)",
  "var(--brand-500)",
  "var(--text-secondary)",
];

/** 可选图标（与 DynamicIcon 注册的 lucide 名对齐） */
const ICON_PRESETS = [
  "Shapes",
  "Tag",
  "Bug",
  "FileText",
  "Target",
  "Zap",
  "Clock",
  "Calendar",
  "Gauge",
  "Link",
];

/** 字段类型选项 */
const FIELD_TYPE_OPTIONS: { value: FieldType; labelKey: string }[] = [
  { value: "text", labelKey: "workItemType.fieldTypeText" },
  { value: "number", labelKey: "workItemType.fieldTypeNumber" },
  { value: "select", labelKey: "workItemType.fieldTypeSelect" },
  { value: "multiselect", labelKey: "workItemType.fieldTypeMultiselect" },
  { value: "date", labelKey: "workItemType.fieldTypeDate" },
  { value: "checkbox", labelKey: "workItemType.fieldTypeCheckbox" },
  { value: "url", labelKey: "workItemType.fieldTypeURL" },
];

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const types = ref<WorkItemType[]>([]);

/** 属性缓存：typeId -> properties */
const propertyMap = ref<Record<number, WorkItemTypeProperty[]>>({});
/** 已展开属性的类型 id 集合 */
const expandedIds = ref<Set<number>>(new Set());

const hasOptionField = (ft: FieldType) => ft === "select" || ft === "multiselect";

async function loadTypes() {
  loading.value = true;
  error.value = "";
  try {
    const res = await workItemTypeApi.list(props.workspaceId, props.projectId);
    types.value = res.items ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t("workItemType.loadFailed");
  } finally {
    loading.value = false;
  }
}

function toggleExpand(type: WorkItemType) {
  const next = new Set(expandedIds.value);
  if (next.has(type.id)) {
    next.delete(type.id);
  } else {
    next.add(type.id);
    if (!(type.id in propertyMap.value)) {
      void loadProperties(type);
    }
  }
  expandedIds.value = next;
}

async function loadProperties(type: WorkItemType) {
  try {
    const res = await workItemTypeApi.listProperties(
      props.workspaceId,
      props.projectId,
      type.id,
    );
    propertyMap.value = { ...propertyMap.value, [type.id]: res.items ?? [] };
  } catch {
    propertyMap.value = { ...propertyMap.value, [type.id]: [] };
  }
}

/* ------------------------------------------------------------------ */
/* 类型表单                                                            */
/* ------------------------------------------------------------------ */

const showTypeModal = ref(false);
const editingType = ref<WorkItemType | null>(null);
const savingType = ref(false);

const typeForm = ref<CreateWorkItemTypeInput>({
  name: "",
  code: "",
  color: COLOR_PRESETS[5],
  icon: "Shapes",
  description: "",
  level: 0,
  is_default: false,
});

function resetTypeForm() {
  typeForm.value = {
    name: "",
    code: "",
    color: COLOR_PRESETS[5],
    icon: "Shapes",
    description: "",
    level: 0,
    is_default: false,
  };
  editingType.value = null;
}

function openCreateType() {
  resetTypeForm();
  showTypeModal.value = true;
}

function openEditType(type: WorkItemType) {
  editingType.value = type;
  typeForm.value = {
    name: type.name,
    code: type.code,
    color: type.color || COLOR_PRESETS[5],
    icon: type.icon || "Shapes",
    description: type.description ?? "",
    level: type.level ?? 0,
    is_default: type.is_default,
  };
  showTypeModal.value = true;
}

async function saveType() {
  if (!typeForm.value.name.trim()) {
    toast.warning(t("workItemType.nameRequired"));
    return;
  }
  if (!typeForm.value.code.trim()) {
    toast.warning(t("workItemType.codeRequired"));
    return;
  }
  savingType.value = true;
  try {
    if (editingType.value) {
      const payload: UpdateWorkItemTypeInput = {
        ...typeForm.value,
        name: typeForm.value.name.trim(),
        code: typeForm.value.code.trim(),
        description: typeForm.value.description?.trim() || undefined,
      };
      await workItemTypeApi.update(
        props.workspaceId,
        props.projectId,
        editingType.value.id,
        payload,
      );
      toast.success(t("workItemType.updateSuccess"));
    } else {
      await workItemTypeApi.create(props.workspaceId, props.projectId, {
        ...typeForm.value,
        name: typeForm.value.name.trim(),
        code: typeForm.value.code.trim(),
        description: typeForm.value.description?.trim() || undefined,
      });
      toast.success(t("workItemType.createSuccess"));
    }
    showTypeModal.value = false;
    resetTypeForm();
    await loadTypes();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workItemType.loadFailed"));
  } finally {
    savingType.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 归档 / 恢复 / 删除                                                  */
/* ------------------------------------------------------------------ */

async function archiveType(type: WorkItemType) {
  if (!await confirm({ title: "归档类型", message: t("workItemType.confirmArchive", { name: type.name }), confirmText: "归档", variant: "danger" })) return;
  try {
    await workItemTypeApi.archive(props.workspaceId, props.projectId, type.id);
    toast.success(t("workItemType.archiveSuccess"));
    await loadTypes();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workItemType.loadFailed"));
  }
}

async function restoreType(type: WorkItemType) {
  try {
    await workItemTypeApi.restore(props.workspaceId, props.projectId, type.id);
    toast.success(t("workItemType.restoreSuccess"));
    await loadTypes();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workItemType.loadFailed"));
  }
}

async function deleteType(type: WorkItemType) {
  if (!await confirm({ title: "删除类型", message: t("workItemType.confirmDelete", { name: type.name }), confirmText: "删除", variant: "danger" })) return;
  try {
    await workItemTypeApi.remove(props.workspaceId, props.projectId, type.id);
    toast.success(t("workItemType.deleteSuccess"));
    await loadTypes();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workItemType.loadFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 属性表单                                                            */
/* ------------------------------------------------------------------ */

const showPropertyModal = ref(false);
const editingProperty = ref<WorkItemTypeProperty | null>(null);
const propertyType = ref<WorkItemType | null>(null);
const savingProperty = ref(false);

const propertyForm = ref<CreatePropertyInput>({
  name: "",
  code: "",
  field_type: "text",
  options: "",
  required: false,
  sort_order: 0,
});

function resetPropertyForm() {
  propertyForm.value = {
    name: "",
    code: "",
    field_type: "text",
    options: "",
    required: false,
    sort_order: 0,
  };
  editingProperty.value = null;
}

function openAddProperty(type: WorkItemType) {
  propertyType.value = type;
  resetPropertyForm();
  showPropertyModal.value = true;
}

function openEditProperty(type: WorkItemType, prop: WorkItemTypeProperty) {
  propertyType.value = type;
  editingProperty.value = prop;
  propertyForm.value = {
    name: prop.name,
    code: prop.code,
    field_type: prop.field_type,
    options: prop.options ?? "",
    required: prop.required,
    sort_order: prop.sort_order ?? 0,
  };
  showPropertyModal.value = true;
}

/** 将逗号分隔字符串序列化为 JSON 数组字符串（后端以 TEXT 存储） */
function serializeOptions(raw: string): string {
  const list = raw
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
  return list.length ? JSON.stringify(list) : "";
}

async function saveProperty() {
  if (!propertyType.value) return;
  if (!propertyForm.value.name.trim()) {
    toast.warning(t("workItemType.propertyNameRequired"));
    return;
  }
  if (!propertyForm.value.code.trim()) {
    toast.warning(t("workItemType.propertyCodeRequired"));
    return;
  }
  savingProperty.value = true;
  try {
    const payload: CreatePropertyInput = {
      ...propertyForm.value,
      name: propertyForm.value.name.trim(),
      code: propertyForm.value.code.trim(),
      options: hasOptionField(propertyForm.value.field_type)
        ? serializeOptions(propertyForm.value.options ?? "")
        : "",
    };
    if (editingProperty.value) {
      // 后端未提供属性更新接口：先删后加，保持列表一致
      await workItemTypeApi.removeProperty(
        props.workspaceId,
        props.projectId,
        propertyType.value.id,
        editingProperty.value.id,
      );
    }
    await workItemTypeApi.addProperty(
      props.workspaceId,
      props.projectId,
      propertyType.value.id,
      payload,
    );
    toast.success(t("workItemType.propertyCreateSuccess"));
    showPropertyModal.value = false;
    resetPropertyForm();
    await loadProperties(propertyType.value);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workItemType.loadFailed"));
  } finally {
    savingProperty.value = false;
  }
}

async function deleteProperty(type: WorkItemType, prop: WorkItemTypeProperty) {
  if (!await confirm({ title: "删除属性", message: t("workItemType.confirmDelete", { name: prop.name }), confirmText: "删除", variant: "danger" })) return;
  try {
    await workItemTypeApi.removeProperty(
      props.workspaceId,
      props.projectId,
      type.id,
      prop.id,
    );
    toast.success(t("workItemType.propertyDeleteSuccess"));
    await loadProperties(type);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workItemType.loadFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function fieldTypeLabel(ft: FieldType): string {
  const hit = FIELD_TYPE_OPTIONS.find((o) => o.value === ft);
  return hit ? t(hit.labelKey) : ft;
}

function levelLabel(level: number): string {
  return level === 1 ? t("workItemType.levelEpic") : t("workItemType.levelTask");
}

const totalCount = computed(() => types.value.length);

onMounted(loadTypes);
</script>

<template>
  <div class="mx-auto max-w-5xl space-y-6">
    <!-- 头部 -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">
          {{ t("workItemType.title") }}
        </h1>
        <p class="mt-1 text-sm text-[var(--text-tertiary)]">
          {{ t("workItemType.subtitle") }}
        </p>
      </div>
      <AppButton @click="openCreateType">
        <DynamicIcon name="Plus" :size="14" /> {{ t("workItemType.newType") }}
      </AppButton>
    </div>

    <!-- 加载 / 错误 / 空态 -->
    <AppLoadingState v-if="loading" :text="t('common.loading')" />
    <AppErrorState v-else-if="error" :message="error" :retry="t('common.retry')" @retry="loadTypes" />

    <template v-else>
      <AppEmptyState
        v-if="!types.length"
        icon="🗂️"
        :title="t('workItemType.emptyTitle')"
        :description="t('workItemType.emptyDesc')"
        :cta-text="t('workItemType.newType')"
        @cta-click="openCreateType"
      />

      <!-- 类型卡片列表 -->
      <div v-else class="space-y-3">
        <div
          v-for="type in types"
          :key="type.id"
          class="rounded-lg border border-[var(--border-subtle)] bg-[var(--surface-1)]"
        >
          <!-- 类型主行 -->
          <div class="flex flex-wrap items-center gap-3 px-4 py-3">
            <span
              class="inline-block h-3.5 w-3.5 shrink-0 rounded-full"
              :style="{ backgroundColor: type.color || 'var(--brand-500)' }"
            />
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span class="truncate text-sm font-semibold">{{ type.name }}</span>
                <span class="font-mono text-xs text-[var(--text-tertiary)]">{{ type.code }}</span>
                <AppBadge variant="info">{{ levelLabel(type.level) }}</AppBadge>
                <AppBadge v-if="type.is_default" variant="brand">
                  {{ t("workItemType.default") }}
                </AppBadge>
                <AppBadge v-if="!type.is_active" variant="default">
                  {{ t("workItemType.inactive") }}
                </AppBadge>
              </div>
              <p
                v-if="type.description"
                class="mt-0.5 truncate text-xs text-[var(--text-tertiary)]"
              >
                {{ type.description }}
              </p>
            </div>

            <span class="shrink-0 text-xs text-[var(--text-tertiary)]">
              {{
                t("workItemType.propertyCount", {
                  count: propertyMap[type.id]?.length ?? 0,
                })
              }}
            </span>

            <div class="flex shrink-0 items-center gap-2">
              <AppButton size="sm" variant="ghost" @click="toggleExpand(type)">
                {{ t("workItemType.properties") }}
              </AppButton>
              <AppButton size="sm" variant="secondary" @click="openEditType(type)">
                {{ t("common.edit") }}
              </AppButton>
              <AppButton
                v-if="type.is_active"
                size="sm"
                variant="ghost"
                @click="archiveType(type)"
              >
                {{ t("workItemType.archiveType") }}
              </AppButton>
              <AppButton
                v-else
                size="sm"
                variant="ghost"
                @click="restoreType(type)"
              >
                {{ t("workItemType.restoreType") }}
              </AppButton>
              <AppButton size="sm" variant="danger" @click="deleteType(type)">
                {{ t("common.delete") }}
              </AppButton>
            </div>
          </div>

          <!-- 属性展开区 -->
          <div
            v-if="expandedIds.has(type.id)"
            class="border-t border-[var(--border-subtle)] px-4 py-3"
          >
            <div class="mb-2 flex items-center justify-between">
              <span class="text-xs font-semibold text-[var(--text-secondary)]">
                {{ t("workItemType.properties") }}
              </span>
              <AppButton size="sm" variant="secondary" @click="openAddProperty(type)">
                <DynamicIcon name="Plus" :size="14" /> {{ t("workItemType.addProperty") }}
              </AppButton>
            </div>

            <p
              v-if="!propertyMap[type.id] || !propertyMap[type.id].length"
              class="py-2 text-xs text-[var(--text-tertiary)]"
            >
              {{ t("workItemType.noProperties") }}
            </p>

            <div v-else class="space-y-1.5">
              <div
                v-for="prop in propertyMap[type.id]"
                :key="prop.id"
                class="flex items-center gap-3 rounded-md border border-[var(--border-subtle)] px-3 py-2"
              >
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="truncate text-sm font-medium">{{ prop.name }}</span>
                    <span class="font-mono text-xs text-[var(--text-tertiary)]">
                      {{ prop.code }}
                    </span>
                    <AppBadge variant="default">{{ fieldTypeLabel(prop.field_type) }}</AppBadge>
                    <AppBadge v-if="prop.required" variant="warning">
                      {{ t("workItemType.required") }}
                    </AppBadge>
                  </div>
                  <p
                    v-if="hasOptionField(prop.field_type) && prop.options"
                    class="mt-0.5 truncate text-xs text-[var(--text-tertiary)]"
                  >
                    {{ prop.options }}
                  </p>
                </div>
                <div class="flex shrink-0 items-center gap-2">
                  <span class="text-xs text-[var(--text-tertiary)]">
                    #{{ prop.sort_order }}
                  </span>
                  <button
                    class="text-xs text-[var(--text-secondary)] hover:text-[var(--brand-600)]"
                    @click="openEditProperty(type, prop)"
                  >
                    {{ t("common.edit") }}
                  </button>
                  <button
                    class="text-xs text-[var(--danger)]"
                    @click="deleteProperty(type, prop)"
                  >
                    {{ t("common.delete") }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <p class="text-xs text-[var(--text-tertiary)]">
        {{ t("workItemType.propertyCount", { count: totalCount }) }}
      </p>
    </template>

    <!-- 类型新建/编辑弹窗 -->
    <AppModal
      :visible="showTypeModal"
      :title="editingType ? t('workItemType.editType') : t('workItemType.newType')"
      width="560px"
      @close="showTypeModal = false"
    >
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("workItemType.typeName") }} *
            </label>
            <input
              v-model="typeForm.name"
              class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
              :placeholder="t('workItemType.typeNamePlaceholder')"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("workItemType.typeCode") }} *
            </label>
            <input
              v-model="typeForm.code"
              class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 font-mono text-sm"
              :placeholder="t('workItemType.typeCodePlaceholder')"
            />
          </div>
        </div>

        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("workItemType.color") }}
          </label>
          <div class="flex items-center gap-2">
            <input
              v-model="typeForm.color"
              type="color"
              class="h-8 w-10 cursor-pointer rounded border border-[var(--border-subtle)] p-0.5"
            />
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="c in COLOR_PRESETS"
                :key="c"
                class="h-6 w-6 rounded-full border border-black/10"
                :class="{ 'ring-2 ring-[var(--brand-500)] ring-offset-1': typeForm.color === c }"
                :style="{ backgroundColor: c }"
                @click="typeForm.color = c"
              />
            </div>
          </div>
        </div>

        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("workItemType.icon") }}
          </label>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="ic in ICON_PRESETS"
              :key="ic"
              class="flex h-8 w-8 items-center justify-center rounded-md border border-[var(--border-subtle)] text-[var(--text-secondary)] hover:bg-[var(--surface-2)]"
              :class="{
                'border-[var(--brand-500)] bg-[var(--brand-50)] text-[var(--brand-600)]':
                  typeForm.icon === ic,
              }"
              :title="ic"
              @click="typeForm.icon = ic"
            >
              <DynamicIcon :name="ic" :size="16" />
            </button>
          </div>
        </div>

        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("workItemType.description") }}
          </label>
          <textarea
            v-model="typeForm.description"
            rows="2"
            class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
            :placeholder="t('workItemType.descriptionPlaceholder')"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("workItemType.level") }}
            </label>
            <select
              v-model.number="typeForm.level"
              class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
            >
              <option :value="0">{{ t("workItemType.levelTask") }}</option>
              <option :value="1">{{ t("workItemType.levelEpic") }}</option>
            </select>
          </div>
          <div class="flex flex-col justify-end gap-2">
            <label class="flex items-center gap-2 text-sm">
              <input v-model="typeForm.is_default" type="checkbox" />
              <span>{{ t("workItemType.isDefault") }}</span>
            </label>
          </div>
        </div>
      </div>

      <template #footer>
        <AppButton variant="secondary" @click="showTypeModal = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton :loading="savingType" @click="saveType">
          {{ t("common.save") }}
        </AppButton>
      </template>
          </AppModal>

    <!-- 属性新建/编辑弹窗 -->
    <AppModal
      :visible="showPropertyModal"
      :title="
        editingProperty
          ? t('workItemType.editProperty')
          : t('workItemType.addProperty')
      "
      width="520px"
      @close="showPropertyModal = false"
    >
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("workItemType.propertyName") }} *
            </label>
            <input
              v-model="propertyForm.name"
              class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
              :placeholder="t('workItemType.propertyNamePlaceholder')"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("workItemType.propertyCode") }} *
            </label>
            <input
              v-model="propertyForm.code"
              class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 font-mono text-sm"
              :placeholder="t('workItemType.propertyCodePlaceholder')"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("workItemType.fieldType") }}
            </label>
            <select
              v-model="propertyForm.field_type"
              class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
            >
              <option
                v-for="opt in FIELD_TYPE_OPTIONS"
                :key="opt.value"
                :value="opt.value"
              >
                {{ t(opt.labelKey) }}
              </option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("workItemType.sortOrder") }}
            </label>
            <input
              v-model.number="propertyForm.sort_order"
              type="number"
              class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
            />
          </div>
        </div>

        <div v-if="hasOptionField(propertyForm.field_type)">
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("workItemType.options") }}
          </label>
          <input
            v-model="propertyForm.options"
            class="w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
            :placeholder="t('workItemType.optionsPlaceholder')"
          />
        </div>

        <label class="flex items-center gap-2 text-sm">
          <input v-model="propertyForm.required" type="checkbox" />
          <span>{{ t("workItemType.required") }}</span>
        </label>
      </div>

      <template #footer>
        <AppButton variant="secondary" @click="showPropertyModal = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton :loading="savingProperty" @click="saveProperty">
          {{ t("common.save") }}
        </AppButton>
      
      </template>
    </AppModal>
  </div>
</template>


