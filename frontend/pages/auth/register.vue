<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const { loading, error, run } = useSubmit()

const form = reactive({
  displayName: '',
  email: '',
  password: '',
  asFreelancer: true,
  asClient: false,
})

async function submit() {
  const ok = await run(() => auth.register({ ...form }))
  if (ok !== undefined) await navigateTo((route.query.next as string) || '/dashboard')
}
</script>

<template>
  <AuthCard mode="register">
    <form class="flex flex-col gap-3" @submit.prevent="submit">
      <input v-model="form.displayName" class="input" placeholder="Имя" required />
      <input v-model="form.email" type="email" class="input" placeholder="Email" required />
      <input v-model="form.password" type="password" class="input" placeholder="Пароль (мин. 8)" minlength="8" required />
      <div class="mt-1 flex flex-col gap-2 rounded-xl bg-slate-50 p-3 text-sm text-slate-600">
        <span class="font-medium">Я хочу</span>
        <label class="flex items-center gap-2">
          <input v-model="form.asFreelancer" type="checkbox" /> выполнять заказы
        </label>
        <label class="flex items-center gap-2">
          <input v-model="form.asClient" type="checkbox" /> размещать заказы
        </label>
      </div>
      <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
      <button class="btn-accent mt-1" :disabled="loading || (!form.asFreelancer && !form.asClient)">
        {{ loading ? 'Создаём…' : 'Создать аккаунт' }}
      </button>
    </form>
    <p class="mt-4 text-sm text-slate-500">
      Уже есть аккаунт? <NuxtLink to="/auth/login">Войти</NuxtLink>
    </p>
  </AuthCard>
</template>
