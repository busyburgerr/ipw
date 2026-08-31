<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const { loading, error, run } = useSubmit()

const form = reactive({ email: '', password: '' })

async function submit() {
  const ok = await run(() => auth.login(form.email, form.password))
  if (ok !== undefined) await navigateTo((route.query.next as string) || '/dashboard')
}
</script>

<template>
  <div class="mx-auto max-w-sm">
    <h1 class="mb-6">Вход</h1>
    <form class="flex flex-col gap-4" @submit.prevent="submit">
      <div class="field">
        <label class="label">Email</label>
        <input v-model="form.email" type="email" class="input" required />
      </div>
      <div class="field">
        <label class="label">Пароль</label>
        <input v-model="form.password" type="password" class="input" required />
      </div>
      <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
      <button class="btn-primary" :disabled="loading">{{ loading ? 'Входим…' : 'Войти' }}</button>
    </form>
    <p class="mt-4 text-sm text-slate-500">
      Нет аккаунта? <NuxtLink to="/auth/register">Зарегистрироваться</NuxtLink>
    </p>
  </div>
</template>
