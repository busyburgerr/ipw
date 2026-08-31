import { defineStore } from 'pinia'

export interface User {
  id: string
  email: string
  displayName: string
  avatarUrl: string
  isFreelancer: boolean
  isClient: boolean
  isAdmin: boolean
  emailVerified: boolean
}

interface AuthResponse {
  user: User
  accessToken: string
  refreshToken: string
  expiresIn: number
}

const ACCESS_KEY = 'ipw.access'
const REFRESH_KEY = 'ipw.refresh'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    accessToken: '' as string,
    refreshToken: '' as string,
    ready: false,
  }),
  getters: {
    isAuthed: (s) => !!s.accessToken && !!s.user,
  },
  actions: {
    loadTokens() {
      if (import.meta.client) {
        this.accessToken = localStorage.getItem(ACCESS_KEY) || ''
        this.refreshToken = localStorage.getItem(REFRESH_KEY) || ''
      }
    },
    persist() {
      if (import.meta.client) {
        localStorage.setItem(ACCESS_KEY, this.accessToken)
        localStorage.setItem(REFRESH_KEY, this.refreshToken)
      }
    },
    setSession(r: AuthResponse) {
      this.user = r.user
      this.accessToken = r.accessToken
      this.refreshToken = r.refreshToken
      this.persist()
    },
    clear() {
      this.user = null
      this.accessToken = ''
      this.refreshToken = ''
      if (import.meta.client) {
        localStorage.removeItem(ACCESS_KEY)
        localStorage.removeItem(REFRESH_KEY)
      }
    },
    async register(body: {
      email: string
      password: string
      displayName: string
      asFreelancer: boolean
      asClient: boolean
    }) {
      const { $apiRaw } = useNuxtApp()
      const r = await $apiRaw<AuthResponse>('/auth/register', { method: 'POST', body })
      this.setSession(r)
    },
    async login(email: string, password: string) {
      const { $apiRaw } = useNuxtApp()
      const r = await $apiRaw<AuthResponse>('/auth/login', { method: 'POST', body: { email, password } })
      this.setSession(r)
    },
    async refresh() {
      if (!this.refreshToken) throw new Error('no refresh token')
      const { $apiRaw } = useNuxtApp()
      const r = await $apiRaw<AuthResponse>('/auth/refresh', {
        method: 'POST',
        body: { refreshToken: this.refreshToken },
      })
      this.setSession(r)
    },
    async fetchMe() {
      const { $apiRaw } = useNuxtApp()
      this.user = await $apiRaw<User>('/auth/me', {
        headers: { Authorization: `Bearer ${this.accessToken}` },
      })
    },
    async logout() {
      const { $apiRaw } = useNuxtApp()
      try {
        await $apiRaw('/auth/logout', { method: 'POST', body: { refreshToken: this.refreshToken } })
      } catch {
        /* ignore */
      }
      this.clear()
    },
    async init() {
      if (this.ready) return
      this.loadTokens()
      if (this.accessToken) {
        try {
          await this.fetchMe()
        } catch {
          try {
            await this.refresh()
            await this.fetchMe()
          } catch {
            this.clear()
          }
        }
      }
      this.ready = true
    },
  },
})
