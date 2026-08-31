import { useAuthStore } from '~/stores/auth'

/**
 * Provides two helpers:
 *   $apiRaw — fetch against the API base with the shared error shape unwrapped.
 *   $api    — same, plus Bearer auth and a one-shot refresh-and-retry on 401.
 */
export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig()
  const base = config.public.apiBase

  const parseError = (e: any) => {
    const msg = e?.data?.error?.message || e?.data?.message || e?.message || 'Ошибка запроса'
    const err = new Error(msg) as Error & { statusCode?: number; code?: string }
    err.statusCode = e?.statusCode || e?.response?.status
    err.code = e?.data?.error?.code
    return err
  }

  const apiRaw = async <T>(path: string, opts: any = {}): Promise<T> => {
    try {
      return await $fetch<T>(base + path, opts)
    } catch (e) {
      throw parseError(e)
    }
  }

  const api = async <T>(path: string, opts: any = {}): Promise<T> => {
    const auth = useAuthStore()
    const run = (token: string) =>
      apiRaw<T>(path, {
        ...opts,
        headers: { ...(opts.headers || {}), ...(token ? { Authorization: `Bearer ${token}` } : {}) },
      })

    try {
      return await run(auth.accessToken)
    } catch (e: any) {
      if (e?.statusCode === 401 && auth.refreshToken) {
        try {
          await auth.refresh()
        } catch {
          auth.clear()
          throw e
        }
        return await run(auth.accessToken)
      }
      throw e
    }
  }

  return { provide: { api, apiRaw } }
})
