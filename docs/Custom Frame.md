# Custom Frames — Independently Deployable Prompts

## Context

Today a "frame" in the cropper is one of 7 hardcoded vector shapes (`square`,
`circle`, `heart`, etc.) — each a parametric Skia `SkPath` (`shapes.ts`) used as
a clip mask to turn a photo into a transparent-background PNG cutout
(`create-cutout.ts`). There's no way for users to make their own.

Goal: let users create their **own** frames, saved per-user and reused across
sessions, via two methods — (1) **upload a PNG** whose **opaque area becomes the
cutout shape** (alpha mask, not a decorative overlay), and (2) **freehand
drawing** a closed shape. Entry is via "Draw" / "Upload" chips appended to the
existing picker, opening pageSheet modals (pattern: `books/item-picker.tsx`).

The unifying abstraction is a `FrameSpec` discriminated union that replaces the
bare `ShapeId` in screen state, overlay, and `createCutout`. Drawn frames flow
through the **existing `clipPath` pipeline** (stored as an SVG path string in
0..1 unit space); image frames use a **Skia `DstIn` alpha composite**.

```ts
export type FrameSpec =
  | { kind: 'builtin'; id: ShapeId }
  | { kind: 'svg'; svgPath: string; frameId?: string }    // drawn
  | { kind: 'image'; maskUri: string; frameId?: string };  // uploaded
```

## How the work is sliced

Four prompts, each independently mergeable and deployable (each leaves the app
working and shippable). **Prompt 1** and **Prompt 2** are foundations with no
user-visible change. **Prompt 3** (drawing) and **Prompt 4** (upload) each
deliver a complete end-to-end user feature and are independent of each other —
either can ship first once 1 & 2 are in.

```
Prompt 1 (FrameSpec refactor) ─┐
                               ├─→ Prompt 3 (Draw)    ← ships independently
Prompt 2 (persistence layer) ──┤
                               └─→ Prompt 4 (Upload)  ← ships independently
```

---

## PROMPT 1 — Introduce the `FrameSpec` abstraction (pure refactor)

**Deployable because:** no behavior change — only `builtin` frames exist and the
UI is identical. It just reshapes internals so later prompts plug in cleanly.

**Scope / files:**
- `apps/mobile/src/cropper/shapes.ts`: add the `FrameSpec` type and
  `buildFramePath(spec, box): SkPath | null` (handles `builtin` via existing
  `buildShapePath`; `svg` via `Skia.Path.MakeFromSVGString` + a translate/scale
  matrix mapping 0..1 → box; returns `null` for `image`).
- `apps/mobile/src/cropper/create-cutout.ts`: change signature to
  `createCutout(imageUri, frame: FrameSpec)`; replace `buildShapePath(shape,…)`
  with `buildFramePath(frame, insetBox(CUTOUT_SIZE))`; throw a friendly error if
  it returns null. (Only `builtin`/`svg` vector path is exercised here for now.)
- `apps/mobile/src/cropper/shape-overlay.tsx`: prop `shape` → `frame: FrameSpec`;
  `path = buildFramePath(frame, insetBox(size))`; unchanged dim/clear/stroke
  render for non-null paths.
- `apps/mobile/src/cropper/shape-picker.tsx`: keep rendering `SHAPES`, but
  `onSelect` now emits `{ kind:'builtin', id }`; `selected: FrameSpec`, active
  match on `selected.kind==='builtin' && selected.id===id`.
- `apps/mobile/src/cropper/cropper-screen.tsx`: state
  `useState<FrameSpec>({ kind:'builtin', id:'square' })`; thread `frame` to
  overlay/picker and `createCutout`/`uploadCutout`; update `captureHandledError`
  context to `{ kind: frame.kind }`.
- `apps/mobile/src/cropper/upload-cutout.ts`: signature
  `uploadCutout(cutout, frame: FrameSpec)`; set `item_name` from `frame.id` for
  builtins (preserve current values).

