<script setup lang="ts">
defineProps<{ mode: 'login' | 'register' }>()
</script>

<template>
  <div class="flex min-h-[70vh] items-center justify-center">
    <div class="card w-full max-w-sm">
      <div class="mb-5 flex items-baseline gap-3">
        <NuxtLink
          to="/auth/login"
          class="text-lg font-bold no-underline"
          :class="mode === 'login' ? 'text-ink' : 'text-slate-400'"
        >Авторизация</NuxtLink>
        <NuxtLink
          to="/auth/register"
          class="text-sm no-underline"
          :class="mode === 'register' ? 'font-bold text-ink' : 'text-slate-400'"
        >Регистрация</NuxtLink>
      </div>
      <slot />
    </div>
  </div>
</template>
