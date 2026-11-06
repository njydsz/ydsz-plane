<script setup lang="ts">
/**
 * Projects-as-Code 视图。
 *
 * 以 YAML 声明式定义项目（状态 / 标签 / 模块 / 成员 / 设置），
 * 支持保存配置、校验解析、与既有项目对比、应用下发、以及反向导出。
 *
 * 能力来自后端 projectcode 域：parse / diff / apply / export / configs CRUD。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  projectCodeApi,
  type ProjectCodeConfig,
  type ProjectCodeDiffItem,
  type ProjectCodeDiffResult,
  type ProjectSpec,
} from "@/api/services/projectcode";
import { workspaceApi, type Project } from "@/api/services/workspace";
import { AppBadge, AppButton, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer } from "@/components";
import { useTranslation } from "@/composables/useTranslation";
import { formatRelativeTime } from "@/lib/formatTime";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const { t } = useTranslation();
const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

/* ------------------------------------------------------------------ */
/* 列表数据                                                            */
/* ------------------------------------------------------------------ */

const configs = ref<ProjectCodeConfig[]>([]);
const projects = ref<Project[]>([]);
const loading = ref(true);
const error = ref("");

/** 项目 ID → 项目信息（用于展示「关联项目」链接） */
const projectMap = computed(() => {
  const map = new Map<number, Project>();
  for (const p of projects.value) map.set(p.id, p);
  return map;
});

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const [cfgs, ps] = await Promise.all([
      projectCodeApi.listConfigs(workspaceId.value),
      workspaceApi.listProjects(workspaceId.value).catch(() => [] as Project[]),
    ]);
    configs.value = cfgs ?? [];
    projects.value = ps ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t("projectCode.loadFailed", "加载失败");
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function toProject(projectId?: number | null) {
  if (!projectId || !workspaceId.value) return undefined;
  router.push(`/${workspaceId.value}/projects/${projectId}`);
}

/* ------------------------------------------------------------------ */
/* YAML 编辑器                                                         */
/* ------------------------------------------------------------------ */

const showEditor = ref(false);
/** 当前打开的配置（编辑模式）；null 表示新建 */
const editingId = ref<number | null>(null);
const yamlText = ref("");
const configName = ref("");
const editorError = ref("");
const parsing = ref(false);

/** 内置示例 YAML，帮助用户快速上手 */
const SAMPLE_YAML = `version: "1.0"
name: "Apollo"
key: "APOLLO"
color: "var(--brand-500)"
icon: "🚀"
description: "Apollo program"
states:
  - name: "待办"
    group: "backlog"
    color: "var(--brand-200)"
  - name: "进行中"
    group: "started"
    color: "var(--warning-500)"
  - name: "已完成"
    group: "completed"
    color: "var(--success-500)"
labels:
  - name: "后端"
    color: "var(--info-500)"
modules:
  - name: "API"
    lead: "alice@example.com"
members:
  - email: "alice@example.com"
    role: "admin"
settings:
  default_assignee: "alice@example.com"
  estimate_points: [1, 2, 3, 5, 8]
`;

function openCreateEditor() {
  editingId.value = null;
  yamlText.value = SAMPLE_YAML;
  configName.value = "";
  editorError.value = "";
  showEditor.value = true;
}

function openEditEditor(cfg: ProjectCodeConfig) {
  editingId.value = cfg.id;
  yamlText.value = cfg.yaml;
  configName.value = cfg.name === cfg.spec?.name ? "" : cfg.name;
  editorError.value = "";
  showEditor.value = true;
}

function closeEditor() {
  showEditor.value = false;
  parsing.value = false;
}

/* ---------------- 校验（parse） ---------------- */

const parseResult = ref<ProjectSpec | null>(null);
const parsingBusy = ref(false);

async function handleParse() {
  if (!yamlText.value.trim()) {
    editorError.value = t("projectCode.editor.yamlRequired", "请输入 YAML 内容");
    return;
  }
  parsingBusy.value = true;
  editorError.value = "";
  parseResult.value = null;
  try {
    parseResult.value = await projectCodeApi.parse(
      workspaceId.value,
      yamlText.value,
    );
    toast.success(t("projectCode.parse.ok", "校验通过"));
  } catch (e: unknown) {
    editorError.value =
      e instanceof Error ? e.message : t("projectCode.parse.failed", "YAML 校验失败");
  } finally {
    parsingBusy.value = false;
  }
}

/* ---------------- 保存配置 ---------------- */

const saving = ref(false);

async function handleSave() {
  if (!yamlText.value.trim()) {
    editorError.value = t("projectCode.editor.yamlRequired", "请输入 YAML 内容");
    return;
  }
  saving.value = true;
  editorError.value = "";
  try {
    await projectCodeApi.saveConfig(workspaceId.value, {
      yaml: yamlText.value,
      name: configName.value.trim() || undefined,
      project_id: selectedProjectId.value ?? undefined,
    });
    toast.success(t("projectCode.save.ok", "配置已保存"));
    closeEditor();
    await load();
  } catch (e: unknown) {
    editorError.value =
      e instanceof Error ? e.message : t("projectCode.save.failed", "保存失败");
  } finally {
    saving.value = false;
  }
}

