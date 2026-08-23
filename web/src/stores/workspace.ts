import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useStorage } from '@vueuse/core'

export type TemplateCard = { id: string; name: string; category: string; scenarios: string[]; accent: string; active: boolean }
export type Snapshot = { id: string; version: number; reason: string; createdAt: string; fields: number }

export const useWorkspaceStore = defineStore('workspace', () => {
  const templates = ref<TemplateCard[]>([
    { id: 'modern', name: '清晰双栏', category: '通用', scenarios: ['社招', '技术'], accent: '#0f766e', active: true },
    { id: 'focused', name: '技术聚焦', category: '工程', scenarios: ['技术'], accent: '#2563eb', active: false },
    { id: 'compact', name: '校园精简', category: '校招', scenarios: ['校招'], accent: '#b45309', active: false }
  ])
  const draft = useStorage('cry076.resume-draft.v1', {
    full_name: '林未',
    email: 'lin@example.test',
    phone: '138 0000 0000',
    summary: '专注可靠服务与工程效率。',
    experiences: ['平台服务重构', '可观测性治理'],
    skills: ['Go', 'PostgreSQL', 'Kubernetes']
  }, window.localStorage, { mergeDefaults: true })
  const snapshots = ref<Snapshot[]>([
    { id: 's3', version: 7, reason: '切换到技术模板前', createdAt: '今天 14:36', fields: 18 },
    { id: 's2', version: 5, reason: '补充项目经历', createdAt: '昨天 20:18', fields: 16 },
    { id: 's1', version: 2, reason: '首次完整填写', createdAt: '8 月 20 日', fields: 12 }
  ])
  const privacy = useStorage<Record<string, boolean>>('cry076.export-privacy.v1', {
    full_name: true,
    email: true,
    phone: false,
    address: false,
    attachments: false
  }, window.localStorage, { mergeDefaults: true })
  const selectedTemplate = computed(() => templates.value.find((item) => item.active)!)
  const selectedPrivateFields = computed(() => Object.entries(privacy.value).filter(([, included]) => included).map(([field]) => field))
  function selectTemplate(id: string) { templates.value = templates.value.map((item) => ({ ...item, active: item.id === id })) }
  function updateField(field: keyof typeof draft.value, value: string | string[]) { draft.value = { ...draft.value, [field]: value } }
  function togglePrivacy(field: string) { privacy.value[field] = !privacy.value[field] }
  return { templates, draft, snapshots, privacy, selectedTemplate, selectedPrivateFields, selectTemplate, updateField, togglePrivacy }
})
