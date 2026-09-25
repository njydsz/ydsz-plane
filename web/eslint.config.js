import pluginVue from "eslint-plugin-vue";
import tseslint from "@typescript-eslint/parser";
import vueParser from "vue-eslint-parser";
import tseslintPlugin from "@typescript-eslint/eslint-plugin";

export default [
  ...pluginVue.configs["flat/recommended"],
  {
    files: ["**/*.{ts,vue}"],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tseslint,
        ecmaVersion: "latest",
        sourceType: "module",
        extraFileExtensions: [".vue"],
      },
    },
    rules: {
      "vue/multi-word-component-names": "off",
      "vue/max-attributes-per-line": "off",
      "vue/singleline-html-element-content-newline": "off",
      "vue/html-self-closing": "off",
      "vue/html-indent": "off",
      // .vue 中 <script setup> 的顶层类型导入由 vue-tsc 负责校验，eslint 不再重复要求
      "no-unused-vars": "off",
    },
  },
  // 叠加 @typescript-eslint/recommended-type-checked（P2-5）
  // 禁用 6 条噪声较大的规则，保留真正依赖类型信息、能捕获真实 bug 的规则，便于后续分批治理
  {
    files: ["src/**/*.{ts,vue}"],
    languageOptions: {
      parserOptions: {
        project: ["./tsconfig.app.json"],
      },
    },
    plugins: {
      "@typescript-eslint": tseslintPlugin,
    },
    rules: {
      ...tseslintPlugin.configs["recommended-type-checked"].rules,
      // 待分批修复：大量 API/响应使用 any，加类型标注后再开启
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-unsafe-assignment": "off",
      "@typescript-eslint/no-unsafe-member-access": "off",
      "@typescript-eslint/no-unsafe-call": "off",
      "@typescript-eslint/no-unsafe-return": "off",
      "@typescript-eslint/no-unsafe-argument": "off",
      // 业务逻辑中大量回调为 fire-and-forget，待补充 void 或 await 后再启用
      "@typescript-eslint/no-floating-promises": "off",
      // Vue 事件处理器返回 Promise 常为误报，待确认后再启用
      "@typescript-eslint/no-misused-promises": "off",
    },
  },
  {
    ignores: [
      "dist",
      "dist-*",
      "_tmp_*",
      "node_modules",
      "coverage",
      "playwright-report",
      "test-results",
      "e2e/.auth",
    ],
  },
];
