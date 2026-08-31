<template>
  <div class="wrapper">
    <NuxtLayout>
      <div class="container">
        <NuxtPage/>
      </div>
    </NuxtLayout>
  </div>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import ResumeComponent from "~/components/ResumeComponent.vue";
import VacancyComponent from "~/pages/vacancies/components/VacancyComponent.vue";

export default defineComponent({
  components: {
    ResumeComponent,
    VacancyComponent,
  },
  head() {
    definePageMeta({
      layout: 'default',
    })
  },
  // setup() {
  //   console.log("bongaclamms ^)")
  // }
})
</script>

<style scoped>
* {
  font-family: Inter, sans-serif;
  font-style: normal;
  font-weight: 700;
  line-height: normal;
}

.container {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-height: calc(100vh - 80px);
  margin: 20px 0;
  gap: 35px;
  color: #fff;
  width: 100%;
}
</style>