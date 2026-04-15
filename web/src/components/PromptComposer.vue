<script setup lang="ts">
import { computed } from 'vue'

import { ratioOptions, resolutionOptions } from '../lib/generation-options'
import { useGenerationStore } from '../stores/generation'
import { useToastStore } from '../stores/toast'

const generationStore = useGenerationStore()
const toastStore = useToastStore()
const imageNumOptions = computed(() => [1, 2, 3, 4])

async function handleSubmit() {
  try {
    await generationStore.submit()
    toastStore.push({
      title: '任务已提交',
      description: '生成请求已经进入队列，可以去任务队列页继续跟踪结果。',
      tone: 'success',
    })
  } catch (error) {
    toastStore.push({
      title: '提交失败',
      description: error instanceof Error ? error.message : '表单提交失败，请稍后重试。',
      tone: 'error',
    }, 3600)
  }
}
</script>

<template>
  <div class="panel panel--dense composer-panel">
    <div class="composer-panel__head composer-panel__head--stacked">
      <div>
        <p class="eyebrow">Control panel</p>
      </div>
    </div>

    <label class="field">
      <textarea
        :value="generationStore.draft.prompt"
        rows="7"
        placeholder="例如：matte aluminum chair, soft side lighting, clean studio background, minimal product photography"
        @input="
          generationStore.updateDraft({
            prompt: ($event.target as HTMLTextAreaElement).value,
          })
        "
      />
    </label>

    <div class="field">
      <span>画幅比例</span>
      <div class="option-pill-group">
        <button
          v-for="option in ratioOptions"
          :key="option"
          type="button"
          class="option-pill"
          :class="{ 'option-pill--active': generationStore.draft.ratio === option }"
          @click="generationStore.updateDraft({ ratio: option })"
        >
          {{ option }}
        </button>
      </div>
      <small class="field__hint">横图适合场景，竖图适合人物，方图适合封面。</small>
    </div>

    <div class="field">
      <span>清晰度</span>
      <div class="option-pill-group">
        <button
          v-for="option in resolutionOptions"
          :key="option"
          type="button"
          class="option-pill"
          :class="{ 'option-pill--active': generationStore.draft.resolution === option }"
          @click="generationStore.updateDraft({ resolution: option })"
        >
          {{ option }}
        </button>
      </div>
      <small class="field__hint">清晰度越高，耗时通常越长，首轮建议先用 1K 或 2K 试词。</small>
    </div>

    <div class="field">
      <span>出图数量</span>
      <div class="option-pill-group">
        <button
          v-for="option in imageNumOptions"
          :key="option"
          type="button"
          class="option-pill"
          :class="{ 'option-pill--active': generationStore.draft.imageNum === option }"
          @click="generationStore.updateDraft({ imageNum: option })"
        >
          {{ option }}
        </button>
      </div>
      <small class="field__hint">多张适合找方向，单张适合快速试词。</small>
    </div>

    <div class="composer-actions composer-actions--single">
      <button class="button button--primary button--wide" :disabled="generationStore.isSubmitting" @click="handleSubmit">
        {{ generationStore.isSubmitting ? '正在提交...' : '立即生成' }}
      </button>
    </div>
  </div>
</template>
