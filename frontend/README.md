# IPW frontend

The original "IT Professionals Work" Nuxt design, rewired to the freelance
marketplace API.

- Nuxt 3, plain axios + js-cookie (no Pinia, no socket.io)
- Auth: access token in the `ipw` cookie, refresh token in `ipw_refresh`,
  transparent refresh-on-401 in `composables/useApi.ts`
- The old blue-gradient look is kept: `assets/css/style.css` plus the
  per-page CSS under `pages/**/assets/css/`

## Run

```sh
# backend must be running (see ../backend/README.md)
npm install
npm run dev            # http://localhost:3000
```

Override the API base: `NUXT_PUBLIC_API_BASE=https://api.example.com/api/v1 npm run dev`

## Route map (old -> new)

| route | was | now |
|---|---|---|
| `/projects` | вакансии | список проектов |
| `/projects/:id` | вакансия | проект + отклик / список откликов |
| `/projects/new`, `/projects/:id/edit` | создание вакансии | создание / редактирование проекта |
| `/freelancers`, `/freelancers/:id` | резюме | каталог фрилансеров / профиль |
| `/clients/:id` | компания | профиль заказчика |
| `/profile` | профиль + резюме | кабинет: профиль фрилансера и заказчика |
| `/requests` | отклики (HR) | мои проекты (заказчик) |
| `/my-proposals` | — | мои отклики (фрилансер) |
| `/contracts`, `/contracts/:id` | — | контракты, этапы, escrow, отзывы, споры |
| `/wallet` | — | баланс и выплаты |
| `/admin` | — | арбитраж споров |
