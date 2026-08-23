<script setup lang="ts">
import { computed, ref } from 'vue'
import { AlertCircle, ChevronLeft, ChevronRight, Save } from '@lucide/vue'
import PageHeader from '../components/PageHeader.vue'
import { useWorkspaceStore } from '../stores/workspace'

const store = useWorkspaceStore()
const steps = ['求职目标', '基本资料', '工作经历', '技能与项目', '附件']
const current = ref(1)
const summary = computed({ get: () => store.draft.summary, set: (value) => store.updateField('summary', value) })
</script>

<template>
  <section>
    <PageHeader title="填写向导" description="内容独立于模板保存，切换版式时不会丢失已填写字段。"><button class="button secondary"><Save :size="17" />保存快照</button></PageHeader>
    <div class="wizard-layout">
      <ol class="step-list"><li v-for="(step, index) in steps" :key="step" :class="{ active: current === index, done: current > index }"><button @click="current = index"><span>{{ index + 1 }}</span><div><strong>{{ step }}</strong><small>{{ current > index ? '已填写' : index === current ? '正在编辑' : '待填写' }}</small></div></button></li></ol>
      <form class="editor-panel" @submit.prevent>
        <div class="panel-heading"><div><span class="eyebrow">第 {{ current + 1 }} 步</span><h2>{{ steps[current] }}</h2></div><span class="autosave">自动保存于 14:42</span></div>
        <div v-if="current === 1" class="form-grid"><label><span>姓名</span><input :value="store.draft.full_name" @input="store.updateField('full_name', ($event.target as HTMLInputElement).value)" /></label><label><span>邮箱</span><input :value="store.draft.email" @input="store.updateField('email', ($event.target as HTMLInputElement).value)" /></label><label><span>联系电话</span><input :value="store.draft.phone" @input="store.updateField('phone', ($event.target as HTMLInputElement).value)" /></label><label class="wide"><span>个人概述</span><textarea v-model="summary" rows="5"></textarea><small>{{ summary.length }} / 300</small></label></div>
        <div v-else class="empty-editor"><strong>{{ steps[current] }}内容已保留</strong><p>当前模板会按字段规范展示这一部分。</p></div>
        <div class="notice warning"><AlertCircle :size="19" /><div><strong>有 2 个字段尚未映射</strong><p>“开源社区贡献”和“期望工作城市”会保留在草稿中，但不会出现在当前模板预览。</p></div></div>
        <footer class="form-actions"><button class="button secondary" :disabled="current === 0" @click="current--"><ChevronLeft :size="17" />上一步</button><button class="button primary" :disabled="current === steps.length - 1" @click="current++">下一步<ChevronRight :size="17" /></button></footer>
      </form>
    </div>
  </section>
</template>
