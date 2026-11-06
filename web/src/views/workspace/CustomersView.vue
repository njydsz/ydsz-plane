<script setup lang="ts">
/**
 * CustomersView — 客户记录管理。
 *
 * 支持：列表 + 阶段/等级/行业筛选 + 搜索 + 新建/编辑/删除。
 */
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import {
  customerApi,
  type Customer,
  type CustomerStage,
  type CustomerTier,
} from "@/api/services/customer";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppBadge, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer, DynamicIcon } from "@/components";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const router = useRouter();
const wsStore = useWorkspaceStore();
const wsId = computed(() => wsStore.current?.id ?? 0);

/* ---- 数据 ---- */

const customers = ref<Customer[]>([]);
const loading = ref(true);
const error = ref("");
const searchText = ref("");

/* ---- 筛选 ---- */

const stageFilter = ref<string>("");
const tierFilter = ref<string>("");
const industryFilter = ref<string>("");

const industries = computed(() => {
  const set = new Set<string>();
  customers.value.forEach((c) => {
    if (c.industry) set.add(c.industry);
  });
  return Array.from(set);
});

/* ---- 创建/编辑 ---- */

const showForm = ref(false);
const editingId = ref<number | null>(null);
const saving = ref(false);
const form = ref({
  name: "",
  email: "",
  website: "",
  industry: "",
  stage: "lead" as CustomerStage,
  tier: "free" as CustomerTier,
  region: "",
  employees: 0,
  revenue: "",
  account_manager_id: 0,
  notes: "",
});

const stageOptions: Array<{ value: CustomerStage; label: string }> = [
  { value: "lead", label: "线索" },
  { value: "prospect", label: "潜在" },
  { value: "customer", label: "客户" },
  { value: "churned", label: "流失" },
];
const tierOptions: Array<{ value: CustomerTier; label: string }> = [
  { value: "free", label: "免费版" },
  { value: "pro", label: "专业版" },
  { value: "enterprise", label: "企业版" },
];

function stageLabel(s: string): string {
  return stageOptions.find((o) => o.value === s)?.label ?? s;
}
function tierLabel(t: string): string {
  return tierOptions.find((o) => o.value === t)?.label ?? t;
}
function stageVariant(s: string): "default" | "success" | "warning" | "info" | "brand" | "danger" {
  switch (s) {
    case "customer":
      return "success";
    case "prospect":
      return "info";
    case "churned":
      return "danger";
    default:
      return "default";
  }
}

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const res = await customerApi.list(wsId.value, {
      stage: stageFilter.value || undefined,
      tier: tierFilter.value || undefined,
      industry: industryFilter.value || undefined,
      q: searchText.value || undefined,
    });
    customers.value = (res as { items?: Customer[] }).items ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载客户列表失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function openCreate() {
  editingId.value = null;
  form.value = {
    name: "",
    email: "",
    website: "",
    industry: "",
    stage: "lead",
    tier: "free",
    region: "",
    employees: 0,
    revenue: "",
    account_manager_id: 0,
    notes: "",
  };
  showForm.value = true;
}

function openEdit(c: Customer) {
  editingId.value = c.id;
  form.value = {
    name: c.name,
    email: c.email,
    website: c.website,
    industry: c.industry,
    stage: c.stage,
    tier: c.tier,
    region: c.region,
    employees: c.employees,
    revenue: c.revenue,
    account_manager_id: c.account_manager_id,
    notes: c.notes,
  };
  showForm.value = true;
}

