<script setup lang="ts">
/**
 * GrantConditionBuilder — 条件表达式可视化构造器（递归组件）。
 *
 * 支持：
 *  - 逻辑分组：and（all）/ or（any）/ not（inner）
 *  - 比较节点：eq / ne / in / contains / gt / lt（field + value / values）
 *  - 嵌套任意层级；增删子条件、切换操作符。
 *
 * Props:
 *  - modelValue: 当前构造器根节点
 * Events:
 *  - update:modelValue: 节点变更
 */
import { computed } from "vue";
import type { GrantField, GrantOp } from "@/api/services/grant";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useI18n } from "@/composables/useI18n";

const { t } = useI18n();

/* ------------------------------------------------------------------ */
/* 构造器节点模型                                                      */
/* ------------------------------------------------------------------ */

export interface ConditionNode {
  key: string;
  op: GrantOp;
  field: GrantField | "";
  /** 比较值（字符串形式，序列化时按字段类型转换） */
  value: string;
  /** in 操作符的候选值（逗号分隔字符串） */
  valuesText: string;
  children: ConditionNode[];
  inner: ConditionNode | null;
}

const props = withDefaults(defineProps<{
  modelValue: ConditionNode;
  /** 是否为根节点（根节点不显示删除按钮） */
  root?: boolean;
}>(), { root: true });

const emit = defineEmits<{
  "update:modelValue": [node: ConditionNode];
}>();

/** 当前节点 */
const node = computed<ConditionNode>(() => props.modelValue);

/** 递归替换节点（不可变更新，保证 v-model 语义正确）。 */
function updateNode(root: ConditionNode, key: string, next: ConditionNode | null): ConditionNode {
  if (root.key === key) {
    return next ?? root;
  }
  if (root.inner) {
    const inner = updateNode(root.inner, key, next);
    if (inner !== root.inner) return { ...root, inner };
  }
  const idx = root.children.findIndex((c) => c.key === key);
  if (idx >= 0) {
    if (next === null) {
      const children = root.children.slice();
      children.splice(idx, 1);
      return { ...root, children };
    }
    const children = root.children.slice();
    children[idx] = next;
    return { ...root, children };
  }
  // 深度递归
  for (let i = 0; i < root.children.length; i++) {
    const updated = updateNode(root.children[i], key, next);
    if (updated !== root.children[i]) {
      const children = root.children.slice();
      children[i] = updated;
      return { ...root, children };
    }
  }
  return root;
}

function mutate(fn: (node: ConditionNode) => ConditionNode | null) {
  emit("update:modelValue", updateNode(props.modelValue, props.modelValue.key, fn(props.modelValue)));
}

/** 子节点变更（递归事件冒泡到根） */
function onChildUpdate(key: string, next: ConditionNode | null) {
  emit("update:modelValue", updateNode(props.modelValue, key, next));
}

/* ------------------------------------------------------------------ */
/* 常量                                                                */
/* ------------------------------------------------------------------ */

const FIELD_OPTIONS: { value: GrantField; label: string }[] = [
  { value: "projectId", label: "projectId" },
  { value: "type", label: "type" },
  { value: "stateGroup", label: "stateGroup" },
  { value: "priority", label: "priority" },
  { value: "creatorId", label: "creatorId" },
  { value: "assigneeId", label: "assigneeId" },
  { value: "labelId", label: "labelId" },
  { value: "workspaceId", label: "workspaceId" },
  { value: "moduleId", label: "moduleId" },
  { value: "sprintId", label: "sprintId" },
];

const OP_OPTIONS: { value: GrantOp; label: string }[] = [
  { value: "eq", label: "等于 (eq)" },
  { value: "ne", label: "不等于 (ne)" },
  { value: "in", label: "属于 (in)" },
  { value: "contains", label: "包含 (contains)" },
  { value: "gt", label: "大于 (gt)" },
  { value: "lt", label: "小于 (lt)" },
  { value: "and", label: "且 (and)" },
  { value: "or", label: "或 (or)" },
  { value: "not", label: "非 (not)" },
];

function usesValues(op: GrantOp): boolean {
  return op === "in";
}

function isLogical(op: GrantOp): boolean {
  return op === "and" || op === "or" || op === "not";
}

/* ------------------------------------------------------------------ */
/* 节点操作                                                            */
/* ------------------------------------------------------------------ */

