<script setup lang="ts">
/**
 * 需求/任务/缺陷详情页 — 展示描述、状态流转、活动日志与工时记录。
 */

import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { issueApi, type Issue, type IssueActivity, type State, type TimeLog } from "@/api/services/issue";
import { workspaceApi, type Workspace, type Member } from "@/api/services/workspace";
import { versionApi, type Version } from "@/api/services/version";
import { attachmentApi, type Attachment } from "@/api/services/attachment";
import { toast } from "@/lib/toast";
import { useAuthStore } from "@/stores/auth";
import { useWorkspaceStore } from "@/stores/workspace";
import RichTextEditor from "@/components/RichTextEditor.vue";
import CommentList from "@/components/CommentList.vue";
import AttachmentUploader from "@/components/AttachmentUploader.vue";
import RelationPanel from "./RelationPanel.vue";
import ReviewPanel from "@/components/ReviewPanel.vue";
import IssueCreateModal from "./IssueCreateModal.vue";
import { AppLoadingState, AppErrorState, AppEmptyState, IssueSocialBar } from "@/components";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
}>();

const router = useRouter();

const auth = useAuthStore();
const wsStore = useWorkspaceStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

/** 是否允许编辑/删除：owner/admin 或分配给自己的需求/任务/缺陷且拥有 edit_own 权限 */
const canEditIssue = computed(() => {
  if (!issue.value) return false;
  if (wsStore.hasPermission("issue:edit_all")) return true;
  if (wsStore.hasPermission("issue:edit_own") && issue.value.assignees.includes(currentUserId.value)) {
    return true;
  }
  return false;
});

const canTransition = computed(() => wsStore.hasPermission("issue:transition"));

const ws = ref<Workspace | null>(null);
const issue = ref<Issue | null>(null);
const states = ref<State[]>([]);
const activities = ref<IssueActivity[]>([]);
const loading = ref(true);
const error = ref("");
const transitionError = ref("");
const showTransitionMenu = ref(false);

// --- 工时 ---
const timeLogs = ref<TimeLog[]>([]);
const totalMinutes = ref(0);
const showTimeLogForm = ref(false);
const newSpentDate = ref(new Date().toISOString().slice(0, 10));
const newDurationHours = ref(1);
const newDurationMinutes = ref(0);
const newTimeDesc = ref("");
const timeLogError = ref("");
const timeLogSubmitting = ref(false);
const timeLogsLoading = ref(false);

// --- 版本 / 成员（缺陷行内编辑用） ---
const versions = ref<Version[]>([]);
const versionsLoading = ref(false);
const members = ref<Member[]>([]);
const membersLoading = ref(false);

// --- 行内编辑 ---
const editField = ref<string | null>(null);
const editValue = ref("");
const editSaving = ref(false);
const editError = ref("");

// 描述编辑器（TipTap）状态
const editingDesc = ref(false);
const descHtml = ref("");
const descJsonValue = ref("{}");
const descSaving = ref(false);
const descError = ref("");
const descEditor = ref<InstanceType<typeof RichTextEditor> | null>(null);

/** 粘贴图片到描述编辑器：上传附件并将图片插入编辑器光标位置。 */
async function handleDescPasteImage(file: File) {
  if (!ws.value) return;
  // 仅处理编辑器激活状态下的粘贴
  if (!editingDesc.value) return;
  try {
    // 1. 获取预签名 URL
    const presigned = await attachmentApi.getPresignedUploadURL(
      ws.value.id, props.projectId,
      {
        file_name: file.name || `paste-${Date.now()}.png`,
        content_type: file.type || "image/png",
        entity_type: "issue",
        entity_id: props.issueId,
      },
    );

    // 2. PUT 直传 MinIO
    const ok = await new Promise<boolean>((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open("PUT", presigned.upload_url);
      xhr.setRequestHeader("Content-Type", file.type || "image/png");
      xhr.onload = () => (xhr.status >= 200 && xhr.status < 300 ? resolve(true) : reject(new Error(`upload ${xhr.status}`)));
      xhr.onerror = () => reject(new Error("network error"));
      xhr.send(file);
    });
    if (!ok) throw new Error("上传失败");

    // 3. 确认写入 DB
    const confirmed = await attachmentApi.confirmUpload(ws.value.id, props.projectId, {
      file_name: file.name || `paste-${Date.now()}.png`,
      content_type: file.type || "image/png",
      file_size: file.size,
      entity_type: "issue",
      entity_id: props.issueId,
      storage_key: presigned.storage_key,
    });
    const _att: Attachment = confirmed.attachment;

    // 4. 获取预签名查看 URL（列表接口返回带 URL 的附件）
    const list = await attachmentApi.listAttachments(ws.value.id, props.projectId, "issue", props.issueId);
    const withUrl = list.results.find((a) => a.id === _att.id);
    const imageUrl = withUrl?.storage_url ?? "";
    if (!imageUrl) throw new Error("图片 URL 未就绪");

    // 5. 插入图片到编辑器光标处
    const ed = descEditor.value?.editor;
    if (ed) {
      ed.chain().focus().setImage({ src: imageUrl }).run();
    }
    toast.success("图片已插入");
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : "图片上传失败";
    descError.value = msg;
    toast.error(msg);
  }
}

// 一键提缺陷
const showDefectModal = ref(false);

// --- 子需求/任务/缺陷（WBS 树） ---
const subIssues = ref<Issue[]>([]);
const subIssuesLoading = ref(false);
const showSubIssueModal = ref(false);
const subIssueParentId = ref(0);
const expandedSubIssues = ref<Set<number>>(new Set());
const childrenMap = ref<Record<number, Issue[]>>({});
const childrenLoadingSet = ref<Set<number>>(new Set());

async function loadSubIssues() {
  if (!ws.value) return;
  subIssuesLoading.value = true;
  try {
    const res = await issueApi.listIssues(ws.value.id, props.projectId, {
      parent_id: props.issueId,
    });
    subIssues.value = res.results;
    // 初始展开第一层，孙级懒加载
    expandedSubIssues.value = new Set(subIssues.value.map((s) => s.id));
  } catch {
    // 非关键模块静默忽略
  } finally {
    subIssuesLoading.value = false;
  }
}

/** 懒加载某节点下的子需求/任务/缺陷，完成后自动展开该节点 */
async function loadChildrenOf(id: number) {
  if (!ws.value) return;
  childrenLoadingSet.value.add(id);
  try {
    const res = await issueApi.listIssues(ws.value.id, props.projectId, { parent_id: id });
    childrenMap.value = { ...childrenMap.value, [id]: res.results };
  } catch {
    childrenMap.value = { ...childrenMap.value, [id]: [] };
  } finally {
    childrenLoadingSet.value.delete(id);
    // 加载完成后把该节点标记为已展开
    const next = new Set(expandedSubIssues.value);
    next.add(id);
    expandedSubIssues.value = next;
  }
}