**Verify:** `pnpm typecheck`, `pnpm lint`; extend `create-cutout.test.ts` mock
for `buildFramePath`; manual run — all 7 shapes behave exactly as before.

---

## PROMPT 2 — Persistence layer for custom frames (backend, no UI)

**Deployable because:** adds tables/bucket/API that are simply unused until a UI
prompt lands. No runtime behavior change.

**Scope / files:**
- New migration `packages/supabase/migrations/20260606000001_custom_frames.sql`:
  `content.custom_frames` (`id`, `user_id` FK cascade, `kind` check
  `'drawn'|'image'`, `name`, `svg_path` text nullable, `mask_image_key`
  varchar nullable, timestamps, `deleted_at`, a payload CHECK enforcing the
  right column per kind), indexes (user_id, created_at desc, partial
  deleted_at), `set_updated_at` trigger, and owner-only RLS (mirror
  `content.items` in `20260605000003_content_rls.sql`). Also add
  `items.custom_frame_id uuid references content.custom_frames(id) on delete
  set null`.
- New migration `…20260606000002_frames_bucket.sql`: copy the `cutouts` block
  from `20260605000004_storage_buckets.sql` with `bucket_id='frames'` (private,
  per-user-folder RLS).
- Run `pnpm gen:types` to update `packages/shared-types/src/database.types.ts`
  (do not hand-edit).
- New `apps/mobile/src/cropper/frames-api.ts` (follow `books/api.ts`):
  `listCustomFrames`, `createDrawnFrame(name, svgPath)`,
  `createImageFrame(name, maskImageKey)`, `deleteCustomFrame(id)` (soft delete).
- New `apps/mobile/src/cropper/frames-hooks.ts` (follow `books/hooks.ts`):
  `useCustomFrames` + create/delete mutations invalidating `['custom-frames']`.
- Extend `apps/mobile/src/lib/storage.ts`: `signFrameUrl(key)` (mirror
  `signCutoutUrl`, `from('frames')`).

**Verify:** `pnpm db:start` → apply migrations → `pnpm gen:types`; SQL linter
(`packages/supabase/scripts/lint-sql.mjs`); `pnpm typecheck`; new
`frames-api.test.ts` / `frames-hooks.test.tsx` (mirror books tests); Supabase MCP
`execute_sql` sanity insert proving the payload CHECK.

---

## PROMPT 3 — Freehand drawing frames (end-to-end feature)

**Depends on:** Prompts 1 & 2. **Deployable because:** delivers a complete user
flow — draw → save → reuse — using the existing clip pipeline; independent of the
upload feature.

**Scope / files:**
- New `apps/mobile/src/cropper/draw-frame.tsx`: `<Modal pageSheet>` with a square
  Skia `<Canvas>`. Capture with `Gesture.Pan()` + `GestureDetector`
  (gesture-handler/reanimated/worklets already installed): `onBegin` moveTo,
  `onUpdate` lineTo/`quadTo` smoothing; live stroke + faint filled preview.
  **Done**: reject < ~3 segments / zero-area bounds (inline error); `close()`;
  normalize to 0..1 via matrix from `getBounds()` (inverse of `buildFramePath`);
  `toSVGString()`; name input; `useCreateDrawnFrame`; `onSaved` + close. Wrap in
  `GestureHandlerRootView` only if root doesn't already (check `app/_layout.tsx`).
- `shape-picker.tsx`: render `useCustomFrames()` drawn-frame chips (small Skia
  thumbnail via `buildFramePath`; long-press → `useDeleteCustomFrame`) and a
  trailing "+ Draw" (`pencil-outline`) action chip → `onRequestDraw`.
- `cropper-screen.tsx`: `drawOpen` state; render `<DrawFrameModal>`; `onSaved`
  sets `frame` to `{ kind:'svg', svgPath, frameId }`; React Query refresh shows
  the new chip.
