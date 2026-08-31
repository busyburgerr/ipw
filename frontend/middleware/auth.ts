import { useAuthStore } from '~/stores/auth'

export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  await auth.init()
  if (!auth.isAuthed) {
    return navigateTo(`/auth/login?next=${encodeURIComponent(to.fullPath)}`)
  }
})
