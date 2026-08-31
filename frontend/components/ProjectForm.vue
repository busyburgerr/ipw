<script setup lang="ts">
const props = defineProps<{ initial?: any; submitLabel: string; error?: string; loading?: boolean }>()
const emit = defineEmits<{ submit: [payload: any] }>()
const { $apiRaw } = useNuxtApp()

interface Category { id: string; slug: string; name: string }
interface Skill { id: string; slug: string; name: string }

const { data: cats } = await useAsyncData('pf-categories', () =>
  $apiRaw<{ categories: Category[] }>('/catalog/categories'),
)

const form = reactive({
  title: props.initial?.title || '',
  description: props.initial?.description || '',
  categoryId: props.initial?.categoryId || '',
  budgetType: props.initial?.budgetType || 'fixed',
  fixedAmount: props.initial?.fixedAmountCents ? props.initial.fixedAmountCents / 100 : '',
  hourlyMin: props.initial?.hourlyRateMinCents ? props.initial.hourlyRateMinCents / 100 : '',
  hourlyMax: props.initial?.hourlyRateMaxCents ? props.initial.hourlyRateMaxCents / 100 : '',
  experienceLevel: props.initial?.experienceLevel || 'any',
  skillIds: new Set<string>(props.initial?.skillIds || []),
})

const skills = ref<Skill[]>([])
watch(
  () => form.categoryId,
  async (id) => {
    const slug = cats.value?.categories.find((c) => c.id === id)?.slug
    skills.value = slug
      ? (await $apiRaw<{ skills: Skill[] }>(`/catalog/skills?category=${slug}`)).skills
      : []
  },
  { immediate: true },
)

function toggleSkill(sid: string) {
  form.skillIds.has(sid) ? form.skillIds.delete(sid) : form.skillIds.add(sid)
}

function submit() {
  const payload: any = {
    title: form.title,
    description: form.description,
    categoryId: form.categoryId || '',
    budgetType: form.budgetType,
    experienceLevel: form.experienceLevel,
    skillIds: [...form.skillIds],
  }
  if (form.budgetType === 'fixed') {
    payload.fixedAmountCents = toCents(form.fixedAmount)
  } else {
    payload.hourlyRateMinCents = toCents(form.hourlyMin)
    payload.hourlyRateMaxCents = toCents(form.hourlyMax)
  }
  emit('submit', payload)
}
</script>

<template>
  <form class="flex flex-col gap-4" @submit.prevent="submit">
    <div class="field">
      <label class="label">Название</label>
      <input v-model="form.title" class="input" minlength="5" required />
    </div>
    <div class="field">
      <label class="label">Описание задачи</label>
      <textarea v-model="form.description" class="input" rows="6" required />
    </div>
    <div class="flex flex-wrap gap-3">
      <div class="field flex-1">
        <label class="label">Категория</label>
        <select v-model="form.categoryId" class="input">
          <option value="">— не указана —</option>
          <option v-for="c in cats?.categories || []" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </div>
      <div class="field">
        <label class="label">Уровень</label>
        <select v-model="form.experienceLevel" class="input">
          <option value="any">любой</option>
          <option value="entry">начинающий</option>
          <option value="intermediate">средний</option>
          <option value="expert">эксперт</option>
        </select>
      </div>
    </div>

    <div v-if="skills.length" class="field">
      <label class="label">Навыки</label>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="s in skills"
          :key="s.id"
          type="button"
          class="pill border"
          :class="form.skillIds.has(s.id)
            ? 'border-brand bg-brand/10 text-brand-700'
            : 'border-slate-300 text-slate-600'"
          @click="toggleSkill(s.id)"
        >{{ s.name }}</button>
      </div>
    </div>

    <div class="field">
      <label class="label">Тип бюджета</label>
      <div class="flex gap-4 text-sm">
        <label class="flex items-center gap-2">
          <input v-model="form.budgetType" type="radio" value="fixed" /> фиксированный
        </label>
        <label class="flex items-center gap-2">
          <input v-model="form.budgetType" type="radio" value="hourly" /> почасовой
        </label>
      </div>
    </div>

    <div v-if="form.budgetType === 'fixed'" class="field w-48">
      <label class="label">Сумма, ₽</label>
      <input v-model="form.fixedAmount" type="number" min="1" step="0.01" class="input" required />
    </div>
    <div v-else class="flex gap-3">
      <div class="field w-40">
        <label class="label">от, ₽/час</label>
        <input v-model="form.hourlyMin" type="number" min="1" step="0.01" class="input" required />
      </div>
      <div class="field w-40">
        <label class="label">до, ₽/час</label>
        <input v-model="form.hourlyMax" type="number" min="1" step="0.01" class="input" required />
      </div>
    </div>

    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
    <button class="btn-primary w-fit" :disabled="loading">{{ submitLabel }}</button>
  </form>
</template>
