<template>
  <div class="container">
    <div class="container-head">
      <h2>Новый проект</h2>
    </div>
    <div class="container-content">
      <form class="create-vacancy_form" @submit.prevent="submit">
        <div class="specialize-block">
          <h3>Название</h3>
          <input v-model="form.title" type="text" minlength="5" required />
        </div>
        <div class="desc-block">
          <h3>Описание задачи</h3>
          <textarea v-model="form.description" rows="6" required />
        </div>
        <div class="location-block">
          <h3>Категория</h3>
          <select v-model="form.categoryId">
            <option value="">— не указана —</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="skills-block" v-if="catSkills.length">
          <h3>Ключевые навыки</h3>
          <div class="skill-input_content">
            <button
              v-for="s in catSkills"
              :key="s.id"
              type="button"
              class="skill-item"
              :class="{ on: form.skillIds.has(s.id) }"
              @click="form.skillIds.has(s.id) ? form.skillIds.delete(s.id) : form.skillIds.add(s.id)"
            >{{ s.name }}</button>
          </div>
        </div>
        <div class="exp-block">
          <h3>Требуемый уровень</h3>
          <select v-model="form.experienceLevel">
            <option value="any">любой</option>
            <option value="entry">junior</option>
            <option value="intermediate">middle</option>
            <option value="expert">senior</option>
          </select>
        </div>
        <div class="salary-block">
          <h3>Бюджет</h3>
          <select v-model="form.budgetType">
            <option value="fixed">фиксированная цена</option>
            <option value="hourly">почасовая оплата</option>
          </select>
          <input v-if="form.budgetType === 'fixed'" v-model="form.fixedAmount" type="number" min="1" step="0.01" placeholder="Сумма, ₽" required />
          <div v-else class="hourly">
            <input v-model="form.hourlyMin" type="number" min="1" step="0.01" placeholder="от, ₽/час" required />
            <input v-model="form.hourlyMax" type="number" min="1" step="0.01" placeholder="до, ₽/час" required />
          </div>
        </div>
        <p v-if="error" class="err">{{ error }}</p>
        <div class="btn-block">
          <NuxtLink class="btn back" to="/requests">Назад</NuxtLink>
          <button class="btn save" type="submit" :disabled="saving">Создать черновик</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()

const { data: cats } = await useAsyncData('new-cats', () =>
  api.get('/catalog/categories').then((r) => r.data),
)
const categories = computed(() => cats.value?.categories || [])

const form = reactive({
  title: '', description: '', categoryId: '',
  budgetType: 'fixed', fixedAmount: '', hourlyMin: '', hourlyMax: '',
  experienceLevel: 'any', skillIds: new Set<string>(),
})

const catSkills = ref<any[]>([])
watch(
  () => form.categoryId,
  async (id) => {
    const slug = categories.value.find((c: any) => c.id === id)?.slug
    catSkills.value = slug ? (await api.get(`/catalog/skills?category=${slug}`)).data.skills : []
  },
)

const saving = ref(false)
const error = ref('')
async function submit() {
  saving.value = true
  error.value = ''
  const body: any = {
    title: form.title,
    description: form.description,
    categoryId: form.categoryId || '',
    budgetType: form.budgetType,
    experienceLevel: form.experienceLevel,
    skillIds: [...form.skillIds],
  }
  if (form.budgetType === 'fixed') body.fixedAmountCents = toCents(form.fixedAmount)
  else {
    body.hourlyRateMinCents = toCents(form.hourlyMin)
    body.hourlyRateMaxCents = toCents(form.hourlyMax)
  }
  try {
    const { data } = await api.post('/projects', body)
    await navigateTo(`/projects/${data.id}/edit?created=1`)
  } catch (e: any) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
@import './assets/css/style.css';

h1, h2, h3, h4, p { margin: 0; padding: 0; }
.container { padding: 20px 0; }
select {
  padding: 10px; border-radius: 5px; border: 0.5px solid rgba(0, 0, 0, 0.5);
  font-size: 16px; font-weight: 500; align-self: stretch;
}
.hourly { display: flex; gap: 10px; align-self: stretch; }
.hourly input { flex: 1; }
.skill-item { border: none; cursor: pointer; }
.skill-item.on { background: #4e92f8; color: #fff; }
.btn.back, .btn.save { text-decoration: none; cursor: pointer; color: #fff; }
.btn.back { color: #000; }
.err { color: #b91c1c; font-weight: 600; align-self: stretch; }
</style>