let nodeSeq = 0;
function nextKey(): string {
  nodeSeq += 1;
  return `n${Date.now()}_${nodeSeq}`;
}

function createComparisonNode(): ConditionNode {
  return {
    key: nextKey(),
    op: "eq",
    field: "projectId",
    value: "",
    valuesText: "",
    children: [],
    inner: null,
  };
}

function createLogicalNode(op: "and" | "or" | "not"): ConditionNode {
  const node: ConditionNode = {
    key: nextKey(),
    op,
    field: "",
    value: "",
    valuesText: "",
    children: [],
    inner: null,
  };
  if (op === "not") {
    node.inner = createComparisonNode();
  } else {
    node.children = [createComparisonNode()];
  }
  return node;
}

function changeOp(op: GrantOp) {
  const cur = node.value;
  mutate(() => {
    if (op === "not") {
      return { ...cur, op, inner: createComparisonNode(), children: [] };
    }
    if (op === "and" || op === "or") {
      return { ...cur, op, children: [createComparisonNode()], inner: null };
    }
    return {
      ...cur,
      op,
      field: cur.field || "projectId",
      children: [],
      inner: null,
    };
  });
}

function setField(field: GrantField) {
  mutate(() => ({ ...node.value, field }));
}

function setValue(value: string) {
  mutate(() => ({ ...node.value, value }));
}

function setValuesText(valuesText: string) {
  mutate(() => ({ ...node.value, valuesText }));
}

function addComparison() {
  mutate(() => ({ ...node.value, children: [...node.value.children, createComparisonNode()] }));
}

function addGroup(op: "and" | "or") {
  mutate(() => ({
    ...node.value,
    children: [...node.value.children, createLogicalNode(op)],
  }));
}

function removeNode() {
  mutate(() => null);
}

/* ------------------------------------------------------------------ */
/* 与后端 Expr 互转                                                    */
/* ------------------------------------------------------------------ */

/** 逗号 / 中文逗号分隔的 values 文本 */
function splitValues(text: string): string[] {
  return text
    .split(/[,，]/)
    .map((v) => v.trim())
    .filter((v) => v.length > 0);
}

/** 按字段类型转换值（数字字段转 number，其余保持字符串） */
function coerceValue(field: GrantField, raw: string): unknown {
  const numericFields: GrantField[] = [
    "projectId",
    "priority",
    "creatorId",
    "assigneeId",
    "labelId",
    "workspaceId",
    "moduleId",
    "sprintId",
  ];
  if (numericFields.includes(field)) {
    const n = Number(raw);
    return Number.isNaN(n) ? raw : n;
  }
  return raw;
}

/** 将构造器节点序列化为后端 Expr；不完整时返回 null */
function nodeToExpr(n: ConditionNode): import("@/api/services/grant").GrantExpr | null {
  if (isLogical(n.op)) {
    if (n.op === "not") {
      if (!n.inner) return null;
      const inner = nodeToExpr(n.inner);
      if (!inner) return null;
      return { op: "not", inner };
    }
    if (n.children.length === 0) return null;
    const subs = n.children.map(nodeToExpr);
    if (subs.some((s) => s === null)) return null;
    if (n.op === "and") return { op: "and", all: subs as import("@/api/services/grant").GrantExpr[] };
    return { op: "or", any: subs as import("@/api/services/grant").GrantExpr[] };
  }
  if (!n.field) return null;
  if (usesValues(n.op)) {
    const values = splitValues(n.valuesText);
    if (values.length === 0) return null;
    return { op: n.op, field: n.field, values };
  }
  if (n.value.trim() === "") return null;
  return { op: n.op, field: n.field, value: coerceValue(n.field, n.value) };
}

/** 将后端 Expr 反序列化为构造器节点 */
function exprToNode(expr: import("@/api/services/grant").GrantExpr | null | undefined): ConditionNode {
  if (!expr || typeof expr.op !== "string") return createComparisonNode();
  if (expr.op === "and" || expr.op === "or") {
    const n = createLogicalNode(expr.op);
    const subs = expr.op === "and" ? expr.all : expr.any;
    n.children = (subs ?? []).map(exprToNode);
    if (n.children.length === 0) n.children = [createComparisonNode()];
    return n;
  }
  if (expr.op === "not") {
    const n = createLogicalNode("not");
    n.inner = exprToNode(expr.inner ?? null);
    return n;
  }
  const n = createComparisonNode();
  n.op = expr.op;
  if (expr.field) n.field = expr.field;
  if (expr.op === "in") {
    n.valuesText = (expr.values ?? []).map((v) => String(v)).join(", ");
  } else if (expr.value !== undefined && expr.value !== null) {
    n.value = String(expr.value);
  }
  return n;
}

