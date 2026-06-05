// Renders a cropper cutout by signing its private Storage key on demand.
// Falls back to a placeholder tile while signing or if the URL can't be made.

import { Ionicons } from '@expo/vector-icons';
import { useEffect, useState } from 'react';
import {
  Image,
  type ImageStyle,
  type StyleProp,
  StyleSheet,
  View,
  type ViewStyle,
} from 'react-native';

import { signCutoutUrl } from '@/lib/storage';
import { colors, radius } from '@/theme/colors';

export function CutoutThumb({
  storageKey,
  style,
}: {
  storageKey: string;
  // Callers pass layout-only styles (size/margin) shared by both the image and
  // the placeholder tile.
  style?: StyleProp<ViewStyle>;
}) {
  const imageStyle = style as StyleProp<ImageStyle>;
  const [url, setUrl] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    signCutoutUrl(storageKey).then((signed) => {
      if (active) setUrl(signed);
    });
    return () => {
      active = false;
    };
  }, [storageKey]);

  if (!url) {
    return (
      <View style={[styles.placeholder, style]}>
        <Ionicons name="image-outline" size={24} color={colors.textHint} />
      </View>
    );
  }

  return (
    <Image source={{ uri: url }} style={[styles.image, imageStyle]} resizeMode="contain" />
  );
}

const styles = StyleSheet.create({
  placeholder: {
    backgroundColor: colors.surface,
    borderRadius: radius.sm,
    alignItems: 'center',
    justifyContent: 'center',
  },
  image: {
    borderRadius: radius.sm,
    backgroundColor: colors.surface,
  },
});
