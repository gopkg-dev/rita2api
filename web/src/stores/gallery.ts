import { defineStore } from 'pinia'
import { ref } from 'vue'

import { fetchGallery } from '../lib/api'
import type { GenerationTask } from '../types'

export const useGalleryStore = defineStore('gallery', () => {
  const items = ref<GenerationTask[]>([])
  const total = ref(0)
  const isLoading = ref(false)
  const errorMessage = ref('')

  async function load(page = 1, limit = 18) {
    isLoading.value = true
    errorMessage.value = ''
    try {
      const payload = await fetchGallery(page, limit)
      items.value = payload.items
      total.value = payload.total
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : '公开画廊加载失败'
    } finally {
      isLoading.value = false
    }
  }

  return {
    items,
    total,
    isLoading,
    errorMessage,
    load,
  }
})
