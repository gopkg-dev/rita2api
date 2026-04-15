<script setup lang="ts">
import { onMounted } from 'vue'

import SiteShell from '../components/SiteShell.vue'
import TaskCard from '../components/TaskCard.vue'
import { useGalleryStore } from '../stores/gallery'
import { useGenerationStore } from '../stores/generation'

const galleryStore = useGalleryStore()
const generationStore = useGenerationStore()

onMounted(async () => {
  await galleryStore.load()
})
</script>

<template>
  <SiteShell>
    <article v-if="galleryStore.errorMessage" class="panel gallery-state">
      <p class="eyebrow">Load status</p>
      <h2>画廊暂时没有加载出来</h2>
      <p class="panel__lead">{{ galleryStore.errorMessage }}</p>
    </article>

    <article v-else-if="galleryStore.isLoading" class="panel gallery-state">
      <p class="eyebrow">Loading</p>
      <h2>正在整理公开作品</h2>
      <p class="panel__lead">画廊内容加载后会自动出现在下面。</p>
    </article>

    <section class="gallery-grid gallery-grid--masonry">
      <TaskCard
        v-for="task in galleryStore.items"
        :key="task.id"
        :task="task"
        compact
        @retry="generationStore.retry"
      />
    </section>
  </SiteShell>
</template>
