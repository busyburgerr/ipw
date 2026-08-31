<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const { $api } = useNuxtApp()
const { loading, error, run } = useSubmit()

async function create(payload: any) {
  const project = await run(() => $api<any>('/projects', { method: 'POST', body: payload }))
  if (project) await navigateTo(`/projects/${project.id}/edit?created=1`)
}
</script>

<template>
  <div class="mx-auto max-w-2xl">
    <h1 class="mb-6">Новый проект</h1>
    <ProjectForm
      submit-label="Создать черновик"
      :loading="loading"
      :error="error"
      @submit="create"
    />
    <p class="mt-3 text-sm text-slate-500">
      Проект создаётся черновиком. После проверки опубликуйте его, чтобы получать отклики.
    </p>
  </div>
</template>
