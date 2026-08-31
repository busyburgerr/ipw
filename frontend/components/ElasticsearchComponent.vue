<template>
  <div class="container">
    <input class="searchbar" type="search" placeholder="Поиск" v-model="query" @input="checkAndRedirect"/>
    <div class="nav_list" v-for="user in users" :key="user.id">
      <!-- Вывод результатов поиска -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import axios from "axios";

const query = ref('')
const users = ref([])

const checkAndRedirect = () => {
    if (query.value.includes("print('Hello World')")) {
    window.alert("Пока сайт нет, но есть это уведомление)")
  } else {
    searchUser()
  }
}
const searchUser = async () => {
  const apiUrl = `http://localhost:5000/search/v1/search/user/${query.value}`;
  try {
    if (query.value === "") {
      return;
    }
    await axios.get(apiUrl)
        .then(response => {
          users.value = response.data;
          console.log(response.data)
        })
        .catch(error => {
          console.error(error);
        });
  } catch (e) {
    console.log(e)
  }
};
</script>

<style scoped>
.container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  align-self: stretch;
}
.searchbar {
  display: flex;
  width: 200px;
  padding: 10px 15px;
  justify-content: space-between;
  align-items: center;
  border-radius: 10px;
  background: #FFF;
  outline: none;
  border: none;
}
.searchbar::placeholder {
  color: rgba(0, 0, 0, 0.50);
  font-family: Inter, sans-serif;
  font-size: 12px;
  font-style: normal;
  font-weight: 500;
  line-height: normal;
}
</style>