// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { describe, expect, it, vi } from 'vitest'

const { loadMock, retryMock } = vi.hoisted(() => ({
  loadMock: vi.fn().mockResolvedValue(undefined),
  retryMock: vi.fn(),
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
    },
    template: '<article class="task-card-stub">{{ task.id }}</article>',
  }),
}))

vi.mock('../stores/gallery', () => ({
  useGalleryStore: () => ({
    load: loadMock,
    errorMessage: '',
    isLoading: false,
    items: [
      { id: 'task-1' },
      { id: 'task-2' },
      { id: 'task-3' },
    ],
  }),
}))

vi.mock('../stores/generation', () => ({
  useGenerationStore: () => ({
    retry: retryMock,
  }),
}))

import GalleryView from './GalleryView.vue'

describe('GalleryView', () => {
  it('renders gallery items inside a masonry stage', () => {
    const wrapper = mount(GalleryView)

    expect(wrapper.find('.gallery-grid').exists()).toBe(true)
    expect(wrapper.find('.gallery-grid--masonry').exists()).toBe(true)
    expect(wrapper.findAll('.task-card-stub')).toHaveLength(3)
  })
})
