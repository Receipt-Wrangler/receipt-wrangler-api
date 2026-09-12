import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_nav.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/scan_nav_item.dart';

import '../helpers/permission_test_helpers.dart';
import '../helpers/receipt_entry_test_helpers.dart';
import '../helpers/receipt_form_test_helpers.dart';

const _filler = NavDestinationItem(
  id: 'filler',
  destination:
      NavigationDestination(icon: Icon(Icons.receipt), label: 'Receipts'),
);

/// The single definition of the add/scan button. Both bottom navs mount what
/// this produces, so what it decides here is what every screen shows.
void main() {
  const quickScan = api.Permission.groupPeriodReceiptsPeriodQuickScan;
  const create = api.Permission.groupPeriodReceiptsPeriodCreate;

  late NavDestinationItem? item;
  late GlobalKey anchorKey;

  // One controller per test, closed afterwards. Built here rather than inside
  // the route builder below, which can re-run: a fresh controller per rebuild
  // would leave BottomNav subscribed to the previous one. Mirrors
  // test/shared/widgets/bottom_nav_test.dart.
  //
  // Broadcast, unlike that file's: the slot is omitted when the user holds
  // neither permission, so BottomNav is never mounted and never listens. A
  // single-subscription controller's close() future does not complete until
  // something has listened, which hangs tearDown forever.
  late StreamController<int> indexSelectedController;

  setUp(() => indexSelectedController = StreamController<int>.broadcast());
  tearDown(() => indexSelectedController.close());

  Future<void> pumpSlot(
    WidgetTester tester, {
    required bool aiEnabled,
    required List<api.Permission> permissions,
  }) async {
    anchorKey = GlobalKey();
    item = null;

    final router = GoRouter(
      initialLocation: '/groups/5/receipts',
      routes: [
        GoRoute(
          path: '/groups/:groupId/receipts',
          builder: (context, state) {
            item = buildScanNavItem(context, anchorKey);
            final built = item;
            return Scaffold(
              bottomNavigationBar: built == null
                  ? null
                  : BottomNav(
                      // NavigationBar needs two destinations; the filler stands
                      // in for the real navs' unconditional ones.
                      items: [built, _filler],
                      onDestinationSelected: (_) {},
                      getInitialSelectedIndex: () => 0,
                      indexSelectedController: indexSelectedController,
                    ),
            );
          },
        ),
      ],
    );

    await tester.pumpWidget(pumpReceiptEntryApp(
      router: router,
      aiEnabled: aiEnabled,
      permissions: seededPermissions(group: {5: permissions}),
      groups: [buildGroup(id: 5, name: 'Household')],
    ));
    await tester.pump();
  }

  testWidgets('reads "Scan" with the scanner icon when Quick Scan can run',
      (tester) async {
    await pumpSlot(tester, aiEnabled: true, permissions: [quickScan, create]);

    expect(item, isNotNull);
    expect(find.text('Scan'), findsOneWidget);
    expect(find.byIcon(Icons.document_scanner), findsOneWidget);
    expect(find.byIcon(Icons.add), findsNothing);
  });

  testWidgets('falls back to "Add" when the ai flag is off', (tester) async {
    await pumpSlot(tester, aiEnabled: false, permissions: [quickScan, create]);

    expect(find.text('Add'), findsOneWidget);
    expect(find.byIcon(Icons.add), findsOneWidget);
    expect(find.byIcon(Icons.document_scanner), findsNothing,
        reason: 'promising a scanner the install cannot run is the whole thing '
            'this label change avoids');
  });

  testWidgets('falls back to "Add" without the quick-scan permission',
      (tester) async {
    await pumpSlot(tester, aiEnabled: true, permissions: [create]);

    expect(find.text('Add'), findsOneWidget);
  });

  testWidgets('still reads "Scan" for a user who may scan but not create',
      (tester) async {
    await pumpSlot(tester, aiEnabled: true, permissions: [quickScan]);

    expect(find.text('Scan'), findsOneWidget);
  });

  testWidgets('is not built at all when the user can do neither',
      (tester) async {
    await pumpSlot(tester,
        aiEnabled: true,
        permissions: [api.Permission.groupPeriodReceiptsPeriodRead]);

    expect(item, isNull,
        reason: 'an affordance that could only ever be refused is worse than '
            'no affordance');
    expect(find.text('Scan'), findsNothing);
    expect(find.text('Add'), findsNothing);
  });

  testWidgets('a hold opens the menu with the permitted entries',
      (tester) async {
    await pumpSlot(tester, aiEnabled: true, permissions: [quickScan, create]);

    await tester.longPress(find.text('Scan'));
    await tester.pumpAndSettle();

    expect(find.text(quickScanLabel), findsOneWidget);
    expect(find.text(addManualReceiptLabel), findsOneWidget);
    expect(find.text(uploadPhotoLabel), findsOneWidget);
      expect(find.text(uploadFileLabel), findsOneWidget);
  });

  testWidgets("a hold on the blocked slot offers only what the user can do",
      (tester) async {
    await pumpSlot(tester, aiEnabled: false, permissions: [quickScan, create]);

    await tester.longPress(find.text('Add'));
    await tester.pumpAndSettle();

    expect(find.text(addManualReceiptLabel), findsOneWidget);
    expect(find.text(quickScanLabel), findsNothing);
    expect(find.text(uploadPhotoLabel), findsNothing);
      expect(find.text(uploadFileLabel), findsNothing);
  });

  testWidgets('names the hold gesture for assistive tech', (tester) async {
    await pumpSlot(tester, aiEnabled: true, permissions: [quickScan, create]);

    final handle = tester.ensureSemantics();
    expect(
      find.bySemanticsLabel(RegExp('Scan')),
      findsWidgets,
    );
    expect(
      tester
          .getSemantics(find.byIcon(Icons.document_scanner))
          .hint,
      contains('hold'),
      reason: 'a long-press is undiscoverable without being described',
    );
    handle.dispose();
  });

  testWidgets("Material's own long-press tooltip is disabled on the slot",
      (tester) async {
    // A NavigationDestination shows a tooltip on long press (falling back to
    // its label), which would put a competing long-press recognizer in the
    // gesture arena against the one that opens the menu.
    await pumpSlot(tester, aiEnabled: true, permissions: [quickScan, create]);

    expect(item!.destination.tooltip, '');

    await tester.longPress(find.text('Scan'));
    await tester.pumpAndSettle();

    // The filler destination keeps its default tooltip; only the scan slot's
    // must be gone.
    expect(
      find.byWidgetPredicate(
          (w) => w is Tooltip && (w.message == 'Scan' || w.message == 'Add')),
      findsNothing,
      reason: 'no tooltip should be competing for the hold',
    );
    expect(find.text(quickScanLabel), findsOneWidget,
        reason: 'the hold reached the menu');
  });
}
