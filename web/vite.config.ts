import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: '127.0.0.1',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // 注意：不使用 cookieDomainRewrite。
        // Go 的 http.SetCookie 默认不设 Domain（host-only cookie），
        // 浏览器会按当前 origin 直接存储；如果强制重写成固定域名，
        // 当访问 origin 与重写目标不一致时浏览器会拒绝 Set-Cookie。
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('/vue/') || id.includes('/vue-router/') || id.includes('/pinia/')) return 'vue';
            if (id.includes('/element-plus/')) return 'element-plus';
            if (id.includes('/echarts') || id.includes('/vue-echarts/')) return 'echarts';
            if (id.includes('/@tiptap/') || id.includes('/prosemirror')) return 'tiptap';
          }
        },
      },
    },
  },
})
