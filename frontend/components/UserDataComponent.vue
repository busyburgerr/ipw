<template>
  <div class="user-data_component">
    <div class="userData_header">
      <h1 class="fillData">Заполните данные</h1>
      <h3 class="fillData_second">Осталось совсем немного</h3>
    </div>
    <form class="userData_form">
      <label for="name">Введите имя</label>
      <input v-model="name" type="text" name="name" id="name" placeholder="прим. Иван Иванов">
      <label for="tag">Введите тэг</label>
      <input v-model="tag" type="text" name="tag" id="tag" placeholder="прим. @ipw">
      <label for="date">Введите дату рождения</label>
      <input v-model="birthday" type="text" name="date" id="date" placeholder="прим. 01.01.1987">
      <label for="location">Введите город</label>
      <input v-model="location" type="text" name="location" id="location" placeholder="прим. Уфа">
      <label for="gender">Выберите гендер</label>
      <select v-model="gender" name="gender" id="gender-select">
        <option value="Мужчина">Мужчина</option>
        <option value="Женщина">Женщина</option>
        <option value="Другое">Другое</option>
      </select>
      <label for="number">Введите номер телефона</label>
      <input v-model="number" type="tel" name="number" id="number" placeholder="прим. +7(999)999-99-99">
    </form>
    <div class="footer">
      <div class="btnForm">
        <button class="btn cancelBtn" @click="closeModal">Отмена</button>
        <button class="btn saveBtn" @click="saveBtn">Сохранить</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import {ref} from 'vue'


defineProps({
  closeModal: {
    type: Function,
    required: true
  }
})

const name = ref("")
const tag = ref("")
const birthday = ref("")
const location = ref("")
const gender = ref('')
const number = ref("")

const saveBtn = async () => {
  try {
    const response = await axios.put('http://localhost:5000/data/v1/user/update',{
      tag: tag.value,
      birthday: birthday.value,
      location: location.value,
      gender: gender.value,
      number: number.value,
    }, {withCredentials: true})
    console.log(response.data)
    window.location.reload()
  } catch (e) {
    console.log(e)
  }
}


</script>

<style scoped>
@import "assets/css/UserData.css";
</style>