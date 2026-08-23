<script setup lang="ts">
import { computed, ref } from 'vue'
import { Archive, GitCompare, RotateCcw } from '@lucide/vue'
import PageHeader from '../components/PageHeader.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { useWorkspaceStore } from '../stores/workspace'

const store = useWorkspaceStore()
const selected = ref<string[]>(['s3', 's2'])
const canCompare = computed(() => selected.value.length === 2)
function toggle(id: string) { selected.value = selected.value.includes(id) ? selected.value.filter((item) => item !== id) : [...selected.value.slice(-1), id] }
</script>
<template>
  <section>
    <PageHeader title="草稿历史" description="比较、恢复或归档任意版本，恢复前会自动保留当前快照。"><button class="button primary" :disabled="!canCompare"><GitCompare :size="17" />比较所选版本</button></PageHeader>
    <div class="history-list"><article v-for="snapshot in store.snapshots" :key="snapshot.id" class="history-row"><label class="check-control"><input type="checkbox" :checked="selected.includes(snapshot.id)" @change="toggle(snapshot.id)" /><span></span></label><div class="version-mark">v{{ snapshot.version }}</div><div class="history-main"><div><h2>{{ snapshot.reason }}</h2><StatusBadge v-if="snapshot.id === 's3'" tone="success" label="当前基线" /></div><p>{{ snapshot.createdAt }} · {{ snapshot.fields }} 个字段</p></div><div class="row-actions"><button class="icon-button" title="恢复版本"><RotateCcw :size="18" /></button><button class="icon-button" title="归档版本"><Archive :size="18" /></button></div></article></div>
    <div class="comparison-band"><div><span class="eyebrow">版本差异</span><h2>v5 → v7</h2></div><div class="diff-summary"><span class="added">+ 3 个字段</span><span class="changed">~ 2 处修改</span><span class="removed">- 1 个字段</span></div></div>
  </section>
</template>
