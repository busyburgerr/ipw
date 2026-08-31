<template>
  <div class="contract" v-if="contract">
    <NuxtLink to="/contracts" class="back">← контракты</NuxtLink>

    <div class="head card">
      <div>
        <h1>Контракт · {{ money(contract.agreedAmountCents, contract.currency) }}</h1>
        <p class="sub">{{ ru(contract.type) }} · начат {{ fmtDate(contract.startedAt) }}</p>
      </div>
      <span class="status">{{ ru(contract.status) }}</span>
    </div>

    <div v-if="contract.status === 'disputed'" class="card warn">
      По контракту открыт спор — операции с этапами заморожены до решения арбитра.
    </div>

    <p v-if="error" class="err">{{ error }}</p>

    <section class="card">
      <div class="section-head">
        <h2>Этапы</h2>
        <button
          v-if="isClient && contract.status === 'active' && contract.type === 'fixed'"
          class="btn ghost"
          @click="showAdd = !showAdd"
        >＋ Этап</button>
      </div>

      <form v-if="showAdd" class="ms-form" @submit.prevent="addMilestone">
        <input v-model="ms.title" placeholder="Название этапа" required />
        <textarea v-model="ms.description" rows="2" placeholder="Описание" />
        <div class="ms-row">
          <input v-model="ms.amount" type="number" min="1" step="0.01" placeholder="Сумма, ₽" required />
          <input v-model="ms.dueDate" type="date" />
        </div>
        <button class="btn blue" :disabled="busy">Добавить</button>
      </form>

      <p v-if="!milestones.length" class="muted">Этапов пока нет</p>
      <div v-for="m in milestones" :key="m.id" class="ms">
        <div class="ms-top">
          <div>
            <div class="ms-title">{{ m.sequence }}. {{ m.title }}</div>
            <div v-if="m.description" class="muted">{{ m.description }}</div>
          </div>
          <div class="ms-right">
            <b>{{ money(m.amountCents, contract.currency) }}</b>
            <span class="status sm">{{ ru(m.status) }}</span>
          </div>
        </div>
        <p v-if="m.deliverableNote" class="note">Сдача: {{ m.deliverableNote }}</p>
        <div v-if="contract.status === 'active'" class="ms-btns">
          <button v-if="isClient && m.status === 'pending'" class="btn blue" @click="fund(m.id)">Оплатить в escrow</button>
          <button v-if="isClient && m.status === 'pending'" class="btn ghost" @click="msAction(m.id, 'cancel')">Отменить</button>
          <button v-if="isFreelancer && m.status === 'funded'" class="btn blue" @click="submitWork(m.id)">Сдать работу</button>
          <button v-if="isClient && m.status === 'funded'" class="btn ghost" @click="msAction(m.id, 'refund')">Вернуть средства</button>
          <template v-if="isClient && m.status === 'submitted'">
            <button class="btn blue" @click="msAction(m.id, 'approve')">Принять и выплатить</button>
            <button class="btn ghost" @click="msAction(m.id, 'request-changes')">На доработку</button>
          </template>
        </div>
      </div>
    </section>

    <div class="actions">
      <button v-if="isClient && contract.status === 'active' && allResolved" class="btn navy" @click="complete">
        Завершить контракт
      </button>
      <button v-if="contract.status === 'active'" class="btn ghost" @click="showDispute = !showDispute">
        Открыть спор
      </button>
    </div>

    <form v-if="showDispute" class="card ms-form" @submit.prevent="raiseDispute">
      <h2>Спор</h2>
      <select v-model="disputeMs">
        <option value="">весь контракт</option>
        <option v-for="m in milestones" :key="m.id" :value="m.id">{{ m.sequence }}. {{ m.title }}</option>
      </select>
      <textarea v-model="disputeReason" rows="3" placeholder="Суть спора" required />
      <button class="btn danger" :disabled="busy">Открыть спор</button>
    </form>

    <section v-if="contract.status === 'completed'" class="card">
      <h2>Отзывы</h2>
      <form v-if="!myReview" class="ms-form" @submit.prevent="submitReview">
        <select v-model.number="rev.rating">
          <option v-for="n in 5" :key="n" :value="n">{{ n }} ★</option>
        </select>
        <textarea v-model="rev.comment" rows="3" placeholder="Комментарий" />
        <button class="btn blue" :disabled="busy">Оставить отзыв</button>
        <p class="muted">Отзыв станет виден, когда его оставят обе стороны (или через 14 дней).</p>
      </form>
      <div v-for="r in reviews" :key="r.id" class="review">
        <div class="stars">{{ '★'.repeat(r.rating) }}{{ '☆'.repeat(5 - r.rating) }}</div>
        <span class="muted">{{ r.published ? 'опубликован' : 'ждём вторую сторону' }}</span>
        <p v-if="r.comment">{{ r.comment }}</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()
const { user } = useAuth()
const id = useRoute().params.id as string

