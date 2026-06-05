// `react-native-url-polyfill` must be imported before supabase-js so that the
// global URL implementation it relies on exists in the React Native runtime.
import 'react-native-url-polyfill/auto';

import AsyncStorage from '@react-native-async-storage/async-storage';
import type { Database } from '@scrappd/shared-types';
import { createClient } from '@supabase/supabase-js';
import { AppState } from 'react-native';

import { env } from './env';

export const supabase = createClient<Database>(
  env.supabaseUrl,
  env.supabaseAnonKey,
  {
    auth: {
      // Persist the session to AsyncStorage so it survives app reloads.
      storage: AsyncStorage,
      autoRefreshToken: true,
      persistSession: true,
      // No URL-based session detection on native (that's a web-only concern).
      detectSessionInUrl: false,
    },
  },
);

// Supabase recommends only auto-refreshing the token while the app is in the
// foreground; pause it in the background to avoid unnecessary work.
AppState.addEventListener('change', (state) => {
  if (state === 'active') {
    supabase.auth.startAutoRefresh();
  } else {
    supabase.auth.stopAutoRefresh();
  }
});
