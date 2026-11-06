<script setup lang="ts">
/**
 * CustomerRequestsView — 客户请求管理。
 *
 * 支持：列表 + 状态/优先级/客户筛选 + 新建/编辑 + 关联工作项 + 状态流转。
 */
import { computed, onMounted, ref } from "vue";

import {
  customerApi,
  type Customer,
  type CustomerRequest,
  type RequestStatus,
  type RequestPriority,
  type RequestSource,
} from "@/api/services/customer";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppBadge, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer, DynamicIcon } from "@/components";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const wsStore = useWorkspaceStore();
const wsId = computed(() => wsStore.current?.id ?? 0);

/* ---- 数据 ---- */

const requests = ref<CustomerRequest[]>([]);
const customers = ref<Customer[]>([]);
const loading = ref(true);
const error = ref("");

const statusFilter = ref<string>("");
const priorityFilter = ref<string>("");
const customerFilter = ref<number>(0);

/* ---- 表单 ---- */

const showForm = ref(false);
const editingId = ref<number | null>(null);
const saving = ref(false);
const form = ref({
  customer_id: 0,
  title: "",
  description: "",
  status: "new" as RequestStatus,
  priority: "medium" as RequestPriority,
  source: "email" as RequestSource,
  assigned_to: 0,
});

const statusOptions: Array<{ value: RequestStatus; label: string }> = [
  { value: "new", label: "新建" },
  { value: "acknowledged", label: "已确认" },
  { value: "in_progress", label: "处理中" },
  { value: "resolved", label: "已解决" },
  { value: "closed", label: "已关闭" },
];
const priorityOptions: Array<{ value: RequestPriority; label: string }> = [
  { value: "urgent", label: "紧急" },
  { value: "high", label: "高" },
  { value: "medium", label: "中" },
  { value: "low", label: "低" },
];
const sourceOptions: Array<{ value: RequestSource; label: string }> = [
  { value: "email", label: "邮件" },
  { value: "form", label: "表单" },
  { value: "in_app", label: "站内" },
  { value: "phone", label: "电话" },
];

function statusLabel(s: string): string {
  return statusOptions.find((o) => o.value === s)?.label ?? s;
}
function statusVariant(s: string): "default" | "success" | "info" | "warning" | "brand" | "danger" {
  switch (s) {
    case "resolved":
    case "closed":
      return "success";
    case "in_progress":
      return "info";
    case "acknowledged":
      return "warning";
    default:
      return "default";
  }
}
function priorityLabel(p: string): string {
  return priorityOptions.find((o) => o.value === p)?.label ?? p;
}
function priorityVariant(p: string): "default" | "success" | "info" | "warning" | "brand" | "danger" {
  switch (p) {
    case "urgent":
      return "danger";
    case "high":
      return "warning";
    default:
      return "default";
  }
}
function sourceLabel(s: string): string {
  return sourceOptions.find((o) => o.value === s)?.label ?? s;
}

function customerName(id: number): string {
  return customers.value.find((c) => c.id === id)?.name || `客户 #${id}`;
}

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const [reqRes, custRes] = await Promise.all([
      customerApi.listRequests(wsId.value, {
        customer_id: customerFilter.value || undefined,
        status: statusFilter.value || undefined,
        priority: priorityFilter.value || undefined,
      }),
      customerApi.list(wsId.value, { limit: 200 }),
    ]);
    requests.value = (reqRes as { items?: CustomerRequest[] }).items ?? [];
    customers.value = (custRes as { items?: Customer[] }).items ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载客户请求失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function openCreate() {
  editingId.value = null;
  form.value = {
    customer_id: customerFilter.value || 0,
    title: "",
    description: "",
    status: "new",
    priority: "medium",
    source: "email",
    assigned_to: 0,
  };
  showForm.value = true;
}

function openEdit(r: CustomerRequest) {
  editingId.value = r.id;
  form.value = {
    customer_id: r.customer_id,
    title: r.title,
    description: r.description,
    status: r.status,
    priority: r.priority,
    source: r.source,
    assigned_to: r.assigned_to,
  };
  showForm.value = true;
}

