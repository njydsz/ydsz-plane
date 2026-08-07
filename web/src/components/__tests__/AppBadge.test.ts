/**
 * AppBadge 组件单元测试。
 *
 * 覆盖：默认/自定义变体 class、dot 模式、插槽内容。
 */
import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";

import AppBadge from "../AppBadge.vue";

describe("AppBadge", () => {
  it("渲染插槽内容", () => {
    const wrapper = mount(AppBadge, {
      slots: { default: "进行中" },
    });
    expect(wrapper.text()).toContain("进行中");
  });

  it("默认使用 default 变体", () => {
    const wrapper = mount(AppBadge, {
      slots: { default: "x" },
    });
    expect(wrapper.classes()).toContain("app-badge--default");
  });

  it("按 variant prop 应用对应 class", () => {
    const wrapper = mount(AppBadge, {
      props: { variant: "success" },
      slots: { default: "完成" },
    });
    expect(wrapper.classes()).toContain("app-badge--success");
  });

  it("dot 模式附加圆点 class", () => {
    const wrapper = mount(AppBadge, {
      props: { dot: true },
    });
    expect(wrapper.classes()).toContain("dot");
  });
});
