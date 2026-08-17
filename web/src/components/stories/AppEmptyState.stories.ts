import type { Meta, StoryObj } from "@storybook/vue3"
import AppEmptyState from "../AppEmptyState.vue"

const meta: Meta<typeof AppEmptyState> = {
  title: "Components/AppEmptyState",
  component: AppEmptyState,
  tags: ["autodocs"],
  argTypes: {
    scenario: {
      control: "select",
      options: ["default", "issues", "projects", "sprints", "modules", "search", "intake", "notifications", "labels", "members", "analytics", "views", "inbox", "api-token", "webhooks", "error"],
    },
    illustrationSize: {
      control: "select",
      options: ["sm", "md", "lg"],
    },
    title: { control: "text" },
    description: { control: "text" },
    icon: { control: "text" },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const IssuesEmpty: Story = {
  args: {
    scenario: "issues",
  },
}

export const ProjectsEmpty: Story = {
  args: {
    scenario: "projects",
  },
}

export const SprintsEmpty: Story = {
  args: {
    scenario: "sprints",
  },
}

export const SearchEmpty: Story = {
  args: {
    scenario: "search",
  },
}

export const NotificationsEmpty: Story = {
  args: {
    scenario: "notifications",
  },
}

export const Custom: Story = {
  args: {
    icon: "🚀",
    title: "开始你的旅程",
    description: "创建一个新项目来管理任务和迭代。",
  },
}

export const WithActions: Story = {
  render: () => ({
    components: { AppEmptyState },
    template: `
      <AppEmptyState scenario="issues">
        <template #cta>
          <button style="padding:6px 14px;border-radius:6px;background:#3b82f6;color:#fff;border:none;cursor:pointer;font:inherit;font-size:13px;font-weight:500;">创建需求/任务/缺陷</button>
        </template>
        <template #secondary>
          <button style="padding:6px 14px;border-radius:6px;background:transparent;color:#6b7280;border:1px solid #d1d5db;cursor:pointer;font:inherit;font-size:13px;">导入数据</button>
        </template>
      </AppEmptyState>
    `,
  }),
}

export const ErrorState: Story = {
  args: {
    scenario: "error",
  },
}

export const Small: Story = {
  args: {
    scenario: "issues",
    illustrationSize: "sm",
  },
}

export const Large: Story = {
  args: {
    scenario: "projects",
    illustrationSize: "lg",
  },
}
