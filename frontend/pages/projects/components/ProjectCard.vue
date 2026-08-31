<template>
  <div class="vacancies_item_block">
    <div class="vacancies_item_header">
      <div class="vacancies_item_logo" />
      <div class="vacancies_item_vacanciesInfo">
        <div class="vacancies_item_vacanciesInfo_header">
          <span class="vacancies_item_name">{{ budget }}</span>
          <div class="vacancies_item_tag">{{ ru(project.status) }}</div>
        </div>
        <NuxtLink :to="`/projects/${project.id}`" class="vacancies_item_levelDirection">
          {{ project.title }}
        </NuxtLink>
        <div class="vacancies_item_location">
          {{ ru(project.experienceLevel) }} · {{ project.proposalsCount }} откликов
        </div>
      </div>
    </div>
    <div class="vacancies_item_description">{{ project.description }}</div>
    <div class="vacancies_item_btnBlock">
      <NuxtLink class="btn send" :to="`/projects/${project.id}`">Смотреть проект</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ project: any }>()

const budget = computed(() => {
  const p = props.project
  if (p.budgetType === 'fixed') return money(p.fixedAmountCents, p.currency)
  return `${money(p.hourlyRateMinCents, p.currency)}–${money(p.hourlyRateMaxCents, p.currency)} / час`
})
</script>

<style scoped>
@import '../assets/css/vacancy-component.css';

.vacancies_item_levelDirection {
  font-size: 20px;
  text-decoration: none;
}
.send {
  text-decoration: none;
}
</style>
