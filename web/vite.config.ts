/// <reference types="vitest/config" />
import { fileURLToPath, URL } from "node:url";

import vue from "@vitejs/plugin-vue";
import { VitePWA } from "vite-plugin-pwa";
import { defineConfig } from "vite";

export default defineConfig(({ mode }) => ({
  plugins: [
    vue(),
    // 仅在 production 模式启用 PWA，避免 dev 模式下 Service Worker 与 HMR 冲突导致页面闪烁
    ...(mode === "production"
      ? [
          VitePWA({
            registerType: "autoUpdate",
            // Service Worker 策略
            workbox: {
              globPatterns: ["**/*.{js,css,html,ico,png,svg,woff2}"],
              // 运行时缓存策略
              runtimeCaching: [
                {
                  // API 请求：网络优先，离线回退缓存
                  urlPattern: /^\/api\/.*/i,
                  handler: "NetworkFirst",
                  options: {
                    cacheName: "api-cache",
                    expiration: {
                      maxEntries: 100,
                      maxAgeSeconds: 60 * 60, // 1 小时
                    },
                    networkTimeoutSeconds: 5,
                  },
                },
                {
                  // 静态资源：缓存优先
                  urlPattern: /\.(?:js|css|woff2|png|svg|ico)$/i,
                  handler: "CacheFirst",
                  options: {
                    cacheName: "static-assets",
                    expiration: {
                      maxEntries: 200,
                      maxAgeSeconds: 30 * 24 * 60 * 60, // 30 天
                    },
                  },
                },
              ],
            },
            // PWA Manifest
            manifest: {
              name: "Ydsz Plane — 开源项目管理平台",
              short_name: "Ydsz Plane",
              description: "面向中国软件团队的开源项目管理平台",
              theme_color: "#3b82f6",
              background_color: "#ffffff",
              display: "standalone",
              orientation: "any",
              scope: "/",
              start_url: "/",
              icons: [
                {
                  src: "/pwa-192x192.png",
                  sizes: "192x192",
                  type: "image/png",
                },
                {
                  src: "/pwa-512x512.png",
                  sizes: "512x512",
                  type: "image/png",
                },
                {
                  src: "/pwa-512x512.png",
                  sizes: "512x512",
                  type: "image/png",
                  purpose: "any maskable",
                },
              ],
              // 快捷方式
              shortcuts: [
                {
                  name: "工作台",
                  short_name: "工作台",
                  description: "查看个人工作台",
                  url: "/workspace",
                },
                {
                  name: "搜索",
                  short_name: "搜索",
                  description: "全局搜索",
                  url: "/search",
                },
              ],
              // 关联应用
              related_applications: [],
              prefer_related_applications: false,
            },
          }),
        ]
      : []),
  ],
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
    host: "127.0.0.1",
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        // 重写后端 Set-Cookie 的 domain，使浏览器能正确接收
        // 后端在 localhost 设置 cookie，但浏览器在 127.0.0.1 访问 Vite，需统一域名
        cookieDomainRewrite: {
          "localhost": "127.0.0.1",
        },
      },
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
}));
