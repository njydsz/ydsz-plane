/**
 * FavoritesStore — 收藏/置顶管理。
 *
 * 用户可以收藏项目、迭代、页面，置顶在侧边栏快速访问区。
 * 数据持久化到 localStorage（per workspace scope）。
 */

import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";

export type FavoriteType = "project" | "sprint" | "page" | "issue";

export interface FavoriteItem {
  id: string;
  type: FavoriteType;
  /** 目标 ID（projectId / sprintId 等） */
  targetId: number;
  /** 显示名 */
  label: string;
  /** 图标 emoji */
  icon: string;
  /** 路由路径 */
  path: string;
  /** 收藏时间戳（用于排序） */
  addedAt: number;
}

const STORAGE_PREFIX = "ydsz.favorites";

function storageKey(wsId: number): string {
  return `${STORAGE_PREFIX}.ws${wsId}`;
}

function loadFromStorage(wsId: number): FavoriteItem[] {
  try {
    const raw = localStorage.getItem(storageKey(wsId));
    if (!raw) return [];
    const parsed = JSON.parse(raw) as FavoriteItem[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export const useFavoritesStore = defineStore("favorites", () => {
  const items = ref<FavoriteItem[]>([]);
  const currentWsId = ref<number | null>(null);

  /** 当前工作空间的收藏（按添加时间倒序 → 最近添加置顶） */
  const favorites = computed(() => [...items.value].sort((a, b) => b.addedAt - a.addedAt));

  /** 按类型分组 */
  const projects = computed(() => favorites.value.filter((f) => f.type === "project"));
  const pages = computed(() => favorites.value.filter((f) => f.type === "page"));

  /** 切换工作空间时加载对应收藏 */
  function setWorkspace(wsId: number | null) {
    currentWsId.value = wsId;
    items.value = wsId != null ? loadFromStorage(wsId) : [];
  }

  /** 持久化到 localStorage */
  function persist() {
    if (currentWsId.value == null) return;
    localStorage.setItem(storageKey(currentWsId.value), JSON.stringify(items.value));
  }

  function add(input: Omit<FavoriteItem, "addedAt">): boolean {
    if (items.value.some((f) => f.id === input.id)) return false;
    items.value.push({ ...input, addedAt: Date.now() });
    persist();
    return true;
  }

  function remove(id: string): void {
    items.value = items.value.filter((f) => f.id !== id);
    persist();
  }

  function toggle(item: Omit<FavoriteItem, "addedAt">): boolean {
    if (items.value.some((f) => f.id === item.id)) {
      remove(item.id);
      return false;
    }
    add(item);
    return true;
  }

  function isFavorite(id: string): boolean {
    return items.value.some((f) => f.id === id);
  }

  /** 置顶/置底排序时使用 */
  function reorder(fromIdx: number, toIdx: number): void {
    const arr = [...items.value];
    const [moved] = arr.splice(fromIdx, 1);
    if (moved) {
      arr.splice(toIdx, 0, moved);
      items.value = arr;
      persist();
    }
  }

  // 自动持久化（备用，避免忘记调用 persist）
  watch(items, persist, { deep: true });

  return {
    items,
    favorites,
    projects,
    pages,
    currentWsId,
    setWorkspace,
    add,
    remove,
    toggle,
    isFavorite,
    reorder,
  };
});
