<script setup lang="ts">
/**
 * ImportDialog — CSV / XLSX 工作项导入弹窗。
 *
 * 步骤:
 *   1. 选择文件 (input[type=file])  →  自动解析 header 预览列
 *   2. 为每一列选择目标 Plane 字段（下拉，仅展示白名单字段）
 *   3. 勾选「增量导入」（启用时强制 external_id 列必须被映射）
 *   4. 提交 → 调用 issueApi.importIssues → 显示结果
 */
import { computed, ref, watch } from "vue";

import {
  type ImportColumnMapping,
  type ImportResult,
  IMPORT_FIELD_OPTIONS,
  issueApi,
} from "@/api/services/issue";
import { toast } from "@/lib/toast";

const props = defineProps<{
  visible: boolean;
  wsId: number;
  projectId: number;
}>();

const emit = defineEmits<{
  close: [];
  imported: [result: ImportResult];
}>();

// ---------- 步骤状态 ----------
const STEP = { SELECT: 1, MAPPING: 2, RESULT: 3 } as const;
const step = ref<number>(STEP.SELECT);

// 文件 / 解析
const file = ref<File | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const headers = ref<string[]>([]);          // 原始 CSV 表头
const previewRows = ref<string[][]>([]);    // 前 3 行数据（预览）
const parseError = ref("");
const parsing = ref(false);

// 映射
const mappings = ref<MappingRow[]>([]);

// 增量导入
const incremental = ref(false);

// 提交
const submitting = ref(false);
const result = ref<ImportResult | null>(null);

interface MappingRow {
  column_name: string;
  field: string; // 空字符串 = 不导入
}

/** 是否已将某一列映射到 external_id */
const hasExternalIdMapping = computed(() =>
  mappings.value.some((m) => m.field === "external_id"),
);

/** 表单校验是否通过мож */
const canSubmit = computed(() => {
  if (!file.value) return false;
  // 至少 name 列被映射
  const hasName = mappings.value.some((m) => m.field === "name");
  if (!hasName) return false;
  // 增量导入需要 external_id
  if (incremental.value && !hasExternalIdMapping.value) return false;
  return true;
});

/** 文件选择 → 解析 header + 前几行（CSV 本地解析，XLSX 走后端预览） */
async function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const f = input.files?.[0];
  if (!f) return;
  file.value = f;
  parseError.value = "";
  parsing.value = true;

  if (f.name.toLowerCase().endsWith(".xlsx")) {
    try {
      const r = await issueApi.previewImport(props.wsId, props.projectId, f);
      applyHeaders(r.headers ?? [], r.preview_rows ?? []);
    } catch (err) {
      parseError.value = err instanceof Error ? err.message : "文件解析失败";
      parsing.value = false;
    }
    return;
  }

  parseCSV(f);
}

/** 用后端/本地解析出的 header 与预览行填充映射步骤 */
function applyHeaders(headerList: string[], rows: string[][]) {
  headers.value = headerList;
  previewRows.value = rows;
  // 默认映射：表头与 IMPORT_FIELD_LABELS 匹配的自动选上
  mappings.value = headerList.map((h) => ({
    column_name: h,
    field: guessField(h),
  }));
  step.value = STEP.MAPPING;
  parsing.value = false;
}

/** 用浏览器原生 CSV 解析（简化版，不做 full RFC 4180） */
function parseCSV(f: File) {
  const reader = new FileReader();
  reader.onload = () => {
    const text = reader.result as string;
    const lines = splitCSLines(text);
    if (lines.length === 0) {
      parseError.value = "文件为空";
      parsing.value = false;
      return;
    }
    const sep = detectSeparator(lines[0]);
    const headerList = splitLine(lines[0], sep);
    const rows = lines.slice(1, 4).map((l) => splitLine(l, sep));
    applyHeaders(headerList, rows);
  };
  reader.onerror = () => {
    parseError.value = "文件读取失败";
    parsing.value = false;
  };
  reader.readAsText(f, "UTF-8");
}

