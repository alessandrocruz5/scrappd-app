// Horizontal strip of shape choices shown under the camera viewport.

import { Ionicons } from '@expo/vector-icons';
import { ScrollView, StyleSheet, Text, TouchableOpacity } from 'react-native';

import { colors, radius, spacing } from '@/theme/colors';

import { SHAPES, type ShapeId } from './shapes';

type ShapePickerProps = {
  selected: ShapeId;
  onSelect: (shape: ShapeId) => void;
  disabled?: boolean;
};

export function ShapePicker({ selected, onSelect, disabled }: ShapePickerProps) {
  return (
    <ScrollView
      horizontal
      showsHorizontalScrollIndicator={false}
      contentContainerStyle={styles.row}
    >
      {SHAPES.map((shape) => {
        const active = shape.id === selected;
        return (
          <TouchableOpacity
            key={shape.id}
            onPress={() => onSelect(shape.id)}
            disabled={disabled}
            activeOpacity={0.8}
            style={[styles.chip, active && styles.chipActive]}
          >
            <Ionicons
              // Icons come from a fixed Ionicons subset; cast keeps the glyph
              // map happy without widening ShapeMeta.icon to the full union.
              name={shape.icon as keyof typeof Ionicons.glyphMap}
              size={22}
              color={active ? colors.white : colors.textSecondary}
            />
            <Text style={[styles.label, active && styles.labelActive]}>
              {shape.label}
            </Text>
          </TouchableOpacity>
        );
      })}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  row: {
    gap: spacing.sm,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  chip: {
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.xs,
    minWidth: 72,
    paddingVertical: spacing.sm,
    paddingHorizontal: spacing.md,
    borderRadius: radius.md,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.surface,
  },
  chipActive: {
    backgroundColor: colors.primary,
    borderColor: colors.primary,
  },
  label: {
    fontSize: 12,
    fontWeight: '600',
    color: colors.textSecondary,
  },
  labelActive: {
    color: colors.white,
  },
});
