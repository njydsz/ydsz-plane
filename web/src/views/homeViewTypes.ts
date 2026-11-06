/**
 * HomeView — 快捷操作项模型。
 *
 * 从 HomeView.vue 提取为独立 .ts 文件：vue-tsc 对「在 .vue 中定义 interface
 * 并在同文件模板中使用其他 computed」存在类型推断顺序问题，
 * 提取后消除误报。
 */
export interface QuickActionItem {
  type: string;
  label: string;
  icon: string;
  route: string;
}
