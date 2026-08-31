<template>
  <div class="auth-wrap">
    <form class="authForm" v-if="isLogin" @submit.prevent="submitLogin">
      <h1>Авторизация <span>/</span> Регистрация</h1>
      <input type="email" v-model="email" placeholder="Почта" required />
      <input type="password" v-model="password" placeholder="Пароль" required />
      <p v-if="error" class="err">{{ error }}</p>
      <button class="btn primary" type="submit" :disabled="loading">Войти</button>
      <p>Нет аккаунта?
        <button class="btn link" type="button" @click="isLogin = false">Зарегистрироваться</button>
      </p>
    </form>

    <form class="authForm" v-else @submit.prevent="submitRegister">
      <h1>Авторизация <span>/</span> Регистрация</h1>
      <input type="text" v-model="name" placeholder="Имя / фамилия" required />
      <input type="email" v-model="email" placeholder="E-mail" required />
      <input type="password" v-model="password" placeholder="Пароль (мин. 8)" minlength="8" required />
      <div class="roles">
        <label><input type="checkbox" v-model="asFreelancer" /> выполнять заказы</label>
        <label><input type="checkbox" v-model="asClient" /> размещать заказы</label>
      </div>
      <p v-if="error" class="err">{{ error }}</p>
      <button class="btn primary" type="submit" :disabled="loading || (!asFreelancer && !asClient)">
        Зарегистрироваться
      </button>
      <p>Уже есть аккаунт?
        <button class="btn link" type="button" @click="isLogin = true">Войти</button>
      </p>
    </form>
  </div>
</template>

<script setup lang="ts">
const { login, register, isAuthed, init } = useAuth()
const route = useRoute()

await init()
if (isAuthed.value) await navigateTo('/profile')

const isLogin = ref(true)
const name = ref('')
const email = ref('')
const password = ref('')
const asFreelancer = ref(true)
const asClient = ref(false)
const error = ref('')
const loading = ref(false)

const dest = () => (route.query.next as string) || '/profile'

async function submitLogin() {
  loading.value = true
  error.value = ''
  try {
    await login(email.value, password.value)
    await navigateTo(dest())
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
async function submitRegister() {
  loading.value = true
  error.value = ''
  try {
    await register({
      displayName: name.value,
      email: email.value,
      password: password.value,
      asFreelancer: asFreelancer.value,
      asClient: asClient.value,
    })
    await navigateTo(dest())
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
* { color: #000; font-style: normal; line-height: normal; }
.auth-wrap {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 250px);
  width: 100%;
}
h1 { margin: 0; font-size: 24px; font-weight: 500; }
.authForm {
  display: flex;
  padding: 70px;
  flex-direction: column;
  align-items: center;
  gap: 18px;
  border-radius: 30px;
  background: #fff;
  width: 340px;
}
input[type='email'],
input[type='password'],
input[type='text'] {
  width: 100%;
  padding: 15px 10px;
  border-radius: 15px;
  border: 0.8px solid rgba(0, 0, 0, 0.6);
  background: rgba(230, 230, 227, 0.4);
}
input::placeholder { color: rgba(0, 0, 0, 0.5); font-size: 12px; font-weight: 500; }
.roles { display: flex; flex-direction: column; gap: 8px; align-self: stretch; font-size: 14px; font-weight: 500; }
.primary {
  width: 100%;
  padding: 14px;
  border-radius: 12px;
  background: #4e92f8;
  color: #fff;
  font-weight: 600;
  border: none;
  cursor: pointer;
}
.link { color: #4e92f8; font-weight: 600; text-decoration: underline; padding: 0; }
.err { color: #b91c1c; font-size: 13px; font-weight: 500; align-self: stretch; }
</style>
