<template>
  <div class="container">
    <div class="create-company_block">
      <h2 class="create-company_head">Создание профиля компании</h2>
      <div class="create-company_content">
        <form class="create-company_form" enctype="multipart/form-data" @submit.prevent="createCompanyHandler">

          <!--          <label for="photo" v-if="showUpLoad" ref="fileInput" class="logo"-->
          <!--                 style="display:flex; align-items: center; justify-content: center">-->
          <!--            <img src="@/assets/img/camera.svg" alt="add-photo"/>-->
          <!--            <input class="logo" type="file" ref="fileInput2" name="photo" id="photo" @change="handleFileChange"/>-->
          <!--          </label>-->
          <input type="file" id="photo" @change="handlePhotoUpload" accept="image/*">
          <input type="text" v-model="companyName" placeholder="Название компании">
          <input type="text" v-model="companyTag" placeholder="Тэг компании">
          <textarea placeholder="Описание компании" v-model="companyDescription"></textarea>
          <input type="text" v-model="companyLocation" placeholder="Местоположение">
          <input type="text" v-model="companyWebSite" placeholder="Веб-сайт">
          <input type="email" v-model="companyEmail" placeholder="Email">
          <input type="text" v-model="companyPhone" placeholder="Телефон">
          <button class="btn success" type="submit">Создать</button>
        </form>
      </div>
    </div>
  </div>
  <div class="modal" v-if="uploadModal">
    <UploadPhotoComponent :closeUploadModal="closeUploadModal"/>
  </div>
</template>

<script setup>
import {ref} from 'vue'
import axios from "axios";
import UploadPhotoComponent from './components/UploadPhotoComponent.vue'

let companyName = ref('')
let companyTag = ref('')
let companyDescription = ref('')
let companyLocation = ref('')
let companyWebSite = ref('')
let companyEmail = ref('')
let companyPhone = ref('')
let companyPhoto = ref(null)
const router = useRouter()

const handlePhotoUpload = (event) => {
  companyPhoto.value = event.target.files[0];
}

const createCompanyHandler = async () => {
  const formData = new FormData()
  formData.append("name", companyName.value)
  formData.append("photo", companyPhoto.value)
  formData.append("tag", companyTag.value)
  formData.append("email", companyEmail.value)
  formData.append("phone", companyPhone.value)
  formData.append("location", companyLocation.value)
  formData.append("description", companyDescription.value)
  try {
    const res = await axios.post(`http://localhost:5000/data/v1/user/create-company`,
      formData, {withCredentials: true, headers: {"Content-Type": "multipart/form-data"}})
    console.log(res)
    router.push('/profile')
  } catch (e) {
    console.log(e)
  }
}
const uploadModal = ref(false)
const openUploadModal = () => {
  uploadModal.value = true
}
const closeUploadModal = () => {
  uploadModal.value = false
}

let imagePreviewURL = ref(null)
let showUpLoad = ref(true)
const handleFileChange = (event) => {
  const file = event.target.files[0];
  const reader = new FileReader();
  reader.onload = () => {
    imagePreviewURL.value = reader.result;
    showUpLoad.value = false
  };
  reader.readAsDataURL(file);
}

defineComponent({
  UploadPhotoComponent
})
</script>

<style scoped>
@import "./assets/css/style.css";
</style>