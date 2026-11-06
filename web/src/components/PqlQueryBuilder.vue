<script setup lang="ts">
/**
 * PqlQueryBuilder — PQL 查询构建器。
 *
 * 双模式：
 *  - 可视化：字段 + 操作符 + 值 的行组合，AND/OR 逻辑；
 *  - 文本：直接编写 PQL，实时校验并显示生成的 SQL。
 *
 * 输出：通过 update:modelValue 向上抛出 PQL 字符串。
 */
import { onMounted, ref, watch } from "vue";

import { pqlApi, type PQLField, type PQLParseResult } from "@/api/services/pql";

const props = withDefaults(defineProps<{
  modelValue: string;
  workspaceId: number;
  /** 紧凑模式（用于嵌入筛选栏） */
  compact?: boolean;
}>(), {
  compact: false,
});

const emit = defineEmits<{
  (e: "update:modelValue", v: string): void;
  (e: "change", v: string): void;
}>();

const mode = ref<"visual" | "text">("text");
const raw = ref(props.modelValue);
const fields = ref<PQLField[]>([]);
const result = ref<PQLParseResult | null>(null);
const validating = ref(false);
const showHelp = ref(false);

/* ---- 可视化条件 ---- */

interface Condition {
  id: number;
  field: string;
  op: string;
  value: string;
}
interface Group {
  id: number;
  op: "AND" | "OR";
  conditions: Condition[];
}

const groups = ref<Group[]>([]);
let nextId = 1;

const opOptions: Record<string, string[]> = {
  text: ["=", "!=", "~"],
  number: ["=", "!=", ">", ">=", "<", "<="],
  date: ["=", "!=", ">", ">=", "<", "<="],
  enum: ["=", "!=", "in"],
  user: ["=", "!="],
};

const opLabels: Record<string, string> = {
  "=": "等于",
  "!=": "不等于",
  "~": "包含",
  ">": "大于",
  ">=": "大于等于",
  "<": "小于",
  "<=": "小于等于",
  in: "属于",
};

function addGroup() {
  groups.value.push({
    id: nextId++,
    op: groups.value.length === 0 ? "AND" : "AND",
    conditions: [{ id: nextId++, field: "", op: "=", value: "" }],
  });
}

function removeGroup(gid: number) {
  groups.value = groups.value.filter((g) => g.id !== gid);
}

function addCondition(g: Group) {
  g.conditions.push({ id: nextId++, field: "", op: "=", value: "" });
}

function removeCondition(g: Group, cid: number) {
  g.conditions = g.conditions.filter((c) => c.id !== cid);
  if (!g.conditions.length) removeGroup(g.id);
}

function fieldOf(name: string): PQLField | undefined {
  return fields.value.find((f) => f.name === name);
}

function onFieldChange(c: Condition) {
  const f = fieldOf(c.field);
  const ops = f ? opOptions[f.type] ?? ["="] : ["="];
  if (!ops.includes(c.op)) c.op = ops[0];
}

/** 可视化 → PQL 文本 */
function visualToPql(): string {
  const parts: string[] = [];
  groups.value.forEach((g, gi) => {
    const conds = g.conditions
      .filter((c) => c.field && c.value)
      .map((c) => {
        const f = fieldOf(c.field);
        const isUser = f?.type === "user";
        const val = isUser && c.value.toLowerCase() === "me" ? "me" : `"${c.value}"`;
        return `${c.field} ${c.op} ${val}`;
      });
    if (!conds.length) return;
    const joined = conds.join(` ${g.op} `);
    parts.push(gi === 0 ? joined : `${g.op} (${joined})`);
  });
  return parts.join(" ");
}

watch(groups, () => {
  const pql = visualToPql();
  raw.value = pql;
  emit("update:modelValue", pql);
  emit("change", pql);
}, { deep: true });

watch(() => props.modelValue, (v) => {
  if (v !== raw.value) {
    raw.value = v;
    void validate(v);
  }
});

/* ---- 校验 ---- */

async function validate(q?: string) {
  const query = (q ?? raw.value).trim();
  if (!query) {
    result.value = null;
    return;
  }
  validating.value = true;
  try {
    result.value = await pqlApi.validate(props.workspaceId, query);
  } catch {
    result.value = { valid: false, error: "校验失败", limit: 100, offset: 0 };
  } finally {
    validating.value = false;
  }
}

let validateTimer: ReturnType<typeof setTimeout> | null = null;
function onRawInput() {
  emit("update:modelValue", raw.value);
  emit("change", raw.value);
  if (validateTimer) clearTimeout(validateTimer);
  validateTimer = setTimeout(() => validate(), 400);
}

onMounted(async () => {
  try {
    const res = await pqlApi.listFields(props.workspaceId);
    fields.value = (res as { fields?: PQLField[] }).fields ?? [];
  } catch {
    fields.value = [];
  }
  if (raw.value) void validate(raw.value);
});

const examples = [
  'state = "in_progress" AND priority = "high"',
  'assignee = me AND target_date <= "2026-12-31"',
  'type = "defect" AND severity >= 4',
  'title ~ "登录" OR description ~ "鉴权"',
  'label in ("backend", "urgent") AND NOT state_group = "cancelled"',
  'state_group = "started" sort:updated_at desc limit:20',
];

</script>

