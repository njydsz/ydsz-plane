<script lang="ts">
/**
 * BurndownChart — 燃尽图 SVG 组件（无外部图表依赖）。
 *
 * 根据迭代每日快照渲染三条折线：剩余点数（actual）、已完成点数（done）
 * 与理想线（ideal_line）。支持 startDate / endDate 边界与空数据占位。
 */
import { defineComponent, h } from "vue";

/** 燃尽图单日数据点 */
export interface BPoint {
  date: string;
  done_points: number;
  remaining: number;
  ideal_line: number;
}

export default defineComponent({
  name: "BurndownChart",
  props: {
    points: { type: Array as () => BPoint[], required: true },
    startDate: { type: String, default: undefined },
    endDate: { type: String, default: undefined },
  },
  setup(props) {
    return () => {
      const W = 600, H = 220, P = 30;
      const pts = props.points;
      if (!pts.length) return h("div", { class: "empty" }, "暂无数据");

      const maxVal = Math.max(...pts.map((p) => Math.max(p.remaining, p.ideal_line)), 1);
      const stepX = (W - P * 2) / Math.max(pts.length - 1, 1);
      const xp = (i: number) => P + i * stepX;
      const yp = (v: number) => H - P - (v / maxVal) * (H - P * 2);

      const toPath = (fn: (p: BPoint) => number) =>
        pts.map((p, i) => `${i === 0 ? "M" : "L"}${xp(i)},${yp(fn(p))}`).join(" ");

      return h(
        "svg",
        { viewBox: `0 0 ${W} ${H}`, class: "burndown-svg", width: "100%" },
        [
          ...[0, 0.25, 0.5, 0.75, 1].map((f) =>
            h("line", { x1: P, y1: yp(maxVal * f), x2: W - P, y2: yp(maxVal * f), class: "grid" })
          ),
          h("line", { x1: P, y1: H - P, x2: W - P, y2: H - P, class: "axis" }),
          h("line", { x1: P, y1: P, x2: P, y2: H - P, class: "axis" }),
          h("path", { d: toPath((p) => p.ideal_line), class: "ideal" }),
          h("path", { d: toPath((p) => p.done_points), class: "done" }),
          h("path", { d: toPath((p) => p.remaining), class: "remaining" }),
          ...pts.map((p, i) => h("circle", { cx: xp(i), cy: yp(p.remaining), r: 3, class: "dot" })),
          ...pts.map((p, i) => {
            const d = new Date(p.date);
            return h("text", { x: xp(i), y: H - 8, "text-anchor": "middle", class: "axis-label" }, `${d.getMonth() + 1}/${d.getDate()}`);
          }),
          ...[0, maxVal * 0.5, maxVal].map((v) =>
            h("text", { x: P - 5, y: yp(v) + 3, "text-anchor": "end", class: "axis-label" }, `${Math.round(v)}pt`)
          ),
        ]
      );
    };
  },
});
</script>

<style scoped>
.burndown-svg { background: var(--surface-2); border-radius: var(--radius-sm); display: block; }
.burndown-svg .grid { stroke: var(--border-subtle); stroke-width: 1; }
.burndown-svg .axis { stroke: var(--text-tertiary); stroke-width: 1; }
.burndown-svg .axis-label { font-size: 10px; fill: var(--text-tertiary); }
.burndown-svg .ideal { stroke: var(--warning-500); stroke-width: 1.5; stroke-dasharray: 4 3; fill: none; }
.burndown-svg .remaining { stroke: var(--brand-500); stroke-width: 2; fill: none; }
.burndown-svg .done { stroke: var(--success-500); stroke-width: 2; fill: none; }
.burndown-svg .dot { fill: var(--brand-500); }
.empty { text-align: center; color: var(--text-tertiary); padding: 24px 0; }
</style>
