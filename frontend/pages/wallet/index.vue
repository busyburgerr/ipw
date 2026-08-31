<template>
  <div class="wallet">
    <h1>Кошелёк</h1>

    <div class="cards">
      <div class="card">
        <span>Доступно к выводу</span>
        <b>{{ money(w?.availableCents, w?.currency) }}</b>
      </div>
      <div class="card">
        <span>В обработке</span>
        <b>{{ money(w?.pendingPayoutCents, w?.currency) }}</b>
      </div>
    </div>

    <form class="card form" @submit.prevent="request">
      <h2>Запросить вывод</h2>
      <div class="row">
        <input v-model="form.amount" type="number" min="1000" step="0.01" placeholder="Сумма, ₽" required />
        <select v-model="form.method">
          <option value="sbp">СБП</option>
          <option value="card">Карта</option>
          <option value="manual">Другое</option>
        </select>
        <input v-model="form.destination" placeholder="Телефон / номер карты" required />
      </div>
      <p v-if="error" class="err">{{ error }}</p>
      <button class="btn" :disabled="busy">Запросить</button>
      <span class="muted">Минимум 1000 ₽. Выплаты обрабатываются вручную.</span>
    </form>

    <div v-if="payouts.length" class="history">
      <h2>История</h2>
      <div v-for="p in payouts" :key="p.id" class="row-item">
        <div>
          <b>{{ money(p.amountCents, p.currency) }}</b>
          <span class="muted"> · {{ p.method }} · {{ p.destination }}</span>
          <p v-if="p.note" class="muted sm">{{ p.note }}</p>
        </div>
        <span class="status">{{ ru(p.status) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()

const { data: wallet, refresh: rw } = await useAsyncData('wallet', () =>
  api.get('/me/wallet').then((r) => r.data),
)
const { data: po, refresh: rp } = await useAsyncData('payouts', () =>
  api.get('/me/payouts').then((r) => r.data),
)
const w = computed(() => wallet.value)
const payouts = computed(() => po.value?.payouts || [])

const form = reactive({ amount: '', method: 'sbp', destination: '' })
const busy = ref(false)
const error = ref('')
async function request() {
  busy.value = true
  error.value = ''
  try {
    await api.post('/me/payouts', {
      amountCents: toCents(form.amount),
      method: form.method,
      destination: form.destination,
    })
    form.amount = ''
    form.destination = ''
    await Promise.all([rw(), rp()])
  } catch (e: any) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.wallet { width: 100%; display: flex; flex-direction: column; gap: 18px; }
.cards { display: flex; gap: 16px; flex-wrap: wrap; }
.card { background: #fff; color: #000; border-radius: 20px; padding: 22px; }
.cards .card { flex: 1; min-width: 200px; display: flex; flex-direction: column; gap: 6px; }
.cards .card b { font-size: 26px; }
.cards .card span { color: #6b7280; font-weight: 500; }
.form { display: flex; flex-direction: column; gap: 12px; }
.row { display: flex; gap: 12px; flex-wrap: wrap; }
.row > * { flex: 1; min-width: 140px; border: 1px solid #d1d5db; border-radius: 10px; padding: 10px; font-family: Inter, sans-serif; }
.btn { align-self: flex-start; background: #4e92f8; color: #fff; border: none; border-radius: 10px; padding: 10px 24px; font-weight: 600; cursor: pointer; }
.muted { color: #9ca3af; font-weight: 500; }
.muted.sm { font-size: 12px; }
.err { color: #b91c1c; font-weight: 600; }
.history { background: #fff; color: #000; border-radius: 20px; padding: 22px; display: flex; flex-direction: column; gap: 12px; }
.row-item { display: flex; justify-content: space-between; align-items: center; border-top: 1px solid #eef1f4; padding-top: 12px; }
.row-item:first-of-type { border-top: none; padding-top: 0; }
.status { background: #eef1f4; color: #374151; padding: 4px 12px; border-radius: 999px; font-size: 12px; font-weight: 600; }
</style>
