import type { Meta, StoryObj } from "@storybook/vue3"
import AppToast from "../AppToast.vue"

/**
 * AppToast 依赖全局 reactive 消息队列 (lib/toast.ts)，
 * 无法在隔离 args 下展示具体消息内容。
 * 此处通过 render 函数手动注入消息状态演示各类型 Toast 外观。
 */

const meta: Meta<typeof AppToast> = {
  title: "Components/AppToast",
  component: AppToast,
  tags: ["autodocs"],
  parameters: {
    layout: "padded",
    docs: {
      description: {
        component:
          "AppToast 消费全局 `toasts` reactive 队列。实际使用时通过 `toast.success('消息')` / `toast.error('错误')` 等方法触发。此处展示各类型 Toast 的视觉样式。",
      },
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const AllTypes: Story = {
  render: () => ({
    components: { AppToast },
    template: `
      <div style="display:flex;flex-direction:column;gap:10px;max-width:380px;">
        <!-- Success -->
        <div style="display:flex;align-items:flex-start;gap:10px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #16a34a;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
          <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#16a34a;flex-shrink:0;">✓</span>
          <div style="display:flex;flex-direction:column;gap:2px;">
            <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">成功</span>
            <span style="font-size:13px;color:#1f2937;">操作已成功完成</span>
          </div>
        </div>
        <!-- Error -->
        <div style="display:flex;align-items:flex-start;gap:10px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #ef4444;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
          <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#ef4444;flex-shrink:0;">✕</span>
          <div style="display:flex;flex-direction:column;gap:2px;">
            <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">错误</span>
            <span style="font-size:13px;color:#1f2937;">操作失败，请重试</span>
          </div>
        </div>
        <!-- Info -->
        <div style="display:flex;align-items:flex-start;gap:10px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #3b82f6;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
          <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#3b82f6;flex-shrink:0;">ℹ</span>
          <div style="display:flex;flex-direction:column;gap:2px;">
            <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">提示</span>
            <span style="font-size:13px;color:#1f2937;">这是一条信息提示</span>
          </div>
        </div>
        <!-- Warning -->
        <div style="display:flex;align-items:flex-start;gap:10px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #f59e0b;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
          <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#f59e0b;flex-shrink:0;">!</span>
          <div style="display:flex;flex-direction:column;gap:2px;">
            <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">警告</span>
            <span style="font-size:13px;color:#1f2937;">请注意操作风险</span>
          </div>
        </div>
        <!-- Loading -->
        <div style="display:flex;align-items:flex-start;gap:10px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #3b82f6;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
          <span style="width:18px;height:18px;border:2px solid #e5e7eb;border-top-color:#3b82f6;border-radius:50%;animation:spin 0.65s linear infinite;flex-shrink:0;margin-top:1px;"></span>
          <div style="display:flex;flex-direction:column;gap:2px;">
            <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">处理中</span>
            <span style="font-size:13px;color:#1f2937;">正在执行操作，请稍候...</span>
          </div>
        </div>
      </div>
    `,
  }),
}

export const Stacked: Story = {
  render: () => ({
    components: { AppToast },
    template: `
      <div style="position:relative;width:380px;height:300px;">
        <div style="position:absolute;top:16px;right:16px;display:flex;flex-direction:column;gap:8px;">
          <div style="display:flex;align-items:flex-start;gap:10px;min-width:260px;max-width:380px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #16a34a;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
            <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#16a34a;flex-shrink:0;">✓</span>
            <div style="display:flex;flex-direction:column;gap:2px;">
              <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">成功</span>
              <span style="font-size:13px;color:#1f2937;">第一条通知</span>
            </div>
          </div>
          <div style="display:flex;align-items:flex-start;gap:10px;min-width:260px;max-width:380px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #ef4444;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
            <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#ef4444;flex-shrink:0;">✕</span>
            <div style="display:flex;flex-direction:column;gap:2px;">
              <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">错误</span>
              <span style="font-size:13px;color:#1f2937;">第二条通知</span>
            </div>
          </div>
          <div style="display:flex;align-items:flex-start;gap:10px;min-width:260px;max-width:380px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #3b82f6;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
            <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#3b82f6;flex-shrink:0;">ℹ</span>
            <div style="display:flex;flex-direction:column;gap:2px;">
              <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">提示</span>
              <span style="font-size:13px;color:#1f2937;">第三条通知</span>
            </div>
          </div>
        </div>
      </div>
    `,
  }),
}

export const SingleToast: Story = {
  render: () => ({
    components: { AppToast },
    template: `
      <div style="max-width:380px;">
        <div style="display:flex;align-items:flex-start;gap:10px;min-width:260px;max-width:380px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #16a34a;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
          <span style="width:20px;height:20px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:12px;color:#fff;background:#16a34a;flex-shrink:0;">✓</span>
          <div style="flex:1;display:flex;flex-direction:column;gap:2px;">
            <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">成功</span>
            <span style="font-size:13px;color:#1f2937;">保存需求成功，ID 为 PROJ-1024</span>
          </div>
          <button style="background:none;border:none;padding:0 2px;font-size:15px;line-height:1;color:#9ca3af;cursor:pointer;">×</button>
        </div>
      </div>
    `,
  }),
}

export const LoadingSpinner: Story = {
  render: () => ({
    components: { AppToast },
    template: `
      <style>@keyframes toast-spin { to { transform: rotate(360deg); } }</style>
      <div style="max-width:380px;">
        <div style="display:flex;align-items:flex-start;gap:10px;min-width:260px;max-width:380px;padding:10px 12px;border-radius:8px;background:#fff;border:1px solid #d1d5db;border-left:3px solid #3b82f6;box-shadow:0 4px 16px rgba(0,0,0,0.1);">
          <span style="width:18px;height:18px;border:2px solid #e5e7eb;border-top-color:#3b82f6;border-radius:50%;animation:toast-spin 0.65s linear infinite;flex-shrink:0;margin-top:1px;"></span>
          <div style="flex:1;display:flex;flex-direction:column;gap:2px;">
            <span style="font-size:11px;font-weight:600;color:#9ca3af;text-transform:uppercase;letter-spacing:0.04em;">处理中</span>
            <span style="font-size:13px;color:#1f2937;">正在上传附件，请稍候...</span>
          </div>
        </div>
      </div>
    `,
  }),
}
