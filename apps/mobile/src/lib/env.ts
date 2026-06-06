// Centralised access to the public Expo env vars, with a friendly error if a
// developer forgot to create their .env from .env.example.
const supabaseUrl = process.env.EXPO_PUBLIC_SUPABASE_URL;
const supabaseAnonKey = process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY;

if (!supabaseUrl || !supabaseAnonKey) {
  throw new Error(
    'Missing Supabase config. Copy apps/mobile/.env.example to apps/mobile/.env ' +
      'and set EXPO_PUBLIC_SUPABASE_URL and EXPO_PUBLIC_SUPABASE_ANON_KEY.',
  );
}

// Optional: Sentry crash reporting. Unset in local dev / Expo Go so the SDK
// stays dormant; supplied as a build-time var on the web (Vercel) and native
// (EAS) production builds. When absent, src/lib/sentry.ts is a no-op.
const sentryDsn = process.env.EXPO_PUBLIC_SENTRY_DSN || undefined;

export const env = {
  supabaseUrl,
  supabaseAnonKey,
  sentryDsn,
} as const;
