/**
 * 工作空间 Pinia store — 管理工作空间列表、当前空间（含 slug 解析）与角色判断。
 */
import { defineStore } from "pinia";

import { workspaceApi, type Workspace } from "@/api/services/workspace";

/** 工作空间 store：list 为全部空间，current 为当前进入的空间，slug 为 URL 中的空间标识 */
export const useWorkspaceStore = defineStore("workspace", {
  state: () => ({
    list: [] as Workspace[],
    current: null as Workspace | null,
    loaded: false,
    slug: "" as string,
  }),
  getters: {
    /** 当前空间 slug（优先取 current，未加载时退回 URL slug） */
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
    /** 设置当前空间与 URL slug */
    setCurrent(ws: Workspace | null, slug = "") {
      this.current = ws;
      this.slug = ws?.slug ?? slug;
    },
    /** 按 slug 解析并切换当前空间（命中缓存直接返回） */
    async resolveBySlug(slug: string) {
      if (this.current?.slug === slug) return this.current;
      await this.load();
      const ws = this.list.find((w) => w.slug === slug) ?? null;
      this.setCurrent(ws, slug);
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
