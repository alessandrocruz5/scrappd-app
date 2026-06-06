# Scrappd Mobile (Expo)

The Scrappd React Native app, built with Expo + expo-router. It talks to
Supabase directly: Auth for sign-in/up, Postgres (with RLS) for data, and
Storage for shape cutouts. There is no custom API server.

This app replaces the retired Flutter app (`scrappd-mobile/`).

## Stack

- **Expo** (SDK 56) + **expo-router** (file-based routing, typed routes)
- **@supabase/supabase-js** — session persisted to AsyncStorage with auto-refresh
- **@tanstack/react-query** — server state
- **zustand** — auth/client state
- **@scrappd/shared-types** — generated Supabase types (workspace package)

## Setup

1. From the repo root, install deps: `pnpm install`
2. Start the local Supabase stack: `pnpm --filter @scrappd/supabase db:start`
3. Create the env file: `cp apps/mobile/.env.example apps/mobile/.env`
   - Copy the API URL + anon key from the `db:start` output (or `supabase status`).
   - **Physical device / Expo Go:** replace `127.0.0.1` with your computer's
     LAN IP so the phone can reach the local stack.
4. Run the app: `pnpm --filter mobile dev`
5. Open in **Expo Go** on your iPhone (scan the QR), or press `i` for the iOS
   simulator.

## Environment variables

The app reads **build-time** public vars (the `EXPO_PUBLIC_` prefix inlines them
into the JS bundle); they are read in `src/lib/env.ts`, which throws at startup
if the required Supabase pair is missing:

| Variable                        | Required | Purpose                                          |
| ------------------------------- | -------- | ------------------------------------------------ |
| `EXPO_PUBLIC_SUPABASE_URL`      | yes      | Supabase project API URL                         |
| `EXPO_PUBLIC_SUPABASE_ANON_KEY` | yes      | Supabase anon/publishable key (RLS-safe)         |
| `EXPO_PUBLIC_SENTRY_DSN`        | no       | Sentry DSN; enables crash reporting when present |

Source maps for production builds also use three **secret**, build-time-only
vars (read by the Sentry Metro/Expo plugin, **not** inlined into the bundle):

| Variable            | Purpose                                             |
| ------------------- | --------------------------------------------------- |
| `SENTRY_ORG`        | Sentry org slug                                     |
| `SENTRY_PROJECT`    | Sentry project slug                                 |
| `SENTRY_AUTH_TOKEN` | Auth token used to upload source maps (keep secret) |

How they are supplied per target:

- **Local dev:** `apps/mobile/.env` (copy from `.env.example`). Git-ignored.
- **Live / hosted Supabase:** point `.env` at the hosted project
  (`https://gujldqovvhjbctzelark.supabase.co` + its anon key). The anon key is
  safe to ship in the client — Row Level Security, not key secrecy, protects the
  data.
- **Web (Vercel):** set both vars under Project → Settings → Environment
  Variables so they are present at `expo export` build time.
- **Native (EAS):** set them in the `eas.json` build profile `env` block, or as
  EAS project secrets, so they are baked into each build.

`.env` is git-ignored (root `.gitignore` ignores `.env`); only `.env.example`
is committed. Never commit real values.

## Scripts

| Command                          | Description                          |
| -------------------------------- | ------------------------------------ |
| `pnpm --filter mobile dev`       | Start the Expo dev server            |
| `pnpm --filter mobile lint`      | Lint with eslint-config-expo         |
| `pnpm --filter mobile typecheck` | Type-check with `tsc --noEmit`       |
| `pnpm --filter mobile build`     | Export the JS bundle (`expo export`) |

## Layout

```
app/
  _layout.tsx          Providers (React Query, SafeArea) + auth gate
  index.tsx            Entry redirect based on session
  (auth)/              login, register, forgot-password
  (tabs)/              Cropper, Books, Profile tab shell
src/
  lib/supabase.ts      Supabase client (AsyncStorage + auto-refresh)
  lib/query-client.ts  React Query client
  lib/env.ts           Validated EXPO_PUBLIC_* config
  stores/auth-store.ts zustand auth store (mirrors the old Flutter AuthProvider)
  cropper/             Instant shape cropper (Skia, client-side)
  editor/              Skia page editor + client-side export (export-page.ts)
  books/               Books / Pages / Items data + UI
  components/          Shared UI (buttons, fields, splash)
  theme/colors.ts      Brand palette carried over from the Flutter app
```

## Page export

Pages are exported entirely on-device — there is no server render endpoint. The
editor snapshots the composed page view (the Skia background plus the cutout
overlays) to a high-res PNG with `react-native-view-shot`, then saves it to the
photo library (`expo-media-library`) and offers the system share sheet
(`expo-sharing`). This is the React Native port of the old
`page_export_service.dart`, which downloaded a server-rendered image. See
`src/editor/export-page.ts`.

## Crash reporting (Sentry)

Production observability runs through Sentry (`@sentry/react-native`), wired up
in `src/lib/sentry.ts` and initialised once in `app/_layout.tsx`. It is
**opt-in**: nothing is sent unless `EXPO_PUBLIC_SENTRY_DSN` is set, so local dev
and Expo Go stay silent.

- **Unhandled render errors** are caught by a root `Sentry.ErrorBoundary` in
  `app/_layout.tsx`, which reports the crash and shows the friendly
  `src/components/error-fallback.tsx` UI (with a "Try again" reset) instead of a
  blank screen. The root component is wrapped with `Sentry.wrap` for navigation
  and touch breadcrumbs.
- **Handled errors** in the key flows are reported via `captureHandledError`
  with a `feature` tag: cutout upload (`cropper.upload`), editor persistence
  (`editor.persist.*`, the optimistic mutations that otherwise roll back
  silently), and page export (`editor.export`).
- **Source maps:** `metro.config.js` uses Sentry's `getSentryExpoConfig` and
  `app.json` includes the `@sentry/react-native/expo` plugin, so production
  web/native builds emit source maps and upload them when `SENTRY_AUTH_TOKEN` /
  `SENTRY_ORG` / `SENTRY_PROJECT` are present (the upload is skipped otherwise).

To verify end-to-end, run a build with a real DSN, trigger a test exception
(e.g. throw in a button handler), and confirm it lands in the Sentry dashboard
with readable (de-minified) stack frames.

## Auth flow

`app/_layout.tsx` hydrates the persisted session and subscribes to
`supabase.auth.onAuthStateChange`. An `AuthGate` routes between the `(auth)`
stack and the `(tabs)` shell based on session state — the equivalent of the old
Flutter `root_screen.dart` / `main_shell.dart`.

Email confirmation is disabled in the local Supabase config, so sign-up signs
the user straight in. In **production** confirmations are enabled (see
`packages/supabase/config.toml`), so sign-up surfaces a "check your email"
notice instead — which requires an SMTP sender to be wired on the hosted
project. The `scrappd://` deep-link scheme is allow-listed for the
password-reset / confirmation callbacks (the reset route lands in P4).
