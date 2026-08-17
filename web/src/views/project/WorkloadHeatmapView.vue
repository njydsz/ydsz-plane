<template>
  <div class="workload-heatmap">
    <!-- 页头 -->
    <div class="wh-header">
      <div class="wh-header__left">
        <h2 class="wh-title">工作量热力图</h2>
        <p class="wh-subtitle">项目成员在指定日期范围内的工时分布</p>
      </div>
      <div class="wh-header__right">
        <div class="wh-date-range">
          <label>日期范围</label>
          <div class="wh-date-inputs">
            <input
              v-model="dateFrom"
              type="date"
              class="wh-date-input"
              :max="dateTo || undefined"
              @change="loadData"
            />
            <span class="wh-date-sep">至</span>
            <input
              v-model="dateTo"
              type="date"
              class="wh-date-input"
              :min="dateFrom || undefined"
              @change="loadData"
            />
            <button class="wh-btn wh-btn--ghost" @click="setRange(7)">近 7 天</button>
            <button class="wh-btn wh-btn--ghost" @click="setRange(30)">近 30 天</button>
            <button class="wh-btn wh-btn--ghost" @click="setRange(90)">近 90 天</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 汇总卡片 -->
    <div class="wh-summary" v-if="data">
      <div class="wh-summary__card">
        <span class="wh-summary__label">总工时</span>
        <span class="wh-summary__value">{{ formatHours(data.summary.total_hours) }}</span>
      </div>
      <div class="wh-summary__card">
        <span class="wh-summary__label">参与人数</span>
        <span class="wh-summary__value">{{ data.summary.total_members }} 人</span>
      </div>
      <div class="wh-summary__card">
        <span class="wh-summary__label">活跃天数</span>
        <span class="wh-summary__value">{{ data.summary.total_days }} 天</span>
      </div>
      <div class="wh-summary__card">
        <span class="wh-summary__label">日均工时</span>
        <span class="wh-summary__value">{{ formatHours(data.summary.daily_average_hours) }}</span>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="wh-loading">
      <span class="wh-spinner"></span>
      <span>加载中...</span>
    </div>

    <!-- 空状态 -->
    <div v-else-if="data && data.entries.length === 0" class="wh-empty">
      <p>该日期范围内暂无工时记录</p>
      <p class="wh-empty__hint">成员在工作项上登记工时后，热力图将自动展示</p>
    </div>

    <!-- 热力图主体 -->
    <div v-else-if="data" class="wh-body">
      <!-- 热力图网格 -->
      <div class="wh-grid-container">
        <div class="wh-grid-header">
          <div class="wh-grid-label-col"></div>
          <div class="wh-grid-months">
            <span
              v-for="month in months"
              :key="month.key"
              class="wh-month-label"
              :style="{ gridColumn: `${month.startCol} / span ${month.span}` }"
            >
              {{ month.label }}
            </span>
          </div>
        </div>

        <div class="wh-grid-scroll">
          <div
            v-for="member in data.members"
            :key="member.user_id"
            class="wh-grid-row"
          >
            <div class="wh-member-label" :title="`用户 ${member.user_id}`">
              <span class="wh-member-avatar">{{ getAvatar(member.user_id) }}</span>
              <span class="wh-member-name">#{{ member.user_id }}</span>
              <span class="wh-member-hours">{{ formatHours(member.total_hours) }}</span>
            </div>
            <div class="wh-cells">
              <div
                v-for="day in allDays"
                :key="day"
                class="wh-cell"
                :class="getCellClass(member.user_id, day)"
                :title="getCellTooltip(member.user_id, day)"
                :style="getCellStyle(member.user_id, day)"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- 图例 -->
      <div class="wh-legend">
        <span class="wh-legend-label">少</span>
        <div
          v-for="level in 5"
          :key="level"
          class="wh-legend-cell"
          :class="`wh-cell--level-${level - 1}`"
        ></div>
        <span class="wh-legend-label">多</span>
      </div>
    </div>

    <!-- 成员排行 -->
    <div v-if="data && data.members.length > 0" class="wh-ranking">
      <h3 class="wh-ranking__title">成员工时排行</h3>
      <div class="wh-ranking__list">
        <div
          v-for="(member, idx) in data.members"
          :key="member.user_id"
          class="wh-ranking__item"
        >
          <span class="wh-ranking__rank" :class="idx < 3 ? `wh-ranking__rank--top${idx + 1}` : ''">
            {{ idx + 1 }}
          </span>
          <span class="wh-member-avatar">{{ getAvatar(member.user_id) }}</span>
          <span class="wh-ranking__name">#{{ member.user_id }}</span>
          <div class="wh-ranking__bar-wrap">
            <div
              class="wh-ranking__bar"
              :style="{ width: getBarWidth(member.total_hours) }"
            ></div>
          </div>
          <span class="wh-ranking__hours">{{ formatHours(member.total_hours) }}</span>
          <span class="wh-ranking__days">{{ member.day_count }} 天</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { workloadApi, type WorkloadHeatmapData } from "@/api/services/workload";

