<script setup lang="ts">
/**
 * 工作项详情页 — 展示描述、状态流转、活动日志与工时记录。
 */

import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { issueApi, type Issue, type IssueActivity, type State, type TimeLog } from "@/api/services/issue";
import { workspaceApi, type Workspace } from "@/api/services/workspace";
import { attachmentApi, type Attachment } from "@/api/services/attachment";
import { toast } from "@/lib/toast";
import { useAuthStore } from "@/stores/auth";
import RichTextEditor from "@/components/RichTextEditor.vue";
import CommentList from "@/components/CommentList.vue";
import AttachmentUploader from "@/components/AttachmentUploader.vue";
import RelationPanel from "./RelationPanel.vue";
import IssueCreateModal from "./IssueCreateModal.vue";
import { AppLoadingState, AppErrorState, AppEmptyState, IssueSocialBar } from "@/components";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
}>();

const router = useRouter();

const auth = useAuthStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

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
  if (!ws.value || !confirm("确定要归档该工作项吗？")) return;
  try {
    await issueApi.deleteIssue(ws.value.id, props.projectId, props.issueId);
    toast.success("工作项已归档");
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
  return ({ requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[type] ?? type;
}

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
});
</script>

<template>
  <div class="issue-detail">
    <header class="issue-detail__header">
      <button class="btn btn--ghost" @click="goBack">← 返回看板</button>
      <div class="issue-detail__actions">
        <button
          v-if="issue && issue.type_code === 'requirement'"
          class="btn btn--sm"
          @click="showDefectModal = true"
        >
          🐛 提缺陷
        </button>
        <button class="btn btn--danger" @click="doDelete">归档</button>
      </div>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!issue"
      title="工作项不存在或已被删除"
      description="请检查工作项 ID 是否正确"
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
        <h1 v-else class="issue-detail__name editable" @click="startEdit('name', issue.name)">
          {{ issue.name }}
          <span class="edit-hint">✎</span>
        </h1>

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
          <span v-else class="issue-detail__priority editable" @click="startEdit('priority', issue.priority)">
            优先级: {{ ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[issue.priority] ?? issue.priority }}
            <span class="edit-hint">✎</span>
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
        </div>

        <!-- 社交反馈栏：投票 / 表情反应 / 关注 -->
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
            <div v-if="!editingDesc" class="section-head__actions">
              <button class="btn btn--sm btn--ghost" @click="startEditDesc">编辑</button>
            </div>
          </div>
          <!-- 编辑模式：TipTap 富文本编辑器 -->
          <div v-if="editingDesc" class="edit-row">
            <RichTextEditor
              ref="descEditor"
              v-model:content-html="descHtml"
              v-model:content-json="descJsonValue"
              placeholder="输入工作项描述..."
              :min-height="'200px'"
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
            <div v-if="issue.reproduce_steps" class="field-row">
              <span class="field-label">复现步骤:</span>
              <span class="field-value">{{ JSON.stringify(issue.reproduce_steps) }}</span>
            </div>
            <div v-if="issue.environment" class="field-row">
              <span class="field-label">环境:</span>
              <span class="field-value">{{ JSON.stringify(issue.environment) }}</span>
            </div>
            <div v-if="issue.root_cause_category" class="field-row">
              <span class="field-label">根因分类:</span>
              <span class="field-value">{{ issue.root_cause_category }}</span>
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

        <!-- 流转操作 -->
        <div class="issue-detail__section">
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
                {{ act.verb === "created" ? "创建了工作项" : act.verb === "transitioned" ? `流转状态: ${act.old_value} → ${act.new_value}` : `${act.field}: ${act.old_value} → ${act.new_value}` }}
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

        <!-- 关联关系 -->
        <RelationPanel
          v-if="ws"
          :workspace-id="ws.id"
          :project-id="props.projectId"
          :issue-id="props.issueId"
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
</style>
