import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';

import 'helpers/feature_flags.dart';
import 'helpers/login.dart';
import 'helpers/nav.dart';
import 'helpers/permission_fixtures.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/receipt_test_helpers.dart';

/// Permission gating of the receipt-entry affordances against a real backend.
///
/// The widget suite covers the full gate matrix; this proves the same decisions
/// hold when the permissions come from real roles over the wire rather than a
/// seeded model -- including the case the widget tests cannot reach, where the
/// backend's own role definitions decide what the user holds.
///
/// The slot is a direct action now: a **tap** scans (or opens the manual form
/// when Quick Scan is blocked), and a **hold** opens the menu. What the menu
/// offers is what these assert.
///
/// Note: every account owns a personal "My Receipts" group with create allowed,
/// so the "held in any group" fallback is always satisfiable -- the deny cases
/// are therefore exercised per-group, from inside a group.
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  /// Holds the scan/add slot and returns once its menu is on screen.
  Future<void> openEntryMenu(WidgetTester tester) async {
    await pumpUntilFound(tester, scanNavSlot());
    await tester.longPress(scanNavSlot());
    for (int i = 0; i < 5; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }
  }

  testWidgets(
    'entry: a Legacy Viewer gets no scan/add slot at all in their group',
    (tester) async {
      final fixture = await provisionPermUser(roleName: 'Legacy Viewer');
      await loginAs(
        tester,
        username: fixture.username,
        password: fixture.password,
      );
      await enterGroup(tester, fixture.groupName!);

      // A Viewer holds neither group.receipts.create nor quick-scan here, so
      // there is no action to offer -- the destination is omitted rather than
      // shown and then refused.
      expect(scanNavSlot(), findsNothing);
      // The receipts screen's overflow is gated on the same reading.
      expect(find.byKey(const ValueKey('receipt-entry-overflow-menu')),
          findsNothing);
    },
  );

  // Asserts the slot's menu OFFERS manual entry, not that tapping opens the
  // form -- that belongs where the feature flag is pinned, since it decides
  // whether a tap scans or falls through. Covered by quick_scan_entry_gated_test
  // (both blocked reasons) and receipt_entry_menu_reopen_test.
  testWidgets(
    'entry: a Legacy Editor is offered manual entry from the slot menu',
    (tester) async {
      final fixture = await provisionPermUser(roleName: 'Legacy Editor');
      await loginAs(
        tester,
        username: fixture.username,
        password: fixture.password,
      );
      await enterGroup(tester, fixture.groupName!);

      await openEntryMenu(tester);
      await pumpUntilFound(tester, find.text(addManualReceiptLabel));
      expect(find.text(addManualReceiptLabel), findsOneWidget);
    },
  );

  testWidgets(
    'entry: group-select offers add when the user can create in any group',
    (tester) async {
      // A fresh account owns its personal "My Receipts" group, so the
      // "held in any group" check passes with no current group selected.
      final fixture = await provisionPermUser();
      await loginAs(
        tester,
        username: fixture.username,
        password: fixture.password,
      );

      await openEntryMenu(tester);
      await pumpUntilFound(tester, find.text(addManualReceiptLabel));
      expect(find.text(addManualReceiptLabel), findsOneWidget);
    },
  );

  testWidgets(
    'entry: no Quick Scan without group.receipts.quick-scan (flag on)',
    (tester) async {
      await enableAiPoweredReceiptsForTest();
      // Legacy Editor MINUS quick-scan: keeps receipts.create (so the slot is
      // still there and manual entry is still offered) but not quick-scan, so
      // Quick Scan is hidden by the permission rather than the flag.
      final fixture = await provisionGroupMemberWithoutPermission(
        'group.receipts.quick-scan',
        baselineRole: 'Legacy Editor',
      );
      await loginAs(
        tester,
        username: fixture.username,
        password: fixture.password,
      );
      await enterGroup(tester, fixture.groupName!);

      // The slot advertises what it will do: no scanner, so no "Scan" label.
      // `.hitTestable()` on both halves matters here: inside a group TWO navs
      // are mounted (the group-select shell sits under the group shell), and the
      // hidden one scopes to "held in any group" -- which this user satisfies
      // through their personal "My Receipts" group, so it legitimately reads
      // "Scan". Only the visible nav describes the group they are in.
      expect(find.text('Add').hitTestable(), findsWidgets);
      expect(find.text('Scan').hitTestable(), findsNothing);

      await openEntryMenu(tester);
      await pumpUntilFound(tester, find.text(addManualReceiptLabel));
      expect(find.text(addManualReceiptLabel), findsOneWidget,
          reason: 'still holds receipts.create');
      expect(find.text(quickScanLabel), findsNothing,
          reason: 'lacks group.receipts.quick-scan');
      expect(find.text(uploadPhotoLabel), findsNothing,
          reason: 'the picker entries feed Quick Scan, so they need the same '
              'permission');
      expect(find.text(uploadFileLabel), findsNothing,
          reason: 'the picker entries feed Quick Scan, so they need the same '
              'permission');
    },
  );

  testWidgets(
    'entry: Quick Scan shown with group.receipts.quick-scan (flag on)',
    (tester) async {
      await enableAiPoweredReceiptsForTest();
      // A full Legacy Editor holds group.receipts.quick-scan.
      final fixture = await provisionPermUser(roleName: 'Legacy Editor');
      await loginAs(
        tester,
        username: fixture.username,
        password: fixture.password,
      );
      await enterGroup(tester, fixture.groupName!);

      expect(find.text('Scan').hitTestable(), findsWidgets,
          reason: 'the slot advertises the direct scan action');

      await openEntryMenu(tester);
      await pumpUntilFound(tester, find.text(quickScanLabel));
      expect(find.text(quickScanLabel), findsOneWidget);
      expect(find.text(uploadPhotoLabel), findsOneWidget);
      expect(find.text(uploadFileLabel), findsOneWidget);
    },
  );
}
