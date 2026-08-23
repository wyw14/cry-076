<script setup lang="ts">
import { computed, ref } from 'vue'
import { Check, Search, SlidersHorizontal } from '@lucide/vue'
import PageHeader from '../components/PageHeader.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { useWorkspaceStore } from '../stores/workspace'

const store = useWorkspaceStore()
const search = ref('')
const scenario = ref('全部场景')
const scenarios = ['全部场景', '社招', '技术', '校招']
const visibleTemplates = computed(() => store.templates.filter((item) => {
  const matchesText = item.name.includes(search.value) || item.category.includes(search.value)
  const matchesScenario = scenario.value === '全部场景' || item.scenarios.includes(scenario.value)
  return matchesText && matchesScenario
}))
</script>

<template>
  <section>
    <PageHeader title="模板库" description="选择与当前求职目标匹配的版式和字段规范。"><button class="button secondary"><SlidersHorizontal :size="17" />管理分类</button></PageHeader>
    <div class="filter-bar"><label class="search-control"><Search :size="17" /><input v-model="search" placeholder="搜索模板" /></label><div class="segmented"><button v-for="item in scenarios" :key="item" :class="{ active: scenario === item }" @click="scenario = item">{{ item }}</button></div></div>
    <div class="template-grid">
      <article v-for="template in visibleTemplates" :key="template.id" class="template-card" :class="{ selected: template.active }">
        <div class="template-preview" :style="{ '--accent': template.accent }"><div class="preview-title"></div><div class="preview-columns"><div></div><div></div></div><span>{{ template.name.slice(0, 1) }}</span></div>
        <div class="template-meta"><div><h2>{{ template.name }}</h2><p>{{ template.category }} · {{ template.scenarios.join(' / ') }}</p></div><StatusBadge v-if="template.active" tone="success" label="正在使用" /></div>
        <button class="button full" :class="template.active ? 'secondary' : 'primary'" @click="store.selectTemplate(template.id)"><Check v-if="template.active" :size="17" />{{ template.active ? '已选择' : '应用模板' }}</button>
      </article>
    </div>
  </section>
</template>
