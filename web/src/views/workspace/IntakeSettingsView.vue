<script setup lang="ts">
/**
 * 收件箱设置 — 提报渠道管理与工单审核（对标 Jira Service Management / TAPD 提报）。
 * 能力：渠道 CRUD + 公开链接复制 + 工单列表 + 接受/拒绝/归档 + 转正为正式工作项。
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

const loading = ref(true);
const error = ref("");

// ---- 渠道 ----
const channels = ref<IntakeChannel[]>([]);
const showChannelForm = ref(false);
const saving = ref(false);
const channelForm = ref({ name: "", description: "", slug: "", project_id: undefined as number | undefined });

// ---- 工单 ----
const issues = ref<IntakeIssue[]>([]);
const statusFilter = ref("");
const detail = ref<IntakeIssue | null>(null);
const showDetail = ref(false);
const promoteForm = ref({ type_code: "requirement", severity: 3, found_phase: "内测" });
const showPromote = ref(false);

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    const [ch, iss] = await Promise.all([
      intakeApi.listChannels(workspaceId.value),
      intakeApi.listIssues(workspaceId.value, { limit: 200 }),
    ]);
    channels.value = ch.results ?? [];
    issues.value = iss.results ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadIssues() {
  try {
    const r = await intakeApi.listIssues(workspaceId.value, {
      status: statusFilter.value || undefined,
      limit: 200,
    });
    issues.value = r.results ?? [];
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "加载工单失败");
  }
}

async function createChannel() {
  if (!channelForm.value.name.trim()) { toast.warning("请输入渠道名称"); return; }
  saving.value = true;
  try {
    await intakeApi.createChannel(workspaceId.value, channelForm.value);
    channelForm.value = { name: "", description: "", slug: "", project_id: undefined };
    showChannelForm.value = false;
    toast.success("渠道已创建");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "创建失败");
  } finally {
    saving.value = false;
  }
}

async function toggleChannel(ch: IntakeChannel) {
  try {
    await intakeApi.updateChannel(workspaceId.value, ch.id, { is_active: !ch.is_active });
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function removeChannel(ch: IntakeChannel) {
  if (!confirm(`确定删除渠道「${ch.name}」？相关工单将保留。`)) return;
  try {
    await intakeApi.removeChannel(workspaceId.value, ch.id);
    toast.success("渠道已删除");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "删除失败");
  }
}

function copyLink(ch: IntakeChannel) {
  const base = `${location.origin}/intake/${workspaceId.value}/${ch.slug}`;
  void navigator.clipboard?.writeText(base).then(
    () => toast.success("公开提报链接已复制"),
    () => toast.error("复制失败，请手动复制"),
  );
}

async function openDetail(it: IntakeIssue) {
  detail.value = it;
  showDetail.value = true;
  showPromote.value = false;
}

async function flow(it: IntakeIssue, action: "accept" | "reject" | "archive") {
  try {
    if (action === "accept") await intakeApi.acceptIssue(workspaceId.value, it.id);
    else if (action === "reject") await intakeApi.rejectIssue(workspaceId.value, it.id);
    else await intakeApi.archiveIssue(workspaceId.value, it.id);
    toast.success("操作成功");
    showDetail.value = false;
    await loadIssues();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function promote() {
  if (!detail.value) return;
  try {
    const updated = await intakeApi.promoteIssue(workspaceId.value, detail.value.id, {
      type_code: promoteForm.value.type_code,
      severity: promoteForm.value.type_code === "defect" ? promoteForm.value.severity : undefined,
      found_phase: promoteForm.value.type_code === "defect" ? promoteForm.value.found_phase : undefined,
    });
    toast.success(`已转正：${updated.linked_entity_identifier ?? "工作项"}`);
    showDetail.value = false;
    await loadIssues();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "转正失败");
  }
}

const STATUS_TEXT: Record<string, string> = {
  open: "待处理", accepted: "已接受", rejected: "已拒绝", archived: "已归档",
};
const PRIORITY_TEXT: Record<string, string> = {
  urgent: "紧急", high: "高", medium: "中", low: "低", none: "无",
};

onMounted(load);
</script>

<template>
  <div class="mx-auto max-w-6xl space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">收件箱</h1>
      <button
        class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)]"
        @click="showChannelForm = !showChannelForm"
      >
        {{ showChannelForm ? "取消" : "新建渠道" }}
      </button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 4" :key="i" class="h-14 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 创建渠道表单 -->
      <section v-if="showChannelForm" class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
        <h2 class="text-sm font-semibold text-[var(--text-secondary)]">新建提报渠道</h2>
        <div class="grid gap-3 md:grid-cols-2">
          <div>
            <label class="text-xs text-[var(--text-tertiary)]">名称（必填）</label>
            <input v-model="channelForm.name" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="如：产品需求收集箱" />
          </div>
          <div>
            <label class="text-xs text-[var(--text-tertiary)]">链接标识（留空自动生成）</label>
            <input v-model="channelForm.slug" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="product-intake" />
          </div>
          <div class="md:col-span-2">
            <label class="text-xs text-[var(--text-tertiary)]">描述</label>
            <textarea v-model="channelForm.description" rows="2" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="说明该渠道的用途与提交规范"></textarea>
          </div>
        </div>
        <button
          :disabled="saving"
          class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
          @click="createChannel"
        >
          {{ saving ? "创建中…" : "创建" }}
        </button>
      </section>

      <!-- 渠道列表 -->
      <section class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
        <h2 class="text-sm font-semibold text-[var(--text-secondary)]">提报渠道（{{ channels.length }}）</h2>
        <AppEmptyState v-if="!channels.length" title="还没有提报渠道" desc="创建渠道后，把公开链接分享给提交者即可匿名提报" />
        <div v-else class="space-y-2">
          <div
            v-for="ch in channels"
            :key="ch.id"
            class="flex items-center justify-between rounded-md border border-[var(--border-subtle)] px-3 py-2.5"
          >
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium">{{ ch.name }}</span>
                <span class="rounded bg-[var(--bg-secondary)] px-1.5 py-0.5 text-xs text-[var(--text-tertiary)]">{{ ch.slug }}</span>
                <span
                  class="rounded px-1.5 py-0.5 text-xs"
                  :class="ch.is_active ? 'bg-[var(--success-bg,#e6f7ee)] text-[var(--success,#16a34a)]' : 'bg-[var(--bg-secondary)] text-[var(--text-tertiary)]'"
                >{{ ch.is_active ? "启用中" : "已停用" }}</span>
              </div>
              <p class="mt-0.5 truncate text-xs text-[var(--text-tertiary)]">
                {{ ch.description || "无描述" }} · 工单 {{ ch.issue_count ?? 0 }}
              </p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <button class="text-xs text-[var(--text-secondary)] hover:text-[var(--brand-600)]" title="复制公开链接" @click="copyLink(ch)">复制链接</button>
              <button class="text-xs text-[var(--text-secondary)] hover:text-[var(--brand-600)]" title="启停" @click="toggleChannel(ch)">{{ ch.is_active ? "停用" : "启用" }}</button>
              <button class="text-xs text-[var(--danger,#ef4444)] hover:text-[var(--danger)]" @click="removeChannel(ch)">删除</button>
            </div>
          </div>
        </div>
      </section>

      <!-- 工单列表 -->
      <section class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold text-[var(--text-secondary)]">提报工单（{{ issues.length }}）</h2>
          <select v-model="statusFilter" class="rounded-md border border-[var(--border-subtle)] px-2 py-1.5 text-sm" @change="loadIssues">
            <option value="">全部状态</option>
            <option value="open">待处理</option>
            <option value="accepted">已接受</option>
            <option value="rejected">已拒绝</option>
            <option value="archived">已归档</option>
          </select>
        </div>

        <AppEmptyState v-if="!issues.length" title="暂无工单" desc="把公开提报链接分享出去，提交后会自动出现在这里" />
        <div v-else class="space-y-2">
          <div
            v-for="it in issues"
            :key="it.id"
            class="flex cursor-pointer items-center justify-between rounded-md border border-[var(--border-subtle)] px-3 py-2.5 transition-colors hover:bg-[var(--bg-secondary)]"
            @click="openDetail(it)"
          >
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium">{{ it.name }}</span>
                <span class="font-mono text-xs text-[var(--text-tertiary)]">{{ it.tracking_id }}</span>
                <span
                  class="rounded px-1.5 py-0.5 text-xs"
                  :class="{
                    'bg-[var(--warning-bg,#fef3c7)] text-[var(--warning,#d97706)]': it.status === 'open',
                    'bg-[var(--success-bg,#e6f7ee)] text-[var(--success,#16a34a)]': it.status === 'accepted',
                    'bg-[var(--bg-secondary)] text-[var(--text-tertiary)]': it.status === 'rejected' || it.status === 'archived',
                  }"
                >{{ STATUS_TEXT[it.status] }}</span>
              </div>
              <p class="mt-0.5 truncate text-xs text-[var(--text-tertiary)]">
                {{ it.channel_name || "未知渠道" }} · {{ it.submitter_email }} · 优先级 {{ PRIORITY_TEXT[it.priority] ?? it.priority }}
              </p>
            </div>
            <span v-if="it.linked_entity_identifier" class="shrink-0 text-xs text-[var(--brand-600)]">{{ it.linked_entity_identifier }}</span>
          </div>
        </div>
      </section>
    </template>

    <!-- 工单详情抽屉 -->
    <div v-if="showDetail && detail" class="fixed inset-0 z-50 flex justify-end bg-black/30" @click.self="showDetail = false">
      <div class="h-full w-full max-w-md space-y-4 overflow-y-auto bg-[var(--color-background-primary,#fff)] p-6 shadow-xl">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="text-lg font-semibold">{{ detail.name }}</h3>
            <p class="font-mono text-xs text-[var(--text-tertiary)]">{{ detail.tracking_id }} · {{ STATUS_TEXT[detail.status] }}</p>
          </div>
          <button class="text-xl leading-none text-[var(--text-tertiary)]" @click="showDetail = false">×</button>
        </div>

        <dl class="space-y-2 text-sm">
          <div class="flex justify-between"><dt class="text-[var(--text-tertiary)]">渠道</dt><dd>{{ detail.channel_name || "—" }}</dd></div>
          <div class="flex justify-between"><dt class="text-[var(--text-tertiary)]">提交人</dt><dd>{{ detail.submitter_name || "匿名" }}（{{ detail.submitter_email }}）</dd></div>
          <div class="flex justify-between"><dt class="text-[var(--text-tertiary)]">优先级</dt><dd>{{ PRIORITY_TEXT[detail.priority] ?? detail.priority }}</dd></div>
          <div class="flex justify-between"><dt class="text-[var(--text-tertiary)]">提交时间</dt><dd>{{ new Date(detail.created_at).toLocaleString() }}</dd></div>
          <div v-if="detail.linked_entity_identifier" class="flex justify-between"><dt class="text-[var(--text-tertiary)]">转正后</dt><dd class="text-[var(--brand-600)]">{{ detail.linked_entity_identifier }}</dd></div>
        </dl>

        <div>
          <h4 class="mb-1 text-xs font-semibold text-[var(--text-tertiary)]">描述</h4>
          <pre class="whitespace-pre-wrap rounded-md bg-[var(--bg-secondary)] p-3 text-sm">{{ detail.description || "（无）" }}</pre>
        </div>

        <!-- 转正表单 -->
        <div v-if="showPromote" class="space-y-3 rounded-md border border-[var(--border-subtle)] p-3">
          <h4 class="text-sm font-semibold">转正为正式工作项</h4>
          <select v-model="promoteForm.type_code" class="w-full rounded-md border border-[var(--border-subtle)] px-2 py-2 text-sm">
            <option value="requirement">需求</option>
            <option value="task">任务</option>
            <option value="defect">缺陷</option>
          </select>
          <template v-if="promoteForm.type_code === 'defect'">
            <select v-model.number="promoteForm.severity" class="w-full rounded-md border border-[var(--border-subtle)] px-2 py-2 text-sm">
              <option :value="1">1-致命</option>
              <option :value="2">2-严重</option>
              <option :value="3">3-一般</option>
              <option :value="4">4-提示</option>
              <option :value="5">5-建议</option>
            </select>
            <input v-model="promoteForm.found_phase" class="w-full rounded-md border border-[var(--border-subtle)] px-2 py-2 text-sm" placeholder="发现阶段，如：内测" />
          </template>
          <div class="flex gap-2">
            <button class="flex-1 rounded-md bg-[var(--brand-600)] px-3 py-2 text-sm text-white hover:bg-[var(--brand-700)]" @click="promote">确认转正</button>
            <button class="rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" @click="showPromote = false">取消</button>
          </div>
        </div>

        <div class="flex flex-wrap gap-2 pt-2">
          <template v-if="detail.status === 'open'">
            <button class="rounded-md bg-[var(--success,#16a34a)] px-3 py-2 text-sm text-white" @click="flow(detail, 'accept')">接受</button>
            <button class="rounded-md bg-[var(--danger,#ef4444)] px-3 py-2 text-sm text-white" @click="flow(detail, 'reject')">拒绝</button>
            <button class="rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" @click="flow(detail, 'archive')">归档</button>
          </template>
          <button
            v-if="detail.status === 'accepted' && !detail.linked_entity_id"
            class="rounded-md bg-[var(--brand-600)] px-3 py-2 text-sm text-white"
            @click="showPromote = !showPromote"
          >{{ showPromote ? "取消转正" : "转正" }}</button>
          <button v-if="detail.status === 'accepted' || detail.status === 'rejected'" class="rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" @click="flow(detail, 'archive')">归档</button>
        </div>
      </div>
    </div>
  </div>
</template>
