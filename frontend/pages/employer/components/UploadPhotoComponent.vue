<template>
<!--  @submit.prevent="uploadPhoto"-->
  <form class="upload-form" enctype="multipart/form-data">
    <h2 class="head">Загрузить фото</h2>
    <div class="content">
      <input ref="fileInput" id="file-input" type="file" name="photo"/>
      <label for="file-input"></label>
      <div class="upload-btn_block">
        <button class="btn success" @click="closeUploadModal">Закрыть</button>
      </div>
    </div>
  </form>
</template>

<script setup>
import axios from 'axios'
import { defineProps, ref } from 'vue'

defineProps({
  closeUploadModal: {
    type: Function,
    required: true
  }
})
const fileInput = ref(null)

// const uploadPhoto = async () => {
//   const formData = new FormData()
//   formData.append('photo', fileInput.value.files[0])
//
//   try {
//     const response = await axios.put('http://localhost:5000/data/v1/user/upload-photo', formData, { withCredentials: true })
//     console.log(response.data)
//     window.location.reload()
//   } catch (e) {
//     console.log(e)
//   }
// }
</script>

<style scoped>
.upload-form {
  display: flex;
  width: 507px;
  height: 164px;
  padding: 20px 30px;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  background: #FFF;
  color: #000;
  border-radius: 20px;
}
.head {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  align-self: stretch;
  font-size: 24px;
  font-weight: 500;
}
.content {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 10px;
  flex: 1 0 0;
  align-self: stretch;
}

.upload-btn_block {
  display: flex;
  gap: 20px;
}
</style>