<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
const auth = useAuthStore()
</script>

<template>
  <section class="flex flex-col items-center gap-6 py-12 text-center">
    <h1 class="max-w-2xl text-4xl font-bold tracking-tight text-ink">
      Находите специалистов и работайте по безопасной сделке
    </h1>
    <p class="max-w-xl text-slate-600">
      Заказчик размещает проект, фрилансер отправляет отклик. Деньги за этап
      резервируются в escrow и выплачиваются после приёмки работы.
    </p>
    <div class="flex gap-3">
      <NuxtLink to="/projects" class="btn-primary no-underline">Смотреть проекты</NuxtLink>
      <NuxtLink v-if="!auth.isAuthed" to="/auth/register" class="btn-ghost no-underline">
        Создать аккаунт
      </NuxtLink>
      <NuxtLink v-else to="/dashboard" class="btn-ghost no-underline">В кабинет</NuxtLink>
    </div>

    <div class="mt-8 grid w-full max-w-3xl gap-4 text-left sm:grid-cols-3">
      <div class="card">
        <p class="font-semibold">1. Проект и отклики</p>
        <p class="mt-1 text-sm text-slate-600">Опишите задачу и бюджет, соберите предложения.</p>
      </div>
      <div class="card">
        <p class="font-semibold">2. Контракт и этапы</p>
        <p class="mt-1 text-sm text-slate-600">Примите оффер, разбейте работу на этапы с escrow.</p>
      </div>
      <div class="card">
        <p class="font-semibold">3. Приёмка и выплата</p>
        <p class="mt-1 text-sm text-slate-600">Проверьте результат — деньги уходят фрилансеру.</p>
      </div>
    </div>
  </section>
</template>
