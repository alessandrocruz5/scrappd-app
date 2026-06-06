// Pure transform / stacking math for the page editor, split out from the
// gesture component (canvas-item.tsx) and the editor screen (page-editor.tsx)
// so the persistence-facing arithmetic can be unit-tested without a Skia canvas
// or a live gesture. These functions take plain numbers and return the exact
// patch that gets written to content.page_items.

export type LiveTransform = {
  x: number;
  y: number;
  width: number;
  height: number;
  // Degrees.
  rotation: number;
};

export type RoundedTransform = {
  position_x: number;
  position_y: number;
  width: number;
  height: number;
  rotation: number;
};

// Snap a live (sub-pixel) gesture transform to the integer-ish values we
// persist. Position and size are whole pixels; rotation keeps two decimals so a
// twist doesn't churn the row with meaningless precision.
export function roundTransform(t: LiveTransform): RoundedTransform {
  return {
    position_x: Math.round(t.x),
    position_y: Math.round(t.y),
    width: Math.round(t.width),
    height: Math.round(t.height),
    rotation: Math.round(t.rotation * 100) / 100,
  };
}

// The z_index that brings `selected` above everything else, or null when it's
// already on top (no write needed). Returns one past the current max.
export function bringToFrontZ(
  items: readonly { z_index: number }[],
  selected: { z_index: number },
): number | null {
  if (items.length === 0) return null;
  const max = Math.max(...items.map((i) => i.z_index));
  return selected.z_index === max ? null : max + 1;
}

// The z_index that sends `selected` behind everything else, or null when it's
// already at the back. Returns one below the current min.
export function sendToBackZ(
  items: readonly { z_index: number }[],
  selected: { z_index: number },
): number | null {
  if (items.length === 0) return null;
  const min = Math.min(...items.map((i) => i.z_index));
  return selected.z_index === min ? null : min - 1;
}
