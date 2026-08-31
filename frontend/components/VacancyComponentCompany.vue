<template>
  <div class="vacancy_item">
    <div :class="$style.vacancies_item_block">
      <div :class="$style.vacancies_item_header">
        <div :class="$style.vacancies_item_logo"></div>
        <div :class="$style.vacancies_item_vacanciesInfo">
          <div :class="$style.vacancies_item_vacanciesInfo_header">
            <div :class="$style.vacancies_item_name">IT Professionals Work</div>
            <div :class="$style.vacancies_item_tag">@ipw</div>
          </div>
          <div :class="$style.vacancies_item_levelDirection">Senior Golang Developer</div>
          <div :class="$style.vacancies_item_location">Ufa / Full Day</div>
          <!-- <div class="vacancies_item_levelDirection"></div> ?? >> email -->
        </div>
      </div>
      <div :class="$style.vacancies_item_description">
        Мы знаем, что такое скорбь, сидим вместе, создаем нашу элиту.
        Донец сусципит элементум нунц, сэд хендрерит метус конгуэ сэд.
        Нулла фасилиси. Преддверие и внутренняя оболочка.
        Ничто не имеет силы, ничто не говорит о том, что есть, конец всему сущему.
        Сид ид ниси в justo ultrices tempor. Преддверие вариуса сагиттиса маттиса.
        Сед виверра велит, чтобы я рисковал, а пеллентеск был на месте.
      </div>
      <div :class="$style.vacancies_item_skills">
        UI/UX дизайнер, Средний (Middle) • Figma Design • UI/UX дизайн •
        Брендирование • Проектирование интерфейсов • Исследование пользователя •
        Дизайн продукта • Дизайн иконок • Разработка фирменного стиля • Дизайн логотипов •
        Прототипирование
      </div>
    </div>
  </div>
</template>

<style module>
@import 'assets/css/vacancyCompany.module.css';
</style>