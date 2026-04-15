import { describe, expect, it } from 'vitest'

import { ratioOptions, resolutionOptions } from './generation-options'
import { filterHiddenTasks, mergeTasksByFreshness } from './task-cache'
import type { GenerationTask } from '../types'

function makeTask(overrides: Partial<GenerationTask>): GenerationTask {
  return {
    id: 'task-1',
    prompt: 'chrome flower',
    ratio: '1:1',
    resolution: '1K',
    imageNum: 1,
    status: 'queued',
    parentMessageId: '',
    messageId: '',
    resultUrl: '',
    errorMessage: '',
    isPublic: false,
    createdAt: '2026-04-14T00:00:00Z',
    updatedAt: '2026-04-14T00:00:00Z',
    finishedAt: '',
    ...overrides,
  }
}

describe('mergeTasksByFreshness', () => {
  it('keeps the fresher server task when ids collide', () => {
    const localTasks = [
      makeTask({
        id: 'task-1',
        status: 'running',
        updatedAt: '2026-04-14T00:00:01Z',
      }),
    ]

    const serverTasks = [
      makeTask({
        id: 'task-1',
        status: 'succeeded',
        resultUrl: 'https://img.example/final.png',
        updatedAt: '2026-04-14T00:00:05Z',
      }),
    ]

    const merged = mergeTasksByFreshness(localTasks, serverTasks)

    expect(merged).toHaveLength(1)
    expect(merged[0].status).toBe('succeeded')
    expect(merged[0].resultUrl).toBe('https://img.example/final.png')
  })

  it('keeps local queued tasks that do not exist on the server yet', () => {
    const localTasks = [
      makeTask({
        id: 'task-local',
        status: 'queued',
        updatedAt: '2026-04-14T00:00:03Z',
      }),
    ]

    const merged = mergeTasksByFreshness(localTasks, [])

    expect(merged).toHaveLength(1)
    expect(merged[0].id).toBe('task-local')
  })

  it('filters tasks hidden by local deletion', () => {
    const tasks = [
      makeTask({ id: 'task-1' }),
      makeTask({ id: 'task-2', updatedAt: '2026-04-14T00:00:03Z' }),
    ]

    const filtered = filterHiddenTasks(tasks, ['task-2'])

    expect(filtered).toHaveLength(1)
    expect(filtered[0].id).toBe('task-1')
  })

  it('exposes the supported ratio buttons for the homepage composer', () => {
    expect(ratioOptions).toEqual(['1:1', '2:3', '3:2', '3:4', '16:9', '4:3', '9:16'])
  })

  it('exposes the supported resolution buttons for the homepage composer', () => {
    expect(resolutionOptions).toEqual(['1K', '2K', '4K'])
  })
})
