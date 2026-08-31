<script setup lang="ts">
definePageMeta({ middleware: 'admin' })
const { $api } = useNuxtApp()
const { data, refresh } = await useAsyncData('admin-disputes', () =>
  $api<{ disputes: any[] }>('/admin/disputes'),
)
const { loading, error, run } = useSubmit()

async function claim(id: string) {
  await run(() => $api(`/admin/disputes/${id}/claim`, { method: 'POST' }))
  await refresh()
}
async function resolve(id: string, outcome: 'client' | 'freelancer') {
  const note = prompt('Комментарий к решению:') ?? ''
  await run(() => $api(`/admin/disputes/${id}/resolve`, { method: 'POST', body: { outcome, note } }))
  await refresh()
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <h1>Арбитраж — открытые споры</h1>
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
    <EmptyState v-if="!data?.disputes?.length" title="Открытых споров нет" />
    <div v-for="d in data?.disputes || []" :key="d.id" class="card flex flex-col gap-2">
      <div class="flex items-center justify-between">
        <NuxtLink :to="`/contracts/${d.contractId}`" class="text-sm">Контракт →</NuxtLink>
        <StatusPill :status="d.status" />
      </div>
      <p class="text-sm text-slate-700">{{ d.reason }}</p>
      <p class="text-xs text-slate-400">
        этап: {{ d.milestoneId || 'весь контракт' }} · {{ relTime(d.createdAt) }}
      </p>
      <div class="flex flex-wrap gap-2 pt-1">
        <button v-if="d.status === 'open'" class="btn-ghost btn-sm" :disabled="loading" @click="claim(d.id)">
          Взять в работу
        </button>
        <button class="btn-primary btn-sm" :disabled="loading" @click="resolve(d.id, 'freelancer')">
          В пользу фрилансера
        </button>
        <button class="btn-danger btn-sm" :disabled="loading" @click="resolve(d.id, 'client')">
          В пользу заказчика
        </button>
      </div>
    </div>
  </div>
</template>
