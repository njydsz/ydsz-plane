/**
 * 工作空间 Pinia store — 管理工作空间列表、当前空间、角色与权限判断。
 */
import { defineStore } from "pinia";

import { workspaceApi, type Workspace, type MyRoleResponse, type RoleDefinition, type ProjectModuleToggles } from "@/api/services/workspace";

/** 工作空间 store 状态。显式声明 state 类型，规避 Pinia 对 Set 字段的推断歧义。 */
interface WorkspaceState {
  list: Workspace[];
  current: Workspace | null;
  loaded: boolean;
  slug: string;
  /** 当前用户在该工作空间的权限码集合 */
  permissions: Set<string>;
  /** 当前用户在该工作空间的角色详情 */
  roleDetail: RoleDefinition | null;
  /** 当前项目的功能模块开关（null 表示未加载） */
  projectModules: ProjectModuleToggles | null;
}

/** 工作空间 store */
export const useWorkspaceStore = defineStore("workspace", {
  state: (): WorkspaceState => ({
    list: [],
    current: null,
    loaded: false,
    slug: "",
    permissions: new Set<string>(),
    roleDetail: null,
    projectModules: null,
  }),
  getters: {
    /** 当前空间 ID */
    currentId: (s) => s.current?.id ?? 0,
    /** 当前空间 slug */
    currentSlug: (s) => s.current?.slug ?? s.slug,
    /** 当前用户在当前空间的角色 slug */
    currentRole: (s) => s.current?.role ?? "",
    /** 当前用户角色级别数值（便于比大小） */
    currentLevel(): number {
      return this.roleDetail?.level ?? 0;
    },
    /** 是否 owner */
    isOwner: (s) => s.current?.role === "owner",
    /** 是否 owner 或 admin */
    canManage: (s) => ["owner", "admin"].includes(s.current?.role ?? ""),
    /** 是否 owner/admin/pm */
    canAdminister: (s) => ["owner", "admin", "pm"].includes(s.current?.role ?? ""),
    /** 是否技术角色（techlead / dev） */
    isTechRole: (s) => ["techlead", "dev"].includes(s.current?.role ?? ""),
    /** 检查当前项目的某功能模块是否启用（未加载时默认 true） */
    isProjectModuleEnabled: (s) => (moduleKey: keyof ProjectModuleToggles) => {
      if (!s.projectModules) return true;
      return s.projectModules[moduleKey] !== false;
    },
  },
  actions: {
    /** 加载当前用户在该工作空间的权限 + 角色详情 */
    async loadPermissions(): Promise<void> {
      if (!this.currentId) return;
      try {
        const resp: MyRoleResponse = await workspaceApi.getMyRole(this.currentId);
        this.permissions = new Set(resp.permissions);
        this.roleDetail = resp.role;
      } catch {
        this.permissions = new Set();
        this.roleDetail = null;
      }
    },
    /** 检查是否拥有某权限 */
    hasPermission(perm: string): boolean {
      return this.permissions.has(perm);
    },
    /** 检查是否拥有任意一项权限 */
    hasAnyPermission(perms: string[]): boolean {
      return perms.some((p) => this.permissions.has(p));
    },
    /**
     * 判断菜单项是否可见：
     *  1) 若定义了 permissions：满足任一权限可见
     *  2) 若定义了 minLevel：角色级别 ≥ 该值可见
     *  3) 都未定义：始终可见
     */
    canSeeMenu(item: { permissions?: string[]; minLevel?: number }): boolean {
      if (item.permissions && item.permissions.length > 0 && !this.hasAnyPermission(item.permissions)) {
        return false;
      }
      if (item.minLevel !== undefined && this.currentLevel < item.minLevel) {
        return false;
      }
      return true;
    },
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
      this.permissions = new Set();
      this.roleDetail = null;
    },
    /** 设置当前项目的功能模块开关（由项目详情页加载时调用） */
    setProjectModules(modules: ProjectModuleToggles | null) {
      this.projectModules = modules;
    },
    /** 按 ID 解析并切换当前空间（命中缓存直接返回） */
    async resolveById(id: number) {
      if (this.current?.id === id) {
        if (this.permissions.size === 0) await this.loadPermissions();
        return this.current;
      }
      await this.load();
      const ws = this.list.find((w) => w.id === id) ?? null;
      if (ws) {
        this.setCurrent(ws);
        await this.loadPermissions();
        return ws;
      }
      try {
        const fresh = await workspaceApi.get(id);
        this.setCurrent(fresh);
        await this.loadPermissions();
        return fresh;
      } catch {
        this.setCurrent(null);
        return null;
      }
    },
    /** 按 slug 解析（兼容旧链接） */
    async resolveBySlug(slug: string) {
      if (this.current?.slug === slug) {
        if (this.permissions.size === 0) await this.loadPermissions();
        return this.current;
      }
      await this.load();
      const ws = this.list.find((w) => w.slug === slug) ?? null;
      this.setCurrent(ws);
      if (ws) await this.loadPermissions();
      return ws;
    },
    /** 新建空间后插入列表头部并切换为当前空间 */
    add(ws: Workspace) {
      this.list.unshift(ws);
      this.setCurrent(ws);
    },
  },
});
