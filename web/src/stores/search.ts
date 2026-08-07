import { defineStore } from 'pinia'
import { ref } from 'vue'
import { searchApi, type SearchResponse } from '@/api/services/search'

export const useSearchStore = defineStore('search', () => {
  const query = ref('')
  const results = ref<SearchResponse | null>(null)
  const loading = ref(false)
  const open = ref(false)

  async function search(wsId: number | string, q: string) {
    if (!q.trim()) {
      results.value = null
      return
    }
    query.value = q
    loading.value = true
    try {
      results.value = await searchApi.searchWorkspace(wsId, { q, limit: 10 })
    } catch {
      results.value = null
    } finally {
      loading.value = false
    }
  }

  function toggle() {
    open.value = !open.value
    if (!open.value) {
      results.value = null
      query.value = ''
    }
  }

  function clear() {
    query.value = ''
    results.value = null
    open.value = false
  }

  return { query, results, loading, open, search, toggle, clear }
})
