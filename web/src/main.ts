/**
 * 应用入口 — 创建 Vue 应用实例并装配 Pinia 与路由，挂载到 #app。
 */
import { createPinia } from "pinia";
import { createApp } from "vue";

import App from "./App.vue";
import router from "./router";

import "./design/tokens.css";

const app = createApp(App);
app.use(createPinia());
app.use(router);
app.mount("#app");
