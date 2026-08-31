<script setup lang="ts">
const props = defineProps<{ status: string }>()
</script>

<template>
  <span class="pill" :class="statusColor(props.status)">{{ statusRu(props.status) }}</span>
</template>
