import { ref } from 'vue'

/** Wraps an async action with loading + error state for forms and buttons. */
export function useSubmit() {
  const loading = ref(false)
  const error = ref('')

  async function run<T>(fn: () => Promise<T>): Promise<T | undefined> {
    loading.value = true
    error.value = ''
    try {
      return await fn()
    } catch (e: any) {
      error.value = e?.message || 'Ошибка'
      return undefined
    } finally {
      loading.value = false
    }
  }

  return { loading, error, run }
}
