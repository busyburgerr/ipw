<script setup lang="ts">
const { $apiRaw } = useNuxtApp()
const route = useRoute()
const router = useRouter()

const q = ref((route.query.q as string) || '')
const category = ref((route.query.category as string) || '')
const budgetType = ref((route.query.budgetType as string) || '')

interface Category { id: string; slug: string; name: string; children: Category[] }
const { data: cats } = await useAsyncData('categories', () =>
  $apiRaw<{ categories: Category[] }>('/catalog/categories'),
)

const query = computed(() => {
  const p = new URLSearchParams()
  if (q.value) p.set('q', q.value)
  if (category.value) p.set('category', category.value)
  if (budgetType.value) p.set('budgetType', budgetType.value)
  return p.toString()
})

const { data, pending } = await useAsyncData(
  'projects',
  () => $apiRaw<{ projects: any[]; total: number }>(`/projects${query.value ? '?' + query.value : ''}`),
  { watch: [query] },
)

let t: ReturnType<typeof setTimeout>
watch(query, () => {
  clearTimeout(t)
  t = setTimeout(
    () => router.replace({ query: Object.fromEntries(new URLSearchParams(query.value)) }),
    300,
  )
})
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1>Проекты</h1>
      <span class="text-sm text-slate-500">{{ data?.total ?? 0 }} открытых</span>
    </div>

    <div class="flex flex-wrap gap-3">
      <input v-model="q" placeholder="Поиск по названию…" class="input min-w-[220px] flex-1" />
      <select v-model="category" class="input">
        <option value="">Все категории</option>
        <option v-for="c in cats?.categories || []" :key="c.id" :value="c.slug">{{ c.name }}</option>
      </select>
      <select v-model="budgetType" class="input">
        <option value="">Любой бюджет</option>
        <option value="fixed">Фиксированный</option>
        <option value="hourly">Почасовой</option>
      </select>
    </div>

    <p v-if="pending" class="text-sm text-slate-500">Загрузка…</p>
    <EmptyState v-else-if="!data?.projects?.length" title="Ничего не найдено" hint="Измените фильтры." />
    <div v-else class="flex flex-col gap-3">
      <ProjectCard v-for="p in data.projects" :key="p.id" :project="p" />
    </div>
  </div>
</template>
