// Uploads a finished cutout to Supabase Storage and records it as an item.
//
// Storage layout mirrors the bucket RLS in
// packages/supabase/migrations/...storage_buckets.sql: objects live under
// "<auth.uid()>/<file>" in the private 'cutouts' bucket, and policies restrict
// every user to their own folder. There's no AI/processing step, so the item
// row is inserted already 'completed' (no polling).

import { supabase } from '@/lib/supabase';

import type { Cutout } from './create-cutout';
import type { ShapeId } from './shapes';

export type SavedCutout = {
  itemId: string;
  storagePath: string;
};

export async function uploadCutout(
  cutout: Cutout,
  shape: ShapeId,
): Promise<SavedCutout> {
  const {
    data: { user },
    error: userError,
  } = await supabase.auth.getUser();
  if (userError || !user) {
    throw new Error('You need to be signed in to save a cutout.');
  }

  const fileName = `${Date.now()}-${Math.random().toString(36).slice(2, 10)}.png`;
  const storagePath = `${user.id}/${fileName}`;

  // supabase-js wants an ArrayBuffer in React Native; the Skia-encoded bytes
  // already sit in an exact-length buffer.
  const { error: uploadError } = await supabase.storage
    .from('cutouts')
    .upload(storagePath, cutout.bytes.buffer as ArrayBuffer, {
      contentType: 'image/png',
      upsert: false,
    });
  if (uploadError) {
    throw new Error(`Upload failed: ${uploadError.message}`);
  }

  // The bucket is private; store the object path as the key and a stable
  // public-URL form for reference. The Books/Items UI signs URLs on demand
  // when it displays cutouts.
  const { data: publicUrlData } = supabase.storage
    .from('cutouts')
    .getPublicUrl(storagePath);
  const imageUrl = publicUrlData.publicUrl;

  const { data: item, error: insertError } = await supabase
    .schema('content')
    .from('items')
    .insert({
      user_id: user.id,
      original_image_key: storagePath,
      original_image_url: imageUrl,
      processed_image_key: storagePath,
      processed_image_url: imageUrl,
      processing_status: 'completed',
      mime_type: 'image/png',
      original_width: cutout.width,
      original_height: cutout.height,
      original_file_size_bytes: cutout.bytes.byteLength,
      item_category: 'cutout',
      item_name: shape,
    })
    .select('id')
    .single();

  if (insertError || !item) {
    // Best-effort cleanup so we don't leave an orphaned object behind.
    await supabase.storage.from('cutouts').remove([storagePath]);
    throw new Error(
      `Saved the image but couldn't record the item: ${insertError?.message ?? 'unknown error'}`,
    );
  }

  return { itemId: item.id, storagePath };
}
