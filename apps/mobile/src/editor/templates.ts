// Layout templates and background patterns for the page editor.
//
// A template is just a named pairing of a background colour and a guide
// pattern, ported from the Flutter editor's _PageTemplate (Clean / Grid /
// Split) and extended with a couple more. Applying one writes background_color
// + background_pattern to content.pages, and stores the template id in
// layout_template so the chosen chip can be restored on reload. The patterns
// themselves are painted with Skia (see page-background.tsx).

export type PatternId = 'none' | 'grid' | 'dots' | 'split';

export type Template = {
  id: string;
  name: string;
  backgroundColor: string;
  pattern: PatternId;
};

export const TEMPLATES: Template[] = [
  { id: 'clean', name: 'Clean', backgroundColor: '#F9FAFB', pattern: 'none' },
  { id: 'grid', name: 'Grid', backgroundColor: '#FFFFFF', pattern: 'grid' },
  { id: 'dots', name: 'Dots', backgroundColor: '#FFFDF7', pattern: 'dots' },
  { id: 'split', name: 'Split', backgroundColor: '#F8FAFC', pattern: 'split' },
];

// A small palette of solid background swatches the user can apply on top of a
// template, matching the 7-char hex stored by content.pages.background_color.
export const BACKGROUND_SWATCHES: string[] = [
  '#FFFFFF',
  '#F9FAFB',
  '#F6E8C9',
  '#FCE7F3',
  '#E0F2FE',
  '#DCFCE7',
  '#1B0E03',
];

export function isPatternId(value: string | null | undefined): value is PatternId {
  return (
    value === 'none' || value === 'grid' || value === 'dots' || value === 'split'
  );
}
