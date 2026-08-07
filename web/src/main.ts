/**
 * 应用入口 — 创建 Vue 应用实例并装配 Pinia、路由、i18n，挂载到 #app。
 *
 * 装配顺序（遵循 Vue 插件最佳实践）:
 *   1. Pinia（状态管理）
 *   2. vue-i18n（国际化）
 *   3. Router（路由）
 *   4. 全局 CSS 设计令牌
 */
import { createPinia } from "pinia";
import { createApp } from "vue";

import App from "./App.vue";
import router from "./router";
import { i18n } from "./locales";

import "./design/tokens.css";

const app = createApp(App);
app.use(createPinia());
app.use(i18n);
app.use(router);
app.mount("#app");
