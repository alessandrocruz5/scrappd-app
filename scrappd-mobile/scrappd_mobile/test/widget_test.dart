// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:scrappd_mobile/main.dart';
import 'package:scrappd_mobile/core/storage/token_storage.dart';

void main() {
  testWidgets('Shows login screen when unauthenticated',
      (WidgetTester tester) async {
    TestWidgetsFlutterBinding.ensureInitialized();
    SharedPreferences.setMockInitialValues({});

    final tokenStorage = TokenStorage();
    await tokenStorage.init();

    await tester.pumpWidget(ScrappdApp(tokenStorage: tokenStorage));
    await tester.pumpAndSettle();

    expect(find.text('Welcome back'), findsOneWidget);
  });
}
