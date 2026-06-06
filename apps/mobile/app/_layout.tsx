import { QueryClientProvider } from '@tanstack/react-query';
import { Stack, useRouter, useSegments } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { SplashScreen } from '@/components/splash-screen';
import { queryClient } from '@/lib/query-client';
import { useRecoveryLink } from '@/lib/recovery-link';
import { useAuthStore } from '@/stores/auth-store';
import { colors } from '@/theme/colors';

// Routes the user between the auth stack and the tab shell based on session
// state — the React Native equivalent of the Flutter root_screen.dart gate.
function AuthGate() {
  const status = useAuthStore((s) => s.status);
  const isPasswordRecovery = useAuthStore((s) => s.isPasswordRecovery);
  const segments = useSegments();
  const router = useRouter();

  // Consume any incoming password-recovery deep link (native + web).
  useRecoveryLink();

  useEffect(() => {
    if (status === 'unknown') return;

    const inAuthGroup = segments[0] === '(auth)';

    // A recovery link establishes a real session, so without this the user
    // would be dropped straight into the app. Pin them on the reset screen
    // until they pick a new password (which clears the flag).
    if (isPasswordRecovery) {
      const onResetScreen =
        inAuthGroup && (segments as string[]).includes('reset-password');
      if (!onResetScreen) {
        router.replace('/(auth)/reset-password');
      }
      return;
    }

    if (status === 'unauthenticated' && !inAuthGroup) {
      router.replace('/(auth)/login');
    } else if (status === 'authenticated' && inAuthGroup) {
      router.replace('/(tabs)');
    }
  }, [status, segments, router, isPasswordRecovery]);

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

export default function RootLayout() {
  const initialize = useAuthStore((s) => s.initialize);

  useEffect(() => {
    // Hydrate the session and subscribe to auth changes; clean up on unmount.
    const unsubscribe = initialize();
    return unsubscribe;
  }, [initialize]);

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <QueryClientProvider client={queryClient}>
        <SafeAreaProvider>
          <StatusBar style="light" />
          <AuthGate />
        </SafeAreaProvider>
      </QueryClientProvider>
    </GestureHandlerRootView>
  );
}
