<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const route = useRoute()
const { $api } = useNuxtApp()
const id = route.params.id as string
const { loading, error, run } = useSubmit()

const { data: project } = await useAsyncData(`project-edit-${id}`, () => $api<any>(`/projects/${id}`))

async function save(payload: any) {
  const updated = await run(() => $api<any>(`/projects/${id}`, { method: 'PUT', body: payload }))
  if (updated) await navigateTo(`/projects/${id}`)
}

async function publish() {
  await run(() => $api(`/projects/${id}/publish`, { method: 'POST' }))
  await navigateTo(`/projects/${id}`)
}
</script>

<template>
  <div v-if="project" class="mx-auto max-w-2xl">
    <div class="mb-6 flex items-center justify-between">
      <h1>Редактирование проекта</h1>
      <StatusPill :status="project.status" />
    </div>

    <div v-if="route.query.created" class="mb-4 rounded-lg bg-emerald-50 p-3 text-sm text-emerald-800">
      Черновик создан. Проверьте детали и опубликуйте.
    </div>

    <ProjectForm
      :initial="project"
      submit-label="Сохранить"
      :loading="loading"
      :error="error"
      @submit="save"
    />

    <div v-if="project.status === 'draft'" class="mt-6 flex items-center gap-3 border-t border-slate-200 pt-6">
      <button class="btn-primary" :disabled="loading" @click="publish">Опубликовать</button>
      <span class="text-sm text-slate-500">После публикации проект увидят фрилансеры.</span>
    </div>
  </div>
</template>
