import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { createGeneration, fetchHistory, fetchTask, retryTask } from '../lib/api'
import {
  loadCachedTasks,
  loadDraftState,
  loadHiddenTaskIds,
  saveCachedTasks,
  saveDraftState,
  saveHiddenTaskIds,
} from '../lib/cache'
import { filterHiddenTasks, mergeTasksByFreshness } from '../lib/task-cache'
import { useSessionStore } from './session'
import type { GenerationTask, TaskEventPayload } from '../types'

const streams = new Map<string, EventSource>()

export const useGenerationStore = defineStore('generation', () => {
  const tasks = ref<GenerationTask[]>([])
  const hiddenTaskIds = ref<string[]>([])
  const isSubmitting = ref(false)
  const hydrated = ref(false)
  const draft = ref({
    prompt: '',
    ratio: '1:1',
    resolution: '1K',
    imageNum: 1,
  })

  const featuredTask = computed(() => tasks.value.find((task) => task.status === 'succeeded') ?? tasks.value[0])
  const pendingTasks = computed(() =>
    tasks.value.filter((task) => task.status === 'queued' || task.status === 'running'),
  )

  function stopStream(taskId: string) {
    const stream = streams.get(taskId)
    if (!stream) {
      return
    }

    stream.close()
    streams.delete(taskId)
  }

  function applyTask(task: GenerationTask) {
    if (hiddenTaskIds.value.includes(task.id)) {
      return
    }

    tasks.value = mergeTasksByFreshness(tasks.value, [task])
    void saveCachedTasks(tasks.value)
  }

  async function initialize() {
    if (hydrated.value) {
      return
    }

    const sessionStore = useSessionStore()
    await sessionStore.bootstrap()

    const [cachedTasks, cachedDraft, storedHiddenTaskIds] = await Promise.all([
      loadCachedTasks(),
      loadDraftState(),
      loadHiddenTaskIds(),
    ])
    hiddenTaskIds.value = storedHiddenTaskIds
    tasks.value = filterHiddenTasks(cachedTasks, hiddenTaskIds.value)

    if (cachedDraft) {
      draft.value = cachedDraft
    } else {
      draft.value.ratio = sessionStore.defaultRatio
      draft.value.resolution = sessionStore.defaultResolution
      draft.value.imageNum = sessionStore.defaultImageNum
    }

    await refreshHistory()
    reconnectPendingTasks()
    hydrated.value = true
  }

  async function refreshHistory() {
    const history = await fetchHistory()
    tasks.value = filterHiddenTasks(
      mergeTasksByFreshness(tasks.value, history.items),
      hiddenTaskIds.value,
    )
    await saveCachedTasks(tasks.value)
  }

  async function saveDraft() {
    await saveDraftState(draft.value)
  }

  function updateDraft(patch: Partial<typeof draft.value>) {
    draft.value = { ...draft.value, ...patch }
    void saveDraft()
  }

  async function submit() {
    const trimmedPrompt = draft.value.prompt.trim()
    if (!trimmedPrompt) {
      throw new Error('请输入提示词')
    }

    isSubmitting.value = true
    try {
      const task = await createGeneration({
        prompt: trimmedPrompt,
        ratio: draft.value.ratio,
        resolution: draft.value.resolution,
        imageNum: draft.value.imageNum,
      })
      applyTask(task)
      subscribeToTask(task.id)
    } finally {
      isSubmitting.value = false
    }
  }

  async function refreshTask(taskId: string) {
    const task = await fetchTask(taskId)
    applyTask(task)
  }

  async function retry(taskId: string) {
    const task = await retryTask(taskId)
    applyTask(task)
    subscribeToTask(task.id)
  }

  async function deleteLocalTask(taskId: string) {
    stopStream(taskId)

    if (!hiddenTaskIds.value.includes(taskId)) {
      hiddenTaskIds.value = [...hiddenTaskIds.value, taskId]
    }

    tasks.value = tasks.value.filter((task) => task.id !== taskId)

    await saveHiddenTaskIds(hiddenTaskIds.value)
    await saveCachedTasks(tasks.value)
  }

  function reconnectPendingTasks() {
    for (const task of tasks.value) {
      if (task.status === 'queued' || task.status === 'running') {
        subscribeToTask(task.id)
      }
    }
  }

  function subscribeToTask(taskId: string) {
    if (streams.has(taskId)) {
      return
    }

    if (hiddenTaskIds.value.includes(taskId)) {
      return
    }

    const stream = new EventSource(`/api/v1/generations/${taskId}/stream`, { withCredentials: true })
    streams.set(taskId, stream)

    stream.onmessage = (message) => {
      const payload = JSON.parse(message.data) as TaskEventPayload
      applyTask(payload.task)

      if (
        payload.type === 'done' ||
        payload.task.status === 'succeeded' ||
        payload.task.status === 'failed'
      ) {
        stream.close()
        streams.delete(taskId)
      }
    }

    stream.onerror = async () => {
      stream.close()
      streams.delete(taskId)
      await refreshTask(taskId).catch(() => undefined)
    }
  }

  return {
    tasks,
    draft,
    isSubmitting,
    hydrated,
    featuredTask,
    pendingTasks,
    hiddenTaskIds,
    initialize,
    refreshHistory,
    updateDraft,
    submit,
    retry,
    deleteLocalTask,
  }
})
