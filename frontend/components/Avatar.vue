<script setup lang="ts">
const props = withDefaults(
  defineProps<{ src?: string | null; name?: string; size?: number }>(),
  { src: null, name: '', size: 44 },
)

const initials = computed(() => {
  const parts = (props.name || 'IPW').trim().split(/\s+/).slice(0, 2)
  return parts.map((p) => p[0]?.toUpperCase() ?? '').join('') || 'IPW'
})
</script>

<template>
  <span
    class="inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full bg-brand font-bold text-white"
    :style="{ width: size + 'px', height: size + 'px', fontSize: Math.round(size * 0.34) + 'px' }"
  >
    <img v-if="src" :src="src" alt="" class="h-full w-full object-cover" />
    <template v-else>{{ initials }}</template>
  </span>
</template>
