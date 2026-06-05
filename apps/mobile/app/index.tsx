import { Redirect } from 'expo-router';

import { useAuthStore } from '@/stores/auth-store';

// Entry route. Once the session is known, send authenticated users into the
// tab shell; everyone else falls through to the AuthGate, which redirects to
// the login stack.
export default function Index() {
  const status = useAuthStore((s) => s.status);

  if (status === 'authenticated') {
    return <Redirect href="/(tabs)" />;
  }

  return null;
}
