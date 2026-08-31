<template>
<!--  <ElasticsearchComponent/>-->
  <div class="container">
    <div class="r_item" v-for="r in resumes" :key="r.id">
      <ResumeComponent :resume="r"/>
    </div>
  </div>
</template>

<script lang="ts">
import {defineComponent, onMounted, ref} from "vue";
import axios from "axios";

export default defineComponent({
  setup() {
    const resumes = ref([]);
    onMounted(async () => {
      try {
        const response = await axios.get(`http://localhost:5000/resume/v1/resumes`);
        resumes.value = response.data;
        console.log(resumes.value)
      } catch (error) {
        console.error(error);
      }
    });

    return {resumes};
  }
});
</script>

<style scoped>
.container {
  display: flex;
  align-items: flex-start;
  align-content: flex-start;
  gap: 30px 50px;
  align-self: stretch;
  flex-wrap: wrap;
  justify-content: space-between;
  height: calc(100vh - 20px);
}
.r_item {
  display: flex;
  gap: 20px;
  width: calc(50% - 5.4rem);
}
</style>