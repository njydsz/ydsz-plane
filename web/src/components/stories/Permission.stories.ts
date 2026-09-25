/**
 * Permission 组件 Storybook 故事 — S16 P2-1。
 *
 * <Permission> 包裹组件 — 基于后端返回的权限码集合决定是否渲染插槽内容。
 * 三种用法：单一权限、多权限组合、菜单模式。
 */
import type { Meta, StoryObj } from "@storybook/vue3";
import Permission from "../Permission.vue";

const meta: Meta<typeof Permission> = {
  title: "Components/Permission",
  component: Permission,
  tags: ["autodocs"],
  argTypes: {
    permission: { control: "text", description: "单一权限码" },
    permissions: { control: "object", description: "多个权限码数组" },
    match: { control: "select", options: ["all", "any"] },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// 模拟 workspace store 权限已加载的状态
const mockPermissions = ["issue:edit_all", "issue:read", "sprint:read", "member:read"];

export const HasPermission: Story = {
  render: (args) => ({
    components: { Permission },
    setup() {
      return { args };
    },
    template: `
      <Permission v-bind="args">
        <button class="px-3 py-1 bg-blue-600 text-white rounded">编辑按钮</button>
      </Permission>
    `,
  }),
  args: {
    permission: "issue:edit_all",
  },
  parameters: {
    mockData: [{ path: "@/stores/workspace", data: { permissions: new Set(mockPermissions), canManage: true } }],
  },
};

export const NoPermission: Story = {
  render: (args) => ({
    components: { Permission },
    setup() {
      return { args };
    },
    template: `
      <Permission v-bind="args">
        <button class="px-3 py-1 bg-blue-600 text-white rounded">删除按钮</button>
      </Permission>
      <p class="text-xs text-gray-500 mt-2">上面按钮因权限不足被隐藏</p>
    `,
  }),
  args: {
    permission: "workspace:delete",
  },
};

export const MultiplePermissionsAny: Story = {
  render: (args) => ({
    components: { Permission },
    setup() {
      return { args };
    },
    template: `
      <Permission v-bind="args">
        <button class="px-3 py-1 bg-green-600 text-white rounded">审核按钮</button>
      </Permission>
    `,
  }),
  args: {
    permissions: ["issue:approve", "issue:edit_all"],
    match: "any",
  },
};

export const MultiplePermissionsAll: Story = {
  render: (args) => ({
    components: { Permission },
    setup() {
      return { args };
    },
    template: `
      <Permission v-bind="args">
        <button class="px-3 py-1 bg-red-600 text-white rounded">高危操作</button>
      </Permission>
    `,
  }),
  args: {
    permissions: ["workspace:admin", "member:change_role"],
    match: "all",
  },
};
