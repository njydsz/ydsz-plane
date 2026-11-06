/**
 * 页面评论 API — 封装页面级 / 行内评论、线程回复、表情回应与线程解决的 REST 接口。
 *
 * 路由:
 *   GET    /api/v1/workspaces/:ws/pages/:page_id/comments?page_type=project|knowledge
 *   POST   /api/v1/workspaces/:ws/pages/:page_id/comments        {content, parent_id?, block_id?, quote_text?}
 *   PATCH  /api/v1/workspaces/:ws/comments/:id                   {content}
 *   DELETE /api/v1/workspaces/:ws/comments/:id
 *   POST   /api/v1/workspaces/:ws/comments/:id/resolve
 *   POST   /api/v1/workspaces/:ws/comments/:id/unresolve
 *   POST   /api/v1/workspaces/:ws/comments/:id/reactions         {emoji}
 *   DELETE /api/v1/workspaces/:ws/comments/:id/reactions/:emoji
 *   GET    /api/v1/workspaces/:ws/comments/:id/reactions
 *
 * 说明：
 *   - 列表接口返回扁平结构（{ results: PageComment[] }），前端按 parent_id 组装线程；
 *   - 表情回应不内嵌在评论中，需按评论逐条拉取（listReactions）；
 *   - resolve/unresolve 作用于整个线程（后端会向上追溯到根评论）。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 评论归属的页面类型 */
export type PageCommentPageType = "project" | "knowledge";

/** 页面评论（支持页面级与行内评论） */
export interface PageComment {
  id: number;
  workspace_id: number;
  page_id: number;
  page_type: PageCommentPageType;
  /** 回复的评论 ID（线程） */
  parent_id?: number | null;
  author_id: number;
  /** 纯文本内容 */
  content: string;
  /** 行内评论定位：文档块标识 */
  block_id?: string | null;
  /** 行内评论引用的原文 */
  quote_text?: string | null;
  /** 线程已解决时间 */
  resolved_at?: string | null;
  resolved_by?: number;
  mentions?: number[];
  is_edited: boolean;
  /** 审核隐藏状态：true 表示该评论已被审核者隐藏 */
  hidden: boolean;
  /** 最近审核时间 */
  moderated_at?: string | null;
  /** 最近审核人 ID */
  moderated_by?: number | null;
  /** 审核原因（隐藏原因） */
  moderation_reason?: string | null;
  author_name?: string;
  author_avatar?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

/** 评论表情回应 */
export interface PageCommentReaction {
  id: number;
  comment_id: number;
  user_id: number;
  emoji: string;
  created_at: string;
}

/** 创建评论入参 */
export interface CreatePageCommentInput {
  content: string;
  parent_id?: number | null;
  block_id?: string;
  quote_text?: string;
  mentions?: number[];
}

/** 更新评论入参 */
export interface UpdatePageCommentInput {
  content: string;
  mentions?: number[];
}

/** 评论列表查询参数 */
export interface ListPageCommentsParams {
  page_type: PageCommentPageType;
  limit?: number;
  offset?: number;
  /** 是否包含被审核隐藏的评论（管理视图传 true） */
  include_hidden?: boolean;
  /** 仅返回被隐藏的评论（审核队列传 true） */
  hidden_only?: boolean;
}

/** 评论列表响应 */
export interface ListPageCommentsResult {
  results: PageComment[];
}

/** 表情回应列表响应 */
export interface ListReactionsResult {
  results: PageCommentReaction[];
}

/** 评论审核动作：hide 隐藏 / unhide 取消隐藏 */
export type CommentModerationAction = "hide" | "unhide";

/** 评论审核入参 */
export interface ModerateCommentInput {
  action: CommentModerationAction;
  /** 隐藏原因（action=hide 时建议必填，仅管理员/审核者可见） */
  reason?: string;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 页面评论域 API */
export const pageCommentApi = {
  /** 列出页面下的全部评论（扁平结构，前端按 parent_id 组装线程） */
  list: (
    workspaceId: number | string,
    pageId: number | string,
    params: ListPageCommentsParams,
  ) =>
    wrap<ListPageCommentsResult>(
      http.get(`/workspaces/${workspaceId}/pages/${pageId}/comments`, { params }),
    ),

  /** 创建评论（支持回复与行内定位） */
  create: (
    workspaceId: number | string,
    pageId: number | string,
    pageType: PageCommentPageType,
    data: CreatePageCommentInput,
  ) =>
    wrap<PageComment>(
      http.post(`/workspaces/${workspaceId}/pages/${pageId}/comments`, data, {
        params: { page_type: pageType },
      }),
    ),

  /** 编辑评论内容（仅作者本人） */
  update: (
    workspaceId: number | string,
    commentId: number | string,
    data: UpdatePageCommentInput,
  ) =>
    wrap<PageComment>(
      http.patch(`/workspaces/${workspaceId}/comments/${commentId}`, data),
    ),

  /** 删除评论（软删除，仅作者本人） */
  remove: (workspaceId: number | string, commentId: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/comments/${commentId}`)),

  /** 标记评论线程为已解决 */
  resolve: (workspaceId: number | string, commentId: number | string) =>
    wrap<{ ok: boolean }>(
      http.post(`/workspaces/${workspaceId}/comments/${commentId}/resolve`),
    ),

  /** 取消评论线程的解决状态 */
  unresolve: (workspaceId: number | string, commentId: number | string) =>
    wrap<{ ok: boolean }>(
      http.post(`/workspaces/${workspaceId}/comments/${commentId}/unresolve`),
    ),

  /** 列出某条评论的全部表情回应 */
  listReactions: (workspaceId: number | string, commentId: number | string) =>
    wrap<ListReactionsResult>(
      http.get(`/workspaces/${workspaceId}/comments/${commentId}/reactions`),
    ),

  /** 添加 / 切换表情回应（同一用户同一 emoji 幂等） */
  addReaction: (
    workspaceId: number | string,
    commentId: number | string,
    emoji: string,
  ) =>
    wrap<PageCommentReaction>(
      http.post(`/workspaces/${workspaceId}/comments/${commentId}/reactions`, { emoji }),
    ),

  /** 移除表情回应 */
  removeReaction: (
    workspaceId: number | string,
    commentId: number | string,
    emoji: string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/comments/${commentId}/reactions/${encodeURIComponent(emoji)}`,
      ),
    ),

  /**
   * 审核评论（隐藏 / 取消隐藏）。
   *
   * 权限：需要 comment:moderate 权限（工作空间 owner / admin / pm）。
   * 隐藏后该评论对普通成员不可见；带 comment:moderate 时列表传 include_hidden=true
   * 即可看到被隐藏的评论与原因。
   */
  moderateComment: (
    workspaceId: number | string,
    commentId: number | string,
    action: CommentModerationAction,
    reason?: string,
  ) =>
    wrap<{ ok: boolean }>(
      http.post(`/workspaces/${workspaceId}/comments/${commentId}/moderate`, {
        action,
        reason: reason || undefined,
      }),
    ),
};
