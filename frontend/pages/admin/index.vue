<template>
  <div class="admin">
    <h1>Арбитраж — открытые споры</h1>
    <p v-if="error" class="err">{{ error }}</p>
    <p v-if="!items.length" class="hint">Открытых споров нет</p>
    <div v-for="d in items" :key="d.id" class="card">
      <div class="top">
        <NuxtLink :to="`/contracts/${d.contractId}`" class="link">Контракт →</NuxtLink>
        <span class="status">{{ ru(d.status) }}</span>
      </div>
      <p class="reason">{{ d.reason }}</p>
      <p class="muted">этап: {{ d.milestoneId || 'весь контракт' }} · {{ fmtDate(d.createdAt) }}</p>
      <div class="btns">
        <button v-if="d.status === 'open'" class="btn ghost" @click="claim(d.id)">Взять в работу</button>
        <button class="btn blue" @click="resolve(d.id, 'freelancer')">В пользу фрилансера</button>
        <button class="btn danger" @click="resolve(d.id, 'client')">В пользу заказчика</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'admin' })
const api = useApi()
const { data, refresh } = await useAsyncData('admin-disputes', () =>
  api.get('/admin/disputes').then((r) => r.data),
)
const items = computed(() => data.value?.disputes || [])
const error = ref('')

async function claim(id: string) {
  try {
    await api.post(`/admin/disputes/${id}/claim`)
    await refresh()
  } catch (e: any) {
    error.value = e.message
  }
}
async function resolve(id: string, outcome: 'client' | 'freelancer') {
  const note = prompt('Комментарий к решению:') ?? ''
  try {
    await api.post(`/admin/disputes/${id}/resolve`, { outcome, note })
    await refresh()
  } catch (e: any) {
    error.value = e.message
  }
}
</script>

<style scoped>
.admin { width: 100%; display: flex; flex-direction: column; gap: 16px; }
.hint { color: #fff; font-weight: 500; }
.err { color: #fecaca; font-weight: 600; }
.card { background: #fff; color: #000; border-radius: 20px; padding: 22px; display: flex; flex-direction: column; gap: 8px; }
.top { display: flex; justify-content: space-between; align-items: center; }
.link { color: #4e92f8; font-weight: 600; text-decoration: none; }
.reason { font-weight: 500; }
.muted { color: #9ca3af; font-weight: 500; font-size: 13px; }
.btns { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 6px; }
.btn { border: none; border-radius: 10px; padding: 9px 18px; font-weight: 600; cursor: pointer; font-size: 14px; }
.btn.blue { background: #4e92f8; color: #fff; }
.btn.ghost { background: #eef1f4; color: #111827; }
.btn.danger { background: #dc2626; color: #fff; }
.status { background: #eef1f4; color: #374151; padding: 4px 12px; border-radius: 999px; font-size: 12px; font-weight: 600; }
</style>
