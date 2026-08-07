/**
 * AppButton 组件单元测试。
 *
 * 覆盖：变体/尺寸 class、加载态、禁用态、插槽内容、块级布局。
 */
import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";

import AppButton from "../AppButton.vue";

describe("AppButton", () => {
  it("渲染插槽内容", () => {
    const wrapper = mount(AppButton, {
      slots: { default: "创建项目" },
    });
    expect(wrapper.text()).toContain("创建项目");
  });

  it("默认使用 primary 变体与 md 尺寸", () => {
    const wrapper = mount(AppButton, {
      slots: { default: "确定" },
    });
    expect(wrapper.classes()).toContain("app-btn--primary");
    expect(wrapper.classes()).toContain("app-btn--md");
  });

  it("按 variant prop 应用对应 class", () => {
    const wrapper = mount(AppButton, {
      props: { variant: "danger" },
    });
    expect(wrapper.classes()).toContain("app-btn--danger");
  });

  it("按 size prop 应用对应 class", () => {
    const wrapper = mount(AppButton, {
      props: { size: "lg" },
    });
    expect(wrapper.classes()).toContain("app-btn--lg");
  });

  it("block 时附加块级 class", () => {
    const wrapper = mount(AppButton, {
      props: { block: true },
    });
    expect(wrapper.classes()).toContain("block");
  });

  it("loading 时禁用并渲染 spinner", () => {
    const wrapper = mount(AppButton, {
      props: { loading: true },
      slots: { default: "提交中" },
    });
    expect(wrapper.find("button").attributes("disabled")).toBeDefined();
    expect(wrapper.find(".app-btn__spinner").exists()).toBe(true);
  });

  it("disabled prop 直接禁用按钮", () => {
    const wrapper = mount(AppButton, {
      props: { disabled: true },
    });
    expect(wrapper.find("button").attributes("disabled")).toBeDefined();
  });

  it("正确透传 button type", () => {
    const wrapper = mount(AppButton, {
      props: { type: "submit" },
    });
    expect(wrapper.find("button").attributes("type")).toBe("submit");
  });
});
