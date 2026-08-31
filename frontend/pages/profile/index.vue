<template>
  <div class="container" v-if="user">
    <div class="container-head">
      <div class="main-profile">
        <div class="user-photo" />
        <div class="user-dataBtn_block">
          <div class="user-info static">
            <div class="user-nameEdit_block">
              <h2 class="user-name">{{ user.displayName || user.email }}</h2>
            </div>
            <div class="user-tagAgeGender_block">
              <div class="user-tag">{{ user.email }}</div>
            </div>
            <div class="user-locationStatus_block">
              <div class="user-location">
                <span v-if="user.isFreelancer">фрилансер</span>
                <span v-if="user.isFreelancer && user.isClient"> · </span>
                <span v-if="user.isClient">заказчик</span>
                <span v-if="user.isAdmin"> · админ</span>
              </div>
            </div>
          </div>
          <div class="btn-block">
            <NuxtLink class="btn employee" to="/requests" v-if="user.isClient">Мои проекты</NuxtLink>
            <NuxtLink class="btn employee" to="/my-proposals" v-if="user.isFreelancer">Мои отклики</NuxtLink>
            <NuxtLink class="btn employee" to="/contracts">Контракты</NuxtLink>
          </div>
        </div>
      </div>
    </div>

    <div class="container-content">
      <!-- freelancer profile -->
      <div class="about-me_block" v-if="user.isFreelancer">
        <div class="about-me_head"><h2>Профиль фрилансера</h2></div>
        <form class="edit-form" @submit.prevent="saveFreelancer">
          <input v-model="frl.headline" placeholder="Заголовок (напр. Backend-разработчик на Go)" />
          <textarea v-model="frl.bio" rows="3" placeholder="О себе" />
          <div class="row">
            <input v-model="frl.hourlyRate" type="number" min="0" placeholder="Ставка, ₽/час" />
            <select v-model="frl.availability">
              <option value="available">открыт к работе</option>
              <option value="limited">частично занят</option>
              <option value="unavailable">не ищу</option>
            </select>
            <input v-model="frl.location" placeholder="Локация" />
          </div>
          <select v-model="frl.primaryCategoryId">
            <option value="">— основная категория —</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <div class="skills-pick" v-if="catSkills.length">
            <button
              v-for="s in catSkills"
              :key="s.id"
              type="button"
              class="skill-chip"
              :class="{ on: frl.skillIds.has(s.id) }"
              @click="frl.skillIds.has(s.id) ? frl.skillIds.delete(s.id) : frl.skillIds.add(s.id)"
            >{{ s.name }}</button>
          </div>
          <input v-model="frl.languages" placeholder="Языки через запятую (Русский, English)" />
          <p v-if="frlMsg" :class="frlErr ? 'err' : 'ok'">{{ frlMsg }}</p>
          <button class="btn save" :disabled="frlSaving">Сохранить</button>
        </form>
      </div>

      <!-- client profile -->
      <div class="about-me_block" v-if="user.isClient">
        <div class="about-me_head"><h2>Профиль заказчика</h2></div>
        <form class="edit-form" @submit.prevent="saveClient">
          <input v-model="cln.companyName" placeholder="Компания" />
          <textarea v-model="cln.about" rows="3" placeholder="О компании" />
          <div class="row">
            <input v-model="cln.website" placeholder="Сайт" />
            <input v-model="cln.location" placeholder="Локация" />
          </div>
          <p v-if="clnMsg" :class="clnErr ? 'err' : 'ok'">{{ clnMsg }}</p>
          <button class="btn save" :disabled="clnSaving">Сохранить</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
const api = useApi()
const { user } = useAuth()

const { data: cats } = await useAsyncData('cab-cats', () =>
  api.get('/catalog/categories').then((r) => r.data),
)
const categories = computed(() => cats.value?.categories || [])

