import type { Meta, StoryObj } from "@storybook/vue3"
import AppModal from "../AppModal.vue"

const meta: Meta<typeof AppModal> = {
  title: "Components/AppModal",
  component: AppModal,
  tags: ["autodocs"],
  argTypes: {
    visible: { control: "boolean" },
    title: { control: "text" },
    width: { control: "text" },
  },
  args: {
    visible: true,
    title: "弹窗标题",
    width: "480px",
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  render: (args) => ({
    components: { AppModal },
    setup() { return { args } },
    template: `
      <AppModal v-bind="args">
        <p style="margin:0;color:var(--text-secondary);line-height:1.6;">这是弹窗正文内容。用户可以阅读详细说明并执行操作。</p>
      </AppModal>
    `,
  }),
}

export const WithFooter: Story = {
  render: (args) => ({
    components: { AppModal },
    setup() { return { args } },
    template: `
      <AppModal v-bind="args">
        <p style="margin:0;color:var(--text-secondary);line-height:1.6;">确认执行此操作？该操作不可撤销。</p>
        <template #footer>
          <button style="padding:6px 14px;border-radius:6px;background:transparent;color:#6b7280;border:1px solid #d1d5db;cursor:pointer;font:inherit;font-size:13px;">取消</button>
          <button style="padding:6px 14px;border-radius:6px;background:#ef4444;color:#fff;border:none;cursor:pointer;font:inherit;font-size:13px;font-weight:500;">删除</button>
        </template>
      </AppModal>
    `,
  }),
}

export const Wide: Story = {
  args: {
    title: "详细信息",
    width: "720px",
  },
  render: (args) => ({
    components: { AppModal },
    setup() { return { args } },
    template: `
      <AppModal v-bind="args">
        <div style="color:var(--text-secondary);line-height:1.7;font-size:14px;">
          <p style="margin-top:0;">这是一个较宽的弹窗，适合展示更多内容，例如表单或详情面板。</p>
          <p>宽度可设置为任意 CSS 值，如 720px、900px 或 90vw 等。</p>
        </div>
      </AppModal>
    `,
  }),
}

export const CustomHeader: Story = {
  args: {
    title: "",
    width: "480px",
  },
  render: (args) => ({
    components: { AppModal },
    setup() { return { args } },
    template: `
      <AppModal v-bind="args">
        <template #header>
          <div style="display:flex;align-items:center;gap:8px;">
            <span style="font-size:18px;">🚀</span>
            <strong style="font-size:16px;">自定义 Header Slot</strong>
          </div>
        </template>
        <p style="margin:0;color:var(--text-secondary);line-height:1.6;">使用 #header 插槽可以完全自定义弹窗头部内容，例如添加图标或额外的操作按钮。</p>
      </AppModal>
    `,
  }),
}

export const NoTitle: Story = {
  args: {
    title: "",
  },
  render: (args) => ({
    components: { AppModal },
    setup() { return { args } },
    template: `
      <AppModal v-bind="args">
        <p style="margin:0;color:var(--text-secondary);line-height:1.6;">无标题弹窗，适用于简短的确认提示或纯内容展示。</p>
      </AppModal>
    `,
  }),
}
