import type { Meta, StoryObj } from "@storybook/vue3"
import CommentItem from "../CommentItem.vue"

/**
 * CommentList 依赖 Pinia store 和 API 调用，无法在隔离 story 中独立渲染。
 * 此处使用 CommentItem 子组件展示评论列表的完整视觉效果，
 * 包括嵌套回复、编辑态提示等。
 */

const sampleComments = [
  {
    id: 1,
    workspace_id: 1,
    project_id: 1,
    issue_id: 1,
    content_json: {},
    content_html: "<p>用户反馈登录失败，请排查。</p>",
    content_stripped: "用户反馈登录失败，请排查。",
    created_by: 200,
    creator_name: "张三",
    creator_avatar: "",
    mentions: [],
    parent_id: null,
    is_edited: false,
    edited_at: null,
    created_at: new Date(Date.now() - 7200000).toISOString(),
    updated_at: new Date(Date.now() - 7200000).toISOString(),
  },
  {
    id: 2,
    workspace_id: 1,
    project_id: 1,
    issue_id: 1,
    content_json: {},
    content_html: "<p>收到，正在排查，初步判断是 SSO 服务超时。</p>",
    content_stripped: "收到，正在排查，初步判断是 SSO 服务超时。",
    created_by: 300,
    creator_name: "李四",
    creator_avatar: "",
    mentions: [],
    parent_id: 1,
    is_edited: false,
    edited_at: null,
    created_at: new Date(Date.now() - 5400000).toISOString(),
    updated_at: new Date(Date.now() - 5400000).toISOString(),
  },
  {
    id: 3,
    workspace_id: 1,
    project_id: 1,
    issue_id: 1,
    content_json: {},
    content_html: "<p>SSO 侧确认是证书过期导致，已更新证书。请验证。</p>",
    content_stripped: "SSO 侧确认是证书过期导致，已更新证书。请验证。",
    created_by: 200,
    creator_name: "张三",
    creator_avatar: "",
    mentions: [],
    parent_id: 1,
    is_edited: false,
    edited_at: null,
    created_at: new Date(Date.now() - 3600000).toISOString(),
    updated_at: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: 4,
    workspace_id: 1,
    project_id: 1,
    issue_id: 1,
    content_json: {},
    content_html: "<p>验证通过，已关闭。感谢排查！</p>",
    content_stripped: "验证通过，已关闭。感谢排查！",
    created_by: 100,
    creator_name: "我",
    creator_avatar: "",
    mentions: [],
    parent_id: 2,
    is_edited: true,
    edited_at: new Date(Date.now() - 3000000).toISOString(),
    created_at: new Date(Date.now() - 1800000).toISOString(),
    updated_at: new Date(Date.now() - 1800000).toISOString(),
  },
]

const meta: Meta<typeof CommentItem> = {
  title: "Components/CommentList",
  component: CommentItem,
  tags: ["autodocs"],
  parameters: {
    docs: {
      description: {
        component:
          "CommentList 依赖 Pinia store 和 API 接口，此 Story 使用 CommentItem 模拟评论列表的完整视觉效果（含嵌套回复）。",
      },
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

export const ThreadWithReplies: Story = {
  render: () => ({
    components: { CommentItem },
    setup() {
      return { comments: sampleComments, currentUserId: 100 }
    },
    template: `
      <div style="max-width:600px;">
        <div style="font-size:14px;font-weight:600;color:var(--text-primary);margin-bottom:16px;">
          评论 <span style="font-size:12px;font-weight:500;color:var(--text-tertiary);background:var(--surface-3);padding:1px 7px;border-radius:10px;">{{ comments.length }}</span>
        </div>
        <CommentItem :comment="comments[0]" :current-user-id="currentUserId" />
        <div style="margin-left:36px;padding-left:12px;border-left:2px solid var(--border-subtle);">
          <CommentItem :comment="comments[1]" :current-user-id="currentUserId" reply-hint="张三" />
          <div style="margin-left:36px;padding-left:12px;border-left:2px solid var(--border-subtle);">
            <CommentItem :comment="comments[3]" :current-user-id="currentUserId" reply-hint="李四" />
          </div>
          <CommentItem :comment="comments[2]" :current-user-id="currentUserId" reply-hint="张三" />
        </div>
      </div>
    `,
  }),
}

export const SingleComment: Story = {
  render: () => ({
    components: { CommentItem },
    setup() {
      return { comment: sampleComments[0], currentUserId: 100 }
    },
    template: `
      <div style="max-width:600px;">
        <div style="font-size:14px;font-weight:600;color:var(--text-primary);margin-bottom:16px;">评论 <span style="font-size:12px;font-weight:500;color:var(--text-tertiary);background:var(--surface-3);padding:1px 7px;border-radius:10px;">1</span></div>
        <CommentItem :comment="comment" :current-user-id="currentUserId" />
      </div>
    `,
  }),
}

export const CompactList: Story = {
  render: () => ({
    components: { CommentItem },
    setup() {
      return { comments: sampleComments.slice(0, 2), currentUserId: 100 }
    },
    template: `
      <div style="max-width:500px;">
        <CommentItem v-for="c in comments" :key="c.id" :comment="c" :current-user-id="currentUserId" />
        <div style="text-align:center;padding:12px;font-size:12px;color:var(--text-tertiary);">
          还有 {{ comments.length - 2 }} 条评论...
        </div>
      </div>
    `,
  }),
}

export const DeeplyNested: Story = {
  render: () => ({
    components: { CommentItem },
    setup() {
      return { comments: sampleComments, currentUserId: 100 }
    },
    template: `
      <div style="max-width:600px;">
        <CommentItem :comment="comments[0]" :current-user-id="currentUserId" />
        <div style="margin-left:36px;padding-left:12px;border-left:2px solid var(--border-subtle);">
          <CommentItem :comment="comments[1]" :current-user-id="currentUserId" reply-hint="张三" />
        </div>
        <div style="margin-left:72px;padding-left:12px;border-left:2px solid var(--border-subtle);">
          <CommentItem :comment="comments[3]" :current-user-id="currentUserId" reply-hint="李四" />
        </div>
        <div style="margin-left:36px;padding-left:12px;border-left:2px solid var(--border-subtle);">
          <CommentItem :comment="comments[2]" :current-user-id="currentUserId" reply-hint="张三" />
        </div>
      </div>
    `,
  }),
}

export const OnlyReplies: Story = {
  args: {
    comment: sampleComments[3],
  },
  render: (args) => ({
    components: { CommentItem },
    setup() { return { args, currentUserId: 100 } },
    template: `
      <div style="max-width:600px;margin-left:36px;padding-left:12px;border-left:2px solid var(--border-subtle);">
        <CommentItem v-bind="args" :current-user-id="currentUserId" reply-hint="李四" />
      </div>
    `,
  }),
}
