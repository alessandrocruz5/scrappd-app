-- 20260605000001_content_schema.sql
-- ============================================================================
-- Content schema: Books -> Pages -> Items (ported from the Go Postgres schema
-- in backend/migrations/000002 + 000003, with the Project -> Book rename).
--
-- Notes vs. the Go schema:
--   * content.projects        -> content.books
--   * content.pages.project_id -> content.pages.book_id
--   * The bespoke auth.users table from the Go schema is dropped entirely;
--     Supabase Auth owns auth.users now. FKs reference auth.users(id).
--   * ML/processing columns on items are retained but unused in the default
--     flow; items now default processing_status = 'completed'.
--   * usage_tracking is kept as-is (retained for the future premium tier).
-- ============================================================================

create schema if not exists content;

-- ----------------------------------------------------------------------------
-- Shared updated_at trigger function
-- ----------------------------------------------------------------------------
create or replace function content.set_updated_at()
returns trigger
language plpgsql
as $$
begin
    new.updated_at = now();
    return new;
end;
$$;

-- ----------------------------------------------------------------------------
-- content.books  (was content.projects)
-- ----------------------------------------------------------------------------
create table content.books (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references auth.users(id) on delete cascade,
    title varchar(200) not null,
    description text,
    cover_image_url text,

    -- Privacy & sharing
    visibility varchar(20) not null default 'private'
        check (visibility in ('private', 'unlisted', 'public')),
    is_template boolean not null default false,
    template_price decimal(10, 2),

    -- Stats
    view_count integer not null default 0,
    like_count integer not null default 0,
    fork_count integer not null default 0,

    -- Metadata
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    published_at timestamptz,
    deleted_at timestamptz
);

create index idx_books_user_id on content.books(user_id);
create index idx_books_visibility on content.books(visibility);
create index idx_books_is_template on content.books(is_template) where is_template = true;
create index idx_books_created_at on content.books(created_at desc);
create index idx_books_deleted_at on content.books(deleted_at) where deleted_at is null;

create trigger update_books_updated_at before update on content.books
    for each row execute function content.set_updated_at();

-- ----------------------------------------------------------------------------
-- content.pages  (project_id -> book_id)
-- ----------------------------------------------------------------------------
create table content.pages (
    id uuid primary key default gen_random_uuid(),
    book_id uuid not null references content.books(id) on delete cascade,
    page_number integer not null,
    title varchar(200),

    -- Canvas configuration
    canvas_width integer not null default 1080,
    canvas_height integer not null default 1920,
    background_color varchar(7) not null default '#FFFFFF',
    background_image_url text,
    background_pattern varchar(50),

    -- Template data
    layout_template jsonb,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    unique (book_id, page_number)
);

create index idx_pages_book_id on content.pages(book_id);

create trigger update_pages_updated_at before update on content.pages
    for each row execute function content.set_updated_at();

-- ----------------------------------------------------------------------------
-- content.items  (ML columns retained but unused; default completed)
-- ----------------------------------------------------------------------------
create table content.items (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references auth.users(id) on delete cascade,

    -- Original image
    original_image_key varchar(500) not null,
    original_image_url text not null,
    original_file_size_bytes bigint,
    original_width integer,
    original_height integer,

    -- Processed image
    processed_image_key varchar(500),
    processed_image_url text,
    processed_file_size_bytes bigint,

    -- ML processing (retained from Go schema, unused in the default cropper
    -- flow). Cutouts are uploaded ready-to-use, so default to 'completed'.
    processing_status varchar(20) not null default 'completed'
        check (processing_status in ('pending', 'processing', 'completed', 'failed')),
    ml_model_version varchar(50),
    processing_started_at timestamptz,
    processing_completed_at timestamptz,
    processing_error text,

    -- Metadata
    mime_type varchar(50),
    item_name varchar(200),
    item_category varchar(50),
    tags text[],

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create index idx_items_user_id on content.items(user_id);
create index idx_items_processing_status on content.items(processing_status);
create index idx_items_created_at on content.items(created_at desc);
create index idx_items_deleted_at on content.items(deleted_at) where deleted_at is null;

create trigger update_items_updated_at before update on content.items
    for each row execute function content.set_updated_at();

-- ----------------------------------------------------------------------------
-- content.page_items
-- ----------------------------------------------------------------------------
create table content.page_items (
    id uuid primary key default gen_random_uuid(),
    page_id uuid not null references content.pages(id) on delete cascade,
    item_id uuid not null references content.items(id) on delete cascade,

    -- Position & transformation
    position_x decimal(10, 2) not null,
    position_y decimal(10, 2) not null,
    width decimal(10, 2) not null,
    height decimal(10, 2) not null,
    rotation decimal(6, 2) not null default 0,
    z_index integer not null default 0,
    opacity decimal(3, 2) not null default 1.0,

    -- Filters/effects
    filters jsonb,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_page_items_page_id on content.page_items(page_id);
create index idx_page_items_item_id on content.page_items(item_id);
create index idx_page_items_z_index on content.page_items(z_index);

create trigger update_page_items_updated_at before update on content.page_items
    for each row execute function content.set_updated_at();

-- ----------------------------------------------------------------------------
-- content.usage_tracking  (kept for the future premium/freemium limits)
-- ----------------------------------------------------------------------------
create table content.usage_tracking (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references auth.users(id) on delete cascade,

    -- Usage period (monthly)
    period_start date not null,
    period_end date not null,

    -- Counters
    items_processed integer not null default 0,
    items_limit integer, -- NULL means unlimited (Pro users)

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    unique (user_id, period_start)
);

create index idx_usage_tracking_user_id on content.usage_tracking(user_id);
create index idx_usage_tracking_period on content.usage_tracking(period_start, period_end);
create index idx_usage_tracking_user_current on content.usage_tracking(user_id, period_start, period_end);

create trigger update_usage_tracking_updated_at before update on content.usage_tracking
    for each row execute function content.set_updated_at();

-- ----------------------------------------------------------------------------
-- Expose the content schema to the PostgREST API roles. RLS (added in a later
-- migration) still governs row visibility; these grants only allow the roles
-- to reach the tables at all.
-- ----------------------------------------------------------------------------
grant usage on schema content to anon, authenticated, service_role;

grant select on all tables in schema content to anon;
grant select, insert, update, delete on all tables in schema content to authenticated;
grant all on all tables in schema content to service_role;

alter default privileges in schema content grant select on tables to anon;
alter default privileges in schema content
    grant select, insert, update, delete on tables to authenticated;
alter default privileges in schema content grant all on tables to service_role;
