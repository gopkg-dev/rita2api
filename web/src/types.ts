export type TaskStatus = 'queued' | 'running' | 'succeeded' | 'failed'

export interface GenerationTask {
  id: string
  prompt: string
  ratio: string
  resolution: string
  imageNum: number
  status: TaskStatus
  parentMessageId: string
  messageId: string
  resultUrl: string
  errorMessage: string
  isPublic: boolean
  createdAt: string
  updatedAt: string
  finishedAt: string
}

export interface TaskEventPayload {
  type: string
  task: GenerationTask
}

export interface TaskListResponse {
  total: number
  items: GenerationTask[]
}

export interface BootstrapPayload {
  session: {
    token: string
  }
  defaults: {
    ratio: string
    resolution: string
    imageNum: number
  }
  gallery: {
    enabled: boolean
  }
}
