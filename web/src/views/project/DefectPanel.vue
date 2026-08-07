<script setup lang="ts">
/**
 * 缺陷面板 — 展示版本的缺陷列表与修复状态。
 */

import { onMounted, ref, computed } from "vue";

import { versionApi, type BugVersionView } from "@/api/services/version";
import { analyticsApi } from "@/api/services/analytics";
import { AppBadge, AppLoadingState, AppErrorState, AppEmptyState } from "@/components";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  versionId: number;
}>();

const defects = ref<BugVersionView[]>([]);
const total = ref(0);
const loading = ref(true);
const error = ref("");

const severityLabel: Record<number, string> = {
  0: "致命",
  1: "严重",
  2: "一般",
  3: "轻微",
  4: "建议",
};

const severityVariant: Record<number, "danger" | "warning" | "info" | "default" | "brand"> = {
  0: "danger",
  1: "danger",
  2: "warning",
  3: "info",
  4: "default",
};

const stateGroupLabel: Record<string, string> = {
  backlog: "待办",
  unstarted: "未开始",
  started: "进行中",
  completed: "已完成",
  cancelled: "已取消",
};

const stateGroupVariant: Record<string, "default" | "warning" | "success" | "danger"> = {
  backlog: "default",
  unstarted: "default",
  started: "warning",
  completed: "success",
  cancelled: "danger",
};

/* severity summary */
const summary = computed(() => {
  const map: Record<number, number> = {};
  defects.value.forEach((d) => {
    const s = d.severity ?? 4;
    map[s] = (map[s] ?? 0) + 1;
  });
  return map;
});

let wsIdVal = 0;
async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.get(props.workspaceId);
  wsIdVal = ws.id;
  return wsIdVal;
}

/* 导出（CSV / xlsx）— 按当前版本过滤 */
const showExportDropdown = ref(false);

function openExport(format: string) {
  if (!wsIdVal) return;
  window.open(
    analyticsApi.exportUrl(wsIdVal, props.projectId, format, { version_id: props.versionId }),
    "_blank",
  );
  showExportDropdown.value = false;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const res = await versionApi.getDefectPanel(wsId, props.projectId, props.versionId);
    defects.value = res.results;
    total.value = res.total;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载缺陷面板失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="defect-panel">
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState v-else-if="defects.length === 0" icon="🐛" title="该版本暂无关联缺陷" description="版本发布后将自动关联缺陷" />
    <template v-else>
      <!-- Summary -->
      <div class="defect-panel__summary">
        <span class="defect-panel__summary-text">
          共 {{ total }} 个缺陷
        </span>
        <span
          v-for="(count, sev) in summary"
          :key="sev"
          class="defect-panel__summary-count"
        >
          {{ severityLabel[Number(sev)] ?? sev }}: {{ count }}
        </span>
        <div class="defect-panel__export">
          <div class="defect-panel__export-wrap" @mouseleave="showExportDropdown = false">
            <button
              class="defect-panel__export-btn"
              @mouseenter="showExportDropdown = true"
            >
              导出
            </button>
            <div v-if="showExportDropdown" class="defect-panel__export-menu">
              <a
                class="defect-panel__export-item"
                href="#"
                @click.prevent="openExport('csv')"
              >
                导出 CSV
              </a>
              <a
                class="defect-panel__export-item"
                href="#"
                @click.prevent="openExport('xlsx')"
              >
                导出 Excel (.xlsx)
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="defect-panel__table-wrap">
        <table class="defect-panel__table">
          <thead>
            <tr>
              <th>标识</th>
              <th>标题</th>
              <th>严重程度</th>
              <th>状态</th>
              <th>发现版本</th>
              <th>修复版本</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="d in defects"
              :key="d.issue_id"
              class="defect-panel__row"
              :class="{
                'defect-panel__row--critical': d.severity != null && d.severity <= 1,
              }"
            >
              <td>
                <code class="defect-panel__id">{{ d.identifier }}</code>
              </td>
              <td>
                <span class="defect-panel__name">{{ d.name }}</span>
              </td>
              <td>
                <AppBadge
                  v-if="d.severity != null"
                  :variant="severityVariant[d.severity] ?? 'default'"
                >
                  {{ severityLabel[d.severity] ?? d.severity }}
                </AppBadge>
                <span v-else class="defect-panel__na">-</span>
              </td>
              <td>
                <AppBadge
                  v-if="d.state_group"
                  :variant="stateGroupVariant[d.state_group] ?? 'default'"
                >
                  {{ stateGroupLabel[d.state_group] ?? d.state_name }}
                </AppBadge>
                <span v-else>{{ d.state_name }}</span>
              </td>
              <td>
                <code class="defect-panel__ver">{{ d.found_version ?? '-' }}</code>
              </td>
              <td>
                <code class="defect-panel__ver">{{ d.fix_version ?? '-' }}</code>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<style scoped>
.defect-panel {
  font-size: 13px;
}

.defect-panel__error {
  text-align: center;
  padding: 24px;
  color: var(--danger-500);
}
.defect-panel__error p { margin: 0 0 8px; }
.defect-panel__retry-btn {
  font-size: 12px;
  padding: 4px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  cursor: pointer;
}

.defect-panel__loading {
  text-align: center;
  padding: 24px;
  color: var(--text-tertiary);
}

.defect-panel__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
}
.defect-panel__summary-text {
  font-weight: 500;
  color: var(--text-secondary);
  margin-right: 8px;
}
.defect-panel__summary-count {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}
.defect-panel__export {
  margin-left: auto;
}
.defect-panel__export-wrap {
  position: relative;
}
.defect-panel__export-btn {
  padding: 4px 12px;
  font-size: 12px;
  font-family: inherit;
  border: 1px solid var(--brand-500);
  border-radius: var(--radius-sm);
  background: var(--brand-500);
  color: var(--text-on-brand);
  cursor: pointer;
  transition: background 0.15s;
}
.defect-panel__export-btn:hover {
  background: var(--brand-600);
}
.defect-panel__export-menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 150px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-popover);
  z-index: 50;
  overflow: hidden;
}
.defect-panel__export-item {
  display: block;
  padding: 8px 14px;
  font-size: 13px;
  color: var(--text-primary);
  text-decoration: none;
  transition: background 0.1s;
}
.defect-panel__export-item:hover {
  background: var(--surface-2);
}
.defect-panel__export-item + .defect-panel__export-item {
  border-top: 1px solid var(--border-subtle);
}

.defect-panel__empty {
  text-align: center;
  padding: 24px;
  color: var(--text-tertiary);
}

.defect-panel__table-wrap {
  overflow-x: auto;
}
.defect-panel__table {
  width: 100%;
  border-collapse: collapse;
}
.defect-panel__table th {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-align: left;
  padding: 6px 10px;
  border-bottom: 2px solid var(--border-subtle);
  white-space: nowrap;
}
.defect-panel__table td {
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-subtle);
  vertical-align: middle;
}
.defect-panel__row:hover {
  background: var(--surface-2);
}
.defect-panel__row--critical {
  border-left: 3px solid var(--danger-500);
}
.defect-panel__id {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
  background: var(--surface-2);
  padding: 1px 6px;
  border-radius: 3px;
}
.defect-panel__name {
  max-width: 240px;
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.defect-panel__ver {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
}
.defect-panel__na {
  color: var(--text-placeholder);
}
</style>
