import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastTone = 'info' | 'success' | 'warning' | 'error'

export interface ToastItem {
  id: number
  title: string
  description: string
  tone: ToastTone
}

let nextToastID = 1

export const useToastStore = defineStore('toast', () => {
  const items = ref<ToastItem[]>([])

  function push(input: Omit<ToastItem, 'id'>, duration = 2800) {
    const item: ToastItem = {
      id: nextToastID++,
      ...input,
    }

    items.value = [...items.value, item]

    window.setTimeout(() => {
      dismiss(item.id)
    }, duration)
  }

  function dismiss(id: number) {
    items.value = items.value.filter((item) => item.id !== id)
  }

  return {
    items,
    push,
    dismiss,
  }
})
