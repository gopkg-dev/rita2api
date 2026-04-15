<script setup lang="ts">
import { computed } from 'vue'

import type { GenerationTask } from '../types'

const props = defineProps<{
  task: GenerationTask
  compact?: boolean
  adaptivePreview?: boolean
  showRemoveAction?: boolean
}>()

const emit = defineEmits<{
  retry: [taskId: string]
  remove: [taskId: string]
}>()

const statusLabel = computed(() => {
  switch (props.task.status) {
    case 'queued':
      return '排队中'
    case 'running':
      return '生成中'
    case 'succeeded':
      return '已完成'
    case 'failed':
      return '失败'
  }
})

const previewAspectRatio = computed(() => {
  const [width, height] = props.task.ratio.split(':').map((value) => Number.parseFloat(value))

  if (
    Number.isFinite(width) &&
    Number.isFinite(height) &&
    width > 0 &&
    height > 0
  ) {
    return `${width} / ${height}`
  }

  return '1 / 1'
})

const previewStyle = computed(() => {
  if (props.compact || props.adaptivePreview) {
    return undefined
  }

  return {
    aspectRatio: previewAspectRatio.value,
  }
})

</script>

<template>
  <article
    class="task-card"
    :class="{
      'task-card--compact': compact,
      'task-card--adaptive': adaptivePreview,
    }"
  >
    <div
      class="task-card__preview"
      :class="{
        'task-card__preview--compact': compact,
        'task-card__preview--adaptive': adaptivePreview,
      }"
      :style="previewStyle"
    >
      <img
        v-if="task.resultUrl"
        :src="task.resultUrl"
        :alt="task.prompt"
        class="task-card__image"
        :class="{ 'task-card__image--contain': compact || adaptivePreview }"
        loading="lazy"
      />
      <span class="task-card__status" :data-status="task.status">{{ statusLabel }}</span>
    </div>

    <div class="task-card__body">
      <p class="task-card__prompt">{{ task.prompt }}</p>

      <div class="task-card__meta">
        <span>{{ task.ratio }}</span>
        <span>{{ task.resolution }}</span>
        <span>{{ task.imageNum }} 图</span>
      </div>

      <p class="task-card__hint">
        {{
          task.status === 'queued'
            ? '任务已进入队列，结果会自动刷新。'
            : task.status === 'running'
              ? '任务正在生成，保持页面打开可以看到同步结果。'
              : task.status === 'failed'
                ? '失败后可以直接重试，或删除当前浏览器中的本地记录。'
                : '结果已经就绪，支持查看原图和继续生成。'
        }}
      </p>

      <p v-if="task.errorMessage" class="task-card__error">{{ task.errorMessage }}</p>

      <div class="task-card__actions">
        <a
          v-if="task.resultUrl"
          class="task-card__button"
          :href="task.resultUrl"
          target="_blank"
          rel="noreferrer"
        >
          查看大图
        </a>
        <button
          v-if="task.status === 'failed'"
          class="task-card__button"
          type="button"
          @click="emit('retry', task.id)"
        >
          重新生成
        </button>
        <button
          v-if="showRemoveAction"
          class="task-card__ghost"
          type="button"
          @click="emit('remove', task.id)"
        >
          删除本地记录
        </button>
      </div>
    </div>
  </article>
</template>
