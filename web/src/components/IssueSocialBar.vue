<script setup lang="ts">
/**
 * IssueSocialBar — 工作项社交反馈栏（表情反应 + 投票 + 关注）。
 *
 * 参考 Plane / Linear 的轻量协作反馈交互：
 *  - 表情反应：点击即切换，聚合展示各表情计数，hover 高亮
 *  - 投票：赞成/反对切换，显示净分
 *  - 关注：订阅工作项动态（状态变更/评论时收到通知）
 */
import { computed, onMounted, ref } from "vue";

import { issueApi, type ReactionSummary, type VoteSummary } from "@/api/services/issue";
import { useAuthStore } from "@/stores/auth";
import { promiseToast } from "@/lib/toast";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
  /** 当前用户是否已关注（由父级 Issue.watchers 推导） */
  initialWatching?: boolean;
}>();

const auth = useAuthStore();

const reactions = ref<ReactionSummary[]>([]);
const vote = ref<VoteSummary>({ upvotes: 0, downvotes: 0, score: 0, voted: null });
const watching = ref(!!props.initialWatching);
const loading = ref(true);
const busy = ref(false);

/** 预设表情面板（点击 + 号展开） */
const emojiPanelOpen = ref(false);
const emojiChoices = ["👍", "👀", "🎉", "❤️", "😄", "🔥", "🚀", "👏"];

const currentUserId = computed(() => auth.user?.id ?? 0);

const canInteract = computed(() => currentUserId.value > 0 && !busy.value);

async function load() {
  loading.value = true;
  try {
    const [reactionsRes, voteRes] = await Promise.all([
      issueApi.listReactions(props.workspaceId, props.projectId, props.issueId),
      issueApi.voteSummary(props.workspaceId, props.projectId, props.issueId),
    ]);
    reactions.value = reactionsRes.results ?? [];
    vote.value = voteRes;
  } catch {
    // 静默失败：社交反馈为增强功能，不阻塞页面
  } finally {
    loading.value = false;
  }
}

async function toggleReaction(emoji: string) {
  if (!canInteract.value) return;
  busy.value = true;
  try {
    const existing = reactions.value.find((r) => r.reaction_type === emoji);
    if (existing?.reacted) {
      await promiseToast(
        issueApi.removeReaction(props.workspaceId, props.projectId, props.issueId, emoji),
        { loading: "取消反应...", success: () => "已取消", error: () => "操作失败" },
      );
    } else {
      await promiseToast(
        issueApi.addReaction(props.workspaceId, props.projectId, props.issueId, emoji),
        { loading: "添加反应...", success: () => "已添加", error: () => "操作失败" },
      );
    }
    await load();
  } finally {
    busy.value = false;
    emojiPanelOpen.value = false;
  }
}

async function castVote(v: 1 | -1) {
  if (!canInteract.value) return;
  busy.value = true;
  try {
    if (vote.value.voted === v) {
      // 再次点击取消投票
      await promiseToast(
        issueApi.removeVote(props.workspaceId, props.projectId, props.issueId),
        { loading: "撤销投票...", success: () => "已撤销", error: () => "操作失败" },
      );
    } else {
      await promiseToast(
        issueApi.vote(props.workspaceId, props.projectId, props.issueId, v),
        { loading: "投票中...", success: () => (v === 1 ? "已赞成" : "已反对"), error: () => "操作失败" },
      );
    }
    await load();
  } finally {
    busy.value = false;
  }
}

