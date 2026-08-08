import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
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
})
