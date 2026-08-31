<template>
  <header class="header">
    <div class="header-logoNotify_block">
      <div class="header_logo">
        <NuxtLink to="/">IPW</NuxtLink>
      </div>
      <NuxtLink class="header_requests" to="/requests" v-if="usrRole == 'HR'">Отклики</NuxtLink>
      <!--      <button class="header_notification" @click="openModal">notify</button>-->
    </div>
    <nav class="navbar">
<!--      <div v-if="route.path != '/' && route.path != '/profile' && route.path != '/auth'">-->
<!--        <NuxtLink class="btn" to="/profile">Профиль</NuxtLink>-->
<!--      </div>-->
      <!--      <NuxtLink to="/messenger">Чат</NuxtLink>-->
      <NuxtLink to="/resumes">Резюме</NuxtLink>
      <NuxtLink to="/vacancies">Вакансии</NuxtLink>
      <NuxtLink :class="{'disable':isActive}" v-if="isActive" :to="isActive ? '/' : '/about'">О нас</NuxtLink>
    </nav>
  </header>
  <div class="modal" v-if="showModal">
    <NotifyModalComponent :closeModal="closeModal"/>
  </div>
  <slot/>
  <footer>
    <NuxtLink class="police-privacy" to="/policy">Политика конфиденциальности</NuxtLink>
  </footer>
</template>

<script setup lang="ts">
import {defineComponent, ref} from "vue";
import NotifyModalComponent from "/components/NotifyModalComponent.vue"
import ResumeComponent from "~/components/ResumeComponent.vue";
import VacancyComponent from "~/pages/vacancies/components/VacancyComponent.vue";
import axios from "axios";

const usrRole = ref('')

onMounted(async () => {
  try {
    const res = await axios.get('http://localhost:5000/data/v1/user', {withCredentials: true})
    usrRole.value = res.data.role.role_name
  } catch (e) {
    console.log(e)
  }
})

defineComponent({
  components: {
    NotifyModalComponent,
  },
})
let isActive = ref(true)
const route = useRoute()
const router = useRouter()
// onMounted(() => {
//   if (route.path == '/messenger') {
//     router.push('/')
//   }
// })
let showModal = ref(false)
const openModal = () => {
  showModal.value = true
}
const closeModal = () => {
  showModal.value = false
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
  color: #FFF;
  font-size: 16px;
  font-style: normal;
  font-weight: 500;
  line-height: normal;
}
</style>