import { Ionicons } from '@expo/vector-icons';
import type { ReactNode } from 'react';
import {
  ActivityIndicator,
  StyleSheet,
  Text,
  TextInput,
  type TextInputProps,
  TouchableOpacity,
  View,
} from 'react-native';

import { colors, radius, spacing } from '@/theme/colors';

type AppButtonProps = {
  label: string;
  onPress: () => void;
  loading?: boolean;
  disabled?: boolean;
  variant?: 'primary' | 'text';
};

export function AppButton({
  label,
  onPress,
  loading = false,
  disabled = false,
  variant = 'primary',
}: AppButtonProps) {
  const isPrimary = variant === 'primary';
  return (
    <TouchableOpacity
      onPress={onPress}
      disabled={disabled || loading}
      activeOpacity={0.85}
      style={[
        isPrimary ? styles.primaryButton : styles.textButton,
        (disabled || loading) && isPrimary ? styles.buttonDisabled : null,
      ]}
    >
      {loading ? (
        <ActivityIndicator color={isPrimary ? colors.white : colors.primary} />
      ) : (
        <Text style={isPrimary ? styles.primaryButtonText : styles.textButtonText}>
          {label}
        </Text>
      )}
    </TouchableOpacity>
  );
}

type AppTextFieldProps = TextInputProps & {
  label: string;
};

export function AppTextField({ label, style, ...rest }: AppTextFieldProps) {
  return (
    <View style={styles.fieldWrapper}>
      <Text style={styles.fieldLabel}>{label}</Text>
      <TextInput
        placeholderTextColor={colors.textHint}
        style={[styles.input, style]}
        {...rest}
      />
    </View>
  );
}

export function FormError({ message }: { message: string }) {
  return (
    <View style={styles.errorBox}>
      <Ionicons name="alert-circle-outline" size={18} color={colors.error} />
      <Text style={styles.errorText}>{message}</Text>
    </View>
  );
}

export function FormNotice({ message }: { message: string }) {
  return (
    <View style={styles.noticeBox}>
      <Ionicons name="information-circle-outline" size={18} color={colors.secondary} />
      <Text style={styles.noticeText}>{message}</Text>
    </View>
  );
}

export function Placeholder({
  icon,
  title,
  subtitle,
}: {
  icon: keyof typeof Ionicons.glyphMap;
  title: string;
  subtitle: string;
}): ReactNode {
  return (
    <View style={styles.placeholder}>
      <Ionicons name={icon} size={56} color={colors.accent} />
      <Text style={styles.placeholderTitle}>{title}</Text>
      <Text style={styles.placeholderSubtitle}>{subtitle}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  primaryButton: {
    backgroundColor: colors.primary,
    borderRadius: radius.md,
    paddingVertical: spacing.lg,
    alignItems: 'center',
    justifyContent: 'center',
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  primaryButtonText: {
    color: colors.white,
    fontSize: 16,
    fontWeight: '600',
  },
  textButton: {
    paddingVertical: spacing.md,
    alignItems: 'center',
  },
  textButtonText: {
    color: colors.primary,
    fontSize: 15,
    fontWeight: '500',
  },
  fieldWrapper: {
    marginBottom: spacing.lg,
  },
  fieldLabel: {
    color: colors.textSecondary,
    fontSize: 14,
    fontWeight: '600',
    marginBottom: spacing.sm,
  },
  input: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: radius.md,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    fontSize: 16,
    color: colors.textPrimary,
  },
  errorBox: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    backgroundColor: 'rgba(239, 68, 68, 0.1)',
    borderRadius: radius.sm,
    padding: spacing.md,
    marginBottom: spacing.lg,
  },
  errorText: {
    color: colors.error,
    flex: 1,
    fontSize: 14,
  },
  noticeBox: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    backgroundColor: 'rgba(81, 4, 32, 0.08)',
    borderRadius: radius.sm,
    padding: spacing.md,
    marginBottom: spacing.lg,
  },
  noticeText: {
    color: colors.secondary,
    flex: 1,
    fontSize: 14,
  },
  placeholder: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: spacing.xl,
    gap: spacing.md,
  },
  placeholderTitle: {
    fontSize: 22,
    fontWeight: '700',
    color: colors.textPrimary,
  },
  placeholderSubtitle: {
    fontSize: 15,
    color: colors.textSecondary,
    textAlign: 'center',
  },
});
