<script setup lang="ts">
/**
 * 收件箱 Intake — 公开提交渠道管理 + 工单审核/转换。
 * 数据来源：intakeApi（后端 intake 域）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  intakeApi,
  type IntakeChannel,
  type IntakeIssue,
} from "@/api/services/intake";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const tab = ref<"channels" | "issues">("channels");
const loading = ref(true);
const error = ref("");
const channels = ref<IntakeChannel[]>([]);
const issues = ref<IntakeIssue[]>([]);
const showForm = ref(false);
const saving = ref(false);

const form = ref({ slug: "", name: "", description: "" });

async function loadChannels() {
  loading.value = true;
  error.value = "";
  try {
    const r = await intakeApi.listChannels(workspaceId.value, { limit: 100 });
    channels.value = r.items ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadIssues() {
  loading.value = true;
  error.value = "";
  try {
    const r = await intakeApi.listIssues(workspaceId.value, { limit: 100 });
    issues.value = r.items ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function switchTab(t: "channels" | "issues") {
  tab.value = t;
  if (t === "channels") loadChannels();
  else loadIssues();
}

async function createChannel() {
  if (!form.value.slug.trim() || !form.value.name.trim()) {
    toast.warning("请填写 slug 和名称");
    return;
  }
  saving.value = true;
  try {
    await intakeApi.createChannel(workspaceId.value, form.value);
    form.value = { slug: "", name: "", description: "" };
    showForm.value = false;
    await loadChannels();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "创建失败");
  } finally {
    saving.value = false;
  }
}

async function toggleChannel(c: IntakeChannel) {
  try {
    await intakeApi.updateChannel(workspaceId.value, c.id, { is_active: !c.is_active });
    await loadChannels();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function review(issue: IntakeIssue, action: "approve" | "reject") {
  try {
    await intakeApi.reviewIssue(workspaceId.value, issue.id, { action });
    toast.success(action === "approve" ? "已通过审核" : "已拒绝");
    await loadIssues();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function convert(issue: IntakeIssue) {
  const projectId = Number(prompt("输入目标项目 ID："));
  if (!projectId) return;
  try {
    const r = await intakeApi.convertIssue(workspaceId.value, issue.id, {
      target_project_id: projectId,
      target_issue_type: issue.issue_type === "bug" ? "defect" : issue.issue_type,
    });
    toast.success(`已转换为 ${r.identifier}`);
    await loadIssues();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "转换失败");
  }
}

function statusLabel(s: string): string {
  const map: Record<string, string> = {
    pending: "待审核", reviewed: "已审核", converted: "已转换",
    rejected: "已拒绝", archived: "已归档",
  };
  return map[s] ?? s;
}

onMounted(() => loadChannels());
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">收件箱</h1>
      <button
        v-if="tab === 'channels'"
        class="rounded-md bg-[var(--brand-600)] px-3 py-1.5 text-sm font-medium text-white hover:bg-[var(--brand-700)]"
        @click="showForm = !showForm"
      >
        {{ showForm ? "取消" : "新建渠道" }}
      </button>
    </div>

    <!-- Tab 切换 -->
    <div class="flex gap-2">
      <button
        v-for="t in [{ k: 'channels', l: '渠道' }, { k: 'issues', l: '工单' }] as const"
        :key="t.k"
        class="rounded-md px-3 py-1.5 text-sm"
        :class="tab === t.k
          ? 'bg-[var(--brand-600)] text-white'
          : 'border border-[var(--border-subtle)] text-[var(--text-secondary)]'"
        @click="switchTab(t.k)"
      >
        {{ t.l }}
      </button>
    </div>

    <!-- 创建渠道表单 -->
    <div
      v-if="tab === 'channels' && showForm"
      class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4"
    >
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">Slug</label>
          <input v-model="form.slug" type="text" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="product-feedback" />
        </div>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">名称</label>
          <input v-model="form.name" type="text" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="产品反馈" />
        </div>
      </div>
      <div>
        <label class="text-xs text-[var(--text-tertiary)]">描述</label>
        <input v-model="form.description" type="text" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" />
      </div>
      <button
        :disabled="saving"
        class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
        @click="createChannel"
      >
        {{ saving ? "创建中…" : "创建" }}
      </button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 3" :key="i" class="h-14 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="tab === 'channels' ? loadChannels() : loadIssues()" />

    <!-- 渠道列表 -->
    <template v-else-if="tab === 'channels'">
      <AppEmptyState
        v-if="channels.length === 0"
        title="暂无收件箱渠道"
        description="创建公开渠道，让外部用户提交需求与缺陷。"
      />
      <div v-else class="space-y-3">
        <div
          v-for="c in channels"
          :key="c.id"
          class="flex items-center justify-between rounded-md border border-[var(--border-subtle)] p-4"
        >
          <div class="flex items-center gap-3">
            <div>
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-[var(--text-primary)]">{{ c.name }}</span>
                <span
                  class="rounded-full px-2 py-0.5 text-xs"
                  :class="c.is_active ? 'bg-[var(--brand-50)] text-[var(--brand-600)]' : 'bg-[var(--surface-2)] text-[var(--text-tertiary)]'"
                >
                  {{ c.is_active ? "启用" : "停用" }}
                </span>
              </div>
              <div class="mt-0.5 font-mono text-xs text-[var(--text-tertiary)]">
                /intake/{{ c.slug }}
              </div>
              <div class="mt-0.5 text-xs text-[var(--text-tertiary)]">{{ c.description }}</div>
            </div>
          </div>
          <button class="text-xs text-[var(--brand-600)]" @click="toggleChannel(c)">
            {{ c.is_active ? "停用" : "启用" }}
          </button>
        </div>
      </div>
    </template>

    <!-- 工单列表 -->
    <template v-else>
      <AppEmptyState
        v-if="issues.length === 0"
        title="暂无工单"
        description="外部用户通过公开渠道提交的工单会显示在这里。"
      />
      <div v-else class="space-y-3">
        <div
          v-for="i in issues"
          :key="i.id"
          class="rounded-md border border-[var(--border-subtle)] p-4"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="font-mono text-xs text-[var(--brand-600)]">{{ i.tracking_id }}</span>
              <span class="text-sm font-medium text-[var(--text-primary)]">{{ i.title }}</span>
              <span class="rounded-full bg-[var(--surface-2)] px-2 py-0.5 text-xs text-[var(--text-secondary)]">
                {{ statusLabel(i.status) }}
              </span>
            </div>
            <div v-if="i.status === 'pending'" class="flex gap-2">
              <button class="text-xs text-[var(--brand-600)]" @click="review(i, 'approve')">通过</button>
              <button class="text-xs text-[var(--danger, #ef4444)]" @click="review(i, 'reject')">拒绝</button>
            </div>
            <button
              v-else-if="i.status === 'reviewed'"
              class="text-xs text-[var(--brand-600)]"
              @click="convert(i)"
            >
              转换为工作项
            </button>
          </div>
          <div class="mt-1 text-xs text-[var(--text-tertiary)]">
            {{ i.submitter_name }}（{{ i.submitter_email }}）· {{ i.issue_type }}
          </div>
          <div v-if="i.description" class="mt-1 text-sm text-[var(--text-secondary)] line-clamp-2">
            {{ i.description }}
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
