/**
 * AppAvatar 组件单元测试。
 *
 * 覆盖：name 首字符占位、src 图片渲染、size class、aria 可访问性。
 */
import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";

import AppAvatar from "../AppAvatar.vue";

describe("AppAvatar", () => {
  it("无 src 时渲染名称首字符作为占位（大写）", () => {
    const wrapper = mount(AppAvatar, {
      props: { name: "zhangsan" },
    });
    expect(wrapper.find(".app-avatar__fallback").text()).toBe("Z");
  });

  it("提供 src 时渲染 img 而非占位符", () => {
    const wrapper = mount(AppAvatar, {
      props: { name: "lisi", src: "/logo.png" },
    });
    const img = wrapper.find("img.app-avatar__img");
    expect(img.exists()).toBe(true);
    expect(img.attributes("src")).toBe("/logo.png");
    expect(wrapper.find(".app-avatar__fallback").exists()).toBe(false);
  });

  it("按 size prop 应用对应 class", () => {
    const wrapper = mount(AppAvatar, {
      props: { name: "a", size: "lg" },
    });
    expect(wrapper.classes()).toContain("app-avatar--lg");
  });

  it("暴露可访问的 role 与 aria-label", () => {
    const wrapper = mount(AppAvatar, {
      props: { name: "wangwu" },
    });
    expect(wrapper.find("span.app-avatar").attributes("role")).toBe("img");
    expect(wrapper.find("span.app-avatar").attributes("aria-label")).toBe(
      "wangwu",
    );
  });
});
