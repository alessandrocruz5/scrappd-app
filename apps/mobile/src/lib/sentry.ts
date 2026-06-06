// Production crash reporting + error tracking, centralised so the rest of the
// app never imports the Sentry SDK directly.
//
// Sentry is opt-in: it only initialises when EXPO_PUBLIC_SENTRY_DSN is set
// (production web/native builds). In local dev / Expo Go the DSN is unset, so
// `initSentry` is a no-op and `captureHandledError` silently does nothing — no
// noise, no network, no dashboard pollution from development runs.

import * as Sentry from '@sentry/react-native';

import { env } from './env';

export const sentryEnabled = !!env.sentryDsn;

let initialised = false;

// Wire up the SDK once, as early as possible in the root layout. Safe to call
// when the DSN is missing (becomes a no-op) and idempotent across fast-refresh.
export function initSentry(): void {
  if (!sentryEnabled || initialised) return;
  initialised = true;

  Sentry.init({
    dsn: env.sentryDsn,
    // Surface the environment in the dashboard so dev/staging crashes (if a DSN
    // is ever wired there) don't mix with production.
    environment: __DEV__ ? 'development' : 'production',
    // Performance tracing is off by default — we only need crash + error
    // reporting for launch. Bump this once we want transaction sampling.
    tracesSampleRate: 0,
  });
}

// Report a handled (caught) error to Sentry with a little structured context so
// failures in a known flow — cutout upload, editor persistence, page export —
// are grouped and searchable. A no-op when Sentry is disabled.
export function captureHandledError(
  error: unknown,
  context: { feature: string; [key: string]: unknown },
): void {
  if (!sentryEnabled) return;
  const err = error instanceof Error ? error : new Error(String(error));
  Sentry.captureException(err, {
    tags: { feature: context.feature },
    extra: context,
  });
}

// Re-exported so screens can use the boundary / wrapper without importing the
// SDK package name directly.
export const ErrorBoundary = Sentry.ErrorBoundary;
export const wrapWithSentry = Sentry.wrap;
