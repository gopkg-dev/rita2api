// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { computed, defineComponent } from 'vue'
import { describe, expect, it, vi } from 'vitest'

const storeMock = vi.hoisted(() => ({
  tasks: [
    { id: '1', status: 'succeeded', prompt: 'done' },
    { id: '2', status: 'queued', prompt: 'queued' },
    { id: '3', status: 'failed', prompt: 'failed' },
  ],
  initialize: vi.fn().mockResolvedValue(undefined),
  refreshHistory: vi.fn().mockResolvedValue(undefined),
  retry: vi.fn().mockResolvedValue(undefined),
  deleteLocalTask: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../stores/generation', () => ({
  useGenerationStore: () => ({
    tasks: storeMock.tasks,
    pendingTasks: computed(() => storeMock.tasks.filter((task) => task.status === 'queued' || task.status === 'running')),
    initialize: storeMock.initialize,
    refreshHistory: storeMock.refreshHistory,
    retry: storeMock.retry,
    deleteLocalTask: storeMock.deleteLocalTask,
  }),
}))

vi.mock('../stores/toast', () => ({
  useToastStore: () => ({
    push: vi.fn(),
  }),
}))

vi.mock('../components/SiteShell.vue', () => ({
  default: defineComponent({
    name: 'SiteShellStub',
    template: '<div class="site-shell"><slot /></div>',
  }),
}))

vi.mock('../components/TaskCard.vue', () => ({
  default: defineComponent({
    name: 'TaskCardStub',
    props: {
      task: {
        type: Object,
        required: true,
      },
      adaptivePreview: {
        type: Boolean,
        default: false,
      },
    },
    template:
      '<div class="task-card-stub" :class="{ \'task-card-stub--adaptive\': adaptivePreview }">{{ task.prompt }}</div>',
  }),
}))

import QueueView from './QueueView.vue'

describe('QueueView', () => {
  it('renders a single-column queue layout with filters and tasks', () => {
    const wrapper = mount(QueueView)

    expect(wrapper.find('.queue-stage').exists()).toBe(true)
    expect(wrapper.find('.queue-toolbar').exists()).toBe(true)
    expect(wrapper.findAll('.queue-filter')).toHaveLength(5)
    expect(wrapper.findAll('.task-card-stub')).toHaveLength(3)
    expect(wrapper.findAll('.task-card-stub--adaptive')).toHaveLength(3)
  })
})
