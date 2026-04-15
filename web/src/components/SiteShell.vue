<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import bananaLogo from '../assets/banana.svg'

const links = [
  { label: '首页', to: '/' },
  { label: '任务队列', to: '/queue' },
  { label: '公开画廊', to: '/gallery' },
  { label: 'API接口', to: '/api-docs' },
]

const route = useRoute()
const isHome = computed(() => route.name === 'home')
</script>

<template>
  <div class="shell" :class="{ 'shell--home': isHome }">
    <div class="shell__glow shell__glow--left" aria-hidden="true"></div>
    <div class="shell__glow shell__glow--right" aria-hidden="true"></div>
    <header v-if="!isHome" class="shell__header" :class="{ 'shell__header--home': isHome }">
      <RouterLink class="shell__brand" to="/">
        <span class="shell__brand-mark">
          <img :src="bananaLogo" alt="Nano banana" />
        </span>
        <div>
          <p>Nano banana</p>
          <span>Nano banana workspace</span>
        </div>
      </RouterLink>

      <nav class="shell__nav" :class="{ 'shell__nav--home': isHome }">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="shell__nav-link"
        >
          {{ link.label }}
        </RouterLink>
      </nav>
    </header>

    <main class="shell__content" :class="{ 'shell__content--home': isHome }">
      <slot />
    </main>
  </div>
</template>