/* ---------------- 目标项目选择（diff / apply） ---------------- */

const selectedProjectId = ref<number | null>(null);

/* ---------------- 对比（diff） ---------------- */

const showDiffPanel = ref(false);
const diffing = ref(false);
const diffError = ref("");
const diffResult = ref<ProjectCodeDiffResult | null>(null);

interface DiffGroup {
  key: string;
  label: string;
  variant: "success" | "warning" | "danger" | "info";
  items: ProjectCodeDiffItem[];
}

const RESOURCE_LABELS: Record<string, string> = {
  project: t("projectCode.diff.resource.project", "项目"),
  state: t("projectCode.diff.resource.state", "状态"),
  label: t("projectCode.diff.resource.label", "标签"),
  module: t("projectCode.diff.resource.module", "模块"),
  member: t("projectCode.diff.resource.member", "成员"),
  setting: t("projectCode.diff.resource.setting", "设置"),
};

const KIND_META: Record<
  string,
  { label: string; variant: "success" | "warning" | "danger" | "info" }
> = {
  add: {
    label: t("projectCode.diff.kind.add", "新增"),
    variant: "success",
  },
  update: {
    label: t("projectCode.diff.kind.update", "更新"),
    variant: "warning",
  },
  remove: {
    label: t("projectCode.diff.kind.remove", "移除"),
    variant: "danger",
  },
  noop: {
    label: t("projectCode.diff.kind.noop", "无变化"),
    variant: "info",
  },
};

const diffGroups = computed<DiffGroup[]>(() => {
  const res = diffResult.value;
  if (!res) return [];
  const groups = new Map<string, DiffGroup>();
  for (const item of res.items) {
    const key = item.resource || "other";
    let g = groups.get(key);
    if (!g) {
      g = {
        key,
        label: RESOURCE_LABELS[key] ?? key,
        variant: "info",
        items: [],
      };
      groups.set(key, g);
    }
    g.items.push(item);
  }
  return Array.from(groups.values());
});

/** 将 diff 字段值安全序列化为可读文本（避免对象默认 stringification）。 */
function displayValue(value: unknown): string {
  if (value === undefined || value === null) return "—";
  if (typeof value === "string") return value || "—";
  if (
    typeof value === "number" ||
    typeof value === "boolean" ||
    typeof value === "bigint"
  ) {
    return value.toString();
  }
  // 对象 / 数组：JSON 序列化兜底
  try {
    return JSON.stringify(value) ?? "—";
  } catch {
    return "—";
  }
}

function describeChange(item: ProjectCodeDiffItem): string {
  const kind = KIND_META[item.kind] ?? { label: item.kind, variant: "info" as const };
  const name = item.name ? `「${item.name}」` : "";
  if (item.field) {
    return `${kind.label} ${RESOURCE_LABELS[item.resource] ?? item.resource} ${name} · ${
      item.field
    }: ${displayValue(item.old)} → ${displayValue(item.new)}`;
  }
  return `${kind.label} ${RESOURCE_LABELS[item.resource] ?? item.resource} ${name}`;
}

async function handleDiff() {
  if (!selectedProjectId.value) {
    diffError.value = t("projectCode.diff.selectProject", "请先选择要对比的项目");
    return;
  }
  if (!yamlText.value.trim()) {
    diffError.value = t("projectCode.editor.yamlRequired", "请输入 YAML 内容");
    return;
  }
  diffing.value = true;
  diffError.value = "";
  diffResult.value = null;
  try {
    diffResult.value = await projectCodeApi.diff(
      workspaceId.value,
      selectedProjectId.value,
      yamlText.value,
    );
    showDiffPanel.value = true;
  } catch (e: unknown) {
    diffError.value =
      e instanceof Error ? e.message : t("projectCode.diff.failed", "对比失败");
  } finally {
    diffing.value = false;
  }
}

/* ---------------- 应用（apply） ---------------- */

const showApplyConfirm = ref(false);
const applying = ref(false);

function openApplyConfirm() {
  if (!yamlText.value.trim()) {
    editorError.value = t("projectCode.editor.yamlRequired", "请输入 YAML 内容");
    return;
  }
  editorError.value = "";
  showApplyConfirm.value = true;
}

async function handleApply() {
  applying.value = true;
  try {
    const res = await projectCodeApi.apply(
      workspaceId.value,
      yamlText.value,
      selectedProjectId.value ?? undefined,
    );
    toast.success(
      res.created
        ? t("projectCode.apply.created", "项目已创建")
        : t("projectCode.apply.updated", "项目已更新"),
    );
    showApplyConfirm.value = false;
    closeEditor();
    await load();
    // 应用成功后跳转到对应项目
    if (res.project_id) toProject(res.project_id);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("projectCode.apply.failed", "应用失败"));
  } finally {
    applying.value = false;
  }
}

