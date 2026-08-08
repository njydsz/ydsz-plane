/**
 * empty-state-copy.ts — 空状态文案与插画的集中管理库。
 *
 * 目的：
 *  1. 保证空状态文案在全站风格一致、语气统一
 *  2. 方便本地化 / 品牌调性迭代
 *  3. AppEmptyState 组件通过此库自动匹配标题与描述
 *
 * 文案风格：
 *  - 标题简短一句话（通常 6-10 字）
 *  - 描述一句引导性陈述 + 可操作建议
 *  - 避免"空""无"等负面词汇堆砌
 */

export type EmptyScenario =
  | "default"
  | "issues"
  | "projects"
  | "sprints"
  | "modules"
  | "search"
  | "intake"
  | "notifications"
  | "labels"
  | "members"
  | "analytics"
  | "views"
  | "inbox"
  | "api-token"
  | "webhooks"
  | "error"
  | "gantt"
  | "calendar"
  | "pages"
  | "cycles"
  | "automation"
  | "comments"
  | "versions"
  | "workflows";

/** 插画色彩主题名（与 IllustrationTheme token 对齐） */
export type IllustrationPalette = "brand" | "success" | "warning" | "neutral";

export interface EmptyStateCopy {
  title: string;
  description: string;
  /** 主题色偏好（用于 SVG 插画取色） */
  palette: IllustrationPalette;
  /** 主 CTA 文案（可选，供外部直接使用） */
  ctaLabel?: string;
}

const copy: Record<Exclude<EmptyScenario, "default" | "error">, EmptyStateCopy> = {
  issues: {
    title: "还没有工作项",
    description: "创建第一个需求、任务或缺陷，开始管理你的项目进度。",
    palette: "brand",
    ctaLabel: "创建工作项",
  },
  projects: {
    title: "还没有项目",
    description: "创建一个项目来组织工作项，与团队协作推进。",
    palette: "brand",
    ctaLabel: "创建项目",
  },
  sprints: {
    title: "还没有迭代",
    description: "规划你的第一个冲刺，设定目标和时间范围。",
    palette: "warning",
    ctaLabel: "创建迭代",
  },
  modules: {
    title: "还没有模块",
    description: "将复杂项目拆分为更小的模块，便于分工与管理。",
    palette: "neutral",
    ctaLabel: "创建模块",
  },
  search: {
    title: "未找到匹配结果",
    description: "尝试调整搜索关键字或筛选条件。",
    palette: "neutral",
  },
  intake: {
    title: "收件箱是空的",
    description: "外部提交的工作项会出现在这里，等待你审核与转正。",
    palette: "brand",
  },
  notifications: {
    title: "暂无通知",
    description: "当有新的工作项、评论或提及时，你会收到通知。",
    palette: "neutral",
  },
  labels: {
    title: "还没有标签",
    description: "创建标签来分类和标记工作项，方便筛选。",
    palette: "brand",
    ctaLabel: "创建标签",
  },
  members: {
    title: "还没有成员",
    description: "邀请团队成员加入项目，共同协作。",
    palette: "success",
    ctaLabel: "邀请成员",
  },
  analytics: {
    title: "暂无数据可分析",
    description: "当项目有进度变更后，这里会显示效能趋势。",
    palette: "brand",
  },
  views: {
    title: "还没有自定义视图",
    description: "保存当前的筛选和分组条件，创建专属视图。",
    palette: "neutral",
    ctaLabel: "创建视图",
  },
  inbox: {
    title: "收件箱是空的",
    description: "所有外部提交的工作项会显示在此。",
    palette: "neutral",
  },
  "api-token": {
    title: "还没有 API 令牌",
    description: "创建令牌以通过 API 访问工作空间数据。",
    palette: "neutral",
    ctaLabel: "生成令牌",
  },
  webhooks: {
    title: "还没有 Webhook",
    description: "配置 Webhook 将事件推送到外部系统。",
    palette: "brand",
    ctaLabel: "添加 Webhook",
  },
  gantt: {
    title: "甘特图暂无数据",
    description: "为工作项设置开始和结束时间后，这里将展示甘特视图。",
    palette: "brand",
  },
  calendar: {
    title: "日历暂无安排",
    description: "设置工作项的截止日期，日程将显示在日历上。",
    palette: "brand",
  },
  pages: {
    title: "还没有文档",
    description: "创建项目文档、设计笔记或团队知识库。",
    palette: "brand",
    ctaLabel: "创建文档",
  },
  cycles: {
    title: "还没有周期",
    description: "设置固定长度的时间盒，迭代推进项目。",
    palette: "brand",
    ctaLabel: "创建周期",
  },
  automation: {
    title: "还没有自动化规则",
    description: "创建规则以在特定触发条件下自动执行操作。",
    palette: "brand",
    ctaLabel: "添加规则",
  },
  comments: {
    title: "暂无评论",
    description: "开始对话，评论将显示在此。",
    palette: "neutral",
  },
  versions: {
    title: "还没有版本",
    description: "发布版本来标记重要的里程碑。",
    palette: "success",
    ctaLabel: "创建版本",
  },
  workflows: {
    title: "还没有工作流",
    description: "设计审批流或阶段流程以规范工作项推进。",
    palette: "brand",
    ctaLabel: "创建工作流",
  },
};

/** 获取空状态文案（返回副本以便调用方自由修改） */
export function getEmptyCopy(scenario: EmptyScenario): EmptyStateCopy {
  if (scenario === "error") {
    return { title: "出错了", description: "加载数据时遇到问题，请稍后重试或联系管理员。", palette: "neutral" };
  }
  if (scenario === "default") {
    return { title: "暂无数据", description: "", palette: "neutral" };
  }
  const c = copy[scenario];
  return c ? { ...c } : { title: "暂无数据", description: "", palette: "neutral" };
}

export default copy;
