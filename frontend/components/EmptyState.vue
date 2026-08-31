<script setup lang="ts">
defineProps<{ title: string; hint?: string }>()
</script>

<template>
  <div class="card flex flex-col items-center gap-2 py-12 text-center">
    <p class="font-medium text-slate-700">{{ title }}</p>
    <p v-if="hint" class="max-w-sm text-sm text-slate-500">{{ hint }}</p>
    <slot />
  </div>
</template>
