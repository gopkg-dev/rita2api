const HIDDEN_TASK_IDS_KEY = 'rita-hidden-task-ids'

export interface StorageLike {
  getItem(key: string): string | null
  setItem(key: string, value: string): void
}

export function loadHiddenTaskIdsSnapshot(storage: StorageLike): string[] {
  const raw = storage.getItem(HIDDEN_TASK_IDS_KEY)
  if (!raw) {
    return []
  }

  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed.filter((value): value is string => typeof value === 'string') : []
  } catch {
    return []
  }
}

export function saveHiddenTaskIdsSnapshot(storage: StorageLike, hiddenTaskIds: string[]): void {
  storage.setItem(HIDDEN_TASK_IDS_KEY, JSON.stringify(hiddenTaskIds))
}

export function getBrowserStorage(): StorageLike | undefined {
  if (typeof window === 'undefined' || !window.localStorage) {
    return undefined
  }

  return window.localStorage
}

export function createMemoryStorage(): StorageLike {
  const data = new Map<string, string>()

  return {
    getItem(key: string) {
      return data.get(key) ?? null
    },
    setItem(key: string, value: string) {
      data.set(key, value)
    },
  }
}
