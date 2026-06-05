import { Ionicons } from '@expo/vector-icons';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';

import { colors, spacing } from '@/theme/colors';

// Shown while the auth store is resolving the persisted session (status
// 'unknown'). Mirrors the Flutter root_screen.dart splash.
export function SplashScreen() {
  return (
    <View style={styles.container}>
      <Ionicons name="sparkles" size={64} color={colors.white} />
      <Text style={styles.title}>Scrapp&apos;d</Text>
      <ActivityIndicator color={colors.white} style={styles.spinner} />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.secondary,
    gap: spacing.lg,
  },
  title: {
    fontSize: 40,
    fontWeight: '700',
    color: colors.white,
    letterSpacing: -0.5,
  },
  spinner: {
    marginTop: spacing.sm,
  },
});
