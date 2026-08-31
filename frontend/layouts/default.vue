<template>
  <header class="header">
    <div class="header-logoNotify_block">
      <div class="header_logo">
        <NuxtLink to="/">IPW</NuxtLink>
      </div>
      <NuxtLink class="header_requests" to="/requests" v-if="user?.isClient">Отклики</NuxtLink>
    </div>
    <nav class="navbar">
      <NuxtLink to="/projects">Проекты</NuxtLink>
      <NuxtLink to="/freelancers">Фрилансеры</NuxtLink>
      <template v-if="isAuthed">
        <NuxtLink to="/profile">Профиль</NuxtLink>
        <NuxtLink to="/wallet" v-if="user?.isFreelancer">Кошелёк</NuxtLink>
        <NuxtLink to="/admin" v-if="user?.isAdmin">Арбитраж</NuxtLink>
        <button class="btn" @click="onLogout">Выйти</button>
      </template>
      <NuxtLink to="/auth" v-else>Войти</NuxtLink>
    </nav>
  </header>

  <slot />

  <footer>
    <NuxtLink class="police-privacy" to="/policy">Политика конфиденциальности</NuxtLink>
  </footer>
</template>

<script setup lang="ts">
const { user, isAuthed, init, logout } = useAuth()
await init()

async function onLogout() {
  await logout()
  await navigateTo('/')
}
</script>

<style scoped>
.header-logoNotify_block {
  display: flex;
  align-items: center;
  gap: 10px;
}

footer {
  width: 100%;
  padding: 20px;
}

.police-privacy {
  font-family: Inter, sans-serif;
  font-weight: 500;
  font-size: 12px;
}

.header_requests {
  color: #fff;
  font-size: 16px;
  font-weight: 500;
}

.navbar .btn {
  font-family: Raleway, sans-serif;
  font-size: 18px;
  font-weight: 700;
}
</style>
