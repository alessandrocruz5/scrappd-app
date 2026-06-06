import { QueryClientProvider } from '@tanstack/react-query';
import { Stack, useRouter, useSegments } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { ErrorFallback } from '@/components/error-fallback';
import { SplashScreen } from '@/components/splash-screen';
import { queryClient } from '@/lib/query-client';
import { ErrorBoundary, initSentry, wrapWithSentry } from '@/lib/sentry';
import { useAuthStore } from '@/stores/auth-store';
import { colors } from '@/theme/colors';

// Initialise crash reporting before anything renders. No-op unless
// EXPO_PUBLIC_SENTRY_DSN is set (production web/native builds).
initSentry();

// Routes the user between the auth stack and the tab shell based on session
// state — the React Native equivalent of the Flutter root_screen.dart gate.
function AuthGate() {
  const status = useAuthStore((s) => s.status);
  const segments = useSegments();
  const router = useRouter();

  useEffect(() => {
    if (status === 'unknown') return;

    const inAuthGroup = segments[0] === '(auth)';

    if (status === 'unauthenticated' && !inAuthGroup) {
      router.replace('/(auth)/login');
    } else if (status === 'authenticated' && inAuthGroup) {
      router.replace('/(tabs)');
    }
  }, [status, segments, router]);

  if (status === 'unknown') {
    return <SplashScreen />;
  }

  // A Stack at the root so the Books -> Pages -> editor flow pushes screens with
  // native headers and back gestures. The tab shell and auth stack render their
  // own headers, so they opt out here.
  return (
    <Stack
      screenOptions={{
        headerStyle: { backgroundColor: colors.primary },
        headerTintColor: colors.white,
        headerTitleStyle: { fontWeight: '700' },
        contentStyle: { backgroundColor: colors.background },
      }}
    >
      <Stack.Screen name="index" options={{ headerShown: false }} />
      <Stack.Screen name="(auth)" options={{ headerShown: false }} />
      <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
      <Stack.Screen name="book/[id]" options={{ title: 'Book' }} />
      <Stack.Screen name="page/[id]" options={{ title: 'Page' }} />
    </Stack>
  );
}

function RootLayout() {
  const initialize = useAuthStore((s) => s.initialize);

  useEffect(() => {
    // Hydrate the session and subscribe to auth changes; clean up on unmount.
    const unsubscribe = initialize();
    return unsubscribe;
  }, [initialize]);

  return (
    // The boundary reports unhandled render errors to Sentry and shows a
    // friendly fallback instead of a blank screen. `showDialog` is native-only
    // and off by default — we use our own fallback UI.
    <ErrorBoundary
      fallback={({ resetError }) => <ErrorFallback resetError={resetError} />}
    >
      <GestureHandlerRootView style={{ flex: 1 }}>
        <QueryClientProvider client={queryClient}>
          <SafeAreaProvider>
            <StatusBar style="light" />
            <AuthGate />
          </SafeAreaProvider>
        </QueryClientProvider>
      </GestureHandlerRootView>
    </ErrorBoundary>
  );
}

// `Sentry.wrap` enables touch/navigation breadcrumbs and ties the root into the
// SDK. It's a transparent pass-through when Sentry isn't initialised.
export default wrapWithSentry(RootLayout);
