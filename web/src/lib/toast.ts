/**
 * toast — 全局消息提示系统（操作反馈）。
 *
 * 用法：
 *   import { toast } from "@/lib/toast";
 *   toast.success("保存成功");
 *   toast.error("网络异常");
 *   toast.info("正在处理...");
 *
 * Promise Toast（异步操作自动管理状态）：
 *   import { promiseToast } from "@/lib/toast";
 *   await promiseToast(
 *     api.updateIssue(id, data),
 *     {
 *       loading: "正在保存...",
 *       success: () => `已保存 ${data.name}`,
 *       error: () => "保存失败，请重试",
 *     }
 *   );
 *
 * 渲染由 AppToast.vue 负责；本模块只维护消息队列与订阅。
 */
import { reactive } from "vue";

/** 消息类型（success / error / info / warning / loading）。 */
export type ToastType = "success" | "error" | "info" | "warning" | "loading";

/** 单条消息项（id + type + 内容 + 持续时长，<=0 表示常驻）。 */
export interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
  duration: number; // ms；<=0 表示常驻（需手动关闭）
}

/** Promise Toast 状态文案配置（支持字符串或函数）。 */
export interface PromiseToastMessages<T> {
  loading: string | (() => string);
  success: string | ((data: T) => string);
  error: string | ((err: unknown) => string);
}

/** 消息队列（reactive 以便组件渲染） */
export const toasts = reactive<ToastItem[]>([]);

let nextId = 1;
const DEFAULT_DURATION = 3000;

function resolveMessage(msg: string | ((...args: any[]) => string), ...args: any[]): string {
  return typeof msg === "function" ? msg(...args) : msg;
}

function push(type: ToastType, message: string, duration = DEFAULT_DURATION): number {
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

/** 更新已有消息（用于 Promise Toast 状态切换）。 */
export function update(id: number, type: ToastType, message: string, duration = DEFAULT_DURATION) {
  const item = toasts.find((t) => t.id === id);
  if (item) {
    item.type = type;
    item.message = message;
    item.duration = duration;
    // 如果原来是 loading（常驻），现在变成 auto-dismiss，设置定时器
    if (duration > 0) {
      setTimeout(() => dismiss(id), duration);
    }
  }
}

/**
 * Promise Toast — 异步操作自动管理 loading / success / error 三态。
 *
 * 工作流程：
 *   1. 立即展示 loading 消息（常驻不自动关闭）
 *   2. promise 完成后自动切换为 success 或 error
 *   3. success/error 后按默认时长自动关闭
 *
 * 注意事项：
 *   - 如 promise 抛出异常（reject），会自动展示 error 消息并重新抛出
 *   - 返回 promise 的 resolve 值，可链式处理
 */
export async function promiseToast<T>(
  promise: Promise<T>,
  messages: PromiseToastMessages<T>,
): Promise<T> {
  const loadingMsg = resolveMessage(messages.loading);
  const id = push("loading", loadingMsg, 0); // 0 = 常驻

  try {
    const result = await promise;
    const successMsg = resolveMessage(messages.success as any, result);
    update(id, "success", successMsg, DEFAULT_DURATION);
    return result;
  } catch (err) {
    const errorMsg = resolveMessage(messages.error as any, err);
    update(id, "error", errorMsg, DEFAULT_DURATION + 1000); // error 稍长一点
    throw err;
  }
}

/** 全局消息提示对象 — success / error / info / warning 四方法 + dismiss + promiseToast。 */
export const toast = {
  success: (msg: string, duration?: number) => push("success", msg, duration),
  error: (msg: string, duration?: number) => push("error", msg, duration ?? 4000),
  info: (msg: string, duration?: number) => push("info", msg, duration),
  warning: (msg: string, duration?: number) => push("warning", msg, duration),

  /** 常驻型 loading 消息（需手动通过 dismiss 关闭或配合 promiseToast 使用）。 */
  loading: (msg: string) => push("loading", msg, 0),

  /** Promise Toast — 操作反馈增强。 */
  promise: promiseToast,
};
