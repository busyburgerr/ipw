<template>
  <div class="my-proposals">
    <h1>Мои отклики</h1>
    <p v-if="err" class="err">{{ err }}</p>
    <p v-if="!items.length" class="hint">Вы ещё не откликались</p>
    <div v-else class="list">
      <div v-for="p in items" :key="p.id" class="row">
        <div class="row-main">
          <NuxtLink :to="`/projects/${p.projectId}`" class="link">Проект →</NuxtLink>
          <div class="amt">{{ money(p.bidAmountCents) }}
            <span v-if="p.estimatedDays" class="sub"> · ~{{ p.estimatedDays }} дн.</span>
          </div>
          <p class="letter">{{ p.coverLetter }}</p>
        </div>
        <div class="row-side">
          <span class="status">{{ ru(p.status) }}</span>
          <button
            v-if="['pending', 'shortlisted'].includes(p.status)"
            class="btn withdraw"
            @click="withdraw(p.id)"
          >Отозвать</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()
const { data, refresh } = await useAsyncData('my-proposals', () =>
  api.get('/me/proposals').then((r) => r.data),
)
const items = computed(() => data.value?.proposals || [])
const err = ref('')

async function withdraw(id: string) {
  try {
    await api.post(`/proposals/${id}/withdraw`)
    await refresh()
  } catch (e: any) {
    err.value = e.message
  }
}
</script>

<style scoped>
.my-proposals { width: 100%; display: flex; flex-direction: column; gap: 18px; }
.hint { color: #fff; font-weight: 500; }
.err { color: #fecaca; font-weight: 600; }
.list { display: flex; flex-direction: column; gap: 12px; }
.row {
  background: #fff; color: #000; border-radius: 16px; padding: 18px 22px;
  display: flex; justify-content: space-between; gap: 20px;
}
.link { color: #4e92f8; font-weight: 600; text-decoration: none; }
.amt { font-weight: 700; margin-top: 6px; }
.sub { font-weight: 500; color: #6b7280; }
.letter { font-size: 13px; font-weight: 500; color: #374151; margin-top: 6px; }
.row-side { display: flex; flex-direction: column; align-items: flex-end; gap: 10px; }
.status { background: #eef1f4; color: #374151; padding: 4px 12px; border-radius: 999px; font-size: 12px; font-weight: 600; }
.withdraw { background: #eef1f4; border: none; border-radius: 8px; padding: 6px 14px; font-weight: 600; cursor: pointer; }
</style>
