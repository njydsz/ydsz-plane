<script setup lang="ts">
/**
 * PqlConsoleView — PQL 查询控制台。
 *
 * 功能：
 *  - 编写 / 校验 / 解释 PQL 查询；
 *  - 可视化条件构建器；
 *  - 字段与操作符参考；
 *  - 示例查询库；
 *  - 保存常用查询（localStorage）。
 */
import { computed, onMounted, ref } from "vue";

import { pqlApi, type PQLField, type PQLParseResult } from "@/api/services/pql";
import { AppPageContainer, PqlQueryBuilder } from "@/components";
import { useWorkspaceStore } from "@/stores/workspace";
import { useTranslation } from "@/composables/useTranslation";
import { toast } from "@/lib/toast";

const { t } = useTranslation();
const wsStore = useWorkspaceStore();
const wsId = computed(() => wsStore.current?.id ?? 0);

const query = ref("");
const result = ref<PQLParseResult | null>(null);
const fields = ref<PQLField[]>([]);
const sortFields = ref<Record<string, boolean>>({});
const loading = ref(false);

/* ---- 保存的查询 ---- */

const STORAGE_KEY = "ydsz-pql-saved";
const saved = ref<Array<{ name: string; query: string }>>([]);

function loadSaved() {
  try {
    saved.value = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? "[]");
  } catch {
    saved.value = [];
  }
}

function saveQuery() {
  const name = window.prompt("查询名称:");
  if (!name) return;
  saved.value = [...saved.value, { name, query: query.value }];
  localStorage.setItem(STORAGE_KEY, JSON.stringify(saved.value));
  toast.success("查询已保存");
}

function removeSaved(idx: number) {
  saved.value = saved.value.filter((_, i) => i !== idx);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(saved.value));
}

onMounted(async () => {
  loadSaved();
  if (!wsId.value) return;
  try {
    const res = await pqlApi.listFields(wsId.value);
    fields.value = (res as { fields?: PQLField[] }).fields ?? [];
    sortFields.value = (res as { sort_fields?: Record<string, boolean> }).sort_fields ?? {};
  } catch {
    /* 忽略 */
  }
});

/* ---- 操作 ---- */

async function explain() {
  if (!query.value.trim()) {
    toast.error("请输入查询");
    return;
  }
  loading.value = true;
  try {
    result.value = await pqlApi.explain(wsId.value, query.value);
  } catch (e: unknown) {
    result.value = { valid: false, error: e instanceof Error ? e.message : "解释失败", limit: 100, offset: 0 };
  } finally {
    loading.value = false;
  }
}

function applyQuery(q: string) {
  query.value = q;
}

const examples = [
  { name: "进行中的高优先级", query: 'state_group = "started" AND priority = "high"' },
  { name: "指派给我", query: 'assignee = me' },
  { name: "本周到期", query: 'target_date <= "2026-12-31" AND state_group != "completed"' },
  { name: "高严重度缺陷", query: 'type = "defect" AND severity >= 4' },
  { name: "标题包含登录", query: 'title ~ "登录"' },
  { name: "多标签", query: 'label in ("backend", "urgent")' },
  { name: "排序分页", query: 'state_group = "started" sort:updated_at desc limit:20' },
];
</script>

