<script setup lang="ts">
/**
 * 版本发布页 — 发布前检查清单、发布说明生成与确认发布。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { versionApi, type Version, type DeliveryReport, type ChecklistItem } from "@/api/services/version";
import { AppBadge, AppButton, AppLoadingState, AppErrorState, ProgressBar } from "@/components";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const versionId = computed(() => Number(route.params.versionId));

const STEPS = [
  { key: "checklist", label: "检查清单", num: 1 },
  { key: "notes", label: "Release Notes", num: 2 },
  { key: "confirm", label: "确认发布", num: 3 },
] as const;

type StepKey = (typeof STEPS)[number]["key"];

const currentStep = ref<StepKey>("checklist");
const version = ref<Version | null>(null);
const report = ref<DeliveryReport | null>(null);
const loading = ref(true);
const releasing = ref(false);
const error = ref("");

// form state
const draftOverride = ref("");
const forceChecklist = ref(false);
const addKnown = ref(true);

// checklist inline editing
const newChecklistLabel = ref("");
const savingChecklist = ref(false);

let wsIdVal = 0;

/* ---------- computed ---------- */

const checklistAllDone = computed(() => {
  if (!version.value?.checklist?.length) return true;
  return version.value.checklist.every((c) => !c.required || c.checked);
});

const qualityPassed = computed(() => {
  return (version.value?.quality?.critical_bugs ?? 0) === 0;
});

const canProceedFromChecklist = computed(() => {
  return checklistAllDone.value || forceChecklist.value;
});

const canPublish = computed(() => {
  return canProceedFromChecklist.value && qualityPassed.value;
});

