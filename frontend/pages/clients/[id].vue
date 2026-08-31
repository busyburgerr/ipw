<template>
  <div class="container" v-if="c">
    <div class="container-head">
      <div class="company">
        <div class="logo" />
        <div class="company-data">
          <div class="name">{{ c.companyName || c.displayName || 'Заказчик' }}</div>
          <div class="location">{{ c.location }}</div>
        </div>
      </div>
      <div class="company-contacts">
        <div class="contacts-info">
          <div class="stat"><b>{{ c.hiresCount }}</b> наймов</div>
          <div class="stat">★ {{ c.ratingAvg ? c.ratingAvg.toFixed(1) : '—' }} ({{ c.ratingCount }})</div>
          <div class="stat" v-if="c.paymentVerified">платежи подтверждены</div>
        </div>
      </div>
    </div>

    <div class="container-content">
      <div class="about-company">
        <div class="about-company_head">О компании</div>
        <div class="about-company_content">{{ c.about || '—' }}</div>
        <a v-if="c.website" :href="c.website" target="_blank" class="site">{{ c.website }}</a>
      </div>

      <div class="stack" v-if="reviews.length">
        <div class="stack-head"><h3>Отзывы фрилансеров</h3></div>
        <div class="reviews">
          <div class="review" v-for="r in reviews" :key="r.id">
            <div class="stars">{{ '★'.repeat(r.rating) }}{{ '☆'.repeat(5 - r.rating) }}</div>
            <p v-if="r.comment">{{ r.comment }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const api = useApi()
const id = useRoute().params.id as string

const { data: c } = await useAsyncData(`cln-${id}`, () =>
  api.get(`/profiles/clients/${id}`).then((r) => r.data).catch(() => null),
)
const { data: rev } = await useAsyncData(`cln-rev-${id}`, () =>
  api.get(`/users/${id}/reviews`).then((r) => r.data).catch(() => ({ reviews: [] })),
)
const reviews = computed(() => rev.value?.reviews || [])
</script>

<style scoped>
@import './assets/css/company-[id].css';

* { font-family: Inter, sans-serif; }
h1, h2, h3, h4, p { margin: 0; padding: 0; }

.company-contacts { background: #fff; border-radius: 40px; padding: 40px; color: #000; }
.contacts-info { display: flex; flex-direction: column; gap: 12px; }
.stat { font-size: 16px; font-weight: 500; }
.about-company_content { white-space: break-spaces; font-size: 18px; font-weight: 500; }
.site { color: #4e92f8; font-weight: 600; margin-top: 10px; display: inline-block; }
.reviews { display: flex; flex-direction: column; gap: 14px; width: 100%; }
.review { border: 1px solid #e5e7eb; border-radius: 16px; padding: 14px; }
.stars { color: #f59e0b; }
.review p { font-size: 14px; font-weight: 500; margin-top: 4px; }
</style>
