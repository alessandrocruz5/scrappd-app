import { useState } from 'react';
import {
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import {
  AppButton,
  AppTextField,
  FormError,
  FormNotice,
} from '@/components/ui';
import { useAuthStore } from '@/stores/auth-store';
import { colors, spacing } from '@/theme/colors';

export default function ResetPasswordScreen() {
  const updatePassword = useAuthStore((s) => s.updatePassword);
  const signOut = useAuthStore((s) => s.signOut);
  const isSubmitting = useAuthStore((s) => s.isSubmitting);
  const isPasswordRecovery = useAuthStore((s) => s.isPasswordRecovery);
  // A bad/expired link fails to establish a recovery session and parks the
  // reason here, so surface it instead of an empty form the user can't use.
  const linkError = useAuthStore((s) => s.errorMessage);

  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const onSubmit = async () => {
    setLocalError(null);
    if (password.length < 8) {
      setLocalError('Choose a password of at least 8 characters.');
      return;
    }
    if (password !== confirm) {
      setLocalError('The passwords do not match.');
      return;
    }
    const result = await updatePassword(password);
    if (!result.ok) {
      setLocalError(result.message ?? 'Could not update your password.');
      return;
    }
    // On success isPasswordRecovery clears and the AuthGate routes the now
    // fully-authenticated user into the app.
  };

  if (!isPasswordRecovery) {
    return (
      <SafeAreaView style={styles.safe}>
        <View style={styles.content}>
          <Text style={styles.title}>Reset link needed</Text>
          <Text style={styles.subtitle}>
            Open this screen from the password-reset email link to continue.
          </Text>
          {linkError ? (
            <View style={styles.form}>
              <FormError message={linkError} />
            </View>
          ) : null}
          <View style={styles.form}>
            <AppButton label="Back to log in" onPress={signOut} />
          </View>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.safe}>
      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <ScrollView
          contentContainerStyle={styles.content}
          keyboardShouldPersistTaps="handled"
        >
          <Text style={styles.title}>Choose a new password</Text>
          <Text style={styles.subtitle}>
            Enter a new password for your account.
          </Text>

          <View style={styles.form}>
            <AppTextField
              label="New password"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
              autoComplete="password-new"
              textContentType="newPassword"
            />
            <AppTextField
              label="Confirm new password"
              value={confirm}
              onChangeText={setConfirm}
              secureTextEntry
              autoComplete="password-new"
              textContentType="newPassword"
            />

            <FormNotice message="After saving, you'll be signed in with your new password." />
            {localError ? <FormError message={localError} /> : null}

            <AppButton
              label="Save new password"
              onPress={onSubmit}
              loading={isSubmitting}
            />
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  flex: { flex: 1 },
  content: { flexGrow: 1, padding: spacing.xl },
  title: {
    fontSize: 32,
    fontWeight: '700',
    color: colors.textPrimary,
    marginTop: spacing.xxl,
  },
  subtitle: {
    fontSize: 16,
    color: colors.textSecondary,
    marginTop: spacing.sm,
  },
  form: { marginTop: spacing.xxl },
});
