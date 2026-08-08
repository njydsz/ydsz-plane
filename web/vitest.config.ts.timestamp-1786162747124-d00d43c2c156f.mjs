// vitest.config.ts
import { defineConfig } from "file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vitest@2.1.9_@types+node@20_1a4be881fb770578351b4d0bdae59b55/node_modules/vitest/dist/config.js";
import vue from "file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/@vitejs+plugin-vue@5.2.4_vi_7d31ea48d12dbbc9d2bb928927ee7dee/node_modules/@vitejs/plugin-vue/dist/index.mjs";
import { fileURLToPath, URL } from "node:url";
var __vite_injected_original_import_meta_url = "file:///D:/Code/open/ydsz-plane/web/vitest.config.ts";
var vitest_config_default = defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", __vite_injected_original_import_meta_url))
    }
  },
  test: {
    environment: "happy-dom",
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    exclude: ["e2e/**", "node_modules/**", "dist/**"],
    globals: false,
    coverage: {
      provider: "v8",
      include: ["src/components/**", "src/lib/**", "src/stores/**", "src/api/**"]
    }
  }
});
export {
  vitest_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZXN0LmNvbmZpZy50cyJdLAogICJzb3VyY2VzQ29udGVudCI6IFsiY29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2Rpcm5hbWUgPSBcIkQ6XFxcXENvZGVcXFxcb3BlblxcXFx5ZHN6LXBsYW5lXFxcXHdlYlwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiRDpcXFxcQ29kZVxcXFxvcGVuXFxcXHlkc3otcGxhbmVcXFxcd2ViXFxcXHZpdGVzdC5jb25maWcudHNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL0Q6L0NvZGUvb3Blbi95ZHN6LXBsYW5lL3dlYi92aXRlc3QuY29uZmlnLnRzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSAndml0ZXN0L2NvbmZpZydcbmltcG9ydCB2dWUgZnJvbSAnQHZpdGVqcy9wbHVnaW4tdnVlJ1xuaW1wb3J0IHsgZmlsZVVSTFRvUGF0aCwgVVJMIH0gZnJvbSAnbm9kZTp1cmwnXG5cbmV4cG9ydCBkZWZhdWx0IGRlZmluZUNvbmZpZyh7XG4gIHBsdWdpbnM6IFt2dWUoKV0sXG4gIHJlc29sdmU6IHtcbiAgICBhbGlhczoge1xuICAgICAgJ0AnOiBmaWxlVVJMVG9QYXRoKG5ldyBVUkwoJy4vc3JjJywgaW1wb3J0Lm1ldGEudXJsKSksXG4gICAgfSxcbiAgfSxcbiAgdGVzdDoge1xuICAgIGVudmlyb25tZW50OiAnaGFwcHktZG9tJyxcbiAgICBpbmNsdWRlOiBbJ3NyYy8qKi8qLnt0ZXN0LHNwZWN9Lnt0cyx0c3h9J10sXG4gICAgZXhjbHVkZTogWydlMmUvKionLCAnbm9kZV9tb2R1bGVzLyoqJywgJ2Rpc3QvKionXSxcbiAgICBnbG9iYWxzOiBmYWxzZSxcbiAgICBjb3ZlcmFnZToge1xuICAgICAgcHJvdmlkZXI6ICd2OCcsXG4gICAgICBpbmNsdWRlOiBbJ3NyYy9jb21wb25lbnRzLyoqJywgJ3NyYy9saWIvKionLCAnc3JjL3N0b3Jlcy8qKicsICdzcmMvYXBpLyoqJ10sXG4gICAgfSxcbiAgfSxcbn0pXG4iXSwKICAibWFwcGluZ3MiOiAiO0FBQWlSLFNBQVMsb0JBQW9CO0FBQzlTLE9BQU8sU0FBUztBQUNoQixTQUFTLGVBQWUsV0FBVztBQUZzSSxJQUFNLDJDQUEyQztBQUkxTixJQUFPLHdCQUFRLGFBQWE7QUFBQSxFQUMxQixTQUFTLENBQUMsSUFBSSxDQUFDO0FBQUEsRUFDZixTQUFTO0FBQUEsSUFDUCxPQUFPO0FBQUEsTUFDTCxLQUFLLGNBQWMsSUFBSSxJQUFJLFNBQVMsd0NBQWUsQ0FBQztBQUFBLElBQ3REO0FBQUEsRUFDRjtBQUFBLEVBQ0EsTUFBTTtBQUFBLElBQ0osYUFBYTtBQUFBLElBQ2IsU0FBUyxDQUFDLCtCQUErQjtBQUFBLElBQ3pDLFNBQVMsQ0FBQyxVQUFVLG1CQUFtQixTQUFTO0FBQUEsSUFDaEQsU0FBUztBQUFBLElBQ1QsVUFBVTtBQUFBLE1BQ1IsVUFBVTtBQUFBLE1BQ1YsU0FBUyxDQUFDLHFCQUFxQixjQUFjLGlCQUFpQixZQUFZO0FBQUEsSUFDNUU7QUFBQSxFQUNGO0FBQ0YsQ0FBQzsiLAogICJuYW1lcyI6IFtdCn0K
