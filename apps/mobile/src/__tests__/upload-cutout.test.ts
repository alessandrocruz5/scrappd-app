/* eslint-disable import/first -- jest.mock() must be hoisted above the imports it stubs. */
// Covers uploadCutout: the auth guard, the storage-upload failure path, the
// happy path (object uploaded under "<uid>/<file>" + an items row inserted),
// and the orphan-cleanup that removes the uploaded object when the DB insert
// fails.

import { createSupabaseMock } from '@/test-utils/supabase-mock';

const mockDb = createSupabaseMock();
// Lazy getter: the factory is hoisted above `mockDb`'s initialisation, so it
// must not dereference it until first access (which happens at call time).
jest.mock('@/lib/supabase', () => ({
  get supabase() {
    return mockDb.supabase;
  },
}));

import { uploadCutout } from '@/cropper/upload-cutout';
import type { Cutout } from '@/cropper/create-cutout';

const cutout: Cutout = {
  bytes: new Uint8Array([1, 2, 3, 4, 5]),
  base64: 'AQIDBAU=',
  width: 1080,
  height: 1080,
};

beforeEach(() => {
  mockDb.queue.length = 0;
  mockDb.calls.length = 0;
  // Reset the storage spies' call history so per-test `.mock.calls[0]` reads
  // the call this test made, not one leaked from a previous test.
  mockDb.storage.upload.mockReset();
  mockDb.storage.remove.mockReset();
  mockDb.storage.getPublicUrl.mockReset();
  mockDb.getUser.mockResolvedValue({
    data: { user: { id: 'user-1' } },
    error: null,
  });
  mockDb.storage.upload.mockResolvedValue({ data: {}, error: null });
  mockDb.storage.remove.mockResolvedValue({ data: {}, error: null });
  mockDb.storage.getPublicUrl.mockReturnValue({
    data: { publicUrl: 'https://example.test/object.png' },
  });
});

describe('uploadCutout', () => {
  it('requires a signed-in user', async () => {
    mockDb.getUser.mockResolvedValue({ data: { user: null }, error: null });
    await expect(uploadCutout(cutout, 'circle')).rejects.toThrow(
      /signed in to save a cutout/,
    );
    expect(mockDb.storage.upload).not.toHaveBeenCalled();
  });

  it('throws a friendly error when the storage upload fails', async () => {
    mockDb.storage.upload.mockResolvedValue({
      data: null,
      error: { message: 'quota exceeded' },
    });
    await expect(uploadCutout(cutout, 'circle')).rejects.toThrow(
      'Upload failed: quota exceeded',
    );
  });

  it('uploads under the user folder and records a completed item', async () => {
    mockDb.queue.push({ data: { id: 'item-9' }, error: null });

    const result = await uploadCutout(cutout, 'star');

    expect(result.itemId).toBe('item-9');
    // Object path is "<uid>/<file>.png".
    expect(result.storagePath).toMatch(/^user-1\/.+\.png$/);

    const [path, , opts] = mockDb.storage.upload.mock.calls[0];
    expect(path).toBe(result.storagePath);
    expect(opts).toMatchObject({ contentType: 'image/png', upsert: false });

    const insert = mockDb.calls.find((c) => c.method === 'insert')
      ?.args[0] as Record<string, unknown>;
    expect(insert).toMatchObject({
      user_id: 'user-1',
      processing_status: 'completed',
      item_category: 'cutout',
      item_name: 'star',
      mime_type: 'image/png',
      original_width: 1080,
      original_height: 1080,
      original_file_size_bytes: 5,
      original_image_key: result.storagePath,
    });
  });

  it('cleans up the orphaned object when the item insert fails', async () => {
    mockDb.queue.push({ data: null, error: { message: 'rls denied' } });

    await expect(uploadCutout(cutout, 'square')).rejects.toThrow(
      /record the item: rls denied/,
    );
    // The uploaded object is removed so nothing is left dangling.
    expect(mockDb.storage.remove).toHaveBeenCalledTimes(1);
    const [removedPaths] = mockDb.storage.remove.mock.calls[0];
    expect(removedPaths).toHaveLength(1);
  });
});
