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

export const env = {
  supabaseUrl,
  supabaseAnonKey,
} as const;
