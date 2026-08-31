import { useAuthStore } from '~/stores/auth'

// Restore the session on first client load.
export default defineNuxtPlugin(async () => {
  await useAuthStore().init()
})
