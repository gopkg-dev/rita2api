import { describe, expect, it } from 'vitest'

import {
  createMemoryStorage,
  loadHiddenTaskIdsSnapshot,
  saveHiddenTaskIdsSnapshot,
} from './hidden-task-storage'

describe('hidden task local snapshot', () => {
  it('persists deleted task ids synchronously', () => {
    const storage = createMemoryStorage()

    saveHiddenTaskIdsSnapshot(storage, ['task-a', 'task-b'])

    expect(loadHiddenTaskIdsSnapshot(storage)).toEqual(['task-a', 'task-b'])
  })

  it('falls back to an empty array when stored data is invalid', () => {
    const storage = createMemoryStorage()
    storage.setItem('rita-hidden-task-ids', '{broken json')

    expect(loadHiddenTaskIdsSnapshot(storage)).toEqual([])
  })
})
