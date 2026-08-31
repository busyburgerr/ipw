<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
import { useAuthStore } from '~/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const { $api } = useNuxtApp()
const id = route.params.id as string

const { data, refresh, error: loadError } = await useAsyncData(`contract-${id}`, () =>
  $api<{ contract: any; milestones: any[] }>(`/contracts/${id}`),
)

const contract = computed(() => data.value?.contract)
const milestones = computed(() => data.value?.milestones || [])
const isClient = computed(() => auth.user?.id === contract.value?.clientId)
const isFreelancer = computed(() => auth.user?.id === contract.value?.freelancerId)
const allResolved = computed(
  () =>
    milestones.value.length > 0 &&
    milestones.value.every((m) => ['approved', 'released', 'cancelled'].includes(m.status)),
)

const { loading, error, run } = useSubmit()

async function act(fn: () => Promise<any>) {
  const r = await run(fn)
  if (r !== undefined) await refresh()
  return r
}

// milestone form (client)
const showAdd = ref(false)
const ms = reactive({ title: '', description: '', amount: '', dueDate: '' })
async function addMilestone() {
  const r = await act(() =>
    $api(`/contracts/${id}/milestones`, {
      method: 'POST',
      body: {
        title: ms.title,
        description: ms.description,
        amountCents: toCents(ms.amount),
        dueDate: ms.dueDate || null,
      },
    }),
  )
  if (r !== undefined) {
    Object.assign(ms, { title: '', description: '', amount: '', dueDate: '' })
    showAdd.value = false
  }
}

async function fund(mid: string) {
  const payment = await run(() => $api<any>(`/milestones/${mid}/fund`, { method: 'POST' }))
  if (!payment) return
  if (payment.paymentUrl?.includes('/dev/payments/')) {
    await run(() => $fetch(payment.paymentUrl, { method: 'POST' }))
    await refresh()
  } else {
    window.open(payment.paymentUrl, '_blank')
  }
}

const milestoneAction = (mid: string, action: string) =>
  act(() => $api(`/milestones/${mid}/${action}`, { method: 'POST' }))

async function submitWork(mid: string) {
  const note = prompt('Комментарий к сдаче (ссылки, доступы):') ?? ''
  await act(() => $api(`/milestones/${mid}/submit`, { method: 'POST', body: { deliverableNote: note } }))
}

async function completeContract() {
  await act(() => $api(`/contracts/${id}/complete`, { method: 'POST' }))
}

// dispute
const showDispute = ref(false)
const disputeReason = ref('')
const disputeMilestone = ref('')
async function raiseDispute() {
  const r = await act(() =>
    $api(`/contracts/${id}/disputes`, {
      method: 'POST',
      body: { reason: disputeReason.value, milestoneId: disputeMilestone.value || '' },
    }),
  )
  if (r !== undefined) {
    showDispute.value = false
    disputeReason.value = ''
  }
}

// reviews
const reviews = ref<any[]>([])
const myReview = computed(() => reviews.value.find((r) => r.reviewerId === auth.user?.id))
async function loadReviews() {
  if (contract.value?.status !== 'completed') return
  reviews.value = (await $api<{ reviews: any[] }>(`/contracts/${id}/reviews`)).reviews
}
watch(contract, loadReviews, { immediate: true })

const rev = reactive({ rating: 5, comment: '' })
async function submitReview() {
  const r = await run(() =>
    $api(`/contracts/${id}/reviews`, { method: 'POST', body: { rating: rev.rating, comment: rev.comment } }),
  )
  if (r !== undefined) {
    rev.comment = ''
    await loadReviews()
  }
}
</script>

