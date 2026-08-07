/**
 * 工作空间 Pinia store — 管理工作空间列表、当前空间（含 ID 解析）与角色判断。
 */
import { defineStore } from "pinia";

import { workspaceApi, type Workspace } from "@/api/services/workspace";

/** 工作空间 store：list 为全部空间，current 为当前进入的空间 */
export const useWorkspaceStore = defineStore("workspace", {
  state: () => ({
    list: [] as Workspace[],
    current: null as Workspace | null,
    loaded: false,
    slug: "" as string,
  }),
  getters: {
    /** 当前空间 ID */
    currentId: (s) => s.current?.id ?? 0,
    /** 当前空间 slug（兼容旧引用，优先取 current） */
    currentSlug: (s) => s.current?.slug ?? s.slug,
    /** 当前用户在当前空间的角色 */
    currentRole: (s) => s.current?.role ?? "",
    /** 是否拥有空间管理权限（owner / admin） */
    canManage: (s) => ["owner", "admin"].includes(s.current?.role ?? ""),
  },
  actions: {
    /** 拉取工作空间列表（已加载且非强制时跳过） */
    async load(force = false) {
      if (this.loaded && !force) return;
      this.list = await workspaceApi.list();
      this.loaded = true;
    },
    /** 设置当前空间 */
    setCurrent(ws: Workspace | null) {
      this.current = ws;
      this.slug = ws?.slug ?? "";
    },
    /** 按 ID 解析并切换当前空间（命中缓存直接返回） */
    async resolveById(id: number) {
      if (this.current?.id === id) return this.current;
      await this.load();
      const ws = this.list.find((w) => w.id === id) ?? null;
      if (ws) {
        this.setCurrent(ws);
        return ws;
      }
      // 缓存未命中（可能新加入的空间），走 API
      try {
        const fresh = await workspaceApi.get(id);
        this.setCurrent(fresh);
        return fresh;
      } catch {
        this.setCurrent(null);
        return null;
      }
    },
    /** 按 slug 解析（兼容旧链接） */
    async resolveBySlug(slug: string) {
      if (this.current?.slug === slug) return this.current;
      await this.load();
      const ws = this.list.find((w) => w.slug === slug) ?? null;
      this.setCurrent(ws);
      return ws;
    },
    /** 新建空间后插入列表头部并切换为当前空间 */
    add(ws: Workspace) {
      this.list.unshift(ws);
      this.current = ws;
      this.slug = ws.slug;
    },
  },
});
