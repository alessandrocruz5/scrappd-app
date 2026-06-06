# @scrappd/supabase

Supabase CLI project for Scrappd: local Postgres/Auth/Storage stack, database
migrations, and edge functions.

## Layout

- `config.toml` — Supabase CLI configuration for the local stack.
- `migrations/` — SQL migrations applied with `supabase db push` / `supabase db reset`.
- `functions/` — Supabase Edge Functions (Deno). Currently just
  `remove-background`, a **dormant premium placeholder** that gates on
  `profiles.subscription_tier` and will later call the Python ML service
  (`apps/ml-service`). See `functions/remove-background/README.md`.
- `tests/` — SQL verification scripts (e.g. RLS checks with two users).

## Schema

Ported from the legacy Go Postgres schema (`backend/migrations/*.sql`) with the
revamp changes (see `REVAMP_PLAN.md`, Phase 1):

- **`content` schema** — `books` (was `projects`), `pages` (`project_id` →
  `book_id`), `items`, `page_items`, `usage_tracking`. The `content` schema is
  exposed to the PostgREST API (see `[api] schemas` in `config.toml`). Items
  default to `processing_status = 'completed'`; the ML/processing columns are
  retained but unused in the default cropper flow.
- **`public.profiles`** — 1:1 with `auth.users` (display_name, bio, avatar_url,
  subscription_tier). Supabase Auth owns `auth.users`; the bespoke user columns
  from the Go schema live here now. A trigger inserts a profile row on signup.
- **RLS** — owner-only read/write on `books`/`pages`/`items`/`page_items` via
  `auth.uid()`, plus public read where `books.visibility = 'public'`.
- **Storage** — private buckets `cutouts` and `exports` with per-user RLS
  (objects are namespaced under `<bucket>/<auth.uid()>/...`).

## Environments

- **Local (dev):** the Docker stack started with `pnpm db:start`. `config.toml`
  configures it.
- **Live (hosted):** a single hosted Supabase project, `gujldqovvhjbctzelark`
  (`https://gujldqovvhjbctzelark.supabase.co`), is the live backend. A dedicated
  dev/prod split is deferred (free-tier active-project cap) — see the migration
  plan. The mobile app points at it via `apps/mobile/.env` (and Vercel/EAS build
  env for the published builds).

Current state of the hosted project (verified):

- All 4 migrations applied (`content` schema, `profiles` + signup trigger, RLS,
  storage buckets). **Do not re-apply** — diff against `migrations/` before any
  `db push`.
- Private `cutouts` and `exports` buckets exist with per-user RLS.
- The `remove-background` edge function is deployed (dormant: 403 free /
  501 premium).
- Production Auth (Site URL, redirect URLs, SMTP, `enable_confirmations`) still
  needs to be set in the dashboard / pushed — see the production notes in
  `config.toml`. These cannot be committed (SMTP password is a secret).

## Verifying RLS

With the local stack running, run the two-user RLS checks:

```bash
psql "$(supabase status -o env | grep '^DB_URL=' | cut -d= -f2- | tr -d '\"')" \
  -f tests/rls_verification.sql
```

It seeds two users + a public/private book and asserts that cross-user private
data is invisible/unwritable, public books are readable by others and by anon,
and storage objects are owner-scoped.

## Common commands

Run from the repo root:

```bash
pnpm db:start     # start the local Supabase stack (Docker)
pnpm db:push      # apply migrations to the linked project
pnpm gen:types    # generate TS types into packages/shared-types
```

Or from this package with the bundled CLI:

```bash
pnpm --filter @scrappd/supabase exec supabase <command>
```

## Notes

The Supabase CLI requires Docker to be running for `supabase start`. It manages
its own containers — the repo's `docker-compose.yml` now only holds the dormant
ML service (behind the `ml` profile), so there are no port conflicts with the
Supabase stack.
