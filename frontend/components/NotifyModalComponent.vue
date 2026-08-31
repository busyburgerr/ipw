<template>
  <div class="container">
    <div class="notice-head">
      <h2>Уведомление!</h2>
      <h4>@notice</h4>
    </div>
    <div class="notice-content">
      <p>Здесь какой-то контент</p>
    </div>
    <div class="notice-footer">
      <div class="btn success" @click="closeModal">Закрыть</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  closeModal: {
    type: Function,
    required: true
  }
})
</script>

<style scoped>

.container {
  display: flex;
  //width: 800px;
  padding: 20px;
  flex-direction: column;
  align-items: flex-start;
  gap: 20px;
  border-radius: 20px;
  background: #FFF;
  color: #000;
}

.notice-head {
  display: flex;
  align-items: center;
  gap: 10px;
  align-self: stretch;
}

.notice-head > h4 {
  color: rgba(0, 0, 0, 0.50);
  font-size: 16px;
  font-weight: 500;
}

.notice-content {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 20px;
  align-self: stretch;
}

.notice-footer {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 10px;
  align-self: stretch;
}
</style>