/* ---------- data ---------- */

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.get(workspaceId.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const [v, r] = await Promise.all([
      versionApi.getVersion(wsId, projectId.value, versionId.value),
      versionApi.getDeliveryReport(wsId, projectId.value, versionId.value).catch(() => null),
    ]);
    version.value = v;
    report.value = r;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

/* ---------- step navigation ---------- */

const stepIndex = computed(() => STEPS.findIndex((s) => s.key === currentStep.value));

function goTo(step: StepKey) {
  if (step === "notes" && !canProceedFromChecklist.value) return;
  currentStep.value = step;
}

function next() {
  const idx = stepIndex.value;
  if (idx < STEPS.length - 1) {
    goTo(STEPS[idx + 1].key);
  }
}

function prev() {
  const idx = stepIndex.value;
  if (idx > 0) {
    currentStep.value = STEPS[idx - 1].key;
  }
}

/* ---------- checklist inline edit ---------- */

/** 当前 checklist 副本（用于本地编辑同步到 UI） */
const editingChecklist = computed({
  get: () => version.value?.checklist ?? [],
  set: (val: ChecklistItem[]) => {
    if (version.value) version.value.checklist = val;
  },
});

/** 持久化当前 checklist 到后端 */
async function persistChecklist(items: ChecklistItem[]) {
  if (!version.value) return;
  savingChecklist.value = true;
  try {
    const wsId = await resolveWsId();
    const updated = await versionApi.updateVersion(wsId, projectId.value, versionId.value, {
      checklist: items,
      version: version.value.version ?? 0,
    });
    version.value = updated;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "保存清单失败";
  } finally {
    savingChecklist.value = false;
  }
}

/** 切换某条 checked 状态 */
async function toggleCheckItem(item: ChecklistItem) {
  const list = editingChecklist.value.map((c) =>
    c.id === item.id ? { ...c, checked: !c.checked } : c,
  );
  editingChecklist.value = list;
  await persistChecklist(list);
}

/** 切换某条 required 状态 */
async function toggleRequired(item: ChecklistItem) {
  const list = editingChecklist.value.map((c) =>
    c.id === item.id ? { ...c, required: !c.required } : c,
  );
  editingChecklist.value = list;
  await persistChecklist(list);
}

/** 删除某条 checklist */
async function removeCheckItem(item: ChecklistItem) {
  const list = editingChecklist.value.filter((c) => c.id !== item.id);
  editingChecklist.value = list;
  await persistChecklist(list);
}

/** 新增 checklist 条目 */
async function addCheckItem() {
  const label = newChecklistLabel.value.trim();
  if (!label) return;
  const newItem: ChecklistItem = {
    id: `chk-${Date.now()}`,
    label,
    required: false,
    checked: false,
  };
  const list = [...editingChecklist.value, newItem];
  editingChecklist.value = list;
  newChecklistLabel.value = "";
  await persistChecklist(list);
}

/* ---------- publish ---------- */

async function release() {
  if (!version.value) return;
  releasing.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    await versionApi.releaseVersion(wsId, projectId.value, versionId.value, {
      draft_override: draftOverride.value || undefined,
      force_checklist: forceChecklist.value,
      add_known_issues_to_notes: addKnown.value,
    });
    router.push({ name: "version-detail", params: { versionId: versionId.value } });
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "发布失败";
  } finally {
    releasing.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="release-wizard">
    <!-- Loading -->
    <AppLoadingState v-if="loading" text="正在加载发布向导..." />

    <!-- Error -->
    <AppErrorState
      v-else-if="error && !version"
      :message="error"
      @retry="load"
    />

    <template v-else-if="version">
      <!-- Header -->
      <header class="release-wizard__header">
        <div>
          <h1 class="release-wizard__title">发布向导</h1>
          <p class="release-wizard__subtitle">{{ version.name }} · {{ version.semver }}</p>
        </div>
      </header>

      <!-- Step indicator -->
      <nav class="stepper">
        <div
          v-for="(step, i) in STEPS"
          :key="step.key"
          class="stepper__step"
          :class="{
            'stepper__step--active': currentStep === step.key,
            'stepper__step--done': stepIndex > i,
            'stepper__step--disabled':
              step.key === 'notes' && !canProceedFromChecklist,
          }"
        >
          <div class="stepper__dot">
            <span v-if="stepIndex > i">✓</span>
            <span v-else>{{ step.num }}</span>
          </div>
          <span class="stepper__label">{{ step.label }}</span>
          <div v-if="i < STEPS.length - 1" class="stepper__line"></div>
        </div>
      </nav>

      <!-- Error banner -->
      <div v-if="error" class="release-wizard__banner error">
        {{ error }}
      </div>

      <!-- ==================== Step 1: Checklist ==================== -->
      <div v-if="currentStep === 'checklist'" class="step-content">
        <section class="step-section">
          <h2 class="step-section__title">
            ① 发布检查清单
            <span v-if="checklistAllDone" class="step-section__badge step-section__badge--pass">
              全部完成
            </span>
            <span v-else class="step-section__badge step-section__badge--warn">
              有待完成项
            </span>
          </h2>
          <p class="step-section__desc">
            确保所有必做检查项已完成，这是发布流程的第一步。
          </p>

          <!-- 可编辑的检查清单 -->
          <div v-if="editingChecklist.length > 0" class="checklist-editor">
            <ul class="checklist">
              <li
                v-for="item in editingChecklist"
                :key="item.id"
                class="checklist__item"
                :class="{
                  'checklist__item--done': item.checked,
                  'checklist__item--required': item.required && !item.checked,
                }"
              >
                <label class="checklist__check">
                  <input
                    type="checkbox"
                    :checked="item.checked"
                    :disabled="savingChecklist"
                    @change="toggleCheckItem(item)"
                  />
                  <span class="checklist__icon">{{ item.checked ? '✓' : '○' }}</span>
                </label>
                <span class="checklist__text">{{ item.label }}</span>
                <!-- Required toggle -->
                <label class="checklist__required-toggle" title="标记为必过项">
                  <input
                    type="checkbox"
                    :checked="item.required"
                    :disabled="savingChecklist"
                    @change="toggleRequired(item)"
                  />
                  <span class="checklist__required-label">必做</span>
                </label>
                <button
                  class="checklist__remove"
                  title="删除"
                  :disabled="savingChecklist"
                  @click="removeCheckItem(item)"
                >
                  ×
                </button>
              </li>
            </ul>
          </div>
          <p v-else class="step-section__empty">暂无检查项。使用下方按钮添加。</p>

          <!-- 新增条目输入区 -->
          <div class="checklist-add">
            <input
              v-model="newChecklistLabel"
              class="checklist-add__input"
              placeholder="新增检查项..."
              :disabled="savingChecklist"
              @keydown.enter.prevent="addCheckItem"
            />
            <button
              class="btn btn--sm"
              :disabled="!newChecklistLabel.trim() || savingChecklist"
              @click="addCheckItem"
            >
              {{ savingChecklist ? '保存中...' : 'Add Checklist Item' }}
            </button>
          </div>

          <label class="force-toggle">
            <input v-model="forceChecklist" type="checkbox" />
            <span>强制跳过清单校验（管理员豁免）</span>
          </label>
        </section>
      </div>

      <!-- ==================== Step 2: Release Notes ==================== -->
      <div v-if="currentStep === 'notes'" class="step-content">
        <section class="step-section">
          <h2 class="step-section__title">② Release Notes 预览与编辑</h2>
          <p class="step-section__desc">
            系统将自动从已完成的工作项和缺陷生成 Release Notes，你也可以手动编辑。
          </p>

          <label class="notes-toggle">
            <input v-model="addKnown" type="checkbox" />
            <span>在 Release Notes 中附带已知未关闭问题列表</span>
          </label>

          <label class="notes-field">
            <span class="notes-field__label">自定义 Notes（可选）</span>
            <textarea
              v-model="draftOverride"
              placeholder="留空则使用自动生成的 Release Notes。填写内容将完全覆盖自动生成结果。"
              class="notes-field__textarea"
              rows="10"
            ></textarea>
          </label>

          <div v-if="!draftOverride" class="notes-preview">
            <div class="notes-preview__header">自动生成预览（摘要）</div>
            <div class="notes-preview__body">
              <p>系统将在发布时自动生成包含以下内容的 Release Notes：</p>
              <ul>
                <li>✅ 已完成需求与任务</li>
                <li>🐛 已修复缺陷</li>
                <li v-if="addKnown">⚠️ 已知问题（未关闭）</li>
              </ul>
            </div>
          </div>
          <div v-else class="notes-preview">
            <div class="notes-preview__header">自定义 Notes 预览</div>
            <pre class="notes-preview__md">{{ draftOverride }}</pre>
          </div>
        </section>
      </div>

      <!-- ==================== Step 3: Confirm ==================== -->
      <div v-if="currentStep === 'confirm'" class="step-content">
        <section class="step-section">
          <h2 class="step-section__title">③ 确认发布</h2>
          <p class="step-section__desc">
            请确认以下信息无误后点击「确认发布」按钮。
          </p>

          <!-- Quality gate -->
          <div class="confirm-block">
            <div class="confirm-block__header">
              <span>质量门禁</span>
              <AppBadge :variant="qualityPassed ? 'success' : 'danger'">
                {{ qualityPassed ? '通过' : '不通过' }}
              </AppBadge>
            </div>
            <div class="confirm-block__body">
              <div class="confirm-stat">
                <span class="confirm-stat__label">致命/严重未关闭缺陷</span>
                <span
                  class="confirm-stat__value"
                  :class="qualityPassed ? 'text-success' : 'text-danger'"
                >{{ version.quality?.critical_bugs ?? 0 }}</span>
              </div>
              <div class="confirm-stat">
                <span class="confirm-stat__label">修复率</span>
                <span class="confirm-stat__value">
                  {{ Math.round((version.quality?.fix_rate ?? 0) * 100) }}%
                </span>
              </div>
            </div>
          </div>

          <!-- Delivery summary -->
          <div v-if="report" class="confirm-block">
            <div class="confirm-block__header">
              <span>交付摘要</span>
              <AppBadge :variant="report.eligible_to_release ? 'success' : 'warning'">
                {{ report.eligible_to_release ? '满足准出条件' : '未满足准出条件' }}
              </AppBadge>
            </div>
            <div class="confirm-block__body">
              <div class="confirm-stats-grid">
                <div class="confirm-stat">
                  <span class="confirm-stat__label">迭代数</span>
                  <span class="confirm-stat__value">{{ report.sprint_count }}</span>
                </div>
                <div class="confirm-stat">
                  <span class="confirm-stat__label">已完成 / 总工作项</span>
                  <span class="confirm-stat__value">{{ report.completed_issues }} / {{ report.total_issues }}</span>
                </div>
                <div class="confirm-stat">
                  <span class="confirm-stat__label">已完成 / 总故事点</span>
                  <span class="confirm-stat__value">{{ Math.round(report.completed_points) }} / {{ Math.round(report.total_points) }}</span>
                </div>
                <div class="confirm-stat">
                  <span class="confirm-stat__label">已修复 / 总缺陷</span>
                  <span class="confirm-stat__value">{{ report.fixed_bug_count }} / {{ report.bug_count }}</span>
                </div>
                <div class="confirm-stat">
                  <span class="confirm-stat__label">通过率</span>
                  <span class="confirm-stat__value">{{ Math.round(report.pass_rate * 100) }}%</span>
                </div>
                <div class="confirm-stat">
                  <span class="confirm-stat__label">进度</span>
                  <span class="confirm-stat__value">{{ Math.round((version.progress?.completion_rate ?? 0) * 100) }}%</span>
                </div>
              </div>

              <!-- Progress bar -->
              <div class="confirm-progress">
                <ProgressBar
                  :percent="Math.round((version.progress?.completion_rate ?? 0) * 100)"
                  size="md"
                  :color="(version.progress?.completion_rate ?? 0) >= 0.8 ? 'var(--success-500)' : 'var(--warning-500)'"
                />
              </div>

              <!-- Checklist summary -->
              <div class="confirm-checklist-summary">
                <span class="confirm-checklist-summary__label">检查清单</span>
                <span>
                  {{ (version.checklist ?? []).filter(c => c.checked).length }}/{{ version.checklist?.length ?? 0 }} 项完成
                  <template v-if="forceChecklist">（已强制跳过校验）</template>
                </span>
              </div>
            </div>
          </div>

          <!-- Warning if not eligible -->
          <div
            v-if="!canPublish"
            class="release-wizard__banner warn"
          >
            <template v-if="!checklistAllDone && !forceChecklist">
              ⚠️ 检查清单有未完成的必做项
            </template>
            <template v-if="!qualityPassed">
              ⚠️ 存在未关闭的致命/严重缺陷，质量门禁未通过
            </template>
            <template v-if="!checklistAllDone && !forceChecklist && !qualityPassed">
              <br />
            </template>
          </div>
        </section>
      </div>

      <!-- Navigation buttons -->
      <footer class="release-wizard__footer">
        <AppButton
          v-if="currentStep !== 'checklist'"
          variant="secondary"
          @click="prev"
        >
          上一步
        </AppButton>
        <div class="release-wizard__footer-spacer"></div>
        <AppButton
          variant="secondary"
          @click="router.back()"
        >
          取消
        </AppButton>
        <AppButton
          v-if="currentStep === 'checklist'"
          variant="primary"
          :disabled="!canProceedFromChecklist"
          @click="next"
        >
          {{ !canProceedFromChecklist ? '请先完成检查清单' : '下一步' }}
        </AppButton>
        <AppButton
          v-if="currentStep === 'notes'"
          variant="primary"
          @click="next"
        >
          下一步：确认发布
        </AppButton>
        <AppButton
          v-if="currentStep === 'confirm'"
          variant="primary"
          :loading="releasing"
          :disabled="!canPublish"
          @click="release"
        >
          {{ releasing ? '发布中…' : '确认发布' }}
        </AppButton>
      </footer>
    </template>
  </div>
