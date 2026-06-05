-- 20260605000002_profiles.sql
-- ============================================================================
-- public.profiles
--
-- Replaces the bespoke per-user columns that lived on the Go schema's
-- auth.users table (display_name, bio, avatar, subscription_tier, ...).
-- Supabase Auth owns auth.users; profile data lives here, 1:1 with a user.
-- A trigger inserts a profile row whenever a new auth user signs up.
-- ============================================================================

create table public.profiles (
    id uuid primary key references auth.users(id) on delete cascade,
    display_name text,
    bio text,
    avatar_url text,
    subscription_tier text not null default 'free'
        check (subscription_tier in ('free', 'pro', 'creator')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- Reuse the content schema's updated_at helper for consistency.
create trigger update_profiles_updated_at before update on public.profiles
    for each row execute function content.set_updated_at();

-- ----------------------------------------------------------------------------
-- Insert a profile row on signup.
--
-- SECURITY DEFINER so it runs as the function owner (the migration/superuser
-- role) regardless of which role triggers the auth.users insert. The display
-- name is seeded from the signup metadata when present.
-- ----------------------------------------------------------------------------
create or replace function public.handle_new_user()
returns trigger
language plpgsql
security definer
set search_path = ''
as $$
begin
    insert into public.profiles (id, display_name)
    values (
        new.id,
        coalesce(
            new.raw_user_meta_data ->> 'display_name',
            new.raw_user_meta_data ->> 'full_name'
        )
    );
    return new;
end;
$$;

create trigger on_auth_user_created
    after insert on auth.users
    for each row execute function public.handle_new_user();

-- ----------------------------------------------------------------------------
-- RLS for profiles: readable by anyone (needed to show author names on public
-- books), but a user may only insert/update their own row. Deletes cascade
-- from auth.users, so no delete policy is exposed.
-- ----------------------------------------------------------------------------
alter table public.profiles enable row level security;

create policy "Profiles are viewable by everyone"
    on public.profiles for select
    using (true);

create policy "Users can insert their own profile"
    on public.profiles for insert to authenticated
    with check (auth.uid() = id);

create policy "Users can update their own profile"
    on public.profiles for update to authenticated
    using (auth.uid() = id)
    with check (auth.uid() = id);
