<template>
  <div class="container-head">
    <div class="main-profile">
      <button class="btn user-photo" style="background: none" v-if="userData.photo != null" @click="openUploadModal">
        <img class="user-photo" alt="user-photo" :src="`data:image/jpeg;base64,${userData.photo}`">
      </button>
      <button class="btn user-photo" @click="openUploadModal" v-else><img src="../assets/img/camera.svg" alt="add-photo"></button>
      <div class="user-dataBtn_block">
        <div class="user-info">
          <div class="user-nameEdit_block">
            <h2 class="user-name">{{ userData["name"] }}</h2>
            <button class="btn" @click="openModal"><img src="../assets/img/editBtnWhite.svg" alt="edit-btn"></button>
          </div>
          <div class="user-tagAgeGender_block">
            <div class="user-tag" v-if="userData.tag == ''"></div>
            <div class="user-tag" v-else>{{ userData["tag"] }}</div>
            <div class="user-ageGender" v-if="userData.birthday != ''"></div>
            <div class="user-ageGender" v-else-if="Age !== Age">{{ userData["gender"] }}</div>
            <div class="user-ageGender" v-else>{{ userData["gender"] }}, {{ Age }}</div>
          </div>
          <div class="user-locationStatus_block" v-bind:class="{ 'null': userData.location == '' }">
            <div class="user-location" v-if="userData.location == ''"></div>
            <div class="user-location" v-else>
              <span><img src="../assets/img/location.svg" alt="location"></span>
              {{ userData["location"] }}
            </div>
            <div class="user-status" v-if="userData.status == ''"></div>
            <div class="user-status" v-else>{{ userData["status"] }}</div>
          </div>
          <div class="user-emailPhone_block">
            <div class="user-email">{{ userData["email"] }}</div>
            <div class="user-phone"></div>
          </div>
        </div>
        <div class="btn-block" v-if="userRole != 'HR'">
          <NuxtLink class="btn employee" to="/employer">Работодателям</NuxtLink>
          <button class="btn download-cv disable">Скачать CV</button>
        </div>
        <div class="btn-block" v-else>
          <!--            <NuxtLink class="btn employee" to="/employer">Работодателям</NuxtLink>-->
          <button class="btn download-cv disable">Скачать CV</button>
          <button class="btn logout-btn" @click="logoutHandler">Выйти</button>
        </div>
      </div>
      <!--        <div class="user-social"></div>-->
    </div>
    <div class="company">
      <img class="logo" alt="logo" :src="`data:image/jpeg;base64,${companyData.photo}`">
      <div class="company-data">
        <div class="company-name">{{ companyData["name"] }}</div>
        <div class="company-location">
          {{ companyData["location"] }}
        </div>
        <NuxtLink class="btn success company-about_btn" :to="`/company/${companyData.id}`">Подробнее</NuxtLink>
      </div>
    </div>
  </div>
  <div class="container-content">
    <div class="about-me_block">
      <div class="about-me_head">
        <h2>Обо мне</h2>
        <button class="btn" @click="toggleEditModeDesc"><img src="../assets/img/editBtn.svg" alt="edit-btn"></button>
      </div>
      <div class="about-me_content">
        <div class="textArea" v-if="editMode">
          <textarea v-model="desc_content" @blur="saveDescription" rows="4"></textarea>
        </div>
        <div class="content" v-else>
          <p>{{ userData["description"] }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from "vue";

const { userData, companyData, Age, userRole, logoutHandler } = defineProps(["userData", "companyData", "Age", "userRole", "logoutHandler"])
const desc_content = ref("")

const editMode = ref(false)
const toggleEditModeDesc = () => {
  editMode.value = !editMode.value
}
const toggleEditModeStack = () => {
  editMode.value = !editMode.value
}
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

.textArea {
  width: 100%;
}

textarea {
  width: 100%;
  border: none;
}

.content {
  width: 100%;
  white-space: break-spaces;
}

.null {
  display: none;
}
</style>