</template>

<style scoped>
.release-wizard {
  max-width: 720px;
}

/* ---- header ---- */
.release-wizard__header {
  margin-bottom: 24px;
}
.release-wizard__title { margin: 0; font-size: 20px; font-weight: 600; }
.release-wizard__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

/* ---- banner ---- */
.release-wizard__banner {
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.release-wizard__banner.error {
  background: rgba(220, 47, 47, 0.08);
  color: var(--danger-500);
  border: 1px solid rgba(220, 47, 47, 0.2);
}
.release-wizard__banner.warn {
  background: rgba(245, 158, 11, 0.08);
  color: var(--warning-500);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

/* ---- stepper ---- */
.stepper {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  padding: 16px 0;
}
.stepper__step {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  position: relative;
}
.stepper__step:last-child { flex: 0; }
.stepper__dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
  background: var(--surface-3);
  color: var(--text-tertiary);
  transition: background 0.2s, color 0.2s;
}
.stepper__step--active .stepper__dot {
  background: var(--brand-500);
  color: var(--text-on-brand);
}
.stepper__step--done .stepper__dot {
  background: var(--success-500);
  color: var(--text-on-brand);
}
.stepper__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-tertiary);
  white-space: nowrap;
}
.stepper__step--active .stepper__label { color: var(--text-primary); }
.stepper__step--done .stepper__label { color: var(--success-500); }
.stepper__step--disabled { opacity: 0.5; }
.stepper__line {
  flex: 1;
  height: 2px;
  background: var(--surface-3);
  margin-left: 12px;
  min-width: 24px;
}
.stepper__step--done .stepper__line {
  background: var(--success-500);
}

