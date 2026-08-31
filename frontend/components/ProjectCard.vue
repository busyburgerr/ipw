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
  <NuxtLink
    :to="`/projects/${project.id}`"
    class="card block no-underline transition hover:-translate-y-0.5 hover:shadow-lg"
  >
    <div class="flex items-start gap-3">
      <Avatar :name="project.title" :size="44" />
      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-3">
          <h2 class="truncate text-base text-ink">{{ project.title }}</h2>
          <StatusPill :status="project.status" />
        </div>
        <p class="mt-1 line-clamp-2 text-sm text-slate-500">{{ project.description }}</p>
      </div>
    </div>
    <div class="mt-4 flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-slate-100 pt-3 text-sm">
      <span class="font-semibold text-navy">{{ budget }}</span>
      <span class="text-slate-400">{{ project.proposalsCount }} откликов</span>
      <span class="text-slate-400">{{ relTime(project.publishedAt || project.createdAt) }}</span>
    </div>
  </NuxtLink>
</template>