async function save() {
  if (!form.value.name.trim()) {
    toast.error("请输入客户名称");
    return;
  }
  saving.value = true;
  try {
    if (editingId.value) {
      await customerApi.update(wsId.value, editingId.value, form.value);
      toast.success("客户已更新");
    } else {
      await customerApi.create(wsId.value, form.value);
      toast.success("客户已创建");
    }
    showForm.value = false;
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function remove(c: Customer) {
  if (!await confirm({ title: "删除客户", message: `确认删除客户「${c.name}」？此操作不可撤销。`, confirmText: "删除", variant: "danger" })) return;
  try {
    await customerApi.remove(wsId.value, c.id);
    toast.success("已删除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

function openDetail(c: Customer) {
  router.push(`/${wsId.value}/customers/${c.id}`);
}
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="customers">
    <!-- 头部 -->
    <div class="customers__header">
      <div>
        <h1 class="customers__title">客户</h1>
        <p class="customers__subtitle">管理客户记录，跟踪客户需求与关联工作项</p>
      </div>
      <button class="customers__primary" @click="openCreate">
        <DynamicIcon name="Plus" :size="14" /> 新建客户
      </button>
    </div>

    <!-- 筛选栏 -->
    <div class="customers__filters">
      <input
        v-model="searchText"
        class="customers__search"
        placeholder="搜索客户名称..."
        @keyup.enter="load"
      />
      <select v-model="stageFilter" class="customers__select" @change="load">
        <option value="">全部阶段</option>
        <option v-for="o in stageOptions" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>
      <select v-model="tierFilter" class="customers__select" @change="load">
        <option value="">全部等级</option>
        <option v-for="o in tierOptions" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>
      <select v-model="industryFilter" class="customers__select" @change="load">
        <option value="">全部行业</option>
        <option v-for="i in industries" :key="i" :value="i">{{ i }}</option>
      </select>
      <button class="customers__ghost" @click="load">查询</button>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!customers.length"
      title="暂无客户"
      description="创建客户记录以跟踪需求"
    >
      <button class="customers__primary" @click="openCreate"><DynamicIcon name="Plus" :size="14" /> 新建客户</button>
    </AppEmptyState>

    <div v-else class="customers__table">
      <div class="customers__table-head">
        <span>名称</span>
        <span>阶段</span>
        <span>等级</span>
        <span>行业</span>
        <span>邮箱</span>
        <span>请求数</span>
        <span>操作</span>
      </div>
      <div
        v-for="c in customers"
        :key="c.id"
        class="customers__row"
        @click="openDetail(c)"
      >
        <span class="customers__name">{{ c.name }}</span>
        <span>
          <AppBadge :variant="stageVariant(c.stage)">{{ stageLabel(c.stage) }}</AppBadge>
        </span>
        <span>{{ tierLabel(c.tier) }}</span>
        <span class="customers__muted">{{ c.industry || "—" }}</span>
        <span class="customers__muted">{{ c.email || "—" }}</span>
        <span>{{ c.request_count ?? 0 }}</span>
        <span class="customers__actions" @click.stop>
          <button class="customers__link" @click="openEdit(c)">编辑</button>
          <button class="customers__link customers__link--danger" @click="remove(c)">
            删除
          </button>
        </span>
      </div>
    </div>

    <!-- 创建/编辑弹窗 -->
    <AppModal
      :visible="showForm"
      :title="editingId ? '编辑客户' : '新建客户'"
      @close="showForm = false"
    >
      <div class="customers__form">
        <label class="customers__label">名称 *</label>
        <input v-model="form.name" class="customers__input" placeholder="客户名称" />

        <div class="customers__grid">
          <div>
            <label class="customers__label">邮箱</label>
            <input v-model="form.email" class="customers__input" placeholder="name@company.com" />
          </div>
          <div>
            <label class="customers__label">网站</label>
            <input v-model="form.website" class="customers__input" placeholder="https://" />
          </div>
        </div>

        <div class="customers__grid">
          <div>
            <label class="customers__label">阶段</label>
            <select v-model="form.stage" class="customers__input">
              <option v-for="o in stageOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
          <div>
            <label class="customers__label">等级</label>
            <select v-model="form.tier" class="customers__input">
              <option v-for="o in tierOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
        </div>

        <div class="customers__grid">
          <div>
            <label class="customers__label">行业</label>
            <input v-model="form.industry" class="customers__input" placeholder="例如：金融" />
          </div>
          <div>
            <label class="customers__label">地区</label>
            <input v-model="form.region" class="customers__input" placeholder="例如：华东" />
          </div>
        </div>

        <div class="customers__grid">
          <div>
            <label class="customers__label">员工数</label>
            <input
              v-model.number="form.employees"
              type="number"
              class="customers__input"
              min="0"
            />
          </div>
          <div>
            <label class="customers__label">营收</label>
            <input v-model="form.revenue" class="customers__input" placeholder="例如：100万" />
          </div>
        </div>

        <label class="customers__label">备注</label>
        <textarea
          v-model="form.notes"
          class="customers__textarea"
          rows="3"
          placeholder="客户背景、沟通记录等"
        />
      </div>

      <template #footer>
        <button class="customers__ghost" @click="showForm = false">取消</button>
        <button class="customers__primary" :disabled="saving" @click="save">
          {{ saving ? "保存中..." : "保存" }}
        </button>
      </template>
    </AppModal>
  
    </div>
  </AppPageContainer>
    </AppModal>
  </div>


<style scoped>
.customers {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
  
  
}

.customers__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.customers__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--txt-primary);
}

.customers__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--txt-secondary);
}

.customers__primary {
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

.customers__primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.customers__ghost {
  padding: 8px 14px;
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.customers__filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.customers__search {
  flex: 1;
  min-width: 200px;
  padding: 8px 12px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.customers__select {
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.customers__table {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  overflow: hidden;
  background: var(--bg-surface);
}

.customers__table-head,
.customers__row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 1.5fr 0.8fr 1.2fr;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}

.customers__table-head {
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-tertiary);
  background: var(--bg-surface-2);
  border-bottom: 1px solid var(--border-subtle);
}

.customers__row {
  font-size: 13px;
  color: var(--txt-primary);
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
}

.customers__row:last-child {
  border-bottom: none;
}

.customers__row:hover {
  background: var(--bg-surface-2);
}

.customers__name {
  font-weight: 500;
}

.customers__muted {
  color: var(--txt-tertiary);
}

.customers__actions {
  display: flex;
  gap: 10px;
}

.customers__link {
  font-size: 12px;
  color: var(--txt-secondary);
  background: none;
  border: none;
  cursor: pointer;
}

.customers__link:hover {
  color: var(--accent);
}

.customers__link--danger:hover {
  color: var(--danger);
}

.customers__form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.customers__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.customers__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-secondary);
}

.customers__input,
.customers__textarea {
  width: 100%;
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}
</style>
