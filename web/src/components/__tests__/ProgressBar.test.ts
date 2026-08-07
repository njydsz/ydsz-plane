/**
 * ProgressBar 组件单元测试。
 *
 * 覆盖：百分比渲染、边界钳制（0-100）、label 覆盖、自定义颜色、尺寸/动画 class。
 */
import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";

import ProgressBar from "../ProgressBar.vue";

describe("ProgressBar", () => {
  it("渲染百分比标签", () => {
    const wrapper = mount(ProgressBar, {
      props: { percent: 42.4 },
    });
    expect(wrapper.find(".progress__label").text()).toBe("42%");
  });

  it("label prop 覆盖百分比显示", () => {
    const wrapper = mount(ProgressBar, {
      props: { percent: 50, label: "进行中" },
    });
    expect(wrapper.find(".progress__label").text()).toBe("进行中");
  });

  it("将百分比钳制在 0-100 之间", () => {
    const over = mount(ProgressBar, { props: { percent: 150 } });
    expect(over.find(".progress__bar").attributes("style")).toContain(
      "width: 100%",
    );
    const under = mount(ProgressBar, { props: { percent: -20 } });
    expect(under.find(".progress__bar").attributes("style")).toContain(
      "width: 0%",
    );
  });

  it("应用自定义颜色", () => {
    const wrapper = mount(ProgressBar, {
      props: { percent: 60, color: "red" },
    });
    expect(wrapper.find(".progress__bar").attributes("style")).toContain(
      "background: red",
    );
  });

  it("按 size 应用 class", () => {
    const wrapper = mount(ProgressBar, {
      props: { percent: 10, size: "lg" },
    });
    expect(wrapper.classes()).toContain("progress--lg");
  });

  it("animated/striped 时附加对应 class", () => {
    const wrapper = mount(ProgressBar, {
      props: { percent: 10, animated: true, striped: true },
    });
    expect(wrapper.find(".progress__bar").classes()).toContain(
      "progress__bar--animated",
    );
    expect(wrapper.find(".progress__bar").classes()).toContain(
      "progress__bar--striped",
    );
  });
});
