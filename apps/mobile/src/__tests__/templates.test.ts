import {
  BACKGROUND_SWATCHES,
  TEMPLATES,
  isPatternId,
} from '@/editor/templates';

describe('isPatternId', () => {
  it('accepts every known pattern id', () => {
    for (const id of ['none', 'grid', 'dots', 'split']) {
      expect(isPatternId(id)).toBe(true);
    }
  });

  it('rejects unknown strings and nullish values', () => {
    expect(isPatternId('lines')).toBe(false);
    expect(isPatternId('')).toBe(false);
    expect(isPatternId(null)).toBe(false);
    expect(isPatternId(undefined)).toBe(false);
  });
});

const HEX = /^#[0-9A-Fa-f]{6}$/;

describe('template data', () => {
  it('pairs every template with a valid pattern id and 7-char hex colour', () => {
    for (const template of TEMPLATES) {
      expect(isPatternId(template.pattern)).toBe(true);
      expect(template.backgroundColor).toMatch(HEX);
    }
  });

  it('exposes only valid 7-char hex background swatches', () => {
    for (const swatch of BACKGROUND_SWATCHES) {
      expect(swatch).toMatch(HEX);
    }
  });
});
