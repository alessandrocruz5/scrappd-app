/* eslint-disable import/first -- jest.mock() must be hoisted above the imports it stubs. */
// Covers exportPageView: the "canvas not ready" guard, the capture call, and
// the independent save-to-library / share branches (each gated on a permission
// or availability check, neither fatal to the other).

import type { RefObject } from 'react';
import type { View } from 'react-native';

// Prefixed with `mock` so jest's hoisted factories may reference them.
const mockCaptureRef = jest.fn();
const mockRequestPermissionsAsync = jest.fn();
const mockSaveToLibraryAsync = jest.fn();
const mockIsAvailableAsync = jest.fn();
const mockShareAsync = jest.fn();

jest.mock('react-native-view-shot', () => ({
  captureRef: (...args: unknown[]) => mockCaptureRef(...args),
}));
jest.mock('expo-media-library', () => ({
  requestPermissionsAsync: () => mockRequestPermissionsAsync(),
  saveToLibraryAsync: (uri: string) => mockSaveToLibraryAsync(uri),
}));
jest.mock('expo-sharing', () => ({
  isAvailableAsync: () => mockIsAvailableAsync(),
  shareAsync: (...args: unknown[]) => mockShareAsync(...args),
}));

import { exportPageView } from '@/editor/export-page';

const ref = { current: {} } as RefObject<View | null>;

beforeEach(() => {
  jest.clearAllMocks();
  mockCaptureRef.mockResolvedValue('file:///tmp/page.png');
  mockRequestPermissionsAsync.mockResolvedValue({ granted: true });
  mockSaveToLibraryAsync.mockResolvedValue(undefined);
  mockIsAvailableAsync.mockResolvedValue(true);
  mockShareAsync.mockResolvedValue(undefined);
});

describe('exportPageView', () => {
  it('throws when the canvas ref is not mounted yet', async () => {
    const empty = { current: null } as RefObject<View | null>;
    await expect(
      exportPageView(empty, { width: 1080, height: 1920 }),
    ).rejects.toThrow(/page canvas isn’t ready/);
    expect(mockCaptureRef).not.toHaveBeenCalled();
  });

  it('captures at the requested design resolution as a PNG', async () => {
    await exportPageView(ref, {
      width: 1080,
      height: 1920,
      fileName: 'scrappd-page-1',
    });
    expect(mockCaptureRef).toHaveBeenCalledWith(
      ref,
      expect.objectContaining({
        format: 'png',
        width: 1080,
        height: 1920,
        result: 'tmpfile',
        fileName: 'scrappd-page-1',
      }),
    );
  });

  it('saves to the library and shares when both are available', async () => {
    const result = await exportPageView(ref, { width: 100, height: 100 });
    expect(mockSaveToLibraryAsync).toHaveBeenCalledWith('file:///tmp/page.png');
    expect(mockShareAsync).toHaveBeenCalledTimes(1);
    expect(result).toEqual({
      uri: 'file:///tmp/page.png',
      savedToLibrary: true,
      shared: true,
    });
  });

  it('skips the library save when permission is denied (still shares)', async () => {
    mockRequestPermissionsAsync.mockResolvedValue({ granted: false });
    const result = await exportPageView(ref, { width: 100, height: 100 });
    expect(mockSaveToLibraryAsync).not.toHaveBeenCalled();
    expect(result.savedToLibrary).toBe(false);
    expect(result.shared).toBe(true);
  });

  it('skips sharing when the share sheet is unavailable (still saves)', async () => {
    mockIsAvailableAsync.mockResolvedValue(false);
    const result = await exportPageView(ref, { width: 100, height: 100 });
    expect(mockShareAsync).not.toHaveBeenCalled();
    expect(result.shared).toBe(false);
    expect(result.savedToLibrary).toBe(true);
  });
});
