// A book opened to a grid of its pages. Create, reorder, and delete pages
// (content.pages keyed by book_id, under RLS); tapping a page opens the editor
// (stubbed for the next milestone).

import { Ionicons } from '@expo/vector-icons';
import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';

import type { Page } from '@/books/api';
import {
  useBooks,
  useCreatePage,
  useDeletePage,
  usePages,
  useReorderPages,
} from '@/books/hooks';
import { FormError, Placeholder } from '@/components/ui';
import { colors, radius, spacing } from '@/theme/colors';

export default function BookPagesScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const bookId = id ?? '';
  const router = useRouter();

  // The book title comes from the already-cached Books list, no extra fetch.
  const { data: books } = useBooks();
  const book = books?.find((b) => b.id === bookId);

  const {
    data: pages,
    isLoading,
    isError,
    error,
    refetch,
    isRefetching,
  } = usePages(bookId);

  const createPage = useCreatePage(bookId);
  const deletePage = useDeletePage(bookId);
  const reorderPages = useReorderPages(bookId);

  function addPage() {
    createPage.mutate(undefined, {
      onError: (e) => Alert.alert('Couldn’t add page', String(e)),
    });
  }

  function move(page: Page, direction: -1 | 1) {
    if (!pages) return;
    const index = pages.findIndex((p) => p.id === page.id);
    const target = index + direction;
    if (index < 0 || target < 0 || target >= pages.length) return;
    const next = [...pages];
    [next[index], next[target]] = [next[target], next[index]];
    reorderPages.mutate(
      next.map((p) => p.id),
      { onError: (e) => Alert.alert('Couldn’t reorder pages', String(e)) },
    );
  }

  function confirmDelete(page: Page) {
    Alert.alert('Delete page', `Delete page ${page.page_number}?`, [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Delete',
        style: 'destructive',
        onPress: () =>
          deletePage.mutate(page.id, {
            onError: (e) => Alert.alert('Couldn’t delete page', String(e)),
          }),
      },
    ]);
  }

  function pageOptions(page: Page) {
    Alert.alert(`Page ${page.page_number}`, undefined, [
      { text: 'Move up', onPress: () => move(page, -1) },
      { text: 'Move down', onPress: () => move(page, 1) },
      {
        text: 'Delete',
        style: 'destructive',
        onPress: () => confirmDelete(page),
      },
      { text: 'Cancel', style: 'cancel' },
    ]);
  }

  return (
    <View style={styles.container}>
      <Stack.Screen options={{ title: book?.title ?? 'Book' }} />

      {isError ? (
        <View style={styles.errorWrap}>
          <FormError
            message={
              error instanceof Error ? error.message : 'Failed to load pages.'
            }
          />
        </View>
      ) : null}

      {isLoading ? (
        <View style={styles.centered}>
          <ActivityIndicator color={colors.primary} size="large" />
        </View>
      ) : pages && pages.length > 0 ? (
        <FlatList
          data={pages}
          keyExtractor={(page) => page.id}
          numColumns={2}
          contentContainerStyle={styles.grid}
          columnWrapperStyle={styles.row}
          refreshControl={
            <RefreshControl refreshing={isRefetching} onRefresh={refetch} />
          }
          renderItem={({ item }) => (
            <TouchableOpacity
              style={styles.tile}
              activeOpacity={0.85}
              onPress={() => router.push(`/page/${item.id}`)}
              onLongPress={() => pageOptions(item)}
            >
              <View style={styles.tileCanvas}>
                <Ionicons
                  name="document-outline"
                  size={28}
                  color={colors.textHint}
                />
              </View>
              <View style={styles.tileFooter}>
                <Text style={styles.tileLabel}>Page {item.page_number}</Text>
                <TouchableOpacity
                  hitSlop={10}
                  onPress={() => pageOptions(item)}
                >
                  <Ionicons
                    name="ellipsis-horizontal"
                    size={18}
                    color={colors.textHint}
                  />
                </TouchableOpacity>
              </View>
            </TouchableOpacity>
          )}
        />
      ) : (
        <View style={styles.emptyWrap}>
          <Placeholder
            icon="documents-outline"
            title="No pages yet"
            subtitle="Add your first page to start building this book."
          />
        </View>
      )}

      <TouchableOpacity
        style={styles.fab}
        activeOpacity={0.85}
        onPress={addPage}
        disabled={createPage.isPending}
      >
        {createPage.isPending ? (
          <ActivityIndicator color={colors.white} />
        ) : (
          <Ionicons name="add" size={28} color={colors.white} />
        )}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  centered: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  errorWrap: { paddingHorizontal: spacing.lg, paddingTop: spacing.lg },
  emptyWrap: { flex: 1 },
  grid: { padding: spacing.md },
  row: { gap: spacing.md },
  tile: {
    flex: 1,
    margin: spacing.xs,
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    borderWidth: 1,
    borderColor: colors.border,
    overflow: 'hidden',
  },
  tileCanvas: {
    aspectRatio: 3 / 4,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
  },
  tileFooter: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
  },
  tileLabel: { fontSize: 14, fontWeight: '600', color: colors.textPrimary },
  fab: {
    position: 'absolute',
    right: spacing.xl,
    bottom: spacing.xl,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    elevation: 4,
    shadowColor: colors.black,
    shadowOpacity: 0.2,
    shadowRadius: 6,
    shadowOffset: { width: 0, height: 3 },
  },
});
