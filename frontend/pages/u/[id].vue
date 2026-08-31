<script setup lang="ts">
const route = useRoute()
const { $apiRaw } = useNuxtApp()
const id = route.params.id as string

const { data: frl } = await useAsyncData(`u-frl-${id}`, () =>
  $apiRaw<any>(`/profiles/freelancers/${id}`).catch(() => null),
)
const { data: cln } = await useAsyncData(`u-cln-${id}`, () =>
  $apiRaw<any>(`/profiles/clients/${id}`).catch(() => null),
)
const { data: reviews } = await useAsyncData(`u-reviews-${id}`, () =>
  $apiRaw<{ reviews: any[] }>(`/users/${id}/reviews`).catch(() => ({ reviews: [] })),
)
</script>

<template>
  <div class="flex flex-col gap-6">
    <EmptyState v-if="!frl && !cln" title="Профиль не найден" />

    <div v-if="frl" class="card flex flex-col gap-3">
      <div class="flex items-center justify-between">
        <div>
          <h1>{{ frl.displayName || 'Фрилансер' }}</h1>
          <p class="text-slate-600">{{ frl.headline }}</p>
        </div>
        <RatingStars :value="frl.ratingAvg" :count="frl.ratingCount" />
      </div>
      <p v-if="frl.bio" class="whitespace-pre-wrap text-sm text-slate-700">{{ frl.bio }}</p>
      <div class="flex flex-wrap gap-x-6 gap-y-1 text-sm text-slate-500">
        <span v-if="frl.hourlyRateCents">{{ money(frl.hourlyRateCents, frl.currency) }} / час</span>
        <span>{{ statusRu(frl.availability) }}</span>
        <span v-if="frl.location">{{ frl.location }}</span>
        <span>{{ frl.jobsCompleted }} завершённых</span>
      </div>
      <div v-if="frl.skills?.length" class="flex flex-wrap gap-2">
        <span v-for="s in frl.skills" :key="s.id" class="pill bg-slate-100 text-slate-700">{{ s.name }}</span>
      </div>
    </div>

    <div v-if="cln" class="card flex flex-col gap-2">
      <div class="flex items-center justify-between">
        <h2>{{ cln.companyName || cln.displayName || 'Заказчик' }}</h2>
        <RatingStars :value="cln.ratingAvg" :count="cln.ratingCount" size="sm" />
      </div>
      <p v-if="cln.about" class="text-sm text-slate-700">{{ cln.about }}</p>
      <p class="text-sm text-slate-500">{{ cln.hiresCount }} наймов<span v-if="cln.location"> · {{ cln.location }}</span></p>
    </div>

    <section v-if="reviews?.reviews?.length" class="flex flex-col gap-2">
      <h2>Отзывы</h2>
      <div v-for="r in reviews.reviews" :key="r.id" class="card flex flex-col gap-1">
        <RatingStars :value="r.rating" size="sm" />
        <p v-if="r.comment" class="text-sm text-slate-700">{{ r.comment }}</p>
        <p class="text-xs text-slate-400">{{ date(r.publishedAt) }}</p>
      </div>
    </section>
  </div>
</template>
