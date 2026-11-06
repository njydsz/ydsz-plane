<script setup lang="ts">
/**
 * CustomerDetailView — 客户详情。
 *
 * 展示客户信息、统计（按状态的请求数）、该客户的请求列表。
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  customerApi,
  type Customer,
  type CustomerRequest,
  type CustomerStats,
} from "@/api/services/customer";
import { AppBadge, AppEmptyState, AppErrorState, AppLoadingState, AppPageContainer } from "@/components";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const route = useRoute();
const router = useRouter();
const wsId = computed(() => Number(route.params.workspaceId));
const customerId = computed(() => Number(route.params.id));

const customer = ref<Customer | null>(null);
const stats = ref<CustomerStats | null>(null);
const requests = ref<CustomerRequest[]>([]);
const loading = ref(true);
const error = ref("");

const statusLabelMap: Record<string, string> = {
  new: "新建",
  acknowledged: "已确认",
  in_progress: "处理中",
  resolved: "已解决",
  closed: "已关闭",
};

async function load() {
  if (!wsId.value || !customerId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const [c, s, r] = await Promise.all([
      customerApi.get(wsId.value, customerId.value),
      customerApi.stats(wsId.value, customerId.value),
      customerApi.listRequests(wsId.value, { customer_id: customerId.value }),
    ]);
    customer.value = c;
    stats.value = s;
    requests.value = (r as { items?: CustomerRequest[] }).items ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载客户详情失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
watch(() => customerId.value, load);

async function remove() {
  if (!customer.value) return;
  if (!await confirm({ title: "删除客户", message: `确认删除客户「${customer.value.name}」？`, confirmText: "删除", variant: "danger" })) return;
  try {
    await customerApi.remove(wsId.value, customer.value.id);
    toast.success("已删除");
    router.push(`/${wsId.value}/customers`);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

function back() {
  router.push(`/${wsId.value}/customers`);
}
</script>

<template>
  
  <AppPageContainer max="1000px">
    <div class="cdetail">
    <div class="cdetail__header">
      <button class="cdetail__back" @click="back">← 返回</button>
      <div class="cdetail__head-main">
        <h1 class="cdetail__title">{{ customer?.name || "客户详情" }}</h1>
        <p v-if="customer?.notes" class="cdetail__desc">{{ customer.notes }}</p>
      </div>
      <button class="cdetail__link cdetail__link--danger" @click="remove">
        删除客户
      </button>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState v-else-if="!customer" title="客户不存在" description="该客户可能已被删除" />

    <template v-else>
      <div class="cdetail__columns">
        <!-- 客户信息 -->
        <section class="cdetail__section">
          <h2 class="cdetail__section-title">基本信息</h2>
          <dl class="cdetail__info">
            <div class="cdetail__info-row">
              <dt>阶段</dt>
              <dd>
                <AppBadge
                  :variant="
                    customer.stage === 'customer'
                      ? 'success'
                      : customer.stage === 'churned'
                        ? 'danger'
                        : 'default'
                  "
                >
                  {{ customer.stage }}
                </AppBadge>
              </dd>
            </div>
            <div class="cdetail__info-row">
              <dt>等级</dt>
              <dd>{{ customer.tier }}</dd>
            </div>
            <div class="cdetail__info-row">
              <dt>行业</dt>
              <dd>{{ customer.industry || "—" }}</dd>
            </div>
            <div class="cdetail__info-row">
              <dt>地区</dt>
              <dd>{{ customer.region || "—" }}</dd>
            </div>
            <div class="cdetail__info-row">
              <dt>邮箱</dt>
              <dd>{{ customer.email || "—" }}</dd>
            </div>
            <div class="cdetail__info-row">
              <dt>网站</dt>
              <dd>{{ customer.website || "—" }}</dd>
            </div>
            <div class="cdetail__info-row">
              <dt>员工数</dt>
              <dd>{{ customer.employees || "—" }}</dd>
            </div>
            <div class="cdetail__info-row">
              <dt>营收</dt>
              <dd>{{ customer.revenue || "—" }}</dd>
            </div>
          </dl>
        </section>

        <!-- 统计 -->
        <section class="cdetail__section">
          <h2 class="cdetail__section-title">请求统计</h2>
          <div class="cdetail__stats">
            <div class="cdetail__stat">
              <span class="cdetail__stat-value">{{ stats?.total_requests ?? 0 }}</span>
              <span class="cdetail__stat-label">总请求</span>
            </div>
            <div
              v-for="(count, status) in stats?.by_status ?? {}"
              :key="status"
              class="cdetail__stat"
            >
              <span class="cdetail__stat-value">{{ count }}</span>
              <span class="cdetail__stat-label">
                {{ statusLabelMap[status] ?? status }}
              </span>
            </div>
          </div>
        </section>
      </div>

      <!-- 该客户的请求 -->
      <section class="cdetail__section">
        <h2 class="cdetail__section-title">客户请求</h2>
        <div class="cdetail__list">
          <div
            v-for="r in requests"
            :key="r.id"
            class="cdetail__list-item"
            @click="
              router.push(`/${wsId}/customer-requests?focus=${r.id}`)
            "
          >
            <div class="cdetail__list-item-main">
              <span class="cdetail__list-item-title">{{ r.title }}</span>
              <AppBadge
                :variant="
                  r.status === 'resolved' || r.status === 'closed'
                    ? 'success'
                    : r.status === 'in_progress'
                      ? 'info'
                      : 'default'
                "
              >
                {{ statusLabelMap[r.status] ?? r.status }}
              </AppBadge>
            </div>
            <span class="cdetail__list-item-meta">
              {{ r.priority }} · {{ r.source }}
              <template v-if="r.work_item_id"> · 工作项 #{{ r.work_item_id }}</template>
            </span>
          </div>
        </div>
      </section>
    </template>
    </div>
  </AppPageContainer>
</template>


<style scoped>
.cdetail {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  
  
}

.cdetail__header {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.cdetail__back {
  padding: 6px 12px;
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.cdetail__head-main {
  flex: 1;
}

.cdetail__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--txt-primary);
}

.cdetail__desc {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--txt-secondary);
}

.cdetail__link {
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: none;
  cursor: pointer;
}

.cdetail__link--danger:hover {
  color: var(--danger);
}

.cdetail__columns {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 20px;
}

.cdetail__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--bg-surface);
}

.cdetail__section-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--txt-primary);
}

.cdetail__info {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.cdetail__info-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}

.cdetail__info-row dt {
  color: var(--txt-tertiary);
  flex-shrink: 0;
}

.cdetail__info-row dd {
  margin: 0;
  color: var(--txt-primary);
  text-align: right;
}

.cdetail__stats {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.cdetail__stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 10px 14px;
  border-radius: 8px;
  background: var(--bg-surface-2);
  min-width: 64px;
}

.cdetail__stat-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--txt-primary);
}

.cdetail__stat-label {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.cdetail__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cdetail__list-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.cdetail__list-item:hover {
  background: var(--bg-surface-2);
}

.cdetail__list-item-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.cdetail__list-item-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--txt-primary);
}

.cdetail__list-item-meta {
  font-size: 12px;
  color: var(--txt-tertiary);
}

.cdetail__empty {
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--txt-tertiary);
}
</style>
