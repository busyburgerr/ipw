<template>
  <div class="notice_container">
    <div class="notice_head">
      <div class="head">Уведомение!</div>
      <div class="tag">@notice</div>
    </div>
    <div class="notice_content">
      <div class="notice_info">Сайт на стадии MVP, многие функции могут не работать или работать плохо</div>
      <div class="notice_second-info">
        <div class="email">&#8226 Баги или ошибки? пишите сюда &#8594; <NuxtLink to="mailto:ipwpost@bk.ru">ipwpost@bk.ru</NuxtLink></div>
        <div class="github">&#8226 github &#8594; <NuxtLink to="https://github.com/zzzhubatyrov" target="_blank">@zzzhubatyrov</NuxtLink></div>
      </div>
    </div>
    <div class="notice_footer">
      <button class="btn success_btn" @click="closeModal">Понятно</button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  closeModal: {
    type: Function,
    required: true,
  }
})
</script>

<style scoped>
a {
  color: #000;
}
.notice_container {
  display: flex;
  width: 800px;
  padding: 20px;
  flex-direction: column;
  align-items: flex-start;
  gap: 20px;
  border-radius: 20px;
  background: #FFF;
  color: #000;
}
.notice_head {
  display: flex;
  align-items: center;
  gap: 10px;
  align-self: stretch;
}
.head{
  font-size: 24px;
  font-weight: 700;
}
.tag{
  color: rgba(0, 0, 0, 0.50);
  font-size: 16px;
  font-weight: 500;
}
.notice_content {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  align-self: stretch;
  flex-direction: column;
}
.notice_info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  align-self: stretch;
  font-size: 24px;
  font-weight: 500;
}
.notice_second-info{
  font-size: 20px;
  font-weight: 500;
}
.notice_footer{
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 10px;
  align-self: stretch;
}
.success_btn{
  display: flex;
  padding: 10px 15px;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 10px;
  border-radius: 10px;
  background: #438EFF;
  color: #FFF;
  font-size: 20px;
  font-weight: 500;
}
</style>