/* ---- step content ---- */
.step-content {
  margin-bottom: 24px;
}
.step-section {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 20px;
}
.step-section__title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}
.step-section__badge {
  font-size: 11px;
  font-weight: 500;
  padding: 1px 8px;
  border-radius: 10px;
}
.step-section__badge--pass {
  background: rgba(15, 194, 123, 0.1);
  color: var(--success-500);
}
.step-section__badge--warn {
  background: rgba(245, 158, 11, 0.1);
  color: var(--warning-500);
}
.step-section__desc {
  margin: 0 0 16px;
  font-size: 13px;
  color: var(--text-tertiary);
}
.step-section__empty {
  font-size: 13px;
  color: var(--text-tertiary);
  padding: 16px 0;
  text-align: center;
}

/* ---- checklist ---- */
.checklist {
  list-style: none;
  margin: 0 0 16px;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.checklist__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--success-500);
}
.checklist__item--required {
  border-left-color: var(--danger-500);
  background: rgba(220, 47, 47, 0.04);
}
.checklist__icon {
  font-size: 14px;
  width: 18px;
  text-align: center;
  flex-shrink: 0;
}
.checklist__item--done .checklist__icon { color: var(--success-500); }
.checklist__item--required .checklist__icon { color: var(--danger-500); }
.checklist__text { flex: 1; font-size: 13px; }

