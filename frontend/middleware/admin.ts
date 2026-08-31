import { useAuthStore } from '~/stores/auth'

export default defineNuxtRouteMiddleware(async () => {
  const auth = useAuthStore()
  await auth.init()
  if (!auth.isAuthed) return navigateTo('/auth/login')
  if (!auth.user?.isAdmin) return navigateTo('/')
})
