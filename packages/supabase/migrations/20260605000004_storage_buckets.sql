-- 20260605000004_storage_buckets.sql
-- ============================================================================
-- Storage buckets + per-user RLS.
--
--   * 'cutouts' — shape-cropped PNG cutouts produced by the cropper.
--   * 'exports' — optional rendered page exports.
--
-- Both are private. Objects are namespaced by user id as the first path
-- segment (e.g. "cutouts/<uid>/<file>"), and policies restrict every user to
-- their own folder. This mirrors Supabase's recommended per-user pattern.
-- ============================================================================

insert into storage.buckets (id, name, public)
values
    ('cutouts', 'cutouts', false),
    ('exports', 'exports', false)
on conflict (id) do nothing;

-- ----------------------------------------------------------------------------
-- cutouts: full CRUD on objects under "cutouts/<auth.uid()>/..."
-- ----------------------------------------------------------------------------
create policy "Users can read their own cutouts"
    on storage.objects for select to authenticated
    using (
        bucket_id = 'cutouts'
        and (storage.foldername(name))[1] = auth.uid()::text
    );

create policy "Users can upload their own cutouts"
    on storage.objects for insert to authenticated
    with check (
        bucket_id = 'cutouts'
        and (storage.foldername(name))[1] = auth.uid()::text
    );

create policy "Users can update their own cutouts"
    on storage.objects for update to authenticated
    using (
        bucket_id = 'cutouts'
        and (storage.foldername(name))[1] = auth.uid()::text
    )
    with check (
        bucket_id = 'cutouts'
        and (storage.foldername(name))[1] = auth.uid()::text
    );

create policy "Users can delete their own cutouts"
    on storage.objects for delete to authenticated
    using (
        bucket_id = 'cutouts'
        and (storage.foldername(name))[1] = auth.uid()::text
    );

-- ----------------------------------------------------------------------------
-- exports: same per-user pattern under "exports/<auth.uid()>/..."
-- ----------------------------------------------------------------------------
create policy "Users can read their own exports"
    on storage.objects for select to authenticated
    using (
        bucket_id = 'exports'
        and (storage.foldername(name))[1] = auth.uid()::text
    );

create policy "Users can upload their own exports"
    on storage.objects for insert to authenticated
    with check (
        bucket_id = 'exports'
        and (storage.foldername(name))[1] = auth.uid()::text
    );

create policy "Users can update their own exports"
    on storage.objects for update to authenticated
    using (
        bucket_id = 'exports'
        and (storage.foldername(name))[1] = auth.uid()::text
    )
    with check (
        bucket_id = 'exports'
        and (storage.foldername(name))[1] = auth.uid()::text
    );

create policy "Users can delete their own exports"
    on storage.objects for delete to authenticated
    using (
        bucket_id = 'exports'
        and (storage.foldername(name))[1] = auth.uid()::text
    );
