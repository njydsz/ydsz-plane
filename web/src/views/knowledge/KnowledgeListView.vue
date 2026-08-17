<script setup lang="ts">
/**
 * KnowledgeListView — 知识库空间列表页。
 *
 * 展示工作空间下全部知识库空间（卡片网格），支持：
 *  - 新建空间（名称 / slug 自动生成可改 / 描述 / 私有开关 / 默认权限）
 *  - 删除空间（软删除，confirm 确认）
 *  - 点击卡片进入空间详情（文档管理）
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { knowledgeApi, type KnowledgePage, type KnowledgeSpace, type SpacePermission } from "@/api/services/knowledge";
import { toast } from "@/lib/toast";
import { AppEmptyState, AppErrorState, AppModal } from "@/components";

const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const spaces = ref<KnowledgeSpace[]>([]);

/* ===== 全文检索 ===== */
const searchKeyword = ref("");
const searching = ref(false);
const searchResults = ref<KnowledgePage[]>([]);
const searchDone = ref(false);

/* ===== 新建空间弹窗 ===== */
const showCreate = ref(false);
const creating = ref(false);
const formName = ref("");
const formSlug = ref("");
const slugTouched = ref(false);
const formDescription = ref("");
const formIsPrivate = ref(true);
const formPermission = ref<SpacePermission>("viewer");

const permissionOptions: Array<{ value: SpacePermission; label: string }> = [
  { value: "viewer", label: "只读" },
  { value: "editor", label: "可编辑" },
  { value: "admin", label: "管理员" },
  { value: "owner", label: "所有者" },
];

/** 由名称生成 URL 友好的 slug */
function slugify(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

function permissionLabel(p: SpacePermission): string {
  return permissionOptions.find((o) => o.value === p)?.label ?? p;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    spaces.value = await knowledgeApi.listSpaces(workspaceId.value);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  formName.value = "";
  formSlug.value = "";
  slugTouched.value = false;
  formDescription.value = "";
  formIsPrivate.value = true;
  formPermission.value = "viewer";
  showCreate.value = true;
}

function closeCreate() {
  showCreate.value = false;
}

function onNameInput() {
  // 用户未手动编辑 slug 时自动跟随名称
  if (!slugTouched.value) {
    formSlug.value = slugify(formName.value);
  }
}

function onSlugInput() {
  slugTouched.value = true;
}

async function createSpace() {
  if (!formName.value.trim()) {
    toast.error("请输入空间名称");
    return;
  }
  const slug = formSlug.value.trim() || slugify(formName.value);
  if (!slug) {
    toast.error("请输入有效的链接标识");
    return;
  }
  creating.value = true;
  try {
    const sp = await knowledgeApi.createSpace(workspaceId.value, {
      name: formName.value.trim(),
      slug,
      description: formDescription.value.trim(),
      is_private: formIsPrivate.value,
      default_permission: formPermission.value,
    });
    toast.success("空间已创建");
    closeCreate();
    router.push(`/${workspaceId.value}/knowledge/${sp.id}`);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "创建空间失败");
  } finally {
    creating.value = false;
  }
}

async function deleteSpace(sp: KnowledgeSpace) {
  if (!window.confirm(`确定删除空间「${sp.name}」？该空间下所有文档将被一并删除。`)) return;
  try {
    await knowledgeApi.deleteSpace(workspaceId.value, sp.id);
    spaces.value = spaces.value.filter((s) => s.id !== sp.id);
    toast.success("空间已删除");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除空间失败");
  }
}

function fmtTime(iso: string): string {
  const d = new Date(iso);
  return `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, "0")}/${String(
    d.getDate(),
  ).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(
    d.getMinutes(),
  ).padStart(2, "0")}`;
}

async function runSearch() {
	const kw = searchKeyword.value.trim();
	if (!kw) {
		searchResults.value = [];
		searchDone.value = false;
		return;
	}
	searching.value = true;
	searchDone.value = false;
	try {
		searchResults.value = await knowledgeApi.search(workspaceId.value, kw);
	} catch (e: unknown) {
		toast.error(e instanceof Error ? e.message : "搜索失败");
		searchResults.value = [];
	} finally {
		searching.value = false;
		searchDone.value = true;
	}
}

