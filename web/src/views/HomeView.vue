<script setup lang="ts">
import { computed, onMounted } from 'vue'

import bananaLogo from '../assets/banana.svg'
import PromptComposer from '../components/PromptComposer.vue'
import SiteShell from '../components/SiteShell.vue'
import { useGenerationStore } from '../stores/generation'

const generationStore = useGenerationStore()
const queuedCount = computed(
  () => generationStore.tasks.filter((task) => task.status === 'queued').length,
)
const runningCount = computed(
  () => generationStore.tasks.filter((task) => task.status === 'running').length,
)
const completedCount = computed(
  () => generationStore.tasks.filter((task) => task.status === 'succeeded').length,
)

onMounted(async () => {
  await generationStore.initialize()
})
</script>

<template>
  <SiteShell>
    <section class="home-center">
      <section class="home-hero">
        <p class="pill">🍌 Nano banana</p>
        <p class="home-hero__copy">从一句提示词开始，快速生成高质量图像。</p>
      </section>

      <div class="home-stage">
        <div class="home-stage__overlay">
          <aside class="home-floating-stats" aria-label="首页任务状态">
            <article class="home-logo-card" aria-label="品牌标记">
              <img :src="bananaLogo" alt="Nano banana" />
            </article>
            <article class="home-stat-card">
              <span>总任务</span>
              <strong>{{ generationStore.tasks.length }}</strong>
            </article>
            <article class="home-stat-card">
              <span>排队中</span>
              <strong>{{ queuedCount }}</strong>
            </article>
            <article class="home-stat-card">
              <span>生成中</span>
              <strong>{{ runningCount }}</strong>
            </article>
            <article class="home-stat-card">
              <span>已完成</span>
              <strong>{{ completedCount }}</strong>
            </article>

            <nav class="home-side-links" aria-label="首页快捷入口">
              <RouterLink class="home-side-link-card" to="/queue">
                <span>任务队列</span>
                <strong>查看任务</strong>
              </RouterLink>
              <RouterLink class="home-side-link-card" to="/gallery">
                <span>公开画廊</span>
                <strong>浏览作品</strong>
              </RouterLink>
              <RouterLink class="home-side-link-card" to="/api-docs">
                <span>API接口</span>
                <strong>查看文档</strong>
              </RouterLink>
            </nav>
          </aside>
        </div>

        <div class="home-stage__panel">
          <PromptComposer />
        </div>
      </div>
    </section>
  </SiteShell>
</template>
