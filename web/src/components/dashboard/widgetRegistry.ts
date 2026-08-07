/**
 * Widget 注册表 —— 将 widget_type 映射到渲染组件和中文名。
 *
 * 新增 widget 类型时只需在此文件追加映射即可。
 */
import type { Component } from "vue";

import ProgressOverviewWidget from "./ProgressOverviewWidget.vue";
import PieChartWidget from "./PieChartWidget.vue";
import BarChartWidget from "./BarChartWidget.vue";
import LineChartWidget from "./LineChartWidget.vue";
import StackedBarWidget from "./StackedBarWidget.vue";
import ListTableWidget from "./ListTableWidget.vue";
import ActivityTimelineWidget from "./ActivityTimelineWidget.vue";
import RiskAlertWidget from "./RiskAlertWidget.vue";

import type { WidgetType } from "@/api/services/dashboard";

/** 显示名映射表：widget_type → 中文名（仅用于 UI 展示，不参与后端交互）。 */
export const WIDGET_NAME_MAP: Record<WidgetType, string> = {
  progress_overview: "进度总览",
  burndown: "燃尽图",
  velocity: "速率图",
  priority_split: "优先级分布",
  state_distribution: "状态分布",
  overdue_list: "逾期工作项",
  blocked_list: "阻塞工作项",
  risk_alert: "风险告警",
  recent_activity: "最近动态",
  team_workload: "团队负载",
};

/** Widget 类型 → 渲染组件映射（用于 DashboardView 的动态组件） */
export const WIDGET_COMPONENTS: Record<WidgetType, Component> = {
  progress_overview: ProgressOverviewWidget,
  burndown: LineChartWidget,
  velocity: BarChartWidget,
  priority_split: PieChartWidget,
  state_distribution: BarChartWidget,
  overdue_list: ListTableWidget,
  blocked_list: ListTableWidget,
  risk_alert: RiskAlertWidget,
  recent_activity: ActivityTimelineWidget,
  team_workload: StackedBarWidget,
};
