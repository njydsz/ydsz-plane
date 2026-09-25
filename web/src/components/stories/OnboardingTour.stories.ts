import type { Meta, StoryObj } from "@storybook/vue3"
import OnboardingTour from "../OnboardingTour.vue"

const meta: Meta<typeof OnboardingTour> = {
  title: "Components/OnboardingTour",
  component: OnboardingTour,
  tags: ["autodocs"],
  argTypes: {
    workspaceId: { control: "text" },
    workspaceName: { control: "text" },
  },
  args: {
    workspaceId: "demo-workspace",
    workspaceName: "示例工作空间",
  },
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  render: (args) => ({
    components: { OnboardingTour },
    setup() { return { args } },
    template: `<OnboardingTour v-bind="args" />`,
  }),
}

export const CustomWorkspace: Story = {
  args: {
    workspaceName: "移动端团队",
  },
  render: (args) => ({
    components: { OnboardingTour },
    setup() { return { args } },
    template: `<OnboardingTour v-bind="args" />`,
  }),
}

export const NoWorkspaceName: Story = {
  args: {
    workspaceName: "",
  },
  render: (args) => ({
    components: { OnboardingTour },
    setup() { return { args } },
    template: `<OnboardingTour v-bind="args" />`,
  }),
}

export const StepProgress: Story = {
  render: (args) => ({
    components: { OnboardingTour },
    setup() { return { args } },
    template: `
      <div>
        <p style="padding:16px;font-size:13px;color:var(--text-tertiary);margin:0;">
          引导流程共 5 步。下方展示第一步效果，用户可点击「上一步」「下一步」「跳过引导」进行导航。
        </p>
        <OnboardingTour v-bind="args" />
      </div>
    `,
  }),
}
