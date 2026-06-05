import { Ionicons } from '@expo/vector-icons';
import { StyleSheet, Text, View } from 'react-native';

import { AppButton } from '@/components/ui';
import { useAuthStore } from '@/stores/auth-store';
import { colors, radius, spacing } from '@/theme/colors';

export default function ProfileScreen() {
  const user = useAuthStore((s) => s.user);
  const signOut = useAuthStore((s) => s.signOut);
  const isSubmitting = useAuthStore((s) => s.isSubmitting);

  const displayName =
    (user?.user_metadata?.display_name as string | undefined) ?? null;

  return (
    <View style={styles.container}>
      <View style={styles.card}>
        <View style={styles.avatar}>
          <Ionicons name="person" size={36} color={colors.white} />
        </View>
        {displayName ? <Text style={styles.name}>{displayName}</Text> : null}
        <Text style={styles.email}>{user?.email ?? 'Signed in'}</Text>
      </View>

      <View style={styles.actions}>
        <AppButton label="Log out" onPress={signOut} loading={isSubmitting} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
    padding: spacing.xl,
  },
  card: {
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
    padding: spacing.xl,
    alignItems: 'center',
    gap: spacing.sm,
  },
  avatar: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: spacing.sm,
  },
  name: {
    fontSize: 20,
    fontWeight: '700',
    color: colors.textPrimary,
  },
  email: {
    fontSize: 15,
    color: colors.textSecondary,
  },
  actions: {
    marginTop: spacing.xl,
  },
});
