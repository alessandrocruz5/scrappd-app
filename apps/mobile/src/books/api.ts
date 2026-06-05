// Data access for Books, Pages, and the items that can be placed on a page.
//
// Everything lives in the `content` Postgres schema (see
// packages/supabase/migrations) and is guarded by RLS keyed on auth.uid(), so
// these queries never filter by user_id themselves — the database does it. This
// is the React Native port of the Flutter ProjectsProvider / repository layer,
// now talking to Supabase directly instead of the retired Go API.

import type { Database } from '@scrappd/shared-types';

import { supabase } from '@/lib/supabase';

type ContentTables = Database['content']['Tables'];

export type Book = ContentTables['books']['Row'];
export type Page = ContentTables['pages']['Row'];
export type Item = ContentTables['items']['Row'];
export type PageItem = ContentTables['page_items']['Row'];

const content = () => supabase.schema('content');

// Surface a usable message regardless of where the failure came from, mirroring
// the Flutter mapErrorMessage helper.
function fail(error: { message: string } | null, fallback: string): never {
  throw new Error(error?.message || fallback);
}

// ---------------------------------------------------------------------------
// Books
// ---------------------------------------------------------------------------

export async function listBooks(): Promise<Book[]> {
  const { data, error } = await content()
    .from('books')
    .select('*')
    .is('deleted_at', null)
    .order('created_at', { ascending: false });
  if (error) fail(error, 'Failed to load books.');
  return data ?? [];
}

export async function createBook(
  title: string,
  description?: string | null,
): Promise<Book> {
  const {
    data: { user },
    error: userError,
  } = await supabase.auth.getUser();
  if (userError || !user) {
    throw new Error('You need to be signed in to create a book.');
  }

  const { data, error } = await content()
    .from('books')
    .insert({
      user_id: user.id,
      title: title.trim(),
      description: description?.trim() || null,
    })
    .select('*')
    .single();
  if (error || !data) fail(error, 'Failed to create book.');
  return data;
}

export async function renameBook(id: string, title: string): Promise<Book> {
  const { data, error } = await content()
    .from('books')
    .update({ title: title.trim() })
    .eq('id', id)
    .select('*')
    .single();
  if (error || !data) fail(error, 'Failed to rename book.');
  return data;
}

export async function deleteBook(id: string): Promise<void> {
  // Hard delete; content.pages / content.page_items cascade (see schema).
  const { error } = await content().from('books').delete().eq('id', id);
  if (error) fail(error, 'Failed to delete book.');
}

// ---------------------------------------------------------------------------
// Pages
// ---------------------------------------------------------------------------

export async function listPages(bookId: string): Promise<Page[]> {
  const { data, error } = await content()
    .from('pages')
    .select('*')
    .eq('book_id', bookId)
    .order('page_number', { ascending: true });
  if (error) fail(error, 'Failed to load pages.');
  return data ?? [];
}

export async function createPage(bookId: string): Promise<Page> {
  // page_number is unique per book, so append after the current max.
  const { data: last, error: maxError } = await content()
    .from('pages')
    .select('page_number')
    .eq('book_id', bookId)
    .order('page_number', { ascending: false })
    .limit(1)
    .maybeSingle();
  if (maxError) fail(maxError, 'Failed to create page.');

  const nextNumber = (last?.page_number ?? 0) + 1;

  const { data, error } = await content()
    .from('pages')
    .insert({ book_id: bookId, page_number: nextNumber })
    .select('*')
    .single();
  if (error || !data) fail(error, 'Failed to create page.');
  return data;
}

export async function deletePage(id: string): Promise<void> {
  const { error } = await content().from('pages').delete().eq('id', id);
  if (error) fail(error, 'Failed to delete page.');
}

// Persist a new page order. page_number carries a unique(book_id, page_number)
// constraint, so a naive in-place renumber would collide mid-update. Move every
// row to a negative slot first (always free), then to its final 1-based index.
export async function reorderPages(orderedIds: string[]): Promise<void> {
  for (let i = 0; i < orderedIds.length; i += 1) {
    const { error } = await content()
      .from('pages')
      .update({ page_number: -(i + 1) })
      .eq('id', orderedIds[i]);
    if (error) fail(error, 'Failed to reorder pages.');
  }
  for (let i = 0; i < orderedIds.length; i += 1) {
    const { error } = await content()
      .from('pages')
      .update({ page_number: i + 1 })
      .eq('id', orderedIds[i]);
    if (error) fail(error, 'Failed to reorder pages.');
  }
}

// ---------------------------------------------------------------------------
// Items (cropper cutouts) + placing them on a page
// ---------------------------------------------------------------------------

export async function listItems(): Promise<Item[]> {
  const { data, error } = await content()
    .from('items')
    .select('*')
    .is('deleted_at', null)
    .order('created_at', { ascending: false });
  if (error) fail(error, 'Failed to load your cutouts.');
  return data ?? [];
}

export type PageItemWithItem = PageItem & { item: Item | null };

export async function listPageItems(
  pageId: string,
): Promise<PageItemWithItem[]> {
  // Embed the referenced item (via page_items_item_id_fkey) so callers can
  // render the cutout without a second round-trip.
  const { data, error } = await content()
    .from('page_items')
    .select('*, item:items(*)')
    .eq('page_id', pageId)
    .order('z_index', { ascending: true });
  if (error) fail(error, 'Failed to load page contents.');
  return (data as PageItemWithItem[] | null) ?? [];
}

// Drop a cropper item onto a page at a sensible default size/position. The full
// editor (next milestone) will let the user move and resize it.
export async function addItemToPage(
  pageId: string,
  itemId: string,
): Promise<PageItem> {
  const { data: top, error: topError } = await content()
    .from('page_items')
    .select('z_index')
    .eq('page_id', pageId)
    .order('z_index', { ascending: false })
    .limit(1)
    .maybeSingle();
  if (topError) fail(topError, 'Failed to add item to page.');

  const { data, error } = await content()
    .from('page_items')
    .insert({
      page_id: pageId,
      item_id: itemId,
      position_x: 50,
      position_y: 50,
      width: 200,
      height: 200,
      z_index: (top?.z_index ?? 0) + 1,
    })
    .select('*')
    .single();
  if (error || !data) fail(error, 'Failed to add item to page.');
  return data;
}
