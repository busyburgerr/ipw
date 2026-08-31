<template>
  <div class="container" v-if="project">
    <div class="container-head">
      <h2>Редактирование проекта</h2>
      <div class="st">{{ ru(project.status) }}</div>
    </div>

    <div v-if="route.query.created" class="note">Черновик создан. Проверьте детали и опубликуйте.</div>

    <div class="container-content">
      <form class="create-vacancy_form" @submit.prevent="save">
        <div class="specialize-block">
          <h3>Название</h3>
          <input v-model="form.title" minlength="5" required />
        </div>
        <div class="desc-block">
          <h3>Описание</h3>
          <textarea v-model="form.description" rows="6" required />
        </div>
        <div class="salary-block">
          <h3>Бюджет ({{ ru(project.budgetType) }})</h3>
          <input
            v-if="project.budgetType === 'fixed'"
            v-model="form.fixedAmount"
            type="number" min="1" step="0.01"
          />
          <div v-else class="hourly">
            <input v-model="form.hourlyMin" type="number" min="1" step="0.01" />
            <input v-model="form.hourlyMax" type="number" min="1" step="0.01" />
          </div>
        </div>
        <p v-if="error" class="err">{{ error }}</p>
        <div class="btn-block">
          <button class="btn save" :disabled="busy">Сохранить</button>
        </div>
      </form>

      <div class="publish" v-if="project.status === 'draft'">
        <button class="btn publish-btn" :disabled="busy" @click="publish">Опубликовать проект</button>
        <span>После публикации проект увидят фрилансеры.</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()
const route = useRoute()
const id = route.params.id as string

const { data: project, refresh } = await useAsyncData(`edit-${id}`, () =>
  api.get(`/projects/${id}`).then((r) => r.data),
)

const form = reactive({
  title: project.value?.title || '',
  description: project.value?.description || '',
  fixedAmount: project.value?.fixedAmountCents ? project.value.fixedAmountCents / 100 : '',
  hourlyMin: project.value?.hourlyRateMinCents ? project.value.hourlyRateMinCents / 100 : '',
  hourlyMax: project.value?.hourlyRateMaxCents ? project.value.hourlyRateMaxCents / 100 : '',
})

const busy = ref(false)
const error = ref('')

async function save() {
  busy.value = true
  error.value = ''
  const body: any = {
    title: form.title,
    description: form.description,
    categoryId: project.value.categoryId || '',
    budgetType: project.value.budgetType,
    experienceLevel: project.value.experienceLevel,
    skillIds: project.value.skillIds || [],
  }
  if (project.value.budgetType === 'fixed') body.fixedAmountCents = toCents(form.fixedAmount)
  else {
    body.hourlyRateMinCents = toCents(form.hourlyMin)
    body.hourlyRateMaxCents = toCents(form.hourlyMax)
  }
  try {
    await api.put(`/projects/${id}`, body)
    await navigateTo(`/projects/${id}`)
  } catch (e: any) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}

async function publish() {
  busy.value = true
  error.value = ''
  try {
    await api.post(`/projects/${id}/publish`)
    await navigateTo(`/projects/${id}`)
  } catch (e: any) {
    error.value = e.message
    busy.value = false
  }
}
</script>

<style scoped>
@import '../new/assets/css/style.css';

h1, h2, h3, h4, p { margin: 0; padding: 0; }
.container { padding: 20px 0; }
.st { background: #eef1f4; color: #374151; padding: 4px 14px; border-radius: 999px; font-size: 12px; font-weight: 600; }
.note { background: #dcfce7; color: #166534; border-radius: 12px; padding: 12px 18px; font-weight: 500; }
.hourly { display: flex; gap: 10px; align-self: stretch; }
.hourly input { flex: 1; }
.btn.save { color: #fff; cursor: pointer; }
.publish { display: flex; align-items: center; gap: 14px; margin-top: 10px; color: #000; }
.publish-btn {
  background: #198754; color: #fff; border: none; padding: 10px 24px;
  border-radius: 10px; font-weight: 600; cursor: pointer;
}
.err { color: #b91c1c; font-weight: 600; }
</style>
