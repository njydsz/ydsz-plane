/**
 * formatTime 工具函数单元测试。
 *
 * 覆盖：formatRelativeTime 各个时间桶的边界条件。
 */
import { describe, expect, it } from "vitest";
import { formatRelativeTime } from "../formatTime";

describe("formatRelativeTime", () => {
  const NOW = Date.now();

  it("刚刚（5 秒前）", () => {
    const t = new Date(NOW - 5 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("刚刚");
  });

  it("分钟前（30 分钟前）", () => {
    const t = new Date(NOW - 30 * 60 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("30 分钟前");
  });

  it("小时前（5 小时前）", () => {
    const t = new Date(NOW - 5 * 60 * 60 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("5 小时前");
  });

  it("天前（3 天前）", () => {
    const t = new Date(NOW - 3 * 24 * 60 * 60 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("3 天前");
  });

  it("周前（2 周前）", () => {
    const t = new Date(NOW - 14 * 24 * 60 * 60 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("2 周前");
  });

  it("个月前（3 个月前）", () => {
    const t = new Date(NOW - 90 * 24 * 60 * 60 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("3 个月前");
  });

  it("年前（2 年前）", () => {
    const t = new Date(NOW - 730 * 24 * 60 * 60 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("2 年前");
  });

  it("未来时间回退为刚刚", () => {
    const t = new Date(NOW + 60 * 1000).toISOString();
    expect(formatRelativeTime(t)).toBe("刚刚");
  });

  it("接受 Date 对象", () => {
    const d = new Date(NOW - 30 * 1000);
    expect(formatRelativeTime(d)).toBe("刚刚");
  });
});
