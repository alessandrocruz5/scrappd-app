// React Query hooks over the Books / Pages / Items data layer. Screens stay
// declarative: read with the query hooks, mutate with the mutation hooks, and
// let invalidation keep the lists fresh after a change.

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import {
  addItemToPage,
  createBook,
  createPage,
  deleteBook,
  deletePage,
  deletePageItem,
  getPage,
  listBooks,
  listItems,
  listPageItems,
  listPages,
  renameBook,
  reorderPages,
  updatePage,
  updatePageItem,
  type PageItemPatch,
  type PageItemWithItem,
  type PagePatch,
} from './api';

export const booksKeys = {
  all: ['books'] as const,
  pages: (bookId: string) => ['books', bookId, 'pages'] as const,
  page: (pageId: string) => ['pages', pageId] as const,
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

export function usePage(pageId: string) {
  return useQuery({
    queryKey: booksKeys.page(pageId),
    queryFn: () => getPage(pageId),
    enabled: !!pageId,
  });
}

// Background colour / pattern / template marker. Optimistic so the canvas
// repaints instantly when the user taps a template or swatch.
export function useUpdatePage(pageId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (patch: PagePatch) => updatePage(pageId, patch),
    onMutate: async (patch) => {
      await qc.cancelQueries({ queryKey: booksKeys.page(pageId) });
      const prev = qc.getQueryData(booksKeys.page(pageId));
      qc.setQueryData(booksKeys.page(pageId), (old: unknown) =>
        old ? { ...old, ...patch } : old,
      );
      return { prev };
    },
    onError: (_e, _patch, ctx) => {
      if (ctx?.prev) qc.setQueryData(booksKeys.page(pageId), ctx.prev);
    },
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

// Persist a transform (drag/resize/rotate) or a property change (opacity,
// z_index). Optimistically patches the cached list so the canvas — which reads
// straight from the query — stays the source of truth without a refetch round
// trip mid-interaction. The embedded `item` is preserved across the patch.
export function useUpdatePageItem(pageId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (vars: { id: string; patch: PageItemPatch }) =>
      updatePageItem(vars.id, vars.patch),
    onMutate: async (vars) => {
      await qc.cancelQueries({ queryKey: booksKeys.pageItems(pageId) });
      const prev = qc.getQueryData<PageItemWithItem[]>(
        booksKeys.pageItems(pageId),
      );
      qc.setQueryData<PageItemWithItem[]>(
        booksKeys.pageItems(pageId),
        (old) =>
          old
            ?.map((pi) => (pi.id === vars.id ? { ...pi, ...vars.patch } : pi))
            .sort((a, b) => a.z_index - b.z_index) ?? old,
      );
      return { prev };
    },
    onError: (_e, _vars, ctx) => {
      if (ctx?.prev) qc.setQueryData(booksKeys.pageItems(pageId), ctx.prev);
    },
  });
}

export function useRemovePageItem(pageId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deletePageItem(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: booksKeys.pageItems(pageId) });
      const prev = qc.getQueryData<PageItemWithItem[]>(
        booksKeys.pageItems(pageId),
      );
      qc.setQueryData<PageItemWithItem[]>(
        booksKeys.pageItems(pageId),
        (old) => old?.filter((pi) => pi.id !== id) ?? old,
      );
      return { prev };
    },
    onError: (_e, _id, ctx) => {
      if (ctx?.prev) qc.setQueryData(booksKeys.pageItems(pageId), ctx.prev);
    },
    onSettled: () =>
      qc.invalidateQueries({ queryKey: booksKeys.pageItems(pageId) }),
  });
}
