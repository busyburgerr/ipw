<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const { $api } = useNuxtApp()
const { data } = await useAsyncData('my-projects', () => $api<{ projects: any[] }>('/me/projects'))
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1>Мои проекты</h1>
      <NuxtLink to="/projects/new" class="btn-primary btn-sm no-underline">＋ Новый</NuxtLink>
    </div>
    <EmptyState v-if="!data?.projects?.length" title="Проектов пока нет">
      <NuxtLink to="/projects/new" class="btn-ghost btn-sm no-underline">Разместить проект</NuxtLink>
    </EmptyState>
    <div v-else class="flex flex-col gap-3">
      <NuxtLink
        v-for="p in data.projects"
        :key="p.id"
        :to="p.status === 'draft' ? `/projects/${p.id}/edit` : `/projects/${p.id}`"
        class="card flex items-center justify-between no-underline hover:border-brand"
      >
        <div>
          <p class="font-medium text-slate-900">{{ p.title }}</p>
          <p class="text-sm text-slate-500">{{ p.proposalsCount }} откликов · {{ relTime(p.createdAt) }}</p>
        </div>
        <StatusPill :status="p.status" />
      </NuxtLink>
    </div>
  </div>
</template>
