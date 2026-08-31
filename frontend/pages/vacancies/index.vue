<template>
<!--  <ElasticsearchComponent/>-->
  <div class="container">
    <section v-for="v in vacancies" :key="v.id">
      <VacancyComponent :vacancy="v" />
    </section>
  </div>
</template>

<script lang="ts">
import {defineComponent, ref, onMounted} from "vue";
import axios from "axios";
import VacancyComponent from "~/pages/vacancies/components/VacancyComponent.vue";
export default defineComponent({
  components: {VacancyComponent},
  VacancyComponent,
  setup() {
    const vacancies = ref([]);
    onMounted(async () => {
      try {
        const response = await axios.get(`http://localhost:5000/vacancy/v1/all-vacancies`);
        vacancies.value = response.data;
        console.log(vacancies.value)
      } catch (error) {
        console.error(error);
      }
    });

    return {vacancies};
  }
});
</script>

<style scoped>
.container {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  width: 100%;
  height: calc(100vh - 20px);
}
section {
  width: calc(50% - .7rem);
}
</style>