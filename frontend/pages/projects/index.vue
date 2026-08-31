<template>
  <div class="projects-page">
    <div class="filters">
      <input v-model="q" class="f-input" placeholder="Поиск по проектам…" />
      <select v-model="category" class="f-input">
        <option value="">Все категории</option>
        <option v-for="c in categories" :key="c.id" :value="c.slug">{{ c.name }}</option>
      </select>
      <select v-model="budgetType" class="f-input">
        <option value="">Любой бюджет</option>
        <option value="fixed">Фиксированная цена</option>
        <option value="hourly">Почасовая</option>
      </select>
    </div>

    <p v-if="pending" class="hint">Загрузка…</p>
    <p v-else-if="!items.length" class="hint">Проектов не найдено</p>
    <div v-else class="grid">
      <section v-for="p in items" :key="p.id">
        <ProjectCard :project="p" />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import ProjectCard from './components/ProjectCard.vue'

const api = useApi()
const route = useRoute()
const router = useRouter()

const q = ref((route.query.q as string) || '')
const category = ref((route.query.category as string) || '')
const budgetType = ref((route.query.budgetType as string) || '')

const { data: cats } = await useAsyncData('catalog', () =>
  api.get('/catalog/categories').then((r) => r.data),
)
const categories = computed(() => cats.value?.categories || [])

const query = computed(() => {
  const p = new URLSearchParams()
  if (q.value) p.set('q', q.value)
  if (category.value) p.set('category', category.value)
  if (budgetType.value) p.set('budgetType', budgetType.value)
  return p.toString()
})

const { data, pending } = await useAsyncData(
  'projects',
  () => api.get(`/projects${query.value ? '?' + query.value : ''}`).then((r) => r.data),
  { watch: [query] },
)
const items = computed(() => data.value?.projects || [])

let t: ReturnType<typeof setTimeout>
watch(query, () => {
  clearTimeout(t)
  t = setTimeout(
    () => router.replace({ query: Object.fromEntries(new URLSearchParams(query.value)) }),
    300,
  )
})
</script>

<style scoped>
.projects-page {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 25px;
}
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.f-input {
  padding: 12px 14px;
  border-radius: 12px;
  border: none;
  background: #fff;
  color: #000;
  font-family: Inter, sans-serif;
  font-size: 14px;
  min-width: 200px;
}
.filters .f-input:first-child {
  flex: 1;
}
.grid {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}
.grid section {
  width: calc(50% - 0.7rem);
}
.hint {
  color: #fff;
  font-weight: 500;
}
@media (max-width: 720px) {
  .grid section {
    width: 100%;
  }
}
</style>
