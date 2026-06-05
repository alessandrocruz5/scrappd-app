import { StyleSheet, View } from 'react-native';

import { Placeholder } from '@/components/ui';
import { colors } from '@/theme/colors';

// Placeholder for the Books list (Book -> Pages -> Items) — built in a later
// milestone against Supabase.
export default function BooksScreen() {
  return (
    <View style={styles.container}>
      <Placeholder
        icon="book-outline"
        title="Your Books"
        subtitle="Your scrapbooks will live here. Create your first one soon."
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
});