/* ---------------- 删除配置 ---------------- */

async function removeConfig(cfg: ProjectCodeConfig) {
  if (!await confirm({
    title: "删除配置",
    message: t("projectCode.delete.confirm", `确认删除配置「${cfg.name}」？此操作不可撤销。`),
    confirmText: "删除",
    variant: "danger",
  })) {
    return;
  }
  try {
    await projectCodeApi.deleteConfig(workspaceId.value, cfg.id);
    toast.success(t("projectCode.delete.ok", "已删除"));
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("projectCode.delete.failed", "删除失败"));
  }
}

/* ---------------- 导出 ---------------- */

const showExport = ref(false);
const exporting = ref(false);
const exportError = ref("");
const exportYaml = ref("");
const exportSpec = ref<ProjectSpec | null>(null);
const exportProjectId = ref<number | null>(null);

async function handleExport() {
  if (!exportProjectId.value) {
    exportError.value = t("projectCode.export.selectProject", "请选择要导出的项目");
    return;
  }
  exporting.value = true;
  exportError.value = "";
  try {
    const res = await projectCodeApi.exportProject(
      workspaceId.value,
      exportProjectId.value,
    );
    exportYaml.value = res.yaml;
    exportSpec.value = res.spec;
  } catch (e: unknown) {
    exportError.value =
      e instanceof Error ? e.message : t("projectCode.export.failed", "导出失败");
  } finally {
    exporting.value = false;
  }
}

function closeExport() {
  showExport.value = false;
  exportYaml.value = "";
  exportSpec.value = null;
  exportError.value = "";
  exportProjectId.value = null;
}

async function copyExport() {
  if (!exportYaml.value) return;
  try {
    await navigator.clipboard.writeText(exportYaml.value);
    toast.success(t("common.copied", "已复制"));
  } catch {
    // Clipboard API 在非安全上下文（http）下不可用，降级为选中提示
    toast.warning(t("projectCode.export.copyFallback", "复制失败，请手动全选文本复制"));
  }
}

/* ------------------------------------------------------------------ */
/* 小工具                                                              */
/* ------------------------------------------------------------------ */

function shortSha(sha?: string): string {
  return sha ? sha.slice(0, 8) : "—";
}

