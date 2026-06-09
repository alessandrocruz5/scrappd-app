# Scrappd — Migration Audit & Live-Deployment Plan

## Context

Scrappd was migrated from a **Flutter + Go (REST) + Python ML** stack to an **Expo (React Native) + Supabase** monorepo (see `REVAMP_PLAN.md`). This task audits whether that migration is actually complete, lists what is still missing (including env vars), and lays out a sequenced, **prompt-by-prompt** path to a live deployment — **web first, then iOS and Android**.

**Hosted Supabase decision:** **single project** — reuse the existing `gujldqovvhjbctzelark` as the live backend, with the local Docker stack (`pnpm db:start`) as the dev environment. A dedicated `scrappd-prod` (two-project dev/prod split) is deferred: the account is currently at the Supabase free-tier active-project cap (2), and a split isn't worth blocking launch on pre-users. Revisit once there are real users or a second developer (free a slot or upgrade tier, then promote: existing → staging, new → prod, with per-environment env vars in Vercel/EAS). Before P1 pushes anything, inspect the project (`list_migrations` / `list_tables`) to confirm whether the 4 migrations are already applied and avoid double-applying.

**Audit verdict: the migration is _code-complete_ but _not deployment-ready._** All 7 milestone PRs (#22–#29) are merged to `main`. No Dart/Go/Flutter remnants remain; there are no leftover REST endpoints or old API base URLs. The app talks to Supabase directly, RLS is comprehensive, TypeScript is `strict` with zero `@ts-ignore`, and CI (lint, typecheck, Jest+coverage, Prettier, Expo web export, ML byte-compile) is green. What's missing is everything between "runs on a local stack in Expo Go" and "live in front of real users."

---

## Audit results

### ✅ In place (verified)
- **Monorepo**: pnpm 10.33 + Turborepo; `apps/mobile` (Expo SDK 56, RN 0.85.3, React 19, expo-router), `apps/ml-service` (dormant FastAPI/BiRefNet), `packages/supabase` (CLI project), `packages/shared-types` (generated DB types).
- **Features**: auth (sign in/up), instant Skia shape cropper → upload to `cutouts` bucket + `items` row, books/pages CRUD, Skia page editor (drag/pinch/rotate/z-index/opacity/backgrounds), client-side export (save to library + share).
- **Supabase schema**: 4 migrations — `content.{books,pages,items,page_items,usage_tracking}`, `public.profiles` (+ signup trigger), full RLS (owner-only + public-read), private `cutouts`/`exports` buckets with per-user RLS. Edge function `remove-background` (entitlement gate only, dormant).
- **Client wiring**: `src/lib/supabase.ts` (AsyncStorage session + auto-refresh), `src/lib/env.ts` (runtime validation), React Query + Zustand, signed-URL helper for private cutouts.

### ❌ Missing / blocking live deployment
1. **Hosted Supabase not wired to the app.** An existing project (`gujldqovvhjbctzelark`) is available but the app isn't pointed at it — `.env.example` still targets `127.0.0.1:54321` with the demo anon key, and we have not confirmed whether the 4 migrations/buckets/edge function are applied there.
2. **No real `.env`.** Only `.env.example` exists. Production needs real values (see env table below).
3. **Production auth not configured.** `config.toml` has `site_url=http://127.0.0.1:3000`, localhost redirect URLs, `enable_confirmations=false`, and no SMTP — password reset / confirmation emails can't be sent in prod.
4. **No password-reset screen or deep-link handler.** `auth-store.ts` exposes `resetPasswordForEmail`, but there's no `(auth)/reset` route and no `scrappd://` callback handling.
5. **No app assets.** No `apps/mobile/assets/` — no icon, adaptive icon, or splash image. App stores reject builds without these.
6. **No EAS config.** No `eas.json`, no `extra.eas.projectId` in `app.json`. Native binaries can't be built/submitted.
7. **No web deployment.** Web export is validated in CI but never published; no Vercel project/config.
8. **No crash reporting / analytics / structured logging** (no Sentry, no error boundary).
9. **Thin tests.** 3 unit test files (~129 LOC); cropper/editor/books/export untested.
10. **Minor**: lint not wired for `packages/supabase` & `packages/shared-types`; `exports` bucket unused (export is device-only).

### Environment variables — full inventory
| Variable | Where | Status | Notes |
|---|---|---|---|
| `EXPO_PUBLIC_SUPABASE_URL` | mobile build + Vercel build | needs prod value | hosted project URL |
| `EXPO_PUBLIC_SUPABASE_KEY` | mobile build + Vercel build | needs prod value | publishable/anon key |
| `EXPO_PUBLIC_SENTRY_DSN` | mobile build + Vercel build | **new** | added in Sentry prompt |
| `SUPABASE_URL` / `SUPABASE_ANON_KEY` | edge function | auto-injected | already used by `remove-background` |
| `ML_SERVICE_URL` | edge function | future (dormant) | only when premium ships |
| SMTP host/user/pass/sender | Supabase Auth (hosted) | **new** | e.g. Resend/SendGrid for prod emails |
| `EXPO_TOKEN` | GitHub Actions (CD) | **new** | EAS build/submit from CI |
| `VERCEL_TOKEN` / org / project IDs | GitHub Actions (CD) | **new** | web deploy from CI |
| `SENTRY_AUTH_TOKEN` / `SENTRY_ORG` / `SENTRY_PROJECT` | build (source maps) | **new** | sourcemap upload |
| Apple ASC API key + Team ID | EAS Submit (iOS) | **new** | store submission |
| Google Play service-account JSON | EAS Submit (Android) | **new** | store submission |

---

## The plan — independently deployable Claude prompts

Sequenced web-first. **P1 is the only hard prerequisite for everything else.** P3–P7 are independent of each other and can run in parallel/any order once P1 lands. P8/P9 (native builds) depend on P1 + P3 (assets). Each prompt is self-contained, ends in a commit/push to its own branch, and is independently verifiable. Shared preamble for every prompt: _"Repo `scrappd-app`. Monorepo: Expo app in `apps/mobile`, Supabase project in `packages/supabase`, dormant FastAPI in `apps/ml-service`, shared TS in `packages/shared-types`. Stack: Expo SDK 56 + Supabase (Auth/Postgres+RLS/Storage). ML is dormant. Work on a new branch, then commit and push."_

### P1 — Wire the existing hosted Supabase + production config *(foundation)*
```
Wire Scrappd to the existing hosted Supabase project gujldqovvhjbctzelark.
- First inspect the project with the Supabase MCP tools (list_migrations, list_tables on public + content): confirm whether the 4 migrations from packages/supabase/migrations are already applied. Apply only what's missing — do NOT double-apply. Confirm content.* tables, public.profiles + signup trigger, and RLS exist remotely. Create the 'cutouts' and 'exports' storage buckets with the per-user RLS from migration 4 if absent.
- Deploy the remove-background edge function (it stays dormant: 403 free / 501 premium).
- Configure production Auth: set Site URL and additional redirect URLs to the future web domain + the scrappd:// scheme; wire an SMTP provider so password-reset/confirmation emails send; decide enable_confirmations and document it. Mirror these into packages/supabase/config.toml.
- Add apps/mobile/.env from the hosted project URL + anon key. Document in apps/mobile/README.md how prod env vars are supplied. DO NOT commit the .env (keep .gitignored); commit only updated .env.example guidance.
Verify: from a throwaway script or the app against the hosted URL, sign up a user (profile row auto-created), upload a cutout, create a book/page, and confirm RLS blocks a second user from reading the first user's private book.
```

### P2 — Deploy Expo web to Vercel *(first live URL; depends on P1)*
```
Ship the Expo web build to Vercel as the first live Scrappd surface.
- Configure the Expo web export for static hosting and create the Vercel project (use the Vercel MCP tools). Set EXPO_PUBLIC_SUPABASE_URL and EXPO_PUBLIC_SUPABASE_KEY (and EXPO_PUBLIC_SENTRY_DSN if P5 has landed) as Vercel build-time env vars. Add a vercel.json / build config that runs `pnpm --filter mobile exec expo export --platform web` and serves the static output with SPA routing fallback.
- IMPORTANT web caveat: the cropper depends on expo-camera, which is limited on web. Make the cropper gracefully fall back to expo-image-picker (file upload) when running on web so the core crop→upload flow works in a browser. Gate any camera-only UI behind Platform checks.
- Set the hosted Supabase Site URL / redirect URLs to the deployed Vercel domain.
Verify: open the Vercel URL, sign up/in, run a crop (via image picker on web), create a book/page, place an item, export. Confirm session persists across reload.
```

### P3 — App icon, adaptive icon & splash assets *(store prerequisite)*
```
Add Scrappd's brand assets and finalize app.json for store builds.
- Create apps/mobile/assets/ with icon.png (1024x1024), adaptive-icon.png (foreground), splash.png, and favicon.png, using the brand palette already in app.json (#510420 / #F6E8C9). Wire them in app.json (icon, ios.icon, android.adaptiveIcon, expo-splash-screen plugin, web.favicon).
- Bump version, set iOS buildNumber / Android versionCode, and add extra.eas.projectId placeholder wiring (left blank until P8/P9 create the EAS project).
Verify: `expo export` succeeds; `expo-doctor` reports no missing-asset warnings; icons render in the web build.
```

### P4 — Password reset screen + deep linking *(independent feature)*
```
Complete the auth flow with password reset and deep linking.
- Add an (auth)/forgot-password screen (calls auth-store.resetPasswordForEmail) and an (auth)/reset screen that consumes the recovery token and calls supabase.auth.updateUser. Link them from the login screen.
- Handle the scrappd:// recovery deep link (and the web https reset URL) via expo-linking + expo-router so the email link lands on the reset screen on native and web. Register the redirect URLs in Supabase Auth config (coordinate with P1).
Verify: request a reset email, follow the link on web and on a device, set a new password, sign in with it.
```

### P5 — Sentry crash reporting + error boundary *(independent)*
```
Add production observability to the Expo app.
- Integrate sentry-expo / @sentry/react-native. Read EXPO_PUBLIC_SENTRY_DSN from env (extend src/lib/env.ts to make it optional). Wrap the root layout in a Sentry error boundary and a friendly fallback UI. Capture handled errors in the cropper upload, editor persistence, and export paths.
- Configure source-map upload (SENTRY_AUTH_TOKEN/ORG/PROJECT) for native and web builds; document the new env vars in README and .env.example.
Verify: trigger a test exception in dev/staging and confirm it appears in the Sentry dashboard with readable stack frames.
```

### P6 — Expand test coverage + lint the remaining packages *(independent)*
```
Raise the safety net before store launch.
- Add component/integration tests (jest + @testing-library/react-native) for: shape path building & cutout creation, books/pages CRUD hooks (mock supabase), page-editor transform persistence, and export-page guards. Mock @supabase/supabase-js and Skia where needed.
- Re-enable coverage thresholds in jest.config.js at a realistic floor and raise over time.
- Wire real lint for packages/supabase (SQL/format) and packages/shared-types so `pnpm lint` covers them.
Verify: `pnpm test` and `pnpm lint` pass locally and in CI; coverage meets the new threshold.
```

### P7 — CI/CD: auto-deploy web + build pipeline scaffolding *(independent; depends on P2)*
```
Extend .github/workflows so merges to main deploy automatically.
- Add a deploy-web job that builds the Expo web export and deploys to Vercel on push to main (using VERCEL_TOKEN + project/org IDs from GitHub Secrets). Keep it separate from the existing validation jobs.
- Add a manually-triggered (workflow_dispatch) eas-build job stub that runs `eas build`/`eas submit` once EAS is set up (P8/P9), reading EXPO_TOKEN from secrets. Document all required GitHub Secrets in the workflow and README.
Verify: a merge to main redeploys the Vercel site; the eas-build job is present and gated behind manual dispatch.
```

### P8 — EAS Build + TestFlight (iOS) *(depends on P1, P3)*
```
Set up iOS production builds via EAS and ship to TestFlight.
- Run eas init (creates the EAS project; fill extra.eas.projectId in app.json). Add eas.json with development, preview, and production profiles. Store EXPO_PUBLIC_SUPABASE_URL/ANON_KEY (+ Sentry DSN) as EAS environment variables/secrets for the build.
- Configure iOS credentials (bundleIdentifier com.scrappd.app) and eas submit with an App Store Connect API key. Produce a production build and submit to TestFlight.
Verify: a TestFlight build installs on a device, points at the hosted Supabase, and the full crop→book→export flow works end to end.
```

### P9 — EAS Build + Play internal track (Android) *(depends on P1, P3)*
```
Set up Android production builds via EAS and ship to the Play internal testing track.
- Add the Android production profile to eas.json (package com.scrappd.app, versionCode). Reuse the EAS env vars/secrets from P8. Configure a Google Play service-account JSON for eas submit.
- Produce a production AAB and submit to the Play internal testing track.
Verify: the internal-track build installs on an Android device, points at hosted Supabase, and the full flow works end to end.
```

### Later (out of scope for v1, kept dormant)
- **Premium ML**: deploy `apps/ml-service` to Cloud Run, set `ML_SERVICE_URL`, wire `remove-background` to forward + store results, add upgrade/paywall UI. Schema (`usage_tracking`, `subscription_tier`) already supports it.

---

## Verification (end to end)
1. **P1**: hosted Supabase reachable; two-user RLS check passes (reuse `packages/supabase/tests/rls_verification.sql`); buckets + edge function present.
2. **P2**: Vercel URL live; sign-up → crop (image picker) → book/page → export works in a browser; session persists.
3. **P3**: `expo export` + `expo-doctor` clean; icons/splash render.
4. **P4**: reset email round-trips on web and device.
5. **P5**: test exception visible in Sentry with source-mapped frames.
6. **P6**: `pnpm test`/`pnpm lint` green at the new coverage floor.
7. **P7**: merge to main auto-redeploys web.
8. **P8/P9**: TestFlight + Play internal builds run the full flow against hosted Supabase.

## Notes on sequencing & risk
- **Order**: P1 → P2 (live web fast) → P3/P4/P5/P6 in parallel → P7 → P8 → P9.
- **Web camera risk**: the cropper's `expo-camera` path is weak on web; P2's image-picker fallback is essential for the web release to be usable.
- **Secrets hygiene**: anon key is public by design (RLS enforces auth); never ship the Supabase service-role key in the app or web bundle. All store/CI credentials live in EAS/GitHub/Vercel secret stores, never committed.