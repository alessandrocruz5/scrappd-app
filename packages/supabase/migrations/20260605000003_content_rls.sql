-- 20260605000003_content_rls.sql
-- ============================================================================
-- Row Level Security for the content schema.
--
-- This replaces every handler-level ownership check that used to live in the
-- Go API. The rules:
--
--   * Owner-only read/write on books, pages, items, page_items via auth.uid().
--   * Public READ where the owning book has visibility = 'public'.
--
-- Ownership for pages / page_items is derived through their book; items carry
-- user_id directly and are exposed publicly only when used on a public book.
-- ============================================================================

alter table content.books enable row level security;
alter table content.pages enable row level security;
alter table content.items enable row level security;
alter table content.page_items enable row level security;

-- ----------------------------------------------------------------------------
-- books
-- ----------------------------------------------------------------------------
create policy "Books are readable by owner or when public"
    on content.books for select
    using (
        auth.uid() = user_id
        or visibility = 'public'
    );

create policy "Users can insert their own books"
    on content.books for insert to authenticated
    with check (auth.uid() = user_id);

create policy "Users can update their own books"
    on content.books for update to authenticated
    using (auth.uid() = user_id)
    with check (auth.uid() = user_id);

create policy "Users can delete their own books"
    on content.books for delete to authenticated
    using (auth.uid() = user_id);

-- ----------------------------------------------------------------------------
-- pages  (ownership/visibility derived from the parent book)
-- ----------------------------------------------------------------------------
create policy "Pages are readable by book owner or when book is public"
    on content.pages for select
    using (
        exists (
            select 1 from content.books b
            where b.id = pages.book_id
              and (b.user_id = auth.uid() or b.visibility = 'public')
        )
    );

create policy "Users can insert pages into their own books"
    on content.pages for insert to authenticated
    with check (
        exists (
            select 1 from content.books b
            where b.id = pages.book_id and b.user_id = auth.uid()
        )
    );

create policy "Users can update pages in their own books"
    on content.pages for update to authenticated
    using (
        exists (
            select 1 from content.books b
            where b.id = pages.book_id and b.user_id = auth.uid()
        )
    )
    with check (
        exists (
            select 1 from content.books b
            where b.id = pages.book_id and b.user_id = auth.uid()
        )
    );

create policy "Users can delete pages in their own books"
    on content.pages for delete to authenticated
    using (
        exists (
            select 1 from content.books b
            where b.id = pages.book_id and b.user_id = auth.uid()
        )
    );

-- ----------------------------------------------------------------------------
-- items  (carry user_id directly; public read when used on a public book)
-- ----------------------------------------------------------------------------
create policy "Items are readable by owner or when used on a public book"
    on content.items for select
    using (
        auth.uid() = user_id
        or exists (
            select 1
            from content.page_items pi
            join content.pages pg on pg.id = pi.page_id
            join content.books b on b.id = pg.book_id
            where pi.item_id = items.id and b.visibility = 'public'
        )
    );

create policy "Users can insert their own items"
    on content.items for insert to authenticated
    with check (auth.uid() = user_id);

create policy "Users can update their own items"
    on content.items for update to authenticated
    using (auth.uid() = user_id)
    with check (auth.uid() = user_id);

create policy "Users can delete their own items"
    on content.items for delete to authenticated
    using (auth.uid() = user_id);

-- ----------------------------------------------------------------------------
-- page_items  (ownership/visibility derived from page -> book)
-- ----------------------------------------------------------------------------
create policy "Page items are readable by book owner or when book is public"
    on content.page_items for select
    using (
        exists (
            select 1
            from content.pages pg
            join content.books b on b.id = pg.book_id
            where pg.id = page_items.page_id
              and (b.user_id = auth.uid() or b.visibility = 'public')
        )
    );

create policy "Users can insert page items into their own books"
    on content.page_items for insert to authenticated
    with check (
        exists (
            select 1
            from content.pages pg
            join content.books b on b.id = pg.book_id
            where pg.id = page_items.page_id and b.user_id = auth.uid()
        )
    );

create policy "Users can update page items in their own books"
    on content.page_items for update to authenticated
    using (
        exists (
            select 1
            from content.pages pg
            join content.books b on b.id = pg.book_id
            where pg.id = page_items.page_id and b.user_id = auth.uid()
        )
    )
    with check (
        exists (
            select 1
            from content.pages pg
            join content.books b on b.id = pg.book_id
            where pg.id = page_items.page_id and b.user_id = auth.uid()
        )
    );

create policy "Users can delete page items in their own books"
    on content.page_items for delete to authenticated
    using (
        exists (
            select 1
            from content.pages pg
            join content.books b on b.id = pg.book_id
            where pg.id = page_items.page_id and b.user_id = auth.uid()
        )
    );

-- ----------------------------------------------------------------------------
-- usage_tracking  (strictly owner-only; no public access)
-- ----------------------------------------------------------------------------
alter table content.usage_tracking enable row level security;

create policy "Users can read their own usage"
    on content.usage_tracking for select
    using (auth.uid() = user_id);

create policy "Users can write their own usage"
    on content.usage_tracking for all to authenticated
    using (auth.uid() = user_id)
    with check (auth.uid() = user_id);
