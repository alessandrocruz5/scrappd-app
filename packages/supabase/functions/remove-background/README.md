# `remove-background` Edge Function (placeholder)

A **dormant, premium-gated placeholder** for AI background removal. It exists so
the future premium tier has a wired entry point, but it does **no real ML work
yet**.

## What it does today

1. Requires an `Authorization: Bearer <user-jwt>` header.
2. Resolves the caller with `supabase.auth.getUser()`.
3. Reads their `public.profiles.subscription_tier`.
4. Gates:
   - **free** → `403` `{ upgrade_required: true }`
   - **pro / creator** → `501` (entitled, but the pipeline is dormant)

## What it will do later

For entitled users it will forward the image to the dormant Python FastAPI
BiRefNet service (`apps/ml-service`, deployed to Cloud Run), store the processed
PNG in the private `cutouts` Storage bucket, and update the `content.items` row.
See the inline comments in `index.ts` for the intended flow and the
`ML_SERVICE_URL` env var it will use.

## Run / deploy

```bash
# Serve locally (needs the local stack: pnpm db:start)
pnpm --filter @scrappd/supabase exec supabase functions serve remove-background

# Deploy to the linked project
pnpm --filter @scrappd/supabase exec supabase functions deploy remove-background
```

`SUPABASE_URL` and `SUPABASE_ANON_KEY` are injected automatically by the
Supabase runtime. `verify_jwt` is enabled in `config.toml`, so unauthenticated
requests are rejected before the function body even runs.

## Try it

```bash
# Non-pro users are rejected with 403:
curl -i -X POST "$SUPABASE_URL/functions/v1/remove-background" \
  -H "Authorization: Bearer <free-user-jwt>"
# → HTTP/1.1 403  { "error": "Background removal is a premium feature.", ... }
```
