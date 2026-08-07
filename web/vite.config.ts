/// <reference types="vitest/config" />
import { fileURLToPath, URL } from "node:url";

import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  test: {
    // 前端组件/单元测试环境
    environment: "happy-dom",
    globals: true,
    include: ["src/**/*.{test,spec}.ts"],
    // 覆盖率阈值门禁：全量 >= 50%，函数 >= 40%
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      include: ["src/components/**/*.vue", "src/stores/**/*.ts"],
      thresholds: {
        global: {
          lines: 50,
          functions: 40,
          branches: 40,
          statements: 50,
        },
      },
    },
  },
  build: {
    // 将重量级可视化库单独拆包，避免打入业务 chunk 影响首屏加载
    rollupOptions: {
      output: {
        manualChunks: {
          echarts: ["echarts", "vue-echarts"],
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
});
