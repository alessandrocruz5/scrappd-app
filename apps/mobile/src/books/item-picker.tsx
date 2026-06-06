// Modal that lists the cutouts the cropper has saved (content.items) so the
// user can drop one onto a page. This is the wiring between the cropper output
// and the page editor: pick an item -> a page_items row is created.

import { Ionicons } from '@expo/vector-icons';
import {
  ActivityIndicator,
  FlatList,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';

import { Placeholder } from '@/components/ui';
import { colors, spacing } from '@/theme/colors';

import { CutoutThumb } from './cutout-thumb';
import { useItems } from './hooks';

export function ItemPicker({
  visible,
  onClose,
  onPick,
}: {
  visible: boolean;
  onClose: () => void;
  onPick: (itemId: string) => void;
}) {
  const { data: items, isLoading } = useItems();

  return (
    <Modal
      visible={visible}
      animationType="slide"
      presentationStyle="pageSheet"
      onRequestClose={onClose}
    >
      <View style={styles.container}>
        <View style={styles.header}>
          <Text style={styles.title}>Add a cutout</Text>
          <TouchableOpacity onPress={onClose} hitSlop={12}>
            <Ionicons name="close" size={26} color={colors.textPrimary} />
          </TouchableOpacity>
        </View>

        {isLoading ? (
          <ActivityIndicator
            color={colors.primary}
            style={styles.loader}
            size="large"
          />
        ) : !items || items.length === 0 ? (
          <Placeholder
            icon="cut-outline"
            title="No cutouts yet"
            subtitle="Use the Cropper tab to snap a cutout, then add it here."
          />
        ) : (
          <FlatList
            data={items}
            keyExtractor={(item) => item.id}
            numColumns={3}
            contentContainerStyle={styles.grid}
            columnWrapperStyle={styles.row}
            renderItem={({ item }) => (
              <TouchableOpacity
                style={styles.tile}
                activeOpacity={0.8}
                onPress={() => onPick(item.id)}
              >
                <CutoutThumb
                  storageKey={
                    item.processed_image_key ?? item.original_image_key
                  }
                  style={styles.thumb}
                />
                {item.item_name ? (
                  <Text style={styles.tileLabel} numberOfLines={1}>
                    {item.item_name}
                  </Text>
                ) : null}
              </TouchableOpacity>
            )}
          />
        )}
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: spacing.lg,
  },
  title: { fontSize: 20, fontWeight: '700', color: colors.textPrimary },
  loader: { marginTop: spacing.xxl },
  grid: { padding: spacing.sm },
  row: { gap: spacing.sm },
  tile: {
    flex: 1 / 3,
    margin: spacing.xs,
  },
  thumb: {
    aspectRatio: 1,
    width: '100%',
  },
  tileLabel: {
    marginTop: spacing.xs,
    fontSize: 12,
    color: colors.textSecondary,
    textAlign: 'center',
  },
});
