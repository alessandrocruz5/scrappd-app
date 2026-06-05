// The aiming mask drawn over the live camera viewport.
//
// A dim layer covers the whole square viewport with the selected shape punched
// clear (BlendMode.Clear inside an offscreen layer), plus a bright outline so
// the user can line the shape up against a pattern or poster. The same
// `buildShapePath` + `insetBox` feed the cutout renderer, so this preview is
// what gets saved.

import { useMemo } from 'react';
import { Canvas, Fill, Group, Path } from '@shopify/react-native-skia';

import { colors } from '@/theme/colors';

import { buildShapePath, insetBox, type ShapeId } from './shapes';

type ShapeOverlayProps = {
  shape: ShapeId;
  // Side length of the (square) camera viewport in px.
  size: number;
};

export function ShapeOverlay({ shape, size }: ShapeOverlayProps) {
  const path = useMemo(
    () => buildShapePath(shape, insetBox(size)),
    [shape, size],
  );

  return (
    <Canvas style={{ width: size, height: size }} pointerEvents="none">
      {/* Dim everything, then clear the shape's interior on its own layer. */}
      <Group layer>
        <Fill color="rgba(27, 14, 3, 0.5)" />
        <Path path={path} color="black" blendMode="clear" />
      </Group>
      {/* Bright aiming outline on top. */}
      <Path
        path={path}
        style="stroke"
        strokeWidth={size * 0.01}
        color={colors.accentAlt2}
      />
    </Canvas>
  );
}
