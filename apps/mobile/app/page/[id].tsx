// Page editor — stubbed for the next milestone. For now it shows the cutouts
// placed on this page and lets the user add more from the cropper's library,
// which is the wiring this milestone is responsible for: pick an item ->
// content.page_items row created under RLS.

import { Ionicons } from '@expo/vector-icons';
import { Stack, useLocalSearchParams } from 'expo-router';
import { useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';

import { CutoutThumb } from '@/books/cutout-thumb';
import { useAddItemToPage, usePageItems } from '@/books/hooks';
import { ItemPicker } from '@/books/item-picker';
import { FormError, FormNotice, Placeholder } from '@/components/ui';
import { colors, radius, spacing } from '@/theme/colors';

export default function PageEditorScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const pageId = id ?? '';

  const { data: pageItems, isLoading, isError, error } = usePageItems(pageId);
  const addItem = useAddItemToPage(pageId);

  const [pickerOpen, setPickerOpen] = useState(false);

  function handlePick(itemId: string) {
    addItem.mutate(itemId, {
      onSuccess: () => setPickerOpen(false),
      onError: (e) => Alert.alert('Couldn’t add cutout', String(e)),
    });
  }

  return (
    <View style={styles.container}>
      <Stack.Screen options={{ title: 'Page' }} />

      <View style={styles.noticeWrap}>
        <FormNotice message="Full page editor is coming soon. For now, add cutouts from your library." />
      </View>

      {isError ? (
        <View style={styles.noticeWrap}>
          <FormError
            message={error instanceof Error ? error.message : 'Failed to load page.'}
          />
        </View>
      ) : null}

      {isLoading ? (
        <View style={styles.centered}>
          <ActivityIndicator color={colors.primary} size="large" />
        </View>
      ) : pageItems && pageItems.length > 0 ? (
        <FlatList
          data={pageItems}
          keyExtractor={(pi) => pi.id}
          numColumns={3}
          contentContainerStyle={styles.grid}
          columnWrapperStyle={styles.row}
          renderItem={({ item: pi }) =>
            pi.item ? (
              <CutoutThumb
                storageKey={pi.item.processed_image_key ?? pi.item.original_image_key}
                style={styles.thumb}
              />
            ) : (
              <View style={[styles.thumb, styles.missing]}>
                <Ionicons name="help-outline" size={22} color={colors.textHint} />
              </View>
            )
          }
        />
      ) : (
        <View style={styles.emptyWrap}>
          <Placeholder
            icon="images-outline"
            title="Empty page"
            subtitle="Tap “Add cutout” to place your first item on this page."
          />
        </View>
      )}

      <TouchableOpacity
        style={styles.addButton}
        activeOpacity={0.85}
        onPress={() => setPickerOpen(true)}
        disabled={addItem.isPending}
      >
        {addItem.isPending ? (
          <ActivityIndicator color={colors.white} />
        ) : (
          <>
            <Ionicons name="add" size={20} color={colors.white} />
            <Text style={styles.addButtonText}>Add cutout</Text>
          </>
        )}
      </TouchableOpacity>

      <ItemPicker
        visible={pickerOpen}
        onClose={() => setPickerOpen(false)}
        onPick={handlePick}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  centered: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  noticeWrap: { paddingHorizontal: spacing.lg, paddingTop: spacing.lg },
  emptyWrap: { flex: 1 },
  grid: { padding: spacing.sm },
  row: { gap: spacing.sm },
  thumb: {
    flex: 1 / 3,
    aspectRatio: 1,
    margin: spacing.xs,
  },
  missing: {
    backgroundColor: colors.surface,
    borderRadius: radius.sm,
    alignItems: 'center',
    justifyContent: 'center',
  },
  addButton: {
    position: 'absolute',
    right: spacing.xl,
    bottom: spacing.xl,
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    backgroundColor: colors.primary,
    borderRadius: radius.xl,
    paddingHorizontal: spacing.xl,
    paddingVertical: spacing.md,
    elevation: 4,
    shadowColor: colors.black,
    shadowOpacity: 0.2,
    shadowRadius: 6,
    shadowOffset: { width: 0, height: 3 },
  },
  addButtonText: { color: colors.white, fontSize: 15, fontWeight: '600' },
});
