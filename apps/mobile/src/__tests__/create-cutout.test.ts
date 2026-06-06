/* eslint-disable import/first -- jest.mock() must be hoisted above the imports it stubs. */
// Drives the on-device cutout pipeline (decode → offscreen surface → clip →
// snapshot → encode) against a fully mocked Skia, asserting both the happy-path
// output and every guard that turns a missing/failed Skia primitive into a
// user-facing Error.

// Mutable state the Skia mock reads at call time, so each test can swap in a
// null image / surface / encode result to exercise a specific guard.
const mockSkiaState: {
  image: unknown;
  surface: unknown;
  bytes: Uint8Array | null;
  base64: string;
} = { image: null, surface: null, bytes: null, base64: '' };

jest.mock('@shopify/react-native-skia', () => ({
  Skia: {
    Data: { fromURI: jest.fn(async () => ({})) },
    Image: { MakeImageFromEncoded: jest.fn(() => mockSkiaState.image) },
    Surface: { MakeOffscreen: jest.fn(() => mockSkiaState.surface) },
    Color: jest.fn(() => 0),
    XYWHRect: (x: number, y: number, w: number, h: number) => [x, y, w, h],
    Paint: jest.fn(() => ({ setAntiAlias: jest.fn() })),
    Path: {
      Make: () => ({
        addRect() {},
        addCircle() {},
        moveTo() {},
        lineTo() {},
        cubicTo() {},
        close() {},
        op() {},
      }),
    },
  },
  ClipOp: { Intersect: 1 },
  ImageFormat: { PNG: 4 },
  PathOp: { Difference: 7 },
}));

import { createCutout, CUTOUT_SIZE } from '@/cropper/create-cutout';

function freshSurface() {
  const snapshot = {
    encodeToBytes: jest.fn(() => mockSkiaState.bytes),
    encodeToBase64: jest.fn(() => mockSkiaState.base64),
  };
  const canvas = {
    clear: jest.fn(),
    save: jest.fn(),
    clipPath: jest.fn(),
    drawImageRect: jest.fn(),
    restore: jest.fn(),
  };
  return {
    snapshot,
    canvas,
    surface: {
      getCanvas: () => canvas,
      flush: jest.fn(),
      makeImageSnapshot: () => snapshot,
    },
  };
}

beforeEach(() => {
  const { surface } = freshSurface();
  mockSkiaState.image = { width: () => 2000, height: () => 1000 };
  mockSkiaState.surface = surface;
  mockSkiaState.bytes = new Uint8Array([1, 2, 3, 4]);
  mockSkiaState.base64 = 'QUJDRA==';
});

describe('createCutout', () => {
  it('produces a square PNG cutout at CUTOUT_SIZE', async () => {
    const result = await createCutout('file:///photo.jpg', 'circle');
    expect(result.width).toBe(CUTOUT_SIZE);
    expect(result.height).toBe(CUTOUT_SIZE);
    expect(result.bytes).toEqual(new Uint8Array([1, 2, 3, 4]));
    expect(result.base64).toBe('QUJDRA==');
  });

  it('clips through the shape path before drawing the image', async () => {
    const { surface, canvas } = freshSurface();
    mockSkiaState.surface = surface;
    await createCutout('file:///photo.jpg', 'heart');
    // The clip must happen, and the image must be drawn within a save/restore.
    expect(canvas.save).toHaveBeenCalled();
    expect(canvas.clipPath).toHaveBeenCalled();
    expect(canvas.drawImageRect).toHaveBeenCalledTimes(1);
    expect(canvas.restore).toHaveBeenCalled();
  });

  it('throws a friendly error when the image cannot be decoded', async () => {
    mockSkiaState.image = null;
    await expect(createCutout('file:///bad.jpg', 'square')).rejects.toThrow(
      /Could not read that image/,
    );
  });

  it('throws when the offscreen surface cannot be allocated', async () => {
    mockSkiaState.surface = null;
    await expect(createCutout('file:///photo.jpg', 'square')).rejects.toThrow(
      /Could not allocate the drawing surface/,
    );
  });

  it('throws when the snapshot cannot be encoded to bytes', async () => {
    mockSkiaState.bytes = null;
    await expect(createCutout('file:///photo.jpg', 'star')).rejects.toThrow(
      /Could not encode the cutout/,
    );
  });
});
