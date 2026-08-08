import type { Meta, StoryObj } from "@storybook/vue3"
import AppButton from "../AppButton.vue"

const meta: Meta<typeof AppButton> = {
  title: "Components/AppButton",
  component: AppButton,
  tags: ["autodocs"],
  argTypes: {
    variant: {
      control: "select",
      options: ["primary", "secondary", "danger", "ghost"],
    },
    size: {
      control: "select",
      options: ["sm", "md", "lg"],
    },
    loading: { control: "boolean" },
    disabled: { control: "boolean" },
    block: { control: "boolean" },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  args: {
    variant: "primary",
    size: "md",
  },
  render: (args) => ({
    components: { AppButton },
    setup() { return { args } },
    template: "<AppButton v-bind=\"args\">按钮</AppButton>",
  }),
}

export const Secondary: Story = {
  ...Primary,
  args: { variant: "secondary", size: "md" },
}

export const Danger: Story = {
  ...Primary,
  args: { variant: "danger", size: "md" },
}

export const Ghost: Story = {
  ...Primary,
  args: { variant: "ghost", size: "md" },
}

export const Small: Story = {
  ...Primary,
  args: { variant: "primary", size: "sm" },
}

export const Large: Story = {
  ...Primary,
  args: { variant: "primary", size: "lg" },
}

export const Loading: Story = {
  ...Primary,
  args: { variant: "primary", loading: true },
}

export const Disabled: Story = {
  ...Primary,
  args: { variant: "primary", disabled: true },
}

export const Block: Story = {
  ...Primary,
  args: { variant: "primary", block: true },
  parameters: {
    layout: "padded",
  },
}
