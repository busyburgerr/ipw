<template>
  <NuxtLink :to="`/freelancers/${f.userId}`" :class="$style.resume_block" class="card-link">
    <div :class="$style.resume_header_block">
      <div :class="$style.resume_userInfo_block">
        <div :class="$style.resume_logo" />
        <div :class="$style.resume_content_block">
          <div :class="$style.resume_nameTag_block">
            <div :class="$style.resume_name">{{ f.displayName || 'Фрилансер' }}</div>
            <div :class="$style.resume_tag">{{ ru(f.availability) }}</div>
          </div>
          <div :class="$style.resume_secondInfo_block">
            <div :class="$style.resume_levelDirection">{{ f.headline || '—' }}</div>
            <div :class="$style.resume_locationWorkTime">
              {{ f.location || '' }}<span v-if="f.location"> · </span>★ {{ f.ratingAvg ? f.ratingAvg.toFixed(1) : '—' }} ({{ f.ratingCount }})
            </div>
          </div>
        </div>
      </div>
      <div :class="$style.resume_salary">{{ f.hourlyRateCents ? money(f.hourlyRateCents, f.currency) + ' / час' : '' }}</div>
    </div>
    <div :class="$style.resume_description_block">{{ f.bio }}</div>
    <div class="skills">
      <span v-for="s in f.skills" :key="s.id" class="skill">{{ s.name }}</span>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
defineProps<{ f: any }>()
</script>

<style module>
@import '@/assets/css/resume.module.css';
</style>

<style scoped>
.card-link { text-decoration: none; }
.skills { display: flex; flex-wrap: wrap; gap: 6px; }
.skill {
  background: #eef1f4; color: #374151; padding: 3px 12px;
  border-radius: 999px; font-size: 12px; font-weight: 500;
}
</style>
