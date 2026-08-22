<script setup lang="ts">
import { FileText, LayoutTemplate, History, ShieldCheck, Download, Eye, Bell } from '@lucide/vue'
import { RouterLink, RouterView } from 'vue-router'

const navigation = [
  { to: '/templates', label: '模板库', icon: LayoutTemplate },
  { to: '/draft', label: '填写向导', icon: FileText },
  { to: '/preview', label: '实时预览', icon: Eye },
  { to: '/history', label: '草稿历史', icon: History },
  { to: '/exports', label: '导出中心', icon: Download },
  { to: '/privacy', label: '隐私设置', icon: ShieldCheck }
]
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand-block"><div class="brand-mark">履</div><div><strong>履历工坊</strong><span>求职材料工作台</span></div></div>
      <nav aria-label="主导航">
        <RouterLink v-for="item in navigation" :key="item.to" :to="item.to"><component :is="item.icon" :size="18" /><span>{{ item.label }}</span></RouterLink>
      </nav>
      <div class="workspace-status"><span class="status-dot"></span><div><strong>草稿已同步</strong><small>刚刚保存</small></div></div>
    </aside>
    <div class="workspace">
      <header class="topbar"><div><span class="eyebrow">目标方案</span><strong>高级后端工程师 · 社招</strong></div><button class="icon-button" title="通知"><Bell :size="19" /></button><div class="avatar">林</div></header>
      <main><RouterView /></main>
    </div>
  </div>
</template>
