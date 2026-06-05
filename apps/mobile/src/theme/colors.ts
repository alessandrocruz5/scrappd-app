// Scrappd brand palette, carried over from the retired Flutter app's
// theme_constants.dart so the React Native app keeps the same look.
export const colors = {
  primary: '#AD2A1A',
  secondary: '#510420',
  accent: '#C05D23',
  accentAlt: '#CB7D2B',
  accentAlt2: '#D7AD3E',
  black: '#1B0E03',
  white: '#F6E8C9',
  background: '#F6E8C9',
  surface: '#FFFFFF',
  border: '#E5D5B8',
  error: '#EF4444',
  success: '#10B981',
  textPrimary: '#1B0E03',
  textSecondary: '#4A2A1A',
  textHint: '#6E4A33',
} as const;

export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  xxl: 32,
} as const;

export const radius = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
} as const;