/** 判断构造器是否可序列化 */
function nodeIsValid(node: ConditionNode): boolean {
  return nodeToExpr(node) !== null;
}

defineExpose({ nodeToExpr, exprToNode, nodeIsValid, createComparisonNode, createLogicalNode });
</script>

<template>
  <div class="grant-cb">
    <!-- 节点头：操作符 + 删除 -->
    <div class="grant-cb__head">
      <select
        class="grant-cb__select"
        :value="node.op"
        @change="changeOp(($event.target as HTMLSelectElement).value as GrantOp)"
      >
        <option v-for="o in OP_OPTIONS" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>
      <button
        v-if="!$slots.root"
        type="button"
        class="grant-cb__remove"
        @click="removeNode()"
      >
        {{ t("grantRule.remove") }}
      </button>
    </div>

    <!-- 比较节点 -->
    <div v-if="!isLogical(node.op)" class="grant-cb__row">
      <select
        class="grant-cb__select"
        :value="node.field"
        @change="setField(($event.target as HTMLSelectElement).value as GrantField)"
      >
        <option v-for="f in FIELD_OPTIONS" :key="f.value" :value="f.value">
          {{ f.label }}
        </option>
      </select>
      <input
        v-if="usesValues(node.op)"
        class="grant-cb__input"
        type="text"
        :value="node.valuesText"
        :placeholder="t('grantRule.valuesPlaceholder')"
        @input="setValuesText(($event.target as HTMLInputElement).value)"
      />
      <input
        v-else
        class="grant-cb__input"
        type="text"
        :value="node.value"
        :placeholder="t('grantRule.valuePlaceholder')"
        @input="setValue(($event.target as HTMLInputElement).value)"
      />
    </div>

    <!-- not 节点 -->
    <div v-else-if="node.op === 'not'" class="grant-cb__children">
      <GrantConditionBuilder
        v-if="node.inner"
        :model-value="node.inner"
        @update:model-value="onChildUpdate(node.key, $event)"
      />
    </div>

    <!-- and / or 节点 -->
    <div v-else class="grant-cb__children">
      <GrantConditionBuilder
        v-for="child in node.children"
        :key="child.key"
        :model-value="child"
        @update:model-value="onChildUpdate(child.key, $event)"
      />
      <div class="grant-cb__actions">
        <button type="button" class="grant-cb__add" @click="addComparison()">
          <DynamicIcon name="Plus" :size="12" />
          {{ t("grantRule.addCondition") }}
        </button>
        <button type="button" class="grant-cb__add" @click="addGroup('and')">
          <DynamicIcon name="Plus" :size="12" />
          {{ t("grantRule.addAndGroup") }}
        </button>
        <button type="button" class="grant-cb__add" @click="addGroup('or')">
          <DynamicIcon name="Plus" :size="12" />
          {{ t("grantRule.addOrGroup") }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.grant-cb {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.grant-cb__head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.grant-cb__select,
.grant-cb__input {
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
  background: var(--surface-1);
  padding: 4px 8px;
  font-size: 12px;
  color: var(--text-primary);
  outline: none;
}

.grant-cb__select:focus,
.grant-cb__input:focus {
  border-color: var(--brand-500);
}

.grant-cb__input {
  flex: 1;
  min-width: 0;
  font-family: inherit;
}

.grant-cb__row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.grant-cb__remove {
  margin-left: auto;
  font-size: 12px;
  color: var(--danger-500);
  background: none;
  border: none;
  cursor: pointer;
}

.grant-cb__remove:hover {
  text-decoration: underline;
}

.grant-cb__children {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-left: 12px;
  border-left: 2px solid var(--border-default);
}

.grant-cb__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding-top: 4px;
}

.grant-cb__add {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-sm);
  padding: 3px 8px;
  font-size: 12px;
  color: var(--text-secondary);
  background: transparent;
  cursor: pointer;
  font-family: inherit;
}

.grant-cb__add:hover {
  background: var(--surface-2);
}
</style>