- `upload-cutout.ts`: set `item_name` to the frame name and `custom_frame_id`
  for `svg` kind.
- (`create-cutout.ts` already handles `svg` from Prompt 1.)

**Verify:** unit-test the normalize→SVG conversion (bounds → 0..1, empty
rejected); `pnpm typecheck`/`lint`/`test`; manual — draw a heart, confirm chip
persists across app restart, capture yields a correctly clipped cutout. If
gesture capture is flaky under CanvasKit, gate the Draw chip behind
`Platform.OS !== 'web'`.

---

## PROMPT 4 — Uploaded-image frames (end-to-end feature)

**Depends on:** Prompts 1 & 2. **Deployable because:** delivers a complete user
flow — upload PNG → save → reuse — via Skia alpha masking; independent of the
drawing feature.

**Scope / files:**
- `create-cutout.ts`: add the `image` branch. **Order is critical:**
  1. `canvas.saveLayer()` (isolated layer — so `DstIn` composites against the
     photo, NOT the transparent backdrop; getting this wrong erases everything).
  2. Draw photo center-*cover* (existing `drawImageRect`).
  3. Decode mask: `Skia.Image.MakeImageFromEncoded(await Skia.Data.fromURI(frame.maskUri))` (guard null).
  4. `maskPaint.setBlendMode(BlendMode.DstIn)` + antialias.
  5. Draw mask center-*fit* (contain, preserve aspect) so the silhouette isn't cropped.
  6. `canvas.restore()`; then existing snapshot/encode. (Only mask alpha matters.)
- New `apps/mobile/src/cropper/upload-frame.tsx`: `<Modal pageSheet>`;
  `ImagePicker.launchImageLibraryAsync({ mediaTypes:['images'], quality:1 })`;
  preview over checkerboard; validate (reject non-PNG / fully-opaque, hint that a
  transparent PNG is required); name input; upload bytes to `frames` bucket at
  `<user.id>/<rand>.png` (reuse `bytes.buffer as ArrayBuffer` idiom; best-effort
  cleanup on failure); `useCreateImageFrame`; `onSaved` + close.
- `shape-overlay.tsx`: image branch (path null) — `useImage(frame.maskUri)`, dim
  Fill, draw mask `<Image>` with `blendMode="clear"` center-fit (skip stroke).
- `shape-picker.tsx`: image-frame chips with signed-URL thumbnail
  (`signFrameUrl`) + trailing "+ Upload" (`cloud-upload-outline`) chip.
- `cropper-screen.tsx`: `uploadOpen` state; render `<UploadFrameModal>`;
  `processImage` signs the mask key just-in-time → `{ kind:'image', maskUri }`
  for `createCutout` (for a just-uploaded local file, pass the local path to skip
  a round-trip); `onSaved` sets `frame`.
- `upload-cutout.ts`: set `custom_frame_id` for `image` kind.

**Verify:** extend `create-cutout.test.ts` — assert `saveLayer`/`restore` wrap,
`setBlendMode(DstIn)`, two `drawImageRect` calls (add those methods to the Skia
mock); `pnpm typecheck`/`lint`/`test`; manual — upload a transparent PNG, confirm
the cutout matches the opaque silhouette and the overlay shows the punched
silhouette; check web fallback.

---

## Cross-cutting notes
- Private `frames` bucket → reused image masks must be fetched via **signed URL**
  (`signFrameUrl`), resolved in `processImage` before `createCutout`.
- Verify Skia 2.6.4 APIs: `BlendMode.DstIn`, `paint.setBlendMode`, no-arg
  `saveLayer`, `SkPath.transform(Matrix)`, `MakeFromSVGString`/`toSVGString`.
- Frame deletion is soft + `items.custom_frame_id ON DELETE SET NULL`, so
  existing cutouts stay intact.
- Patterns to follow: `books/api.ts`, `books/hooks.ts`, `lib/storage.ts`,
  `upload-cutout.ts`, `books/item-picker.tsx` (modal).