function formatTime(input?: string): string {
  return input ? formatRelativeTime(input) : "—";
}
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="pac">
    <!-- 头部 -->
    <header class="pac__header">
      <div>
        <h1 class="pac__title">{{ t("projectCode.title", "项目即代码") }}</h1>
        <p class="pac__subtitle">
          {{ t("projectCode.subtitle", "以 YAML 声明式定义项目的状态、标签、模块、成员与设置，可校验、对比并一键下发") }}
        </p>
      </div>
      <div class="pac__actions">
        <AppButton variant="secondary" size="sm" @click="showExport = true">
          {{ t("projectCode.header.export", "导出项目") }}
        </AppButton>
        <AppButton variant="primary" size="sm" @click="openCreateEditor">
          {{ t("projectCode.header.import", "导入 YAML") }}
        </AppButton>
      </div>
    </header>

    <!-- 加载中 -->
    <template v-if="loading">
      <AppLoadingState :text="t('common.loading', '加载中...')" />
    </template>

    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <!-- 空状态 -->
    <AppEmptyState
      v-else-if="configs.length === 0"
      :title="t('projectCode.empty.title', '还没有保存的项目配置')"
      :description="
        t('projectCode.empty.description', '点击「导入 YAML」编写第一份项目定义，或从既有项目反向导出。')
      "
    >
      <div class="pac__empty-actions">
        <AppButton variant="primary" size="sm" @click="openCreateEditor">
          {{ t("projectCode.header.import", "导入 YAML") }}
        </AppButton>
        <AppButton variant="secondary" size="sm" @click="showExport = true">
          {{ t("projectCode.header.export", "导出项目") }}
        </AppButton>
      </div>
    </AppEmptyState>

    <!-- 已保存配置列表 -->
    <section v-else class="pac__list">
      <h2 class="pac__section-title">
        {{ t("projectCode.configs.title", "已保存配置") }}
        <span class="pac__count">{{ configs.length }}</span>
      </h2>

      <div class="pac__table-wrap">
        <table class="pac__table">
          <thead>
            <tr>
              <th>{{ t("projectCode.configs.name", "名称") }}</th>
              <th>{{ t("projectCode.configs.version", "版本") }}</th>
              <th>{{ t("projectCode.configs.sha", "SHA") }}</th>
              <th>{{ t("projectCode.configs.updatedAt", "更新时间") }}</th>
              <th>{{ t("projectCode.configs.project", "关联项目") }}</th>
              <th class="pac__th-actions">{{ t("common.actions", "操作") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="cfg in configs" :key="cfg.id">
              <td>
                <div class="pac__name">
                  <span class="pac__name-text">{{ cfg.name || cfg.spec?.name || "—" }}</span>
                  <AppBadge v-if="cfg.spec_key" variant="brand">
                    {{ cfg.spec_key }}
                  </AppBadge>
                </div>
                <div
                  v-if="cfg.spec"
                  class="pac__meta pac__meta--dim"
                >
                  {{ cfg.spec.states?.length ?? 0 }} {{ t("projectCode.summary.states", "状态") }}
                  · {{ cfg.spec.labels?.length ?? 0 }} {{ t("projectCode.summary.labels", "标签") }}
                  · {{ cfg.spec.modules?.length ?? 0 }} {{ t("projectCode.summary.modules", "模块") }}
                </div>
              </td>
              <td>
                <AppBadge variant="default">v{{ cfg.version || "—" }}</AppBadge>
              </td>
              <td>
                <code class="pac__mono">{{ shortSha(cfg.sha256) }}</code>
              </td>
              <td class="pac__meta">{{ formatTime(cfg.updated_at) }}</td>
              <td>
                <button
                  v-if="cfg.project_id && projectMap.has(cfg.project_id)"
                  class="pac__link"
                  @click="toProject(cfg.project_id)"
                >
                  {{ projectMap.get(cfg.project_id)?.identifier }}
                </button>
                <span v-else class="pac__meta pac__meta--dim">—</span>
              </td>
              <td class="pac__td-actions">
                <button class="pac__link" @click="openEditEditor(cfg)">
                  {{ t("common.edit", "编辑") }}
                </button>
                <button
                  class="pac__link pac__link--danger"
                  @click="removeConfig(cfg)"
                >
                  {{ t("common.delete", "删除") }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- YAML 编辑器 Modal -->
    <AppModal
      :visible="showEditor"
      :title="
        editingId
          ? t('projectCode.editor.editTitle', '编辑项目 YAML')
          : t('projectCode.editor.createTitle', '导入项目 YAML')
      "
      width="760px"
      @close="closeEditor"
    >
      <div class="editor">
        <div class="editor__hint">
          <span class="editor__hint-label">
            {{ t("projectCode.editor.hint", "语法提示") }}
          </span>
          <code class="editor__hint-code">
            version / name / key / color / icon / description / states / labels /
            modules / members / settings
          </code>
        </div>

        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.editor.nameLabel", "配置名（可选）") }}
          </span>
          <input
            v-model="configName"
            class="editor__input"
            :placeholder="t('projectCode.editor.namePlaceholder', '留空则使用 spec.name')"
            maxlength="200"
          />
        </label>

        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.editor.yamlLabel", "YAML 内容") }}
            <span class="editor__required">*</span>
          </span>
          <textarea
            v-model="yamlText"
            class="editor__textarea"
            rows="18"
            spellcheck="false"
            :placeholder="SAMPLE_YAML"
          ></textarea>
        </label>

        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.editor.targetProject", "目标项目（对比 / 应用时选择）") }}
          </span>
          <select
            v-model="selectedProjectId"
            class="editor__input"
          >
            <option :value="null">
              {{
                t(
                  "projectCode.editor.targetProjectNew",
                  "不选择 — 应用时将创建新项目",
                )
              }}
            </option>
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.identifier }} · {{ p.name }}
            </option>
          </select>
        </label>

        <div v-if="editorError" class="editor__error">{{ editorError }}</div>

        <!-- 校验结果摘要 -->
        <div v-if="parseResult" class="editor__summary">
          <div class="editor__summary-title">
            {{ t("projectCode.summary.title", "解析结果") }}
          </div>
          <div class="editor__summary-grid">
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ parseResult.key || "—" }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.projectKey", "项目标识") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.states?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.states", "状态") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.labels?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.labels", "标签") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.modules?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.modules", "模块") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.members?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.members", "成员") }}
              </span>
            </div>
          </div>
        </div>

        <!-- 对比结果 -->
        <div v-if="showDiffPanel && diffResult" class="editor__diff">
          <div class="editor__diff-head">
            <span class="editor__summary-title">
              {{ t("projectCode.diff.title", "对比结果") }}
            </span>
            <div class="editor__diff-stats">
              <AppBadge variant="success">
                +{{ diffResult.add_count }}
              </AppBadge>
              <AppBadge variant="warning">
                ~{{ diffResult.update_count }}
              </AppBadge>
              <AppBadge variant="danger">
                -{{ diffResult.remove_count }}
              </AppBadge>
            </div>
          </div>

          <p
            v-if="!diffResult.has_changes"
            class="editor__diff-empty"
          >
            {{ t("projectCode.diff.noChanges", "规范与当前项目一致，无需变更") }}
          </p>

          <div
            v-for="group in diffGroups"
            :key="group.key"
            class="editor__diff-group"
          >
            <div class="editor__diff-group-title">{{ group.label }}</div>
            <ul class="editor__diff-list">
              <li v-for="(item, idx) in group.items" :key="idx">
                <AppBadge :variant="KIND_META[item.kind]?.variant ?? 'info'">
                  {{ KIND_META[item.kind]?.label ?? item.kind }}
                </AppBadge>
                <span>{{ describeChange(item) }}</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="editor__footer">
          <div class="editor__footer-left">
            <AppButton
              variant="ghost"
              size="sm"
              :loading="parsingBusy"
              @click="handleParse"
            >
              {{ t("projectCode.action.validate", "校验") }}
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              :loading="diffing"
              @click="handleDiff"
            >
              {{ t("projectCode.action.diff", "对比") }}
            </AppButton>
          </div>
          <div class="editor__footer-right">
            <AppButton variant="secondary" size="sm" @click="closeEditor">
              {{ t("common.cancel", "取消") }}
            </AppButton>
            <AppButton
              variant="secondary"
              size="sm"
              :loading="saving"
              @click="handleSave"
            >
              {{ t("projectCode.action.save", "保存配置") }}
            </AppButton>
            <AppButton
              variant="primary"
              size="sm"
              @click="openApplyConfirm"
            >
              {{ t("projectCode.action.apply", "应用") }}
            </AppButton>
          </div>
        </div>
          </AppModal>

    <!-- 应用确认 Modal -->
    <AppModal
      :visible="showApplyConfirm"
      :title="t('projectCode.apply.confirmTitle', '确认应用规范')"
      width="480px"
      @close="showApplyConfirm = false"
    >
      <p class="pac__confirm-text">
        {{
          selectedProjectId
            ? t(
                "projectCode.apply.confirmUpdate",
                "将把当前 YAML 规范应用到所选项目，可能新增 / 修改项目的状态、标签、模块与成员。是否继续？",
              )
            : t(
                "projectCode.apply.confirmCreate",
                "将依据当前 YAML 规范创建一个新项目，并写入其状态、标签、模块与成员。是否继续？",
              )
        }}
      </p>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showApplyConfirm = false">
          {{ t("common.cancel", "取消") }}
        </AppButton>
        <AppButton
          variant="danger"
          size="sm"
          :loading="applying"
          @click="handleApply"
        >
          {{ t("common.confirm", "确认") }}
        </AppButton>
      
    </AppModal>

    <!-- 导出 Modal -->
    <AppModal
      :visible="showExport"
      :title="t('projectCode.export.title', '导出项目为 YAML')"
      width="720px"
      @close="closeExport"
    >
      <div class="export">
        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.export.selectLabel", "选择项目") }}
            <span class="editor__required">*</span>
          </span>
          <select v-model="exportProjectId" class="editor__input">
            <option :value="null">
              {{ t("projectCode.export.selectPlaceholder", "请选择要导出的项目") }}
            </option>
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.identifier }} · {{ p.name }}
            </option>
          </select>
        </label>

        <div class="export__actions">
          <AppButton
            variant="primary"
            size="sm"
            :loading="exporting"
            @click="handleExport"
          >
            {{ t("projectCode.export.generate", "生成 YAML") }}
          </AppButton>
          <AppButton
            v-if="exportYaml"
            variant="secondary"
            size="sm"
            @click="copyExport"
          >
            {{ t("common.copy", "复制") }}
          </AppButton>
        </div>

        <div v-if="exportError" class="editor__error">{{ exportError }}</div>

        <div v-if="exportSpec" class="editor__summary">
          <div class="editor__summary-title">
            {{ t("projectCode.summary.title", "解析结果") }}
          </div>
          <div class="editor__summary-grid">
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.key || "—" }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.projectKey", "项目标识") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.states?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.states", "状态") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.labels?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.labels", "标签") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.modules?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.modules", "模块") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.members?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.members", "成员") }}
              </span>
            </div>
          </div>
        </div>

        <pre v-if="exportYaml" class="export__yaml">{{ exportYaml }}</pre>
      </div>
    </AppModal>
  
    </div>
  </AppPageContainer>


    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <!-- 空状态 -->
    <AppEmptyState
      v-else-if="configs.length === 0"
      :title="t('projectCode.empty.title', '还没有保存的项目配置')"
      :description="
        t('projectCode.empty.description', '点击「导入 YAML」编写第一份项目定义，或从既有项目反向导出。')
      "
    >
      <div class="pac__empty-actions">
        <AppButton variant="primary" size="sm" @click="openCreateEditor">
          {{ t("projectCode.header.import", "导入 YAML") }}
        </AppButton>
        <AppButton variant="secondary" size="sm" @click="showExport = true">
          {{ t("projectCode.header.export", "导出项目") }}
        </AppButton>
      </div>
    </AppEmptyState>

    <!-- 已保存配置列表 -->
    <section v-else class="pac__list">
      <h2 class="pac__section-title">
        {{ t("projectCode.configs.title", "已保存配置") }}
        <span class="pac__count">{{ configs.length }}</span>
      </h2>

      <div class="pac__table-wrap">
        <table class="pac__table">
          <thead>
            <tr>
              <th>{{ t("projectCode.configs.name", "名称") }}</th>
              <th>{{ t("projectCode.configs.version", "版本") }}</th>
              <th>{{ t("projectCode.configs.sha", "SHA") }}</th>
              <th>{{ t("projectCode.configs.updatedAt", "更新时间") }}</th>
              <th>{{ t("projectCode.configs.project", "关联项目") }}</th>
              <th class="pac__th-actions">{{ t("common.actions", "操作") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="cfg in configs" :key="cfg.id">
              <td>
                <div class="pac__name">
                  <span class="pac__name-text">{{ cfg.name || cfg.spec?.name || "—" }}</span>
                  <AppBadge v-if="cfg.spec_key" variant="brand">
                    {{ cfg.spec_key }}
                  </AppBadge>
                </div>
                <div
                  v-if="cfg.spec"
                  class="pac__meta pac__meta--dim"
                >
                  {{ cfg.spec.states?.length ?? 0 }} {{ t("projectCode.summary.states", "状态") }}
                  · {{ cfg.spec.labels?.length ?? 0 }} {{ t("projectCode.summary.labels", "标签") }}
                  · {{ cfg.spec.modules?.length ?? 0 }} {{ t("projectCode.summary.modules", "模块") }}
                </div>
              </td>
              <td>
                <AppBadge variant="default">v{{ cfg.version || "—" }}</AppBadge>
              </td>
              <td>
                <code class="pac__mono">{{ shortSha(cfg.sha256) }}</code>
              </td>
              <td class="pac__meta">{{ formatTime(cfg.updated_at) }}</td>
              <td>
                <button
                  v-if="cfg.project_id && projectMap.has(cfg.project_id)"
                  class="pac__link"
                  @click="toProject(cfg.project_id)"
                >
                  {{ projectMap.get(cfg.project_id)?.identifier }}
                </button>
                <span v-else class="pac__meta pac__meta--dim">—</span>
              </td>
              <td class="pac__td-actions">
                <button class="pac__link" @click="openEditEditor(cfg)">
                  {{ t("common.edit", "编辑") }}
                </button>
                <button
                  class="pac__link pac__link--danger"
                  @click="removeConfig(cfg)"
                >
                  {{ t("common.delete", "删除") }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- YAML 编辑器 Modal -->
    <AppModal
      :visible="showEditor"
      :title="
        editingId
          ? t('projectCode.editor.editTitle', '编辑项目 YAML')
          : t('projectCode.editor.createTitle', '导入项目 YAML')
      "
      width="760px"
      @close="closeEditor"
    >
      <div class="editor">
        <div class="editor__hint">
          <span class="editor__hint-label">
            {{ t("projectCode.editor.hint", "语法提示") }}
          </span>
          <code class="editor__hint-code">
            version / name / key / color / icon / description / states / labels /
            modules / members / settings
          </code>
        </div>

        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.editor.nameLabel", "配置名（可选）") }}
          </span>
          <input
            v-model="configName"
            class="editor__input"
            :placeholder="t('projectCode.editor.namePlaceholder', '留空则使用 spec.name')"
            maxlength="200"
          />
        </label>

        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.editor.yamlLabel", "YAML 内容") }}
            <span class="editor__required">*</span>
          </span>
          <textarea
            v-model="yamlText"
            class="editor__textarea"
            rows="18"
            spellcheck="false"
            :placeholder="SAMPLE_YAML"
          ></textarea>
        </label>

        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.editor.targetProject", "目标项目（对比 / 应用时选择）") }}
          </span>
          <select
            v-model="selectedProjectId"
            class="editor__input"
          >
            <option :value="null">
              {{
                t(
                  "projectCode.editor.targetProjectNew",
                  "不选择 — 应用时将创建新项目",
                )
              }}
            </option>
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.identifier }} · {{ p.name }}
            </option>
          </select>
        </label>

        <div v-if="editorError" class="editor__error">{{ editorError }}</div>

        <!-- 校验结果摘要 -->
        <div v-if="parseResult" class="editor__summary">
          <div class="editor__summary-title">
            {{ t("projectCode.summary.title", "解析结果") }}
          </div>
          <div class="editor__summary-grid">
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ parseResult.key || "—" }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.projectKey", "项目标识") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.states?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.states", "状态") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.labels?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.labels", "标签") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.modules?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.modules", "模块") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">
                {{ parseResult.members?.length ?? 0 }}
              </span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.members", "成员") }}
              </span>
            </div>
          </div>
        </div>

        <!-- 对比结果 -->
        <div v-if="showDiffPanel && diffResult" class="editor__diff">
          <div class="editor__diff-head">
            <span class="editor__summary-title">
              {{ t("projectCode.diff.title", "对比结果") }}
            </span>
            <div class="editor__diff-stats">
              <AppBadge variant="success">
                +{{ diffResult.add_count }}
              </AppBadge>
              <AppBadge variant="warning">
                ~{{ diffResult.update_count }}
              </AppBadge>
              <AppBadge variant="danger">
                -{{ diffResult.remove_count }}
              </AppBadge>
            </div>
          </div>

          <p
            v-if="!diffResult.has_changes"
            class="editor__diff-empty"
          >
            {{ t("projectCode.diff.noChanges", "规范与当前项目一致，无需变更") }}
          </p>

          <div
            v-for="group in diffGroups"
            :key="group.key"
            class="editor__diff-group"
          >
            <div class="editor__diff-group-title">{{ group.label }}</div>
            <ul class="editor__diff-list">
              <li v-for="(item, idx) in group.items" :key="idx">
                <AppBadge :variant="KIND_META[item.kind]?.variant ?? 'info'">
                  {{ KIND_META[item.kind]?.label ?? item.kind }}
                </AppBadge>
                <span>{{ describeChange(item) }}</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="editor__footer">
          <div class="editor__footer-left">
            <AppButton
              variant="ghost"
              size="sm"
              :loading="parsingBusy"
              @click="handleParse"
            >
              {{ t("projectCode.action.validate", "校验") }}
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              :loading="diffing"
              @click="handleDiff"
            >
              {{ t("projectCode.action.diff", "对比") }}
            </AppButton>
          </div>
          <div class="editor__footer-right">
            <AppButton variant="secondary" size="sm" @click="closeEditor">
              {{ t("common.cancel", "取消") }}
            </AppButton>
            <AppButton
              variant="secondary"
              size="sm"
              :loading="saving"
              @click="handleSave"
            >
              {{ t("projectCode.action.save", "保存配置") }}
            </AppButton>
            <AppButton
              variant="primary"
              size="sm"
              @click="openApplyConfirm"
            >
              {{ t("projectCode.action.apply", "应用") }}
            </AppButton>
          </div>
        </div>
      
    </AppModal>

    <!-- 应用确认 Modal -->
    <AppModal
      :visible="showApplyConfirm"
      :title="t('projectCode.apply.confirmTitle', '确认应用规范')"
      width="480px"
      @close="showApplyConfirm = false"
    >
      <p class="pac__confirm-text">
        {{
          selectedProjectId
            ? t(
                "projectCode.apply.confirmUpdate",
                "将把当前 YAML 规范应用到所选项目，可能新增 / 修改项目的状态、标签、模块与成员。是否继续？",
              )
            : t(
                "projectCode.apply.confirmCreate",
                "将依据当前 YAML 规范创建一个新项目，并写入其状态、标签、模块与成员。是否继续？",
              )
        }}
      </p>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showApplyConfirm = false">
          {{ t("common.cancel", "取消") }}
        </AppButton>
        <AppButton
          variant="danger"
          size="sm"
          :loading="applying"
          @click="handleApply"
        >
          {{ t("common.confirm", "确认") }}
        </AppButton>
      
    </AppModal>

    <!-- 导出 Modal -->
    <AppModal
      :visible="showExport"
      :title="t('projectCode.export.title', '导出项目为 YAML')"
      width="720px"
      @close="closeExport"
    >
      <div class="export">
        <label class="editor__field">
          <span class="editor__label">
            {{ t("projectCode.export.selectLabel", "选择项目") }}
            <span class="editor__required">*</span>
          </span>
          <select v-model="exportProjectId" class="editor__input">
            <option :value="null">
              {{ t("projectCode.export.selectPlaceholder", "请选择要导出的项目") }}
            </option>
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.identifier }} · {{ p.name }}
            </option>
          </select>
        </label>

        <div class="export__actions">
          <AppButton
            variant="primary"
            size="sm"
            :loading="exporting"
            @click="handleExport"
          >
            {{ t("projectCode.export.generate", "生成 YAML") }}
          </AppButton>
          <AppButton
            v-if="exportYaml"
            variant="secondary"
            size="sm"
            @click="copyExport"
          >
            {{ t("common.copy", "复制") }}
          </AppButton>
        </div>

        <div v-if="exportError" class="editor__error">{{ exportError }}</div>

        <div v-if="exportSpec" class="editor__summary">
          <div class="editor__summary-title">
            {{ t("projectCode.summary.title", "解析结果") }}
          </div>
          <div class="editor__summary-grid">
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.key || "—" }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.projectKey", "项目标识") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.states?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.states", "状态") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.labels?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.labels", "标签") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.modules?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.modules", "模块") }}
              </span>
            </div>
            <div class="editor__summary-item">
              <span class="editor__summary-value">{{ exportSpec.members?.length ?? 0 }}</span>
              <span class="editor__summary-label">
                {{ t("projectCode.summary.members", "成员") }}
              </span>
            </div>
          </div>
        </div>

        <pre v-if="exportYaml" class="export__yaml">{{ exportYaml }}</pre>
      </div>
    </AppModal>
  </div>


