<template>
  <div class="container">
    <div class="container-head">
      <div class="company">
        <img class="logo" :src="`data:image/jpeg;base64,${company.photo}`" alt="logo"/>
        <div class="company-data">
          <div class="name">{{ company["name"] }}</div>
          <div class="location">{{ company["location"] }}</div>
          <!--    this block for link on web-site      -->
        </div>
      </div>
      <div class="company-contacts">
        <div class="contacts-info">
          <!--          <div class="vacancy-count"><h3>Вакансии</h3> 200</div>-->
          <!--          <div class="company-size">-->
          <!--            <h3>Размер компании</h3>-->
          <!--            <p>от 100 до 500</p>-->
          <!--          </div>-->
          <div class="main-contacts">
            <div class="phone">
              <p>Номер телефона:</p>
              {{ company["phone"] }}
            </div>
            <div class="email">
              <p>Почта:</p>
              {{ company["email"] }}
            </div>
          </div>
        </div>
        <div class="social-btn_block">
          <div class="social-contacts"></div>
          <!--          <button class="btn want-to-work_btn">{{ wantToWork }}</button>-->
        </div>
      </div>
    </div>
    <div class="container-content">
      <div class="about-company">
        <div class="about-company_head">
          О компании "{{ company["name"] }}"
          <button class="btn" v-if="companyUserID == userID"><img src="~/assets/img/editBtn.svg" alt="edit-btn"></button>
        </div>
        <div class="about-company_content">{{ company["description"] }}</div>
      </div>
      <div class="company-vacancy">
        <div class="company-vacancy_head">
          Вакансии компании
          <NuxtLink class="btn" to="/company/create-vacancy" v-if="companyUserID == userID"><img src="~/assets/img/ic_plus.svg" alt="ic_plus"></NuxtLink>
        </div>
        <div class="not-vacancies" v-if="company.vacancies == null">Нет вакансий</div>
        <div class="not-vacancies" v-else></div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import {defineComponent, onMounted, ref} from "vue";
import axios from "axios";

export default defineComponent({
  setup() {
    const wantToWork = "<Хочу тут работать />"
    const {id} = useRoute().params
    // http://localhost:5000/company/v1/${id}
    const uri = "http://localhost:5000/company/v1/" + id
    const company = ref([])
    const userData = ref([])
    const companyUserID = ref()
    const userID = ref()
    onMounted(async () => {
      try {
        const response = await axios.get(uri)
        company.value = response.data
        companyUserID.value = response.data.userID
        console.log("company-data:", company.value)
      } catch (e) {
        console.log(e)
      }
    })
    onMounted(async () => {
      const response = await axios.get(`http://localhost:5000/data/v1/user`, {withCredentials: true})
      userData.value = response.data
      userID.value = response.data.id
      console.log("user-data:", userData.value)
    })
    return {
      company,
      wantToWork,
      userData,
      companyUserID,
      userID,
    }
  }
})
</script>

<style scoped>
@import "./assets/css/company-[id].css";

* {
  font-family: Inter, sans-serif;
}

h1, h2, h3, h4, p {
  margin: 0;
  padding: 0;
}
</style>