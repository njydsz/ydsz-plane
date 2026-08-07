/**
 * useI18n — 国际化便捷 composable
 *
 * 封装 vue-i18n 的 useI18n，提供:
 *   - t() 翻译函数（支持参数化）
 *   - locale 当前语言
 *   - availableLocales 可用语言列表
 *   - setLocale() 切换语言
 *
 * 使用示例:
 *   const { t, locale } = useI18n();
 *   t('common.save')                    // "保存" / "Save"
 *   t('workspace.list.members', { count: 5 })  // "5 个成员" / "5 members"
 */
import { useI18n as useVueI18n } from "vue-i18n";
import {
  setLocale as setGlobalLocale,
  getLocale,
  type SupportedLocale,
} from "../locales";

/** 国际化便捷 composable — 封装 vue-i18n，提供 t() / locale / setLocale / availableLocales。 */
export function useI18n() {
  const { t, locale, availableLocales, ...rest } = useVueI18n();

  return {
    /** 翻译函数 */
    t,
    /** 当前语言 */
    locale,
    /** 可用语言列表 */
    availableLocales: availableLocales as SupportedLocale[],
    /** 切换语言 */
    setLocale: (loc: SupportedLocale) => setGlobalLocale(loc),
    /** 获取当前语言代码 */
    getLocale,
    ...rest,
  };
}
