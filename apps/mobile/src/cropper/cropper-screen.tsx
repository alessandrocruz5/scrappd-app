// Instant shape cropper.
//
// Aim a shape at a pattern/poster in the live camera, capture, and the framed
// region is clipped through the shape into a transparent PNG cutout (pure
// client-side Skia). The cutout uploads to the private 'cutouts' bucket and an
// 'items' row is inserted as 'completed' — no AI, no polling. A library pick
// runs the same pipeline.
//
// Web caveat: expo-camera's live capture is limited in browsers, so on web we
// skip the camera entirely and drive the same shape -> cutout -> upload
// pipeline from an expo-image-picker file upload. Camera-only UI (the live
// preview, shutter, and permission gates) is gated behind a Platform check.

import { Ionicons } from '@expo/vector-icons';
import { CameraView, useCameraPermissions } from 'expo-camera';
import * as ImagePicker from 'expo-image-picker';
import { useRef, useState } from 'react';
import {
  ActivityIndicator,
  Image,
  Platform,
  StyleSheet,
  Text,
  TouchableOpacity,
  useWindowDimensions,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { AppButton } from '@/components/ui';
import { colors, radius, spacing } from '@/theme/colors';

import { createCutout, type Cutout } from './create-cutout';
import { ShapeOverlay } from './shape-overlay';
import { ShapePicker } from './shape-picker';
import { type ShapeId } from './shapes';
import { uploadCutout } from './upload-cutout';

type Phase = 'aim' | 'working' | 'done';

// Live camera capture is a native-only path; web falls back to file upload.
const isWeb = Platform.OS === 'web';

export function CropperScreen() {
  const { width, height } = useWindowDimensions();
  const viewport = Math.min(width, height * 0.58);

  const [permission, requestPermission] = useCameraPermissions();
  const cameraRef = useRef<CameraView>(null);

  const [shape, setShape] = useState<ShapeId>('square');
  const [phase, setPhase] = useState<Phase>('aim');
  const [result, setResult] = useState<Cutout | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Shared finishing step: shape -> cutout -> upload -> item row.
  async function processImage(uri: string) {
    setPhase('working');
    setError(null);
    try {
      const cutout = await createCutout(uri, shape);
      await uploadCutout(cutout, shape);
      setResult(cutout);
      setPhase('done');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.');
      setPhase('aim');
    }
  }

  async function handleCapture() {
    const photo = await cameraRef.current?.takePictureAsync({ quality: 1 });
    if (photo?.uri) {
      await processImage(photo.uri);
    }
  }

  async function handlePickFromLibrary() {
    const picked = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      quality: 1,
    });
    if (!picked.canceled && picked.assets[0]?.uri) {
      await processImage(picked.assets[0].uri);
    }
  }

  function reset() {
    setResult(null);
    setError(null);
    setPhase('aim');
  }

  // --- Permission gates (native only; web never touches the camera) ---------
  if (!isWeb && !permission) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.primary} />
      </View>
    );
  }

  if (!isWeb && !permission?.granted) {
    return (
      <SafeAreaView style={styles.center}>
        <Ionicons name="camera-outline" size={56} color={colors.accent} />
        <Text style={styles.permissionTitle}>Camera access needed</Text>
        <Text style={styles.permissionText}>
          Scrappd uses the camera so you can aim a shape at a pattern and snap
          an instant cutout.
        </Text>
        <View style={styles.permissionActions}>
          <AppButton label="Allow camera" onPress={requestPermission} />
          <AppButton
            label="Pick from library instead"
            variant="text"
            onPress={handlePickFromLibrary}
          />
        </View>
      </SafeAreaView>
    );
  }

  // --- Result preview -------------------------------------------------------
  if (phase === 'done' && result) {
    return (
      <SafeAreaView style={styles.container} edges={['bottom']}>
        <View style={styles.resultWrap}>
          <View style={[styles.checker, { width: viewport, height: viewport }]}>
            <Image
              source={{ uri: `data:image/png;base64,${result.base64}` }}
              style={{ width: viewport, height: viewport }}
              resizeMode="contain"
            />
          </View>
          <View style={styles.savedBadge}>
            <Ionicons
              name="checkmark-circle"
              size={20}
              color={colors.success}
            />
            <Text style={styles.savedText}>Cutout saved to your items</Text>
          </View>
        </View>
        <View style={styles.actions}>
          <AppButton label="New cutout" onPress={reset} />
        </View>
      </SafeAreaView>
    );
  }

  // --- Web: file upload instead of live camera ------------------------------
  if (isWeb) {
    return (
      <SafeAreaView style={styles.container} edges={['bottom']}>
        <View style={[styles.viewport, { width: viewport, height: viewport }]}>
          <View style={[StyleSheet.absoluteFill, styles.webViewport]}>
            <ShapeOverlay shape={shape} size={viewport} />
          </View>
          {phase === 'working' ? (
            <View style={[StyleSheet.absoluteFill, styles.workingOverlay]}>
              <ActivityIndicator color={colors.white} size="large" />
              <Text style={styles.workingText}>Cutting out…</Text>
            </View>
          ) : null}
        </View>

        <ShapePicker
          selected={shape}
          onSelect={setShape}
          disabled={phase === 'working'}
        />

        {error ? (
          <View style={styles.errorBox}>
            <Ionicons
              name="alert-circle-outline"
              size={18}
              color={colors.error}
            />
            <Text style={styles.errorText}>{error}</Text>
          </View>
        ) : null}

        <View style={styles.webActions}>
          <Text style={styles.webHint}>
            Pick an image and we’ll crop it into the selected shape.
          </Text>
          <AppButton
            label="Choose an image"
            onPress={handlePickFromLibrary}
            disabled={phase === 'working'}
          />
        </View>
      </SafeAreaView>
    );
  }

  // --- Live aiming (native) -------------------------------------------------
  return (
    <SafeAreaView style={styles.container} edges={['bottom']}>
      <View style={[styles.viewport, { width: viewport, height: viewport }]}>
        <CameraView
          ref={cameraRef}
          style={StyleSheet.absoluteFill}
          facing="back"
        />
        <View style={StyleSheet.absoluteFill} pointerEvents="none">
          <ShapeOverlay shape={shape} size={viewport} />
        </View>
        {phase === 'working' ? (
          <View style={[StyleSheet.absoluteFill, styles.workingOverlay]}>
            <ActivityIndicator color={colors.white} size="large" />
            <Text style={styles.workingText}>Cutting out…</Text>
          </View>
        ) : null}
      </View>

      <ShapePicker
        selected={shape}
        onSelect={setShape}
        disabled={phase === 'working'}
      />

      {error ? (
        <View style={styles.errorBox}>
          <Ionicons
            name="alert-circle-outline"
            size={18}
            color={colors.error}
          />
          <Text style={styles.errorText}>{error}</Text>
        </View>
      ) : null}

      <View style={styles.captureRow}>
        <TouchableOpacity
          onPress={handlePickFromLibrary}
          disabled={phase === 'working'}
          style={styles.sideButton}
          activeOpacity={0.8}
        >
          <Ionicons
            name="images-outline"
            size={26}
            color={colors.textSecondary}
          />
        </TouchableOpacity>

        <TouchableOpacity
          onPress={handleCapture}
          disabled={phase === 'working'}
          style={styles.shutter}
          activeOpacity={0.8}
        >
          <View style={styles.shutterInner} />
        </TouchableOpacity>

        {/* Spacer to keep the shutter centred. */}
        <View style={styles.sideButton} />
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.background,
    padding: spacing.xl,
    gap: spacing.md,
  },
  viewport: {
    alignSelf: 'center',
    marginTop: spacing.lg,
    borderRadius: radius.lg,
    overflow: 'hidden',
    backgroundColor: colors.black,
  },
  webViewport: {
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.surface,
  },
  workingOverlay: {
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
    backgroundColor: 'rgba(27, 14, 3, 0.55)',
  },
  workingText: {
    color: colors.white,
    fontSize: 15,
    fontWeight: '600',
  },
  captureRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.xxl,
    marginTop: 'auto',
    paddingBottom: spacing.lg,
  },
  sideButton: {
    width: 56,
    height: 56,
    alignItems: 'center',
    justifyContent: 'center',
  },
  shutter: {
    width: 76,
    height: 76,
    borderRadius: 38,
    borderWidth: 4,
    borderColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.surface,
  },
  shutterInner: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.primary,
  },
  errorBox: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    marginHorizontal: spacing.lg,
    backgroundColor: 'rgba(239, 68, 68, 0.1)',
    borderRadius: radius.sm,
    padding: spacing.md,
  },
  errorText: {
    color: colors.error,
    flex: 1,
    fontSize: 14,
  },
  permissionTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: colors.textPrimary,
  },
  permissionText: {
    fontSize: 15,
    color: colors.textSecondary,
    textAlign: 'center',
  },
  permissionActions: {
    alignSelf: 'stretch',
    marginTop: spacing.lg,
    gap: spacing.sm,
  },
  resultWrap: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.lg,
  },
  checker: {
    borderRadius: radius.lg,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.surface,
  },
  savedBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  savedText: {
    fontSize: 15,
    fontWeight: '600',
    color: colors.textSecondary,
  },
  actions: {
    paddingHorizontal: spacing.xl,
    paddingBottom: spacing.lg,
  },
  webActions: {
    marginTop: 'auto',
    paddingHorizontal: spacing.xl,
    paddingBottom: spacing.lg,
    gap: spacing.md,
  },
  webHint: {
    fontSize: 15,
    color: colors.textSecondary,
    textAlign: 'center',
  },
});
