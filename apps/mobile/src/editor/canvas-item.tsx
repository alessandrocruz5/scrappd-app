// A single cutout on the page: drag to move, pinch to resize, two-finger twist
// to rotate, tap to select. Gestures run on the UI thread (Gesture Handler +
// Reanimated shared values) for 60fps manipulation; the committed transform is
// handed back to JS on gesture end so the editor can persist it to
// content.page_items. The image itself is a plain RN <Image> of the signed
// cutout, layered over the Skia page background.

import { useEffect } from 'react';
import { Image, StyleSheet, View } from 'react-native';
import { Gesture, GestureDetector } from 'react-native-gesture-handler';
import Animated, {
  runOnJS,
  useAnimatedStyle,
  useSharedValue,
} from 'react-native-reanimated';

import type { PageItemPatch, PageItemWithItem } from '@/books/api';
import { colors, radius } from '@/theme/colors';

import { roundTransform } from './transform';
import { useSignedUrl } from './use-signed-url';

const MIN_SIZE = 60;
const RAD_TO_DEG = 180 / Math.PI;

export function CanvasItem({
  pageItem,
  canvasWidth,
  canvasHeight,
  stackIndex,
  selected,
  onSelect,
  onCommit,
}: {
  pageItem: PageItemWithItem;
  canvasWidth: number;
  canvasHeight: number;
  // Paint order within the canvas. Driven by the item's position in the
  // z_index-sorted list (always >= 1) so even an item sent to the back stays
  // above the empty-canvas deselect layer.
  stackIndex: number;
  selected: boolean;
  onSelect: () => void;
  onCommit: (patch: PageItemPatch) => void;
}) {
  const key =
    pageItem.item?.processed_image_key ?? pageItem.item?.original_image_key;
  const url = useSignedUrl(key);

  // Live transform, seeded from the persisted row.
  const tx = useSharedValue(pageItem.position_x);
  const ty = useSharedValue(pageItem.position_y);
  const w = useSharedValue(pageItem.width);
  const h = useSharedValue(pageItem.height);
  const rot = useSharedValue(pageItem.rotation);

  // Gesture-start snapshots.
  const startX = useSharedValue(0);
  const startY = useSharedValue(0);
  const startW = useSharedValue(0);
  const startH = useSharedValue(0);
  const startRot = useSharedValue(0);

  // Re-seed when the persisted values change from outside a gesture (opacity /
  // z-index edits, a refetch after add/delete). Our own commits keep the cache
  // equal to these shared values, so this never fights an in-flight drag.
  useEffect(() => {
    tx.value = pageItem.position_x;
    ty.value = pageItem.position_y;
    w.value = pageItem.width;
    h.value = pageItem.height;
    rot.value = pageItem.rotation;
  }, [
    pageItem.position_x,
    pageItem.position_y,
    pageItem.width,
    pageItem.height,
    pageItem.rotation,
    tx,
    ty,
    w,
    h,
    rot,
  ]);

  const maxSize = Math.max(canvasWidth, canvasHeight);

  function commit() {
    onCommit(
      roundTransform({
        x: tx.value,
        y: ty.value,
        width: w.value,
        height: h.value,
        rotation: rot.value,
      }),
    );
  }

  const pan = Gesture.Pan()
    .onStart(() => {
      startX.value = tx.value;
      startY.value = ty.value;
      runOnJS(onSelect)();
    })
    .onUpdate((e) => {
      const maxX = Math.max(0, canvasWidth - w.value);
      const maxY = Math.max(0, canvasHeight - h.value);
      tx.value = Math.min(Math.max(startX.value + e.translationX, 0), maxX);
      ty.value = Math.min(Math.max(startY.value + e.translationY, 0), maxY);
    })
    .onEnd(() => runOnJS(commit)());

  const pinch = Gesture.Pinch()
    .onStart(() => {
      startW.value = w.value;
      startH.value = h.value;
    })
    .onUpdate((e) => {
      w.value = Math.min(Math.max(startW.value * e.scale, MIN_SIZE), maxSize);
      h.value = Math.min(Math.max(startH.value * e.scale, MIN_SIZE), maxSize);
    })
    .onEnd(() => runOnJS(commit)());

  const rotation = Gesture.Rotation()
    .onStart(() => {
      startRot.value = rot.value;
    })
    .onUpdate((e) => {
      rot.value = startRot.value + e.rotation * RAD_TO_DEG;
    })
    .onEnd(() => runOnJS(commit)());

  const tap = Gesture.Tap().onEnd(() => runOnJS(onSelect)());

  const gesture = Gesture.Simultaneous(tap, pan, pinch, rotation);

  const animatedStyle = useAnimatedStyle(() => ({
    width: w.value,
    height: h.value,
    transform: [
      { translateX: tx.value },
      { translateY: ty.value },
      { rotateZ: `${rot.value}deg` },
    ],
  }));

  return (
    <GestureDetector gesture={gesture}>
      <Animated.View
        style={[
          styles.item,
          animatedStyle,
          { opacity: pageItem.opacity, zIndex: stackIndex },
          selected ? styles.selected : null,
        ]}
      >
        {url ? (
          <Image
            source={{ uri: url }}
            style={styles.image}
            resizeMode="contain"
          />
        ) : (
          <View style={styles.placeholder} />
        )}
      </Animated.View>
    </GestureDetector>
  );
}

const styles = StyleSheet.create({
  item: {
    position: 'absolute',
    left: 0,
    top: 0,
  },
  image: {
    width: '100%',
    height: '100%',
  },
  placeholder: {
    flex: 1,
    borderRadius: radius.sm,
    backgroundColor: 'rgba(0,0,0,0.06)',
  },
  selected: {
    borderWidth: 2,
    borderColor: colors.primary,
    borderRadius: radius.sm,
  },
});
