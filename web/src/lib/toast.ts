/**
 * toast — 轻量级全局消息提示（操作成功/失败反馈）。
 *
 * 用法：
 *   import { toast } from "@/lib/toast";
 *   toast.success("保存成功");
 *   toast.error("网络异常");
 *   toast.info("正在处理...");
 *
 * 渲染由 AppToast.vue 负责；本模块只维护消息队列与订阅。
 */
import { reactive } from "vue";

export type ToastType = "success" | "error" | "info" | "warning";

export interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
  duration: number; // ms；<=0 表示常驻（需手动关闭）
}

/** 消息队列（reactive 以便组件渲染） */
export const toasts = reactive<ToastItem[]>([]);

let nextId = 1;
const DEFAULT_DURATION = 3000;

function push(type: ToastType, message: string, duration = DEFAULT_DURATION) {
  const id = nextId++;
  toasts.push({ id, type, message, duration });

  if (duration > 0) {
    setTimeout(() => dismiss(id), duration);
  }
  return id;
}

/** 手动关闭某条消息 */
export function dismiss(id: number) {
  const idx = toasts.findIndex((t) => t.id === id);
  if (idx >= 0) toasts.splice(idx, 1);
}

export const toast = {
  success: (msg: string, duration?: number) => push("success", msg, duration),
  error: (msg: string, duration?: number) => push("error", msg, duration ?? 4000),
  info: (msg: string, duration?: number) => push("info", msg, duration),
  warning: (msg: string, duration?: number) => push("warning", msg, duration),
};
