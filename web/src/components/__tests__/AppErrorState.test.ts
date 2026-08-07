/**
 * AppErrorState 组件单元测试。
 *
 * 覆盖：错误文案渲染、重试按钮触发事件、retry 为空时隐藏按钮。
 */
import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";

import AppErrorState from "../AppErrorState.vue";

describe("AppErrorState", () => {
  it("渲染错误信息", () => {
    const wrapper = mount(AppErrorState, {
      props: { message: "加载失败" },
    });
    expect(wrapper.find(".app-error__text").text()).toBe("加载失败");
  });

  it("默认显示重试按钮并触发 retry 事件", async () => {
    const wrapper = mount(AppErrorState, {
      props: { message: "加载失败" },
    });
    const btn = wrapper.find(".app-error__btn");
    expect(btn.exists()).toBe(true);
    expect(btn.text()).toBe("重试");
    await btn.trigger("click");
    expect(wrapper.emitted("retry")).toHaveLength(1);
  });

  it("retry 设为空字符串时隐藏重试按钮", () => {
    const wrapper = mount(AppErrorState, {
      props: { message: "加载失败", retry: "" },
    });
    expect(wrapper.find(".app-error__btn").exists()).toBe(false);
  });
});
