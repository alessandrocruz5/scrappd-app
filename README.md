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

## Dormant ML service

The Python service is **not** part of normal development. To run it locally for
premium experiments:

```bash
docker compose --profile ml up    # builds + runs apps/ml-service on :8000
```

It will eventually be invoked from the `remove-background` Supabase Edge
Function for `pro` / `creator` users only.
