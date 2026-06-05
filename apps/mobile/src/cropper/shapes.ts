// Parametric Skia Path builders for the instant shape cropper.
//
// Every builder takes a bounding `Box` and returns a fresh `SkPath` whose
// silhouette fits inside that box. The same builders drive both the live
// aiming overlay (drawn into the camera viewport) and the offscreen cutout
// surface (clipped before snapshotting), so a shape always looks identical on
// screen and in the saved PNG.

import { PathOp, Skia, type SkPath } from '@shopify/react-native-skia';

export type ShapeId =
  | 'square'
  | 'rectangle'
  | 'circle'
  | 'heart'
  | 'star'
  | 'scallop'
  | 'stamp';

export type Box = {
  x: number;
  y: number;
  width: number;
  height: number;
};

export type ShapeMeta = {
  id: ShapeId;
  label: string;
  // Ionicons glyph used in the shape picker.
  icon: string;
};

// Order shown in the picker. Icons are approximate stand-ins from the Ionicons
// set; the real silhouette is whatever `buildShapePath` draws.
export const SHAPES: ShapeMeta[] = [
  { id: 'square', label: 'Square', icon: 'square-outline' },
  { id: 'rectangle', label: 'Rectangle', icon: 'tablet-landscape-outline' },
  { id: 'circle', label: 'Circle', icon: 'ellipse-outline' },
  { id: 'heart', label: 'Heart', icon: 'heart-outline' },
  { id: 'star', label: 'Star', icon: 'star-outline' },
  { id: 'scallop', label: 'Scallop', icon: 'flower-outline' },
  { id: 'stamp', label: 'Stamp', icon: 'mail-outline' },
];

/**
 * The shape's bounding box inside a square canvas of `size`, leaving a margin
 * so strokes/edges never touch the frame. Shared by the overlay and the cutout
 * renderer so they stay in lockstep.
 */
export function insetBox(size: number, marginRatio = 0.06): Box {
  const margin = size * marginRatio;
  return {
    x: margin,
    y: margin,
    width: size - margin * 2,
    height: size - margin * 2,
  };
}

export function buildShapePath(shape: ShapeId, box: Box): SkPath {
  switch (shape) {
    case 'square':
      return square(box);
    case 'rectangle':
      return rectangle(box);
    case 'circle':
      return circle(box);
    case 'heart':
      return heart(box);
    case 'star':
      return star(box);
    case 'scallop':
      return scallop(box);
    case 'stamp':
      return stamp(box);
  }
}

// ---------------------------------------------------------------------------
// Builders
// ---------------------------------------------------------------------------

// Centred square with side = the smaller box dimension.
function square(box: Box): SkPath {
  const path = Skia.Path.Make();
  const side = Math.min(box.width, box.height);
  const x = box.x + (box.width - side) / 2;
  const y = box.y + (box.height - side) / 2;
  path.addRect(Skia.XYWHRect(x, y, side, side));
  return path;
}

// Fills the whole box (the box itself is already wider than tall when the
// caller wants a landscape rectangle; here we use a 4:3 letterbox inside it).
function rectangle(box: Box): SkPath {
  const path = Skia.Path.Make();
  const w = box.width;
  const h = Math.min(box.height, box.width * 0.72);
  const y = box.y + (box.height - h) / 2;
  path.addRect(Skia.XYWHRect(box.x, y, w, h));
  return path;
}

// Centred circle with radius = half the smaller box dimension.
function circle(box: Box): SkPath {
  const path = Skia.Path.Make();
  const r = Math.min(box.width, box.height) / 2;
  path.addCircle(box.x + box.width / 2, box.y + box.height / 2, r);
  return path;
}

// Symmetric heart built from four cubic segments over a unit template, scaled
// into the box. Unit coordinates are y-down to match Skia's canvas.
function heart(box: Box): SkPath {
  const path = Skia.Path.Make();
  const px = (u: number) => box.x + u * box.width;
  const py = (v: number) => box.y + v * box.height;

  path.moveTo(px(0.5), py(0.32));
  // Left lobe.
  path.cubicTo(px(0.5), py(0.1), px(0.08), py(0.08), px(0.06), py(0.36));
  // Left side down to the bottom tip.
  path.cubicTo(px(0.05), py(0.58), px(0.32), py(0.74), px(0.5), py(0.94));
  // Right side back up.
  path.cubicTo(px(0.68), py(0.74), px(0.95), py(0.58), px(0.94), py(0.36));
  // Right lobe back to the centre dip.
  path.cubicTo(px(0.92), py(0.08), px(0.5), py(0.1), px(0.5), py(0.32));
  path.close();
  return path;
}

// Five-point star. Outer vertices on a circle, inner vertices at the golden
// ratio so the points stay crisp. Starts at the top (-90°).
function star(box: Box): SkPath {
  const path = Skia.Path.Make();
  const cx = box.x + box.width / 2;
  const cy = box.y + box.height / 2;
  const outer = Math.min(box.width, box.height) / 2;
  const inner = outer * 0.382;
  const points = 5;

  for (let i = 0; i < points * 2; i += 1) {
    const radius = i % 2 === 0 ? outer : inner;
    const angle = -Math.PI / 2 + (i * Math.PI) / points;
    const x = cx + radius * Math.cos(angle);
    const y = cy + radius * Math.sin(angle);
    if (i === 0) {
      path.moveTo(x, y);
    } else {
      path.lineTo(x, y);
    }
  }
  path.close();
  return path;
}

// Scalloped circle (flower edge): a base disc unioned with a ring of small
// discs. Skia's default non-zero winding fills the union, so the overlapping
// circles read as rounded petals around the rim.
function scallop(box: Box): SkPath {
  const path = Skia.Path.Make();
  const cx = box.x + box.width / 2;
  const cy = box.y + box.height / 2;
  const r = Math.min(box.width, box.height) / 2;

  const petals = 12;
  const petalRadius = r * 0.26;
  // Push petal centres out so each bump just reaches the box edge.
  const ringRadius = r - petalRadius;

  path.addCircle(cx, cy, ringRadius);
  for (let i = 0; i < petals; i += 1) {
    const angle = (i * 2 * Math.PI) / petals;
    const px = cx + ringRadius * Math.cos(angle);
    const py = cy + ringRadius * Math.sin(angle);
    path.addCircle(px, py, petalRadius);
  }
  return path;
}

// Postage stamp: a rectangle with semicircular perforations bitten out of every
// edge, leaving the classic toothed border. Built as a true geometric
// difference so there are no stray bumps outside the rectangle.
function stamp(box: Box): SkPath {
  const side = Math.min(box.width, box.height);
  const x = box.x + (box.width - side) / 2;
  const y = box.y + (box.height - side) / 2;

  const base = Skia.Path.Make();
  base.addRect(Skia.XYWHRect(x, y, side, side));

  const teeth = 9;
  const step = side / teeth;
  const toothRadius = step * 0.42;

  const holes = Skia.Path.Make();
  for (let i = 0; i < teeth; i += 1) {
    const offset = step * (i + 0.5);
    // Top and bottom edges.
    holes.addCircle(x + offset, y, toothRadius);
    holes.addCircle(x + offset, y + side, toothRadius);
    // Left and right edges.
    holes.addCircle(x, y + offset, toothRadius);
    holes.addCircle(x + side, y + offset, toothRadius);
  }

  // base := base - holes
  base.op(holes, PathOp.Difference);
  return base;
}