<template>
  <div v-if="loadError" class="card text-rose-600">{{ loadError.message }}</div>
  <div v-else-if="contract" class="flex flex-col gap-6">
    <div>
      <NuxtLink to="/me/contracts" class="text-sm">← контракты</NuxtLink>
      <div class="mt-2 flex items-start justify-between gap-4">
        <h1>Контракт · {{ money(contract.agreedAmountCents, contract.currency) }}</h1>
        <StatusPill :status="contract.status" />
      </div>
      <p class="mt-1 text-sm text-slate-500">
        {{ contract.type === 'fixed' ? 'фиксированная цена' : 'почасовой' }} ·
        начат {{ date(contract.startedAt) }}
      </p>
    </div>

    <div v-if="contract.status === 'disputed'" class="rounded-lg bg-rose-50 p-4 text-sm text-rose-800">
      По контракту открыт спор. Операции с этапами заморожены до решения арбитра.
    </div>

    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>

    <!-- milestones -->
    <section class="flex flex-col gap-3">
      <div class="flex items-center justify-between">
        <h2>Этапы</h2>
        <button
          v-if="isClient && contract.status === 'active' && contract.type === 'fixed'"
          class="btn-ghost btn-sm"
          @click="showAdd = !showAdd"
        >＋ Этап</button>
      </div>

      <form v-if="showAdd" class="card flex flex-col gap-3" @submit.prevent="addMilestone">
        <div class="field">
          <label class="label">Название этапа</label>
          <input v-model="ms.title" class="input" required />
        </div>
        <div class="field">
          <label class="label">Описание</label>
          <textarea v-model="ms.description" class="input" rows="2" />
        </div>
        <div class="flex gap-3">
          <div class="field w-40">
            <label class="label">Сумма, ₽</label>
            <input v-model="ms.amount" type="number" min="1" step="0.01" class="input" required />
          </div>
          <div class="field w-44">
            <label class="label">Срок</label>
            <input v-model="ms.dueDate" type="date" class="input" />
          </div>
        </div>
        <button class="btn-primary w-fit" :disabled="loading">Добавить</button>
      </form>

      <EmptyState v-if="!milestones.length" title="Этапов пока нет"
        :hint="isClient ? 'Добавьте первый этап, чтобы начать работу.' : 'Заказчик ещё не создал этапы.'" />

      <div v-for="m in milestones" :key="m.id" class="card flex flex-col gap-2">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="font-medium text-slate-900">{{ m.sequence }}. {{ m.title }}</p>
            <p v-if="m.description" class="text-sm text-slate-600">{{ m.description }}</p>
          </div>
          <div class="flex flex-col items-end gap-1">
            <b>{{ money(m.amountCents, contract.currency) }}</b>
            <StatusPill :status="m.status" />
          </div>
        </div>
        <p v-if="m.deliverableNote" class="rounded bg-slate-50 p-2 text-sm text-slate-700">
          Сдача: {{ m.deliverableNote }}
        </p>

        <div v-if="contract.status === 'active'" class="flex flex-wrap gap-2 pt-1">
          <button v-if="isClient && m.status === 'pending'" class="btn-accent btn-sm" @click="fund(m.id)">
            Оплатить в escrow
          </button>
          <button v-if="isClient && m.status === 'pending'" class="btn-ghost btn-sm" @click="milestoneAction(m.id, 'cancel')">
            Отменить
          </button>
          <button v-if="isFreelancer && m.status === 'funded'" class="btn-accent btn-sm" @click="submitWork(m.id)">
            Сдать работу
          </button>
          <button v-if="isClient && m.status === 'funded'" class="btn-ghost btn-sm" @click="milestoneAction(m.id, 'refund')">
            Вернуть средства
          </button>
          <template v-if="isClient && m.status === 'submitted'">
            <button class="btn-accent btn-sm" @click="milestoneAction(m.id, 'approve')">Принять и выплатить</button>
            <button class="btn-ghost btn-sm" @click="milestoneAction(m.id, 'request-changes')">На доработку</button>
          </template>
        </div>
      </div>
    </section>

    <!-- contract actions -->
    <div class="flex flex-wrap gap-3 border-t border-slate-200 pt-4">
      <button
        v-if="isClient && contract.status === 'active' && allResolved"
        class="btn-primary"
        @click="completeContract"
      >Завершить контракт</button>
      <button
        v-if="contract.status === 'active'"
        class="btn-ghost"
        @click="showDispute = !showDispute"
      >Открыть спор</button>
    </div>

    <form v-if="showDispute" class="card flex flex-col gap-3" @submit.prevent="raiseDispute">
      <h2>Спор по контракту</h2>
      <div class="field">
        <label class="label">Этап (необязательно)</label>
        <select v-model="disputeMilestone" class="input">
          <option value="">весь контракт</option>
          <option v-for="m in milestones" :key="m.id" :value="m.id">{{ m.sequence }}. {{ m.title }}</option>
        </select>
      </div>
      <div class="field">
        <label class="label">Суть спора</label>
        <textarea v-model="disputeReason" class="input" rows="3" required />
      </div>
      <button class="btn-danger w-fit" :disabled="loading">Открыть спор</button>
    </form>

    <!-- reviews -->
    <section v-if="contract.status === 'completed'" class="flex flex-col gap-3 border-t border-slate-200 pt-4">
      <h2>Отзывы</h2>
      <form v-if="!myReview" class="card flex flex-col gap-3" @submit.prevent="submitReview">
        <div class="field w-40">
          <label class="label">Оценка</label>
          <select v-model.number="rev.rating" class="input">
            <option v-for="n in 5" :key="n" :value="n">{{ n }} ★</option>
          </select>
        </div>
        <div class="field">
          <label class="label">Комментарий</label>
          <textarea v-model="rev.comment" class="input" rows="3" />
        </div>
        <button class="btn-primary w-fit" :disabled="loading">Оставить отзыв</button>
        <p class="text-xs text-slate-500">
          Отзыв станет виден после того, как обе стороны его оставят (или через 14 дней).
        </p>
      </form>

      <div v-for="r in reviews" :key="r.id" class="card flex flex-col gap-1">
        <div class="flex items-center justify-between">
          <RatingStars :value="r.rating" size="sm" />
          <span class="text-xs text-slate-400">
            {{ r.published ? 'опубликован' : 'ожидает второй стороны' }}
          </span>
        </div>
        <p v-if="r.comment" class="text-sm text-slate-700">{{ r.comment }}</p>
      </div>
    </section>
  </div>
</template>
