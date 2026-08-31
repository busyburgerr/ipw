<template>
  <div v-if="project" class="container">
    <div class="container-head">
      <div class="company">
        <div class="logo" />
        <div class="company-data">
          <div class="name">Проект</div>
          <div class="location">{{ ru(project.status) }}</div>
          <NuxtLink
            v-if="isOwner && project.status === 'draft'"
            class="btn about-company_btn"
            :to="`/projects/${project.id}/edit`"
          >Редактировать</NuxtLink>
          <button
            v-else-if="isOwner && ['draft', 'open'].includes(project.status)"
            class="btn about-company_btn"
            @click="closeProject"
          >Закрыть проект</button>
        </div>
      </div>

      <div class="main-vacancy">
        <div class="skills-block">
          <div class="skills-head">Требуемые навыки</div>
          <div class="skills-content">
            <div class="skill-item" v-for="s in skillNames" :key="s">{{ s }}</div>
            <div v-if="!skillNames.length" class="location">не указаны</div>
          </div>
        </div>
        <div class="vacancy-info">
          <div class="vacancy-data">
            <div class="vacancy-titleSalary_block">
              <div class="vacancy-title">{{ project.title }}</div>
              <div class="vacancy-salary">{{ budget }}</div>
            </div>
            <div class="employment-exp_block">
              <div class="vacancy-description-item">Уровень: <span>{{ ru(project.experienceLevel) }}</span></div>
              <div class="vacancy-description-item">{{ project.proposalsCount }} откликов</div>
            </div>
          </div>
          <button v-if="canPropose" class="btn req-btn" @click="showForm = !showForm">Откликнуться</button>
        </div>
      </div>
    </div>

    <div class="container-content">
      <div class="about-company_block">
        <h2 class="about-company_head">О проекте</h2>
        <div class="about-company_content">{{ project.description }}</div>
      </div>

      <div v-if="canPropose && showForm" class="about-company_block">
        <h2 class="about-company_head">Ваш отклик</h2>
        <form class="prop-form" @submit.prevent="submitProposal">
          <textarea v-model="bid.coverLetter" placeholder="Сопроводительное письмо" rows="4" required />
          <div class="prop-row">
            <input v-model="bid.amount" type="number" min="1" step="0.01" placeholder="Ваша ставка, ₽" required />
            <input v-model="bid.days" type="number" min="1" placeholder="Срок, дней" />
          </div>
          <p v-if="formError" class="err">{{ formError }}</p>
          <button class="btn success" type="submit" :disabled="sending">Отправить отклик</button>
        </form>
      </div>
      <div v-else-if="submitted" class="about-company_block">
        <div class="about-company_content">
          Отклик отправлен. Статус — в разделе <NuxtLink to="/my-proposals">«Мои отклики»</NuxtLink>.
        </div>
      </div>

      <div v-if="isOwner" class="about-company_block">
        <h2 class="about-company_head">Отклики ({{ proposals.length }})</h2>
        <p v-if="actError" class="err">{{ actError }}</p>
        <div class="prop-list">
          <div v-for="p in proposals" :key="p.id" class="prop-card">
            <div class="prop-card_head">
              <b>{{ money(p.bidAmountCents, project.currency) }}</b>
              <span class="location">{{ ru(p.status) }}</span>
            </div>
            <p class="prop-card_letter">{{ p.coverLetter }}</p>
            <div v-if="['pending', 'shortlisted'].includes(p.status)" class="prop-card_btns">
              <button class="btn success" @click="accept(p.id)">Принять</button>
              <button v-if="p.status === 'pending'" class="btn ghost" @click="decide(p.id, 'shortlist')">В шорт-лист</button>
              <button class="btn ghost" @click="decide(p.id, 'decline')">Отклонить</button>
            </div>
          </div>
          <div v-if="!proposals.length" class="location">Пока нет откликов</div>
        </div>
      </div>

      <div v-else-if="!isAuthed.value" class="about-company_block">
        <div class="about-company_content">
          <NuxtLink to="/auth">Войдите</NuxtLink>, чтобы откликнуться на проект.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const api = useApi()