// 简化的 CSV 解析辅助
function splitCSLines(text: string): string[] {
  // 处理 BOM
  if (text.charCodeAt(0) === 0xFEFF) text = text.slice(1);
  return text.split(/\r?\n/).filter((l, i, arr) => !(i === arr.length - 1 && l === ""));
}
function detectSeparator(firstLine: string): "," | ";" | "\t" {
  const counts = [
    { sep: "," as const, n: (firstLine.match(/,/g) || []).length },
    { sep: ";" as const, n: (firstLine.match(/;/g) || []).length },
    { sep: "\t" as const, n: (firstLine.match(/\t/g) || []).length },
  ];
  return counts.sort((a, b) => b.n - a.n)[0].sep;
}
function splitLine(line: string, sep: "," | ";" | "\t"): string[] {
  // 简化：不处理引号内分隔符（CSV 引号解析较复杂，此处够用即可）
  return line.split(sep).map((s) => s.trim().replace(/^"|"$/g, ""));
}

/** 根据表头猜测目标字段 */
function guessField(header: string): string {
  const h = header.toLowerCase();
  if (/(名称|标题|题目|name)/.test(h)) return "name";
  if (/(描述|描述|description)/.test(h)) return "description";
  if (/(优先级|priority)/.test(h)) return "priority";
  if (/(严重|severity)/.test(h)) return "severity";
  if (/(类型|type)/.test(h)) return "type_code";
  if (/(状态|state)/.test(h)) return "state_name";
  if (/(指派|负责人|assign)/.test(h)) return "assignee_emails";
  if (/(模块|module)/.test(h)) return "module_names";
  if (/(标签|label)/.test(h)) return "label_names";
  if (/(外部|external|来源编号)/.test(h)) return "external_id";
  if (/(来源|source)/.test(h)) return "source";
  if (/(发现.*阶段|found.*phase)/.test(h)) return "found_phase";
  if (/(发现.*版本|found.*version)/.test(h)) return "found_version";
  if (/(修复.*版本|fix.*version)/.test(h)) return "fix_version";
  if (/(父.*|parent)/.test(h)) return "parent_identifier";
  if (/(根因|root.*cause)/.test(h)) return "root_cause_category";
  if (/(分类|category)/.test(h)) return "category";
  if (/(点数|point|故事点)/.test(h)) return "point";
  return "";
}

/** 切换增量导入联动 */
watch(incremental, (val) => {
  if (val) {
    // 自动把某一列表头匹配为 external_id（如无映射）
    if (!hasExternalIdMapping.value) {
      const candidate = headers.value.find((h) =>
        /(外部|external|ext[^a-z]?id)/i.test(h),
      );
      if (candidate) {
        const row = mappings.value.find((m) => m.column_name === candidate);
        if (row) row.field = "external_id";
      }
    }
  }
});

function triggerFileInput() {
  fileInput.value?.click();
}

function reset() {
  file.value = null;
  headers.value = [];
  previewRows.value = [];
  mappings.value = [];
  incremental.value = false;
  result.value = null;
  parseError.value = "";
  step.value = STEP.SELECT;
}

function close() {
  reset();
  emit("close");
}

async function submit() {
  if (!file.value || !canSubmit.value) return;
  submitting.value = true;
  result.value = null;

  const activeMappings: ImportColumnMapping[] = mappings.value
    .filter((m) => m.field !== "")
    .map((m) => ({ column_name: m.column_name, field: m.field }));

  try {
    const r = await issueApi.importIssues(
      props.wsId,
      props.projectId,
      file.value,
      activeMappings,
      incremental.value,
    );
    result.value = r;
    step.value = STEP.RESULT;
    if (r.failed > 0) {
      toast.warning(`导入完成：${r.succeeded} 成功，${r.failed} 失败`);
    } else {
      toast.success(`导入完成：共 ${r.succeeded} 项（新建 ${r.created}，更新 ${r.updated}）`);
    }
    emit("imported", r);
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : "导入失败";
    toast.error(msg);
  } finally {
    submitting.value = false;
  }
}

function selectFieldName(row: MappingRow, event: Event) {
  const target = event.target as HTMLSelectElement;
  row.field = target.value;
}
</script>

