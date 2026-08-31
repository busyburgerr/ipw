/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './components/**/*.{vue,js,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './app.vue',
    './error.vue',
  ],
  theme: {
    extend: {
      colors: {
        ink: '#1b2430',
        // blue accent (buttons inside cards, links, "Откликнуться")
        brand: {
          DEFAULT: '#4a8fe0',
          600: '#3f80e0',
          700: '#356fc4',
        },
        // dark navy (primary CTA — "Зарегистрироваться")
        navy: {
          DEFAULT: '#2f3e57',
          700: '#263248',
        },
        // page gradient stops
        sky: {
          from: '#6fa8ea',
          to: '#3f80e0',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        card: '0 10px 30px -12px rgba(20, 40, 80, 0.25)',
      },
    },
  },
  plugins: [],
}