const route = useRoute();
const wsId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const loading = ref(false);
const data = ref<WorkloadHeatmapData | null>(null);
const dateFrom = ref("");
const dateTo = ref("");

// 初始化日期范围为最近 30 天
function initDateRange() {
  const now = new Date();
  dateTo.value = formatDate(now);
  dateFrom.value = formatDate(new Date(now.getTime() - 29 * 24 * 60 * 60 * 1000));
}

function setRange(days: number) {
  const now = new Date();
  dateTo.value = formatDate(now);
  dateFrom.value = formatDate(new Date(now.getTime() - (days - 1) * 24 * 60 * 60 * 1000));
  loadData();
}

function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10);
}

function formatHours(h: number): string {
  if (h < 1) return `${Math.round(h * 60)} 分钟`;
  return `${h.toFixed(1)} 小时`;
}

// 生成日期范围内的所有日期
const allDays = computed(() => {
  if (!dateFrom.value || !dateTo.value) return [];
  const days: string[] = [];
  const start = new Date(dateFrom.value);
  const end = new Date(dateTo.value);
  for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
    days.push(formatDate(new Date(d)));
  }
  return days;
});

// 月份标签
const months = computed(() => {
  const result: { key: string; label: string; startCol: number; span: number }[] = [];
  if (allDays.value.length === 0) return result;

  let currentMonth = "";
  let startIdx = 0;
  let dayCount = 0;

  allDays.value.forEach((day, idx) => {
    const month = day.slice(0, 7); // YYYY-MM
    if (month !== currentMonth) {
      if (currentMonth) {
        result.push({
          key: currentMonth,
          label: formatMonthLabel(currentMonth),
          startCol: startIdx + 1,
          span: dayCount,
        });
      }
      currentMonth = month;
      startIdx = idx;
      dayCount = 1;
    } else {
      dayCount++;
    }
  });
  if (currentMonth) {
    result.push({
      key: currentMonth,
      label: formatMonthLabel(currentMonth),
      startCol: startIdx + 1,
      span: dayCount,
    });
  }
  return result;
});

function formatMonthLabel(ym: string): string {
  const [y, m] = ym.split("-");
  return `${y}年${Number(m)}月`;
}

// 构建查找表：user_id -> date -> entry
const entryMap = computed(() => {
  const map = new Map<number, Map<string, { total_hours: number; total_minutes: number; log_count: number; issue_count: number }>>();
  if (!data.value) return map;
  for (const e of data.value.entries) {
    if (!map.has(e.user_id)) map.set(e.user_id, new Map());
    map.get(e.user_id)!.set(e.spent_date, {
      total_hours: e.total_hours,
      total_minutes: e.total_minutes,
      log_count: e.log_count,
      issue_count: e.issue_count,
    });
  }
  return map;
});

// 最大单日工时（用于颜色等级计算）
const maxDailyHours = computed(() => {
  let max = 0;
  for (const e of data.value?.entries ?? []) {
    if (e.total_hours > max) max = e.total_hours;
  }
  return max || 8; // 默认 8 小时为满格
});

function getCellClass(userId: number, day: string): string {
  const entry = entryMap.value.get(userId)?.get(day);
  if (!entry) return "wh-cell--level-0";
  const level = getLevel(entry.total_hours);
  return `wh-cell--level-${level}`;
}

function getCellStyle(_userId: number, _day: string): Record<string, string> {
  return {};
}

function getLevel(hours: number): number {
  const max = maxDailyHours.value;
  if (hours <= 0) return 0;
  const ratio = hours / max;
  if (ratio <= 0.25) return 1;
  if (ratio <= 0.5) return 2;
  if (ratio <= 0.75) return 3;
  return 4;
}

function getCellTooltip(userId: number, day: string): string {
  const entry = entryMap.value.get(userId)?.get(day);
  if (!entry) return `${day}\n#${userId}\n无记录`;
  return `${day}\n#${userId}\n${formatHours(entry.total_hours)} (${entry.log_count} 条记录, ${entry.issue_count} 个工作项)`;
}

function getAvatar(userId: number): string {
  return String(userId).slice(-2).padStart(2, "0");
}

function getBarWidth(hours: number): string {
  if (!data.value?.members[0]) return "0%";
  const max = data.value.members[0].total_hours;
  if (max <= 0) return "0%";
  return `${Math.max(2, (hours / max) * 100)}%`;
}

async function loadData() {
  if (!dateFrom.value || !dateTo.value) return;
  loading.value = true;
  try {
    data.value = await workloadApi.getHeatmap(wsId, projectId, {
      date_from: dateFrom.value,
      date_to: dateTo.value,
    });
  } catch (err) {
    console.error("加载热力图失败:", err);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  initDateRange();
  loadData();
});
</script>

<style scoped>
.workload-heatmap {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
}

