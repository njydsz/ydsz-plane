import type { Meta, StoryObj } from "@storybook/vue3"
import AppAvatar from "../AppAvatar.vue"

const meta: Meta<typeof AppAvatar> = {
  title: "Components/AppAvatar",
  component: AppAvatar,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    src: { control: "text" },
    size: { control: "select", options: ["sm", "md", "lg"] },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Medium: Story = {
  args: { name: "张三", size: "md" },
}

export const Small: Story = {
  args: { name: "李四", size: "sm" },
}

export const Large: Story = {
  args: { name: "王五", size: "lg" },
}

export const WithImage: Story = {
  args: {
    name: "用户",
    src: "https://i.pravatar.cc/150?img=1",
    size: "md",
  },
}

export const Group: Story = {
  render: () => ({
    components: { AppAvatar },
    template: `
      <div style="display:flex;gap:4px;align-items:center;">
        <AppAvatar name="张三" size="sm" />
        <AppAvatar name="李四" size="sm" />
        <AppAvatar name="王五" size="sm" />
        <AppAvatar name="赵六" size="sm" />
      </div>
    `,
  }),
}
