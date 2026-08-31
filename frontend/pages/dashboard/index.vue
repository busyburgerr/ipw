<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
import { useAuthStore } from '~/stores/auth'
const auth = useAuthStore()
const { $api } = useNuxtApp()

const { data: contracts } = await useAsyncData('dash-contracts', () =>
  $api<{ contracts: any[] }>('/me/contracts'),
)
const { data: wallet } = await useAsyncData(
  'dash-wallet',
  () => $api<any>('/me/wallet').catch(() => null),
)
</script>

<template>
  <div class="flex flex-col gap-6">
    <h1>Кабинет</h1>
    <p class="text-slate-600">
      {{ auth.user?.displayName }} ·
      <span v-if="auth.user?.isFreelancer">фрилансер</span>
      <span v-if="auth.user?.isFreelancer && auth.user?.isClient"> и </span>
      <span v-if="auth.user?.isClient">заказчик</span>
    </p>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <NuxtLink v-if="auth.user?.isClient" to="/projects/new" class="card no-underline hover:border-brand">
        <p class="font-semibold text-slate-900">＋ Разместить проект</p>
        <p class="mt-1 text-sm text-slate-500">Опишите задачу и соберите отклики</p>
      </NuxtLink>
      <NuxtLink v-if="auth.user?.isClient" to="/me/projects" class="card no-underline hover:border-brand">
        <p class="font-semibold text-slate-900">Мои проекты</p>
        <p class="mt-1 text-sm text-slate-500">Черновики, публикации, отклики</p>
      </NuxtLink>
      <NuxtLink v-if="auth.user?.isFreelancer" to="/me/proposals" class="card no-underline hover:border-brand">
        <p class="font-semibold text-slate-900">Мои отклики</p>
        <p class="mt-1 text-sm text-slate-500">Статусы отправленных предложений</p>
      </NuxtLink>
      <NuxtLink to="/me/contracts" class="card no-underline hover:border-brand">
        <p class="font-semibold text-slate-900">Контракты</p>
        <p class="mt-1 text-sm text-slate-500">{{ contracts?.contracts?.length || 0 }} всего</p>
      </NuxtLink>
      <NuxtLink v-if="auth.user?.isFreelancer" to="/me/wallet" class="card no-underline hover:border-brand">
        <p class="font-semibold text-slate-900">Кошелёк</p>
        <p class="mt-1 text-sm text-slate-500">
          {{ wallet ? money(wallet.availableCents, wallet.currency) : '—' }} доступно
        </p>
      </NuxtLink>
      <NuxtLink to="/profile/edit" class="card no-underline hover:border-brand">
        <p class="font-semibold text-slate-900">Профиль</p>
        <p class="mt-1 text-sm text-slate-500">Публичная информация, навыки, портфолио</p>
      </NuxtLink>
    </div>

    <div v-if="contracts?.contracts?.length" class="flex flex-col gap-3">
      <h2>Активные контракты</h2>
      <NuxtLink
        v-for="c in contracts.contracts.filter((x) => ['active', 'disputed'].includes(x.status))"
        :key="c.id"
        :to="`/contracts/${c.id}`"
        class="card flex items-center justify-between no-underline hover:border-brand"
      >
        <span class="text-sm">{{ money(c.agreedAmountCents, c.currency) }} · {{ c.type }}</span>
        <StatusPill :status="c.status" />
      </NuxtLink>
    </div>
  </div>
</template>