async function loadVersions() {
  if (!ws.value || versionsLoading.value) return;
  versionsLoading.value = true;
  try {
    const res = await versionApi.listVersions(ws.value.id, props.projectId);
    versions.value = res.results ?? [];
  } catch {
    // 非关键模块静默失败
  } finally {
    versionsLoading.value = false;
  }
}

async function loadMembers() {
  if (!ws.value || membersLoading.value) return;
  membersLoading.value = true;
  try {
    members.value = await workspaceApi.listMembers(ws.value.id);
  } catch {
    // 非关键模块静默失败
  } finally {
    membersLoading.value = false;
  }
}

function toggleSubIssue(id: number) {
  const next = new Set(expandedSubIssues.value);
  if (next.has(id)) {
    next.delete(id);
    expandedSubIssues.value = next;
    return;
  }
  // 第一次展开该节点？若尚未加载孙级，先懒加载再展开
  if (childrenMap.value[id] === undefined) {
    void loadChildrenOf(id);
    return;
  }
  next.add(id);
  expandedSubIssues.value = next;
}

function openSubIssue(parentId: number) {
  subIssueParentId.value = parentId;
  showSubIssueModal.value = true;
}

function formatSubIssueState(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.name ?? "未知";
}

function formatSubIssueColor(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.color ?? "#8DA2C2";
}

