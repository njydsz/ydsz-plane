/**
 * i18n 国际化配置
 *
 * 设计参考:
 *   - vue-i18n v10 组合式 API
 *   - 字节跳动国际化最佳实践（语言包按模块拆分、懒加载）
 *   - 语言检测: localStorage > navigator.language > 默认 zh-CN
 *
 * 使用方式:
 *   - 模板: {{ $t('common.save') }}
 *   - 脚本: const { t } = useI18n(); t('auth.login.title')
 *   - 参数化: t('workspace.list.members', { count: 5 })
 */
import { createI18n } from "vue-i18n";
import zhCN from "./zh-CN";
import enUS from "./en-US";

/** 支持的语言列表 */
export const SUPPORTED_LOCALES = [
  { code: "zh-CN", name: "简体中文", flag: "🇨🇳" },
  { code: "en-US", name: "English", flag: "🇺🇸" },
] as const;

/** 支持的语言代码类型（与 SUPPORTED_LOCALES 的 code 字段对齐）。 */
export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number]["code"];

/** 从 localStorage 或浏览器检测用户语言 */
function detectLocale(): SupportedLocale {
  // 1. 用户手动选择的语言
  const stored = localStorage.getItem("ydsz-locale");
  if (stored === "zh-CN" || stored === "en-US") {
    return stored;
  }

  // 2. 浏览器首选语言
  const navLang = navigator.language;
  if (navLang.startsWith("zh")) return "zh-CN";
  if (navLang.startsWith("en")) return "en-US";

  // 3. 默认中文
  return "zh-CN";
}

/** 创建 i18n 实例 */
export const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: detectLocale(),
  fallbackLocale: "zh-CN",
  messages: {
    "zh-CN": zhCN,
    "en-US": enUS,
  },
  // 全局配置
  silentTranslationWarn: import.meta.env.PROD, // 生产环境静默缺失翻译警告
  missingWarn: import.meta.env.DEV, // 开发环境显示缺失翻译警告
  fallbackWarn: import.meta.env.DEV,
});

/**
 * 切换语言
 * 同时持久化到 localStorage 和设置 html[lang] 属性
 */
export function setLocale(locale: SupportedLocale): void {
  i18n.global.locale.value = locale;
  localStorage.setItem("ydsz-locale", locale);
  document.documentElement.setAttribute("lang", locale);
}

/** 获取当前语言 */
export function getLocale(): SupportedLocale {
  return i18n.global.locale.value as SupportedLocale;
}
