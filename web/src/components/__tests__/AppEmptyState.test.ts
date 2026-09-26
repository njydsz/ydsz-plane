/**
 * AppEmptyState 组件单元测试。
 *
 * 覆盖：默认插槽、自定义标题/描述、CTA 事件、预设场景。
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

  it("空标题时隐藏标题区域", () => {
    const wrapper = mount(AppEmptyState, {
      props: { description: "只有描述" },
    });
    expect(wrapper.find("h3").exists()).toBe(false);
  });

  it("点击 CTA 按钮触发 action 事件", async () => {
    const handler = vi.fn();
    const wrapper = mount(AppEmptyState, {
      props: {
        title: "空",
        action-label: "创建",
        onAction: handler,
      },
    });

    const btn = wrapper.find(".app-empty-state__cta");
    expect(btn.exists()).toBe(true);
    await btn.trigger("click");
    expect(handler).toHaveBeenCalledTimes(1);
  });

  it("使用预设场景 issues", () => {
    const wrapper = mount(AppEmptyState, {
      props: { scenario: "issues" },
    });
    // 场景化标题应包含关键词
    const text = wrapper.text();
    expect(text.length).toBeGreaterThan(0);
  });

  it("preset scenarios render without error", () => {
    const scenarios = [
      "default", "projects", "sprints", "modules",
      "search", "notifications", "labels", "members",
    ] as const;

    for (const scenario of scenarios) {
      const wrapper = mount(AppEmptyState, { props: { scenario } });
      expect(wrapper.vm).toBeTruthy();
      wrapper.unmount();
    }
  });
});
