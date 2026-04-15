import { defineStore } from 'pinia'
import { ref } from 'vue'

import { fetchBootstrap } from '../lib/api'

export const useSessionStore = defineStore('session', () => {
  const bootstrapped = ref(false)
  const sessionToken = ref('')
  const defaultRatio = ref('1:1')
  const defaultResolution = ref('1K')
  const defaultImageNum = ref(1)
  const galleryEnabled = ref(true)

  async function bootstrap() {
    if (bootstrapped.value) {
      return
    }

    const payload = await fetchBootstrap()
    sessionToken.value = payload.session.token
    defaultRatio.value = payload.defaults.ratio
    defaultResolution.value = payload.defaults.resolution
    defaultImageNum.value = payload.defaults.imageNum
    galleryEnabled.value = payload.gallery.enabled
    bootstrapped.value = true
  }

  return {
    bootstrapped,
    sessionToken,
    defaultRatio,
    defaultResolution,
    defaultImageNum,
    galleryEnabled,
    bootstrap,
  }
})
