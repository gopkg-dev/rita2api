<script setup lang="ts">
import { storeToRefs } from 'pinia'

import { useToastStore } from '../stores/toast'

const toastStore = useToastStore()
const { items } = storeToRefs(toastStore)
</script>

<template>
  <div class="toast-viewport" aria-live="polite" aria-atomic="true">
    <article
      v-for="item in items"
      :key="item.id"
      class="toast-card"
      :data-tone="item.tone"
    >
      <div class="toast-card__copy">
        <strong>{{ item.title }}</strong>
        <p>{{ item.description }}</p>
      </div>
      <button
        type="button"
        class="toast-card__dismiss"
        title="关闭提示"
        @click="toastStore.dismiss(item.id)"
      >
        关闭
      </button>
    </article>
  </div>
</template>
