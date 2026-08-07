/**
 * 时间格式化工具 —— 提供相对时间与绝对时间的友好展示。
 *
 * 约定：
 *  - < 60s         → "刚刚"
 *  - < 60min       → "N 分钟前"
 *  - < 24h         → "N 小时前"
 *  - < 7d          → "N 天前"
 *  - < 30d         → "N 周前"
 *  - < 365d        → "N 个月前"
 *  - >= 365d       → "N 年前"
 *  - 超过 30 天也可回退为 YYYY/MM/DD。
 */

const SECOND = 1000
const MINUTE = 60 * SECOND
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR
const WEEK = 7 * DAY
const MONTH = 30 * DAY
const YEAR = 365 * DAY

/** 将 ISO 时间戳或 Date 格式化为相对时间文字列 */
export function formatRelativeTime(input: string | Date | number): string {
  const date = typeof input === 'string' ? new Date(input) : new Date(input)
  const now = Date.now()
  const diff = now - date.getTime()

  if (diff < 0) return '刚刚'
  if (diff < MINUTE) return '刚刚'
  if (diff < HOUR) return `${Math.floor(diff / MINUTE)} 分钟前`
  if (diff < DAY) return `${Math.floor(diff / HOUR)} 小时前`
  if (diff < WEEK) return `${Math.floor(diff / DAY)} 天前`
  if (diff < MONTH) return `${Math.floor(diff / WEEK)} 周前`
  if (diff < YEAR) return `${Math.floor(diff / MONTH)} 个月前`
  return `${Math.floor(diff / YEAR)} 年前`
}

/**
 * 将 ISO 时间戳格式化为绝对日期（YYYY/MM/DD），
 * 用于超过 30 天的通知的 fallback 展示。
 */
export function formatAbsoluteDate(input: string | Date | number): string {
  const date = typeof input === 'string' ? new Date(input) : new Date(input)
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}/${m}/${d}`
}
