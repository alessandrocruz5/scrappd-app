// A small cross-platform text-input dialog. React Native's Alert.prompt is
// iOS-only, so Books/Pages use this for create + rename instead.

import { useState } from 'react';
import { Modal, StyleSheet, Text, View } from 'react-native';

import { AppButton, AppTextField } from '@/components/ui';
import { colors, radius, spacing } from '@/theme/colors';

type NameDialogProps = {
  visible: boolean;
  title: string;
  label: string;
  placeholder?: string;
  initialValue?: string;
  confirmLabel?: string;
  loading?: boolean;
  onCancel: () => void;
  onSubmit: (value: string) => void;
};

export function NameDialog({
  visible,
  title,
  label,
  placeholder,
  initialValue = '',
  confirmLabel = 'Save',
  loading = false,
  onCancel,
  onSubmit,
}: NameDialogProps) {
  const [value, setValue] = useState(initialValue);

  // Reset the field each time the dialog opens, using the "adjust state during
  // render from previous props" pattern instead of an effect.
  const [wasVisible, setWasVisible] = useState(visible);
  if (visible !== wasVisible) {
    setWasVisible(visible);
    if (visible) setValue(initialValue);
  }

  const trimmed = value.trim();

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onCancel}
    >
      <View style={styles.backdrop}>
        <View style={styles.card}>
          <Text style={styles.title}>{title}</Text>
          <AppTextField
            label={label}
            value={value}
            onChangeText={setValue}
            placeholder={placeholder}
            autoFocus
            autoCapitalize="sentences"
          />
          <AppButton
            label={confirmLabel}
            onPress={() => onSubmit(trimmed)}
            loading={loading}
            disabled={trimmed.length === 0}
          />
          <AppButton label="Cancel" variant="text" onPress={onCancel} />
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.45)',
    justifyContent: 'center',
    padding: spacing.xl,
  },
  card: {
    backgroundColor: colors.background,
    borderRadius: radius.lg,
    padding: spacing.xl,
  },
  title: {
    fontSize: 18,
    fontWeight: '700',
    color: colors.textPrimary,
    marginBottom: spacing.lg,
  },
});
