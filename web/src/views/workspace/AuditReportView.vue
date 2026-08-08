<script setup lang="ts">
/**
 * 审计日志报表页 — 展示工作空间管理操作日志，仅 owner/admin 可访问。
 * 参考：GitHub Audit Log、GitLab Event Log。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { auditApi, type AuditLogEntry } from "@/api/services/audit";
import { workspaceApi, type Workspace } from "@/api/services/workspace";

const route = useRoute();
const wsId = computed(() => Number(route.params.workspaceId));

const ws = ref<Workspace | null>(null);
const logs = ref<AuditLogEntry[]>([]);
const loading = ref(true);
const error = ref("");
const limit = ref(100);

// 按 action 分组统计
const actionStats = computed(() => {
  const map = new Map<string, number>();
  for (const l of logs.value) {
    const key = l.action.split(".")[0] || "other";
    map.set(key, (map.get(key) ?? 0) + 1);
  }
  return [...map.entries()]
    .map(([action, count]) => ({ action, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 8);
});

// 按 actor 分组统计
const actorStats = computed(() => {
  const map = new Map<number, { name: string; count: number }>();
  for (const l of logs.value) {
    const id = l.actor_id;
    if (!map.has(id)) map.set(id, { name: l.actor_name || `#${id}`, count: 0 });
    map.get(id)!.count++;
  }
  return [...map.values()].sort((a, b) => b.count - a.count).slice(0, 6);
});

function actionLabel(action: string): string {
  const labels: Record<string, string> = {
    "workspace.create": "创建工作空间",
    "workspace.update": "更新工作空间",
    "workspace.archive": "归档工作空间",
    "invitation.send": "发送邀请",
    "invitation.accept": "接受邀请",
    "invitation.revoke": "撤销邀请",
    "member.role_change": "修改成员角色",
    "member.remove": "移除成员",
    "project.create": "创建项目",
    "project.update": "更新项目",
    "project.archive": "归档项目",
    "issue.create": "创建工作项",
    "issue.update": "更新工作项",
    "issue.delete": "删除工作项",
    "issue.transition": "状态流转",
    "version.release": "发布版本",
    "automation.create": "创建自动化",
    "webhook.create": "创建 Webhook",
  };
  for (const [prefix, label] of Object.entries(labels)) {
    if (action.startsWith(prefix)) return label;
  }
  return action;
}

function actionColor(action: string): string {
  if (action.includes("delete") || action.includes("archive") || action.includes("remove") || action.includes("revoke")) return "var(--danger-500)";
  if (action.includes("create")) return "var(--success-500)";
  if (action.includes("update") || action.includes("role_change")) return "var(--warning-500)";
  return "var(--brand-500)";
}

function formatTime(ts: string): string {
  const d = new Date(ts);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "刚刚";
  if (diffMin < 60) return `${diffMin} 分钟前`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr} 小时前`;
  const diffDay = Math.floor(diffHr / 24);
  if (diffDay < 30) return `${diffDay} 天前`;
  return d.toLocaleDateString("zh-CN");
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    ws.value = await workspaceApi.get(wsId.value);
    logs.value = await auditApi.list(ws.value.id, limit.value);
  } catch (e: any) {
    error.value = e.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="audit">
    <header class="audit__header">
      <div>
        <h1 v-if="ws">{{ ws.name }} · 审计日志</h1>
        <h1 v-else>审计日志</h1>
        <p class="hint">工作空间管理操作全量审计（仅 owner / admin 可见）</p>
      </div>
      <div class="actions">
        <select v-model.number="limit" class="page-size" @change="load">
          <option :value="50">50 条</option>
          <option :value="100">100 条</option>
          <option :value="200">200 条</option>
        </select>
        <button class="btn" @click="load">刷新</button>
      </div>
    </header>

    <!-- 空状态 -->
    <div v-if="!loading && !error && logs.length === 0" class="empty">
      <p>暂无审计日志</p>
      <p class="hint">工作空间的管理操作会在此处记录</p>
    </div>

    <!-- 加载与错误 -->
    <div v-if="loading" class="loading">加载中...</div>
    <div v-if="error" class="error">{{ error }}</div>

    <!-- 内容 -->
    <div v-if="!loading && logs.length > 0" class="audit__content">
      <!-- 概览卡片 -->
      <div class="summary-grid">
        <div class="summary-card">
          <div class="summary-card__label">总记录数</div>
          <div class="summary-card__value">{{ logs.length }}</div>
        </div>
        <div class="summary-card">
          <div class="summary-card__label">最近操作者</div>
          <div class="summary-card__value summary-card__value--sm">
            {{ actorStats[0]?.name || "-" }}
          </div>
        </div>
        <div class="summary-card">
          <div class="summary-card__label">操作类型</div>
          <div class="summary-card__value">{{ actionStats.length }}</div>
        </div>
        <div class="summary-card">
          <div class="summary-card__label">活跃操作者</div>
          <div class="summary-card__value">{{ actorStats.length }}</div>
        </div>
      </div>

      <div class="audit__split">
        <!-- 操作分布 -->
        <div class="panel">
          <div class="panel__title">操作分布</div>
          <div class="bar-list">
            <div v-for="s in actionStats" :key="s.action" class="bar-row">
              <span class="bar-row__label">{{ s.action }}</span>
              <div class="bar-row__track">
                <div
                  class="bar-row__fill"
                  :style="{
                    width: (s.count / logs.length) * 100 + '%',
                    background: actionColor(s.action),
                  }"
                ></div>
              </div>
              <span class="bar-row__count">{{ s.count }}</span>
            </div>
          </div>
        </div>

        <!-- 操作者排行 -->
        <div class="panel">
          <div class="panel__title">操作者排行</div>
          <div class="actor-list">
            <div v-for="(a, idx) in actorStats" :key="a.name" class="actor-row">
              <span class="actor-rank">{{ idx + 1 }}</span>
              <span class="actor-name">{{ a.name }}</span>
              <span class="actor-count">{{ a.count }} 次</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 日志列表 -->
      <div class="panel">
        <div class="panel__title">操作记录</div>
        <div class="log-list">
          <div v-for="l in logs" :key="l.id" class="log-row">
            <span class="log-badge" :style="{ background: actionColor(l.action) }">
              {{ actionLabel(l.action) }}
            </span>
            <span class="log-target">{{ l.target || "-" }}</span>
            <span class="log-actor">{{ l.actor_name || `#${l.actor_id}` }}</span>
            <span v-if="l.ip" class="log-ip">{{ l.ip }}</span>
            <span class="log-time">{{ formatTime(l.created_at) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.audit__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.audit__header h1 { font-size: 20px; margin: 0 0 4px; }
.hint { color: var(--text-tertiary); font-size: 13px; margin: 0; }

.actions { display: flex; gap: 10px; align-items: center; }
.page-size {
  padding: 8px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.loading, .error, .empty {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}
.error { color: var(--danger-500); }

.audit__content { display: flex; flex-direction: column; gap: 16px; }

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.summary-card {
  padding: 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-1);
}

.summary-card__label { font-size: 12px; color: var(--text-tertiary); margin-bottom: 6px; }
.summary-card__value { font-size: 24px; font-weight: 600; color: var(--text-primary); }
.summary-card__value--sm { font-size: 16px; }

.audit__split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.panel {
  padding: 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-1);
}

.panel__title { font-size: 14px; font-weight: 600; color: var(--text-primary); margin-bottom: 12px; }

.bar-list, .actor-list { display: flex; flex-direction: column; gap: 8px; }

.bar-row { display: flex; align-items: center; gap: 10px; }
.bar-row__label { flex: 0 0 80px; font-size: 12px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bar-row__track { flex: 1; height: 10px; background: var(--surface-3); border-radius: 5px; overflow: hidden; }
.bar-row__fill { height: 100%; border-radius: 5px; transition: width 0.3s; }
.bar-row__count { flex: 0 0 30px; text-align: right; font-size: 12px; color: var(--text-tertiary); font-variant-numeric: tabular-nums; }

.actor-row { display: flex; align-items: center; gap: 10px; padding: 4px 0; }
.actor-rank { flex: 0 0 20px; font-size: 12px; color: var(--text-tertiary); }
.actor-name { flex: 1; font-size: 13px; color: var(--text-primary); }
.actor-count { font-size: 12px; color: var(--text-tertiary); }

.log-list { display: flex; flex-direction: column; gap: 4px; max-height: 500px; overflow-y: auto; }
.log-row {
  display: grid;
  grid-template-columns: 110px 1fr 120px 130px 100px;
  gap: 12px;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-subtle);
  font-size: 13px;
}
.log-row:last-child { border-bottom: none; }

.log-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-on-brand);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.log-target { color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.log-actor { color: var(--text-secondary); }
.log-ip { color: var(--text-tertiary); font-family: var(--font-mono); font-size: 11px; }
.log-time { color: var(--text-tertiary); text-align: right; }

:deep(.log-list::-webkit-scrollbar) { width: 6px; }
:deep(.log-list::-webkit-scrollbar-thumb) { background: var(--border-default); border-radius: 3px; }
</style>
