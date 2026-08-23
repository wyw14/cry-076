<script setup lang="ts">
import { ref } from 'vue'
import { CheckCircle2, Download, FileCode2, FileText } from '@lucide/vue'
import PageHeader from '../components/PageHeader.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { useWorkspaceStore } from '../stores/workspace'

const store = useWorkspaceStore()
const format = ref<'html' | 'json'>('html')
</script>
<template>
  <section>
    <PageHeader title="导出中心" description="所有文件在本机生成，并校验模板、快照和隐私字段的一致性。" />
    <div class="export-layout"><div class="export-config"><h2>创建导出</h2><div class="option-grid"><button :class="{ selected: format === 'html' }" @click="format = 'html'"><FileText :size="22" /><strong>HTML 打印版</strong><span>适合浏览器打印为 PDF</span></button><button :class="{ selected: format === 'json' }" @click="format = 'json'"><FileCode2 :size="22" /><strong>结构化 JSON</strong><span>用于本地归档和迁移</span></button></div><dl class="export-summary"><div><dt>模板版本</dt><dd>{{ store.selectedTemplate.name }} · v1</dd></div><div><dt>草稿版本</dt><dd>v7</dd></div><div><dt>包含的隐私字段</dt><dd>{{ store.selectedPrivateFields.join('、') }}</dd></div></dl><button class="button primary full"><Download :size="17" />生成本地文件</button></div><aside class="validation-panel"><CheckCircle2 :size="28" /><h2>一致性检查通过</h2><p>预览与导出使用相同的字段顺序，未选中的隐私字段不会写入文件。</p><ul><li>模板版本可用</li><li>草稿快照未变更</li><li>字段清单与隐私策略一致</li><li>文件哈希可校验</li></ul></aside></div>
    <div class="section-heading"><h2>最近导出</h2></div><div class="export-table"><div class="table-head"><span>文件</span><span>版本</span><span>生成时间</span><span>状态</span><span></span></div><div class="table-row"><span>林未-高级后端工程师.html</span><span>草稿 v7</span><span>今天 14:38</span><StatusBadge tone="success" label="校验通过" /><button class="icon-button" title="下载"><Download :size="18" /></button></div></div>
  </section>
</template>
