# Scrappd Revamp — Expo + Supabase, Books, Instant Shape Cropper, pnpm Monorepo

## Context

Scrappd today is a background-removal scrapbooking app:

- **Frontend:** Flutter (`scrappd-mobile/scrappd_mobile/`) — Provider, Dio, clean architecture.
- **Backend (Go):** Gin REST API (`backend/`) — JWT auth, `projects` → `pages` → `items`, R2 storage, Cloud Tasks async ML, Postgres (pgx).
- **ML (Python):** **already FastAPI** (`scrappd-ml-service/`) — BiRefNet background removal.
- **Not a monorepo.**

This revamp changes direction. Confirmed decisions:

1. **Mobile: Flutter → React Native (Expo).** Familiarity, easy iPhone testing, market share. Flutter app is **replaced and retired**.
2. **Drop AI from the default flow** → future **premium** feature. Python FastAPI service stays (no migration needed — FastAPI is already correct) but **dormant**, invoked later via a Supabase Edge Function.
3. **Instant shape cropper** replaces upload→remove-background: user picks a **shape** (square, circle, rectangle, heart, star, stamp, scalloped circle) and aims it at a pattern/poster. **Cropping is client-side (Skia).**
4. **Rebrand `Project` → `Book`.** Hierarchy: Book → Pages → Items.
5. **Page export client-side (Skia)** — snapshot the page canvas in-app; no server render endpoint.
6. **Backend → Supabase (retire Go).** CTO call: with AI paused and crop+export client-side, the remaining work (auth, CRUD, storage) is pure Supabase territory; a Go service forwarding CRUD to Postgres is negative ROI. Go code is preserved in git history.
7. **pnpm + Turborepo monorepo** orchestrating the Expo app, the Supabase project, and the dormant Python service.

---

## Target architecture

```
Expo app (supabase-js)
   ├── Supabase Auth        (replaces Go auth service entirely)
   ├── Supabase Postgres    (existing schema, migrated; RLS replaces handler-level checks)
   ├── Supabase Storage     (replaces R2 — stores shape cutouts; exports optional)
   └── Supabase Edge Fn     (future premium: calls Python ML on Cloud Run)
Python FastAPI (Cloud Run)  (dormant; background removal for premium tier later)
```

No custom API server. Expo talks to Supabase directly; RLS enforces per-user ownership via `auth.uid()`.

---

## Execution as Claude prompts

Run these as **7 milestone prompts**, each in its own session, each ending in a commit/push to `claude/great-cori-H5Xqp`. Granularity is deliberately milestone-sized (PR-sized), not micro: fewer context reloads, every step independently verifiable. Run them in order — each assumes the previous is committed. Copy a prompt verbatim into a fresh session.

> **Shared preamble** (true for every prompt): *Repo `scrappd-app`, branch `claude/great-cori-H5Xqp`. Monorepo with Expo app in `apps/mobile`, Python ML in `apps/ml-service`, Supabase project in `packages/supabase`, shared TS in `packages/shared-types`. Stack: Expo + Supabase (Auth/Postgres+RLS/Storage) + dormant Python FastAPI ML. AI is paused; cropping and page export are client-side (Skia). Work on the branch, then commit and push.*

### Prompt 1 — Monorepo scaffold (no behavior change)
```
Convert scrappd-app into a pnpm + Turborepo monorepo on branch claude/great-cori-H5Xqp.
- Add root package.json (packageManager: pnpm), pnpm-workspace.yaml (apps/*, packages/*), turbo.json with build/dev/lint/typecheck/db tasks.
- Move scrappd-ml-service/ -> apps/ml-service/ and give it a thin package.json whose lint/test/dev scripts shell to its existing Makefile (keep requirements.txt as the real build). Verify the FastAPI service still boots.
- Create packages/supabase/ as a Supabase CLI project skeleton (config.toml, migrations/, functions/) and packages/shared-types/ (empty TS package).
- Leave the Go backend/ and Flutter scrappd-mobile/ in place for now (later prompts retire them). Remove only Go-specific docker-compose services if they block `supabase start`; otherwise leave compose alone.
- Add root scripts: db:start, db:push, gen:types, dev, build, lint, typecheck.
Verify: pnpm install resolves; turbo runs; ml-service boots. Commit and push.
```

