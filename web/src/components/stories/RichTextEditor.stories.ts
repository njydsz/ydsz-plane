import { defineComponent } from "vue";
import type { Meta, StoryObj } from "@storybook/vue3"

/**
 * RichTextEditor 依赖 TipTap 3 vue 扩展 (useEditor)、Pinia 和多个扩展包，
 * 直接在 Storybook 中加载会因为缺少编辑器扩展完整初始化链而报错。
 * 此处通过 render 函数模拟编辑器外观，集中展示各 toolbar 变体与内容状态，
 * 让读者可以预览各类编辑器布局和交互元素。
 */

const EditorShell = defineComponent({
  props: {
    variant: { type: String, default: "full" },
    showToolbar: { type: Boolean, default: true },
    minHeight: { type: String, default: "120px" },
    readonly: { type: Boolean, default: false },
    content: { type: String, default: "" },
    placeholder: { type: String, default: "输入内容..." },
  },
  template: `
    <div
      class="rich-editor-mock"
      :class="{
        'rich-editor-mock--readonly': readonly,
      }"
      :style="{
        border: readonly ? 'none' : '1px solid var(--border-default, #d1d5db)',
        borderRadius: '8px',
        overflow: 'hidden',
        background: 'var(--surface-1, #fff)',
      }"
    >
      <div
        v-if="showToolbar && variant !== 'compact'"
        style="display:flex;align-items:center;gap:2px;padding:6px 8px;background:var(--surface-2);border-bottom:1px solid var(--border-subtle);flex-wrap:wrap;"
      >
        <template v-if="variant === 'full'">
          <span style="font-size:12px;font-weight:600;color:#4b5563;padding:0 4px;">B</span>
          <span style="font-size:12px;font-style:italic;color:#4b5563;padding:0 4px;">I</span>
          <span style="font-size:12px;text-decoration:underline;color:#4b5563;padding:0 4px;">U</span>
          <span style="font-size:12px;text-decoration:line-through;color:#4b5563;padding:0 4px;">S</span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;font-family:monospace;">{"{ }"}</span>
          <span style="width:1px;height:18px;background:var(--border-subtle);margin:0 4px;"></span>
          <span style="font-size:12px;font-weight:600;color:#4b5563;padding:0 4px;">H2</span>
          <span style="font-size:12px;font-weight:600;color:#4b5563;padding:0 4px;">H3</span>
          <span style="width:1px;height:18px;background:var(--border-subtle);margin:0 4px;"></span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;">•≡</span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;">1≡</span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;">☑</span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;">❝</span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;">&lt;/&gt;</span>
          <span style="width:1px;height:18px;background:var(--border-subtle);margin:0 4px;"></span>
          <span style="font-size:14px;padding:0 2px;">😊</span>
          <span style="width:1px;height:18px;background:var(--border-subtle);margin:0 4px;"></span>
          <span style="font-size:14px;padding:0 2px;">💡</span>
          <span style="font-size:14px;padding:0 2px;">⚠</span>
          <span style="font-size:14px;padding:0 2px;">🚫</span>
          <span style="font-size:14px;padding:0 2px;">✅</span>
          <span style="width:1px;height:18px;background:var(--border-subtle);margin:0 4px;"></span>
          <span style="font-size:14px;padding:0 2px;">🔗</span>
          <span style="font-size:14px;padding:0 2px;">🖼</span>
          <span style="font-size:14px;padding:0 2px;">⊞</span>
        </template>
        <template v-else>
          <span style="font-size:12px;font-weight:600;color:#4b5563;padding:0 4px;">B</span>
          <span style="font-size:12px;font-style:italic;color:#4b5563;padding:0 4px;">I</span>
          <span style="font-size:12px;text-decoration:underline;color:#4b5563;padding:0 4px;">U</span>
          <span style="font-size:12px;text-decoration:line-through;color:#4b5563;padding:0 4px;">S</span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;font-family:monospace;">{"{ }"}</span>
          <span style="width:1px;height:18px;background:var(--border-subtle);margin:0 4px;"></span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;">•≡</span>
          <span style="font-size:12px;color:#4b5563;padding:0 4px;">1≡</span>
          <span style="width:1px;height:18px;background:var(--border-subtle);margin:0 4px;"></span>
          <span style="font-size:14px;padding:0 2px;">🔗</span>
        </template>
      </div>
      <div style="padding:12px;" :style="{ minHeight }">
        <p
          v-if="content"
          style="margin:0;font-size:14px;line-height:1.6;color:var(--text-primary);"
          v-html="content"
        ></p>
        <p
          v-else-if="!readonly"
          style="margin:0;font-size:14px;color:var(--text-tertiary);"
        >
          {{ placeholder }}
        </p>
      </div>
    </div>
  `,
})

