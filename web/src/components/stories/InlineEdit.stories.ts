import type { Meta, StoryObj } from "@storybook/vue3"
import InlineEdit from "../InlineEdit.vue"

const meta: Meta<typeof InlineEdit> = {
  title: "Components/InlineEdit",
  component: InlineEdit,
  tags: ["autodocs"],
  argTypes: {
    modelValue: { control: "text" },
    trigger: { control: "select", options: ["click", "dblclick"] },
    placeholder: { control: "text" },
    maxLength: { control: "number" },
    disabled: { control: "boolean" },
    align: { control: "select", options: ["left", "right", "center"] },
  },
  args: {
    modelValue: "点击即可编辑这段文字",
    trigger: "click",
    placeholder: "点击编辑",
    maxLength: 200,
    disabled: false,
    align: "left",
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  render: (args) => ({
    components: { InlineEdit },
    setup() { return { args } },
    template: `
      <div style="padding:16px;">
        <InlineEdit v-bind="args" />
      </div>
    `,
  }),
}

export const Empty: Story = {
  args: {
    modelValue: "",
    placeholder: "点击添加内容...",
  },
  render: (args) => ({
    components: { InlineEdit },
    setup() { return { args } },
    template: `
      <div style="padding:16px;">
        <InlineEdit v-bind="args" />
      </div>
    `,
  }),
}

export const DblClickTrigger: Story = {
  args: {
    modelValue: "双击我触发编辑",
    trigger: "dblclick",
  },
  render: (args) => ({
    components: { InlineEdit },
    setup() { return { args } },
    template: `
      <div style="padding:16px;">
        <InlineEdit v-bind="args" />
      </div>
    `,
  }),
}

export const Disabled: Story = {
  args: {
    modelValue: "禁用状态，无法编辑",
    disabled: true,
  },
  render: (args) => ({
    components: { InlineEdit },
    setup() { return { args } },
    template: `
      <div style="padding:16px;">
        <InlineEdit v-bind="args" />
      </div>
    `,
  }),
}

export const Centered: Story = {
  args: {
    modelValue: "居中对齐",
    align: "center",
  },
  render: (args) => ({
    components: { InlineEdit },
    setup() { return { args } },
    template: `
      <div style="padding:16px;text-align:center;">
        <InlineEdit v-bind="args" />
      </div>
    `,
  }),
}

export const RightAligned: Story = {
  args: {
    modelValue: "右对齐",
    align: "right",
  },
  render: (args) => ({
    components: { InlineEdit },
    setup() { return { args } },
    template: `
      <div style="padding:16px;text-align:right;">
        <InlineEdit v-bind="args" />
      </div>
    `,
  }),
}

export const WithValidation: Story = {
  args: {
    modelValue: "不能输入空内容",
    maxLength: 20,
  },
  render: (args) => ({
    components: { InlineEdit },
    setup() {
      return {
        args,
        validate: (val: string) => val.length === 0 ? "内容不能为空" : null,
      }
    },
    template: `
      <div style="padding:16px;">
        <InlineEdit v-bind="args" :validate="validate" />
      </div>
    `,
  }),
}
