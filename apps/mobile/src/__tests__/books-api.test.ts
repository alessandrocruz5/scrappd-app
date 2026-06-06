/* eslint-disable import/first -- jest.mock() must be hoisted above the imports it stubs. */
// Exercises the Books / Pages / Items data layer against the chainable Supabase
// mock: the happy path for each query, the auth guard on create, the
// error→Error mapping, and the trickier multi-statement flows (createPage's
// max+1 numbering and reorderPages' two-phase renumber).

import { createSupabaseMock } from '@/test-utils/supabase-mock';

const mockDb = createSupabaseMock();
// Lazy getter: the factory is hoisted above `mockDb`'s initialisation, so it
// must not dereference it until first access (which happens at call time).
jest.mock('@/lib/supabase', () => ({
  get supabase() {
    return mockDb.supabase;
  },
}));

import {
  addItemToPage,
  createBook,
  createPage,
  deleteBook,
  deletePageItem,
  listBooks,
  listPageItems,
  listPages,
  renameBook,
  reorderPages,
  updatePageItem,
} from '@/books/api';

function queue(...results: { data?: unknown; error?: unknown }[]) {
  mockDb.queue.push(...results);
}

function lastInsert() {
  return mockDb.calls.filter((c) => c.method === 'insert').pop()
    ?.args[0] as Record<string, unknown>;
}

beforeEach(() => {
  mockDb.queue.length = 0;
  mockDb.calls.length = 0;
  mockDb.getUser.mockResolvedValue({
    data: { user: { id: 'user-1' } },
    error: null,
  });
});

describe('books', () => {
  it('listBooks returns the rows on success', async () => {
    queue({ data: [{ id: 'b1' }, { id: 'b2' }], error: null });
    await expect(listBooks()).resolves.toEqual([{ id: 'b1' }, { id: 'b2' }]);
    expect(mockDb.calls).toContainEqual({ method: 'from', args: ['books'] });
  });

  it('listBooks coerces a null payload to an empty array', async () => {
    queue({ data: null, error: null });
    await expect(listBooks()).resolves.toEqual([]);
  });

  it('listBooks throws the mapped message on error', async () => {
    queue({ data: null, error: { message: 'boom' } });
    await expect(listBooks()).rejects.toThrow('boom');
  });

  it('createBook requires a signed-in user', async () => {
    mockDb.getUser.mockResolvedValue({ data: { user: null }, error: null });
    await expect(createBook('My Book')).rejects.toThrow(/need to be signed in/);
  });

  it('createBook trims the title and description before inserting', async () => {
    queue({ data: { id: 'b1', title: 'Trip' }, error: null });
    await createBook('  Trip  ', '  notes  ');
    expect(lastInsert()).toMatchObject({
      user_id: 'user-1',
      title: 'Trip',
      description: 'notes',
    });
  });

  it('createBook normalises an empty description to null', async () => {
    queue({ data: { id: 'b1' }, error: null });
    await createBook('Trip', '   ');
    expect(lastInsert()).toMatchObject({ description: null });
  });

  it('createBook surfaces the fallback when no row comes back', async () => {
    queue({ data: null, error: null });
    await expect(createBook('Trip')).rejects.toThrow('Failed to create book.');
  });

  it('renameBook trims and returns the updated row', async () => {
    queue({ data: { id: 'b1', title: 'New' }, error: null });
    await expect(renameBook('b1', '  New  ')).resolves.toEqual({
      id: 'b1',
      title: 'New',
    });
  });

  it('deleteBook resolves on success and throws on error', async () => {
    queue({ error: null });
    await expect(deleteBook('b1')).resolves.toBeUndefined();
    queue({ error: { message: 'cannot delete' } });
    await expect(deleteBook('b1')).rejects.toThrow('cannot delete');
  });
});

describe('pages', () => {
  it('listPages returns the ordered rows', async () => {
    queue({ data: [{ id: 'p1', page_number: 1 }], error: null });
    await expect(listPages('b1')).resolves.toEqual([
      { id: 'p1', page_number: 1 },
    ]);
  });

  it('createPage appends after the current max page_number', async () => {
    // First await: the max-page lookup. Second await: the insert.
    queue(
      { data: { page_number: 4 }, error: null },
      { data: { id: 'p5', page_number: 5 }, error: null },
    );
    const page = await createPage('b1');
    expect(page).toMatchObject({ page_number: 5 });
    expect(lastInsert()).toMatchObject({ book_id: 'b1', page_number: 5 });
  });

  it('createPage starts at page_number 1 for an empty book', async () => {
    queue(
      { data: null, error: null },
      { data: { id: 'p1', page_number: 1 }, error: null },
    );
    await createPage('b1');
    expect(lastInsert()).toMatchObject({ page_number: 1 });
  });

  it('reorderPages renumbers to negative slots first, then 1-based', async () => {
    const ids = ['a', 'b', 'c'];
    // 2N statements (negative pass + positive pass), all succeeding.
    queue(...Array.from({ length: ids.length * 2 }, () => ({ error: null })));
    await reorderPages(ids);

    const updates = mockDb.calls
      .filter((c) => c.method === 'update')
      .map((c) => (c.args[0] as { page_number: number }).page_number);
    // First three drop to negative free slots, last three land 1..3.
    expect(updates).toEqual([-1, -2, -3, 1, 2, 3]);
  });

  it('reorderPages throws if any statement fails', async () => {
    queue({ error: null }, { error: { message: 'conflict' } });
    await expect(reorderPages(['a', 'b'])).rejects.toThrow('conflict');
  });
});

describe('items + page placement', () => {
  it('listPageItems returns the embedded item rows', async () => {
    queue({
      data: [{ id: 'pi1', z_index: 1, item: { id: 'i1' } }],
      error: null,
    });
    await expect(listPageItems('p1')).resolves.toEqual([
      { id: 'pi1', z_index: 1, item: { id: 'i1' } },
    ]);
  });

  it('addItemToPage stacks one above the current top z_index', async () => {
    queue(
      { data: { z_index: 3 }, error: null },
      { data: { id: 'pi1', z_index: 4 }, error: null },
    );
    await addItemToPage('p1', 'i1');
    expect(lastInsert()).toMatchObject({
      page_id: 'p1',
      item_id: 'i1',
      position_x: 50,
      position_y: 50,
      width: 200,
      height: 200,
      z_index: 4,
    });
  });

  it('addItemToPage defaults z_index to 1 on an empty page', async () => {
    queue(
      { data: null, error: null },
      { data: { id: 'pi1', z_index: 1 }, error: null },
    );
    await addItemToPage('p1', 'i1');
    expect(lastInsert()).toMatchObject({ z_index: 1 });
  });

  it('updatePageItem forwards the transform patch and returns the row', async () => {
    queue({ data: { id: 'pi1', position_x: 10 }, error: null });
    await expect(
      updatePageItem('pi1', { position_x: 10, position_y: 20 }),
    ).resolves.toEqual({ id: 'pi1', position_x: 10 });
    const update = mockDb.calls.find((c) => c.method === 'update');
    expect(update?.args[0]).toEqual({ position_x: 10, position_y: 20 });
  });

  it('deletePageItem throws the mapped message on failure', async () => {
    queue({ error: { message: 'nope' } });
    await expect(deletePageItem('pi1')).rejects.toThrow('nope');
  });
});
