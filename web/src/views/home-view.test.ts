// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { computed, defineComponent } from 'vue'
import { describe, expect, it, vi } from 'vitest'

const { initializeMock, tasksMock } = vi.hoisted(() => ({
  initializeMock: vi.fn().mockResolvedValue(undefined),
  tasksMock: [
    { id: '1', status: 'succeeded' },
    { id: '2', status: 'queued' },
    { id: '3', status: 'failed' },
  ],
}))

vi.mock('../stores/generation', () => ({
  useGenerationStore: () => ({
    initialize: initializeMock,
    tasks: tasksMock,
    pendingTasks: computed(() =>
      tasksMock.filter((task) => task.status === 'queued' || task.status === 'running'),
    ),
  }),
}))

vi.mock('../components/SiteShell.vue', () => ({
  default: defineComponent({
    name: 'SiteShellStub',
    template: '<div class="site-shell"><slot /></div>',
  }),
}))

vi.mock('../components/PromptComposer.vue', () => ({
  default: defineComponent({
    name: 'PromptComposerStub',
    template: '<div data-testid="composer-stub">composer</div>',
  }),
}))

import HomeView from './HomeView.vue'

describe('HomeView', () => {
  it('renders homepage floating task stats with side access cards', () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: defineComponent({
            props: {
              to: {
                type: String,
                required: true,
              },
            },
            template: '<a :href="to"><slot /></a>',
          }),
        },
      },
    })

    expect(wrapper.find('.home-hero').exists()).toBe(true)
    expect(wrapper.find('.home-floating-stats').exists()).toBe(true)
    expect(wrapper.find('.home-stage__overlay').exists()).toBe(true)
    expect(wrapper.find('.home-logo-card').exists()).toBe(true)
    expect(wrapper.get('.home-logo-card img').attributes('alt')).toBe('Nano banana')
    expect(wrapper.findAll('.home-stat-card')).toHaveLength(4)
    expect(wrapper.findAll('.home-side-link-card')).toHaveLength(3)
    expect(wrapper.text()).toContain('从一句提示词开始')
    expect(wrapper.text()).toContain('总任务')
    expect(wrapper.text()).toContain('排队中')
    expect(wrapper.text()).toContain('生成中')
    expect(wrapper.text()).toContain('已完成')
    expect(wrapper.find('h1').exists()).toBe(false)
    expect(wrapper.find('[data-testid="composer-stub"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('查看任务')
    expect(wrapper.text()).toContain('浏览作品')
    expect(wrapper.text()).toContain('API接口')
    expect(wrapper.text()).toContain('查看文档')
  })
})
