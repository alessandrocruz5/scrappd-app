// Books tab: the user's scrapbooks. List, create, rename, and delete books
// (content.books under RLS), and open one to its pages. This is the React
// Native port of the Flutter ProjectsProvider / projects list screen.

import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useState } from 'react';
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

import { NameDialog } from '@/books/name-dialog';
import type { Book } from '@/books/api';
import {
  useBooks,
  useCreateBook,
  useDeleteBook,
  useRenameBook,
} from '@/books/hooks';
import { FormError, Placeholder } from '@/components/ui';
import { colors, radius, spacing } from '@/theme/colors';

type DialogState =
  | { mode: 'closed' }
  | { mode: 'create' }
  | { mode: 'rename'; book: Book };

export default function BooksScreen() {
  const router = useRouter();
  const {
    data: books,
    isLoading,
    isError,
    error,
    refetch,
    isRefetching,
  } = useBooks();

  const createBook = useCreateBook();
  const renameBook = useRenameBook();
  const deleteBook = useDeleteBook();

  const [dialog, setDialog] = useState<DialogState>({ mode: 'closed' });

  const closeDialog = () => setDialog({ mode: 'closed' });

  function submitDialog(value: string) {
    if (dialog.mode === 'create') {
      createBook.mutate(
        { title: value },
        {
          onSuccess: closeDialog,
          onError: (e) => Alert.alert('Couldn’t create book', String(e)),
        },
      );
    } else if (dialog.mode === 'rename') {
      renameBook.mutate(
        { id: dialog.book.id, title: value },
        {
          onSuccess: closeDialog,
          onError: (e) => Alert.alert('Couldn’t rename book', String(e)),
        },
      );
    }
  }

  function confirmDelete(book: Book) {
    Alert.alert(
      'Delete book',
      `Delete “${book.title}” and all of its pages? This can’t be undone.`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: () =>
            deleteBook.mutate(book.id, {
              onError: (e) => Alert.alert('Couldn’t delete book', String(e)),
            }),
        },
      ],
    );
  }

  function bookOptions(book: Book) {
    Alert.alert(book.title, undefined, [
      { text: 'Rename', onPress: () => setDialog({ mode: 'rename', book }) },
      {
        text: 'Delete',
        style: 'destructive',
        onPress: () => confirmDelete(book),
      },
      { text: 'Cancel', style: 'cancel' },
    ]);
  }

  if (isLoading) {
    return (
      <View style={[styles.container, styles.centered]}>
        <ActivityIndicator color={colors.primary} size="large" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {isError ? (
        <View style={styles.errorWrap}>
          <FormError
            message={
              error instanceof Error ? error.message : 'Failed to load books.'
            }
          />
        </View>
      ) : null}

      {books && books.length > 0 ? (
        <FlatList
          data={books}
          keyExtractor={(book) => book.id}
          contentContainerStyle={styles.list}
          refreshControl={
            <RefreshControl refreshing={isRefetching} onRefresh={refetch} />
          }
          renderItem={({ item }) => (
            <TouchableOpacity
              style={styles.card}
              activeOpacity={0.85}
              onPress={() => router.push(`/book/${item.id}`)}
              onLongPress={() => bookOptions(item)}
            >
              <View style={styles.cardIcon}>
                <Ionicons name="book" size={22} color={colors.white} />
              </View>
              <View style={styles.cardBody}>
                <Text style={styles.cardTitle} numberOfLines={1}>
                  {item.title}
                </Text>
                {item.description ? (
                  <Text style={styles.cardSubtitle} numberOfLines={1}>
                    {item.description}
                  </Text>
                ) : null}
              </View>
              <TouchableOpacity hitSlop={12} onPress={() => bookOptions(item)}>
                <Ionicons
                  name="ellipsis-vertical"
                  size={20}
                  color={colors.textHint}
                />
              </TouchableOpacity>
            </TouchableOpacity>
          )}
        />
      ) : (
        <View style={styles.emptyWrap}>
          <Placeholder
            icon="book-outline"
            title="Your Books"
            subtitle="Create your first scrapbook to start adding pages."
          />
        </View>
      )}

      <TouchableOpacity
        style={styles.fab}
        activeOpacity={0.85}
        onPress={() => setDialog({ mode: 'create' })}
      >
        <Ionicons name="add" size={28} color={colors.white} />
      </TouchableOpacity>

      <NameDialog
        visible={dialog.mode !== 'closed'}
        title={dialog.mode === 'rename' ? 'Rename book' : 'New book'}
        label="Title"
        placeholder="e.g. Summer 2026"
        confirmLabel={dialog.mode === 'rename' ? 'Rename' : 'Create'}
        initialValue={dialog.mode === 'rename' ? dialog.book.title : ''}
        loading={createBook.isPending || renameBook.isPending}
        onCancel={closeDialog}
        onSubmit={submitDialog}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  centered: { alignItems: 'center', justifyContent: 'center' },
  errorWrap: { paddingHorizontal: spacing.lg, paddingTop: spacing.lg },
  emptyWrap: { flex: 1 },
  list: { padding: spacing.lg, gap: spacing.md },
  card: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    borderWidth: 1,
    borderColor: colors.border,
    padding: spacing.lg,
  },
  cardIcon: {
    width: 44,
    height: 44,
    borderRadius: radius.sm,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  cardBody: { flex: 1 },
  cardTitle: { fontSize: 16, fontWeight: '700', color: colors.textPrimary },
  cardSubtitle: {
    fontSize: 13,
    color: colors.textSecondary,
    marginTop: 2,
  },
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
