import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart' show RenderParagraph;
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:openapi/openapi.dart' show Permission;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/models/user_model.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/top_app_bar.dart';

import '../helpers/auth_test_helpers.dart';
import '../helpers/permission_test_helpers.dart';

api.Claims _claims(String displayName) =>
    api.Claims((b) => b..displayName = displayName);

void main() {
  // Opens the avatar popup menu (located by its stable key) for a caller holding
  // [appPerms], injecting every owned ChangeNotifier via ChangeNotifierProvider.
  Future<void> pumpBar(WidgetTester tester, List<Permission> appPerms) async {
    final authModel = MockAuthModel();
    when(() => authModel.claims).thenReturn(_claims('Admin'));

    await tester.pumpWidget(
      MultiProvider(
        providers: [
          ChangeNotifierProvider<AuthModel>(create: (_) => authModel),
          ChangeNotifierProvider<LoadingModel>(create: (_) => LoadingModel()),
          ChangeNotifierProvider<UserModel>(create: (_) => UserModel()),
          ChangeNotifierProvider<PermissionsModel>(
            create: (_) => seededPermissions(app: appPerms),
          ),
        ],
        child: const MaterialApp(
          home: Scaffold(appBar: TopAppBar(titleText: 'Home')),
        ),
      ),
    );
    await tester.pump();

    // Tap the avatar by key and pump the finite popup-open animation to
    // completion with explicit pumps (no pumpAndSettle).
    await tester.tap(find.byKey(const ValueKey('user-avatar-menu')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));
  }

  testWidgets('shows Reports with app.reports.read', (tester) async {
    await pumpBar(tester, [Permission.appPeriodReportsPeriodRead]);
    expect(find.text('Reports'), findsOneWidget);
  });

  testWidgets('shows Reports with app.reports.readAll (base absent)',
      (tester) async {
    await pumpBar(tester, [Permission.appPeriodReportsPeriodReadAll]);
    expect(find.text('Reports'), findsOneWidget);
  });

  testWidgets('hides Reports without either read permission', (tester) async {
    await pumpBar(tester, [Permission.appPeriodReceiptsPeriodSearch]);
    expect(find.text('Reports'), findsNothing);
    // The menu did open — the ungated items are present.
    expect(find.text('User Profile'), findsOneWidget);
    expect(find.text('Logout'), findsOneWidget);
  });

  testWidgets('renders the items in text colour, not the brand blue',
      (tester) async {
    await pumpBar(tester, [Permission.appPeriodReportsPeriodRead]);

    // Read the scheme actually in force rather than hardcoding colours, so the
    // assertion keeps meaning if the app's palette changes.
    final scheme = Theme.of(tester.element(find.byType(Scaffold))).colorScheme;
    expect(scheme.onSurface, isNot(scheme.primary),
        reason: 'the two must differ or this test proves nothing');

    // The items are TextButtons, which default their foreground to
    // colorScheme.primary — that is what made them render blue. They must use
    // the colour a bare PopupMenuItem would have used instead.
    //
    // Assert the colour the RenderParagraph actually paints, not the declared
    // ButtonStyle: the style is only the mechanism, and a theme or a wrapping
    // DefaultTextStyle could still override what the user ends up seeing.
    for (final label in ['User Profile', 'Reports', 'Logout']) {
      final painted = tester
          .renderObject<RenderParagraph>(find.text(label))
          .text
          .style
          ?.color;
      expect(painted, scheme.onSurface,
          reason: '"$label" is not painted in the ordinary text colour');
    }
  });
}
