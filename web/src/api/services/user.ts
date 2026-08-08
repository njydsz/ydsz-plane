/**
 * 用户域 API — 当前登录用户的个人信息管理。
 *
 * 对标 authApi.me() 的扩展：提供 update 能力，
 * 前端 Profile 设置页通过 PATCH /me 提交修改。
 */
import { http } from "../client";
import type { UserBrief } from "@/stores/auth";

/** 可编辑的用户资料字段（与后端 UserUpdateInput 对齐） */
export interface UserProfileInput {
  display_name?: string;
  avatar_url?: string;
  timezone?: string;
  language?: string;
}

/** 用户域 API — 更新当前登录用户的个人信息。 */
export const userApi = {
  /** 部分更新当前用户资料 */
  update: (input: UserProfileInput) =>
    http.patch<UserBrief>("/me", input).then((r) => ({
      data: r.data,
      requestId: r.headers["x-request-id"],
    })),
};
