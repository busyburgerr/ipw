import Cookies from 'js-cookie'
import { TOKEN_COOKIE, REFRESH_COOKIE } from './useApi'

export interface AuthUser {
  id: string
  email: string
  displayName: string
  avatarUrl: string
  isFreelancer: boolean
  isClient: boolean
  isAdmin: boolean
}

export function useAuth() {
  const user = useState<AuthUser | null>('auth.user', () => null)
  const ready = useState<boolean>('auth.ready', () => false)
  const api = useApi()

  const isAuthed = computed(() => !!user.value)

  function storeTokens(r: { accessToken: string; refreshToken: string }) {
    Cookies.set(TOKEN_COOKIE, r.accessToken, { sameSite: 'lax' })
    Cookies.set(REFRESH_COOKIE, r.refreshToken, { sameSite: 'lax' })
  }

  async function login(email: string, password: string) {
    const { data } = await api.post('/auth/login', { email, password })
    storeTokens(data)
    user.value = data.user
  }

  async function register(body: {
    displayName: string
    email: string
    password: string
    asFreelancer: boolean
    asClient: boolean
  }) {
    const { data } = await api.post('/auth/register', body)
    storeTokens(data)
    user.value = data.user
  }

  async function fetchMe() {
    const { data } = await api.get('/auth/me')
    user.value = data
  }

  async function logout() {
    try {
      await api.post('/auth/logout', { refreshToken: Cookies.get(REFRESH_COOKIE) })
    } catch {
      /* ignore */
    }
    Cookies.remove(TOKEN_COOKIE)
    Cookies.remove(REFRESH_COOKIE)
    user.value = null
  }

  async function init() {
    if (ready.value) return
    if (Cookies.get(TOKEN_COOKIE)) {
      try {
        await fetchMe()
      } catch {
        Cookies.remove(TOKEN_COOKIE)
        Cookies.remove(REFRESH_COOKIE)
      }
    }
    ready.value = true
  }

  return { user, isAuthed, ready, login, register, logout, fetchMe, init }
}
