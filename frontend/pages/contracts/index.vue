<template>
  <div class="contracts">
    <h1>Контракты</h1>
    <p v-if="!items.length" class="hint">Контрактов пока нет</p>
    <div v-else class="list">
      <NuxtLink v-for="c in items" :key="c.id" :to="`/contracts/${c.id}`" class="row">
        <div>
          <div class="amt">{{ money(c.agreedAmountCents, c.currency) }}</div>
          <div class="sub">{{ ru(c.type) }} · с {{ fmtDate(c.startedAt) }}</div>
        </div>
        <span class="status">{{ ru(c.status) }}</span>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()
const { data } = await useAsyncData('contracts', () => api.get('/me/contracts').then((r) => r.data))
const items = computed(() => data.value?.contracts || [])
</script>

<style scoped>
.contracts { width: 100%; display: flex; flex-direction: column; gap: 18px; }
.hint { color: #fff; font-weight: 500; }
.list { display: flex; flex-direction: column; gap: 12px; }
.row {
  background: #fff; color: #000; border-radius: 16px; padding: 18px 24px;
  display: flex; justify-content: space-between; align-items: center; text-decoration: none;
}
.amt { font-weight: 700; }
.sub { font-size: 13px; font-weight: 500; color: #6b7280; margin-top: 4px; }
.status { background: #eef1f4; color: #374151; padding: 4px 12px; border-radius: 999px; font-size: 12px; font-weight: 600; }
</style>
