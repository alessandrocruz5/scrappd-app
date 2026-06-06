# Scrappd

Scrappd is a mobile scrapbooking app. You aim a **shape** (square, circle,
heart, star, …) at a pattern or poster, snap an **instant cutout** (cropped
client-side with Skia — no upload, no server), arrange cutouts on **pages**
inside a **Book**, and **export** a page as a high-res PNG straight from the
phone.

This repository is a **pnpm + Turborepo monorepo**.

## Layout

```
apps/
  mobile/        Expo (React Native) app — the product. Talks to Supabase
                 directly (Auth, Postgres + RLS, Storage). Cropping and page
                 export are client-side (Skia / view-shot).
  ml-service/    Python FastAPI BiRefNet background-removal service. DORMANT —
                 paused for the revamp, slated to return as a premium feature
                 behind the remove-background Edge Function. Not in the default
                 flow.
packages/
  supabase/      Supabase CLI project: config, SQL migrations (the `content`
                 schema + `profiles` + RLS + Storage buckets), and Edge
                 Functions (incl. the dormant `remove-background` placeholder).
  shared-types/  Generated Supabase TypeScript types, shared across the workspace.
```

## Architecture

```
Expo app (supabase-js)
   ├── Supabase Auth        sign-in / sign-up
   ├── Supabase Postgres    Books → Pages → Items (RLS via auth.uid())
   ├── Supabase Storage     private 'cutouts' (+ optional 'exports') buckets
   └── Supabase Edge Fn     remove-background (premium placeholder, dormant)
Python FastAPI (Cloud Run)  background removal for the future premium tier (dormant)
```

There is **no custom API server**. The Expo app talks to Supabase directly and
RLS enforces per-user ownership. The cropper and page export run entirely on
the device.

> **History:** Scrappd began as a Flutter app with a Go (Gin) REST backend and
> an always-on ML background-removal step. The revamp replaced Flutter with
> Expo, retired the Go backend in favour of Supabase, and made cropping + export
> client-side. The Go service and Flutter app are preserved in git history; see
> `REVAMP_PLAN.md`.

## Prerequisites

- Node ≥ 20 and **pnpm 10** (`corepack enable`)
- Docker (for the local Supabase stack)
- The Supabase CLI is bundled as a dev dependency of `@scrappd/supabase`.

## Getting started

```bash
pnpm install                 # install all workspaces
pnpm db:start                # start the local Supabase stack (Docker)
cp apps/mobile/.env.example apps/mobile/.env   # then fill in URL + anon key
pnpm --filter mobile dev     # start the Expo dev server (open in Expo Go)
```

See [`apps/mobile/README.md`](apps/mobile/README.md) for the app,
[`packages/supabase/README.md`](packages/supabase/README.md) for the database
and Edge Functions, and [`apps/ml-service/README.md`](apps/ml-service/README.md)
for the dormant ML service.

## Common scripts (repo root)

| Command          | Description                                          |
| ---------------- | ---------------------------------------------------- |
| `pnpm dev`       | Run `dev` across workspaces (Turborepo)              |
| `pnpm lint`      | Lint all workspaces                                  |
| `pnpm typecheck` | Type-check all workspaces                            |
| `pnpm build`     | Build all workspaces                                 |
| `pnpm db:start`  | Start the local Supabase stack                       |
| `pnpm db:push`   | Apply migrations to the linked Supabase project      |
| `pnpm gen:types` | Regenerate `packages/shared-types` from the database |

## CI/CD

Two GitHub Actions workflows run against this repository.

**`ci.yml`** — runs on every push and PR: lint, typecheck, unit tests (+ coverage artifact), Prettier check, Expo web export, ML byte-compile. All jobs must pass before a PR can be merged.

**`deploy-web.yml`** — runs on every push to `main`: exports the Expo web bundle and deploys it to Vercel as a production deployment. Requires these GitHub Secrets (set under **Settings → Secrets and variables → Actions**):

| Secret                          | Description                                               |
| ------------------------------- | --------------------------------------------------------- |
| `EXPO_PUBLIC_SUPABASE_URL`      | Hosted Supabase project URL (`https://<ref>.supabase.co`) |
| `EXPO_PUBLIC_SUPABASE_ANON_KEY` | Supabase publishable anon key                             |
| `EXPO_PUBLIC_SENTRY_DSN`        | Sentry DSN for error reporting                            |
| `VERCEL_TOKEN`                  | Vercel personal-access or team token                      |
| `VERCEL_ORG_ID`                 | Vercel team/org ID (from Vercel project settings)         |
| `VERCEL_PROJECT_ID`             | Vercel project ID (from Vercel project settings)          |

**`eas-build.yml`** — manually triggered (`workflow_dispatch`): builds native iOS/Android binaries via EAS and optionally submits them to the app stores. Gate this behind the `production` GitHub environment for approval. Additional secrets needed when this is activated (P8/P9):

| Secret                            | Description                                       |
| --------------------------------- | ------------------------------------------------- |
| `EXPO_TOKEN`                      | Expo access token for EAS                         |
| `SENTRY_AUTH_TOKEN`               | Sentry auth token for source-map upload           |
| `SENTRY_ORG`                      | Sentry organisation slug                          |
| `SENTRY_PROJECT`                  | Sentry project slug                               |
| `ASC_APP_ID`                      | Apple App Store Connect app ID (iOS submit)       |
| `ASC_API_KEY_ID`                  | App Store Connect API key ID                      |
| `ASC_API_KEY_ISSUER_ID`           | App Store Connect API key issuer ID               |
| `ASC_API_KEY`                     | Base-64 encoded `.p8` private key                 |
| `GOOGLE_SERVICE_ACCOUNT_KEY_JSON` | Google Play service-account JSON (Android submit) |

## Dormant ML service

The Python service is **not** part of normal development. To run it locally for
premium experiments:

```bash
docker compose --profile ml up    # builds + runs apps/ml-service on :8000
```

It will eventually be invoked from the `remove-background` Supabase Edge
Function for `pro` / `creator` users only.