<template>
  
  <AppPageContainer max="1400px">
    <div class="pqlc">
    <div class="pqlc__header">
      <div>
        <h1 class="pqlc__title">{{ t("pql.title", "PQL 查询控制台") }}</h1>
        <p class="pqlc__subtitle">
          {{ t("pql.subtitle", "使用 Plane Query Language 精确筛选工作项") }}
        </p>
      </div>
      <div class="pqlc__actions">
        <button class="pqlc__btn" @click="saveQuery">保存查询</button>
        <button class="pqlc__btn pqlc__btn--primary" :disabled="loading" @click="explain">
          {{ loading ? "解释中..." : "解释查询" }}
        </button>
      </div>
    </div>

    <div class="pqlc__layout">
      <!-- 左侧：查询构建 -->
      <div class="pqlc__main">
        <div class="pqlc__card">
          <h2 class="pqlc__card-title">查询</h2>
          <PqlQueryBuilder
            v-model="query"
            :workspace-id="wsId"
            @change="result = null"
          />
        </div>

        <!-- 结果 -->
        <div v-if="result" class="pqlc__card">
          <h2 class="pqlc__card-title">编译结果</h2>
          <div
            class="pqlc__badge"
            :class="result.valid ? 'pqlc__badge--ok' : 'pqlc__badge--err'"
          >
            {{ result.valid ? "✓ 合法" : `✗ ${result.error}` }}
          </div>
          <div v-if="result.valid" class="pqlc__result">
            <div class="pqlc__result-row">
              <span class="pqlc__result-label">SQL</span>
              <code class="pqlc__code">{{ result.where_sql || "（无过滤条件）" }}</code>
            </div>
            <div v-if="result.args?.length" class="pqlc__result-row">
              <span class="pqlc__result-label">参数</span>
              <code class="pqlc__code">{{ JSON.stringify(result.args) }}</code>
            </div>
            <div v-if="result.sort_column" class="pqlc__result-row">
              <span class="pqlc__result-label">排序</span>
              <code class="pqlc__code">
                {{ result.sort_column }} {{ result.sort_dir }}
              </code>
            </div>
            <div class="pqlc__result-row">
              <span class="pqlc__result-label">分页</span>
              <code class="pqlc__code">limit {{ result.limit }} offset {{ result.offset }}</code>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：参考 -->
      <aside class="pqlc__side">
        <div class="pqlc__card">
          <h2 class="pqlc__card-title">示例</h2>
          <div class="pqlc__list">
            <button
              v-for="ex in examples"
              :key="ex.query"
              class="pqlc__list-item"
              @click="applyQuery(ex.query)"
            >
              <span class="pqlc__list-item-name">{{ ex.name }}</span>
              <code class="pqlc__list-item-query">{{ ex.query }}</code>
            </button>
          </div>
        </div>

        <div v-if="saved.length" class="pqlc__card">
          <h2 class="pqlc__card-title">已保存</h2>
          <div class="pqlc__list">
            <div
              v-for="(s, i) in saved"
              :key="s.query"
              class="pqlc__list-item pqlc__list-item--row"
            >
              <button class="pqlc__list-item-btn" @click="applyQuery(s.query)">
                <span class="pqlc__list-item-name">{{ s.name }}</span>
                <code class="pqlc__list-item-query">{{ s.query }}</code>
              </button>
              <button class="pqlc__link pqlc__link--danger" @click="removeSaved(i)">
                删除
              </button>
            </div>
          </div>
        </div>

        <div class="pqlc__card">
          <h2 class="pqlc__card-title">可用字段</h2>
          <div class="pqlc__fields">
            <div v-for="f in fields" :key="f.name" class="pqlc__field">
              <span class="pqlc__field-name">{{ f.name }}</span>
              <span class="pqlc__field-type">{{ f.type }}</span>
            </div>
          </div>
        </div>
      </aside>
    </div>
  
    </div>
  </AppPageContainer>
</template>

<style scoped>
.pqlc {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  
  
}

.pqlc__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.pqlc__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--txt-primary);
}

.pqlc__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--txt-secondary);
}

.pqlc__actions {
  display: flex;
  gap: 8px;
}

.pqlc__btn {
  padding: 8px 14px;
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.pqlc__btn--primary {
  color: var(--text-on-color);
  background: var(--accent);
  border-color: var(--accent);
}

.pqlc__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pqlc__layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 20px;
  align-items: start;
}

.pqlc__main,
.pqlc__side {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pqlc__card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--bg-surface);
}

.pqlc__card-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--txt-primary);
}

.pqlc__badge {
  align-self: flex-start;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.pqlc__badge--ok {
  background: var(--success-bg, rgba(16, 185, 129, 0.12));
  color: var(--success);
}

.pqlc__badge--err {
  background: var(--danger-bg, rgba(239, 68, 68, 0.12));
  color: var(--danger);
}

.pqlc__result {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pqlc__result-row {
  display: flex;
  gap: 10px;
  font-size: 12px;
}

.pqlc__result-label {
  width: 44px;
  flex-shrink: 0;
  color: var(--txt-tertiary);
}

.pqlc__code {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  color: var(--txt-secondary);
  word-break: break-all;
}

.pqlc__list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.pqlc__list-item {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: none;
  text-align: left;
  cursor: pointer;
}

.pqlc__list-item:hover {
  background: var(--bg-surface-2);
  border-color: var(--accent);
}

.pqlc__list-item--row {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.pqlc__list-item-btn {
  display: flex;
  flex-direction: column;
  gap: 3px;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  flex: 1;
}

.pqlc__list-item-name {
  font-size: 13px;
  color: var(--txt-primary);
}

.pqlc__list-item-query {
  font-family: ui-monospace, monospace;
  font-size: 11px;
  color: var(--txt-tertiary);
}

.pqlc__link {
  font-size: 12px;
  color: var(--txt-tertiary);
  background: none;
  border: none;
  cursor: pointer;
}

.pqlc__link--danger:hover {
  color: var(--danger);
}

.pqlc__fields {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 320px;
  overflow-y: auto;
}

.pqlc__field {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding: 5px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.pqlc__field:hover {
  background: var(--bg-surface-2);
}

.pqlc__field-name {
  font-family: ui-monospace, monospace;
  color: var(--txt-primary);
}

.pqlc__field-type {
  color: var(--txt-tertiary);
}

@media (max-width: 900px) {
  .pqlc__layout {
    grid-template-columns: 1fr;
  }
}
</style>
