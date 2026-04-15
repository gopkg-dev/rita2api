import type { BootstrapPayload, GenerationTask, TaskListResponse } from '../types'

interface ApiEnvelope<T> {
  data: T
  error?: string
}

async function requestJSON<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  const payload = (await response.json()) as ApiEnvelope<T>
  if (!response.ok) {
    throw new Error(payload.error || 'request failed')
  }

  return payload.data
}

export function fetchBootstrap() {
  return requestJSON<BootstrapPayload>('/api/v1/bootstrap')
}

export function createGeneration(payload: {
  prompt: string
  ratio: string
  resolution: string
  imageNum: number
}) {
  return requestJSON<GenerationTask>('/api/v1/generations', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function fetchHistory(page = 1, limit = 24) {
  return requestJSON<TaskListResponse>(`/api/v1/history?page=${page}&limit=${limit}`)
}

export function fetchGallery(page = 1, limit = 24) {
  return requestJSON<TaskListResponse>(`/api/v1/gallery?page=${page}&limit=${limit}`)
}

export function fetchTask(taskId: string) {
  return requestJSON<GenerationTask>(`/api/v1/generations/${taskId}`)
}

export function retryTask(taskId: string) {
  return requestJSON<GenerationTask>(`/api/v1/generations/${taskId}/retry`, {
    method: 'POST',
  })
}
