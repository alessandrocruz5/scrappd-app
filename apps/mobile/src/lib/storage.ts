// Helpers for the private 'cutouts' Storage bucket.
//
// The cropper uploads cutouts to "<auth.uid()>/<file>.png" in a private bucket
// (see packages/supabase/migrations/...storage_buckets.sql), so the stored
// item URL isn't directly viewable. Sign the object key on demand to display it.

import { supabase } from './supabase';

const CUTOUTS_BUCKET = 'cutouts';
const SIGNED_URL_TTL_SECONDS = 60 * 60; // 1 hour

export async function signCutoutUrl(key: string): Promise<string | null> {
  const { data, error } = await supabase.storage
    .from(CUTOUTS_BUCKET)
    .createSignedUrl(key, SIGNED_URL_TTL_SECONDS);
  if (error || !data) return null;
  return data.signedUrl;
}
