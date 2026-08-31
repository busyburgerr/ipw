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
  <AuthCard mode="login">
    <form class="flex flex-col gap-3" @submit.prevent="submit">
      <input v-model="form.email" type="email" class="input" placeholder="Email" required />
      <input v-model="form.password" type="password" class="input" placeholder="Пароль" required />
      <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
      <button class="btn-accent mt-1" :disabled="loading">{{ loading ? 'Входим…' : 'Войти' }}</button>
    </form>
    <p class="mt-4 text-sm text-slate-500">
      Нет аккаунта? <NuxtLink to="/auth/register">Зарегистрироваться</NuxtLink>
    </p>
  </AuthCard>
</template>
