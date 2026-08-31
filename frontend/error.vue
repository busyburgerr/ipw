<script setup lang="ts">
defineProps<{ error: { statusCode: number; message: string } }>()
</script>

<template>
  <div class="mx-auto flex min-h-screen max-w-md flex-col items-center justify-center gap-4 p-6 text-center">
    <p class="text-5xl font-bold text-brand">{{ error.statusCode }}</p>
    <p class="text-slate-600">{{ error.message || 'Что-то пошло не так' }}</p>
    <button class="btn-primary" @click="clearError({ redirect: '/' })">На главную</button>
  </div>
</template>