function clearSearch() {
	searchKeyword.value = "";
	searchResults.value = [];
	searchDone.value = false;
}

onMounted(load);
</script>

<template>
  <div class="knowledge-list">
    <header class="knowledge-list__header">
      <div>
        <h1>{{ $t("knowledge.title") }}</h1>
        <p class="hint">团队文档与知识沉淀 — 支持无限层级文档树、版本历史与需求/任务/缺陷关联</p>
      </div>
      <div class="knowledge-list__header-right">
        <button class="btn btn--primary" @click="openCreate">＋ {{ $t("knowledge.newSpace") }}</button>
      </div>
    </header>

    <!-- 全文检索 -->
    <div class="kb-search">
      <input
        v-model="searchKeyword"
        class="kb-search__input"
        :placeholder="$t('knowledge.searchPlaceholder')"
        @keydown.enter="runSearch"
      />
      <button class="btn btn--primary" :disabled="searching" @click="runSearch">
        {{ searching ? $t("common.loading") : $t("knowledge.search") }}
      </button>
      <button v-if="searchKeyword" class="btn" @click="clearSearch">{{ $t("common.clear") }}</button>
    </div>

    <!-- 搜索结果 -->
    <div v-if="searchDone && searchKeyword.trim()" class="kb-search__results">
      <div class="kb-search__head">
        <span>{{ $t("knowledge.searchResults") }}: {{ searchResults.length }}</span>
      </div>
      <div v-if="searchResults.length === 0" class="kb-search__empty">
        {{ $t("knowledge.searchEmpty") }}
      </div>
      <div v-else class="kb-search__list">
        <div
          v-for="p in searchResults"
          :key="p.id"
          class="kb-search__item"
          @click="router.push(`/${workspaceId}/knowledge/${p.space_id}/pages/${p.id}`)"
        >
          <span class="kb-search__item-title">{{ p.title }}</span>
          <span class="kb-search__item-path">{{ p.path }}</span>
        </div>
      </div>
    </div>

    <AppErrorState v-if="error" :message="error" @retry="load" />

    <div v-else-if="loading" class="knowledge-list__loading">
      <div v-for="i in 6" :key="i" class="skeleton-card" />
    </div>

    <AppEmptyState
      v-else-if="spaces.length === 0"
      icon="📚"
      :title="$t('knowledge.empty')"
      :description="$t('knowledge.emptyDesc')"
    >
      <button class="btn btn--primary" @click="openCreate">＋ {{ $t("knowledge.newSpace") }}</button>
    </AppEmptyState>

    <div v-else class="knowledge-list__grid">
      <div
        v-for="sp in spaces"
        :key="sp.id"
        class="space-card"
        @click="router.push(`/${workspaceId}/knowledge/${sp.id}`)"
      >
        <div class="space-card__body">
          <div class="space-card__title-row">
            <span class="space-card__title">{{ sp.name }}</span>
            <span v-if="sp.is_private" class="space-card__private" title="私有">🔒</span>
          </div>
          <p class="space-card__desc">{{ sp.description || "暂无描述" }}</p>
          <div class="space-card__meta">
            <span class="space-card__badge">{{ permissionLabel(sp.default_permission) }}</span>
            <span class="space-card__time">更新于 {{ fmtTime(sp.updated_at) }}</span>
          </div>
        </div>
        <div class="space-card__footer">
          <span class="space-card__slug">/{{ sp.slug }}</span>
          <button
            class="space-card__delete"
            title="删除空间"
            @click.stop="deleteSpace(sp)"
          >
