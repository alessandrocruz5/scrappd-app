import { createURL } from 'expo-linking';
import type { Session, User } from '@supabase/supabase-js';
import { create } from 'zustand';

import { supabase } from '@/lib/supabase';

export type AuthStatus = 'unknown' | 'authenticated' | 'unauthenticated';

type AuthResult = { ok: boolean; message?: string };

type AuthState = {
  status: AuthStatus;
  session: Session | null;
  user: User | null;
  isSubmitting: boolean;
  errorMessage: string | null;
  /**
   * True while the user is in the middle of a password-recovery flow: the
   * recovery link has established a temporary session, but the user has not
   * yet chosen a new password. The AuthGate uses this to pin them on the
   * reset screen instead of dropping them into the tab shell.
   */
  isPasswordRecovery: boolean;

  /**
   * Hydrate from any persisted session and subscribe to future auth changes.
   * Returns an unsubscribe function for cleanup. Mirrors the Flutter
   * AuthProvider.initialize() behaviour.
   */
  initialize: () => () => void;
  signIn: (email: string, password: string) => Promise<AuthResult>;
  signUp: (
    email: string,
    password: string,
    displayName?: string,
  ) => Promise<AuthResult>;
  signOut: () => Promise<void>;
  sendPasswordReset: (email: string) => Promise<AuthResult>;
  /**
   * Exchange the access/refresh tokens carried by a recovery deep link for a
   * temporary session, then flag the recovery flow so the user lands on the
   * reset screen.
   */
  beginPasswordRecovery: (
    accessToken: string,
    refreshToken: string,
  ) => Promise<AuthResult>;
  /** Set the new password and exit the recovery flow. */
  updatePassword: (password: string) => Promise<AuthResult>;
  clearError: () => void;
};

// Translate raw Supabase auth errors into something a user can act on,
// mirroring the Flutter app's mapErrorMessage helper.
export function friendlyAuthError(message: string, fallback: string): string {
  const normalized = message.toLowerCase();
  if (normalized.includes('invalid login credentials')) {
    return 'Incorrect email or password.';
  }
  if (
    normalized.includes('already registered') ||
    normalized.includes('already exists')
  ) {
    return 'An account with this email already exists.';
  }
  if (normalized.includes('email not confirmed')) {
    return 'Please confirm your email address before signing in.';
  }
  if (normalized.includes('network')) {
    return 'Network error. Check your connection and try again.';
  }
  return message || fallback;
}

export const useAuthStore = create<AuthState>((set) => ({
  status: 'unknown',
  session: null,
  user: null,
  isSubmitting: false,
  errorMessage: null,
  isPasswordRecovery: false,

  initialize: () => {
    void supabase.auth.getSession().then(({ data }) => {
      set({
        session: data.session,
        user: data.session?.user ?? null,
        status: data.session ? 'authenticated' : 'unauthenticated',
      });
    });

    const { data: listener } = supabase.auth.onAuthStateChange(
      (event, session) => {
        set((state) => ({
          session,
          user: session?.user ?? null,
          status: session ? 'authenticated' : 'unauthenticated',
          // Supabase emits PASSWORD_RECOVERY when it detects a recovery link in
          // the URL (web). On native we set the flag ourselves from the deep
          // link, so preserve it here and only clear it on sign-out.
          isPasswordRecovery:
            event === 'PASSWORD_RECOVERY'
              ? true
              : event === 'SIGNED_OUT'
                ? false
                : state.isPasswordRecovery,
        }));
      },
    );

    return () => listener.subscription.unsubscribe();
  },

  signIn: async (email, password) => {
    set({ isSubmitting: true, errorMessage: null });
    const { error } = await supabase.auth.signInWithPassword({
      email: email.trim(),
      password,
    });
    set({ isSubmitting: false });
    if (error) {
      const message = friendlyAuthError(
        error.message,
        'Login failed. Please try again.',
      );
      set({ errorMessage: message });
      return { ok: false, message };
    }
    // onAuthStateChange will flip status to authenticated.
    return { ok: true };
  },

  signUp: async (email, password, displayName) => {
    set({ isSubmitting: true, errorMessage: null });
    const { data, error } = await supabase.auth.signUp({
      email: email.trim(),
      password,
      options: {
        data: displayName ? { display_name: displayName.trim() } : undefined,
      },
    });
    set({ isSubmitting: false });
    if (error) {
      const message = friendlyAuthError(
        error.message,
        'Registration failed. Please try again.',
      );
      set({ errorMessage: message });
      return { ok: false, message };
    }
    // When email confirmation is enabled there is no session yet; the user
    // must verify before signing in.
    if (!data.session) {
      return {
        ok: true,
        message: 'Check your email to confirm your account, then sign in.',
      };
    }
    return { ok: true };
  },

  signOut: async () => {
    set({ isSubmitting: true });
    await supabase.auth.signOut();
    set({ isSubmitting: false });
    // onAuthStateChange will flip status to unauthenticated.
  },

  sendPasswordReset: async (email) => {
    set({ isSubmitting: true, errorMessage: null });
    const { error } = await supabase.auth.resetPasswordForEmail(email.trim(), {
      redirectTo: createURL('/reset-password'),
    });
    set({ isSubmitting: false });
    if (error) {
      const message = friendlyAuthError(
        error.message,
        'Could not send reset email.',
      );
      set({ errorMessage: message });
      return { ok: false, message };
    }
    return {
      ok: true,
      message:
        'If an account exists for that email, a reset link is on its way.',
    };
  },

  beginPasswordRecovery: async (accessToken, refreshToken) => {
    set({ errorMessage: null });
    const { error } = await supabase.auth.setSession({
      access_token: accessToken,
      refresh_token: refreshToken,
    });
    if (error) {
      const message = friendlyAuthError(
        error.message,
        'This reset link is invalid or has expired. Request a new one.',
      );
      set({ errorMessage: message });
      return { ok: false, message };
    }
    // setSession emits SIGNED_IN (not PASSWORD_RECOVERY), so flag the flow here
    // to keep the AuthGate pinned on the reset screen.
    set({ isPasswordRecovery: true });
    return { ok: true };
  },

  updatePassword: async (password) => {
    set({ isSubmitting: true, errorMessage: null });
    const { error } = await supabase.auth.updateUser({ password });
    set({ isSubmitting: false });
    if (error) {
      const message = friendlyAuthError(
        error.message,
        'Could not update your password. Please try again.',
      );
      set({ errorMessage: message });
      return { ok: false, message };
    }
    // Recovery complete — clearing the flag lets the AuthGate route the now
    // fully-authenticated user into the app.
    set({ isPasswordRecovery: false });
    return { ok: true };
  },

  clearError: () => set({ errorMessage: null }),
}));
