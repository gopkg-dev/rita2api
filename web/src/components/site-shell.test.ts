// @vitest-environment jsdom

import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import SiteShell from './SiteShell.vue'

async function mountSiteShell() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div />' } },
      { path: '/queue', name: 'queue', component: { template: '<div />' } },
      { path: '/gallery', name: 'gallery', component: { template: '<div />' } },
    ],
  })

  await router.push('/queue')
  await router.isReady()

  return mount(SiteShell, {
    global: {
      plugins: [router],
    },
    slots: {
      default: '<div>content</div>',
    },
  })
}

describe('SiteShell', () => {
  it('renders the brand logo from banana.svg', async () => {
    const wrapper = await mountSiteShell()

    const logo = wrapper.get('.shell__brand-mark img')
    expect(logo.attributes('src')).toContain('/src/assets/banana.svg')
    expect(logo.attributes('alt')).toBe('Nano banana')
  })
})
