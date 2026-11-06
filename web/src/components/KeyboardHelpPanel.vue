<script setup lang="ts">
/**
 * KeyboardHelpPanel — 快捷键帮助弹窗（"?" 触发，Esc 关闭）
 *
 * 分组展示：通用 / 导航（G 前缀）/ 创建（N 前缀）/ 工作项操作
 */
import DynamicIcon from "@/components/DynamicIcon.vue";
import { computed } from "vue"

const props = defineProps<{
  shortcuts: { key: string; label: string; description: string; prefix?: string }[]
}>()
const emit = defineEmits<{ (e: "close"): void }>()

interface Group {
  title: string
  items: typeof props.shortcuts
}

const groups = computed<Group[]>(() => {
  const general = props.shortcuts.filter((s) => !s.prefix && !["S", "P", "A", "I", "L", "E", "M", "C"].includes(s.label))
  const nav = props.shortcuts.filter((s) => s.prefix === "g")
  const create = props.shortcuts.filter((s) => s.prefix === "n")
  const workItem = props.shortcuts.filter(
    (s) => !s.prefix && ["S", "P", "A", "I", "L", "E", "M", "C"].includes(s.label),
  )

  const out: Group[] = []
  if (general.length) out.push({ title: "通用", items: general })
  if (nav.length) out.push({ title: "导航（G 前缀）", items: nav })
  if (create.length) out.push({ title: "创建（N 前缀）", items: create })
  if (workItem.length) out.push({ title: "工作项操作", items: workItem })
  return out
})
</script>

<template>
  <Teleport to="body">
    <div class="kh-overlay" role="dialog" aria-modal="true" aria-label="键盘快捷键帮助" @click.self="emit('close')">
      <div class="kh-card">
        <div class="kh-head">
          <h2 class="kh-title">键盘快捷键</h2>
          <button class="kh-close" aria-label="关闭" @click="emit('close')"><DynamicIcon name="X" :size="16" /></button>
        </div>

        <div v-for="g in groups" :key="g.title" class="kh-group">
          <div class="kh-group-title">{{ g.title }}</div>
          <ul class="kh-list">
            <li v-for="s in g.items" :key="`${s.prefix ?? ''}-${s.key}`" class="kh-row">
              <kbd class="kh-key">{{ s.label }}</kbd>
              <span class="kh-desc">{{ s.description }}</span>
            </li>
          </ul>
        </div>

        <p class="kh-tip">按 <kbd class="kh-key">?</kbd> 随时打开此面板</p>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.kh-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
}
.kh-card {
  width: min(620px, 92vw);
  max-height: 80vh;
  overflow-y: auto;
  background: var(--bg-elevated);
  border-radius: 12px;
  padding: 20px 22px;
  box-shadow: 0 18px 60px rgba(0, 0, 0, 0.28);
}
.kh-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.kh-title {
  font-size: 16px;
  font-weight: 600;
}
.kh-close {
  border: none;
  background: transparent;
  font-size: 18px;
  cursor: pointer;
}
.kh-group {
  margin-bottom: 16px;
}
.kh-group:last-of-type {
  margin-bottom: 8px;
}
.kh-group-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 8px;
}
.kh-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.kh-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 0;
}
.kh-key {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  padding: 2px 8px;
  font-family: ui-monospace, monospace;
  font-size: 12px;
  color: var(--txt-primary);
  background: var(--bg-subtle);
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
}
.kh-desc {
  font-size: 13px;
  color: var(--txt-secondary);
}
.kh-tip {
  margin: 12px 0 0;
  font-size: 12px;
  color: var(--txt-tertiary);
  text-align: center;
}
</style>
