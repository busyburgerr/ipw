<template>
  <div class="container">
    <div class="container-head">
    </div>
    <div class="container-content">
      <div class="resume">
        <div class="resume-head">
          <h3>Резюме</h3>
        </div>
        <div class="resume-content"></div>
      </div>
      <div class="stack">
        <div class="stack-head">
          <h3>Стэк</h3>
        </div>
        <div class="stack-content"></div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import {defineComponent, onMounted, ref} from "vue";
import axios from "axios";

const userData = ref([])
let Age = ref()
let currentDate = new Date()
const showModal = ref(false)
const usrStack = ref([])
const openModal = () => {
  showModal.value = true
}
const closeModal = () => {
  showModal.value = false
}
export default defineComponent({
  setup() {
    const {id} = useRoute().params
    const uri = `http://localhost:5000/resume/v1/` + id
    const resume = ref([])
    onMounted(async () => {
      try {
        const response = await axios.get(uri)
        resume.value = response.data
        console.log("resume-data:",resume.value)
      } catch (e) {
        console.log(e)
      }
    })
    return {
      resume,
    }
  }
})
</script>

<style scoped>
@import "assets/css/resumeID.css";
h1, h2, h3, h4, p {
  margin: 0;
  padding: 0;
}
</style>