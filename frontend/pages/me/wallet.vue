<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const { $api } = useNuxtApp()

const { data: wallet, refresh: refreshWallet } = await useAsyncData('wallet', () => $api<any>('/me/wallet'))
const { data: payouts, refresh: refreshPayouts } = await useAsyncData('payouts', () =>
  $api<{ payouts: any[] }>('/me/payouts'),
)

const { loading, error, run } = useSubmit()
const form = reactive({ amount: '', method: 'sbp', destination: '' })

async function request() {
  const r = await run(() =>
    $api('/me/payouts', {
      method: 'POST',
      body: { amountCents: toCents(form.amount), method: form.method, destination: form.destination },
    }),
  )
  if (r !== undefined) {
    form.amount = ''
    form.destination = ''
    await Promise.all([refreshWallet(), refreshPayouts()])
  }
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <h1>Кошелёк</h1>

    <div class="grid gap-4 sm:grid-cols-2">
      <div class="card">
        <p class="text-sm text-slate-500">Доступно к выводу</p>
        <p class="mt-1 text-2xl font-bold">{{ money(wallet?.availableCents, wallet?.currency) }}</p>
      </div>
      <div class="card">
        <p class="text-sm text-slate-500">В обработке (выплаты)</p>
        <p class="mt-1 text-2xl font-bold">{{ money(wallet?.pendingPayoutCents, wallet?.currency) }}</p>
      </div>
    </div>

    <form class="card flex flex-col gap-3" @submit.prevent="request">
      <h2>Запросить вывод</h2>
      <div class="flex flex-wrap gap-3">
        <div class="field w-40">
          <label class="label">Сумма, ₽</label>
          <input v-model="form.amount" type="number" min="1000" step="0.01" class="input" required />
        </div>
        <div class="field w-40">
          <label class="label">Способ</label>
          <select v-model="form.method" class="input">
            <option value="sbp">СБП</option>
            <option value="card">Карта</option>
            <option value="manual">Другое</option>
          </select>
        </div>
        <div class="field flex-1">
          <label class="label">Реквизиты</label>
          <input v-model="form.destination" class="input" placeholder="номер телефона / карты" required />
        </div>
      </div>
      <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
      <button class="btn-primary w-fit" :disabled="loading">Запросить</button>
      <p class="text-xs text-slate-500">Минимальная сумма — 1000 ₽. Выплаты обрабатываются вручную.</p>
    </form>

    <section v-if="payouts?.payouts?.length" class="flex flex-col gap-2">
      <h2>История выплат</h2>
      <div v-for="p in payouts.payouts" :key="p.id" class="card flex items-center justify-between">
        <div class="text-sm">
          <b>{{ money(p.amountCents, p.currency) }}</b>
          <span class="text-slate-500"> · {{ p.method }} · {{ p.destination }}</span>
          <p v-if="p.note" class="text-xs text-slate-400">{{ p.note }}</p>
        </div>
        <StatusPill :status="p.status" />
      </div>
    </section>
  </div>
</template>
