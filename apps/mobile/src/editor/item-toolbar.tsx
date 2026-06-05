// Controls for the currently selected cutout: opacity, z-order, and delete.
// Position/size/rotation are handled directly by gestures on the item; this
// covers the properties that don't map onto a drag.

import { Ionicons } from '@expo/vector-icons';
import { StyleSheet, Text, TouchableOpacity, View } from 'react-native';

import { colors, radius, spacing } from '@/theme/colors';

export function ItemToolbar({
  opacity,
  onOpacityChange,
  onBringToFront,
  onSendToBack,
  onDelete,
}: {
  opacity: number;
  onOpacityChange: (next: number) => void;
  onBringToFront: () => void;
  onSendToBack: () => void;
  onDelete: () => void;
}) {
  const stepOpacity = (delta: number) => {
    const next = Math.round((opacity + delta) * 10) / 10;
    onOpacityChange(Math.min(Math.max(next, 0.1), 1));
  };

  return (
    <View style={styles.bar}>
      <View style={styles.group}>
        <Text style={styles.label}>Opacity</Text>
        <TouchableOpacity
          style={styles.iconButton}
          onPress={() => stepOpacity(-0.1)}
          hitSlop={8}
        >
          <Ionicons name="remove" size={18} color={colors.textPrimary} />
        </TouchableOpacity>
        <Text style={styles.value}>{Math.round(opacity * 100)}%</Text>
        <TouchableOpacity
          style={styles.iconButton}
          onPress={() => stepOpacity(0.1)}
          hitSlop={8}
        >
          <Ionicons name="add" size={18} color={colors.textPrimary} />
        </TouchableOpacity>
      </View>

      <View style={styles.group}>
        <TouchableOpacity
          style={styles.iconButton}
          onPress={onSendToBack}
          hitSlop={8}
        >
          <Ionicons name="chevron-down" size={20} color={colors.textPrimary} />
        </TouchableOpacity>
        <TouchableOpacity
          style={styles.iconButton}
          onPress={onBringToFront}
          hitSlop={8}
        >
          <Ionicons name="chevron-up" size={20} color={colors.textPrimary} />
        </TouchableOpacity>
        <TouchableOpacity
          style={styles.iconButton}
          onPress={onDelete}
          hitSlop={8}
        >
          <Ionicons name="trash-outline" size={20} color={colors.error} />
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  bar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    marginHorizontal: spacing.lg,
    marginBottom: spacing.md,
    gap: spacing.md,
    elevation: 3,
    shadowColor: colors.black,
    shadowOpacity: 0.15,
    shadowRadius: 5,
    shadowOffset: { width: 0, height: 2 },
  },
  group: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  label: {
    fontSize: 13,
    fontWeight: '600',
    color: colors.textSecondary,
  },
  value: {
    fontSize: 13,
    fontWeight: '600',
    color: colors.textPrimary,
    minWidth: 38,
    textAlign: 'center',
  },
  iconButton: {
    width: 34,
    height: 34,
    borderRadius: radius.sm,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
});