### Prompt 2 — Supabase schema, RLS, and Project→Book rename
```
In packages/supabase, build the database from the existing Go Postgres schema (backend/migrations/*.sql) with these changes, on branch claude/great-cori-H5Xqp:
- Rename content.projects -> content.books, and content.pages.project_id -> book_id (plus indexes/FKs).
- Create public.profiles (id FK auth.users, display_name, bio, avatar_url, subscription_tier default 'free') with a trigger inserting a row on signup. Drop the bespoke auth.users columns from the Go schema (Supabase Auth owns users now).
- Keep items/page_items/usage_tracking columns (ML columns retained but unused). Items default processing_status='completed'.
- Add RLS policies on books, pages, items, page_items: owner-only via auth.uid(), plus public read where books.visibility='public'. Add a Storage bucket 'cutouts' (and 'exports') with per-user RLS.
- Run migrations on a local Supabase stack (supabase start) and generate TS types into packages/shared-types (supabase gen types typescript).
Verify: migrations apply cleanly; with two test users, RLS blocks cross-user access and allows public books. Commit and push.
```

### Prompt 3 — Expo app bootstrap: client, auth, navigation shell
```
Create the Expo (TypeScript) app in apps/mobile, replacing the Flutter app, on branch claude/great-cori-H5Xqp.
- Scaffold Expo + expo-router. Add @supabase/supabase-js, @tanstack/react-query, zustand, @react-native-async-storage/async-storage.
- src/lib/supabase.ts: Supabase client with session persisted to AsyncStorage + auto-refresh. Use generated types from packages/shared-types.
- Auth: login + register + password reset screens using supabase.auth; an auth store; an auth gate that routes to the tab shell when signed in (mirror the Flutter root_screen.dart / main_shell.dart behavior).
- expo-router tab shell with three tabs: Cropper, Books, Profile (placeholders for now).
Verify: `pnpm --filter mobile dev`, sign up/in against local Supabase on Expo Go (iPhone), session persists across reload. Commit and push.
```

### Prompt 4 — Instant shape cropper (the centerpiece)
```
Build the instant shape cropper in apps/mobile on branch claude/great-cori-H5Xqp.
- Add @shopify/react-native-skia, expo-camera, expo-image-picker.
- src/cropper/shapes.ts: parametric Skia Path builders for square, rectangle, circle, heart, 5-point star, scalloped circle (flower edge), and stamp (perforated postage edge). Each takes a bounding box and returns a Path.
- Cropper screen: live camera/preview with the selected shape overlaid as an aiming mask; a shape picker. On capture, clip the framed region through the shape Path and makeImageSnapshot to a transparent-background PNG cutout.
- Upload the cutout to the Supabase 'cutouts' Storage bucket and insert an items row (processing_status='completed'). No polling (AI path is off).
Verify: on iPhone via Expo Go, each shape produces a clean transparent PNG cutout that uploads and appears as an item. Commit and push.
```

### Prompt 5 — Books and pages (list + CRUD)
```
Implement Books and Pages in apps/mobile on branch claude/great-cori-H5Xqp.
- Books tab: list user's books, create/rename/delete a book (supabase.from('books') under RLS); open a book to a grid of its pages. Port behavior from the Flutter ProjectsProvider.
- Within a book: create/reorder/delete pages (content.pages keyed by book_id); tapping a page opens the editor (next prompt — stub the route).
- Wire items created by the cropper so they can be picked when adding to a page.
Verify: create a book, add pages, see them persist and reload under RLS. Commit and push.
```

### Prompt 6 — Page editor canvas
```
Build the Skia page editor in apps/mobile on branch claude/great-cori-H5Xqp, porting page_editor_screen.dart.
- Canvas rendering page_items with drag (move), pinch (resize), rotate; z-index ordering; opacity; page background color/image/pattern; layout templates.
- Add items to a page from the user's cutouts; persist transforms to content.page_items (position_x/y, width, height, rotation, z_index, opacity, filters) via supabase updates. Use Reanimated + Gesture Handler + Skia.
Verify: place several cutouts, move/resize/rotate, reload, transforms persisted. Commit and push.
```

