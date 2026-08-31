<template>
  <div>
    <div>{{ notifications }}</div>
    <button @click="sendNotification">Отправить уведомление</button>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
// const sendNotification = async () => {
//   const response = await axios.post('http://localhost:5000/send-notify', {message: "asdasdsa"})
//   console.log(response.data)
//   const read = await axios.get('http://localhost:5000/read-notifications')
//   console.log(read.data)
// }

import {ref, onMounted} from 'vue';

const notifications = ref([]);
onMounted(async () => {
  // Замените URL на ваш сервер Golang + Fiber
  const response = await axios.get('http://localhost:5000/read-notifications');
  console.log(response.data)
  notifications.value = response.data;
});
</script>

<style scoped>
@import "./assets/css/style.css";
</style>