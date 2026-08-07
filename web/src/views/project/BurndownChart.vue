<script setup lang="ts">
import { h } from "vue";

export interface BPoint {
  date: string;
  done_points: number;
  remaining: number;
  ideal_line: number;
}

const props = defineProps<{
  points: BPoint[];
  startDate?: string;
  endDate?: string;
}>();

// functional component: render via setup return
</script>

<script lang="ts">
import { defineComponent, h } from "vue";
import type { BPoint } from "./BurndownChart.vue";

export default defineComponent({
  name: "BurndownChartRender",
  props: {
    points: { type: Array as () => BPoint[], required: true },
  },
  setup(props) {
    return () => renderChart(props.points);
  },
});

function renderChart(pts: BPoint[]) {
  const W = 600, H = 220, P = 30;
  if (!pts.length) return h("div", { class: "empty" }, "暂无");

  const maxVal = Math.max(...pts.map((p) => Math.max(p.remaining, p.ideal_line)), 1);
  const stepX = (W - P * 2) / Math.max(pts.length - 1, 1);
  const x = (i: number) => P + i * stepX;
  const y = (v: number) => H - P - (v / maxVal) * (H - P * 2);

  const toPath = (fn: (p: BPoint) => number) =>
    pts.map((p, i) => `${i === 0 ? "M" : "L"}${x(i)},${y(fn(p))}`).join(" ");

  return h(
    "svg",
    { viewBox: `0 0 ${W} ${H}`, class: "burndown-svg", width: "100%" },
    [
      ...[0, 0.25, 0.5, 0.75, 1].map((f) =>
        h("line", { x1: P, y1: y(maxVal * f), x2: W - P, y2: y(maxVal * f), class: "grid" })
      ),
      h("line", { x1: P, y1: H - P, x2: W - P, y2: H - P, class: "axis" }),
      h("line", { x1: P, y1: P, x2: P, y2: H - P, class: "axis" }),
      h("path", { d: toPath((p) => p.ideal_line), class: "ideal" }),
      h("path", { d: toPath((p) => p.done_points), class: "done" }),
      h("path", { d: toPath((p) => p.remaining), class: "remaining" }),
      ...pts.map((p, i) => h("circle", { cx: x(i), cy: y(p.remaining), r: 3, class: "dot" })),
      ...pts.map((p, i) => {
        const d = new Date(p.date);
        return h("text", { x: x(i), y: H - 8, "text-anchor": "middle", class: "axis-label" }, `${d.getMonth() + 1}/${d.getDate()}`);
      }),
      ...[0, maxVal * 0.5, maxVal].map((v) =>
        h("text", { x: P - 5, y: y(v) + 3, "text-anchor": "end", class: "axis-label" }, `${Math.round(v)}pt`)
      ),
    ]
  );
}
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
