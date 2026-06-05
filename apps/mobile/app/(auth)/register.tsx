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

export default function RegisterScreen() {
  const signUp = useAuthStore((s) => s.signUp);
  const isSubmitting = useAuthStore((s) => s.isSubmitting);

  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const onSubmit = async () => {
    setLocalError(null);
    setNotice(null);
    if (!email.trim() || password.length < 8) {
      setLocalError(
        'Enter your email and a password of at least 8 characters.',
      );
      return;
    }
    const result = await signUp(email, password, displayName || undefined);
    if (!result.ok) {
      setLocalError(result.message ?? 'Registration failed. Please try again.');
      return;
    }
    // When confirmation is required there is no session yet — surface the
    // notice and let the user head back to sign in once verified.
    if (result.message) {
      setNotice(result.message);
    }
    // Otherwise the AuthGate redirects into the tab shell automatically.
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
          <Text style={styles.title}>Create your account</Text>
          <Text style={styles.subtitle}>
            Start building scrapbooks in minutes.
          </Text>

          <View style={styles.form}>
            <AppTextField
              label="Display name (optional)"
              value={displayName}
              onChangeText={setDisplayName}
              autoCapitalize="words"
              textContentType="name"
            />
            <AppTextField
              label="Email"
              value={email}
              onChangeText={setEmail}
              autoCapitalize="none"
              keyboardType="email-address"
              autoComplete="email"
              textContentType="emailAddress"
            />
            <AppTextField
              label="Password"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
              autoComplete="password-new"
              textContentType="newPassword"
            />

            {notice ? <FormNotice message={notice} /> : null}
            {localError ? <FormError message={localError} /> : null}

            <AppButton
              label="Create account"
              onPress={onSubmit}
              loading={isSubmitting}
            />
          </View>

          <View style={styles.footer}>
            <Text style={styles.footerText}>Already have an account?</Text>
            <Link href="/(auth)/login" style={styles.link} replace>
              Log in
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