/* ---- Header ---- */
.wh-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}
.wh-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}
.wh-subtitle {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
}
.wh-date-range {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.wh-date-range label {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 500;
}
.wh-date-inputs {
  display: flex;
  align-items: center;
  gap: 8px;
}
.wh-date-input {
  padding: 6px 10px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
  background: var(--bg-primary);
}
.wh-date-sep {
  font-size: 13px;
  color: var(--text-tertiary);
}
.wh-btn {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: all 0.15s;
}
.wh-btn--ghost:hover {
  background: var(--bg-secondary);
  border-color: var(--border-hover);
}

/* ---- Summary Cards ---- */
.wh-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}
.wh-summary__card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.wh-summary__label {
  font-size: 12px;
  color: var(--text-tertiary);
}
.wh-summary__value {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

/* ---- Loading & Empty ---- */
.wh-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 48px;
  color: var(--text-tertiary);
  font-size: 14px;
}
.wh-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-default);
  border-top-color: var(--brand-500);
  border-radius: 50%;
  animation: wh-spin 0.8s linear infinite;
}
@keyframes wh-spin {
  to { transform: rotate(360deg); }
}
.wh-empty {
  text-align: center;
  padding: 48px;
  color: var(--text-tertiary);
  font-size: 14px;
}
.wh-empty__hint {
  font-size: 12px;
  margin-top: 8px;
}

/* ---- Heatmap Grid ---- */
.wh-body {
  margin-bottom: 32px;
}
.wh-grid-container {
  border: 1px solid var(--border-default);
  border-radius: 8px;
  overflow: hidden;
}
.wh-grid-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-default);
}
.wh-grid-label-col {
  width: 180px;
  flex-shrink: 0;
}
.wh-grid-months {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(14px, 1fr));
  font-size: 11px;
  color: var(--text-tertiary);
}
.wh-month-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.wh-grid-scroll {
  max-height: 400px;
  overflow-y: auto;
  padding: 0 12px 8px;
}
.wh-grid-row {
  display: flex;
  align-items: center;
  padding: 4px 0;
  border-bottom: 1px solid var(--border-subtle);
}
.wh-grid-row:last-child {
  border-bottom: none;
}
.wh-member-label {
  width: 180px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}
.wh-member-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--brand-100);
  color: var(--brand-700);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 600;
  flex-shrink: 0;
}
.wh-member-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.wh-member-hours {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-right: 8px;
}
.wh-cells {
  flex: 1;
  display: flex;
  gap: 2px;
  overflow-x: auto;
}
.wh-cell {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  flex-shrink: 0;
  cursor: default;
  transition: transform 0.1s;
}
.wh-cell:hover {
  transform: scale(1.3);
  outline: 1px solid var(--text-tertiary);
}

/* Heatmap cell levels */
.wh-cell--level-0 {
  background: var(--heat-0, #ebedf0);
}
.wh-cell--level-1 {
  background: var(--heat-1, #9be9a8);
}
.wh-cell--level-2 {
  background: var(--heat-2, #40c463);
}
.wh-cell--level-3 {
  background: var(--heat-3, #30a14e);
}
.wh-cell--level-4 {
  background: var(--heat-4, #216e39);
}

/* ---- Legend ---- */
.wh-legend {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  margin-top: 8px;
  font-size: 11px;
  color: var(--text-tertiary);
}
.wh-legend-label {
  margin: 0 4px;
}
.wh-legend-cell {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}

/* ---- Ranking ---- */
.wh-ranking {
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 16px;
}
.wh-ranking__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 12px;
}
.wh-ranking__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.wh-ranking__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 6px;
  background: var(--bg-secondary);
}
.wh-ranking__rank {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
  font-size: 11px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.wh-ranking__rank--top1 {
  background: #ffd700;
  color: #7a5800;
}
.wh-ranking__rank--top2 {
  background: #c0c0c0;
  color: #4a4a4a;
}
.wh-ranking__rank--top3 {
  background: #cd7f32;
  color: #fff;
}
.wh-ranking__name {
  width: 60px;
  font-size: 13px;
  color: var(--text-primary);
  flex-shrink: 0;
}
.wh-ranking__bar-wrap {
  flex: 1;
  height: 8px;
  background: var(--bg-tertiary);
  border-radius: 4px;
  overflow: hidden;
}
.wh-ranking__bar {
  height: 100%;
  background: var(--brand-500);
  border-radius: 4px;
  transition: width 0.3s ease;
}
.wh-ranking__hours {
  width: 70px;
  text-align: right;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  flex-shrink: 0;
}
.wh-ranking__days {
  width: 40px;
  text-align: right;
  font-size: 11px;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

/* ---- Responsive ---- */
@media (max-width: 768px) {
  .wh-summary {
    grid-template-columns: repeat(2, 1fr);
  }
  .wh-member-label {
    width: 100px;
  }
  .wh-grid-label-col {
    width: 100px;
  }
}
</style>
