<template>
  <div :class="$style.resume_block">
    <div :class="$style.resume_header_block">
      <div :class="$style.resume_userInfo_block">
        <div :class="$style.resume_logo"></div>
        <div :class="$style.resume_content_block">
          <div :class="$style.resume_nameTag_block">
            <div :class="$style.resume_name">{{ resume.userName }}</div>
            <div :class="$style.resume_tag">{{ resume.userTag }}</div>
          </div>
          <div :class="$style.resume_secondInfo_block">
            <NuxtLink :to="`/resumes/${resume.id}`" :class="$style.resume_levelDirection">{{ resume.level }} {{ resume.direction }}</NuxtLink>
            <div :class="$style.resume_locationWorkTime">{{ resume.location }} / {{ resume.status }}</div>
            <div :class="$style.resume_email">{{ resume.userEmail }}</div>
          </div>
        </div>
      </div>
      <div :class="$style.resume_salary">{{ resume.salary }}</div>
    </div>
    <div :class="$style.resume_description_block">{{ resume.description }}</div>
    <div :class="$style.resume_skills_block"></div>
    <div :class="$style.resume_btnBlock">
      <div class="btn" :class="$style.send">Написать</div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { resume } = defineProps(['resume'])
</script>

<style module>
@import '@/assets/css/resume.module.css';
</style>