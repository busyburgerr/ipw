<template>
  <div class="container" v-if="f">
    <div class="container-head">
      <div class="profile-card">
        <div class="pc-photo" />
        <div class="pc-info">
          <div class="pc-name">{{ f.displayName || 'Фрилансер' }}</div>
          <div class="pc-headline">{{ f.headline }}</div>
          <div class="pc-meta">
            <span v-if="f.hourlyRateCents">{{ money(f.hourlyRateCents, f.currency) }} / час</span>
            <span>{{ ru(f.availability) }}</span>
            <span v-if="f.location">{{ f.location }}</span>
            <span>★ {{ f.ratingAvg ? f.ratingAvg.toFixed(1) : '—' }} ({{ f.ratingCount }})</span>
          </div>
        </div>
      </div>
    </div>

    <div class="container-content">
      <div class="resume">
        <div class="resume-head"><h3>О себе</h3></div>
        <div class="resume-body">{{ f.bio || '—' }}</div>
      </div>

      <div class="stack">
        <div class="stack-head"><h3>Навыки</h3></div>
        <div class="stack-body">
          <span class="chip" v-for="s in f.skills" :key="s.id">{{ s.name }}</span>
          <span v-if="!f.skills?.length" class="muted">не указаны</span>
        </div>
      </div>

      <div class="stack" v-if="reviews.length">
        <div class="stack-head"><h3>Отзывы</h3></div>
        <div class="reviews">
          <div class="review" v-for="r in reviews" :key="r.id">
            <div class="stars">{{ '★'.repeat(r.rating) }}{{ '☆'.repeat(5 - r.rating) }}</div>
            <p v-if="r.comment">{{ r.comment }}</p>
            <span class="muted">{{ fmtDate(r.publishedAt) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const api = useApi()
const id = useRoute().params.id as string

const { data: f } = await useAsyncData(`frl-${id}`, () =>
  api.get(`/profiles/freelancers/${id}`).then((r) => r.data).catch(() => null),
)
const { data: rev } = await useAsyncData(`frl-rev-${id}`, () =>
  api.get(`/users/${id}/reviews`).then((r) => r.data).catch(() => ({ reviews: [] })),
)
const reviews = computed(() => rev.value?.reviews || [])
</script>

<style scoped>
@import '../../assets/css/resumeID.css';

h1, h2, h3, h4, p { margin: 0; padding: 0; }

.profile-card {
  display: flex; gap: 24px; align-items: center;
  padding: 40px; border-radius: 40px; background: #1f1f1f; color: #fff; width: 100%;
}
.pc-photo { width: 160px; height: 160px; border-radius: 50%; background: #d9d9d9; flex-shrink: 0; }
.pc-name { font-size: 24px; font-weight: 700; }
.pc-headline { margin-top: 6px; color: rgba(255, 255, 255, 0.8); font-weight: 500; }
.pc-meta { margin-top: 14px; display: flex; flex-wrap: wrap; gap: 16px; font-size: 14px; font-weight: 500; color: rgba(255, 255, 255, 0.75); }

.resume-body { font-size: 16px; font-weight: 500; white-space: break-spaces; }
.stack-body { display: flex; flex-wrap: wrap; gap: 10px; }
.chip { background: #d9d9d9; border-radius: 10px; padding: 5px 20px; font-size: 16px; font-weight: 500; }
.muted { color: #9ca3af; font-weight: 500; }
.reviews { display: flex; flex-direction: column; gap: 14px; width: 100%; }
.review { border: 1px solid #e5e7eb; border-radius: 16px; padding: 14px; display: flex; flex-direction: column; gap: 6px; }
.stars { color: #f59e0b; }
.review p { font-size: 14px; font-weight: 500; }
</style>
