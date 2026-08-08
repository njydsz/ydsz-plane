import type { Meta, StoryObj } from "@storybook/vue3"
import AppBadge from "../AppBadge.vue"

const meta: Meta<typeof AppBadge> = {
  title: "Components/AppBadge",
  component: AppBadge,
  tags: ["autodocs"],
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "brand", "success", "warning", "danger", "info"],
    },
    dot: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: (args) => ({
    components: { AppBadge },
    setup() { return { args } },
    template: "<AppBadge v-bind=\"args\">默认</AppBadge>",
  }),
}

export const Brand: Story = {
  args: { variant: "brand" },
  render: (args) => ({
    components: { AppBadge },
    setup() { return { args } },
    template: "<AppBadge v-bind=\"args\">品牌色</AppBadge>",
  }),
}

export const Success: Story = {
  args: { variant: "success" },
  render: (args) => ({
    components: { AppBadge },
    setup() { return { args } },
    template: "<AppBadge v-bind=\"args\">成功</AppBadge>",
  }),
}

export const Warning: Story = {
  args: { variant: "warning" },
  render: (args) => ({
    components: { AppBadge },
    setup() { return { args } },
    template: "<AppBadge v-bind=\"args\">警告</AppBadge>",
  }),
}

export const Danger: Story = {
  args: { variant: "danger" },
  render: (args) => ({
    components: { AppBadge },
    setup() { return { args } },
    template: "<AppBadge v-bind=\"args\">危险</AppBadge>",
  }),
}

export const Info: Story = {
  args: { variant: "info" },
  render: (args) => ({
    components: { AppBadge },
    setup() { return { args } },
    template: "<AppBadge v-bind=\"args\">信息</AppBadge>",
  }),
}

export const DotOnly: Story = {
  args: { variant: "success", dot: true },
  render: (args) => ({
    components: { AppBadge },
    setup() { return { args } },
    template: '<AppBadge v-bind="args">在线</AppBadge>',
  }),
}