<template>
  <AppModal :visible="visible" title="导入工作项" width="720px" @close="close">
    <!-- Step 1: 选文件 -->
    <div v-if="step === STEP.SELECT" class="import-step">
      <p class="import-step__hint">
        支持 CSV 或 XLSX 文件。第一行应包含表头（字段名）。
      </p>
      <input
        ref="fileInput"
        type="file"
        accept=".csv,.xlsx"
        style="display: none"
        @change="onFileChange"
      />
      <button
        class="btn btn--primary import-upload-btn"
        :disabled="parsing"
        @click="triggerFileInput"
      >
        {{ parsing ? "解析中..." : (file ? file.name || "重新选择文件" : "选择文件") }}
      </button>
      <p v-if="parseError" class="import-error">{{ parseError }}</p>
      <p v-if="file && !parseError" class="import-info">
        已选择：{{ file.name }}
      </p>
    </div>

    <!-- Step 2: 列映射 -->
    <div v-else-if="step === STEP.MAPPING" class="import-step">
      <div class="import-preview">
        <strong>预览</strong>
        <span class="import-preview__name">{{ file?.name }}</span>
        <button class="btn btn--ghost btn--sm" @click="reset">重新选择</button>
      </div>

      <p class="import-step__hint">
        为每一列选择目标 Plane 字段。标 * 的字段为必填。
      </p>

      <table class="import-mapping-table">
        <thead>
          <tr>
            <th class="col-idx">#</th>
            <th class="col-header">CSV 列头</th>
            <th class="col-sample">示例值</th>
            <th class="col-field">目标字段</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, i) in mappings" :key="i">
            <td class="col-idx">{{ i + 1 }}</td>
            <td class="col-header">{{ row.column_name }}</td>
            <td class="col-sample">{{ previewRows[0]?.[i] ?? "-" }}</td>
            <td class="col-field">
              <select
                :value="row.field"
                class="import-select"
                @change="selectFieldName(row, $event)"
              >
                <option value="">-- 不导入 --</option>
                <option
                  v-for="opt in IMPORT_FIELD_OPTIONS"
                  :key="opt.id"
                  :value="opt.id"
                >
                  {{ opt.label }}
                </option>
              </select>
            </td>
          </tr>
        </tbody>
      </table>

      <label class="import-incremental">
        <input v-model="incremental" type="checkbox" />
        <span>
          增量导入（按
          <code>external_id</code>
          识别已有工作项并更新，否则创建新项）
        </span>
      </label>

      <p v-if="incremental && !hasExternalIdMapping" class="import-warning">
        增量导入必须至少映射一列为
        <code>external_id</code>
        。
      </p>
      <p v-else-if="!mappings.some(m => m.field === 'name')" class="import-warning">
        至少需要映射一列为
        <code>name</code>
        （必填字段）。
      </p>

      <div class="import-step__actions">
        <button class="btn btn--ghost" @click="close">取消</button>
        <button
          class="btn btn--primary"
          :disabled="!canSubmit || submitting"
          @click="submit"
        >
          {{ submitting ? "导入中..." : "开始导入" }}
        </button>
      </div>
    </div>

    <!-- Step 3: 结果 -->
    <div v-else-if="step === STEP.RESULT && result" class="import-step">
      <div class="import-result">
        <div class="import-result__card">
          <span class="import-result__num">{{ result.total }}</span>
          <span class="import-result__label">总处理</span>
        </div>
        <div class="import-result__card import-result__card--ok">
          <span class="import-result__num">{{ result.created }}</span>
          <span class="import-result__label">新建</span>
        </div>
        <div class="import-result__card import-result__card--updated">
          <span class="import-result__num">{{ result.updated }}</span>
          <span class="import-result__label">更新</span>
        </div>
        <div class="import-result__card">
          <span class="import-result__num">{{ result.skipped }}</span>
          <span class="import-result__label">跳过</span>
        </div>
        <div class="import-result__card import-result__card--err">
          <span class="import-result__num">{{ result.failed }}</span>
          <span class="import-result__label">失败</span>
        </div>
      </div>

      <div v-if="result.errors.length > 0" class="import-errors">
        <strong>错误详情（前 50 条）：</strong>
        <ul>
          <li v-for="(e, i) in result.errors.slice(0, 50)" :key="i">
            第 {{ e.row }} 行 · {{ e.field }}：{{ e.message }}
          </li>
        </ul>
      </div>

      <div class="import-step__actions">
        <button class="btn btn--ghost" @click="reset">再次导入</button>
        <button class="btn btn--primary" @click="close">完成</button>
      </div>
    </div>
  </AppModal>
