import { Link } from 'expo-router';
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

export default function ForgotPasswordScreen() {
  const sendPasswordReset = useAuthStore((s) => s.sendPasswordReset);
  const isSubmitting = useAuthStore((s) => s.isSubmitting);

  const [email, setEmail] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const onSubmit = async () => {
    setLocalError(null);
    setNotice(null);
    if (!email.trim()) {
      setLocalError('Enter the email address for your account.');
      return;
    }
    const result = await sendPasswordReset(email);
    if (!result.ok) {
      setLocalError(result.message ?? 'Could not send reset email.');
      return;
    }
    setNotice(result.message ?? 'Check your email for a reset link.');
  };

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
          <Text style={styles.title}>Reset your password</Text>
          <Text style={styles.subtitle}>
            Enter your email and we&apos;ll send you a link to set a new
            password.
          </Text>

          <View style={styles.form}>
            <AppTextField
              label="Email"
              value={email}
              onChangeText={setEmail}
              autoCapitalize="none"
              keyboardType="email-address"
              autoComplete="email"
              textContentType="emailAddress"
            />

            {notice ? <FormNotice message={notice} /> : null}
            {localError ? <FormError message={localError} /> : null}

            <AppButton
              label="Send reset link"
              onPress={onSubmit}
              loading={isSubmitting}
            />
          </View>

          <View style={styles.footer}>
            <Text style={styles.footerText}>Remembered it?</Text>
            <Link href="/(auth)/login" style={styles.link} replace>
              Back to log in
            </Link>
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
  footer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    gap: spacing.xs,
    marginTop: spacing.xl,
  },
  footerText: { color: colors.textSecondary, fontSize: 15 },
  link: { color: colors.primary, fontSize: 15, fontWeight: '600' },
});
