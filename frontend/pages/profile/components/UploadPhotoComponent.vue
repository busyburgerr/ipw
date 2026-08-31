<template>
  <form class="upload-form" enctype="multipart/form-data" @submit.prevent="uploadPhoto">
    <h2 class="head">Загрузить фото</h2>
    <div class="content">
      <input ref="fileInput" id="file-input" type="file" name="photo"/>
      <label for="file-input"></label>
      <div class="upload-btn_block">
        <button class="btn success" @click="closeUploadModal">Закрыть</button>
        <button class="btn success" type="submit">Загрузить</button>
      </div>
    </div>
  </form>
</template>

<script setup lang="ts">
import axios from 'axios'
import { defineComponent, ref, defineEmits } from 'vue'

defineProps({
  closeUploadModal: {
    type: Function,
    required: true
  }
})

const fileInput = ref(null)
const formData = new FormData()
const data = ref("")

const isValidImage = (imageType) => {
  const validFormats = ['image/jpeg', 'image/png'];
  return validFormats.includes(imageType);
};
const uploadPhoto = async () => {
  const file = fileInput.value.files[0]
  if (file) {
    if (isValidImage(file['type'])) {
      formData.append('photo', file);
      data.value = file['type'];
    } else {
      throw new Error("Неверный формат файла")
    }
  }
  try {
    await axios.put('http://localhost:5000/data/v1/user/upload-photo', formData, { withCredentials: true })
    window.location.reload()
  } catch (e) {
    console.log(e)
  }
  emitValue(data.value)
}
const emit = defineEmits(['value'])
const emitValue = (value) => {
  emit('value', value)
}

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