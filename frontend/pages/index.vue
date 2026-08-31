<template>
  <div class="home">
    <div class="centName">Заказы, безопасная сделка и выплаты по этапам</div>
    <div class="name">IPW — фриланс-биржа</div>
    <NuxtLink class="btn cta" :to="ctaLink">{{ ctaText }}</NuxtLink>

    <div class="steps">
      <div class="step">
        <span>1</span>
        <p>Заказчик размещает проект, фрилансеры отправляют отклики</p>
      </div>
      <div class="step">
        <span>2</span>
        <p>Принятый оффер → контракт с этапами, деньги в escrow</p>
      </div>
      <div class="step">
        <span>3</span>
        <p>Приёмка этапа — выплата фрилансеру за вычетом комиссии</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { isAuthed } = useAuth()
const ctaLink = computed(() => (isAuthed.value ? '/profile' : '/auth'))
const ctaText = computed(() => (isAuthed.value ? 'В профиль' : 'Регистрация / Авторизация'))
</script>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 25px;
  width: 100%;
  min-height: calc(100vh - 160px);
  text-align: center;
}
.centName {
  color: #fff;
  font-family: Raleway, sans-serif;
  font-size: 22px;
  font-weight: 700;
}
.name {
  font-family: Raleway, sans-serif;
  font-size: 56px;
  color: #fff;
}
.cta {
  font-family: Raleway, sans-serif;
  font-weight: 700;
  font-size: 18px;
  padding: 16px 64px;
  color: #fff;
  background-color: #325181;
  border-radius: 12px;
  text-decoration: none;
}
.steps {
  display: flex;
  gap: 20px;
  margin-top: 30px;
  flex-wrap: wrap;
  justify-content: center;
}
.step {
  background: #fff;
  color: #000;
  border-radius: 20px;
  padding: 24px;
  width: 240px;
  text-align: left;
}
.step span {
  display: block;
  font-size: 28px;
  font-weight: 800;
  color: #4e92f8;
}
.step p {
  margin-top: 8px;
  font-size: 14px;
  font-weight: 500;
}
</style>
