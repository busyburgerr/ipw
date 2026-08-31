<template>
  <div class="container">
    <div class="createResume-left_block">
      <h1 class="logo">IPW</h1>
      <h3 class="second-logo">itprofessionalswork</h3>
      <h2 class="createResume-title">Создание Резюме</h2>
    </div>
    <div class="createResume-right_block">
      <div class="createResume-content_block">
        <h1 class="createResume-head">{{ createResumeHeadTitle }}</h1>
        <div class="createResume-progressDot_block">
          <div class="progress-dots">
            <div v-for="(step, index) in steps" :key="index" :class="dotClass(index)" @click="goToStep(index)">
<!--              {{ index + 1 }}-->
            </div>
          </div>
        </div>
        <div class="createResume-content">
          <div class="step" v-show="currentStep === 0">
            <div class="createResume-specialize_block">
              <h3 class="createResume-specialize_head">Специализация</h3>
              <div class="createResume-specialize_content">
                <input v-model="specializeInput" type="text" class="specialize-input">
              </div>
            </div>
            <div class="createResume-status_block">
              <h3 class="createResume-status_head">Готовность к работе</h3>
              <div class="createResume-status_content">
                <select v-model="statusSelect" class="status-select" name="status-select">
                  <option value="Ищу работу">Ищу работу</option>
                  <option value="Не ищу работу">Не ищу работу</option>
                  <option value="Рассмотрю предложения">Рассмотрю предложения</option>
                </select>
              </div>
            </div>
            <div class="createResume-salary_block">
              <h3 class="createResume-salary_head">Ожидаемое вознаграждение</h3>
              <div class="createResume-salary_content">
                <input v-model="salaryInput" type="text" class="salary-input">
              </div>
            </div>
            <div class="createResume-description_block">
              <h3 class="createResume-description_head">Напишите о себе</h3>
              <div class="createResume-description_content">
                <textarea v-model="descResume" type="text" class="descriptionResume-textarea"></textarea>
              </div>
            </div>
            <div class="createResume-skills_block">
              <div class="createResume-skills_head">
                <h3>Укажите свой стэк</h3>
                <button class="btn">Взять из профиля</button>
              </div>
              <div class="createResume-skills_content">
                <textarea type="text" v-model="skillsValue" class="skillsResume-textarea"></textarea>
              </div>
            </div>
          </div>
<!--          <div v-show="currentStep === 1">Content 2</div>-->
<!--          <div v-show="currentStep === 2">Content 3</div>-->
        </div>
      </div>
      <div class="createResume-btn_block">
        <button class="btn btn-back" @click="prevStep" :disabled="currentStep === 0">Назад</button>
        <button class="btn success" v-if="currentStep == 0" @click="createResume">Сохранить</button>
<!--        :disabled="currentStep === steps.length - 1"-->
      </div>
    </div>

    <div>
      <h1>test request</h1>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import axios from "axios";
import {algorithms} from "iron-webcrypto";

// const steps = ref([1, 2, 3])
const steps = ref([1])
const currentStep = useState<number>('currentStep', () => 0)
const createResumeHeadTitle = "<info />"


const specializeInput = ref("")
// Middle Golang Backend
// const level = specializeInput.value.split(" ")[0]
// const direction = specializeInput.value.split(" ")[1] + " " + specializeInput.value.split(" ")[2]
const statusSelect = ref("")
const salaryInput = ref("")
const descResume = ref("")
const skillsValue = ref("")

const dotClass = (index: any) => {
  return {
    'progress-dot': true,
    'active': index <= currentStep.value
  }
}

const createResume = async () => {
  const direction = specializeInput.value.split(" ")[1] + " " + specializeInput.value.split(" ")[2]
  const level = specializeInput.value.split(" ")[0]
  try {
    const response = axios.post(`http://localhost:5000/data/v1/user/create-resume`, {
      direction: direction,
      level: level,
      salary: salaryInput.value,
      status: statusSelect.value,
      description: descResume.value,
      skills: skillsValue.value,
    }, {withCredentials: true})
    console.log(response)
    // console.log(skillsValue.value)
  } catch (e) {
    console.log(e)
  }
}
const goToStep = (index: any) => {
  currentStep.value = index
}
const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}
const nextStep = () => {
  if (currentStep.value < steps.value.length - 1) {
    currentStep.value++
  }
}
</script>

<style scoped>
@import "./assets/css/style.css";
.progress-dots {
  display: flex;
}

.progress-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background-color: rgba(189, 189, 189, 0.78);
  margin-right: 5px;
}

.progress-dot.active {
  background-color: rgba(0, 102, 255, 0.78);
}

.step {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.success {
  font-size: 16px;
  font-weight: 500;
  color: #fff;
}
.btn-back {
  display: flex;
  padding: 10px 20px;
  flex-direction: column;
  align-items: flex-start;
  border-radius: 10px;
  border: 1px solid rgba(0, 0, 0, 0.30);
  color: #000;
  font-size: 16px;
  font-weight: 500;
}
</style>