<style scoped>
.pac {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  
  
}

.pac__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.pac__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--txt-primary);
}

.pac__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  line-height: 1.5;
  color: var(--txt-secondary);
  max-width: 640px;
}

.pac__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.pac__empty-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

/* ---- 列表 ---- */
.pac__section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--txt-secondary);
}

.pac__count {
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-tertiary);
}

.pac__table-wrap {
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
  background: var(--bg-surface);
  overflow-x: auto;
}

.pac__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.pac__table th {
  text-align: left;
  padding: 10px 14px;
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-tertiary);
  background: var(--bg-layer-1);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.pac__table td {
  padding: 12px 14px;
  border-bottom: 1px solid var(--border-subtle);
  vertical-align: top;
}

.pac__table tbody tr:last-child td {
  border-bottom: none;
}

.pac__table tbody tr:hover {
  background: var(--bg-layer-1-hover);
}

.pac__th-actions,
.pac__td-actions {
  text-align: right;
  white-space: nowrap;
}

.pac__name {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.pac__name-text {
  font-weight: 500;
  color: var(--txt-primary);
}

.pac__meta {
  font-size: 12px;
  color: var(--txt-secondary);
}

.pac__meta--dim {
  color: var(--txt-tertiary);
  margin-top: 2px;
}

.pac__mono {
  font-family: var(--font-mono, monospace);
  font-size: 12px;
  color: var(--txt-secondary);
}

.pac__link {
  background: none;
  border: none;
  padding: 2px 0;
  font-size: 12px;
  color: var(--accent);
  cursor: pointer;
}

.pac__link:hover {
  text-decoration: underline;
}

.pac__link--danger {
  color: var(--danger);
  margin-left: 12px;
}

.pac__confirm-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--txt-secondary);
}