async function toggleWatch() {
  if (!canInteract.value) return;
  busy.value = true;
  try {
    if (watching.value) {
      await promiseToast(
        issueApi.unwatch(props.workspaceId, props.projectId, props.issueId),
        { loading: "取消关注...", success: () => "已取消关注", error: () => "操作失败" },
      );
      watching.value = false;
    } else {
      await promiseToast(
        issueApi.watch(props.workspaceId, props.projectId, props.issueId),
        { loading: "关注中...", success: () => "已关注，变更将通知你", error: () => "操作失败" },
      );
      watching.value = true;
    }
  } finally {
    busy.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="social-bar" aria-label="社交反馈">
    <!-- 投票 -->
    <div class="social-bar__vote" role="group" aria-label="投票">
      <button
        class="vote-btn"
        :class="{ 'vote-btn--active': vote.voted === 1 }"
        :disabled="!canInteract"
        title="赞成"
        @click="castVote(1)"
      >
        <span class="vote-arrow">▲</span>
        <span class="vote-count">{{ vote.upvotes }}</span>
      </button>
      <button
        class="vote-btn"
        :class="{ 'vote-btn--active vote-btn--down': vote.voted === -1 }"
        :disabled="!canInteract"
        title="反对"
        @click="castVote(-1)"
      >
        <span class="vote-arrow">▼</span>
        <span class="vote-count">{{ vote.downvotes }}</span>
      </button>
      <span v-if="vote.score !== 0" class="vote-score" :class="{ 'vote-score--neg': vote.score < 0 }">
        {{ vote.score > 0 ? "+" : "" }}{{ vote.score }}
      </span>
    </div>

    <span class="divider" />

    <!-- 表情反应 -->
    <div class="social-bar__reactions" role="group" aria-label="表情反应">
      <button
        v-for="r in reactions"
        :key="r.reaction_type"
        class="reaction-chip"
        :class="{ 'reaction-chip--active': r.reacted }"
        :disabled="!canInteract"
        :title="r.reacted ? '取消反应' : `添加 ${r.reaction_type}`"
        @click="toggleReaction(r.reaction_type)"
      >
        <span class="reaction-emoji">{{ r.reaction_type }}</span>
        <span class="reaction-count">{{ r.count }}</span>
      </button>

      <!-- 展开表情面板 -->
      <div class="reaction-add">
        <button
          class="reaction-add-btn"
          :disabled="!canInteract"
          title="添加表情"
          @click="emojiPanelOpen = !emojiPanelOpen"
        >
          😊<span class="reaction-plus">+</span>
        </button>
        <Transition name="pop">
          <div v-if="emojiPanelOpen" class="emoji-panel" @click.stop>
            <button
              v-for="e in emojiChoices"
              :key="e"
              class="emoji-item"
              :class="{ 'emoji-item--active': reactions.find((r) => r.reaction_type === e)?.reacted }"
              @click="toggleReaction(e)"
            >
              {{ e }}
            </button>
          </div>
        </Transition>
      </div>
    </div>

    <span class="divider" />

    <!-- 关注 -->
    <button
      class="watch-btn"
      :class="{ 'watch-btn--active': watching }"
      :disabled="!canInteract"
      @click="toggleWatch"
    >
      <span class="watch-icon" aria-hidden="true">👁</span>
      <span>{{ watching ? "已关注" : "关注" }}</span>
    </button>
  </div>
</template>

<style scoped>
.social-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 2px 0;
}

/* ---- 投票 ---- */
.social-bar__vote {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.vote-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-surface-1, #fff);
  color: var(--txt-secondary, #6b7280);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.vote-btn:hover:not(:disabled) {
  border-color: var(--brand-300, #93c5fd);
  color: var(--brand-600, #2563eb);
  background: var(--brand-50, #eff6ff);
}
.vote-btn--active {
  border-color: var(--brand-400, #60a5fa);
  background: var(--brand-50, #eff6ff);
  color: var(--brand-600, #2563eb);
  font-weight: 600;
}
.vote-btn--down.vote-btn--active {
  border-color: var(--danger-300, #fca5a5);
  background: var(--danger-50, #fef2f2);
  color: var(--danger-600, #dc2626);
}
.vote-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.vote-arrow {
  font-size: 9px;
  line-height: 1;
}
.vote-count {
  font-variant-numeric: tabular-nums;
}
.vote-score {
  font-size: 12px;
  font-weight: 600;
  color: var(--success-600, #059669);
  margin-left: 2px;
  font-variant-numeric: tabular-nums;
}
.vote-score--neg {
  color: var(--danger-600, #dc2626);
}

.divider {
  width: 1px;
  height: 18px;
  background: var(--border-subtle, #e5e7eb);
  flex-shrink: 0;
}

/* ---- 表情反应 ---- */
.social-bar__reactions {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.reaction-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: 999px;
  background: var(--bg-surface-1, #fff);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.reaction-chip:hover:not(:disabled) {
  border-color: var(--brand-300, #93c5fd);
  background: var(--bg-layer-1-hover, rgba(0, 0, 0, 0.03));
}
.reaction-chip--active {
  border-color: var(--brand-400, #60a5fa);
  background: var(--brand-50, #eff6ff);
}
.reaction-chip:disabled {
  cursor: not-allowed;
}
.reaction-emoji {
  font-size: 13px;
}
.reaction-count {
  color: var(--txt-secondary, #6b7280);
  font-variant-numeric: tabular-nums;
}

.reaction-add {
  position: relative;
}
.reaction-add-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px dashed var(--border-subtle, #e5e7eb);
  border-radius: 999px;
  background: transparent;
  font-size: 13px;
  cursor: pointer;
  position: relative;
  transition: all 0.15s;
}
.reaction-add-btn:hover:not(:disabled) {
  border-color: var(--brand-300, #93c5fd);
  background: var(--bg-layer-1-hover, rgba(0, 0, 0, 0.03));
}
.reaction-add-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
.reaction-plus {
  position: absolute;
  right: -2px;
  bottom: -3px;
  font-size: 9px;
  color: var(--brand-500, #3b82f6);
  background: var(--bg-surface-1, #fff);
  border-radius: 50%;
}

/* 表情选择面板 */
.emoji-panel {
  position: absolute;
  bottom: 34px;
  left: 0;
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 4px;
  padding: 8px;
  background: var(--bg-surface-1, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  box-shadow: var(--shadow-popover, 0 4px 16px rgba(0, 0, 0, 0.08));
  z-index: 50;
}
.emoji-item {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm, 6px);
  font-size: 16px;
  cursor: pointer;
  transition: background 0.1s;
}
.emoji-item:hover {
  background: var(--bg-layer-1-hover, rgba(0, 0, 0, 0.05));
}
.emoji-item--active {
  background: var(--brand-50, #eff6ff);
}
.pop-enter-active,
.pop-leave-active {
  transition: all 0.15s ease;
}
.pop-enter-from,
.pop-leave-to {
  opacity: 0;
  transform: translateY(4px);
}

/* ---- 关注 ---- */
.watch-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-surface-1, #fff);
  color: var(--txt-secondary, #6b7280);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.watch-btn:hover:not(:disabled) {
  border-color: var(--brand-300, #93c5fd);
  color: var(--brand-600, #2563eb);
}
.watch-btn--active {
  border-color: var(--brand-400, #60a5fa);
  background: var(--brand-50, #eff6ff);
  color: var(--brand-600, #2563eb);
}
.watch-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
.watch-icon {
  font-size: 12px;
}
</style>
