<template>
  <div class="my-projects">
    <div class="head">
      <h1>Мои проекты</h1>
      <NuxtLink class="btn new" to="/projects/new">＋ Новый проект</NuxtLink>
    </div>
    <p v-if="!items.length" class="hint">Проектов пока нет</p>
    <div v-else class="list">
      <NuxtLink
        v-for="p in items"
        :key="p.id"
        class="row"
        :to="p.status === 'draft' ? `/projects/${p.id}/edit` : `/projects/${p.id}`"
      >
        <div>
          <div class="title">{{ p.title }}</div>
          <div class="sub">{{ p.proposalsCount }} откликов · {{ money(p.fixedAmountCents, p.currency) }}</div>
        </div>
        <span class="status">{{ ru(p.status) }}</span>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()
const { data } = await useAsyncData('my-projects', () =>
  api.get('/me/projects').then((r) => r.data),
)
const items = computed(() => data.value?.projects || [])
</script>

<style scoped>
.my-projects { width: 100%; display: flex; flex-direction: column; gap: 20px; }
.head { display: flex; justify-content: space-between; align-items: center; }
.new {
  background: #4e92f8; color: #fff; padding: 10px 20px; border-radius: 10px;
  text-decoration: none; font-weight: 600;
}
.hint { color: #fff; font-weight: 500; }
.list { display: flex; flex-direction: column; gap: 12px; }
.row {
  background: #fff; color: #000; border-radius: 16px; padding: 18px 24px;
  display: flex; justify-content: space-between; align-items: center; text-decoration: none;
}
.title { font-weight: 700; }
.sub { font-size: 13px; font-weight: 500; color: #6b7280; margin-top: 4px; }
.status {
  background: #eef1f4; color: #374151; padding: 4px 12px; border-radius: 999px;
  font-size: 12px; font-weight: 600;
}
</style>