const meta: Meta<typeof EditorShell> = {
  title: "Components/RichTextEditor",
  tags: ["autodocs"],
  parameters: {
    docs: {
      description: {
        component:
          "基于 TipTap 3 的富文本编辑器。支持粗体/斜体/链接/图片/代码/标题/列表/表格/颜色/Callout 等工具栏，以及 @提及、Slash 命令和 AI 辅助功能（续写/改写/纠错）。此 Story 为视觉模拟，实际交互需在应用内运行。",
      },
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  render: (args) => ({
    components: { EditorShell },
    setup() { return { args } },
    template: `<EditorShell v-bind="args" />`,
  }),
  args: {
    variant: "full",
    showToolbar: true,
    placeholder: "输入内容...",
    minHeight: "120px",
  },
}

export const CommentVariant: Story = {
  render: (args) => ({
    components: { EditorShell },
    setup() { return { args } },
    template: `<EditorShell v-bind="args" />`,
  }),
  args: {
    variant: "comment",
    showToolbar: true,
    placeholder: "写一条评论...",
    minHeight: "80px",
  },
}

export const CompactVariant: Story = {
  render: (args) => ({
    components: { EditorShell },
    setup() { return { args } },
    template: `<EditorShell v-bind="args" />`,
  }),
  args: {
    variant: "compact",
    showToolbar: false,
    placeholder: "简短描述...",
    minHeight: "40px",
  },
}

export const ReadOnly: Story = {
  render: (args) => ({
    components: { EditorShell },
    setup() { return { args } },
    template: `<EditorShell v-bind="args" />`,
  }),
  args: {
    variant: "full",
    showToolbar: false,
    readonly: true,
    minHeight: "120px",
    content: `
      <p>这是一篇<strong>已保存的需求描述</strong>，包含以下内容：</p>
      <h2>背景</h2>
      <p>用户希望在系统中实现<em>多租户数据隔离</em>，确保各组织数据安全。</p>
      <h3>技术方案</h3>
      <ul>
        <li>在 PostgREST 层启用 Row Level Security</li>
        <li>每个请求携带 workspace_id 上下文</li>
        <li>使用 tRPC 联邦网关自动注入过滤条件</li>
      </ul>
      <blockquote>参考：<a href="#">多租户架构设计文档</a><br/><code>auth.tenant_id = current_tenant()</code></blockquote>
      <pre><code>const result = await ctx.fetch('/api/items', { headers: { 'x-tenant': tenantId } });</code></pre>
    `,
  },
}

export const WithCallout: Story = {
  render: (args) => ({
    components: { EditorShell },
    setup() { return { args } },
    template: `
      <div>
        <EditorShell v-bind="args" />
        <div style="margin-top:12px;display:flex;gap:8px;flex-direction:column;">
          <div style="border-radius:8px;padding:12px 16px;border-left:3px solid #3b82f6;background:#eef2ff;display:flex;gap:8px;">
            <span>💡</span>
            <span style="font-size:13px;color:#1e40af;">这是一条信息提示框，用于强调重要说明。</span>
          </div>
          <div style="border-radius:8px;padding:12px 16px;border-left:3px solid #ef4444;background:#fef2f2;display:flex;gap:8px;">
            <span>🚫</span>
            <span style="font-size:13px;color:#991b1b;">错误提示框用于警告用户危险操作或不可恢复的状态。</span>
          </div>
        </div>
      </div>
    `,
  }),
  args: {
    variant: "full",
    showToolbar: true,
    readonly: true,
    content: "<p>编辑器内容中带有多条 Callout 提示框。</p>",
  },
}

export const WithMention: Story = {
  render: (args) => ({
    components: { EditorShell },
    setup() { return { args } },
    template: `
      <div>
        <EditorShell v-bind="args" />
        <div style="margin-top:16px;border:1px solid var(--border-subtle);border-radius:8px;padding:12px;background:var(--surface-1);box-shadow:0 4px 16px rgba(0,0,0,0.1);width:240px;">
          <div style="padding:6px 10px;font-size:11px;font-weight:600;color:var(--text-tertiary);text-transform:uppercase;letter-spacing:0.5px;">提及用户</div>
          <div style="display:flex;align-items:center;gap:8px;padding:8px 10px;border-radius:6px;background:var(--brand-50);cursor:pointer;">
            <div style="width:28px;height:28px;border-radius:50%;background:#3b82f6;display:flex;align-items:center;justify-content:center;color:#fff;font-size:12px;font-weight:600;">张</div>
            <span style="font-size:13px;color:var(--text-primary);font-weight:500;">张三</span>
          </div>
          <div style="display:flex;align-items:center;gap:8px;padding:8px 10px;border-radius:6px;cursor:pointer;">
            <div style="width:28px;height:28px;border-radius:50%;background:#10b981;display:flex;align-items:center;justify-content:center;color:#fff;font-size:12px;font-weight:600;">李</div>
            <span style="font-size:13px;color:var(--text-primary);font-weight:500;">李四</span>
          </div>
        </div>
      </div>
    `,
  }),
  args: {
    variant: "full",
    showToolbar: true,
    readonly: true,
    content: "<p>请 @张三 和 @李四 共同评审这个需求。</p>",
  },
}

export const EmptyState: Story = {
  render: (args) => ({
    components: { EditorShell },
    setup() { return { args } },
    template: `<EditorShell v-bind="args" />`,
  }),
  args: {
    variant: "full",
    showToolbar: true,
    placeholder: "输入需求描述...",
    content: "",
  },
}
