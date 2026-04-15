import { openDB } from 'idb'

import {
  getBrowserStorage,
  loadHiddenTaskIdsSnapshot,
  saveHiddenTaskIdsSnapshot,
} from './hidden-task-storage'
import type { GenerationTask } from '../types'

const DB_NAME = 'rati-studio'
const STORE_NAME = 'app-cache'

interface DraftState {
  prompt: string
  ratio: string
  resolution: string
  imageNum: number
}

async function getDB() {
  return openDB(DB_NAME, 1, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME)
      }
    },
  })
}

export async function loadCachedTasks(): Promise<GenerationTask[]> {
  const db = await getDB()
  return ((await db.get(STORE_NAME, 'tasks')) as GenerationTask[] | undefined) ?? []
}

export async function saveCachedTasks(tasks: GenerationTask[]): Promise<void> {
  const db = await getDB()
  await db.put(STORE_NAME, tasks, 'tasks')
}

export async function loadDraftState(): Promise<DraftState | undefined> {
  const db = await getDB()
  return (await db.get(STORE_NAME, 'draft')) as DraftState | undefined
}

export async function loadHiddenTaskIds(): Promise<string[]> {
  const browserStorage = getBrowserStorage()
  if (browserStorage) {
    const snapshot = loadHiddenTaskIdsSnapshot(browserStorage)
    if (snapshot.length > 0) {
      return snapshot
    }
  }

  const db = await getDB()
  return ((await db.get(STORE_NAME, 'hiddenTaskIds')) as string[] | undefined) ?? []
}

export async function saveDraftState(draft: DraftState): Promise<void> {
  const db = await getDB()
  await db.put(STORE_NAME, draft, 'draft')
}

export async function saveHiddenTaskIds(hiddenTaskIds: string[]): Promise<void> {
  const browserStorage = getBrowserStorage()
  if (browserStorage) {
    saveHiddenTaskIdsSnapshot(browserStorage, hiddenTaskIds)
  }

  const db = await getDB()
  await db.put(STORE_NAME, hiddenTaskIds, 'hiddenTaskIds')
}
