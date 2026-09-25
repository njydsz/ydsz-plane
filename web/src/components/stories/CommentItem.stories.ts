import type { Meta, StoryObj } from "@storybook/vue3"
import CommentItem from "../CommentItem.vue"

const sampleComment = {
  id: 1,
  workspace_id: 1,
  project_id: 1,
  issue_id: 1,
  content_json: {},
  content_html: "<p>这个问题已经修复了，可以关闭。</p>",
  content_stripped: "这个问题已经修复了，可以关闭。",
  created_by: 100,
  creator_name: "张三",
  creator_avatar: "",
  mentions: [],
  parent_id: null,
  is_edited: false,
  edited_at: null,
  created_at: new Date(Date.now() - 3600000).toISOString(),
  updated_at: new Date(Date.now() - 3600000).toISOString(),
}

const meta: Meta<typeof CommentItem> = {
  title: "Components/CommentItem",
  component: CommentItem,
  tags: ["autodocs"],
  argTypes: {
    currentUserId: { control: "number" },
    replyHint: { control: "text" },
  },
  args: {
    comment: sampleComment,
    currentUserId: 100,
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args } },
    template: `<CommentItem v-bind="args" />`,
  }),
}

export const OthersComment: Story = {
  args: {
    currentUserId: 999,
    comment: {
      ...sampleComment,
      content_html: "<p>请 @李四 看下这个问题，可能与模块 A 有关。</p>",
      content_stripped: "请 @李四 看下这个问题，可能与模块 A 有关。",
      created_by: 200,
      creator_name: "李四",
      creator_avatar: "",
    },
  },
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args } },
    template: `<CommentItem v-bind="args" />`,
  }),
}

export const OwnComment: Story = {
  args: {
    currentUserId: 100,
    comment: {
      ...sampleComment,
      content_html: "<p>我自己发的评论，可以编辑和删除。</p>",
      content_stripped: "我自己发的评论，可以编辑和删除。",
      created_by: 100,
      creator_name: "我",
    },
  },
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args } },
    template: `<CommentItem v-bind="args" />`,
  }),
}

export const Reply: Story = {
  args: {
    replyHint: "王五",
    comment: {
      ...sampleComment,
      id: 2,
      content_html: "<p>收到，我来排查一下。</p>",
      content_stripped: "收到，我来排查一下。",
      created_by: 300,
      creator_name: "赵六",
      parent_id: 1,
      created_at: new Date(Date.now() - 1800000).toISOString(),
      updated_at: new Date(Date.now() - 1800000).toISOString(),
    },
  },
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args } },
    template: `<CommentItem v-bind="args" />`,
  }),
}

export const Edited: Story = {
  args: {
    comment: {
      ...sampleComment,
      content_html: "<p>更新后的评论内容（已编辑）</p>",
      content_stripped: "更新后的评论内容（已编辑）",
      is_edited: true,
      edited_at: new Date(Date.now() - 600000).toISOString(),
    },
  },
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args } },
    template: `<CommentItem v-bind="args" />`,
  }),
}

export const WithFormattedContent: Story = {
  args: {
    comment: {
      ...sampleComment,
      content_html: `
        <p>请在代码中包含以下修复：</p>
        <pre><code>const result = await fetch(url);</code></pre>
        <p>更多信息请查看 <a href="#">文档链接</a>。</p>
      `,
      content_stripped: "请在代码中包含以下修复：const result = await fetch(url); 更多信息请查看文档链接。",
      created_by: 100,
      creator_name: "张三",
    },
  },
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args } },
    template: `<CommentItem v-bind="args" />`,
  }),
}

export const WithAvatarImage: Story = {
  args: {
    comment: {
      ...sampleComment,
      creator_name: "Alice",
      creator_avatar: "https://i.pravatar.cc/150?img=5",
    },
  },
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args } },
    template: `<CommentItem v-bind="args" />`,
  }),
}
