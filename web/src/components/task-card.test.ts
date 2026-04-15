// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import TaskCard from './TaskCard.vue'
import type { GenerationTask } from '../types'

function createTask(overrides: Partial<GenerationTask> = {}): GenerationTask {
  return {
    id: 'task-1',
    prompt: 'test prompt',
    ratio: '16:9',
    resolution: '1K',
    imageNum: 1,
    status: 'succeeded',
    parentMessageId: '',
    messageId: '',
    resultUrl: 'https://example.com/test.png',
    errorMessage: '',
    isPublic: false,
    createdAt: '',
    updatedAt: '',
    finishedAt: '',
    ...overrides,
  }
}

describe('TaskCard', () => {
  it('uses the task ratio to size the preview area', () => {
    const wrapper = mount(TaskCard, {
      props: {
        task: createTask(),
      },
    })

    const preview = wrapper.get('.task-card__preview')
    expect(preview.attributes('style')).toContain('aspect-ratio: 16 / 9;')
  })

  it('renders the view-large action with button styling', () => {
    const wrapper = mount(TaskCard, {
      props: {
        task: createTask(),
      },
    })

    const action = wrapper.get('a[href="https://example.com/test.png"]')
    expect(action.classes()).toContain('task-card__button')
  })

  it('uses the compact preview mode for gallery cards', () => {
    const wrapper = mount(TaskCard, {
      props: {
        task: createTask(),
        compact: true,
      },
    })

    const preview = wrapper.get('.task-card__preview')
    const image = wrapper.get('.task-card__image')
    expect(preview.classes()).toContain('task-card__preview--compact')
    expect(preview.attributes('style')).toBeUndefined()
    expect(image.classes()).toContain('task-card__image--contain')
  })

  it('uses the adaptive preview mode for queue cards', () => {
    const wrapper = mount(TaskCard, {
      props: {
        task: createTask(),
        adaptivePreview: true,
      },
    })

    const card = wrapper.get('.task-card')
    const preview = wrapper.get('.task-card__preview')
    const image = wrapper.get('.task-card__image')
    expect(card.classes()).toContain('task-card--adaptive')
    expect(preview.classes()).toContain('task-card__preview--adaptive')
    expect(preview.attributes('style')).toBeUndefined()
    expect(image.classes()).toContain('task-card__image--contain')
  })
})
