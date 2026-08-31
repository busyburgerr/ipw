<template>
  <div class="vacancies_item_block">
    <div class="vacancies_item_header">
      <img class="vacancies_item_logo" :src="`data:image/jpeg;base64,${vacancy.companyPhoto}`" alt="company-photo">
      <div class="vacancies_item_vacanciesInfo">
        <div class="vacancies_item_vacanciesInfo_header">
          <NuxtLink class="vacancies_item_name" :to="`/company/${vacancy.companyID}`">{{ vacancy['companyName'] }}</NuxtLink>
          <div class="vacancies_item_tag">{{ vacancy['companyTag'] }}</div>
        </div>
        <NuxtLink :to="`/vacancies/${vacancy.id}`" class="vacancies_item_levelDirection">{{ vacancy['level'] }} {{ vacancy['direction'] }}</NuxtLink>
        <div class="vacancies_item_location">{{ vacancy['location'] }} / {{ vacancy['workTime'] }}</div>
        <!-- <div class="vacancies_item_levelDirection"></div> ?? >> email -->
      </div>
    </div>
    <div class="vacancies_item_description">{{ vacancy['description'] }}</div>
    <div class="vacancies_item_skills">{{ vacancy['skills'] }}</div>
    <div class="vacancies_item_btnBlock">
      <button class="btn send">Откликнуться</button>
      <button class="btn success">Вы уже откликнулись</button>
    </div>
  </div>
</template>

<script setup lang="ts">
const { vacancy } = defineProps(['vacancy'])
</script>

<style scoped>
@import '../assets/css/vacancy-component.css';
</style>