### Prompt 7 — Client-side export, retire Flutter, ML premium placeholder
```
Finish the revamp on branch claude/great-cori-H5Xqp.
- Client-side export: snapshot the page Skia canvas to a high-res PNG; save via expo-media-library and share via expo-sharing (port page_export_service.dart intent). No server render endpoint.
- Delete the Go backend (backend/) and the Flutter app (scrappd-mobile/) now that the Expo app has parity; update docker-compose, READMEs, and CI to the new layout. (Code remains in git history.)
- Add packages/supabase/functions/remove-background as a documented placeholder Edge Function that gates on profiles.subscription_tier and would call apps/ml-service later (no real ML wiring now).
Verify: export a page to the photo library and share sheet on iPhone; repo builds with Go/Flutter removed; edge function placeholder rejects non-pro users. Commit and push.
```

**Ordering notes:** Prompt 2 must land before 3 (the app needs the schema + generated types); 4–6 build on 3. Don't parallelize *across* prompts; do parallelize *within* a session. Prompts 3, 4, 6, 7 need on-device iPhone testing (Expo Go + dev server reachable) — run those locally, not in the remote container. Prompts 3–4 need a Supabase URL + anon key (local `supabase start` or a hosted project).

---

## Phase detail (reference for the prompts above)

### Phase 0 — Monorepo (pnpm + Turborepo)

```
scrappd-app/
├── package.json            # root: packageManager pnpm, turbo
├── pnpm-workspace.yaml      # apps/*, packages/*
├── turbo.json               # build / dev / lint / typecheck / db tasks
├── apps/
│   ├── mobile/              # NEW Expo app (replaces Flutter)
│   └── ml-service/          # moved from scrappd-ml-service (Python, dormant)
├── packages/
│   ├── shared-types/        # `supabase gen types typescript` output + shared TS
│   └── supabase/            # migrations, RLS policies, seed, edge functions, config.toml
```

- Python `ml-service` gets a thin `package.json` whose scripts shell to its Makefile so Turbo can run lint/test; real build stays `requirements.txt`.
- `packages/supabase` holds the Supabase CLI project (`supabase/migrations`, `supabase/functions`, `config.toml`). Root scripts: `pnpm db:start` (local stack), `pnpm db:push`, `pnpm gen:types`.
- Update `docker-compose.yml` to local Supabase stack (or drop it in favor of `supabase start`). Remove Go-specific compose services.

### Phase 1 — Supabase schema (Books rename + RLS)

Port the existing Postgres schema into `packages/supabase/migrations` (data is already Postgres — schema import, not rewrite). Apply the **Project → Book** rename here:

- Tables in `content` schema: `books` (was `projects`), `pages` (`project_id` → `book_id`), `items`, `page_items`. Drop/park `usage_tracking` (only needed when premium AI returns).
- **User profiles:** create `public.profiles` linked to `auth.users.id` (display_name, bio, avatar, subscription_tier). Replaces `auth.users` custom columns from the Go schema. Trigger to insert a profile row on signup.
- **RLS policies** on `books`, `pages`, `items`, `page_items`: owner-only read/write via `auth.uid()` (and a public-read policy for `books.visibility = 'public'`). This replaces every handler-level ownership check in Go.
- **Storage bucket** `cutouts` (and optional `exports`): RLS so users only access their own objects.
- Item model simplifies: a cropped cutout is uploaded straight to Storage and an `items` row is created `processing_status = completed`. No ML columns exercised in the default flow (kept in schema for premium later).
- `supabase gen types typescript` → `packages/shared-types` for end-to-end typing in the app.

### Phase 2 — Expo app (`apps/mobile`), replacing Flutter

