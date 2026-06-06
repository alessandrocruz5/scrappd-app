# Deploying Scrappd web to Vercel (P2)

The Expo app exports a static, single-page web build that is hosted on Vercel.
This is the first live Scrappd surface. The build config lives in
[`vercel.json`](../vercel.json) at the repo root.

## What `vercel.json` does

| Field             | Value                                                  | Why                                                                     |
| ----------------- | ------------------------------------------------------ | ----------------------------------------------------------------------- |
| `installCommand`  | `pnpm install --frozen-lockfile`                       | Installs the whole pnpm workspace (the app needs `@scrappd/shared-types`). |
| `buildCommand`    | `pnpm --filter mobile exec expo export --platform web` | Produces the static web bundle.                                         |
| `outputDirectory` | `apps/mobile/dist`                                     | Where `expo export` writes `index.html` + `_expo/` assets.             |
| `rewrites`        | `/(.*) → /index.html`                                  | SPA fallback so client-side (expo-router) routes resolve on hard loads / reload. |

The app builds with `web.output: "single"` (see `apps/mobile/app.json`), so
the export is a single `index.html` SPA — the rewrite above is what makes deep
links and page reloads work.

## One-time Vercel project setup (dashboard)

The repo config is committed; the project itself is created in the Vercel
dashboard (or via `vercel link`) and pointed at this repo:

1. **Create the project** from the `alessandrocruz5/scrappd-app` GitHub repo.
2. **Root Directory:** leave as the repo root (`./`). `vercel.json` handles the
   monorepo build via the `--filter mobile` install/build commands.
3. **Build & Output settings:** these come from `vercel.json`; no overrides
   needed. The package manager (pnpm 10.33.0) and Node (>= 20) are detected from
   `package.json` (`packageManager` / `engines`).
4. **Environment variables** — add both, for the Production (and Preview)
   environments, so they are inlined at `expo export` time:

   | Variable                        | Value                                          |
   | ------------------------------- | ---------------------------------------------- |
   | `EXPO_PUBLIC_SUPABASE_URL`      | `https://gujldqovvhjbctzelark.supabase.co`     |
   | `EXPO_PUBLIC_SUPABASE_ANON_KEY` | the project's anon / `sb_publishable_…` key    |

   The anon key is safe to expose in the client bundle — Row Level Security, not
   key secrecy, protects the data. (When P5/Sentry lands, also add
   `EXPO_PUBLIC_SENTRY_DSN`.)
5. **Deploy.** The first deploy gives a `*.vercel.app` URL.

## Point Supabase auth at the deployed domain

After the first deploy, in the **Supabase dashboard** for project
`gujldqovvhjbctzelark` → **Authentication → URL Configuration**:

- **Site URL:** the deployed web origin, e.g. `https://<project>.vercel.app`.
- **Redirect URLs:** add the same origin (and, once P4 lands, its
  `/reset` route) plus the `scrappd://` native scheme.

Without this, email links (confirmation / password reset) point at localhost.
The committed `packages/supabase/config.toml` only governs the **local** stack;
the hosted project's auth URLs are dashboard-managed.

## Web behaviour notes

- **Cropper falls back to file upload on web.** `expo-camera`'s live capture is
  limited in browsers, so on web the cropper skips the camera entirely and
  drives the same shape → cutout → upload pipeline from an `expo-image-picker`
  file upload. Camera-only UI (live preview, shutter, permission gates) is gated
  behind `Platform.OS === 'web'` checks in `src/cropper/cropper-screen.tsx`.
- Everything else (auth, books/pages CRUD, Skia page editor, export) runs the
  same code path as native via `react-native-web`.

## Verify

Build locally exactly as Vercel will:

```bash
pnpm install --frozen-lockfile
EXPO_PUBLIC_SUPABASE_URL=https://gujldqovvhjbctzelark.supabase.co \
EXPO_PUBLIC_SUPABASE_ANON_KEY=<anon-key> \
  pnpm --filter mobile exec expo export --platform web
# serve the static output to smoke-test SPA routing:
npx serve apps/mobile/dist -s
```

On the deployed URL: sign up / in, run a crop (via the image picker on web),
create a book + page, place an item, export. Reload the page and confirm the
session persists.
