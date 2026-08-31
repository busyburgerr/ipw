// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },
  css: ['@/assets/css/style.css'],
  ssr: true,
  devServer: { host: '127.0.0.1', port: 3000 },
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:5000/api/v1',
    },
  },
  app: {
    head: {
      title: 'IPW — фриланс-биржа',
      htmlAttrs: { lang: 'ru' },
    },
  },
})
