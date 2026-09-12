import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/receipt_edit_popup_menu.dart';

import 'helpers/api.dart';
import 'helpers/form_actions.dart';
import 'helpers/login.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/receipt_test_helpers.dart';

/// Drives a receipt through every ReceiptStatus value via the
/// view-screen popup menu → Edit form, asserting after each transition that
/// the server-side status matches. Catches regressions where:
///   - the status dropdown ignores user selection in edit mode,
///   - the form submits but doesn't actually persist the new status,
///   - the API client serializes the enum differently than the server expects.
///
/// We compare against the JSON the API returns (`status` is the uppercase
/// enum name) rather than re-driving the UI to read the value back -- the
/// API is the source of truth and a UI round-trip could mask a save bug.
void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  testWidgets('admin can move a receipt through all status transitions',
      (tester) async {
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    final receiptName =
        'e2e-status-${DateTime.now().millisecondsSinceEpoch}';
    final receiptId = await addManualReceiptViaUI(tester, receiptName);
    scheduleReceiptCleanup(receiptId);

    final jwt = await apiLogin();

    // getDefaultReceipt() in lib/utils/receipts.dart sets the initial status
    // to OPEN -- record-and-fail-fast if that ever changes so the rest of
    // the lifecycle assertions stay meaningful.
    final initial = await getReceipt(receiptId, jwt: jwt);
    expect(initial['status'], 'OPEN',
        reason: 'manual receipt should default to OPEN');

    // OPEN is the starting point, so we drive only the remaining transitions
    // plus a round-trip back to OPEN to prove every value is reachable from
    // any other.
    for (final transition in const [
      _StatusTransition('Needs Attention', 'NEEDS_ATTENTION'),
      _StatusTransition('Resolved', 'RESOLVED'),
      _StatusTransition('Draft', 'DRAFT'),
      _StatusTransition('Declined', 'DECLINED'),
      _StatusTransition('Open', 'OPEN'),
    ]) {
      await _changeStatusViaUI(tester, transition.label);

      final updated = await getReceipt(receiptId, jwt: jwt);
      expect(updated['status'], transition.apiValue,
          reason: 'after submitting "${transition.label}" the server '
              'should report status=${transition.apiValue}');
    }
  });
}

class _StatusTransition {
  const _StatusTransition(this.label, this.apiValue);
  final String label;
  final String apiValue;
}

/// Pre-condition: caller is on `/receipts/<id>/view`. Opens the top-right
/// popup menu, taps "Edit", changes the status dropdown to [optionLabel],
/// taps the bottom submit button, and waits to land back on /view.
Future<void> _changeStatusViaUI(WidgetTester tester, String optionLabel) async {
  // Open the view-screen overflow menu and pick "Edit". The PopupMenuButton
  // is gated on `canEditReceipt`, which checks the user's group.receipts.update
  // permission from PermissionsModel -- on cold-boot after navigation,
  // permissions may not be populated yet, so we pump until the menu button is
  // actually mounted instead of tapping immediately.
  final menuButton = find.byType(PopupMenuButton<dynamic>);
  await pumpUntilFound(tester, menuButton);
  await tester.tap(menuButton);
  // The popup scales in; the "Edit" item mounts on the animation's first
  // frame where a tap computed from its center misses (flaked on the iOS
  // simulator: "Offset(1184.8, 99.0) ... would not hit test"). Wait for
  // hittability and drain the open animation -- same hardening as
  // receipt_edit_test.dart and receipt_cost_split_test._navigateToEdit.
  await pumpUntilFound(tester, find.text('Edit').hitTestable());
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
  await tester.tap(find.text('Edit').hitTestable());

  // /edit's destination-mounted marker: the bottom save button, which only
  // renders on edit/add paths -- find.text('Name') matches on /view too,
  // so it cannot prove the navigation happened.
  await pumpUntilFound(tester, find.byType(BottomSubmitButton));

  // The status field sits well below Name on the receipt form; on the
  // 1280x900 test surface it's off-screen until we scroll it into view.
  // Without ensureVisible the dropdown tap silently misses and the
  // option text never renders, so pumpUntilFound inside selectDropdown
  // times out.
  final statusFinder = find.byWidgetPredicate(
    (w) => w is FormBuilderDropdown && w.name == 'status',
  );
  await pumpUntilFound(tester, statusFinder);
  await tester.ensureVisible(statusFinder);
  await tester.pumpAndSettle();

  // Use the proven dropdown helper -- it handles the .last/frame-drain
  // pattern the closed-state child needs.
  await selectDropdown(tester, 'status', optionLabel);

  await tester.pumpAndSettle(const Duration(seconds: 3));
  await tester.tap(find.byType(BottomSubmitButton));
  // /view shell has mounted once ReceiptEditPopupMenu renders -- it
  // only appears on /view per receipt_app_bar_action_builder.dart:56-67.
  await pumpUntilFound(tester, find.byType(ReceiptEditPopupMenu));
}