<template>
  <div class="pql" :class="{ 'pql--compact': compact }">
    <!-- 模式切换 -->
    <div class="pql__modes">
      <button
        class="pql__mode"
        :class="{ 'pql__mode--active': mode === 'text' }"
        @click="mode = 'text'"
      >
        文本
      </button>
      <button
        class="pql__mode"
        :class="{ 'pql__mode--active': mode === 'visual' }"
        @click="mode = 'visual'"
      >
        可视化
      </button>
      <button class="pql__help" @click="showHelp = !showHelp">
        {{ showHelp ? "收起示例" : "示例" }}
      </button>
    </div>

    <!-- 示例 -->
    <div v-if="showHelp" class="pql__examples">
      <button
        v-for="ex in examples"
        :key="ex"
        class="pql__example"
        @click="raw = ex; onRawInput()"
      >
        {{ ex }}
      </button>
    </div>

    <!-- 文本模式 -->
    <div v-if="mode === 'text'" class="pql__text">
      <textarea
        v-model="raw"
        class="pql__textarea"
        :rows="compact ? 2 : 3"
        placeholder='例如：state = "in_progress" AND priority = "high"'
        @input="onRawInput"
      />
      <div v-if="validating" class="pql__status pql__status--pending">校验中...</div>
      <div
        v-else-if="result"
        class="pql__status"
        :class="result.valid ? 'pql__status--ok' : 'pql__status--err'"
      >
        {{ result.valid ? "✓ 语法正确" : `✗ ${result.error}` }}
      </div>
      <div v-if="result?.valid && result.where_sql" class="pql__sql">
        <code>{{ result.where_sql }}</code>
        <span v-if="result.args?.length" class="pql__args">
          参数: {{ JSON.stringify(result.args) }}
        </span>
      </div>
    </div>

    <!-- 可视化模式 -->
    <div v-else class="pql__visual">
      <div v-if="!groups.length" class="pql__empty">
        尚未添加条件
        <button class="pql__link" @click="addGroup">添加条件组</button>
      </div>

      <div v-for="(g, gi) in groups" :key="g.id" class="pql__group">
        <div class="pql__group-head">
          <select v-if="gi > 0" v-model="g.op" class="pql__select pql__select--op">
            <option value="AND">且</option>
            <option value="OR">或</option>
          </select>
          <span v-else class="pql__group-label">条件组</span>
          <button class="pql__link pql__link--danger" @click="removeGroup(g.id)">
            删除组
          </button>
        </div>

        <div v-for="c in g.conditions" :key="c.id" class="pql__cond">
          <select v-model="c.field" class="pql__select" @change="onFieldChange(c)">
            <option value="" disabled>字段</option>
            <option v-for="f in fields" :key="f.name" :value="f.name">{{ f.name }}</option>
          </select>

          <select v-model="c.op" class="pql__select pql__select--op">
            <option
              v-for="op in opOptions[fieldOf(c.field)?.type ?? 'text']"
              :key="op"
              :value="op"
            >
              {{ opLabels[op] ?? op }}
            </option>
          </select>

          <input
            v-if="fieldOf(c.field)?.type === 'user'"
            v-model="c.value"
            class="pql__input"
            placeholder="me 或 用户名"
          />
          <select
            v-else-if="fieldOf(c.field)?.type === 'enum' && fieldOf(c.field)?.values?.length"
            v-model="c.value"
            class="pql__select"
          >
            <option value="" disabled>值</option>
            <option v-for="v in fieldOf(c.field)!.values" :key="v" :value="v">
              {{ v }}
            </option>
          </select>
          <input
            v-else
            v-model="c.value"
            class="pql__input"
            placeholder="值"
          />

          <button class="pql__link pql__link--danger" @click="removeCondition(g, c.id)">
            删除
          </button>
        </div>

        <button class="pql__link" @click="addCondition(g)">+ 添加条件</button>
      </div>

      <button v-if="groups.length" class="pql__link" @click="addGroup">
        + 添加条件组
      </button>
    </div>
  </div>
</template>

<style scoped>
.pql {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pql__modes {
  display: flex;
  gap: 6px;
  align-items: center;
}

.pql__mode {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--bg-surface);
  color: var(--txt-secondary);
  cursor: pointer;
}

.pql__mode--active {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--text-on-color);
}

.pql__help {
  margin-left: auto;
  font-size: 12px;
  color: var(--accent);
  background: none;
  border: none;
  cursor: pointer;
}

.pql__examples {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.pql__example {
  padding: 3px 8px;
  font-size: 11px;
  font-family: ui-monospace, monospace;
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  background: var(--bg-surface-2);
  color: var(--txt-secondary);
  cursor: pointer;
}

.pql__example:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.pql__textarea {
  width: 100%;
  padding: 8px 10px;
  font-family: ui-monospace, monospace;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
  resize: vertical;
}

.pql__status {
  font-size: 12px;
}

.pql__status--ok {
  color: var(--success);
}

.pql__status--err {
  color: var(--danger);
}

.pql__status--pending {
  color: var(--txt-tertiary);
}

.pql__sql {
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--bg-surface-2);
  font-size: 11px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pql__sql code {
  font-family: ui-monospace, monospace;
  color: var(--txt-secondary);
  word-break: break-all;
}

.pql__args {
  color: var(--txt-tertiary);
}

.pql__visual {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pql__empty {
  font-size: 13px;
  color: var(--txt-tertiary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.pql__group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface-2);
}

.pql__group-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pql__group-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-tertiary);
}

.pql__cond {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.pql__select,
.pql__input {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.pql__select {
  min-width: 110px;
}

.pql__select--op {
  min-width: 90px;
}

.pql__input {
  flex: 1;
  min-width: 120px;
}

.pql__link {
  font-size: 12px;
  color: var(--accent);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.pql__link--danger {
  color: var(--txt-tertiary);
}

.pql__link--danger:hover {
  color: var(--danger);
}
</style>
