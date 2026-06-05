-- RLS verification for the content schema + storage, using two users.
--
-- Run against a running local stack:
--   supabase start
--   psql "$(supabase status -o env | grep DB_URL | cut -d= -f2- | tr -d '\"')" \
--     -f tests/rls_verification.sql
--
-- Each persona runs inside ONE transaction so the transaction-local JWT claims
-- (request.jwt.claims) + SET ROLE apply to every statement, exactly like a
-- PostgREST request. Seed rows are inserted as the superuser (bypasses RLS),
-- then reads/writes are attempted as the `authenticated`/`anon` roles.
--
-- Expected: cross-user private data is invisible/unwritable; books with
-- visibility='public' (and their pages/items/page_items) are readable by
-- others and by anon; storage objects are visible only to their owner.

\set ON_ERROR_STOP on
\pset pager off

insert into auth.users (id, email) values
  ('11111111-1111-1111-1111-111111111111', 'alice@test.dev'),
  ('22222222-2222-2222-2222-222222222222', 'bob@test.dev');

\echo '### profiles auto-created by the signup trigger (expect 2 rows):'
select id, subscription_tier from public.profiles order by id;

insert into content.books (id, user_id, title, visibility) values
  ('aaaaaaa1-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'Alice Private', 'private'),
  ('aaaaaaa2-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 'Alice Public',  'public'),
  ('bbbbbbb1-0000-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'Bob Private',   'private');

insert into content.pages (id, book_id, page_number) values
  ('aaaaaaa2-1111-0000-0000-000000000001', 'aaaaaaa2-0000-0000-0000-000000000002', 1);

insert into content.items (id, user_id, original_image_key, original_image_url) values
  ('aaaaaaa2-2222-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'k2', 'u2');

\echo '### items default processing_status (expect completed):'
select distinct processing_status from content.items;

insert into content.page_items (page_id, item_id, position_x, position_y, width, height) values
  ('aaaaaaa2-1111-0000-0000-000000000001', 'aaaaaaa2-2222-0000-0000-000000000001', 0,0,10,10);

insert into storage.objects (bucket_id, name, owner) values
  ('cutouts', '11111111-1111-1111-1111-111111111111/cut1.png', '11111111-1111-1111-1111-111111111111');

\echo ''
\echo '=== AS BOB: must NOT see Alice private; CAN see Alice public ==='
begin;
  set local role authenticated;
  set local request.jwt.claims = '{"sub":"22222222-2222-2222-2222-222222222222","role":"authenticated"}';

  \echo 'books visible to Bob (expect Alice Public + Bob Private, NOT Alice Private):'
  select title, visibility from content.books order by title;

  \echo 'public-book content visible (expect pages=1 items=1 page_items=1):'
  select (select count(*) from content.pages) as pages,
         (select count(*) from content.items) as items,
         (select count(*) from content.page_items) as page_items;

  \echo 'Bob UPDATE on Alice public book (expect UPDATE 0 -- public is read-only):'
  update content.books set title = 'HACKED' where title = 'Alice Public';

  \echo 'Bob INSERT spoofing Alice user_id (expect RLS violation):'
  savepoint sp;
  do $$ begin
    insert into content.books (user_id, title) values ('11111111-1111-1111-1111-111111111111','Spoofed');
    raise notice 'FAIL: spoof insert succeeded';
  exception when others then
    raise notice 'OK: spoof insert blocked (%)', sqlerrm;
  end $$;
  rollback to savepoint sp;

  \echo 'Bob reads Alice cutout (expect 0):'
  select count(*) from storage.objects
    where bucket_id='cutouts' and (storage.foldername(name))[1] = '11111111-1111-1111-1111-111111111111';
commit;

\echo ''
\echo '=== AS ALICE: sees her own data, can write it ==='
begin;
  set local role authenticated;
  set local request.jwt.claims = '{"sub":"11111111-1111-1111-1111-111111111111","role":"authenticated"}';

  \echo 'books visible to Alice (expect her 2 books, NOT Bob Private):'
  select title, visibility from content.books order by title;

  \echo 'Alice updates her own book (expect UPDATE 1):'
  update content.books set description = 'mine' where title = 'Alice Private';

  \echo 'Alice reads her own cutout (expect 1):'
  select count(*) from storage.objects
    where bucket_id='cutouts' and (storage.foldername(name))[1] = auth.uid()::text;
commit;

\echo ''
\echo '=== AS ANON: only public books ==='
begin;
  set local role anon;
  \echo 'books visible to anon (expect ONLY Alice Public):'
  select title, visibility from content.books order by title;
commit;
