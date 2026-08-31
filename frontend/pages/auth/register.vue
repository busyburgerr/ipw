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
  <div class="mx-auto max-w-sm">
    <h1 class="mb-6">Регистрация</h1>
    <form class="flex flex-col gap-4" @submit.prevent="submit">
      <div class="field">
        <label class="label">Имя</label>
        <input v-model="form.displayName" class="input" required />
      </div>
      <div class="field">
        <label class="label">Email</label>
        <input v-model="form.email" type="email" class="input" required />
      </div>
      <div class="field">
        <label class="label">Пароль</label>
        <input v-model="form.password" type="password" class="input" minlength="8" required />
      </div>
      <fieldset class="flex flex-col gap-2">
        <span class="label">Я хочу</span>
        <label class="flex items-center gap-2 text-sm">
          <input v-model="form.asFreelancer" type="checkbox" /> выполнять заказы (фрилансер)
        </label>
        <label class="flex items-center gap-2 text-sm">
          <input v-model="form.asClient" type="checkbox" /> размещать заказы (заказчик)
        </label>
      </fieldset>
      <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>
      <button class="btn-primary" :disabled="loading || (!form.asFreelancer && !form.asClient)">
        {{ loading ? 'Создаём…' : 'Создать аккаунт' }}
      </button>
    </form>
    <p class="mt-4 text-sm text-slate-500">
      Уже есть аккаунт? <NuxtLink to="/auth/login">Войти</NuxtLink>
    </p>
  </div>
</template>
