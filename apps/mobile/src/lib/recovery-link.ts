import { useURL } from 'expo-linking';
import { useEffect } from 'react';
import { Platform } from 'react-native';

import { useAuthStore } from '@/stores/auth-store';

export type RecoveryParams = {
  type: string | null;
  accessToken: string | null;
  refreshToken: string | null;
  error: string | null;
  errorDescription: string | null;
};

/**
 * Pull the recovery tokens out of an incoming auth link. Supabase's implicit
 * flow returns them in the URL fragment (#access_token=...&type=recovery),
 * while errors can arrive in either the query string or the fragment, so we
 * merge both. Works for native `scrappd://` deep links and the web reset URL
 * alike.
 */
export function parseRecoveryParams(url: string): RecoveryParams {
  const params = new URLSearchParams();
  const append = (segment: string | undefined) => {
    if (!segment) return;
    new URLSearchParams(segment).forEach((value, key) =>
      params.set(key, value),
    );
  };

  const hashIndex = url.indexOf('#');
  const queryIndex = url.indexOf('?');

  if (queryIndex !== -1) {
    const end =
      hashIndex !== -1 && hashIndex > queryIndex ? hashIndex : undefined;
    append(url.slice(queryIndex + 1, end));
  }
  if (hashIndex !== -1) {
    append(url.slice(hashIndex + 1));
  }

  return {
    type: params.get('type'),
    accessToken: params.get('access_token'),
    refreshToken: params.get('refresh_token'),
    error: params.get('error'),
    errorDescription: params.get('error_description'),
  };
}

/**
 * Watch for an incoming password-recovery link and hand its tokens to the
 * auth store. On web `useURL` returns the current location (including the hash
 * Supabase appends); on native it fires for the deep link that opened the app.
 */
export function useRecoveryLink(): void {
  const url = useURL();
  const beginPasswordRecovery = useAuthStore((s) => s.beginPasswordRecovery);

  useEffect(() => {
    if (!url) return;
    const { type, accessToken, refreshToken } = parseRecoveryParams(url);
    if (type !== 'recovery' || !accessToken || !refreshToken) return;

    void beginPasswordRecovery(accessToken, refreshToken).then(() => {
      // Strip the tokens from the browser URL so they aren't left in history
      // or copied around. Native deep links have no address bar to clean.
      if (Platform.OS === 'web' && typeof window !== 'undefined') {
        window.history.replaceState(
          null,
          '',
          window.location.pathname + window.location.search,
        );
      }
    });
  }, [url, beginPasswordRecovery]);
}
