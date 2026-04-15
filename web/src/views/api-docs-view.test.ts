// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { describe, expect, it, vi } from 'vitest'

vi.mock('../components/SiteShell.vue', () => ({
  default: defineComponent({
    name: 'SiteShellStub',
    template: '<div class="site-shell"><slot /></div>',
  }),
}))

import ApiDocsView from './ApiDocsView.vue'

describe('ApiDocsView', () => {
  it('renders developer-facing API documentation for public endpoints', () => {
    const wrapper = mount(ApiDocsView)

    expect(wrapper.text()).toContain('API 接口文档')
    expect(wrapper.text()).toContain('/api/v1/bootstrap')
    expect(wrapper.text()).toContain('/api/v1/generations')
    expect(wrapper.text()).toContain('/api/v1/generations/:taskId/stream')
    expect(wrapper.text()).toContain('GET')
    expect(wrapper.text()).toContain('POST')
    expect(wrapper.text()).toContain('EventSource')
    expect(wrapper.text()).toContain('fetch')
  })
})