const route = useRoute()
const { user, isAuthed } = useAuth()
const id = route.params.id as string

const { data: project, refresh } = await useAsyncData(`project-${id}`, () =>
  api.get(`/projects/${id}`).then((r) => r.data).catch(() => null),
)
const { data: allSkills } = await useAsyncData('all-skills', () =>
  api.get('/catalog/skills').then((r) => r.data),
)

const skillMap = computed<Record<string, string>>(() => {
  const m: Record<string, string> = {}
  for (const s of allSkills.value?.skills || []) m[s.id] = s.name
  return m
})
const skillNames = computed(() =>
  (project.value?.skillIds || []).map((x: string) => skillMap.value[x]).filter(Boolean),
)

const isOwner = computed(() => user.value && project.value && user.value.id === project.value.clientId)
const canPropose = computed(
  () => user.value?.isFreelancer && project.value?.status === 'open' && !isOwner.value,
)
const budget = computed(() => {
  const p = project.value
  if (!p) return ''
  return p.budgetType === 'fixed'
    ? money(p.fixedAmountCents, p.currency)
    : `${money(p.hourlyRateMinCents, p.currency)} – ${money(p.hourlyRateMaxCents, p.currency)} / час`
})

const showForm = ref(false)
const submitted = ref(false)
const sending = ref(false)
const formError = ref('')
const bid = reactive({ coverLetter: '', amount: '', days: '' })
async function submitProposal() {
  sending.value = true
  formError.value = ''
  try {
    await api.post(`/projects/${id}/proposals`, {
      coverLetter: bid.coverLetter,
      bidAmountCents: toCents(bid.amount),
      estimatedDays: bid.days ? Number(bid.days) : null,
    })
    showForm.value = false
    submitted.value = true
    await refresh()
  } catch (e: any) {
    formError.value = e.message
  } finally {
    sending.value = false
  }
}

const proposals = ref<any[]>([])
const actError = ref('')
async function loadProposals() {
  if (!isOwner.value) return
  proposals.value = (await api.get(`/projects/${id}/proposals`)).data.proposals
}
onMounted(loadProposals)

async function accept(pid: string) {
  actError.value = ''
  try {
    const { data } = await api.post(`/proposals/${pid}/accept`)
    await navigateTo(`/contracts/${data.id}`)
  } catch (e: any) {
    actError.value = e.message
  }
}
async function decide(pid: string, action: 'shortlist' | 'decline') {
  try {
    await api.post(`/proposals/${pid}/${action}`)
    await loadProposals()
    await refresh()
  } catch (e: any) {
    actError.value = e.message
  }
}
async function closeProject() {
  try {
    await api.post(`/projects/${id}/close`)
    await refresh()
  } catch (e: any) {
    actError.value = e.message
  }
}
</script>

<style scoped>
@import './assets/css/vacancies-[id].css';

* { font-family: Inter, sans-serif; }
h1, h2, h3, h4, p { margin: 0; padding: 0; }

.btn { text-decoration: none; cursor: pointer; }
.err { color: #b91c1c; font-weight: 500; }

.prop-form { display: flex; flex-direction: column; gap: 12px; width: 100%; }
.prop-form textarea,
.prop-form input {
  border: 1px solid rgba(0, 0, 0, 0.25); border-radius: 10px; padding: 12px; font-family: Inter, sans-serif;
}
.prop-row { display: flex; gap: 12px; }
.prop-row input { flex: 1; }

.success {
  padding: 10px 24px; border-radius: 10px; background: #198754; color: #fff; font-weight: 500; border: none;
}
.ghost {
  padding: 10px 24px; border-radius: 10px; background: #eef1f4; color: #000; font-weight: 500; border: none;
}

.prop-list { display: flex; flex-direction: column; gap: 16px; width: 100%; }
.prop-card { border: 1px solid #e5e7eb; border-radius: 16px; padding: 16px; display: flex; flex-direction: column; gap: 8px; }
.prop-card_head { display: flex; justify-content: space-between; align-items: center; }
.prop-card_letter { font-size: 14px; font-weight: 500; white-space: break-spaces; }
.prop-card_btns { display: flex; gap: 10px; }
</style>
