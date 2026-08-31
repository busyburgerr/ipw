<template>
  <div class="freelancers-page">
    <div class="filters">
      <input v-model="q" class="f-input" placeholder="Поиск по специалистам…" />
      <select v-model="category" class="f-input">
        <option value="">Все категории</option>
        <option v-for="c in categories" :key="c.id" :value="c.slug">{{ c.name }}</option>
      </select>
    </div>

    <p v-if="pending" class="hint">Загрузка…</p>
    <p v-else-if="!items.length" class="hint">Специалистов не найдено</p>
    <div v-else class="list">
      <div class="cell" v-for="f in items" :key="f.userId">
        <FreelancerCard :f="f" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import FreelancerCard from '~/components/FreelancerCard.vue'

const api = useApi()
const q = ref('')
const category = ref('')

const { data: cats } = await useAsyncData('frl-cats', () =>
  api.get('/catalog/categories').then((r) => r.data),
)
const categories = computed(() => cats.value?.categories || [])

const query = computed(() => {
  const p = new URLSearchParams()
  if (q.value) p.set('q', q.value)
  if (category.value) p.set('category', category.value)
  return p.toString()
})

const { data, pending } = await useAsyncData(
  'freelancers',
  () => api.get(`/freelancers${query.value ? '?' + query.value : ''}`).then((r) => r.data),
  { watch: [query] },
)
const items = computed(() => data.value?.freelancers || [])
</script>

<style scoped>
.freelancers-page { width: 100%; display: flex; flex-direction: column; gap: 24px; }
.filters { display: flex; gap: 12px; flex-wrap: wrap; }
.f-input {
  padding: 12px 14px; border-radius: 12px; border: none; background: #fff;
  color: #000; font-family: Inter, sans-serif; font-size: 14px; min-width: 220px;
}
.filters .f-input:first-child { flex: 1; }
.list { display: flex; flex-direction: column; gap: 24px; align-items: center; }
.cell { width: 100%; max-width: 815px; }
.hint { color: #fff; font-weight: 500; }
</style>
