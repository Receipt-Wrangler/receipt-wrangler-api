import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/groups/nav/group/group_app_bar.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/receipt_entry_overflow_menu.dart';

import '../helpers/permission_test_helpers.dart';
import '../helpers/receipt_entry_test_helpers.dart';
import '../helpers/receipt_form_test_helpers.dart';

/// The overflow is the accessible route to everything the Scan slot hides behind
/// a long-press, so it must carry exactly the same permission gating -- and it
/// belongs to the receipts screen only, not to every screen the group shell's
/// app bar covers.
void main() {
  const quickScan = api.Permission.groupPeriodReceiptsPeriodQuickScan;
  const create = api.Permission.groupPeriodReceiptsPeriodCreate;
  const overflowKey = ValueKey('receipt-entry-overflow-menu');

  Future<void> pumpMenu(
    WidgetTester tester, {
    required bool aiEnabled,
    required List<api.Permission> permissions,
  }) async {
    final router = GoRouter(
      initialLocation: '/groups/5/receipts',
      routes: [
        GoRoute(
          path: '/groups/:groupId/receipts',
          builder: (context, state) => const Scaffold(
            body: Row(children: [ReceiptEntryOverflowMenu()]),
          ),
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

  /// Pumps the real group shell app bar at [location], which is what decides
  /// whether the overflow belongs on the screen at all.
  Future<void> pumpAppBar(WidgetTester tester, String location) async {
    final router = GoRouter(
      initialLocation: location,
      routes: [
        GoRoute(
          path: '/groups/:groupId/receipts',
          builder: (context, state) =>
              const Scaffold(appBar: GroupAppBar(), body: SizedBox.shrink()),
        ),
        GoRoute(
          path: '/groups/:groupId/dashboards',
          builder: (context, state) =>
              const Scaffold(appBar: GroupAppBar(), body: SizedBox.shrink()),
        ),
      ],
    );

    await tester.pumpWidget(pumpReceiptEntryApp(
      router: router,
      aiEnabled: true,
      permissions: seededPermissions(group: {
        5: [quickScan, create]
      }),
      groups: [buildGroup(id: 5, name: 'Household')],
    ));
    await tester.pump();
  }

  testWidgets('offers every entry the user is permitted', (tester) async {
    await pumpMenu(tester, aiEnabled: true, permissions: [quickScan, create]);

    await tester.tap(find.byKey(overflowKey));
    await tester.pumpAndSettle();

    expect(find.text(quickScanLabel), findsOneWidget);
    expect(find.text(addManualReceiptLabel), findsOneWidget);
    expect(find.text(uploadPhotoLabel), findsOneWidget);
      expect(find.text(uploadFileLabel), findsOneWidget);
  });

  testWidgets('drops the scan entries when Quick Scan cannot run',
      (tester) async {
    await pumpMenu(tester, aiEnabled: false, permissions: [quickScan, create]);

    await tester.tap(find.byKey(overflowKey));
    await tester.pumpAndSettle();

    expect(find.text(addManualReceiptLabel), findsOneWidget);
    expect(find.text(quickScanLabel), findsNothing);
    expect(find.text(uploadPhotoLabel), findsNothing);
      expect(find.text(uploadFileLabel), findsNothing);
  });

  testWidgets('drops manual entry without the create permission',
      (tester) async {
    await pumpMenu(tester, aiEnabled: true, permissions: [quickScan]);

    await tester.tap(find.byKey(overflowKey));
    await tester.pumpAndSettle();

    expect(find.text(addManualReceiptLabel), findsNothing);
    expect(find.text(quickScanLabel), findsOneWidget);
  });

  testWidgets('the whole button is hidden when the user can do neither',
      (tester) async {
    await pumpMenu(tester,
        aiEnabled: true,
        permissions: [api.Permission.groupPeriodReceiptsPeriodRead]);

    expect(find.byKey(overflowKey), findsNothing,
        reason: 'an overflow that could only ever say "no" is worse than none');
  });

  testWidgets('rides on the receipts screen', (tester) async {
    await pumpAppBar(tester, '/groups/5/receipts');

    expect(find.byKey(overflowKey), findsOneWidget);
  });

  testWidgets('is absent from the dashboards screen the app bar also covers',
      (tester) async {
    await pumpAppBar(tester, '/groups/5/dashboards');

    expect(find.byKey(overflowKey), findsNothing);
  });
}
