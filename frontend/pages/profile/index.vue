<template>
  <div class="container" v-if="userRole == 'Пользователь'">
    <div class="container-head">
      <MainProfile :userData="userData" :openUploadModal="openUploadModal" :Age="Age"/>
    </div>
    <div class="container-content">
      <div class="about-me_block">
        <div class="about-me_head">
          <h2>Обо мне</h2>
          <button class="btn" @click="toggleEditModeDesc"><img src="./assets/img/editBtn.svg" alt="edit-btn"></button>
        </div>
        <div class="about-me_content">
          <div class="textArea" v-if="editModeDesc">
            <textarea v-model="content" @blur="saveDescription" rows="4"></textarea>
          </div>
          <div class="content" v-else>
            <p>{{ userData.description }}</p>
          </div>
        </div>
      </div>
      <div class="second-content_block">
        <div class="education-block">
          <div class="education-head">
            <h2>Образование</h2>
            <button class="btn"><img src="./assets/img/editBtn.svg" alt="edit-btn"></button>
          </div>
          <div class="education-content"></div>
        </div>
        <div class="experience-block">
          <div class="experience-head">
            <h2>Опыт работы</h2>
            <button class="btn"><img src="./assets/img/editBtn.svg" alt="edit-btn"></button>
          </div>
          <div class="experience-content"></div>
        </div>
      </div>
      <div class="resume-block">
        <div class="resume-header">
          <h2>Резюме</h2>
          <NuxtLink class="btn" to="/create-resume">
            <img src="./assets/img/editBtn.svg" alt="add-resume">
          </NuxtLink>
        </div>
        <div class="resume-content">
          <div class="resume-item" v-for="resumeItem in userDataResumes">
            <ProfileResumeComponent :profileResumeData="resumeItem"/>
            <hr>
          </div>
        </div>
      </div>
      <div class="stack-block">
        <div class="stack-head">
          <h2>Стэк</h2>
          <button class="btn" @click="toggleEditModeStack"><img src="./assets/img/editBtn.svg" alt="edit-btn"></button>
        </div>
        <div class="stack-content">
          <div class="textArea" v-if="editModeStack">
            <textarea v-model="content_stack" @blur="saveStack" rows="4"></textarea>
          </div>
          <div class="stack-item" v-for="item in usrStack" v-else-if="usrStack.length === 0">
            {{ item }}
          </div>
        </div>
      </div>
    </div>
  </div>
  <div class="container" v-else>
    <HrProfileComponent :user-data="userData" :company-data="companyData" :age="Age" :user-role="userRole"
                        :logout-handler="logoutHandler"/>
    <div class="second-content_block">
      <div class="education-block">
        <div class="education-head">
          <h2>Образование</h2>
          <button class="btn"><img src="./assets/img/editBtn.svg" alt="edit-btn"></button>
        </div>
        <div class="education-content"></div>
      </div>
      <div class="experience-block">
        <div class="experience-head">
          <h2>Опыт работы</h2>
          <button class="btn"><img src="./assets/img/editBtn.svg" alt="edit-btn"></button>
        </div>
        <div class="experience-content"></div>
      </div>
    </div>
    <div class="stack-block">
      <div class="stack-head">
        <h2>Стэк</h2>
        <button class="btn" @click="toggleEditModeStack"><img src="./assets/img/editBtn.svg" alt="edit-btn"></button>
      </div>
      <div class="stack-content">
        <div class="textArea" v-if="editMode">
          <textarea v-model="content_stack" @blur="saveStack" rows="4"></textarea>
        </div>
        <div class="stack-item" v-for="item in usrStack" v-else>
          {{ item }}
        </div>
      </div>
    </div>
    <div class="resume-block">
      <div class="resume-header">
        <h2>Резюме</h2>
        <NuxtLink class="btn" to="/create-resume">
          <img src="./assets/img/editBtn.svg" alt="add-resume">
        </NuxtLink>
      </div>
      <div class="resume-content">
        <div class="resume-item" v-for="resumeItem in userDataResumes">
          <ProfileResumeComponent :userData="userData" :profileResumeData="resumeItem"/>
          <hr>
        </div>
      </div>
    </div>
  </div>
  <!--  <div class="modal" v-if="showModal">-->
  <!--    <UserDataComponent :closeModal="closeModal"/>-->
  <!--  </div>-->
  <div class="modal" v-if="uploadModal">
    <UploadPhotoComponent :closeUploadModal="closeUploadModal" @value="handleFormDataType"/>
  </div>
