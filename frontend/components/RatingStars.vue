<script setup lang="ts">
const props = withDefaults(defineProps<{ value: number; count?: number; size?: 'sm' | 'md' }>(), {
  count: undefined,
  size: 'md',
})
const full = computed(() => Math.round(props.value))
</script>

<template>
  <span class="inline-flex items-center gap-1" :class="size === 'sm' ? 'text-xs' : 'text-sm'">
    <span class="text-amber-500">
      <span v-for="i in 5" :key="i">{{ i <= full ? '★' : '☆' }}</span>
    </span>
    <span class="text-slate-500">
      {{ value ? value.toFixed(1) : '—' }}<template v-if="count != null"> ({{ count }})</template>
    </span>
  </span>
</template>
