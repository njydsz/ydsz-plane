<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">API Tokens</h1>
      <p class="text-sm text-muted-foreground mt-1">
        管理您的个人 API 访问令牌。创建的令牌将用于第三方集成与自动化脚本。
      </p>
    </div>

    <div class="rounded-lg border p-4 space-y-3">
      <h2 class="font-semibold">创建新令牌</h2>
      <div class="flex gap-2">
        <input
          v-model="newTokenName"
          type="text"
          placeholder="令牌名称（例如：CI/CD、个人脚本）"
          class="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm"
        />
        <button
          :disabled="!newTokenName || creating"
          class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          @click="createToken"
        >
          {{ creating ? "创建中..." : "创建令牌" }}
        </button>
      </div>
      <div v-if="createdToken" class="rounded-md border p-3">
        <p class="text-sm font-medium">
          令牌已创建（仅此一次可见）：
        </p>
        <code class="mt-1 block break-all rounded p-2 text-xs font-mono">
          {{ createdToken }}
        </code>
      </div>
    </div>

    <div class="space-y-2">
      <h2 class="font-semibold">我的令牌</h2>
      <div v-if="loading" class="text-sm text-muted-foreground">加载中...</div>
      <div v-else-if="tokens.length === 0" class="text-sm text-muted-foreground">
        暂无 API 令牌
      </div>
      <div v-for="token in tokens" :key="token.id" class="flex items-center justify-between rounded-lg border p-3">
        <div>
          <p class="font-medium text-sm">{{ token.name }}</p>
          <p class="text-xs text-muted-foreground">
            创建于 {{ new Date(token.created_at).toLocaleDateString() }}
          </p>
        </div>
        <button
          class="rounded-md border px-3 py-1 text-xs font-medium hover:bg-muted"
          @click="revokeToken(token.id)"
        >
          吊销
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";

interface Token {
  id: number;
  name: string;
  prefix: string;
  created_at: string;
  expires_at?: string;
}

const tokens = ref<Token[]>([]);
const newTokenName = ref("");
const loading = ref(false);
const creating = ref(false);
const createdToken = ref<string | null>(null);

async function fetchTokens() {
  loading.value = true;
  try {
    const res = await fetch("/api/v1/me/api-tokens", { credentials: "include" });
    if (res.ok) tokens.value = await res.json();
  } finally {
    loading.value = false;
  }
}

async function createToken() {
  creating.value = true;
  createdToken.value = null;
  try {
    const res = await fetch("/api/v1/me/api-tokens", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ name: newTokenName.value }),
    });
    if (res.ok) {
      const data = await res.json();
      createdToken.value = data.token;
      newTokenName.value = "";
      await fetchTokens();
    }
  } finally {
    creating.value = false;
  }
}

async function revokeToken(id: number) {
  if (!confirm("确定要吊销该令牌吗？此操作不可恢复。")) return;
  await fetch(`/api/v1/me/api-tokens/${id}`, {
    method: "DELETE",
    credentials: "include",
  });
  await fetchTokens();
}

onMounted(fetchTokens);
</script>