const { data, refresh } = await useAsyncData(`contract-${id}`, () =>
  api.get(`/contracts/${id}`).then((r) => r.data),
)
const contract = computed(() => data.value?.contract)
const milestones = computed(() => data.value?.milestones || [])
const isClient = computed(() => user.value?.id === contract.value?.clientId)
const isFreelancer = computed(() => user.value?.id === contract.value?.freelancerId)
const allResolved = computed(
  () => milestones.value.length > 0 &&
    milestones.value.every((m: any) => ['approved', 'released', 'cancelled'].includes(m.status)),
)

const busy = ref(false)
const error = ref('')
async function guard(fn: () => Promise<any>) {
  busy.value = true
  error.value = ''
  try {
    await fn()
    await refresh()
  } catch (e: any) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}

const showAdd = ref(false)
const ms = reactive({ title: '', description: '', amount: '', dueDate: '' })
const addMilestone = () =>
  guard(async () => {
    await api.post(`/contracts/${id}/milestones`, {
      title: ms.title, description: ms.description,
      amountCents: toCents(ms.amount), dueDate: ms.dueDate || null,
    })
    Object.assign(ms, { title: '', description: '', amount: '', dueDate: '' })
    showAdd.value = false
  })

async function fund(mid: string) {
  await guard(async () => {
    const { data: pay } = await api.post(`/milestones/${mid}/fund`)
    if (pay.paymentUrl?.includes('/dev/payments/')) {
      await api.post(pay.paymentUrl.replace(useRuntimeConfig().public.apiBase, ''))
    } else {
      window.open(pay.paymentUrl, '_blank')
    }
  })
}

const msAction = (mid: string, action: string) => guard(() => api.post(`/milestones/${mid}/${action}`))

async function submitWork(mid: string) {
  const note = prompt('Комментарий к сдаче (ссылки, доступы):') ?? ''
  await guard(() => api.post(`/milestones/${mid}/submit`, { deliverableNote: note }))
}

const complete = () => guard(() => api.post(`/contracts/${id}/complete`))

const showDispute = ref(false)
const disputeReason = ref('')
const disputeMs = ref('')
const raiseDispute = () =>
  guard(async () => {
    await api.post(`/contracts/${id}/disputes`, { reason: disputeReason.value, milestoneId: disputeMs.value || '' })
    showDispute.value = false
    disputeReason.value = ''
  })

const reviews = ref<any[]>([])
const myReview = computed(() => reviews.value.find((r) => r.reviewerId === user.value?.id))
async function loadReviews() {
  if (contract.value?.status !== 'completed') return
  reviews.value = (await api.get(`/contracts/${id}/reviews`)).data.reviews
}
watch(contract, loadReviews, { immediate: true })

const rev = reactive({ rating: 5, comment: '' })
const submitReview = () =>
  guard(async () => {
    await api.post(`/contracts/${id}/reviews`, { rating: rev.rating, comment: rev.comment })
    rev.comment = ''
    await loadReviews()
  })
</script>

<style scoped>
.contract { width: 100%; display: flex; flex-direction: column; gap: 18px; }
.back { color: #fff; font-weight: 600; text-decoration: none; }
.card { background: #fff; color: #000; border-radius: 24px; padding: 24px; }
.head { display: flex; justify-content: space-between; align-items: flex-start; }
.sub { font-size: 13px; font-weight: 500; color: #6b7280; margin-top: 4px; }
.warn { background: #fee2e2; color: #991b1b; font-weight: 500; }
.err { color: #fecaca; font-weight: 600; }
.section-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.muted { color: #9ca3af; font-weight: 500; }
.note { background: #f3f4f6; border-radius: 8px; padding: 8px 12px; font-size: 14px; font-weight: 500; margin: 8px 0; }

.ms-form { display: flex; flex-direction: column; gap: 10px; margin: 12px 0; }
.ms-form input, .ms-form textarea, .ms-form select {
  border: 1px solid #d1d5db; border-radius: 10px; padding: 10px; font-family: Inter, sans-serif; font-size: 14px;
}
.ms-row { display: flex; gap: 10px; }
.ms-row input { flex: 1; }

.ms { border-top: 1px solid #eef1f4; padding: 14px 0; }
.ms-top { display: flex; justify-content: space-between; gap: 12px; }
.ms-title { font-weight: 700; }
.ms-right { display: flex; flex-direction: column; align-items: flex-end; gap: 6px; }
.ms-btns { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 10px; }

.status { background: #eef1f4; color: #374151; padding: 4px 12px; border-radius: 999px; font-size: 12px; font-weight: 600; }
.status.sm { font-size: 11px; }
.actions { display: flex; gap: 12px; }

.btn { border: none; border-radius: 10px; padding: 9px 20px; font-weight: 600; cursor: pointer; font-size: 14px; }
.btn.blue { background: #4e92f8; color: #fff; }
.btn.navy { background: #2f3e57; color: #fff; }
.btn.ghost { background: #eef1f4; color: #111827; }
.btn.danger { background: #dc2626; color: #fff; }

.review { border-top: 1px solid #eef1f4; padding: 12px 0; }
.stars { color: #f59e0b; }
.review p { font-size: 14px; font-weight: 500; margin-top: 4px; }
</style>