function navigateToIssue(issueId: number) {
  router.push(`/${props.workspaceId}/projects/${props.projectId}/issues/${issueId}`);
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    ws.value = await workspaceApi.get(props.workspaceId);
    const [iss, st, acts] = await Promise.all([
      issueApi.getIssue(ws.value.id, props.projectId, props.issueId),
      issueApi.listStates(ws.value.id, props.projectId),
      issueApi.listActivities(ws.value.id, props.projectId, props.issueId),
    ]);
    issue.value = iss;
    states.value = st;
    activities.value = acts.results;
    // 缺陷类型才需要版本 / 成员下拉数据
    if (iss.type_code === "defect") {
      loadVersions();
      loadMembers();
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function doTransition(toStateId: number) {
  if (!ws.value || !issue.value) return;
  transitionError.value = "";
  showTransitionMenu.value = false;
  try {
    issue.value = await issueApi.transition(ws.value.id, props.projectId, props.issueId, toStateId);
    toast.success("状态已流转");
  } catch (e: unknown) {
    transitionError.value = e instanceof Error ? e.message : "流转失败";
    toast.error(transitionError.value);
  }
}

async function doDelete() {
  if (!ws.value || !confirm("确定要归档该需求/任务/缺陷吗？")) return;
  try {
    await issueApi.deleteIssue(ws.value.id, props.projectId, props.issueId);
    toast.success("需求/任务/缺陷已归档");
    router.push(`/${props.workspaceId}/projects/${props.projectId}/board`);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "删除失败";
    toast.error(error.value);
  }
}

function goBack() {
  router.push(`/${props.workspaceId}/projects/${props.projectId}/board`);
}

function stateName(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.name ?? "未知";
}

function stateColor(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.color ?? "#8DA2C2";
}

function typeLabel(type: string): string {
  return ({ epic: "史诗", requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[type] ?? type;
}

/** 延期原因标签映射 */
const delayReasonLabels: Record<string, string> = {
  scope_change: "需求范围变更",
  resource_lack: "资源不足",
  tech_blocker: "技术阻塞",
  dependency: "依赖延期",
  estimation: "估算不准确",
  priority_shift: "优先级调整",
  external: "外部因素",
  other: "其他",
};

function delayReasonLabel(reason?: string | null): string {
  if (!reason) return "—";
  return delayReasonLabels[reason] ?? reason;
}

/** 任务是否延期（目标日期已过且未完成） */
const isOverdue = computed(() => {
  if (!issue.value) return false;
  if (!issue.value.target_date) return false;
  const state = states.value.find((s) => s.id === issue.value!.state_id);
  if (state && (state.group === "completed" || state.group === "cancelled")) return false;
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const target = new Date(issue.value.target_date);
  return target < today;
});

const availableTransitions = computed(() => {
  if (!issue.value) return [];
  return states.value.filter((s) => s.id !== issue.value!.state_id);
});

async function loadTimeLogs() {
  if (!ws.value) return;
  timeLogsLoading.value = true;
  try {
    const res = await issueApi.listTimeLogs(ws.value.id, props.projectId, props.issueId);
    timeLogs.value = res.results;
    totalMinutes.value = res.results.reduce((sum, tl) => sum + tl.duration_minutes, 0);
  } catch {
    // 非关键模块，静默失败
  } finally {
    timeLogsLoading.value = false;
  }
}

async function submitTimeLog() {
  if (!ws.value || timeLogSubmitting.value) return;
  const totalMins = newDurationHours.value * 60 + newDurationMinutes.value;
  if (totalMins <= 0 || totalMins > 1440) {
    timeLogError.value = "请填写有效的工时（1分钟-24小时）";
    return;
  }
  timeLogSubmitting.value = true;
  timeLogError.value = "";
  try {
    await issueApi.createTimeLog(ws.value.id, props.projectId, props.issueId, {
      spent_date: newSpentDate.value,
      duration_minutes: totalMins,
      description: newTimeDesc.value.trim() || undefined,
    });
    showTimeLogForm.value = false;
    newTimeDesc.value = "";
    newDurationHours.value = 1;
    newDurationMinutes.value = 0;
    toast.success("工时已记录");
    await loadTimeLogs();
  } catch (e: unknown) {
    timeLogError.value = e instanceof Error ? e.message : "记录失败";
    toast.error(timeLogError.value);
  } finally {
    timeLogSubmitting.value = false;
  }
}

async function deleteTimeLog(logId: number) {
  if (!ws.value) return;
  if (!window.confirm("确定删除该条工时记录吗？")) return;
  try {
    await issueApi.deleteTimeLog(ws.value.id, props.projectId, props.issueId, logId);
    toast.success("工时记录已删除");
    await loadTimeLogs();
  } catch (e: unknown) {
    timeLogError.value = e instanceof Error ? e.message : "删除失败";
    toast.error(timeLogError.value);
  }
}

function fmtDuration(mins: number): string {
  if (mins < 60) return `${mins}分钟`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m > 0 ? `${h}小时${m}分钟` : `${h}小时`;
}

/** 格式化小时数（用于 actual_effort / remaining_effort 等以小时存储的字段）。 */
function fmtDurationHours(hours: number): string {
  if (hours < 0.01) return "0 分钟";
  const totalMins = Math.round(hours * 60);
  return fmtDuration(totalMins);
}

// --- 描述编辑器 ---
function startEditDesc() {
  if (!issue.value) return;
  editingDesc.value = true;
  descHtml.value = issue.value.description_html || "";
  descJsonValue.value = "{}";
  descError.value = "";
}

async function saveDesc() {
  if (!ws.value || !issue.value || descSaving.value) return;
  descSaving.value = true;
  descError.value = "";
  try {
    issue.value = await issueApi.updateIssue(
      ws.value.id, props.projectId, props.issueId,
      { description_html: descHtml.value, version: issue.value.version } as Parameters<typeof issueApi.updateIssue>[3],
    );
    editingDesc.value = false;
  } catch (e: unknown) {
    descError.value = e instanceof Error ? e.message : "保存失败";
  } finally {
    descSaving.value = false;
  }
}

function cancelEditDesc() {
  editingDesc.value = false;
  descHtml.value = "";
  descError.value = "";
}

// --- 行内编辑 ---
function startEdit(field: string, currentValue: unknown) {
  editField.value = field;
  editValue.value = String(currentValue ?? "");
  editError.value = "";
}

function cancelEdit() {
  editField.value = null;
  editValue.value = "";
  editError.value = "";
}

async function saveEdit() {
  if (!ws.value || !issue.value || !editField.value || editSaving.value) return;
  editSaving.value = true;
  editError.value = "";

  try {
    const input: Record<string, unknown> = { version: issue.value.version };
    const field = editField.value;

    switch (field) {
      case "name":
        input.name = editValue.value.trim();
        break;
      case "description_html":
        input.description_html = editValue.value;
        break;
      case "priority":
        input.priority = editValue.value;
        break;
      case "delay_reason":
        input.delay_reason = editValue.value || undefined;
        break;
      case "severity":
        input.severity = Number(editValue.value);
        break;
      case "found_phase":
        input.found_phase = editValue.value;
        break;
      case "root_cause_category":
        input.root_cause_category = editValue.value;
        break;
      case "point":
        input.point = editValue.value ? Number(editValue.value) : null;
        break;
      case "fix_version_id":
        input.fix_version_id = editValue.value ? Number(editValue.value) : null;
        break;
      case "found_version_id":
        input.found_version_id = editValue.value ? Number(editValue.value) : null;
        break;
      case "verifier_id":
        input.verifier_id = editValue.value ? Number(editValue.value) : null;
        break;
      case "repr_expect":
        input.reproduce_steps = {
          ...((issue.value?.reproduce_steps as Record<string, unknown>) ?? {}),
          expected: editValue.value,
        };
        break;
      case "repr_actual":
        input.reproduce_steps = {
          ...((issue.value?.reproduce_steps as Record<string, unknown>) ?? {}),
          actual: editValue.value,
        };
        break;
    }

    issue.value = await issueApi.updateIssue(
      ws.value.id, props.projectId, props.issueId,
      input as unknown as Parameters<typeof issueApi.updateIssue>[3],
    );
    cancelEdit();
  } catch (e: unknown) {
    editError.value = e instanceof Error ? e.message : "保存失败";
  } finally {
    editSaving.value = false;
  }
}

onMounted(() => {
  load();
  loadTimeLogs();
  loadSubIssues();
});
</script>

<template>
  <div class="issue-detail">
    <header class="issue-detail__header">
      <button class="btn btn--ghost" @click="goBack">← 返回看板</button>
      <div class="issue-detail__actions">
        <button
          v-if="issue && issue.type_code === 'requirement' && canEditIssue"
          class="btn btn--sm"
          @click="showDefectModal = true"
        >
          🐛 提缺陷
        </button>
        <button v-if="canEditIssue" class="btn btn--danger" @click="doDelete">归档</button>
      </div>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!issue"
      title="需求/任务/缺陷不存在或已被删除"
      description="请检查需求/任务/缺陷 ID 是否正确"
    >
      <button class="btn btn--ghost" @click="goBack">← 返回看板</button>
    </AppEmptyState>

    <div v-else class="issue-detail__body">
      <div class="issue-detail__main">
        <div class="issue-detail__identifier">{{ issue.identifier }}</div>
        <div v-if="editField === 'name'" class="edit-row">
          <input v-model="editValue" class="edit-input" autofocus @keydown.enter="saveEdit" @keydown.escape="cancelEdit" />
          <button class="btn btn--sm btn--primary" :disabled="editSaving || !editValue.trim()" @click="saveEdit">
            {{ editSaving ? "保存中..." : "保存" }}
          </button>
          <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
          <span v-if="editError" class="form-error">{{ editError }}</span>
        </div>
        <h1
          v-if="canEditIssue"
          class="issue-detail__name editable"
          @click="startEdit('name', issue.name)"
        >
          {{ issue.name }}
          <span class="edit-hint">✎</span>
        </h1>
        <h1 v-else class="issue-detail__name">{{ issue.name }}</h1>

        <div class="issue-detail__meta-row">
          <span class="badge" :class="`badge-${issue.type_code}`">{{ typeLabel(issue.type_code) }}</span>
          <span
            class="issue-detail__state-badge"
            :style="{ backgroundColor: stateColor(issue.state_id) }"
          >
            {{ stateName(issue.state_id) }}
          </span>
          <span v-if="editField === 'priority'" class="edit-row edit-row--inline">
            <select v-model="editValue" class="edit-select" @change="saveEdit">
              <option value="urgent">紧急</option>
              <option value="high">高</option>
              <option value="medium">中</option>
              <option value="low">低</option>
              <option value="none">无</option>
            </select>
            <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
          </span>
          <span
            v-else-if="canEditIssue"
            class="issue-detail__priority editable"
            @click="startEdit('priority', issue.priority)"
          >
            优先级: {{ ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[issue.priority] ?? issue.priority }}
            <span class="edit-hint">✎</span>
          </span>
          <span v-else class="issue-detail__priority">
            优先级: {{ ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[issue.priority] ?? issue.priority }}
          </span>
          <span v-if="issue.severity" class="issue-detail__field">严重度: S{{ issue.severity }}</span>
          <span v-if="issue.found_phase" class="issue-detail__field">发现阶段: {{ issue.found_phase }}</span>
          <span v-if="issue.point != null" class="issue-detail__field">点数: {{ issue.point }}</span>
          <span v-if="issue.sprint_id" class="issue-detail__field">
            所属迭代:
            <router-link :to="`/${workspaceId}/projects/${projectId}/sprints/${issue.sprint_id}`" class="link">
              #{{ issue.sprint_id }}
            </router-link>
          </span>
          <!-- 延期原因（任务延期时显示） -->
          <template v-if="issue.type_code === 'task' && (isOverdue || issue.delay_reason)">
            <span v-if="editField === 'delay_reason'" class="edit-row edit-row--inline">
              <select v-model="editValue" class="edit-select" @change="saveEdit">
                <option value="">— 请选择 —</option>
                <option value="scope_change">需求范围变更</option>
                <option value="resource_lack">资源不足</option>
                <option value="tech_blocker">技术阻塞</option>
                <option value="dependency">依赖延期</option>
                <option value="estimation">估算不准确</option>
                <option value="priority_shift">优先级调整</option>
                <option value="external">外部因素</option>
                <option value="other">其他</option>
              </select>
              <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
            </span>
            <span
              v-else-if="canEditIssue"
              class="issue-detail__field issue-detail__delay-reason editable"
              @click="startEdit('delay_reason', issue.delay_reason ?? '')"
            >
              延期原因: <strong>{{ delayReasonLabel(issue.delay_reason) }}</strong>
              <span v-if="isOverdue && !issue.delay_reason" class="delay-reason--required">（必填）</span>
              <span class="edit-hint">✎</span>
            </span>
            <span v-else class="issue-detail__field">
              延期原因: <strong>{{ delayReasonLabel(issue.delay_reason) }}</strong>
            </span>
          </template>
        </div>

        <!-- 关注栏 -->
        <IssueSocialBar
          v-if="ws && issue"
          :workspace-id="ws.id"
          :project-id="projectId"
          :issue-id="issueId"
          :initial-watching="issue.watchers.includes(currentUserId)"
          class="issue-detail__social"
        />

        <div class="issue-detail__section">
          <div class="section-head">
            <h3>描述</h3>
            <div v-if="!editingDesc && canEditIssue" class="section-head__actions">
              <button class="btn btn--sm btn--ghost" @click="startEditDesc">编辑</button>
            </div>
          </div>
          <!-- 编辑模式：TipTap 富文本编辑器 -->
          <div v-if="editingDesc" class="edit-row">
            <RichTextEditor
              ref="descEditor"
              v-model:content-html="descHtml"
              v-model:content-json="descJsonValue"
              placeholder="输入需求/任务/缺陷描述..."
              :min-height="'200px'"
              :workspace-id="ws?.id ?? props.workspaceId"
              :project-id="props.projectId"
              @paste-image="handleDescPasteImage"
            />
            <div class="edit-row__actions">
              <button class="btn btn--sm btn--primary" :disabled="descSaving" @click="saveDesc">{{ descSaving ? "保存中..." : "保存" }}</button>
              <button class="btn btn--sm" :disabled="descSaving" @click="cancelEditDesc">取消</button>
              <span v-if="descError" class="form-error">{{ descError }}</span>
            </div>
          </div>
          <!-- 展示模式：TipTap 只读渲染 -->
          <div v-else-if="issue.description_html" class="issue-detail__desc">
            <RichTextEditor
              :content-html="issue.description_html"
              :editable="false"
            />
          </div>
          <p v-else class="text-muted">暂无描述，点击编辑添加</p>

          <!-- 附件 -->
          <div v-if="ws" class="issue-detail__attachments">
            <AttachmentUploader
              :workspace-id="ws.id"
              :project-id="props.projectId"
              entity-type="issue"
              :entity-id="props.issueId"
            />
          </div>
        </div>

        <div v-if="issue.type_code === 'defect'" class="issue-detail__section">
          <h3>缺陷信息</h3>
          <div class="issue-detail__fields">
            <!-- 复现步骤 -->
            <div v-if="issue.reproduce_steps && issue.reproduce_steps.steps" class="field-row">
              <span class="field-label">复现步骤:</span>
              <span class="field-value">{{ issue.reproduce_steps.steps }}</span>
            </div>

            <!-- 期望结果（行内 textarea） -->
            <div class="field-row field-row--editable">
              <span class="field-label">期望结果:</span>
              <div class="field-value field-value--grow">
                <div
                  v-if="editField === 'repr_expect'"
                  class="edit-row"
                >
                  <textarea
                    v-model="editValue"
                    class="edit-textarea"
                    rows="3"
                    placeholder="描述期望结果..."
                    @keydown.escape="cancelEdit"
                  ></textarea>
                  <div class="edit-row__actions">
                    <button class="btn btn--sm btn--primary" :disabled="editSaving" @click="saveEdit">{{ editSaving ? "保存中..." : "保存" }}</button>
                    <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
                    <span v-if="editError" class="form-error">{{ editError }}</span>
                  </div>
                </div>
                <div
                  v-else-if="canEditIssue"
                  class="editable editable--textarea"
                  @click="startEdit('repr_expect', (issue.reproduce_steps as Record<string, unknown>)?.expected ?? '')"
                >
                  {{ (issue.reproduce_steps as Record<string, unknown>)?.expected || '点击设置...' }}
                  <span class="edit-hint">✎</span>
                </div>
                <span v-else>{{ (issue.reproduce_steps as Record<string, unknown>)?.expected || '—' }}</span>
              </div>
            </div>

            <!-- 实际结果（行内 textarea） -->
            <div class="field-row field-row--editable">
              <span class="field-label">实际结果:</span>
              <div class="field-value field-value--grow">
                <div
                  v-if="editField === 'repr_actual'"
                  class="edit-row"
                >
                  <textarea
                    v-model="editValue"
                    class="edit-textarea"
                    rows="3"
                    placeholder="描述实际结果..."
                    @keydown.escape="cancelEdit"
                  ></textarea>
                  <div class="edit-row__actions">
                    <button class="btn btn--sm btn--primary" :disabled="editSaving" @click="saveEdit">{{ editSaving ? "保存中..." : "保存" }}</button>
                    <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
                    <span v-if="editError" class="form-error">{{ editError }}</span>
                  </div>
                </div>
                <div
                  v-else-if="canEditIssue"
                  class="editable editable--textarea"
                  @click="startEdit('repr_actual', (issue.reproduce_steps as Record<string, unknown>)?.actual ?? '')"
                >
                  {{ (issue.reproduce_steps as Record<string, unknown>)?.actual || '点击设置...' }}
                  <span class="edit-hint">✎</span>
                </div>
                <span v-else>{{ (issue.reproduce_steps as Record<string, unknown>)?.actual || '—' }}</span>
              </div>
            </div>

            <!-- 环境 -->
            <div v-if="issue.environment && issue.environment.value" class="field-row">
              <span class="field-label">环境:</span>
              <span class="field-value">{{ issue.environment.value }}</span>
            </div>

            <!-- 根因分类（行内下拉） -->
            <div class="field-row field-row--editable">
              <span class="field-label">根因分类:</span>
              <div class="field-value field-value--grow">
                <div v-if="editField === 'root_cause_category'" class="edit-row edit-row--inline">
                  <select v-model="editValue" class="edit-select">
                    <option value="code_defect">代码缺陷</option>
                    <option value="design_defect">设计缺陷</option>
                    <option value="environment">环境问题</option>
                    <option value="data">数据问题</option>
                    <option value="requirement">需求理解偏差</option>
                    <option value="test_omission">测试遗漏</option>
                    <option value="documentation">文档缺陷</option>
                    <option value="other">其他</option>
                    <option value="">未分类</option>
                  </select>
                  <button class="btn btn--sm btn--primary" :disabled="editSaving" @click="saveEdit">{{ editSaving ? "保存中..." : "保存" }}</button>
                  <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
                </div>
                <div
                  v-else-if="canEditIssue"
                  class="editable"
                  @click="startEdit('root_cause_category', issue.root_cause_category ?? '')"
                >
                  <span v-if="issue.root_cause_category">
                    {{ ({ code_defect: "代码缺陷", design_defect: "设计缺陷", environment: "环境问题", data: "数据问题", requirement: "需求理解偏差", test_omission: "测试遗漏", documentation: "文档缺陷", other: "其他" } as Record<string, string>)[issue.root_cause_category] ?? issue.root_cause_category }}
                  </span>
                  <span v-else class="add-btn" title="设置根因分类">＋</span>
                  <span class="edit-hint">✎</span>
                </div>
                <span v-else>{{ issue.root_cause_category ? ({ code_defect: "代码缺陷", design_defect: "设计缺陷", environment: "环境问题", data: "数据问题", requirement: "需求理解偏差", test_omission: "测试遗漏", documentation: "文档缺陷", other: "其他" } as Record<string, string>)[issue.root_cause_category] ?? issue.root_cause_category : '—' }}</span>
              </div>
            </div>

            <!-- 发现版本（行内下拉） -->
            <div class="field-row field-row--editable">
              <span class="field-label">发现版本:</span>
              <div class="field-value field-value--grow">
                <div v-if="editField === 'found_version_id'" class="edit-row edit-row--inline">
                  <select v-model="editValue" class="edit-select">
                    <option :value="''">— 未设置 —</option>
                    <option v-for="v in versions" :key="v.id" :value="String(v.id)">
                      {{ v.name }} ({{ v.semver }})
                    </option>
                  </select>
                  <button class="btn btn--sm btn--primary" :disabled="editSaving" @click="saveEdit">{{ editSaving ? "保存中..." : "保存" }}</button>
                  <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
                </div>
                <div
                  v-else-if="canEditIssue"
                  class="editable"
                  @click="startEdit('found_version_id', issue.found_version_id ?? '')"
                >
                  <template v-if="issue.found_version_id">
                    {{ versions.find(v => v.id === issue?.found_version_id)?.name ?? '—' }}
                  </template>
                  <span v-else class="add-btn" title="设置发现版本">＋</span>
                  <span class="edit-hint">✎</span>
                </div>
                <span v-else>{{ issue?.found_version_id ? versions.find(v => v.id === issue?.found_version_id)?.name ?? `#${issue?.found_version_id}` : '—' }}</span>
              </div>
            </div>

            <!-- 修复版本（行内下拉） -->
            <div class="field-row field-row--editable">
              <span class="field-label">修复版本:</span>
              <div class="field-value field-value--grow">
                <div v-if="editField === 'fix_version_id'" class="edit-row edit-row--inline">
                  <select v-model="editValue" class="edit-select">
                    <option :value="''">— 未设置 —</option>
                    <option v-for="v in versions" :key="v.id" :value="String(v.id)">
                      {{ v.name }} ({{ v.semver }})
                    </option>
                  </select>
                  <button class="btn btn--sm btn--primary" :disabled="editSaving" @click="saveEdit">{{ editSaving ? "保存中..." : "保存" }}</button>
                  <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
                </div>
                <div
                  v-else-if="canEditIssue"
                  class="editable"
                  @click="startEdit('fix_version_id', issue.fix_version_id ?? '')"
                >
                  <template v-if="issue.fix_version_id">
                    {{ versions.find(v => v.id === issue?.fix_version_id)?.name ?? '—' }}
                  </template>
                  <span v-else class="add-btn" title="设置修复版本">＋</span>
                  <span class="edit-hint">✎</span>
                </div>
                <span v-else>{{ issue?.fix_version_id ? versions.find(v => v.id === issue?.fix_version_id)?.name ?? `#${issue?.fix_version_id}` : '—' }}</span>
              </div>
            </div>

            <!-- 验证人（行内下拉） -->
            <div class="field-row field-row--editable">
              <span class="field-label">验证人:</span>
              <div class="field-value field-value--grow">
                <div v-if="editField === 'verifier_id'" class="edit-row edit-row--inline">
                  <select v-model="editValue" class="edit-select">
                    <option :value="''">— 未设置 —</option>
                    <option v-for="m in members" :key="m.id" :value="String(m.id)">
                      {{ m.display_name ?? m.email }}
                    </option>
                  </select>
                  <button class="btn btn--sm btn--primary" :disabled="editSaving" @click="saveEdit">{{ editSaving ? "保存中..." : "保存" }}</button>
                  <button class="btn btn--sm" :disabled="editSaving" @click="cancelEdit">取消</button>
                </div>
                <div
                  v-else-if="canEditIssue"
                  class="editable"
                  @click="startEdit('verifier_id', issue.verifier_id ?? '')"
                >
                  <template v-if="issue.verifier_id">
                    {{ members.find(m => m.id === issue?.verifier_id)?.display_name ?? members.find(m => m.id === issue?.verifier_id)?.email ?? `用户 #${issue?.verifier_id}` }}
                  </template>
                  <span v-else class="add-btn" title="设置验证人">＋</span>
                  <span class="edit-hint">✎</span>
                </div>
                <span v-else>{{ issue?.verifier_id ? members.find(m => m.id === issue?.verifier_id)?.display_name ?? `用户 #${issue?.verifier_id}` : '—' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 评论 -->
        <CommentList
          v-if="ws"
          :workspace-id="ws.id"
          :project-id="props.projectId"
          :issue-id="props.issueId"
        />

        <!-- 流转操作（仅拥有 issue:transition 权限的用户可见） -->
        <div v-if="canTransition" class="issue-detail__section">
          <h3>状态流转</h3>
          <div v-if="transitionError" class="form-error">{{ transitionError }}</div>
          <div class="issue-detail__transitions">
            <button
              v-for="st in availableTransitions"
              :key="st.id"
              class="btn btn--state"
              :style="{ borderColor: st.color, color: st.color }"
              @click="doTransition(st.id)"
            >
              → {{ st.name }}
            </button>
          </div>
        </div>
      </div>

      <!-- 侧边栏：活动日志 -->
      <aside class="issue-detail__sidebar">
        <h3>活动日志</h3>
        <div v-if="activities.length === 0" class="text-muted">暂无活动记录</div>
        <div v-else class="activity-timeline">
          <div v-for="act in activities" :key="act.id" class="activity-item">
            <div class="activity-item__icon" :class="`verb-${act.verb}`"></div>
            <div class="activity-item__body">
              <div class="activity-item__text">
                <strong>{{ act.actor_name || "系统" }}</strong>
                {{ act.verb === "created" ? "创建了需求/任务/缺陷" : act.verb === "transitioned" ? `流转状态: ${act.old_value} → ${act.new_value}` : `${act.field}: ${act.old_value} → ${act.new_value}` }}
              </div>
              <div class="activity-item__time">{{ new Date(act.created_at).toLocaleString() }}</div>
            </div>
          </div>
        </div>

        <!-- 工时 -->
        <h3 style="margin-top: 24px">工时</h3>
        <div v-if="totalMinutes > 0" class="timelog-summary">
          累计 {{ fmtDuration(totalMinutes) }}
          <span v-if="issue.actual_effort != null" class="timelog-effort">
            · 实耗 {{ fmtDurationHours(issue.actual_effort) }}
          </span>
          <span v-if="issue.remaining_effort != null" class="timelog-effort">
            · 剩余 {{ fmtDurationHours(issue.remaining_effort) }}
          </span>
        </div>

        <button v-if="!showTimeLogForm" class="btn btn--sm btn--outline" @click="showTimeLogForm = true">
          ＋ 记录工时
        </button>

        <!-- 工时记录表单 -->
        <div v-if="showTimeLogForm" class="timelog-form">
          <div v-if="timeLogError" class="form-error">{{ timeLogError }}</div>
          <div class="timelog-form__row">
            <input
              v-model="newSpentDate"
              type="date"
              class="timelog-input"
              :max="new Date().toISOString().slice(0, 10)"
            />
          </div>
          <div class="timelog-form__row timelog-form__duration">
            <input v-model.number="newDurationHours" type="number" class="timelog-input timelog-input--sm" min="0" max="24" />
            <span class="timelog-label">小时</span>
            <input v-model.number="newDurationMinutes" type="number" class="timelog-input timelog-input--sm" min="0" max="59" step="15" />
            <span class="timelog-label">分钟</span>
          </div>
          <textarea
            v-model="newTimeDesc"
            class="timelog-textarea"
            placeholder="工时描述（可选）"
            rows="2"
          ></textarea>
          <div class="timelog-form__actions">
            <button class="btn btn--sm" :disabled="timeLogSubmitting" @click="showTimeLogForm = false">取消</button>
            <button class="btn btn--sm btn--primary" :disabled="timeLogSubmitting" @click="submitTimeLog">
              {{ timeLogSubmitting ? "保存中..." : "保存" }}
            </button>
          </div>
        </div>

        <!-- 工时列表 -->
        <div v-if="timeLogsLoading" class="text-muted" style="margin-top:8px">加载中...</div>
        <div v-else-if="timeLogs.length > 0" class="timelog-list">
          <div v-for="tl in timeLogs.slice(0, 10)" :key="tl.id" class="timelog-item">
            <div class="timelog-item__meta">
              <span class="timelog-item__date">{{ tl.spent_date.slice(0, 10) }}</span>
              <span class="timelog-item__duration">{{ fmtDuration(tl.duration_minutes) }}</span>
            </div>
            <div v-if="tl.description" class="timelog-item__desc">{{ tl.description }}</div>
            <button class="timelog-item__delete" title="删除工时记录" @click="deleteTimeLog(tl.id)">✕</button>
          </div>
        </div>
        <div v-else-if="!showTimeLogForm" class="text-muted">暂无工时记录</div>

        <!-- 子需求/任务/缺陷 -->
        <div class="sub-issues-section" style="margin-top: 24px">
          <div class="sub-issues-header">
            <h3>子需求/任务/缺陷</h3>
            <button
              v-if="canEditIssue"
              class="btn btn--sm btn--outline"
              @click="openSubIssue(props.issueId)"
            >
              ＋ 添加子需求/任务/缺陷
            </button>
          </div>

          <div v-if="subIssuesLoading" class="text-muted">加载中...</div>
          <div v-else-if="subIssues.length === 0" class="text-muted">暂无子需求/任务/缺陷</div>
          <div v-else class="sub-issues-tree">
            <div v-for="child in subIssues" :key="child.id" class="sub-issue-node">
              <div class="sub-issue-node__row">
                <button
                  class="sub-issue-node__toggle"
                  :class="{ 'sub-issue-node__toggle--collapsed': !expandedSubIssues.has(child.id) }"
                  :aria-expanded="expandedSubIssues.has(child.id)"
                  @click="toggleSubIssue(child.id)"
                >
                  ▶
                </button>
                <span
                  class="sub-issue-node__identifier"
                  @click="navigateToIssue(child.id)"
                >{{ child.identifier }}</span>
                <span class="sub-issue-node__type" :class="`sub-issue-node__type--${child.type_code}`">
                  {{ typeLabel(child.type_code) }}
                </span>
                <span
                  class="sub-issue-node__state"
                  :style="{ backgroundColor: formatSubIssueColor(child.state_id) }"
                >{{ formatSubIssueState(child.state_id) }}</span>
              </div>
              <div class="sub-issue-node__title" @click="navigateToIssue(child.id)">
                {{ child.name }}
              </div>
              <div class="sub-issue-node__actions">
                <button
                  v-if="canEditIssue"
                  class="btn btn--sm btn--ghost"
                  @click="openSubIssue(child.id)"
                >
                  ＋ 子项
                </button>
              </div>

              <!-- 孙级（第二层），懒加载渲染 -->
              <div
                v-if="expandedSubIssues.has(child.id)"
                class="sub-issue-node__children"
              >
                <div v-if="childrenLoadingSet.has(child.id)" class="text-muted sub-issue-node__placeholder">
                  加载中…
                </div>
                <div v-else-if="!childrenMap[child.id]?.length" class="text-muted sub-issue-node__placeholder">
                  无子项
                </div>
                <div
                  v-for="gc in childrenMap[child.id]"
                  :key="gc.id"
                  class="sub-issue-node sub-issue-node--nested"
                >
                  <div class="sub-issue-node__row">
                    <button
                      class="sub-issue-node__toggle"
                      :class="{ 'sub-issue-node__toggle--collapsed': !expandedSubIssues.has(gc.id) }"
                      :aria-expanded="expandedSubIssues.has(gc.id)"
                      @click="toggleSubIssue(gc.id)"
                    >
                      ▶
                    </button>
                    <span
                      class="sub-issue-node__identifier"
                      @click="navigateToIssue(gc.id)"
                    >{{ gc.identifier }}</span>
                    <span class="sub-issue-node__type" :class="`sub-issue-node__type--${gc.type_code}`">
                      {{ typeLabel(gc.type_code) }}
                    </span>
                    <span
                      class="sub-issue-node__state"
                      :style="{ backgroundColor: formatSubIssueColor(gc.state_id) }"
                    >{{ formatSubIssueState(gc.state_id) }}</span>
                  </div>
                  <div class="sub-issue-node__title" @click="navigateToIssue(gc.id)">
                    {{ gc.name }}
                  </div>
                  <div class="sub-issue-node__actions">
                    <button
                      v-if="canEditIssue"
                      class="btn btn--sm btn--ghost"
                      @click="openSubIssue(gc.id)"
                    >
                      ＋ 子项
                    </button>
                  </div>

                  <!-- 曾孙级（第三层，叶子节点不再展开） -->
                  <div
                    v-if="expandedSubIssues.has(gc.id)"
                    class="sub-issue-node__children"
                  >
                    <div v-if="childrenLoadingSet.has(gc.id)" class="text-muted sub-issue-node__placeholder">
                      加载中…
                    </div>
                    <div v-else-if="!childrenMap[gc.id]?.length" class="text-muted sub-issue-node__placeholder">
                      无子项
                    </div>
                    <div
                      v-for="ggc in childrenMap[gc.id]"
                      :key="ggc.id"
                      class="sub-issue-node sub-issue-node--nested"
                    >
                      <div class="sub-issue-node__row">
                        <span class="sub-issue-node__toggle sub-issue-node__toggle--leaf">•</span>
                        <span
                          class="sub-issue-node__identifier"
                          @click="navigateToIssue(ggc.id)"
                        >{{ ggc.identifier }}</span>
                        <span class="sub-issue-node__type" :class="`sub-issue-node__type--${ggc.type_code}`">
                          {{ typeLabel(ggc.type_code) }}
                        </span>
                        <span
                          class="sub-issue-node__state"
                          :style="{ backgroundColor: formatSubIssueColor(ggc.state_id) }"
                        >{{ formatSubIssueState(ggc.state_id) }}</span>
                      </div>
                      <div class="sub-issue-node__title" @click="navigateToIssue(ggc.id)">
                        {{ ggc.name }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 关联关系 -->
        <RelationPanel
          v-if="ws"
          :workspace-id="ws.id"
          :project-id="props.projectId"
          :issue-id="props.issueId"
          style="margin-top: 24px"
        />
        <!-- 需求评审（仅需求类型） -->
        <ReviewPanel
          v-if="ws && issue"
          :workspace-id="ws.id"
          :project-id="props.projectId"
          :issue-id="props.issueId"
          :issue-type="issue.type_code"
          :review-status="issue.review_status"
          style="margin-top: 24px"
        />
      </aside>
    </div>

    <IssueCreateModal
      v-if="ws && issue"
      :workspace-id="ws.id"
      :project-id="props.projectId"
      :visible="showDefectModal"
      preset-type="defect"
      @close="showDefectModal = false"
      @created="showDefectModal = false"
    />

    <!-- 子需求/任务/缺陷创建弹窗 -->
    <IssueCreateModal
      v-if="ws && showSubIssueModal"
      :workspace-id="ws.id"
      :project-id="props.projectId"
      :visible="showSubIssueModal"
      :parent-id="subIssueParentId"
      @close="showSubIssueModal = false"
      @created="showSubIssueModal = false; loadSubIssues()"
    />
  </div>
</template>

<style scoped>
.issue-detail__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.issue-detail__body {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 32px;
}

.issue-detail__identifier {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
  font-weight: 600;
}

.issue-detail__name {
  font-size: 22px;
  font-weight: 600;
  margin: 4px 0 12px;
  color: var(--text-primary);
}

.editable { cursor: pointer; }
.editable:hover .edit-hint { opacity: 1; }

.edit-hint {
  font-size: 12px;
  color: var(--brand-500);
  opacity: 0;
  transition: opacity 0.15s;
  margin-left: 6px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section-head h3 { margin: 0; }

.edit-row {
  margin-bottom: 8px;
}

.edit-row--inline {
  display: inline-flex;
  gap: 6px;
  align-items: center;
}

.edit-input,
.edit-textarea,
.edit-select {
  padding: 6px 10px;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-2);
  border: 1px solid var(--brand-500);
  border-radius: var(--radius-sm);
  outline: none;
}

.edit-input {
  width: 100%;
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 6px;
}

.edit-textarea {
  width: 100%;
  resize: vertical;
  margin-bottom: 6px;
}

.edit-select {
  cursor: pointer;
}

.edit-row__actions {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
}

.issue-detail__meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}

/* 社交反馈栏（投票 / 反应 / 关注） */
.issue-detail__social {
  margin-bottom: 16px;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-weight: 500;
}
.badge-requirement { background: var(--brand-50); color: var(--brand-600); }
.badge-task { background: var(--success-50); color: var(--success-600); }
.badge-defect { background: var(--danger-50); color: var(--danger-600); }

.issue-detail__state-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  color: var(--text-on-brand);
  font-weight: 500;
}

.issue-detail__priority, .issue-detail__field {
  font-size: 12px;
  color: var(--text-secondary);
}

.issue-detail__delay-reason strong {
  color: var(--text-primary);
}

.delay-reason--required {
  color: var(--danger-500, #e5484d);
  font-size: 11px;
  margin-left: 4px;
}

.issue-detail__section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--border-subtle);
}

.issue-detail__section h3 {
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 12px;
  color: var(--text-primary);
}

.issue-detail__desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.issue-detail__fields {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row {
  display: flex;
  gap: 12px;
  font-size: 13px;
}

.field-label {
  color: var(--text-tertiary);
  min-width: 80px;
}

.field-value {
  color: var(--text-secondary);
}

.field-value--grow {
  flex: 1;
  min-width: 0;
}

.field-row--editable .field-value {
  display: flex;
  align-items: flex-start;
}

.editable--textarea {
  white-space: pre-wrap;
  min-height: 2.4em;
}

.add-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--danger-50);
  color: var(--danger-500);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  flex-shrink: 0;
}

.add-btn:hover {
  background: var(--danger-100);
}

.text-muted {
  color: var(--text-tertiary);
  font-size: 13px;
}

.issue-detail__transitions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.form-error { color: var(--danger-500); font-size: 12px; margin-bottom: 8px; }

.btn {
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn--ghost {
  background: none;
  border: none;
  color: var(--brand-500);
  padding: 4px 0;
}

.btn--danger {
  background: var(--danger-50);
  border-color: var(--danger-200);
  color: var(--danger-600);
}

.btn--state {
  background: var(--surface-1);
  font-weight: 500;
}

.activity-timeline {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.activity-item {
  display: flex;
  gap: 10px;
}

.activity-item__icon {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--brand-300);
  margin-top: 6px;
  flex-shrink: 0;
}
.activity-item__icon.verb-created { background: var(--success-500); }
.activity-item__icon.verb-transitioned { background: var(--brand-500); }

.activity-item__text {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.activity-item__time {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

.loading, .error {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}
.error { color: var(--danger-500); }

/* ===== Time Log ===== */
.timelog-summary {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
  font-weight: 500;
}

.timelog-effort {
  color: var(--text-tertiary);
  font-weight: 400;
}

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-family: inherit;
}

.btn--sm:hover:not(:disabled) {
  background: var(--surface-3);
}

.btn--outline {
  border: 1px dashed var(--border-strong);
  background: none;
  color: var(--text-tertiary);
  width: 100%;
  text-align: center;
}

.btn--outline:hover {
  border-color: var(--brand-500);
  color: var(--brand-500);
}

.btn--primary {
  background: var(--brand-500);
  color: var(--text-on-brand);
  border-color: var(--brand-500);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.btn:disabled,
.btn--sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.timelog-form {
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
  margin-top: 8px;
}

.timelog-form__row {
  margin-bottom: 8px;
}

.timelog-form__duration {
  display: flex;
  align-items: center;
  gap: 4px;
}

.timelog-label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0 4px;
}

.timelog-input {
  padding: 5px 8px;
  font-size: 12px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
  width: 100%;
}

.timelog-input:focus {
  border-color: var(--brand-500);
}

.timelog-input--sm {
  width: 72px;
}

.timelog-textarea {
  width: 100%;
  padding: 6px 8px;
  font-size: 12px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
  resize: vertical;
  margin-bottom: 8px;
}

.timelog-textarea:focus {
  border-color: var(--brand-500);
}

.timelog-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}

.timelog-list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.timelog-item {
  padding: 8px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

.timelog-item__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.timelog-item__date {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.timelog-item__duration {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.timelog-item__desc {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.timelog-item {
  position: relative;
}

.timelog-item__delete {
  position: absolute;
  top: 4px;
  right: 6px;
  border: none;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 11px;
  opacity: 0;
  transition: opacity 0.15s;
}
.timelog-item:hover .timelog-item__delete {
  opacity: 1;
}
.timelog-item__delete:hover {
  color: var(--danger-500);
}

/* ===== 子需求/任务/缺陷树 ===== */
.sub-issues-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.sub-issues-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.sub-issues-tree {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sub-issue-node {
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

/* 嵌套子树容器：左侧竖线引导 + 缩进 */
.sub-issue-node__children {
  margin-top: 6px;
  margin-left: 14px;
  padding-left: 10px;
  border-left: 2px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sub-issue-node--nested {
  padding: 6px 8px;
  background: var(--surface-1);
  border-style: dashed;
}

.sub-issue-node__placeholder {
  font-size: 11px;
  padding: 4px 0;
  font-style: italic;
}

.sub-issue-node__toggle--leaf {
  cursor: default;
  color: var(--text-tertiary);
}

.sub-issue-node__row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.sub-issue-node__toggle {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 8px;
  padding: 0;
  transition: transform 0.15s;
}

.sub-issue-node__toggle--collapsed {
  transform: rotate(0deg);
}

.sub-issue-node__toggle:not(.sub-issue-node__toggle--collapsed) {
  transform: rotate(90deg);
}

.sub-issue-node__identifier {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--brand-500);
  cursor: pointer;
  font-weight: 500;
}

.sub-issue-node__identifier:hover {
  text-decoration: underline;
}

.sub-issue-node__type {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  font-weight: 500;
}

.sub-issue-node__type--requirement { background: var(--brand-50); color: var(--brand-600); }
.sub-issue-node__type--task { background: var(--success-50); color: var(--success-600); }
.sub-issue-node__type--defect { background: var(--danger-50); color: var(--danger-600); }

.sub-issue-node__state {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  color: var(--text-on-brand);
  margin-left: auto;
}

.sub-issue-node__title {
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  margin-top: 4px;
  line-height: 1.4;
}

.sub-issue-node__title:hover {
  color: var(--brand-500);
}

.sub-issue-node__actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
}

/* ===== S13 P1: 移动端响应式 ===== */
@media (max-width: 768px) {
  .issue-detail {
    position: fixed;
    inset: 0;
    z-index: 200;
    width: 100% !important;
    max-width: 100% !important;
    border-radius: 0;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }
  .issue-detail__header {
    position: sticky;
    top: 0;
    z-index: 2;
    padding: 12px 16px;
  }
  .issue-detail__actions {
    flex-wrap: wrap;
    gap: 6px;
  }
  .issue-detail__actions button {
    min-height: 44px;
    min-width: 44px;
    flex: 1 1 auto;
    font-size: 13px;
  }
  .issue-detail__body {
    flex-direction: column;
    padding: 12px;
  }
  .issue-detail__main {
    width: 100%;
  }
  .issue-detail__sidebar {
    width: 100% !important;
    border-left: none !important;
    border-top: 1px solid var(--border-subtle, #e4e4e7);
    margin-top: 16px;
    padding-top: 16px;
  }
  /* 触控友好的交互元素 */
  .issue-detail input,
  .issue-detail select,
  .issue-detail textarea {
    font-size: 16px; /* 避免 iOS 自动缩放 */
  }
}

</style>
