<template>
  <div class="container">
    <form class="authForm" v-if="isLogin" @submit.prevent="loginHandler">
      <h1>Авторизация <span>/</span> Регистрация</h1>
      <input type="email" v-model="email" placeholder="Почта" required>
      <input type="password" v-model="password" placeholder="Пароль" required>
      <button class="btn" type="submit">Войти</button>
      <p>Нет аккаунта?
        <button class="btn" @click="switchForm">Зарегистрироваться</button>
      </p>
    </form>

    <!-- Форма регистрации -->
    <form class="authForm" v-else @submit.prevent="registerHandler">
      <h1>Авторизация <span>/</span> Регистрация</h1>
      <input type="text" v-model="name" placeholder="Имя / фамилия" required>
      <input type="email" v-model="email" placeholder="E-mail" required>
      <input type="password" v-model="password" placeholder="Пароль" required>
      <input type="password" v-model="repeatPassword" placeholder="Повторите пароль" required>
      <button class="btn" type="submit">Зарегистрироваться</button>
      <p>
        Уже есть аккаунт?
        <button class="btn" @click="switchForm">Войти</button>
      </p>
    </form>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'

const isLogin = ref(true)
const login = ref('')
const name = ref('')
const email = ref('')
const password = ref('')
const repeatPassword = ref('')
const router = useRouter()

const loginHandler = async () => {
  try {
    const response = await axios.post(
        `http://localhost:5000/auth/v1/login`,
        {
          email: email.value,
          password: password.value
        },
        {
          withCredentials: true
        }
    )
    window.location.reload()
    console.log(response)
  } catch (e) {
    console.log(e)
  }
}
const registerHandler = async () => {
  try {
    const response = await axios.post(
        `http://localhost:5000/auth/v1/register`,
        {
          name: name.value,
          email: email.value,
          password: password.value,
          repeatPassword: password.value
        }, {withCredentials: false}
    )
    window.location.reload()
    console.log(response)
  } catch (e) {
    console.log(e)
  }
}
const switchForm = () => {
  isLogin.value = !isLogin.value
}
onMounted(() => {
  const isAuth = Cookies.get('ipw')
  if (isAuth) {
    router.push('/profile')
  }
})
</script>

<style scoped>
* {
  color: #000;
//text-align: center; font-family: Inter, sans-serif;
  font-style: normal;
  line-height: normal;
}
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 250px);
  width: 100%;
}

h1 {
  margin: 0;
  padding: 0;
  font-size: 24px;
  font-weight: 500;
}

.authForm {
  display: flex;
  padding: 100px;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 20px;
  border-radius: 30px;
  background: #FFF;
}

input {
  display: flex;
  width: 320px;
  padding: 15px 10px;
  align-items: center;
  border-radius: 15px;
  border: .8px solid rgba(0, 0, 0, 0.60);
  background: rgba(230, 230, 227, 0.40);
}

input::placeholder {
  color: rgba(0, 0, 0, 0.50);
  font-size: 12px;
  font-weight: 500;
}
</style>