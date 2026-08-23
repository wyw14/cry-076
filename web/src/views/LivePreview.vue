<script setup lang="ts">
import { computed, ref } from 'vue'
import { Maximize2, Printer, ZoomIn, ZoomOut } from '@lucide/vue'
import PageHeader from '../components/PageHeader.vue'
import { useWorkspaceStore } from '../stores/workspace'

const store = useWorkspaceStore()
const zoom = ref(85)
const phone = computed(() => store.privacy.phone ? store.draft.phone : '')
</script>
<template>
  <section class="preview-page">
    <PageHeader title="实时预览" description="预览与本地导出使用同一模板版本和隐私选择。"><button class="button secondary"><Printer :size="17" />打印预览</button></PageHeader>
    <div class="preview-toolbar"><span>{{ store.selectedTemplate.name }}</span><div><button class="icon-button" title="缩小" @click="zoom = Math.max(60, zoom - 5)"><ZoomOut :size="18" /></button><output>{{ zoom }}%</output><button class="icon-button" title="放大" @click="zoom = Math.min(120, zoom + 5)"><ZoomIn :size="18" /></button><button class="icon-button" title="适应窗口"><Maximize2 :size="18" /></button></div></div>
    <div class="paper-stage"><article class="resume-paper" :style="{ transform: `scale(${zoom / 100})` }"><header><div><h2>{{ store.draft.full_name }}</h2><p>高级后端工程师</p></div><div class="contact"><span>{{ store.draft.email }}</span><span v-if="phone">{{ phone }}</span></div></header><section><h3>个人概述</h3><p>{{ store.draft.summary }}</p></section><section><h3>工作亮点</h3><ul><li v-for="experience in store.draft.experiences" :key="experience">{{ experience }}</li></ul></section><section><h3>技能</h3><div class="skill-row"><span v-for="skill in store.draft.skills" :key="skill">{{ skill }}</span></div></section></article></div>
  </section>
</template>
