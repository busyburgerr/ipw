<template>
  <div class="wrapper">
    <NuxtLayout>
      <div class="container">
        <NuxtPage />
      </div>
    </NuxtLayout>
  </div>
</template>

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
