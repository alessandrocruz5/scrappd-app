/* eslint-disable import/first -- jest.mock() must be hoisted above the imports it stubs. */
// Integration tests for the optimistic React Query hooks. The data layer is
// mocked, so these focus on cache behaviour: that an in-flight edit patches the
// cached list immediately (and re-sorts by z_index), and that a failed edit
// rolls the cache back to its pre-mutation snapshot.

import type { ReactNode } from 'react';
import React from 'react';

import {
  notifyManager,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react-native';

// Run React Query notifications synchronously. By default they're batched onto
// a setTimeout, which can fire after a test finishes — leaving Jest with an
// open handle (and an out-of-act update warning). A synchronous scheduler keeps
// every cache notification inside the test that triggered it.
beforeAll(() => notifyManager.setScheduler((cb) => cb()));

jest.mock('@/books/api', () => ({
  addItemToPage: jest.fn(),
  createBook: jest.fn(),
  createPage: jest.fn(),
  deleteBook: jest.fn(),
  deletePage: jest.fn(),
  deletePageItem: jest.fn(),
  getPage: jest.fn(),
  listBooks: jest.fn(),
  listItems: jest.fn(),
  listPageItems: jest.fn(),
  listPages: jest.fn(),
  renameBook: jest.fn(),
  reorderPages: jest.fn(),
  updatePage: jest.fn(),
  updatePageItem: jest.fn(),
}));

import * as api from '@/books/api';
import {
  booksKeys,
  useRemovePageItem,
  useUpdatePage,
  useUpdatePageItem,
} from '@/books/hooks';

function makeWrapper() {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  createdClients.push(qc);
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  );
  return { qc, wrapper };
}

// Tear every client down after each test so no scheduler/timer survives it.
const createdClients: QueryClient[] = [];
afterEach(() => {
  for (const client of createdClients.splice(0)) client.clear();
});

// A promise whose resolution we control, to freeze a mutation mid-flight and
// inspect the optimistic cache before it settles.
function deferred<T>() {
  let resolve!: (v: T) => void;
  let reject!: (e: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

const PAGE = 'p1';

function pageItem(id: string, z: number) {
  return { id, z_index: z, item: { id: `i-${id}` } } as never;
}

beforeEach(() => jest.clearAllMocks());

describe('useUpdatePageItem (optimistic transform persistence)', () => {
  it('patches and re-sorts the cached list before the request settles', async () => {
    const { qc, wrapper } = makeWrapper();
    qc.setQueryData(booksKeys.pageItems(PAGE), [
      pageItem('a', 1),
      pageItem('b', 2),
    ]);

    const d = deferred<unknown>();
    (api.updatePageItem as jest.Mock).mockReturnValue(d.promise);

    const { result } = renderHook(() => useUpdatePageItem(PAGE), { wrapper });

    act(() => {
      result.current.mutate({ id: 'a', patch: { z_index: 5 } });
    });

    // Optimistically applied + re-sorted: 'b' (2) now before 'a' (5).
    await waitFor(() => {
      const list = qc.getQueryData<{ id: string; z_index: number }[]>(
        booksKeys.pageItems(PAGE),
      );
      expect(list?.map((p) => p.id)).toEqual(['b', 'a']);
      expect(list?.find((p) => p.id === 'a')?.z_index).toBe(5);
    });

    act(() => d.resolve({ id: 'a', z_index: 5 }));
    // Let the mutation settle inside the test so its final re-render doesn't
    // land after teardown (which Jest would flag as an open handle).
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it('preserves the embedded item across an optimistic patch', async () => {
    const { qc, wrapper } = makeWrapper();
    qc.setQueryData(booksKeys.pageItems(PAGE), [pageItem('a', 1)]);
    (api.updatePageItem as jest.Mock).mockResolvedValue({});

    const { result } = renderHook(() => useUpdatePageItem(PAGE), { wrapper });
    act(() => {
      result.current.mutate({ id: 'a', patch: { opacity: 0.5 } });
    });

    await waitFor(() => {
      const list = qc.getQueryData<
        { id: string; opacity: number; item: unknown }[]
      >(booksKeys.pageItems(PAGE));
      expect(list?.[0].item).toEqual({ id: 'i-a' });
      expect(list?.[0].opacity).toBe(0.5);
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it('rolls the cache back to its snapshot when the update fails', async () => {
    const { qc, wrapper } = makeWrapper();
    const original = [pageItem('a', 1), pageItem('b', 2)];
    qc.setQueryData(booksKeys.pageItems(PAGE), original);
    (api.updatePageItem as jest.Mock).mockRejectedValue(new Error('rls'));

    const { result } = renderHook(() => useUpdatePageItem(PAGE), { wrapper });
    act(() => {
      result.current.mutate({ id: 'a', patch: { z_index: 99 } });
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const list = qc.getQueryData<{ id: string; z_index: number }[]>(
      booksKeys.pageItems(PAGE),
    );
    expect(list?.map((p) => p.z_index)).toEqual([1, 2]);
  });
});

describe('useRemovePageItem (optimistic delete)', () => {
  it('drops the item from the cache immediately', async () => {
    const { qc, wrapper } = makeWrapper();
    qc.setQueryData(booksKeys.pageItems(PAGE), [
      pageItem('a', 1),
      pageItem('b', 2),
    ]);
    (api.deletePageItem as jest.Mock).mockResolvedValue(undefined);

    const { result } = renderHook(() => useRemovePageItem(PAGE), { wrapper });
    act(() => {
      result.current.mutate('a');
    });

    await waitFor(() => {
      const list = qc.getQueryData<{ id: string }[]>(booksKeys.pageItems(PAGE));
      expect(list?.map((p) => p.id)).toEqual(['b']);
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it('restores the removed item if the delete fails', async () => {
    const { qc, wrapper } = makeWrapper();
    qc.setQueryData(booksKeys.pageItems(PAGE), [pageItem('a', 1)]);
    (api.deletePageItem as jest.Mock).mockRejectedValue(new Error('boom'));

    const { result } = renderHook(() => useRemovePageItem(PAGE), { wrapper });
    act(() => {
      result.current.mutate('a');
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const list = qc.getQueryData<{ id: string }[]>(booksKeys.pageItems(PAGE));
    expect(list?.map((p) => p.id)).toEqual(['a']);
  });
});

describe('useUpdatePage (optimistic background)', () => {
  it('merges the patch into the cached page before settling', async () => {
    const { qc, wrapper } = makeWrapper();
    qc.setQueryData(booksKeys.page(PAGE), {
      id: PAGE,
      background_color: '#FFFFFF',
    });
    const d = deferred<unknown>();
    (api.updatePage as jest.Mock).mockReturnValue(d.promise);

    const { result } = renderHook(() => useUpdatePage(PAGE), { wrapper });
    act(() => {
      result.current.mutate({ background_color: '#000000' });
    });

    await waitFor(() => {
      const page = qc.getQueryData<{ background_color: string }>(
        booksKeys.page(PAGE),
      );
      expect(page?.background_color).toBe('#000000');
    });

    act(() => d.resolve({ id: PAGE, background_color: '#000000' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});
