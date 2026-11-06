/**
 * useConfirm — 统一确认对话框。
 *
 * 替代原生 window.confirm()，提供一致的视觉与交互。
 * 用法：
 *   const confirm = useConfirm();
 *   if (await confirm({ title: "删除项目", message: "确定删除吗？" })) { ... }
 */
import { ref } from "vue";
import AppModal from "@/components/AppModal.vue";
import AppButton from "@/components/AppButton.vue";
import { createApp, h } from "vue";

export interface ConfirmOptions {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: "danger" | "primary";
}

let confirmState: ReturnType<typeof createConfirmHost> | null = null;

function createConfirmHost() {
  const visible = ref(false);
  const title = ref("");
  const message = ref("");
  const confirmText = ref("确定");
  const cancelText = ref("取消");
  const variant = ref<"danger" | "primary">("danger");
  let resolveFn: ((v: boolean) => void) | null = null;

  function open(opts: ConfirmOptions): Promise<boolean> {
    title.value = opts.title ?? "确认操作";
    message.value = opts.message;
    confirmText.value = opts.confirmText ?? "确定";
    cancelText.value = opts.cancelText ?? "取消";
    variant.value = opts.variant ?? "danger";
    visible.value = true;
    return new Promise<boolean>((resolve) => {
      resolveFn = resolve;
    });
  }

  function onConfirm() {
    visible.value = false;
    resolveFn?.(true);
    resolveFn = null;
  }

  function onCancel() {
    visible.value = false;
    resolveFn?.(false);
    resolveFn = null;
  }

  return { visible, title, message, confirmText, cancelText, variant, open, onConfirm, onCancel };
}

function ensureHost() {
  if (confirmState) return confirmState;
  const host = createConfirmHost();
  const mountEl = document.createElement("div");
  mountEl.id = "app-confirm-host";
  document.body.appendChild(mountEl);
  createApp({
    render: () =>
      h(AppModal, { visible: host.visible.value, title: host.title.value, width: "420px", onClose: host.onCancel }, {
        default: () => h("p", { style: "color: var(--text-secondary); line-height: 1.6; margin: 0;" }, host.message.value),
        footer: () =>
          h("div", { style: "display: flex; gap: 8px; justify-content: flex-end;" }, [
            h(AppButton, { variant: "ghost", onClick: host.onCancel }, () => host.cancelText.value),
            h(AppButton, { variant: host.variant.value, onClick: host.onConfirm }, () => host.confirmText.value),
          ]),
      }),
  }).mount(mountEl);
  confirmState = host;
  return host;
}

export function useConfirm() {
  const host = ensureHost();
  return host.open;
}
