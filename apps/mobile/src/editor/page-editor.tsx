// The Skia page editor — the React Native port of page_editor_screen.dart.
//
// A page is a fixed-aspect canvas: a Skia-painted background (colour + pattern,
// chosen via templates) with cutouts placed on top. Each cutout can be dragged,
// pinched to resize, and twisted to rotate; the selected one exposes opacity,
// z-order, and delete. Every change is persisted to content.page_items /
// content.pages through Supabase, so the layout survives a reload.

import { Ionicons } from '@expo/vector-icons';
import { useMemo, useRef, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  type LayoutChangeEvent,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';

import {
  useAddItemToPage,
  usePage,
  usePageItems,
  useRemovePageItem,
  useUpdatePage,
  useUpdatePageItem,
} from '@/books/hooks';
import { ItemPicker } from '@/books/item-picker';
import { FormError } from '@/components/ui';
import { colors, radius, spacing } from '@/theme/colors';

import { CanvasItem } from './canvas-item';
import { exportPageView } from './export-page';
import { ItemToolbar } from './item-toolbar';
import { PageBackground } from './page-background';
import {
  BACKGROUND_SWATCHES,
  isPatternId,
  TEMPLATES,
  type PatternId,
} from './templates';

export function PageEditor({ pageId }: { pageId: string }) {
  const {
    data: page,
    isLoading: pageLoading,
    isError,
    error,
  } = usePage(pageId);
  const { data: pageItems, isLoading: itemsLoading } = usePageItems(pageId);

  const updatePage = useUpdatePage(pageId);
  const addItem = useAddItemToPage(pageId);
  const updateItem = useUpdatePageItem(pageId);
  const removeItem = useRemovePageItem(pageId);

  const [canvas, setCanvas] = useState({ width: 0, height: 0 });
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [exporting, setExporting] = useState(false);

  // The composed page view we snapshot on export (Skia background + cutouts).
  const pageRef = useRef<View>(null);

  const pattern: PatternId = isPatternId(page?.background_pattern)
    ? page.background_pattern
    : 'none';
  const backgroundColor = page?.background_color ?? '#FFFFFF';
  const aspect =
    page && page.canvas_width > 0 && page.canvas_height > 0
      ? page.canvas_width / page.canvas_height
      : 1080 / 1920;

  const selected = useMemo(
    () => pageItems?.find((pi) => pi.id === selectedId) ?? null,
    [pageItems, selectedId],
  );

  function onCanvasLayout(e: LayoutChangeEvent) {
    const { width, height } = e.nativeEvent.layout;
    setCanvas({ width, height });
  }

  function applyTemplate(templateId: string) {
    const template = TEMPLATES.find((t) => t.id === templateId);
    if (!template) return;
    updatePage.mutate({
      background_color: template.backgroundColor,
      background_pattern: template.pattern,
      layout_template: { template: template.id },
    });
  }

  function applySwatch(hex: string) {
    updatePage.mutate({ background_color: hex });
  }

  function handlePick(itemId: string) {
    addItem.mutate(itemId, {
      onSuccess: () => setPickerOpen(false),
      onError: (e) => Alert.alert('Couldn’t add cutout', String(e)),
    });
  }

  function bringToFront() {
    if (!selected || !pageItems) return;
    const max = Math.max(...pageItems.map((pi) => pi.z_index));
    if (selected.z_index === max) return;
    updateItem.mutate({ id: selected.id, patch: { z_index: max + 1 } });
  }

  function sendToBack() {
    if (!selected || !pageItems) return;
    const min = Math.min(...pageItems.map((pi) => pi.z_index));
    if (selected.z_index === min) return;
    updateItem.mutate({ id: selected.id, patch: { z_index: min - 1 } });
  }

  function changeOpacity(next: number) {
    if (!selected) return;
    updateItem.mutate({ id: selected.id, patch: { opacity: next } });
  }

  function deleteSelected() {
    if (!selected) return;
    const id = selected.id;
    setSelectedId(null);
    removeItem.mutate(id);
  }

  async function handleExport() {
    if (exporting || !page) return;
    // Drop the selection so its highlight border isn't baked into the PNG.
    setSelectedId(null);
    setExporting(true);
    try {
      // Wait a frame for the deselect (and any pending layout) to paint.
      await new Promise((resolve) =>
        requestAnimationFrame(() => resolve(null)),
      );
      // Export at the page's design resolution; fall back to a 2× snapshot of
      // the on-screen canvas when the stored dimensions are missing.
      const width =
        page.canvas_width > 0
          ? page.canvas_width
          : Math.round(canvas.width * 2);
      const height =
        page.canvas_height > 0
          ? page.canvas_height
          : Math.round(canvas.height * 2);

      const result = await exportPageView(pageRef, {
        width,
        height,
        fileName: `scrappd-page-${pageId}`,
      });

      const parts: string[] = [];
      if (result.savedToLibrary) parts.push('Saved to your photo library.');
      if (result.shared) parts.push('Opened the share sheet.');
      Alert.alert(
        'Page exported',
        parts.length > 0
          ? parts.join('\n')
          : 'Your high-res page PNG is ready.',
      );
    } catch (e) {
      Alert.alert(
        'Export failed',
        e instanceof Error ? e.message : 'Could not export this page.',
      );
    } finally {
      setExporting(false);
    }
  }

  if (pageLoading) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator color={colors.primary} size="large" />
      </View>
    );
  }

  if (isError || !page) {
    return (
      <View style={styles.noticeWrap}>
        <FormError
          message={
            error instanceof Error ? error.message : 'Failed to load page.'
          }
        />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {/* Export */}
      <View style={styles.topBar}>
        <TouchableOpacity
          style={styles.exportButton}
          activeOpacity={0.85}
          onPress={handleExport}
          disabled={exporting}
        >
          {exporting ? (
            <ActivityIndicator color={colors.primary} size="small" />
          ) : (
            <Ionicons name="share-outline" size={18} color={colors.primary} />
          )}
          <Text style={styles.exportButtonText}>
            {exporting ? 'Exporting…' : 'Export'}
          </Text>
        </TouchableOpacity>
      </View>

      {/* Templates */}
      <Text style={styles.sectionLabel}>Templates</Text>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.chipRow}
      >
        {TEMPLATES.map((t) => {
          const active =
            (page.layout_template as { template?: string } | null)?.template ===
            t.id;
          return (
            <TouchableOpacity
              key={t.id}
              style={[styles.chip, active ? styles.chipActive : null]}
              onPress={() => applyTemplate(t.id)}
              activeOpacity={0.8}
            >
              <Text
                style={[styles.chipText, active ? styles.chipTextActive : null]}
              >
                {t.name}
              </Text>
            </TouchableOpacity>
          );
        })}
      </ScrollView>

      {/* Background colour swatches */}
      <Text style={styles.sectionLabel}>Background</Text>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.chipRow}
      >
        {BACKGROUND_SWATCHES.map((hex) => (
          <TouchableOpacity
            key={hex}
            style={[
              styles.swatch,
              { backgroundColor: hex },
              backgroundColor.toUpperCase() === hex.toUpperCase()
                ? styles.swatchActive
                : null,
            ]}
            onPress={() => applySwatch(hex)}
            activeOpacity={0.8}
          />
        ))}
      </ScrollView>

      {/* Canvas */}
      <View style={styles.canvasWrap}>
        <View
          ref={pageRef}
          collapsable={false}
          style={[styles.canvas, { aspectRatio: aspect }]}
          onLayout={onCanvasLayout}
        >
          <PageBackground
            width={canvas.width}
            height={canvas.height}
            backgroundColor={backgroundColor}
            pattern={pattern}
            backgroundImageUrl={page.background_image_url}
          />

          {/* Tapping empty canvas clears the selection. */}
          <Pressable
            style={StyleSheet.absoluteFill}
            onPress={() => setSelectedId(null)}
          />

          {itemsLoading ? (
            <View style={styles.canvasLoader}>
              <ActivityIndicator color={colors.primary} />
            </View>
          ) : null}

          {canvas.width > 0 && pageItems
            ? pageItems.map((pi, index) => (
                <CanvasItem
                  key={pi.id}
                  pageItem={pi}
                  canvasWidth={canvas.width}
                  canvasHeight={canvas.height}
                  stackIndex={index + 1}
                  selected={pi.id === selectedId}
                  onSelect={() => setSelectedId(pi.id)}
                  onCommit={(patch) => updateItem.mutate({ id: pi.id, patch })}
                />
              ))
            : null}
        </View>
      </View>

      {selected ? (
        <ItemToolbar
          opacity={selected.opacity}
          onOpacityChange={changeOpacity}
          onBringToFront={bringToFront}
          onSendToBack={sendToBack}
          onDelete={deleteSelected}
        />
      ) : (
        <Text style={styles.hint}>
          Tap a cutout to select. Drag to move, pinch to resize, twist to
          rotate.
        </Text>
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
  topBar: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.md,
  },
  exportButton: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: radius.xl,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.primary,
  },
  exportButtonText: {
    fontSize: 14,
    fontWeight: '700',
    color: colors.primary,
  },
  noticeWrap: { padding: spacing.lg },
  sectionLabel: {
    fontSize: 13,
    fontWeight: '700',
    color: colors.textSecondary,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.md,
    paddingBottom: spacing.xs,
  },
  chipRow: {
    paddingHorizontal: spacing.lg,
    gap: spacing.sm,
    alignItems: 'center',
  },
  chip: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    borderRadius: radius.xl,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
  },
  chipActive: {
    backgroundColor: colors.primary,
    borderColor: colors.primary,
  },
  chipText: { fontSize: 14, fontWeight: '600', color: colors.textSecondary },
  chipTextActive: { color: colors.white },
  swatch: {
    width: 34,
    height: 34,
    borderRadius: radius.sm,
    borderWidth: 1,
    borderColor: colors.border,
  },
  swatchActive: {
    borderWidth: 3,
    borderColor: colors.primary,
  },
  canvasWrap: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: spacing.lg,
  },
  canvas: {
    width: '100%',
    maxHeight: '100%',
    borderRadius: radius.lg,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.surface,
  },
  canvasLoader: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    alignItems: 'center',
    justifyContent: 'center',
  },
  hint: {
    fontSize: 12,
    color: colors.textHint,
    textAlign: 'center',
    paddingHorizontal: spacing.lg,
    marginBottom: spacing.md,
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
