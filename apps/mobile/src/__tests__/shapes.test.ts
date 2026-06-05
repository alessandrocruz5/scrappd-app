import { insetBox } from '@/cropper/shapes';

describe('insetBox', () => {
  it('insets by the default 6% margin on all sides', () => {
    expect(insetBox(1000)).toEqual({
      x: 60,
      y: 60,
      width: 880,
      height: 880,
    });
  });

  it('honours an explicit margin ratio', () => {
    expect(insetBox(1000, 0.06)).toEqual({
      x: 60,
      y: 60,
      width: 880,
      height: 880,
    });
    expect(insetBox(200, 0.1)).toEqual({
      x: 20,
      y: 20,
      width: 160,
      height: 160,
    });
  });

  it('stays symmetric: the box is centred with equal margins', () => {
    const size = 500;
    const box = insetBox(size, 0.08);
    expect(box.x).toBe(box.y);
    expect(box.x + box.width + box.x).toBe(size);
    expect(box.y + box.height + box.y).toBe(size);
  });

  it('collapses to a zero box when the margin takes the whole canvas', () => {
    expect(insetBox(100, 0.5)).toEqual({
      x: 50,
      y: 50,
      width: 0,
      height: 0,
    });
  });
});
