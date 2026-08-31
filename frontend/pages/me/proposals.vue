<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const { $api } = useNuxtApp()
const { data, refresh } = await useAsyncData('my-proposals', () =>
  $api<{ proposals: any[] }>('/me/proposals'),
)
const { error, run } = useSubmit()

async function withdraw(id: string) {
  await run(() => $api(`/proposals/${id}/withdraw`, { method: 'POST' }))
  await refresh()
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <h1>Мои отклики</h1>
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
    <EmptyState v-if="!data?.proposals?.length" title="Вы ещё не откликались">
      <NuxtLink to="/projects" class="btn-ghost btn-sm no-underline">К проектам</NuxtLink>
    </EmptyState>
    <div v-else class="flex flex-col gap-3">
      <div v-for="p in data.proposals" :key="p.id" class="card flex flex-col gap-2">
        <div class="flex items-center justify-between">
          <NuxtLink :to="`/projects/${p.projectId}`" class="text-sm">Проект →</NuxtLink>
          <StatusPill :status="p.status" />
        </div>
        <p class="text-sm"><b>{{ money(p.bidAmountCents) }}</b>
          <span v-if="p.estimatedDays" class="text-slate-500"> · ~{{ p.estimatedDays }} дн.</span>
        </p>
        <p class="line-clamp-2 text-sm text-slate-600">{{ p.coverLetter }}</p>
        <button
          v-if="['pending', 'shortlisted'].includes(p.status)"
          class="btn-ghost btn-sm w-fit"
          @click="withdraw(p.id)"
        >Отозвать</button>
      </div>
    </div>
  </div>
</template>