删除
</button>
        </div>
      </div>
    </div>

    <!-- 新建空间 弹窗 -->
    <AppModal :visible="showCreate" :title="$t('knowledge.newSpace')" width="520px" @close="closeCreate">
      <div class="space-form">
        <div class="form-field">
          <label class="form-label">{{ $t("knowledge.spaceName") }}</label>
          <input
            v-model="formName"
            class="form-input"
            :placeholder="$t('knowledge.spaceNamePlaceholder')"
            @input="onNameInput"
          />
        </div>
        <div class="form-field">
          <label class="form-label">{{ $t("knowledge.slug") }}</label>
          <input
            v-model="formSlug"
            class="form-input form-input--mono"
            :placeholder="$t('knowledge.slugPlaceholder')"
            @input="onSlugInput"
          />
        </div>
        <div class="form-field">
          <label class="form-label">{{ $t("knowledge.description") }}</label>
          <textarea
            v-model="formDescription"
            class="form-textarea"
            rows="3"
            :placeholder="$t('knowledge.descriptionPlaceholder')"
          />
        </div>
        <div class="form-field">
          <label class="form-label">{{ $t("knowledge.defaultPermission") }}</label>
          <select v-model="formPermission" class="form-input">
            <option v-for="opt in permissionOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
        <label class="form-check">
          <input v-model="formIsPrivate" type="checkbox" />
          <span>{{ $t("knowledge.private") }}</span>
        </label>
      </div>
      <template #footer>
        <button class="btn" @click="closeCreate">{{ $t("common.cancel") }}</button>
        <button class="btn btn--primary" :disabled="creating" @click="createSpace">
          {{ creating ? $t("common.loading") : $t("knowledge.createSpace") }}
        </button>
      </template>
    </AppModal>
  </div>
</template>

<style scoped>
.knowledge-list__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.knowledge-list__header h1 {
  font-size: 20px;
  margin: 0 0 4px;
}

.hint {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 0;
}

.knowledge-list__header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* ---- 加载骨架 ---- */
.knowledge-list__loading {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.skeleton-card {
  height: 150px;
  border-radius: var(--radius-md);
  background: linear-gradient(90deg, var(--surface-2) 25%, var(--surface-3) 50%, var(--surface-2) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

@keyframes shimmer {
  from { background-position: 200% 0; }
  to { background-position: -200% 0; }
}

/* ---- 卡片网格 ---- */
.knowledge-list__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.space-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-1);
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.space-card:hover {
  border-color: var(--brand-500);
  box-shadow: var(--shadow-popover);
  transform: translateY(-1px);
}

.space-card__body {
  padding: 16px;
  flex: 1;
}

.space-card__title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.space-card__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.space-card__private {
  font-size: 12px;
  flex-shrink: 0;
}

.space-card__desc {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0 0 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 36px;
}

.space-card__meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.space-card__badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 500;
}

.space-card__time {
  font-size: 11px;
  color: var(--text-tertiary);
}

.space-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-top: 1px solid var(--border-subtle);
  background: var(--surface-2);
}

.space-card__slug {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
}

.space-card__delete {
  border: none;
  background: none;
  color: var(--text-tertiary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  opacity: 0;
  transition: opacity 0.15s;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.space-card:hover .space-card__delete {
  opacity: 1;
}

.space-card__delete:hover {
  color: var(--danger-500);
  background: var(--danger-50);
}

/* ---- 表单 ---- */
.space-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.form-input {
  padding: 8px 10px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
}

.form-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.form-input--mono {
  font-family: var(--font-mono);
}

.form-textarea {
  padding: 8px 10px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
  resize: vertical;
}

.form-textarea:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.form-check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}

/* ---- 按钮（与项目其他视图对齐） ---- */
.btn {
  padding: 8px 16px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}

.btn:hover {
  border-color: var(--brand-500);
  color: var(--brand-600);
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 500;
}

.btn--primary:hover {
  background: var(--brand-600);
  color: var(--text-on-brand);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ---- 全文检索 ---- */
.kb-search {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.kb-search__input {
  flex: 1;
  padding: 8px 10px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
}

.kb-search__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.kb-search__results {
  margin-bottom: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.kb-search__head {
  padding: 10px 16px;
  background: var(--surface-2);
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.kb-search__empty {
  padding: 16px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.kb-search__list {
  max-height: 320px;
  overflow-y: auto;
}

.kb-search__item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: background 0.1s;
}

.kb-search__item:last-child {
  border-bottom: none;
}

.kb-search__item:hover {
  background: var(--brand-50);
}

.kb-search__item-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.kb-search__item-path {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
}
</style>
