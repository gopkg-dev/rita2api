export type ApiDocMethod = 'GET' | 'POST'

export type ApiDocItem = {
  method: ApiDocMethod
  path: string
  title: string
  description: string
  request?: string[]
  response: string[]
}

export const apiDocGroups: Array<{
  title: string
  description: string
  items: ApiDocItem[]
}> = [
  {
    title: '启动与会话',
    description: '初始化匿名会话并返回站点默认配置。',
    items: [
      {
        method: 'GET',
        path: '/api/v1/bootstrap',
        title: '读取启动数据',
        description: '返回会话 token、默认生成参数和画廊开关。',
        response: [
          '{',
          '  "data": {',
          '    "session": { "token": "session_xxx" },',
          '    "defaults": { "ratio": "1:1", "resolution": "1K", "imageNum": 1 },',
          '    "gallery": { "enabled": true }',
          '  }',
          '}',
        ],
      },
      {
        method: 'POST',
        path: '/api/v1/sessions/anonymous',
        title: '创建匿名会话',
        description: '创建或复用匿名会话，并通过 cookie 保持会话状态。',
        response: ['{', '  "data": { "token": "session_xxx" }', '}'],
      },
    ],
  },
  {
    title: '图片生成',
    description: '提交文本生图任务并获取任务快照。',
    items: [
      {
        method: 'POST',
        path: '/api/v1/generations',
        title: '提交生成任务',
        description: '接收提示词、比例、清晰度和出图数量，返回任务对象。',
        request: [
          '{',
          '  "prompt": "matte aluminum chair, soft side lighting",',
          '  "ratio": "1:1",',
          '  "resolution": "1K",',
          '  "imageNum": 1',
          '}',
        ],
        response: [
          '{',
          '  "data": {',
          '    "id": "task_xxx",',
          '    "status": "queued",',
          '    "prompt": "matte aluminum chair, soft side lighting",',
          '    "ratio": "1:1",',
          '    "resolution": "1K",',
          '    "imageNum": 1,',
          '    "resultUrl": ""',
          '  }',
          '}',
        ],
      },
    ],
  },
  {
    title: '任务状态与重试',
    description: '查询单个任务状态、订阅任务流并执行失败重试。',
    items: [
      {
        method: 'GET',
        path: '/api/v1/generations/:taskId',
        title: '读取任务快照',
        description: '根据任务 ID 返回当前任务完整状态。',
        response: [
          '{',
          '  "data": {',
          '    "id": "task_xxx",',
          '    "status": "running",',
          '    "parentMessageId": "msg_parent_xxx",',
          '    "resultUrl": ""',
          '  }',
          '}',
        ],
      },
      {
        method: 'GET',
        path: '/api/v1/generations/:taskId/stream',
        title: '订阅任务事件流',
        description: '使用 SSE 同步接收任务状态变更。',
        response: [
          'event: message',
          'data: {"type":"running","task":{"id":"task_xxx","status":"running"}}',
        ],
      },
      {
        method: 'POST',
        path: '/api/v1/generations/:taskId/retry',
        title: '重试任务',
        description: '基于原任务参数重新创建一个新的生成任务。',
        response: [
          '{',
          '  "data": {',
          '    "id": "task_retry_xxx",',
          '    "status": "queued"',
          '  }',
          '}',
        ],
      },
    ],
  },
  {
    title: '历史与公开画廊',
    description: '分页读取当前会话历史和公开展示区内容。',
    items: [
      {
        method: 'GET',
        path: '/api/v1/history?page=1&limit=20',
        title: '读取历史任务',
        description: '返回当前匿名会话下的任务列表。',
        response: [
          '{',
          '  "data": {',
          '    "items": [{ "id": "task_xxx", "status": "succeeded" }],',
          '    "page": 1,',
          '    "limit": 20',
          '  }',
          '}',
        ],
      },
      {
        method: 'GET',
        path: '/api/v1/gallery?page=1&limit=20',
        title: '读取公开画廊',
        description: '返回公开作品列表。',
        response: [
          '{',
          '  "data": {',
          '    "items": [{ "id": "task_public_xxx", "resultUrl": "https://..." }],',
          '    "page": 1,',
          '    "limit": 20',
          '  }',
          '}',
        ],
      },
    ],
  },
]

export const sseEvents = [
  { type: 'queued', description: '任务已经进入队列。' },
  { type: 'running', description: '上游已受理任务，开始生成。' },
  { type: 'result', description: '任务拿到结果图链接。' },
  { type: 'failed', description: '任务失败，附带错误信息。' },
  { type: 'done', description: '当前任务流结束。' },
]

export const fetchExample = [
  "const response = await fetch('/api/v1/generations', {",
  "  method: 'POST',",
  "  headers: { 'Content-Type': 'application/json' },",
  '  credentials: "include",',
  '  body: JSON.stringify({',
  "    prompt: 'matte aluminum chair, soft side lighting',",
  "    ratio: '1:1',",
  "    resolution: '1K',",
  '    imageNum: 1,',
  '  }),',
  '})',
  '',
  'const payload = await response.json()',
  'console.log(payload.data.id)',
]

export const eventSourceExample = [
  "const stream = new EventSource('/api/v1/generations/task_xxx/stream', {",
  '  withCredentials: true,',
  '})',
  '',
  "stream.onmessage = (event) => {",
  '  const payload = JSON.parse(event.data)',
  '  console.log(payload.type, payload.task.status)',
  '}',
]
