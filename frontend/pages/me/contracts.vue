<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const { $api } = useNuxtApp()
const { data } = await useAsyncData('my-contracts', () => $api<{ contracts: any[] }>('/me/contracts'))
</script>

<template>
  <div class="flex flex-col gap-6">
    <h1>Контракты</h1>
    <EmptyState v-if="!data?.contracts?.length" title="Контрактов пока нет" />
    <div v-else class="flex flex-col gap-3">
      <NuxtLink
        v-for="c in data.contracts"
        :key="c.id"
        :to="`/contracts/${c.id}`"
        class="card flex items-center justify-between no-underline hover:border-brand"
      >
        <div>
          <p class="font-medium text-slate-900">{{ money(c.agreedAmountCents, c.currency) }}</p>
          <p class="text-sm text-slate-500">{{ c.type === 'fixed' ? 'фикс' : 'почасовой' }} · {{ relTime(c.startedAt) }}</p>
        </div>
        <StatusPill :status="c.status" />
      </NuxtLink>
    </div>
  </div>
</template>
