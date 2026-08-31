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
    <header class="border-b border-slate-200 bg-white">
      <div class="mx-auto flex h-14 max-w-5xl items-center gap-6 px-4">
        <NuxtLink to="/" class="font-bold text-ink no-underline">фриланс-биржа</NuxtLink>
        <nav class="flex items-center gap-4 text-sm">
          <NuxtLink
            v-for="item in nav"
            :key="item.to"
            :to="item.to"
            class="no-underline"
            :class="isActive(item.to) ? 'font-semibold text-brand-700' : 'text-slate-600 hover:text-slate-900'"
          >
            {{ item.label }}
          </NuxtLink>
        </nav>
        <div class="ml-auto flex items-center gap-3 text-sm">
          <template v-if="auth.isAuthed">
            <NuxtLink to="/dashboard" class="text-slate-600 no-underline hover:text-slate-900">
              {{ auth.user?.displayName || auth.user?.email }}
            </NuxtLink>
            <button class="btn-ghost btn-sm" @click="logout">Выйти</button>
          </template>
          <template v-else>
            <NuxtLink to="/auth/login" class="text-slate-600 no-underline hover:text-slate-900">Войти</NuxtLink>
            <NuxtLink to="/auth/register" class="btn-primary btn-sm no-underline">Регистрация</NuxtLink>
          </template>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-5xl px-4 py-8">
      <slot />
    </main>
  </div>
</template>