/* ---- 编辑器 ---- */
.editor {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.editor__hint {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border: 1px dashed var(--border-default);
  border-radius: 8px;
  background: var(--bg-layer-1);
}

.editor__hint-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--txt-tertiary);
}

.editor__hint-code {
  font-family: var(--font-mono, monospace);
  font-size: 12px;
  line-height: 1.5;
  color: var(--txt-secondary);
  word-break: break-all;
}

.editor__field {
  display: block;
}

.editor__label {
  display: block;
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-secondary);
}

.editor__required {
  color: var(--danger);
}

.editor__input {
  width: 100%;
  height: 34px;
  padding: 0 10px;
  font-size: 13px;
  color: var(--txt-primary);
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  border-radius: 8px;
}

.editor__input:focus {
  outline: none;
  border-color: var(--accent);
}

.editor__textarea {
  width: 100%;
  min-height: 320px;
  padding: 10px 12px;
  font-family: var(--font-mono, monospace);
  font-size: 12.5px;
  line-height: 1.55;
  color: var(--txt-primary);
  background: var(--bg-layer-1);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  resize: vertical;
  tab-size: 2;
}

.editor__textarea:focus {
  outline: none;
  border-color: var(--accent);
  background: var(--bg-surface);
}

.editor__error {
  font-size: 12px;
  color: var(--danger);
  word-break: break-all;
}

