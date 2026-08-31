import axios, { type AxiosRequestConfig } from 'axios'
import Cookies from 'js-cookie'

export const TOKEN_COOKIE = 'ipw'
export const REFRESH_COOKIE = 'ipw_refresh'

/**
 * Shared axios instance pointed at the marketplace API. Attaches the bearer
 * token and, on a 401, tries a one-shot refresh before failing.
 */
export function useApi() {
  const base = useRuntimeConfig().public.apiBase

  const client = axios.create({ baseURL: base })

  client.interceptors.request.use((cfg) => {
    const token = Cookies.get(TOKEN_COOKIE)
    if (token) cfg.headers.Authorization = `Bearer ${token}`
    return cfg
  })

  let refreshing: Promise<void> | null = null

  const refresh = async () => {
    const rt = Cookies.get(REFRESH_COOKIE)
    if (!rt) throw new Error('no refresh token')
    const { data } = await axios.post(`${base}/auth/refresh`, { refreshToken: rt })
    Cookies.set(TOKEN_COOKIE, data.accessToken, { sameSite: 'lax' })
    Cookies.set(REFRESH_COOKIE, data.refreshToken, { sameSite: 'lax' })
  }

  client.interceptors.response.use(
    (r) => r,
    async (error) => {
      const cfg = error.config as AxiosRequestConfig & { _retried?: boolean }
      if (error.response?.status === 401 && !cfg._retried && Cookies.get(REFRESH_COOKIE)) {
        cfg._retried = true
        try {
          refreshing = refreshing || refresh().finally(() => (refreshing = null))
          await refreshing
          return client(cfg)
        } catch {
          Cookies.remove(TOKEN_COOKIE)
          Cookies.remove(REFRESH_COOKIE)
        }
      }
      const msg =
        error.response?.data?.error?.message || error.response?.data?.message || error.message
      return Promise.reject(new Error(msg))
    },
  )

  return client
}
