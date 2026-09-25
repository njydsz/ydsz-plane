import type { Meta, StoryObj } from "@storybook/vue3"
import MiniGantt from "../MiniGantt.vue"

const meta: Meta<typeof MiniGantt> = {
  title: "Components/MiniGantt",
  component: MiniGantt,
  tags: ["autodocs"],
  argTypes: {
    sprints: { control: "object" },
    versionStart: { control: "text" },
    versionEnd: { control: "text" },
  },
  args: {
    sprints: [
      { id: 1, name: "Sprint 1", startDate: "2025-01-06", endDate: "2025-01-20", progress: 100, status: "completed" },
      { id: 2, name: "Sprint 2", startDate: "2025-01-21", endDate: "2025-02-03", progress: 75, status: "active" },
      { id: 3, name: "Sprint 3", startDate: "2025-02-04", endDate: "2025-02-17", progress: 0, status: "planned" },
    ],
    versionStart: "2025-01-06",
    versionEnd: "2025-02-17",
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  render: (args) => ({
    components: { MiniGantt },
    setup() { return { args } },
    template: `<MiniGantt v-bind="args" />`,
  }),
}

export const SingleSprint: Story = {
  args: {
    sprints: [
      { id: 1, name: "当前迭代", startDate: "2025-03-01", endDate: "2025-03-14", progress: 45, status: "active" },
    ],
    versionStart: "2025-03-01",
    versionEnd: "2025-03-14",
  },
  render: (args) => ({
    components: { MiniGantt },
    setup() { return { args } },
    template: `<MiniGantt v-bind="args" />`,
  }),
}

export const ManySprints: Story = {
  args: {
    sprints: [
      { id: 1, name: "v1.0 基础功能", startDate: "2025-01-01", endDate: "2025-01-28", progress: 100, status: "completed" },
      { id: 2, name: "v1.5 增强", startDate: "2025-01-29", endDate: "2025-02-25", progress: 100, status: "completed" },
      { id: 3, name: "v2.0 重构", startDate: "2025-02-26", endDate: "2025-03-25", progress: 60, status: "active" },
      { id: 4, name: "v2.5 优化", startDate: "2025-03-26", endDate: "2025-04-22", progress: 0, status: "planned" },
      { id: 5, name: "v3.0 规划", startDate: "2025-04-23", endDate: "2025-05-20", progress: 0, status: "planned" },
    ],
    versionStart: "2025-01-01",
    versionEnd: "2025-05-20",
  },
  render: (args) => ({
    components: { MiniGantt },
    setup() { return { args } },
    template: `<MiniGantt v-bind="args" />`,
  }),
}

export const WithoutVersionMarker: Story = {
  args: {
    sprints: [
      { id: 1, name: "Sprint A", startDate: "2025-03-01", endDate: "2025-03-14", progress: 50 },
      { id: 2, name: "Sprint B", startDate: "2025-03-15", endDate: "2025-03-28", progress: 20 },
    ],
    versionStart: "",
    versionEnd: "",
  },
  render: (args) => ({
    components: { MiniGantt },
    setup() { return { args } },
    template: `<MiniGantt v-bind="args" />`,
  }),
}

export const Empty: Story = {
  args: {
    sprints: [],
    versionStart: "",
    versionEnd: "",
  },
  render: (args) => ({
    components: { MiniGantt },
    setup() { return { args } },
    template: `<MiniGantt v-bind="args" />`,
  }),
}

export const WithoutProgress: Story = {
  args: {
    sprints: [
      { id: 1, name: "计划阶段", startDate: "2025-04-01", endDate: "2025-04-15" },
      { id: 2, name: "开发阶段", startDate: "2025-04-16", endDate: "2025-05-15" },
    ],
    versionStart: "",
    versionEnd: "",
  },
  render: (args) => ({
    components: { MiniGantt },
    setup() { return { args } },
    template: `<MiniGantt v-bind="args" />`,
  }),
}
