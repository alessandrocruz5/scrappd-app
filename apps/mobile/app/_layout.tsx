import { QueryClientProvider } from '@tanstack/react-query';
import { Slot, useRouter, useSegments } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { SplashScreen } from '@/components/splash-screen';
import { queryClient } from '@/lib/query-client';
import { useAuthStore } from '@/stores/auth-store';

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

  return <Slot />;
}

export default function RootLayout() {
  const initialize = useAuthStore((s) => s.initialize);

  useEffect(() => {
    // Hydrate the session and subscribe to auth changes; clean up on unmount.
    const unsubscribe = initialize();
    return unsubscribe;
  }, [initialize]);

  return (
    <QueryClientProvider client={queryClient}>
      <SafeAreaProvider>
        <StatusBar style="light" />
        <AuthGate />
      </SafeAreaProvider>
    </QueryClientProvider>
  );
}
