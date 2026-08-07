/**
 * @ydsz/ui — 精选核心基础组件
 *
 * 设计原则:
 *  1) 单一职责、零外部依赖
 *  2) 设计令牌驱动（复用 tokens.css）
 *  3) Props 语义化、TypeScript 严格要求
 *  4) 可访问性优先（aria-*, role, keyboard）
 *  5) 可组合而非可配置
 *
 * 使用方式:
 *   import { AppButton, AppModal } from "@/components";
 */
export { default as AppAvatar } from "./AppAvatar.vue";
export { default as AppBadge } from "./AppBadge.vue";
export { default as AppButton } from "./AppButton.vue";
export { default as AppCard } from "./AppCard.vue";
export { default as AppEmptyState } from "./AppEmptyState.vue";
export { default as AppErrorState } from "./AppErrorState.vue";
export { default as AppLoadingState } from "./AppLoadingState.vue";
export { default as AppModal } from "./AppModal.vue";
export { default as AttachmentUploader } from "./AttachmentUploader.vue";
export { default as CommentForm } from "./CommentForm.vue";
export { default as CommentItem } from "./CommentItem.vue";
export { default as CommentList } from "./CommentList.vue";
export { default as GlobalSearch } from "./GlobalSearch.vue";
export { default as MiniGantt } from "./MiniGantt.vue";
export { default as NotificationBell } from "./NotificationBell.vue";
export { default as ProgressBar } from "./ProgressBar.vue";
