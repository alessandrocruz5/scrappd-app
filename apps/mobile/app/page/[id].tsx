// Page editor route. The Skia canvas, gestures, and persistence live in
// src/editor/page-editor.tsx; this file just resolves the page id from the
// route and mounts it.

import { Stack, useLocalSearchParams } from 'expo-router';

import { PageEditor } from '@/editor/page-editor';

export default function PageEditorScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const pageId = id ?? '';

  return (
    <>
      <Stack.Screen options={{ title: 'Page' }} />
      <PageEditor pageId={pageId} />
    </>
  );
}
