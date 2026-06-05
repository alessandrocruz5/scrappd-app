// React Query hooks over the Books / Pages / Items data layer. Screens stay
// declarative: read with the query hooks, mutate with the mutation hooks, and
// let invalidation keep the lists fresh after a change.

import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';

import {
  addItemToPage,
  createBook,
  createPage,
  deleteBook,
  deletePage,
  listBooks,
  listItems,
  listPageItems,
  listPages,
  renameBook,
  reorderPages,
} from './api';

export const booksKeys = {
  all: ['books'] as const,
  pages: (bookId: string) => ['books', bookId, 'pages'] as const,
  pageItems: (pageId: string) => ['pages', pageId, 'items'] as const,
  items: ['items'] as const,
};

// ---------------------------------------------------------------------------
// Books
// ---------------------------------------------------------------------------

export function useBooks() {
  return useQuery({ queryKey: booksKeys.all, queryFn: listBooks });
}

export function useCreateBook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { title: string; description?: string | null }) =>
      createBook(vars.title, vars.description),
    onSuccess: () => qc.invalidateQueries({ queryKey: booksKeys.all }),
  });
}

export function useRenameBook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { id: string; title: string }) =>
      renameBook(vars.id, vars.title),
    onSuccess: () => qc.invalidateQueries({ queryKey: booksKeys.all }),
  });
}

export function useDeleteBook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteBook(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: booksKeys.all }),
  });
}

// ---------------------------------------------------------------------------
// Pages
// ---------------------------------------------------------------------------

export function usePages(bookId: string) {
  return useQuery({
    queryKey: booksKeys.pages(bookId),
    queryFn: () => listPages(bookId),
    enabled: !!bookId,
  });
}

export function useCreatePage(bookId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => createPage(bookId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: booksKeys.pages(bookId) }),
  });
}

export function useDeletePage(bookId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deletePage(id),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: booksKeys.pages(bookId) }),
  });
}

export function useReorderPages(bookId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderedIds: string[]) => reorderPages(orderedIds),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: booksKeys.pages(bookId) }),
  });
}

// ---------------------------------------------------------------------------
// Items
// ---------------------------------------------------------------------------

export function useItems() {
  return useQuery({ queryKey: booksKeys.items, queryFn: listItems });
}

export function usePageItems(pageId: string) {
  return useQuery({
    queryKey: booksKeys.pageItems(pageId),
    queryFn: () => listPageItems(pageId),
    enabled: !!pageId,
  });
}

export function useAddItemToPage(pageId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (itemId: string) => addItemToPage(pageId, itemId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: booksKeys.pageItems(pageId) }),
  });
}
