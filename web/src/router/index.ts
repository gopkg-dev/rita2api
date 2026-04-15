import ApiDocsView from '../views/ApiDocsView.vue'
import { createRouter, createWebHistory } from 'vue-router'

import GalleryView from '../views/GalleryView.vue'
import HomeView from '../views/HomeView.vue'
import QueueView from '../views/QueueView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/queue', name: 'queue', component: QueueView },
    { path: '/studio', redirect: '/queue' },
    { path: '/history', redirect: '/queue' },
    { path: '/gallery', name: 'gallery', component: GalleryView },
    { path: '/api-docs', name: 'api-docs', component: ApiDocsView },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})