</template>

<script setup lang="ts">
import UserDataComponent from "~/components/UserDataComponent.vue";
import ProfileResumeComponent from "./components/ProfileResumeComponent.vue";
import UploadPhotoComponent from "./components/UploadPhotoComponent.vue";
import HrProfileComponent from "./components/HrProfileComponent.vue"

import {ref} from "vue";
import axios from "axios";
import Cookies from "js-cookie";
import MainProfile from "~/pages/profile/components/MainProfile.vue";

const router = useRouter()
const userData = ref([])
const companyData = ref([])
let Age = ref()
const userDataResumes = ref([])
let currentDate = new Date()
let userRole = ref('')
const editModeDesc = ref(false)
const editModeStack = ref(false)
const desc_content = ref("")
const formDataType = ref("")

const handleFormDataType = (value: any) => {
  formDataType.value = value
  console.log(value)
}
const toggleEditModeDesc = () => {
  editModeDesc.value = !editModeDesc.value
}
const toggleEditModeStack = () => {
  editModeStack.value = !editModeStack.value
}
const saveDescription = async () => {
  try {
    await axios.put("http://localhost:5000/data/v1/user/update", {
      description: desc_content.value,
    }, {withCredentials: true})
  } catch (e) {
    console.log(e)
  }
}
const content_stack = ref("")
const saveStack = async () => {
  try {
    await axios.put("http://localhost:5000/data/v1/user/update", {
      stack: content_stack.value,
    }, {withCredentials: true})
  } catch (e) {
    console.log(e)
  }
}
const usrStack = ref([])
onMounted(async () => {
  try {
    const res = await axios.get(`http://localhost:5000/data/v1/user`, {withCredentials: true})
    userData.value = res.data
    userDataResumes.value = res.data.resumes
    // @ts-ignore
    const date = userData.value.birthday.replace(/\./g, '-')
    const new_date = date.split('-').reverse().join('-')
    const birth = new Date(new_date)
    let age = currentDate.getFullYear() - birth.getFullYear()
    let monthDiff = currentDate.getMonth() - birth.getMonth()
    if (monthDiff < 0 || (monthDiff === 0 && currentDate.getDate() < birth.getDate())) {
      age--;
    }
    Age.value = age
    // @ts-ignore
    userRole.value = userData.value.role["role_name"]
    // @ts-ignore
    if (!userData.value.gender) {
      showModal.value = true;
    }
    // @ts-ignore
    usrStack.value = userData.value["stack"].split(' ')
    console.log(usrStack.value)
    console.log(userData.value)
    // || !userData.value.age || !userData.value.location || !userData.value.gender || !userData.value.number
  } catch (e) {
    console.log(e)
  }
})

onMounted(async () => {
  try {
    const res = await axios.get(`http://localhost:5000/data/v1/user`, {withCredentials: true})
    companyData.value = res.data.company
  } catch (e) {
    console.log(e)
  }
})
const logoutHandler = async () => {
  try {
    await axios.post(`http://localhost:5000/auth/v1/logout`)
        .then(() => {
          Cookies.remove('ipw', {expires: -1})
          window.location.reload()
        })
        .catch(e => console.log(e))
  } catch (e) {
    console.log(e)
  }
}
onMounted(() => {
  const isAuth = Cookies.get("ipw")
  if (!isAuth) {
    router.push('/auth')
  }
})
defineComponent({
  UserDataComponent,
  ProfileResumeComponent,
  UploadPhotoComponent,
  HrProfileComponent
})

const showModal = ref(false)
const uploadModal = ref(false)
const openModal = () => {
  showModal.value = true
}
const closeModal = () => {
  showModal.value = false
}
const openUploadModal = () => {
  uploadModal.value = true
}
const closeUploadModal = () => {
  uploadModal.value = false
}
</script>

<style scoped>
@import './assets/css/profile.css';

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