Stack: **Expo (TypeScript)**, **expo-router**, **@supabase/supabase-js** (auth + data + storage), **@tanstack/react-query** (server cache), **Zustand** (UI state), **@shopify/react-native-skia** (cropper + page canvas + export), **expo-image-picker / expo-camera**, **expo-media-library / expo-sharing**.

- **Supabase client** (`src/lib/supabase.ts`): session persisted with AsyncStorage; auto-refresh. Replaces the Dio `AuthInterceptor`/`TokenStorage` logic.
- **Auth** screens + store: `supabase.auth` (signUp/signInWithPassword/reset). Replaces `AuthProvider`.
- **Books** (replaces Projects): list / create / open; book = grid of pages. Direct `supabase.from('books')` queries under RLS. Port `ProjectsProvider` behavior.
- **Instant Shape Cropper** (centerpiece):
  - Live camera/preview; overlay the chosen **shape mask** the user aims at a pattern/poster.
  - Capture → clip the region through the shape with **Skia** (`Path` mask) → `makeImageSnapshot` → transparent-background PNG cutout.
  - Shapes as parametric Skia `Path` builders in `src/cropper/shapes.ts`: square, rectangle, circle, heart, 5-point star, **scalloped circle** (flower edge), **stamp** (perforated postage edge).
  - Upload cutout to Supabase Storage (`cutouts` bucket) + insert `items` row. No polling in the non-AI path.
- **Page editor** (canvas): port `page_editor_screen.dart` — drag/pinch/rotate `page_items`, z-index, opacity, backgrounds, templates — with Skia + Reanimated + Gesture Handler. Persist via `supabase.from('page_items').update(...)`.
- **Export/Post (client-side):** snapshot the page Skia canvas to a high-res image, save with `expo-media-library`, share with `expo-sharing`. No server render endpoint.
- **Navigation:** expo-router tab shell (Cropper / Books / Profile); mirror `main_shell.dart`.

**Retire Flutter** once parity is verified: delete `scrappd-mobile/` in a dedicated commit.

### Phase 3 — Park ML for premium (no work now, just keep it runnable)

- Keep `apps/ml-service` deployable to Cloud Run; do not wire it into the default flow.
- Stub a `packages/supabase/functions/remove-background` Edge Function (calls the Python service with a service secret), guarded by `profiles.subscription_tier`. Implemented when premium ships — left as a documented placeholder now.

---

## Critical files

- Monorepo: `package.json`, `pnpm-workspace.yaml`, `turbo.json` (new root); `packages/supabase/config.toml`.
- Schema/RLS: `packages/supabase/migrations/*` (books rename, profiles, RLS, storage buckets), generated `packages/shared-types`.
- Expo app: `apps/mobile/src/lib/supabase.ts`, `src/cropper/shapes.ts`, `src/screens/`, `src/stores/`.
- Reuse references (behavior to port): `page_export_service.dart`, `items_remote_datasource.dart` (upload shape), `api_constants.dart`, `auth_interceptor`/`TokenStorage`, `page_editor_screen.dart`.
- Removed: `backend/` (Go) — preserved in git history.

---

## Verification

- **Monorepo:** `pnpm install`; `pnpm db:start` boots local Supabase; `pnpm gen:types` produces TS types; `pnpm typecheck` passes.
- **Schema/RLS:** run migrations on local Supabase; verify a user can only read/write their own books/pages/items (test RLS with two users); public books readable by others.
- **Cropper (Expo):** `pnpm --filter mobile dev` → Expo Go on iPhone. For each shape: aim, capture, confirm a clean transparent-PNG cutout; verify upload to `cutouts` and an `items` row appears.
- **Books/pages/export end-to-end:** create a book → add pages → place cutouts → client-side export → save/share the post artifact.
- **Premium guard (smoke):** confirm the `remove-background` Edge Function placeholder rejects non-`pro` users (no real ML call yet).

## Sequencing

Phase 0 (monorepo) → Phase 1 (Supabase schema + RLS, Book rename) → Phase 2 (Expo: cropper → books → editor → export) → retire Flutter → Phase 3 placeholder. Each phase commits independently to `claude/great-cori-H5Xqp`.
