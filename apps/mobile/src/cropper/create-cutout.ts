// Turns a framed photo into a transparent-background PNG cutout.
//
// Pipeline: decode the source image with Skia, draw a centre-cover crop of it
// into a square offscreen surface, clip that draw through the selected shape
// Path, then snapshot the surface. Everything outside the shape stays fully
// transparent, so the encoded PNG is a clean cutout. This is pure Skia and
// runs entirely on-device — no server round trip.

import { ClipOp, ImageFormat, Skia } from '@shopify/react-native-skia';

import { buildShapePath, insetBox, type ShapeId } from './shapes';

// Square output resolution for every cutout. Big enough to look crisp on a
// page canvas, small enough to upload quickly.
export const CUTOUT_SIZE = 1080;

export type Cutout = {
  bytes: Uint8Array;
  // Same pixels as a base64 PNG, handy for an instant <Image> preview.
  base64: string;
  width: number;
  height: number;
};

/**
 * Build a shaped cutout from an image on disk (camera capture or library pick).
 * Throws if the image can't be decoded or the surface can't be allocated.
 */
export async function createCutout(
  imageUri: string,
  shape: ShapeId,
): Promise<Cutout> {
  const data = await Skia.Data.fromURI(imageUri);
  const image = Skia.Image.MakeImageFromEncoded(data);
  if (!image) {
    throw new Error('Could not read that image. Try another shot.');
  }

  const surface = Skia.Surface.MakeOffscreen(CUTOUT_SIZE, CUTOUT_SIZE);
  if (!surface) {
    throw new Error('Could not allocate the drawing surface.');
  }

  const canvas = surface.getCanvas();
  // Transparent background — the bytes outside the shape must carry no colour.
  canvas.clear(Skia.Color('transparent'));

  const path = buildShapePath(shape, insetBox(CUTOUT_SIZE));

  canvas.save();
  canvas.clipPath(path, ClipOp.Intersect, true);

  // Centre-cover the source into the square output: take the largest centred
  // square of the source and scale it to fill. This matches the square camera
  // viewport, so what the user frames is what they get.
  const imgW = image.width();
  const imgH = image.height();
  const srcSide = Math.min(imgW, imgH);
  const srcRect = Skia.XYWHRect(
    (imgW - srcSide) / 2,
    (imgH - srcSide) / 2,
    srcSide,
    srcSide,
  );
  const destRect = Skia.XYWHRect(0, 0, CUTOUT_SIZE, CUTOUT_SIZE);

  const paint = Skia.Paint();
  paint.setAntiAlias(true);
  canvas.drawImageRect(image, srcRect, destRect, paint);
  canvas.restore();

  surface.flush();
  const snapshot = surface.makeImageSnapshot();

  const bytes = snapshot.encodeToBytes(ImageFormat.PNG, 100);
  const base64 = snapshot.encodeToBase64(ImageFormat.PNG, 100);
  if (!bytes) {
    throw new Error('Could not encode the cutout.');
  }

  return {
    bytes,
    base64,
    width: CUTOUT_SIZE,
    height: CUTOUT_SIZE,
  };
}
