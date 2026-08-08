/**
 * 应用入口 — 创建 Vue 应用实例并装配 Pinia、路由、i18n、PWA，挂载到 #app。
 *
 * 装配顺序（遵循 Vue 插件最佳实践）:
 *   1. Pinia（状态管理）
 *   2. vue-i18n（国际化）
 *   3. Router（路由）
 *   4. PWA Service Worker 注册
 *   5. 全局 CSS 设计令牌
 */
import { createPinia } from "pinia";
import { createApp } from "vue";

import App from "./App.vue";
import router from "./router";
import { i18n } from "./locales";
import { registerSW } from "./pwa";

import "./design/tokens.css";
import { initTheme } from "./lib/theme";
import { vPermission } from "./directives/permission";
import Permission from "./components/Permission.vue";

// 在挂载前应用主题，避免首帧闪烁（FOUC）
initTheme();

const app = createApp(App);
app.use(createPinia());
app.use(i18n);
app.use(router);
// 全局自定义指令 + 组件
app.directive("permission", vPermission);
app.component("Permission", Permission);
app.mount("#app");

// 注册 Service Worker（PWA 离线支持）
registerSW();
