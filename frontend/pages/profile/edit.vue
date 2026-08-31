<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()
const { $api, $apiRaw } = useNuxtApp()

const tab = ref<'freelancer' | 'client'>(auth.user?.isFreelancer ? 'freelancer' : 'client')

// --- freelancer ---
const { data: frl } = await useAsyncData(
  'my-frl-profile',
  () => (auth.user?.isFreelancer ? $api<any>('/me/freelancer-profile').catch(() => null) : Promise.resolve(null)),
)
const { data: cats } = await useAsyncData('profile-cats', () =>
  $apiRaw<{ categories: any[] }>('/catalog/categories'),
)

const f = reactive({
  headline: frl.value?.headline || '',
  bio: frl.value?.bio || '',
  hourlyRate: frl.value?.hourlyRateCents ? frl.value.hourlyRateCents / 100 : '',
  availability: frl.value?.availability || 'available',
  primaryCategoryId: frl.value?.primaryCategoryId || '',
  location: frl.value?.location || '',
  languages: (frl.value?.languages || []).join(', '),
  skillIds: new Set<string>((frl.value?.skills || []).map((s: any) => s.id)),
})

const skills = ref<any[]>([])
watch(
  () => f.primaryCategoryId,
  async (id) => {
    const slug = cats.value?.categories.find((c) => c.id === id)?.slug
    skills.value = slug ? (await $apiRaw<any>(`/catalog/skills?category=${slug}`)).skills : []
  },
  { immediate: true },
)

const frlSubmit = useSubmit()
async function saveFreelancer() {
  await frlSubmit.run(async () => {
    await $api('/me/freelancer-profile', {
      method: 'PUT',
      body: {
        headline: f.headline,
        bio: f.bio,
        hourlyRateCents: toCents(f.hourlyRate || 0),
        availability: f.availability,
        primaryCategoryId: f.primaryCategoryId || '',
        location: f.location,
        languages: f.languages.split(',').map((s) => s.trim()).filter(Boolean),
      },
    })
    await $api('/me/freelancer-profile/skills', {
      method: 'PUT',
      body: { skillIds: [...f.skillIds] },
    })
  })
  saved.value = 'freelancer'
}

// --- client ---
const { data: cln } = await useAsyncData(
  'my-cln-profile',
  () => (auth.user?.isClient ? $api<any>('/me/client-profile').catch(() => null) : Promise.resolve(null)),
)
const c = reactive({
  companyName: cln.value?.companyName || '',
  about: cln.value?.about || '',
  website: cln.value?.website || '',
  location: cln.value?.location || '',
})
const clnSubmit = useSubmit()
async function saveClient() {
  await clnSubmit.run(() => $api('/me/client-profile', { method: 'PUT', body: { ...c } }))
  saved.value = 'client'
}

const saved = ref('')
</script>

<template>
  <div class="mx-auto max-w-2xl">
    <h1 class="mb-6">Профиль</h1>

    <div class="mb-6 flex gap-2 border-b border-slate-200">
      <button
        v-if="auth.user?.isFreelancer"
        class="border-b-2 px-3 py-2 text-sm"
        :class="tab === 'freelancer' ? 'border-brand font-semibold text-brand-700' : 'border-transparent text-slate-500'"
        @click="tab = 'freelancer'"
      >Фрилансер</button>
      <button
        v-if="auth.user?.isClient"
        class="border-b-2 px-3 py-2 text-sm"
        :class="tab === 'client' ? 'border-brand font-semibold text-brand-700' : 'border-transparent text-slate-500'"
        @click="tab = 'client'"
      >Заказчик</button>
    </div>

    <form v-if="tab === 'freelancer'" class="flex flex-col gap-4" @submit.prevent="saveFreelancer">
      <div class="field">
        <label class="label">Заголовок</label>
        <input v-model="f.headline" class="input" placeholder="Backend-разработчик на Go" />
      </div>
      <div class="field">
        <label class="label">О себе</label>
        <textarea v-model="f.bio" class="input" rows="4" />
      </div>
      <div class="flex flex-wrap gap-3">
        <div class="field w-40">
          <label class="label">Ставка, ₽/час</label>
          <input v-model="f.hourlyRate" type="number" min="0" class="input" />
        </div>
        <div class="field w-44">
          <label class="label">Доступность</label>
          <select v-model="f.availability" class="input">
            <option value="available">доступен</option>
            <option value="limited">ограниченно</option>
            <option value="unavailable">занят</option>
          </select>
        </div>
        <div class="field flex-1">
          <label class="label">Локация</label>
          <input v-model="f.location" class="input" />
        </div>
      </div>
      <div class="field">
        <label class="label">Основная категория</label>
        <select v-model="f.primaryCategoryId" class="input">
          <option value="">— не выбрана —</option>
          <option v-for="cat in cats?.categories || []" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
        </select>
      </div>
      <div v-if="skills.length" class="field">
        <label class="label">Навыки</label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="s in skills"
            :key="s.id"
            type="button"
            class="pill border"
            :class="f.skillIds.has(s.id) ? 'border-brand bg-brand/10 text-brand-700' : 'border-slate-300 text-slate-600'"
            @click="f.skillIds.has(s.id) ? f.skillIds.delete(s.id) : f.skillIds.add(s.id)"
          >{{ s.name }}</button>
        </div>
      </div>
      <div class="field">
        <label class="label">Языки (через запятую)</label>
        <input v-model="f.languages" class="input" placeholder="Русский, English" />
      </div>
      <p v-if="frlSubmit.error.value" class="text-sm text-rose-600">{{ frlSubmit.error.value }}</p>
      <p v-if="saved === 'freelancer'" class="text-sm text-emerald-700">Сохранено.</p>
      <button class="btn-primary w-fit" :disabled="frlSubmit.loading.value">Сохранить</button>
    </form>

    <form v-else class="flex flex-col gap-4" @submit.prevent="saveClient">
      <div class="field">
        <label class="label">Компания</label>
        <input v-model="c.companyName" class="input" />
      </div>
      <div class="field">
        <label class="label">О компании</label>
        <textarea v-model="c.about" class="input" rows="4" />
      </div>
      <div class="field">
        <label class="label">Сайт</label>
        <input v-model="c.website" class="input" />
      </div>
      <div class="field">
        <label class="label">Локация</label>
        <input v-model="c.location" class="input" />
      </div>
      <p v-if="clnSubmit.error.value" class="text-sm text-rose-600">{{ clnSubmit.error.value }}</p>
      <p v-if="saved === 'client'" class="text-sm text-emerald-700">Сохранено.</p>
      <button class="btn-primary w-fit" :disabled="clnSubmit.loading.value">Сохранить</button>
    </form>
  </div>
</template>
