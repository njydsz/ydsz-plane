/**
 * useTranslation — 高级国际化 composable
 *
 * 基于 vue-i18n 全局实例（global scope）封装，提供:
 *   - t()    翻译函数（参数化 / 嵌套 key 支持）
 *   - te()   判断 key 是否存在（translation exists）
 *   - tm()   返回嵌套对象（复数 / 命名插槽场景）
 *   - $tc()  复数/插值简化调用（中文语境常用）
 *   - locale 当前语言 ref
 *
 * 使用示例:
 *   const { t, te, tm, locale, $tc } = useTranslation()
 *   t('common.save')                                        // "保存"
 *   t('workspace.list.members', { count: 5 })                // "5 个成员"
 *   te('common.unknown')                                     // false
 *   tm('issue.priority')                                     // { critical: "紧急", ... }
 *   $tc('issue.list.selectedCount', 3)                        // "已选 3 项"
 */
import { useI18n } from "vue-i18n";

import type { SupportedLocale } from "../locales";
import { setLocale as setGlobalLocale, getLocale } from "../locales";

export function useTranslation() {
  const { t, te, tm, locale, messages } = useI18n({ useScope: "global" });

  /**
   * $tc — 复数 / 插值简化调用
   * @param key    翻译 key
   * @param n      数量（用于复数 / 单数区分；可选）
   * @param values 插值参数对象（可选）
   *
   * 与 vue-i18n 内置 $tc 不同，此封装统一处理常见中文插值模式。
   */
  function $tc(key: string, n?: number, values?: Record<string, unknown>): string {
    const interpolation: Record<string, unknown> = { ...values };
    if (n !== undefined) {
      interpolation.count = n;
      interpolation.n = n;
    }
    return t(key, interpolation);
  }

  return {
    /** 翻译函数（模板/脚本通用） */
    t,
    /** 判断 key 是否存在 */
    te,
    /** 返回嵌套对象（复数/命名插槽） */
    tm,
    /** 当前语言 ref */
    locale,
    /** messages 对象（高级场景） */
    messages,
    /** 复数/插值简化函数 */
    $tc,
    /** 切换语言（持久化到 localStorage） */
    setLocale: (loc: SupportedLocale) => setGlobalLocale(loc),
    /** 获取当前语言代码（非响应式） */
    getLocale,
  };
}
