import pluginVue from "eslint-plugin-vue";
import tseslint from "@typescript-eslint/parser";
import vueParser from "vue-eslint-parser";

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
