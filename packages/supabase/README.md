# @scrappd/supabase

Supabase CLI project for Scrappd: local Postgres/Auth/Storage stack, database
migrations, and edge functions.

## Layout

- `config.toml` — Supabase CLI configuration for the local stack.
- `migrations/` — SQL migrations applied with `supabase db push` / `supabase db reset`.
- `functions/` — Supabase Edge Functions (Deno).
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

The Supabase CLI requires Docker to be running for `supabase start`. If the
legacy Go `docker-compose.yml` services bind conflicting ports, stop them
before starting the Supabase stack.
