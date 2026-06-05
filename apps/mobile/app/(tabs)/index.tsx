import { StyleSheet, View } from 'react-native';

import { Placeholder } from '@/components/ui';
import { colors } from '@/theme/colors';

// Placeholder for the instant shape cropper (Skia, client-side) — built in a
// later milestone.
export default function CropperScreen() {
  return (
    <View style={styles.container}>
      <Placeholder
        icon="cut-outline"
        title="Instant Cropper"
        subtitle="Pick a shape, aim it at a pattern, and snap a cutout. Coming soon."
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
});
