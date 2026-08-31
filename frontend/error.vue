<template>
  <div class="container">
      <div class="errContainer">
          <h1>{{ error.statusCode }}</h1>
          <h2>Page Not Found!</h2>
          <p>{{error.message}}</p>
      </div>
  </div>
</template>

<script setup lang="ts">
import {defineProps} from 'vue'
defineProps(['error'])
</script>

<style scoped>
.container {
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 200px 0;
}
.errContainer {
    display: flex;
    max-width: 20vw;
    padding: 100px;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    gap: 20px;
    border-radius: 30px;
    background: #FFF;
    color: #000;
    text-align: center;
    font-family: Inter, sans-serif;
    font-style: normal;
    line-height: normal;
}
h1, h2, p {
    margin: 0;
    padding: 0;
}
h1 {
    font-size: 64px;
    font-weight: 500;
}
h2 {
    font-size: 36px;
    font-weight: 500;
}
p {
    font-size: 20px;
    font-weight: 500;
}
</style>