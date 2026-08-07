import { defineStore } from "pinia";

import { workspaceApi, type Workspace } from "@/api/services/workspace";

export const useWorkspaceStore = defineStore("workspace", {
  state: () => ({
    list: [] as Workspace[],
    current: null as Workspace | null,
    loaded: false,
    slug: "" as string,
  }),
  getters: {
    currentSlug: (s) => s.current?.slug ?? s.slug,
    currentRole: (s) => s.current?.role ?? "",
    canManage: (s) => ["owner", "admin"].includes(s.current?.role ?? ""),
  },
  actions: {
    async load(force = false) {
      if (this.loaded && !force) return;
      this.list = await workspaceApi.list();
      this.loaded = true;
    },
    setCurrent(ws: Workspace | null, slug = "") {
      this.current = ws;
      this.slug = ws?.slug ?? slug;
    },
    async resolveBySlug(slug: string) {
      if (this.current?.slug === slug) return this.current;
      await this.load();
      const ws = this.list.find((w) => w.slug === slug) ?? null;
      this.setCurrent(ws, slug);
      return ws;
    },
    add(ws: Workspace) {
      this.list.unshift(ws);
      this.current = ws;
      this.slug = ws.slug;
    },
  },
});