/* ---- checklist inline editing ---- */
.checklist__check {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
}

.checklist__check input {
  accent-color: var(--brand-500);
}

.checklist__required-toggle {
  display: flex;
  align-items: center;
  gap: 3px;
  cursor: pointer;
  flex-shrink: 0;
}

.checklist__required-toggle input {
  accent-color: var(--danger-500);
}

.checklist__required-label {
  font-size: 11px;
  color: var(--text-tertiary);
  white-space: nowrap;
}

.checklist__required-toggle input:checked + .checklist__required-label {
  color: var(--danger-500);
  font-weight: 600;
}

.checklist__remove {
  background: none;
  border: none;
  color: var(--text-tertiary);
  font-size: 18px;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
  flex-shrink: 0;
  border-radius: 4px;
  transition: color 0.1s, background 0.1s;
}

.checklist__remove:hover:not(:disabled) {
  color: var(--danger-500);
  background: var(--danger-50);
}

.checklist__remove:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.checklist-add {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}

.checklist-add__input {
  flex: 1;
  padding: 6px 10px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-1, #fff);
  color: var(--text-primary, #1f2937);
  outline: none;
}

.checklist-add__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.btn--sm {
  padding: 6px 12px;
  font-size: 12px;
  font-family: inherit;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  border: 1px solid var(--border-default, #d1d5db);
  background: var(--brand-500);
  color: #fff;
  white-space: nowrap;
}

.btn--sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--sm:hover:not(:disabled) {
  background: var(--brand-600);
}

.force-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-tertiary);
  cursor: pointer;
}
.force-toggle input { accent-color: var(--brand-500); }

/* ---- notes step ---- */
.notes-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  margin-bottom: 16px;
}
.notes-toggle input { accent-color: var(--brand-500); }

.notes-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}
.notes-field__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}
.notes-field__textarea {
  font-size: 13px;
  font-family: var(--font-mono);
  padding: 10px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
  color: var(--text-primary);
  resize: vertical;
  min-height: 160px;
}
.notes-field__textarea:focus {
  outline: none;
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.notes-preview {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  overflow: hidden;
}
.notes-preview__header {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary);
  padding: 8px 12px;
  background: var(--surface-2);
  border-bottom: 1px solid var(--border-subtle);
}
.notes-preview__body {
  padding: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}
.notes-preview__body ul { margin: 8px 0 0; padding-left: 20px; }
.notes-preview__md {
  margin: 0;
  padding: 12px;
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-primary);
  white-space: pre-wrap;
  line-height: 1.6;
}

/* ---- confirm step ---- */
.confirm-block {
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  margin-bottom: 12px;
  overflow: hidden;
}
.confirm-block__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--surface-1);
}
.confirm-block__body {
  padding: 14px;
}

.confirm-stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.confirm-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.confirm-stat__label {
  font-size: 11px;
  color: var(--text-tertiary);
}
.confirm-stat__value {
  font-size: 14px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-primary);
}
.confirm-stat__value.text-success { color: var(--success-500); }
.confirm-stat__value.text-danger { color: var(--danger-500); }

.confirm-progress {
  margin-top: 14px;
}

.confirm-checklist-summary {
  display: flex;
  justify-content: space-between;
  margin-top: 14px;
  padding-top: 10px;
  border-top: 1px solid var(--border-subtle);
  font-size: 12px;
  color: var(--text-tertiary);
}
.confirm-checklist-summary__label {
  font-weight: 500;
  color: var(--text-secondary);
}

/* ---- footer ---- */
.release-wizard__footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 0;
  border-top: 1px solid var(--border-subtle);
}
.release-wizard__footer-spacer { flex: 1; }

.text-success { color: var(--success-500); }
.text-danger { color: var(--danger-500); }
</style>
