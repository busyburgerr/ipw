export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuth()
  await auth.init()
  if (!auth.isAuthed.value) {
    return navigateTo(`/auth?next=${encodeURIComponent(to.fullPath)}`)
  }
})