</template>

<style scoped>
.import-step { display: flex; flex-direction: column; gap: 12px; }
.import-step__hint { font-size: 12px; color: var(--text-tertiary); margin: 0; }
.import-step__actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 8px; }

.import-upload-btn { width: 100%; padding: 16px; border: 2px dashed var(--border-default); background: var(--surface-2); border-radius: var(--radius-sm); font-size: 13px; cursor: pointer; }
.import-upload-btn:hover:not(:disabled) { border-color: var(--brand-500); }
.import-upload-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.import-error { color: var(--danger-500); font-size: 12px; margin: 0; }
.import-info { font-size: 12px; color: var(--text-tertiary); margin: 0; }

.import-preview { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.import-preview__name { flex: 1; font-family: var(--font-mono); color: var(--text-primary); }

.import-mapping-table { width: 100%; border-collapse: collapse; font-size: 12px; margin: 4px 0; }
.import-mapping-table th { font-size: 11px; color: var(--text-tertiary); text-transform: uppercase; text-align: left; padding: 4px 6px; border-bottom: 1px solid var(--border-subtle); }
.import-mapping-table td { padding: 4px 6px; border-bottom: 1px solid var(--border-subtle); }
.import-mapping-table .col-idx { width: 32px; color: var(--text-tertiary); }
.import-mapping-table .col-header { width: 28%; }
.import-mapping-table .col-sample { width: 28%; color: var(--text-secondary); }
.import-mapping-table .col-field { width: auto; }

.import-select {
  width: 100%; padding: 3px 6px; font-size: 12px; font-family: inherit;
  background: var(--surface-1); border: 1px solid var(--border-default);
  border-radius: var(--radius-sm); color: var(--text-primary);
}

.import-incremental { display: flex; align-items: flex-start; gap: 6px; font-size: 13px; padding: 8px 0; }
.import-incremental input { margin-top: 2px; accent-color: var(--brand-500); }
.import-warning { font-size: 12px; color: var(--warning-500); margin: 0; }
.import-warning code, .import-incremental code { background: var(--surface-2); padding: 1px 4px; border-radius: 2px; }

.import-result { display: flex; flex-wrap: wrap; gap: 10px; }
.import-result__card {
  flex: 1; min-width: 72px; padding: 12px; background: var(--surface-2); border-radius: var(--radius-sm);
  display: flex; flex-direction: column; align-items: center; gap: 4px;
}
.import-result__card--ok { background: var(--success-50); }
.import-result__card--updated { background: var(--brand-50); }
.import-result__card--err { background: var(--danger-50); }
.import-result__num { font-size: 20px; font-weight: 700; }
.import-result__label { font-size: 11px; color: var(--text-tertiary); text-transform: uppercase; }

.import-errors { max-height: 160px; overflow-y: auto; font-size: 12px; }
.import-errors ul { list-style: none; padding: 0; margin: 4px 0 0; }
.import-errors li { padding: 2px 0; color: var(--text-secondary); border-bottom: 1px solid var(--border-subtle); }

.btn { cursor: pointer; font-family: inherit; border-radius: var(--radius-sm); }
.btn--primary { background: var(--brand-500); color: var(--text-on-brand); padding: 6px 14px; font-size: 13px; border: none; }
.btn--primary:hover:not(:disabled) { background: var(--brand-600); }
.btn--primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn--ghost { background: none; color: var(--text-secondary); padding: 6px 14px; font-size: 13px; border: 1px solid transparent; }
.btn--ghost:hover { background: var(--surface-3); }
.btn--sm { padding: 3px 8px; font-size: 11px; }
</style>
