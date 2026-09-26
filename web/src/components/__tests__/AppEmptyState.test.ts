/**
 * AppEmptyState 组件单元测试。
 *
 * 覆盖：默认场景、自定义标题/描述、CTA 事件、预设场景枚举。
 */
import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";

import AppEmptyState from "../AppEmptyState.vue";

describe("AppEmptyState", () => {
  it("渲染自定义标题与描述", () => {
    const wrapper = mount(AppEmptyState, {
      props: { title: "暂无数据", description: "开始创建第一个需求" },
    });
    expect(wrapper.text()).toContain("暂无数据");
    expect(wrapper.text()).toContain("开始创建第一个需求");
  });

  it("空标题时渲染正常", () => {
    const wrapper = mount(AppEmptyState, {
      props: { description: "只有描述" },
    });
    expect(wrapper.text()).toContain("只有描述");
  });

  it("cta 带 ctaText 属性时渲染按钮", () => {
    const wrapper = mount(AppEmptyState, {
      props: {
        title: "空",
        ctaText: "创建",
      },
    });
    const ctaBtn = wrapper.find(".app-empty__cta");
    expect(ctaBtn.exists()).toBe(true);
    expect(ctaBtn.text()).toContain("创建");
  });

  it("不带 ctaText 时不渲染 CTA 按钮", () => {
    const wrapper = mount(AppEmptyState, {
      props: { title: "空" },
    });
    const ctaBtn = wrapper.find(".app-empty__cta");
    expect(ctaBtn.exists()).toBe(false);
  });

  it("使用预设场景 issues 显示场景化内容", () => {
    const wrapper = mount(AppEmptyState, {
      props: { scenario: "issues" },
    });
    // 场景化标题应包含关键词
    const text = wrapper.text();
    expect(text).toContain("还没有");
  });

  it("所有预设场景渲染无报错", () => {
    const scenarios = [
      "default", "issues", "projects", "sprints", "modules",
      "search", "notifications", "labels", "members",
      "analytics", "views", "inbox", "api-token", "webhooks",
      "error", "gantt", "calendar", "pages", "cycles",
      "automation", "comments",
    ] as const;

    for (const scenario of scenarios) {
      const wrapper = mount(AppEmptyState, { props: { scenario } });
      expect(wrapper.vm).toBeTruthy();
      wrapper.unmount();
    }
  });

  it("compact 模式附加对应 class", () => {
    const wrapper = mount(AppEmptyState, {
      props: { compact: true, title: "test" },
    });
    expect(wrapper.classes()).toContain("app-empty--compact");
  });
});
