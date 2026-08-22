<script setup lang="ts">
import { FileLock2, ShieldCheck } from '@lucide/vue'
import PageHeader from '../components/PageHeader.vue'
import { useWorkspaceStore } from '../stores/workspace'
const store = useWorkspaceStore()
const fields = [{ key: 'full_name', label: '姓名', detail: '导出文件的署名' },{ key: 'email', label: '邮箱', detail: '用于招聘方联系' },{ key: 'phone', label: '联系电话', detail: '默认仅在预览中遮罩' },{ key: 'address', label: '常住地址', detail: '精确地址属于敏感信息' },{ key: 'attachments', label: '受控附件', detail: '证书和作品原文件' }]
</script>
<template>
  <section>
    <PageHeader title="隐私设置" description="控制预览、导出和受控附件访问，所有访问都会写入审计时间线。"><button class="button primary"><ShieldCheck :size="17" />保存设置</button></PageHeader>
    <div class="settings-layout"><div class="settings-list"><div v-for="field in fields" :key="field.key" class="setting-row"><div><strong>{{ field.label }}</strong><p>{{ field.detail }}</p></div><label class="switch"><input type="checkbox" :checked="store.privacy[field.key]" @change="store.togglePrivacy(field.key)" /><span></span></label></div></div><aside class="audit-aside"><FileLock2 :size="25" /><h2>访问审计</h2><p>最近 7 天有 4 次预览、2 次导出，没有附件外部访问。</p><button class="button secondary full">查看审计记录</button></aside></div>
  </section>
</template>
