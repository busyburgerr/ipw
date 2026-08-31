<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
const auth = useAuthStore()
</script>

<template>
  <section class="flex flex-col items-center gap-5 py-20 text-center">
    <p class="text-sm font-medium uppercase tracking-widest text-white/70">
      Фриланс по безопасной сделке
    </p>
    <h1 class="max-w-3xl text-4xl font-extrabold leading-tight text-white sm:text-5xl">
      Находите специалистов и работайте через escrow
    </h1>
    <p class="max-w-xl text-white/80">
      Заказчик размещает проект, фрилансер отправляет отклик. Деньги за этап
      резервируются и выплачиваются только после приёмки работы.
    </p>
    <div class="mt-2 flex flex-wrap justify-center gap-3">
      <NuxtLink to="/projects" class="btn-primary no-underline">Смотреть проекты</NuxtLink>
      <NuxtLink
        :to="auth.isAuthed ? '/dashboard' : '/auth/register'"
        class="btn-ghost no-underline"
      >
        {{ auth.isAuthed ? 'В кабинет' : 'Создать аккаунт' }}
      </NuxtLink>
    </div>

    <div class="mt-12 grid w-full max-w-3xl gap-4 text-left sm:grid-cols-3">
      <div class="card">
        <p class="text-2xl font-extrabold text-brand">1</p>
        <p class="mt-1 font-semibold text-ink">Проект и отклики</p>
        <p class="mt-1 text-sm text-slate-500">Опишите задачу и бюджет, соберите предложения.</p>
      </div>
      <div class="card">
        <p class="text-2xl font-extrabold text-brand">2</p>
        <p class="mt-1 font-semibold text-ink">Контракт и этапы</p>
        <p class="mt-1 text-sm text-slate-500">Примите оффер, разбейте работу на этапы с escrow.</p>
      </div>
      <div class="card">
        <p class="text-2xl font-extrabold text-brand">3</p>
        <p class="mt-1 font-semibold text-ink">Приёмка и выплата</p>
        <p class="mt-1 text-sm text-slate-500">Проверьте результат — деньги уходят фрилансеру.</p>
      </div>
    </div>
  </section>
</template>
