import { bringToFrontZ, roundTransform, sendToBackZ } from '@/editor/transform';

describe('roundTransform', () => {
  it('rounds position and size to whole pixels', () => {
    expect(
      roundTransform({
        x: 10.4,
        y: 20.6,
        width: 199.5,
        height: 60.2,
        rotation: 0,
      }),
    ).toEqual({
      position_x: 10,
      position_y: 21,
      width: 200,
      height: 60,
      rotation: 0,
    });
  });

  it('keeps rotation to two decimal places', () => {
    expect(
      roundTransform({ x: 0, y: 0, width: 0, height: 0, rotation: 12.3456 }),
    ).toMatchObject({ rotation: 12.35 });
    expect(
      roundTransform({ x: 0, y: 0, width: 0, height: 0, rotation: -45.004 }),
    ).toMatchObject({ rotation: -45 });
  });

  it('does not invent precision the persisted row never carries', () => {
    const patch = roundTransform({
      x: 0.49,
      y: -0.49,
      width: 1.5,
      height: 2.5,
      rotation: 359.999,
    });
    expect(Number.isInteger(patch.position_x)).toBe(true);
    expect(Number.isInteger(patch.width)).toBe(true);
    // Banker's-free: Math.round goes half-up.
    expect(patch.width).toBe(2);
    expect(patch.height).toBe(3);
    expect(patch.rotation).toBe(360);
  });
});

describe('bringToFrontZ', () => {
  const items = [{ z_index: 1 }, { z_index: 2 }, { z_index: 5 }];

  it('returns one past the current max for an item not already on top', () => {
    expect(bringToFrontZ(items, { z_index: 1 })).toBe(6);
  });

  it('returns null when the item is already on top (no write needed)', () => {
    expect(bringToFrontZ(items, { z_index: 5 })).toBeNull();
  });

  it('returns null for an empty canvas', () => {
    expect(bringToFrontZ([], { z_index: 0 })).toBeNull();
  });

  it('handles negative z-indexes from prior send-to-backs', () => {
    expect(
      bringToFrontZ([{ z_index: -3 }, { z_index: -1 }], { z_index: -3 }),
    ).toBe(0);
  });
});

describe('sendToBackZ', () => {
  const items = [{ z_index: 1 }, { z_index: 2 }, { z_index: 5 }];

  it('returns one below the current min for an item not already at the back', () => {
    expect(sendToBackZ(items, { z_index: 5 })).toBe(0);
  });

  it('returns null when the item is already at the back', () => {
    expect(sendToBackZ(items, { z_index: 1 })).toBeNull();
  });

  it('returns null for an empty canvas', () => {
    expect(sendToBackZ([], { z_index: 0 })).toBeNull();
  });

  it('keeps walking negative as items are repeatedly sent back', () => {
    expect(sendToBackZ([{ z_index: -2 }, { z_index: 3 }], { z_index: 3 })).toBe(
      -3,
    );
  });
});
