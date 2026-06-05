// Global test setup, loaded after the test framework is installed.

// Stub the Supabase client so importing a store (or anything down its import
// chain) doesn't boot the real client or run env.ts validation, which would
// throw without EXPO_PUBLIC_SUPABASE_* set in the test environment.
jest.mock('@/lib/supabase', () => ({ supabase: {} }));

// Minimal stub for Skia. Modules like src/cropper/shapes.ts import it at the
// top level, but the pure helpers under test never call into it, so a light
// stub keeps those modules importable without loading native/CanvasKit code.
jest.mock('@shopify/react-native-skia', () => ({
  Skia: {
    Path: { Make: () => ({}) },
    XYWHRect: () => ({}),
  },
  PathOp: { Difference: 0 },
}));
