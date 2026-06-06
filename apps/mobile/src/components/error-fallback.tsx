import { Ionicons } from '@expo/vector-icons';
import { StyleSheet, Text, View } from 'react-native';

import { AppButton } from '@/components/ui';
import { colors, spacing } from '@/theme/colors';

// Friendly fallback shown by the root Sentry error boundary when an
// otherwise-unhandled render error crashes the tree. The exception is already
// reported to Sentry by the boundary; here we just keep the user out of a blank
// white screen and offer a one-tap recovery via the boundary's `resetError`.
export function ErrorFallback({ resetError }: { resetError: () => void }) {
  return (
    <View style={styles.container}>
      <Ionicons name="sad-outline" size={64} color={colors.white} />
      <Text style={styles.title}>Something went wrong</Text>
      <Text style={styles.subtitle}>
        Scrapp&apos;d hit an unexpected error. We&apos;ve been notified — try
        again, and if it keeps happening, restart the app.
      </Text>
      <View style={styles.action}>
        <AppButton label="Try again" onPress={resetError} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.secondary,
    padding: spacing.xl,
    gap: spacing.md,
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
    color: colors.white,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 15,
    color: colors.white,
    textAlign: 'center',
    opacity: 0.85,
    lineHeight: 22,
  },
  action: {
    alignSelf: 'stretch',
    marginTop: spacing.lg,
  },
});
