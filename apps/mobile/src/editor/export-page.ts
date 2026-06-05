// Client-side page export — the React Native port of page_export_service.dart.
//
// The Flutter app POSTed to a Go `/pages/:id/render` endpoint and downloaded a
// server-rendered PNG. The revamp drops that server entirely: the page is
// composed on-device (a Skia background with cutout overlays, see
// page-editor.tsx), so we snapshot that exact view to a high-res PNG with
// react-native-view-shot, then hand it to the OS — saved to the photo library
// (expo-media-library) and offered to the share sheet (expo-sharing). No
// network, no render endpoint.

import type { RefObject } from 'react';
import type { View } from 'react-native';

import * as MediaLibrary from 'expo-media-library';
import * as Sharing from 'expo-sharing';
import { captureRef } from 'react-native-view-shot';

export type ExportResult = {
  uri: string;
  savedToLibrary: boolean;
  shared: boolean;
};

export type ExportOptions = {
  // Target pixel dimensions of the PNG. Pass the page's design resolution
  // (canvas_width × canvas_height) so the on-screen view is upscaled to a
  // crisp, full-resolution image rather than the laid-out point size.
  width: number;
  height: number;
  // Base name for the temp file (no extension).
  fileName?: string;
};

export async function exportPageView(
  ref: RefObject<View | null>,
  { width, height, fileName }: ExportOptions,
): Promise<ExportResult> {
  if (!ref.current) {
    throw new Error('The page canvas isn’t ready yet — try again in a moment.');
  }

  // Snapshot the composed page (Skia background + cutout overlays) at the
  // page's design resolution. view-shot upscales the laid-out view to
  // width/height, giving us the high-res export the Flutter `scale: 2.0`
  // render produced — but entirely client-side.
  const uri = await captureRef(ref, {
    format: 'png',
    quality: 1,
    width,
    height,
    result: 'tmpfile',
    fileName,
  });

  // Save to the camera roll when the user grants access. Permission denial is
  // not fatal — the share sheet still lets them keep the file.
  let savedToLibrary = false;
  const permission = await MediaLibrary.requestPermissionsAsync();
  if (permission.granted) {
    await MediaLibrary.saveToLibraryAsync(uri);
    savedToLibrary = true;
  }

  // Offer the system share sheet (AirDrop, Messages, Files, …) when available.
  let shared = false;
  if (await Sharing.isAvailableAsync()) {
    await Sharing.shareAsync(uri, {
      mimeType: 'image/png',
      UTI: 'public.png',
      dialogTitle: 'Share page',
    });
    shared = true;
  }

  return { uri, savedToLibrary, shared };
}
