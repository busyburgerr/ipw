# ipw frontend

Nuxt 3 client for the freelance marketplace.

## Stack

- Nuxt 3 · Vue 3 · Pinia · Tailwind CSS
- Talks to the Go backend at `NUXT_PUBLIC_API_BASE` (default
  `http://localhost:5000/api/v1`)
- Auth: access token in memory + `localStorage`, refresh token rotated
  transparently by `plugins/api.ts` on 401

## Run

```sh
# backend must be running first (see ../backend/README.md)
npm install
npm run dev            # http://localhost:3000
```

Override the API URL:

```sh
NUXT_PUBLIC_API_BASE=https://api.example.com/api/v1 npm run dev
```

## Layout

```
stores/auth.ts        session + user
plugins/api.ts        $api (auth + refresh-retry) / $apiRaw
plugins/auth.client   restores the session on load
middleware/           auth, admin route guards
utils/format.ts       money (cents), dates, RU status labels
components/           StatusPill, RatingStars, ProjectCard, ProjectForm, …
pages/
  index                       landing
  auth/{login,register}
  projects/{index,new,[id],[id]/edit}
  dashboard                   role-aware home
  me/{projects,proposals,contracts,wallet}
  contracts/[id]              milestones, funding, reviews, disputes
  profile/edit                freelancer + client profile
  u/[id]                      public profile
  admin/disputes              arbiter queue
```

## Notes

- `nuxt.config.ts` pins the dev server to `127.0.0.1` (Windows + IPv6
  `localhost` quirk).
- Fund flow: on a stub payment (`LAVA_API_KEY` unset on the backend) the
  contract page auto-confirms the fake payment; with real lava.top it
  opens the payment URL in a new tab.
