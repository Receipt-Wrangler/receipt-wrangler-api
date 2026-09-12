import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/models/user_model.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/top_app_bar.dart';

import '../helpers/auth_test_helpers.dart';
import '../helpers/permission_test_helpers.dart';

/// The loading bar must never evict the toolbar it sits under.
///
/// `AppBar` re-parents its toolbar into a `Column` **only** when `bottom` is
/// non-null, so a null <-> non-null swap changes that slot's runtimeType,
/// fails `Widget.canUpdate` and unmounts `leading`, `title` and every
/// `actions` child. Anything in `actions` that holds its own `BuildContext`
/// across an `await` then fails its `mounted` guard and silently gives up --
/// which is exactly how the receipts-screen overflow menu's Quick Scan and
/// Upload from Gallery items died, since the refresh they await is what raises
/// the bar.
void main() {
  api.Claims claims(String displayName) =>
      api.Claims((b) => b..displayName = displayName);

  /// Mounts a [TopAppBar] carrying one probe action, and returns the probe's
  /// [BuildContext] holder plus the [LoadingModel] driving the bar.
  ///
  /// The probe records its context with `??=`: the point of these tests is that
  /// the element is never torn down and rebuilt, so a later capture would paper
  /// over the very failure under test.
  Future<(List<BuildContext?>, LoadingModel)> pumpBar(
      WidgetTester tester) async {
    final authModel = MockAuthModel();
    when(() => authModel.claims).thenReturn(claims('Admin'));

    final loadingModel = LoadingModel();
    final probe = <BuildContext?>[null];

    await tester.pumpWidget(
      MultiProvider(
        providers: [
          ChangeNotifierProvider<AuthModel>(create: (_) => authModel),
          ChangeNotifierProvider<LoadingModel>.value(value: loadingModel),
          ChangeNotifierProvider<UserModel>(create: (_) => UserModel()),
          ChangeNotifierProvider<PermissionsModel>(
            create: (_) => seededPermissions(),
          ),
        ],
        child: MaterialApp(
          home: Scaffold(
            appBar: TopAppBar(
              titleText: 'Home',
              actions: [
                Builder(builder: (context) {
                  probe[0] ??= context;
                  return const SizedBox.shrink();
                }),
              ],
            ),
          ),
        ),
      ),
    );
    await tester.pump();

    return (probe, loadingModel);
  }

  testWidgets('an action keeps its context when the loading bar appears',
      (tester) async {
    final (probe, loadingModel) = await pumpBar(tester);
    expect(probe[0], isNotNull);

    loadingModel.setIsLoading(true);
    await tester.pump();

    expect(find.byType(LinearProgressIndicator), findsOneWidget);
    expect(probe[0]!.mounted, isTrue,
        reason: 'raising the loading bar unmounted the AppBar toolbar, so an '
            'action captured before an await can no longer be used after it');
  });

  testWidgets('an action keeps its context when the loading bar goes away',
      (tester) async {
    final (probe, loadingModel) = await pumpBar(tester);

    loadingModel.setIsLoading(true);
    await tester.pump();
    loadingModel.setIsLoading(false);
    await tester.pump();

    expect(find.byType(LinearProgressIndicator), findsNothing);
    expect(probe[0]!.mounted, isTrue);
  });

  testWidgets('the idle app bar is still kToolbarHeight tall', (tester) async {
    await pumpBar(tester);

    // The always-mounted bottom must be zero-height when idle, or every screen
    // gains a strip. Test MediaQuery has no top padding, so this is the whole
    // app bar.
    expect(tester.getSize(find.byType(TopAppBar)).height, kToolbarHeight);
  });

  testWidgets('the idle app bar mounts no running animation', (tester) async {
    await pumpBar(tester);

    // An always-mounted LinearProgressIndicator animates forever, which would
    // hang every pumpAndSettle in the suite. The explicit timeout makes that
    // regression fail here in two seconds rather than after the ten-minute
    // default.
    await tester.pumpAndSettle(
      const Duration(milliseconds: 100),
      EnginePhase.sendSemanticsUpdate,
      const Duration(seconds: 2),
    );
  });
}
