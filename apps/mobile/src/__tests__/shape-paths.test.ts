/* eslint-disable import/first -- jest.mock() must be hoisted above the imports it stubs. */
// Verifies buildShapePath dispatches to the right builder and emits the right
// primitive Skia path operations for each shape. We swap in a recording Skia
// mock so the pure geometry can be asserted without CanvasKit: every Path
// records the calls made against it, and XYWHRect echoes its args.

type Op = { method: string; args: number[] };

// The recording path is defined inside the mock factory (jest hoists the
// factory and forbids it referencing out-of-scope names). Each Path records
// every primitive call so the test can assert the emitted geometry.
jest.mock('@shopify/react-native-skia', () => {
  const makePath = () => {
    const ops: { method: string; args: number[] }[] = [];
    const push = (method: string, args: number[]) => {
      ops.push({ method, args });
      return path;
    };
    const path = {
      ops,
      addRect: (rect: number[]) => push('addRect', rect),
      addCircle: (cx: number, cy: number, r: number) =>
        push('addCircle', [cx, cy, r]),
      moveTo: (x: number, y: number) => push('moveTo', [x, y]),
      lineTo: (x: number, y: number) => push('lineTo', [x, y]),
      cubicTo: (...args: number[]) => push('cubicTo', args),
      close: () => push('close', []),
      op: (_other: unknown, pathOp: number) => push('op', [pathOp]),
    };
    return path;
  };
  return {
    Skia: {
      Path: { Make: makePath },
      XYWHRect: (x: number, y: number, w: number, h: number) => [x, y, w, h],
    },
    PathOp: { Difference: 7 },
  };
});

import { buildShapePath, insetBox, SHAPES } from '@/cropper/shapes';

const box = insetBox(1000);

function ops(shape: Parameters<typeof buildShapePath>[0]): Op[] {
  return (buildShapePath(shape, box) as unknown as { ops: Op[] }).ops;
}

function counts(o: Op[]) {
  return o.reduce<Record<string, number>>((acc, { method }) => {
    acc[method] = (acc[method] ?? 0) + 1;
    return acc;
  }, {});
}

describe('SHAPES registry', () => {
  it('every registered shape id builds a path without throwing', () => {
    for (const meta of SHAPES) {
      expect(() => buildShapePath(meta.id, box)).not.toThrow();
    }
  });

  it('exposes a label and icon for each shape', () => {
    for (const meta of SHAPES) {
      expect(meta.label.length).toBeGreaterThan(0);
      expect(meta.icon.length).toBeGreaterThan(0);
    }
  });
});

describe('buildShapePath geometry', () => {
  it('square is a single centred rect sized to the smaller dimension', () => {
    const o = ops('square');
    expect(counts(o)).toEqual({ addRect: 1 });
    const [x, y, w, h] = o[0].args;
    expect(w).toBe(h); // square
    // Centred inside the inset box (which is itself square here).
    expect(x).toBe(box.x);
    expect(y).toBe(box.y);
  });

  it('rectangle is a single letterboxed rect, shorter than the box', () => {
    const o = ops('rectangle');
    expect(counts(o)).toEqual({ addRect: 1 });
    const [, , w, h] = o[0].args;
    expect(w).toBe(box.width);
    expect(h).toBeLessThan(box.height);
  });

  it('circle is a single centred addCircle', () => {
    const o = ops('circle');
    expect(counts(o)).toEqual({ addCircle: 1 });
    const [cx, cy, r] = o[0].args;
    expect(cx).toBe(box.x + box.width / 2);
    expect(cy).toBe(box.y + box.height / 2);
    expect(r).toBe(Math.min(box.width, box.height) / 2);
  });

  it('heart is four cubic segments closed into a loop', () => {
    expect(counts(ops('heart'))).toEqual({
      moveTo: 1,
      cubicTo: 4,
      close: 1,
    });
  });

  it('star is a closed 10-vertex polyline (5 outer + 5 inner points)', () => {
    const c = counts(ops('star'));
    expect(c.moveTo).toBe(1);
    expect(c.lineTo).toBe(9);
    expect(c.close).toBe(1);
  });

  it('scallop unions a base disc with a ring of 12 petals', () => {
    // 1 base ring + 12 petal discs = 13 circles.
    expect(counts(ops('scallop'))).toEqual({ addCircle: 13 });
  });

  it('stamp is a rect with perforations bitten out via a path difference', () => {
    const o = ops('stamp');
    const c = counts(o);
    expect(c.addRect).toBe(1);
    expect(c.op).toBe(1);
    // The subtraction uses the Difference op.
    const opCall = o.find((entry) => entry.method === 'op');
    expect(opCall?.args[0]).toBe(7); // PathOp.Difference from the mock
  });
});
