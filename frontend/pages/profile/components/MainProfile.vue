<template>
  <div class="main-profile">
    <button class="btn" v-if="userData.photo != null" @click="openUploadModal">
      <img class="user-photo" alt="user-photo" :src="`data:image/jpeg;base64,${userData.photo}`">
    </button>
    <button class="btn user-photo" @click="openUploadModal" v-else><img src="../../../assets/img/camera.svg"
                                                                        alt="add-photo"></button>
    <div class="user-dataBtn_block">
      <div class="user-info">
        <div class="user-nameEdit_block">
          <h2 class="user-name">{{ userData['name'] }}</h2>
          <button class="btn" @click="openModal"><img src="../../../assets/img/editBtnWhite.svg" alt="edit-btn"></button>
        </div>
        <div class="user-tagAgeGender_block">
          <div class="user-tag">{{ userData['tag'] }}</div>
          <div class="user-ageGender">{{ userData['gender'] }}, {{ Age }}</div>
        </div>
        <div class="user-locationStatus_block">
          <div class="user-location"><img src="../../../assets/img/location.svg" alt="location">{{ userData['location'] }}
          </div>
          <div class="user-status">{{ userData['status'] }}</div>
        </div>
        <div class="user-emailPhone_block">
          <div class="user-email">{{ userData['email'] }}</div>
          <div class="user-phone"></div>
        </div>
      </div>
      <div class="btn-block">
        <NuxtLink class="btn employee" to="/employer">Работодателям</NuxtLink>
        <button class="btn download-cv disable">Скачать CV</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const {userData, openUploadModal, openModal, Age} = defineProps(['userData', 'openUploadModal', 'openModal', 'Age'])
</script>

<style scoped>
@import '../assets/css/profile.css';

h1, h2, h3, h4, p {
  margin: 0;
  padding: 0;
}

hr {
  margin: 0;
  padding: 0;
}

textarea {
  width: 100%;
  border: none;
}
</style>