<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">个人资料</h1>
      <p class="text-sm text-muted-foreground mt-1">
        您的个人资料信息。这些信息将显示在系统各处。
      </p>
    </div>

    <div class="rounded-lg border p-4 space-y-4">
      <div class="flex items-center gap-4">
        <div class="h-16 w-16 rounded-full bg-primary/10 flex items-center justify-center text-xl font-bold text-primary">
          {{ initials }}
        </div>
        <div>
          <p class="font-semibold">{{ displayName }}</p>
          <p class="text-sm text-muted-foreground">{{ email }}</p>
        </div>
      </div>
      <div v-if="workspace" class="text-sm">
        <span class="text-muted-foreground">当前工作空间：</span>
        <span class="font-medium">{{ workspace }}</span>
      </div>
    </div>

    <div class="rounded-lg border p-4">
      <p class="text-sm text-muted-foreground">
        资料编辑功能即将上线。
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

const displayName = ref("");
const email = ref("");
const workspace = ref("");

const initials = computed(() => {
  const name = displayName.value || email.value || "?";
  return name.charAt(0).toUpperCase();
});

onMounted(async () => {
  try {
    const res = await fetch("/api/v1/me", { credentials: "include" });
    if (res.ok) {
      const data = await res.json();
      displayName.value = data.display_name || data.name || "";
      email.value = data.email || "";
      workspace.value = data.workspace?.name || "";
    }
  } catch {
    // 未登录时忽略
  }
});
</script>
