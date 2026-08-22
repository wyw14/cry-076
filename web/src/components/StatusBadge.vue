<script setup lang="ts">defineProps<{ tone?: 'success' | 'warning' | 'neutral'; label: string }>()</script>
<template><span class="status-badge" :class="`status-${tone ?? 'neutral'}`">{{ label }}</span></template>
