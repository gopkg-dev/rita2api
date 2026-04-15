<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import SiteShell from '../components/SiteShell.vue'
import TaskCard from '../components/TaskCard.vue'
import { useGenerationStore } from '../stores/generation'
import { useToastStore } from '../stores/toast'

type TaskFilter = 'all' | 'queued' | 'running' | 'succeeded' | 'failed'

const generationStore = useGenerationStore()
const toastStore = useToastStore()
const activeFilter = ref<TaskFilter>('all')

const filterOptions = computed<{ value: TaskFilter; label: string }[]>(() => [
  { value: 'all', label: '全部' },
  { value: 'queued', label: '排队中' },
  { value: 'running', label: '生成中' },
  { value: 'succeeded', label: '已完成' },
  { value: 'failed', label: '失败' },
])

const visibleTasks = computed(() => {
  if (activeFilter.value === 'all') {
    return generationStore.tasks
  }

  return generationStore.tasks.filter((task) => task.status === activeFilter.value)
})

onMounted(async () => {
  await generationStore.initialize()
  await generationStore.refreshHistory()
})

async function handleRetry(taskId: string) {
  try {
    await generationStore.retry(taskId)
    toastStore.push({
      title: '已重新提交',
      description: '任务已经重新进入队列，结果会自动同步到当前页面。',
      tone: 'info',
    })
  } catch (error) {
    toastStore.push({
      title: '重试失败',
      description: error instanceof Error ? error.message : '任务重试失败，请稍后再试。',
      tone: 'error',
    }, 3600)
  }
}

async function handleRemove(taskId: string) {
  try {
    await generationStore.deleteLocalTask(taskId)
    toastStore.push({
      title: '本地记录已删除',
      description: '这条记录只会从当前浏览器隐藏，公开画廊和服务端记录保持原样。',
      tone: 'warning',
    })
  } catch (error) {
    toastStore.push({
      title: '删除失败',
      description: error instanceof Error ? error.message : '本地记录删除失败，请稍后重试。',
      tone: 'error',
    }, 3600)
  }
}
</script>

<template>
  <SiteShell>
    <section class="queue-layout">
      <div class="queue-stage">
        <article class="panel queue-toolbar">
          <div class="queue-filters">
            <button
              v-for="filter in filterOptions"
              :key="filter.value"
              type="button"
              class="queue-filter"
              :class="{ 'queue-filter--active': activeFilter === filter.value }"
              :title="`切换到${filter.label}`"
              @click="activeFilter = filter.value"
            >
              {{ filter.label }}
            </button>
          </div>
        </article>

        <div class="task-stack">
          <article v-if="visibleTasks.length === 0" class="panel queue-empty">
            <p class="eyebrow">No tasks</p>
            <h2>当前筛选下还没有任务</h2>
            <p class="panel__lead">你可以回首页继续生成，或者切回“全部”查看完整队列。</p>
          </article>

          <TaskCard
            v-for="task in visibleTasks"
            :key="task.id"
            :task="task"
            :adaptive-preview="true"
            :show-remove-action="true"
            @retry="handleRetry"
            @remove="handleRemove"
          />
        </div>
      </div>
    </section>
  </SiteShell>
</template>
