# @scrappd/supabase

Supabase CLI project for Scrappd: local Postgres/Auth/Storage stack, database
migrations, and edge functions.

## Layout

- `config.toml` — Supabase CLI configuration for the local stack.
- `migrations/` — SQL migrations applied with `supabase db push` / `supabase db reset`.
- `functions/` — Supabase Edge Functions (Deno).

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
