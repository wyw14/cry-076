import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import TemplateLibrary from './views/TemplateLibrary.vue'
import DraftWizard from './views/DraftWizard.vue'
import LivePreview from './views/LivePreview.vue'
import DraftHistory from './views/DraftHistory.vue'
import ExportCenter from './views/ExportCenter.vue'
import PrivacySettings from './views/PrivacySettings.vue'
import './styles.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/templates' },
    { path: '/templates', component: TemplateLibrary },
    { path: '/draft', component: DraftWizard },
    { path: '/preview', component: LivePreview },
    { path: '/history', component: DraftHistory },
    { path: '/exports', component: ExportCenter },
    { path: '/privacy', component: PrivacySettings }
  ]
})

createApp(App).use(createPinia()).use(router).mount('#app')