// --- freelancer ---
const frl = reactive({
  headline: '', bio: '', hourlyRate: '', availability: 'available',
  primaryCategoryId: '', location: '', languages: '',
  skillIds: new Set<string>(),
})
const catSkills = ref<any[]>([])
watch(
  () => frl.primaryCategoryId,
  async (id) => {
    const slug = categories.value.find((c: any) => c.id === id)?.slug
    catSkills.value = slug ? (await api.get(`/catalog/skills?category=${slug}`)).data.skills : []
  },
)

if (user.value?.isFreelancer) {
  try {
    const { data } = await api.get('/me/freelancer-profile')
    Object.assign(frl, {
      headline: data.headline, bio: data.bio,
      hourlyRate: data.hourlyRateCents ? data.hourlyRateCents / 100 : '',
      availability: data.availability === 'unknown' ? 'available' : data.availability,
      primaryCategoryId: data.primaryCategoryId || '',
      location: data.location, languages: (data.languages || []).join(', '),
      skillIds: new Set((data.skills || []).map((s: any) => s.id)),
    })
  } catch { /* no profile yet */ }
}

const frlSaving = ref(false)
const frlMsg = ref('')
const frlErr = ref(false)
async function saveFreelancer() {
  frlSaving.value = true
  frlMsg.value = ''
  try {
    await api.put('/me/freelancer-profile', {
      headline: frl.headline, bio: frl.bio,
      hourlyRateCents: toCents(frl.hourlyRate || 0),
      availability: frl.availability,
      primaryCategoryId: frl.primaryCategoryId || '',
      location: frl.location,
      languages: frl.languages.split(',').map((s) => s.trim()).filter(Boolean),
    })
    await api.put('/me/freelancer-profile/skills', { skillIds: [...frl.skillIds] })
    frlErr.value = false
    frlMsg.value = 'Сохранено'
  } catch (e: any) {
    frlErr.value = true
    frlMsg.value = e.message
  } finally {
    frlSaving.value = false
  }
}

// --- client ---
const cln = reactive({ companyName: '', about: '', website: '', location: '' })
if (user.value?.isClient) {
  try {
    const { data } = await api.get('/me/client-profile')
    Object.assign(cln, {
      companyName: data.companyName, about: data.about,
      website: data.website, location: data.location,
    })
  } catch { /* none yet */ }
}
const clnSaving = ref(false)
const clnMsg = ref('')
const clnErr = ref(false)
async function saveClient() {
  clnSaving.value = true
  clnMsg.value = ''
  try {
    await api.put('/me/client-profile', { ...cln })
    clnErr.value = false
    clnMsg.value = 'Сохранено'
  } catch (e: any) {
    clnErr.value = true
    clnMsg.value = e.message
  } finally {
    clnSaving.value = false
  }
}
</script>

<style scoped>
@import './assets/css/profile.css';

h1, h2, h3, h4, p { margin: 0; padding: 0; }
.user-info.static { height: auto; gap: 10px; color: #fff; }
.main-profile { color: #fff; }

.edit-form { display: flex; flex-direction: column; gap: 12px; width: 100%; }
.edit-form input,
.edit-form textarea,
.edit-form select {
  border: 1px solid rgba(0, 0, 0, 0.25);
  border-radius: 12px;
  padding: 12px;
  font-family: Inter, sans-serif;
  font-size: 14px;
}
.row { display: flex; gap: 12px; flex-wrap: wrap; }
.row > * { flex: 1; min-width: 140px; }
.skills-pick { display: flex; flex-wrap: wrap; gap: 8px; }
.skill-chip {
  border: 1px solid #cbd5e1; border-radius: 999px; padding: 5px 14px;
  font-size: 13px; font-weight: 500; background: #fff; cursor: pointer;
}
.skill-chip.on { background: #4e92f8; color: #fff; border-color: #4e92f8; }
.save {
  align-self: flex-start; padding: 10px 28px; border-radius: 10px;
  background: #198754; color: #fff; font-weight: 600; border: none; cursor: pointer;
}
.err { color: #b91c1c; font-weight: 500; }
.ok { color: #198754; font-weight: 500; }
.btn.employee { text-decoration: none; }
</style>
