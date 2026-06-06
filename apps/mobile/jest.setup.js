// Global test setup, loaded after the test framework is installed.

// Satisfy src/lib/env.ts, which throws at import time when the public Supabase
// vars are missing. Anything that pulls env.ts into its import chain (the
// Supabase client, the Sentry wrapper) would otherwise crash the suite in CI,
// where these aren't set. Dummy values are fine — no test hits a real backend.
process.env.EXPO_PUBLIC_SUPABASE_URL ||= 'http://127.0.0.1:54321';
process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY ||= 'test-anon-key';

// Stub the Supabase client so importing a store (or anything down its import
// chain) doesn't boot the real client.
jest.mock('@/lib/supabase', () => ({ supabase: {} }));

// Mock the Sentry native SDK (not our wrapper) so src/lib/sentry.ts still runs
// — and stays covered — without pulling in native modules. With no DSN set the
// wrapper is a no-op, so these stubs are never actually invoked.
jest.mock('@sentry/react-native', () => ({
  init: jest.fn(),
  captureException: jest.fn(),
  ErrorBoundary: ({ children }) => children,
  wrap: (component) => component,
}));

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
