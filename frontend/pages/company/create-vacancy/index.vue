<template>
  <div class="container">
    <div class="container-head">
      <h2>Создание вакансии</h2>
    </div>
    <div class="container-content">
      <form class="create-vacancy_form">
        <div class="specialize-block">
          <h3 class="specialize-head">Специализация</h3>
          <input v-model="specialize_input" class="specialize-input_content" type="text" />
        </div>
        <div class="salary-block">
          <h3 class="salary-head">Зарплата</h3>
          <input v-model="salary_input" class="salary-input_content" type="text" />
        </div>
        <div class="exp-block">
          <h3 class="exp-head">Требуемый опыт</h3>
          <input v-model="exp_input" class="exp-input_content" type="text" />
        </div>
        <div class="location-block">
          <h3 class="location-head">Регион/Город работы</h3>
          <input v-model="location_input" class="location-input_content" type="text" />
        </div>
        <div class="workTime-block">
          <h3 class="workTime-head">Занятость</h3>
          <input v-model="workTime_input" class="workTime-input_content" type="text" />
        </div>
        <div class="skills-block">
          <h3 class="skill-head">Ключевые навыки</h3>
          <div class="skills-content">
            <input v-model="skill_input" @input="addItem" class="skill-input" type="text" />
            <div class="skill-input_content">
              <div class="skill-item" v-for="item in skills" v-if="skill_input.trim() !== ''">
                {{ item }}
              </div>
            </div>
          </div>
        </div>
        <div class="desc-block">
          <h3 class="desc-head">Описание вакансии</h3>
          <textarea v-model="desc_textArea" class="desc-input_content"></textarea>
        </div>
      </form>
      <div class="btn-block">
        <button class="btn back">Назад</button>
        <button class="btn save" @click="createBtn">Сохранить и продолжить</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import axios from "axios";
import router from "#app/plugins/router";

const specialize_input = ref('')
const salary_input = ref('')
const exp_input = ref('')
const location_input = ref('')
const workTime_input = ref('')
const skill_input = ref('')
const skills = ref()
const desc_textArea = ref('')

const direction = ref('')
const level = ref('')

let route = useRouter()
const createBtn = async () => {
  direction.value = specialize_input.value.split(' ')[1] + ' ' + specialize_input.value.split(' ')[2]
  level.value = specialize_input.value.split(' ')[0]
  try {
    const resp = await axios.post('http://localhost:5000/vacancy/v1/create-vacancy', {
      direction: direction.value,
      level: level.value,
      location: location_input.value,
      workTime: workTime_input.value,
      experience: exp_input.value,
      description: desc_textArea.value,
      skills: skill_input.value,
      salary: salary_input.value,
    }, {withCredentials:true})
    console.log(resp.data)
    route.push('/profile')
  } catch (e) {
    console.log(e)
  }
}

const addItem = () => {
  skills.value = skill_input.value.split(' ')
}

// direction
// level
// location
// workTime
// experience
// description
// skills
// salary

// Go PostgreSQL Docker RabbitMQ JS Vue.js React.js gin Fiber Elasticsearch

// specialize-input_content
// salary-input_content
// exp-input_content
// workTime-input_content
// skill-input_content
// desc-input_content
</script>

<style scoped>
@import "./assets/css/style.css";
h1, h2, h3, h4, p {
  margin: 0;
  padding: 0;
}
</style>