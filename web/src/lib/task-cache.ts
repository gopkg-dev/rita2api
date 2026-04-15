import type { GenerationTask } from '../types'

function toTime(value: string): number {
  const time = Date.parse(value)
  return Number.isNaN(time) ? 0 : time
}

export function mergeTasksByFreshness(
  localTasks: GenerationTask[],
  serverTasks: GenerationTask[],
): GenerationTask[] {
  const merged = new Map<string, GenerationTask>()

  for (const task of localTasks) {
    merged.set(task.id, task)
  }

  for (const task of serverTasks) {
    const current = merged.get(task.id)
    if (!current || toTime(task.updatedAt) >= toTime(current.updatedAt)) {
      merged.set(task.id, task)
    }
  }

  return [...merged.values()].sort((left, right) => toTime(right.updatedAt) - toTime(left.updatedAt))
}

export function filterHiddenTasks(tasks: GenerationTask[], hiddenTaskIds: string[]): GenerationTask[] {
  if (hiddenTaskIds.length === 0) {
    return tasks
  }

  const hiddenSet = new Set(hiddenTaskIds)
  return tasks.filter((task) => !hiddenSet.has(task.id))
}
