<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
const auth = useAuthStore()
const route = useRoute()

async function logout() {
  await auth.logout()
  await navigateTo('/')
}

const nav = computed(() => {
  const items: { to: string; label: string }[] = [{ to: '/projects', label: 'Проекты' }]
  if (!auth.isAuthed) return items
  items.push({ to: '/dashboard', label: 'Кабинет' })
  if (auth.user?.isClient) items.push({ to: '/projects/new', label: 'Разместить проект' })
  if (auth.user?.isFreelancer) items.push({ to: '/me/wallet', label: 'Кошелёк' })
  if (auth.user?.isAdmin) items.push({ to: '/admin/disputes', label: 'Арбитраж' })
  return items
})

const isActive = (to: string) => route.path === to || route.path.startsWith(to + '/')
</script>

<template>
  <div class="min-h-screen">
    <header class="mx-auto flex h-16 max-w-5xl items-center gap-8 px-4">
      <NuxtLink to="/" class="text-xl font-extrabold tracking-wide text-white no-underline">IPW</NuxtLink>
      <nav class="hidden items-center gap-6 text-sm sm:flex">
        <NuxtLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="no-underline transition"
          :class="isActive(item.to) ? 'font-semibold text-white' : 'text-white/70 hover:text-white'"
        >
          {{ item.label }}
        </NuxtLink>
      </nav>
      <div class="ml-auto flex items-center gap-3 text-sm">
        <template v-if="auth.isAuthed">
          <NuxtLink to="/profile/edit" class="text-white/80 no-underline hover:text-white">
            {{ auth.user?.displayName || auth.user?.email }}
          </NuxtLink>
          <button class="btn-ghost btn-sm" @click="logout">Выйти</button>
        </template>
        <template v-else>
          <NuxtLink to="/auth/login" class="text-white/80 no-underline hover:text-white">Войти</NuxtLink>
          <NuxtLink to="/auth/register" class="btn-ghost btn-sm no-underline">Регистрация</NuxtLink>
        </template>
      </div>
    </header>

    <main class="mx-auto max-w-5xl px-4 pb-16 pt-4">
      <slot />
    </main>
  </div>
</template>
