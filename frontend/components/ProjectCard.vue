<script setup lang="ts">
interface Project {
  id: string
  title: string
  description: string
  budgetType: string
  fixedAmountCents: number | null
  hourlyRateMinCents: number | null
  hourlyRateMaxCents: number | null
  currency: string
  experienceLevel: string
  status: string
  proposalsCount: number
  publishedAt: string | null
  createdAt: string
}
const props = defineProps<{ project: Project }>()

const budget = computed(() => {
  const p = props.project
  if (p.budgetType === 'fixed') return money(p.fixedAmountCents, p.currency)
  return `${money(p.hourlyRateMinCents, p.currency)} – ${money(p.hourlyRateMaxCents, p.currency)} / час`
})
</script>

<template>
  <NuxtLink :to="`/projects/${project.id}`" class="card block transition hover:border-brand hover:shadow">
    <div class="flex items-start justify-between gap-3">
      <h2 class="text-slate-900">{{ project.title }}</h2>
      <StatusPill :status="project.status" />
    </div>
    <p class="mt-1 line-clamp-2 text-sm text-slate-600">{{ project.description }}</p>
    <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-slate-500">
      <span class="font-medium text-slate-800">{{ budget }}</span>
      <span>{{ project.proposalsCount }} откликов</span>
      <span>{{ relTime(project.publishedAt || project.createdAt) }}</span>
    </div>
  </NuxtLink>
</template>