.editor__summary {
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
}

.editor__summary-title {
  margin-bottom: 10px;
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-secondary);
}

.editor__summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(90px, 1fr));
  gap: 8px;
}

.editor__summary-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  align-items: center;
  padding: 8px;
  border-radius: 6px;
  background: var(--bg-layer-1);
}

.editor__summary-value {
  font-size: 16px;
  font-weight: 600;
  color: var(--txt-primary);
}

.editor__summary-label {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.editor__diff {
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  padding: 12px;
}

.editor__diff-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.editor__diff-stats {
  display: flex;
  gap: 6px;
}

.editor__diff-empty {
  margin: 0;
  font-size: 13px;
  color: var(--txt-tertiary);
}

.editor__diff-group + .editor__diff-group {
  margin-top: 10px;
}

.editor__diff-group-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-secondary);
  margin-bottom: 4px;
}

.editor__diff-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.editor__diff-list li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 12.5px;
  line-height: 1.5;
  color: var(--txt-primary);
}

.editor__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  width: 100%;
}

.editor__footer-left,
.editor__footer-right {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* ---- 导出 ---- */
.export {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.export__actions {
  display: flex;
  gap: 8px;
}

.export__yaml {
  max-height: 420px;
  overflow: auto;
  margin: 0;
  padding: 12px;
  font-family: var(--font-mono, monospace);
  font-size: 12.5px;
  line-height: 1.55;
  color: var(--txt-primary);
  background: var(--bg-layer-1);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  white-space: pre;
}

@media (max-width: 640px) {
  .pac {
    padding: 16px;
  }

  .pac__header {
    flex-direction: column;
  }

  .editor__footer {
    flex-direction: column;
    align-items: stretch;
  }

  .editor__footer-left,
  .editor__footer-right {
    justify-content: flex-end;
  }
}
</style>
