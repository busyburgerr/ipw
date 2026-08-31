<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const { $api, $apiRaw } = useNuxtApp()
const id = route.params.id as string

const { data: project, refresh, error: loadError } = await useAsyncData(`project-${id}`, () =>
  $apiRaw<any>(`/projects/${id}`, auth.accessToken ? { headers: { Authorization: `Bearer ${auth.accessToken}` } } : {}),
)

const isOwner = computed(() => auth.user && project.value && auth.user.id === project.value.clientId)
const canPropose = computed(
  () => auth.user?.isFreelancer && project.value?.status === 'open' && !isOwner.value,
)

const budget = computed(() => {
  const p = project.value
  if (!p) return ''
  return p.budgetType === 'fixed'
    ? money(p.fixedAmountCents, p.currency)
    : `${money(p.hourlyRateMinCents, p.currency)} – ${money(p.hourlyRateMaxCents, p.currency)} / час`
})

// --- proposal form (freelancer) ---
const bid = reactive({ coverLetter: '', amount: '', days: '' })
const propose = useSubmit()
async function submitProposal() {
  const ok = await propose.run(() =>
    $api(`/projects/${id}/proposals`, {
      method: 'POST',
      body: {
        coverLetter: bid.coverLetter,
        bidAmountCents: toCents(bid.amount),
        estimatedDays: bid.days ? Number(bid.days) : null,
      },
    }),
  )
  if (ok !== undefined) {
    bid.coverLetter = ''
    bid.amount = ''
    bid.days = ''
    await refresh()
    submitted.value = true
  }
}
const submitted = ref(false)

// --- proposals list (owner) ---
const proposals = ref<any[]>([])
async function loadProposals() {
  if (!isOwner.value) return
  const r = await $api<{ proposals: any[] }>(`/projects/${id}/proposals`)
  proposals.value = r.proposals
}
onMounted(loadProposals)

const act = useSubmit()
async function proposalAction(pid: string, action: 'accept' | 'shortlist' | 'decline') {
  const path = action === 'accept' ? `/proposals/${pid}/accept` : `/proposals/${pid}/${action}`
  const res = await act.run(() => $api<any>(path, { method: 'POST' }))
  if (res === undefined) return
  if (action === 'accept') await navigateTo(`/contracts/${res.id}`)
  else {
    await loadProposals()
    await refresh()
  }
}

async function closeProject() {
  await act.run(() => $api(`/projects/${id}/close`, { method: 'POST' }))
  await refresh()
}
</script>

<template>
  <div v-if="loadError" class="card text-rose-600">{{ loadError.message }}</div>
  <div v-else-if="project" class="flex flex-col gap-6">
    <div>
      <NuxtLink to="/projects" class="text-sm">← к проектам</NuxtLink>
      <div class="mt-2 flex items-start justify-between gap-4">
        <h1>{{ project.title }}</h1>
        <StatusPill :status="project.status" />
      </div>
    </div>

    <div class="card flex flex-col gap-4">
      <div class="flex flex-wrap gap-x-8 gap-y-2 text-sm">
        <div><span class="text-slate-500">Бюджет:</span> <b>{{ budget }}</b></div>
        <div><span class="text-slate-500">Тип:</span> {{ project.budgetType === 'fixed' ? 'фикс' : 'почасовой' }}</div>
        <div><span class="text-slate-500">Опыт:</span> {{ project.experienceLevel }}</div>
        <div><span class="text-slate-500">Откликов:</span> {{ project.proposalsCount }}</div>
      </div>
      <p class="whitespace-pre-wrap text-sm text-slate-700">{{ project.description }}</p>
    </div>

    <!-- owner controls -->
    <div v-if="isOwner" class="flex flex-col gap-4">
      <div class="flex items-center justify-between">
        <h2>Отклики</h2>
        <div class="flex gap-2">
          <NuxtLink
            v-if="project.status === 'draft'"
            :to="`/projects/${id}/edit`"
            class="btn-ghost btn-sm no-underline"
          >Редактировать</NuxtLink>
          <button
            v-if="['draft', 'open'].includes(project.status)"
            class="btn-ghost btn-sm"
            @click="closeProject"
          >Закрыть проект</button>
        </div>
      </div>
      <p v-if="act.error.value" class="text-sm text-rose-600">{{ act.error.value }}</p>
      <EmptyState v-if="!proposals.length" title="Пока нет откликов" />
      <div v-for="p in proposals" :key="p.id" class="card flex flex-col gap-2">
        <div class="flex items-center justify-between">
          <b>{{ money(p.bidAmountCents, project.currency) }}</b>
          <StatusPill :status="p.status" />
        </div>
        <p class="text-sm text-slate-500">
          <template v-if="p.estimatedDays">срок ~{{ p.estimatedDays }} дн. · </template>{{ relTime(p.createdAt) }}
        </p>
        <p class="whitespace-pre-wrap text-sm text-slate-700">{{ p.coverLetter }}</p>
        <div v-if="['pending', 'shortlisted'].includes(p.status)" class="flex gap-2 pt-1">
          <button class="btn-primary btn-sm" @click="proposalAction(p.id, 'accept')">Принять</button>
          <button
            v-if="p.status === 'pending'"
            class="btn-ghost btn-sm"
            @click="proposalAction(p.id, 'shortlist')"
          >В шорт-лист</button>
          <button class="btn-ghost btn-sm" @click="proposalAction(p.id, 'decline')">Отклонить</button>
        </div>
      </div>
    </div>

    <!-- freelancer proposal form -->
    <div v-else-if="canPropose" class="card flex flex-col gap-4">
      <h2>Ваш отклик</h2>
      <template v-if="submitted">
        <p class="text-sm text-emerald-700">Отклик отправлен. Следите за статусом в разделе «Мои отклики».</p>
        <NuxtLink to="/me/proposals" class="btn-ghost btn-sm w-fit no-underline">Мои отклики</NuxtLink>
      </template>
      <form v-else class="flex flex-col gap-3" @submit.prevent="submitProposal">
        <div class="field">
          <label class="label">Сопроводительное письмо</label>
          <textarea v-model="bid.coverLetter" class="input" rows="4" required />
        </div>
        <div class="flex gap-3">
          <div class="field flex-1">
            <label class="label">Ставка ({{ project.currency }})</label>
            <input v-model="bid.amount" type="number" min="1" step="0.01" class="input" required />
          </div>
          <div class="field w-32">
            <label class="label">Срок, дней</label>
            <input v-model="bid.days" type="number" min="1" class="input" />
          </div>
        </div>
        <p v-if="propose.error.value" class="text-sm text-rose-600">{{ propose.error.value }}</p>
        <button class="btn-primary w-fit" :disabled="propose.loading.value">Отправить отклик</button>
      </form>
    </div>

    <div v-else-if="!auth.isAuthed" class="card text-sm text-slate-600">
      <NuxtLink to="/auth/login">Войдите</NuxtLink>, чтобы откликнуться на проект.
    </div>
  </div>
</template>