async function save() {
  if (!form.value.title.trim()) {
    toast.error("请输入请求标题");
    return;
  }
  if (!form.value.customer_id) {
    toast.error("请选择客户");
    return;
  }
  saving.value = true;
  try {
    if (editingId.value) {
      await customerApi.updateRequest(wsId.value, editingId.value, form.value);
      toast.success("请求已更新");
    } else {
      await customerApi.createRequest(wsId.value, form.value);
      toast.success("请求已创建");
    }
    showForm.value = false;
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function remove(r: CustomerRequest) {
  if (!await confirm({ title: "删除请求", message: `确认删除请求「${r.title}」？`, confirmText: "删除", variant: "danger" })) return;
  try {
    await customerApi.removeRequest(wsId.value, r.id);
    toast.success("已删除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

async function linkWorkItem(r: CustomerRequest) {
  const idStr = window.prompt("输入工作项 ID:");
  if (!idStr) return;
  const type = window.prompt("工作项类型 (requirement/task/defect):", "task") || "task";
  try {
    await customerApi.linkToWorkItem(wsId.value, r.id, Number(idStr), type);
    toast.success("已关联工作项");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "关联失败");
  }
}

async function setStatus(r: CustomerRequest, status: RequestStatus) {
  try {
    await customerApi.updateRequest(wsId.value, r.id, { status });
    toast.success("状态已更新");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "更新失败");
  }
}
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="creqs">
    <!-- 头部 -->
    <div class="creqs__header">
      <div>
        <h1 class="creqs__title">客户请求</h1>
        <p class="creqs__subtitle">跟踪客户提交的需求，关联到工作项处理</p>
      </div>
      <button class="creqs__primary" @click="openCreate"><DynamicIcon name="Plus" :size="14" /> 新建请求</button>
    </div>

    <!-- 筛选 -->
    <div class="creqs__filters">
      <select v-model="statusFilter" class="creqs__select" @change="load">
        <option value="">全部状态</option>
        <option v-for="o in statusOptions" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>
      <select v-model="priorityFilter" class="creqs__select" @change="load">
        <option value="">全部优先级</option>
        <option v-for="o in priorityOptions" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>
      <select v-model="customerFilter" class="creqs__select" @change="load">
        <option :value="0">全部客户</option>
        <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!requests.length"
      title="暂无客户请求"
      description="记录客户提出的需求与反馈"
    >
      <button class="creqs__primary" @click="openCreate"><DynamicIcon name="Plus" :size="14" /> 新建请求</button>
    </AppEmptyState>

    <div v-else class="creqs__list">
      <div v-for="r in requests" :key="r.id" class="creqs__card">
        <div class="creqs__card-top">
          <span class="creqs__card-title">{{ r.title }}</span>
          <span class="creqs__badges">
            <AppBadge :variant="statusVariant(r.status)">{{ statusLabel(r.status) }}</AppBadge>
            <AppBadge :variant="priorityVariant(r.priority)">
              {{ priorityLabel(r.priority) }}
            </AppBadge>
          </span>
        </div>

        <p v-if="r.description" class="creqs__desc">{{ r.description }}</p>

        <div class="creqs__card-meta">
          <span>客户：{{ customerName(r.customer_id) }}</span>
          <span>来源：{{ sourceLabel(r.source) }}</span>
          <span v-if="r.work_item_id">工作项：#{{ r.work_item_id }}</span>
          <span v-else>未关联工作项</span>
        </div>

        <div class="creqs__card-actions">
          <button class="creqs__link" @click="openEdit(r)">编辑</button>
          <button class="creqs__link" @click="linkWorkItem(r)">关联工作项</button>
          <button
            v-if="r.status === 'new'"
            class="creqs__link"
            @click="setStatus(r, 'acknowledged')"
          >
            确认
          </button>
          <button
            v-if="r.status === 'acknowledged'"
            class="creqs__link"
            @click="setStatus(r, 'in_progress')"
          >
            开始处理
          </button>
          <button
            v-if="r.status === 'in_progress'"
            class="creqs__link"
            @click="setStatus(r, 'resolved')"
          >
            标记解决
          </button>
          <button class="creqs__link creqs__link--danger" @click="remove(r)">删除</button>
        </div>
      </div>
    </div>

    <!-- 创建/编辑弹窗 -->
    <AppModal
      :visible="showForm"
      :title="editingId ? '编辑客户请求' : '新建客户请求'"
      @close="showForm = false"
    >
      <div class="creqs__form">
        <label class="creqs__label">客户 *</label>
        <select v-model="form.customer_id" class="creqs__input">
          <option :value="0" disabled>选择客户</option>
          <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>

        <label class="creqs__label">标题 *</label>
        <input v-model="form.title" class="creqs__input" placeholder="请求标题" />

        <label class="creqs__label">描述</label>
        <textarea
          v-model="form.description"
          class="creqs__textarea"
          rows="3"
          placeholder="客户的具体诉求"
        />

        <div class="creqs__grid">
          <div>
            <label class="creqs__label">状态</label>
            <select v-model="form.status" class="creqs__input">
              <option v-for="o in statusOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
          <div>
            <label class="creqs__label">优先级</label>
            <select v-model="form.priority" class="creqs__input">
              <option v-for="o in priorityOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
        </div>

        <label class="creqs__label">来源</label>
        <select v-model="form.source" class="creqs__input">
          <option v-for="o in sourceOptions" :key="o.value" :value="o.value">
            {{ o.label }}
          </option>
        </select>
      </div>

      <template #footer>
        <button class="creqs__ghost" @click="showForm = false">取消</button>
        <button class="creqs__primary" :disabled="saving" @click="save">
          {{ saving ? "保存中..." : "保存" }}
        </button>
      </template>
    </AppModal>
  
    </div>
  </AppPageContainer>
    </AppModal>
  </div>


<style scoped>
.creqs {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
  
  
}

.creqs__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.creqs__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--txt-primary);
}

.creqs__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--txt-secondary);
}

.creqs__primary {
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

.creqs__primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.creqs__ghost {
  padding: 8px 14px;
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.creqs__filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.creqs__select {
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.creqs__list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
}

.creqs__card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--bg-surface);
}

.creqs__card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.creqs__card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--txt-primary);
}

.creqs__badges {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.creqs__desc {
  margin: 0;
  font-size: 13px;
  color: var(--txt-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.creqs__card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
  color: var(--txt-tertiary);
}

.creqs__card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding-top: 8px;
  border-top: 1px solid var(--border-subtle);
}

.creqs__link {
  font-size: 12px;
  color: var(--txt-secondary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.creqs__link:hover {
  color: var(--accent);
}

.creqs__link--danger:hover {
  color: var(--danger);
}

.creqs__form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.creqs__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.creqs__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-secondary);
}

.creqs__input,
.creqs__textarea {
  width: 100%;
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}
</style>
