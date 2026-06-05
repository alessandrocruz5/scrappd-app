// The page substrate, painted with Skia: a solid fill, an optional cover
// background image, and the selected guide pattern (grid / dots / split). This
// is the static layer; the interactive cutouts are Reanimated views laid over
// it (see canvas-item.tsx). Keeping the background in Skia means the same
// renderer that cuts shapes also draws the page, and it composes cleanly with a
// future client-side page export snapshot.

import {
  Canvas,
  Fill,
  Group,
  Image as SkiaImage,
  Path,
  Skia,
  useImage,
} from '@shopify/react-native-skia';
import { useMemo } from 'react';

import type { PatternId } from './templates';

const PATTERN_COLOR = '#D8D2C4';
const GRID_STEP = 48;
const DOT_STEP = 40;
const DOT_RADIUS = 2;

function buildPatternPath(pattern: PatternId, width: number, height: number) {
  const path = Skia.Path.Make();
  if (width <= 0 || height <= 0) return path;

  if (pattern === 'grid') {
    for (let x = GRID_STEP; x < width; x += GRID_STEP) {
      path.moveTo(x, 0);
      path.lineTo(x, height);
    }
    for (let y = GRID_STEP; y < height; y += GRID_STEP) {
      path.moveTo(0, y);
      path.lineTo(width, y);
    }
  } else if (pattern === 'split') {
    const mid = width / 2;
    path.moveTo(mid, 0);
    path.lineTo(mid, height);
    const third = height / 3;
    path.moveTo(0, third);
    path.lineTo(width, third);
  } else if (pattern === 'dots') {
    for (let x = DOT_STEP; x < width; x += DOT_STEP) {
      for (let y = DOT_STEP; y < height; y += DOT_STEP) {
        path.addCircle(x, y, DOT_RADIUS);
      }
    }
  }
  return path;
}

export function PageBackground({
  width,
  height,
  backgroundColor,
  pattern,
  backgroundImageUrl,
}: {
  width: number;
  height: number;
  backgroundColor: string;
  pattern: PatternId;
  backgroundImageUrl?: string | null;
}) {
  const image = useImage(backgroundImageUrl ?? null);
  const path = useMemo(
    () => buildPatternPath(pattern, width, height),
    [pattern, width, height],
  );
  // Dots read as filled, lines as strokes.
  const filled = pattern === 'dots';

  if (width <= 0 || height <= 0) return null;

  return (
    <Canvas style={{ width, height }}>
      <Fill color={backgroundColor} />
      {image ? (
        <SkiaImage
          image={image}
          x={0}
          y={0}
          width={width}
          height={height}
          fit="cover"
        />
      ) : null}
      <Group>
        <Path
          path={path}
          color={PATTERN_COLOR}
          style={filled ? 'fill' : 'stroke'}
          strokeWidth={1}
        />
      </Group>
    </Canvas>
  );
}
