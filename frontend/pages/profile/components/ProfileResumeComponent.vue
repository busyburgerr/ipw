<template>
  <div class="resume-item_block">
    <div class="resume-item_head">
      <div class="user-data">
        <h3 class="level-direction">{{ profileResumeData['level'] }} {{ profileResumeData['direction'] }} </h3>
        <div class="name-tag_block">
          <h3>{{ userData['name'] }}</h3>
          <h3>{{ userData['tag'] }}</h3>
        </div>
        <h3 class="status">{{ profileResumeData['status'] }}</h3>
        <h3 class="email">{{ userData['email'] }}</h3>
      </div>
      <div class="salary">{{ profileResumeData['salary'] }}₽ на руки</div>
    </div>
    <div class="resume-item_content">
      <p class="about-resume">{{ profileResumeData['description'] }}</p>
      <div class="skills">
        <div class="skill-item" v-for="skill in skills">
          {{ skill }}
        </div>
      </div>
      <div class="btn-block">
        <button class="btn delete-btn">Удалить</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const {profileResumeData, userData} = defineProps(["profileResumeData", "userData"])
const skills = profileResumeData.skills.split(" ")
</script>

<style scoped>
@import "../assets/css/resumeComponent.css";
h1, h2, h3, h4, p {
  margin: 0;
  padding: 0;
}
</style>