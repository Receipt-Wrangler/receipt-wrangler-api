// Flow #8 -- Quick Scan happy path.
//
// Quick Scan is the AI-assisted bulk-receipt entry flow. This covers the
// gallery route into it: hold the bottom-nav scan slot, pick "Upload from
// Gallery", choose one or more images, fill a per-image form (group, paid by,
// status), and submit. The backend queues each image as an async OCR/AI
// extraction job that materializes into a receipt.
//
// The camera route is the slot's plain tap and is covered by
// quick_scan_entry_test.dart, which drives the mocked document scanner.
//
// PRECONDITION: the demo and local backends both have
// `featureConfig.aiPoweredReceipts: true`. If false, the in-app
// flow shows an error snackbar instead of the bottom sheet (see
// mobile/lib/shared/functions/quick_scan.dart:227-231).
//
// Runs on every target. It drives the **file** source, which
// `installFileSelectorMock` intercepts by swapping the platform interface
// before any platform code runs. (The photo source is not asserted here: on
// Linux `image_picker` delegates to `file_selector`, so the two sources are
// indistinguishable there.)

import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';

import 'helpers/feature_flags.dart';
import 'helpers/file_selector_mock.dart';
import 'helpers/form_actions.dart';
import 'helpers/login.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/receipt_test_helpers.dart';
import 'helpers/users.dart';

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  testWidgets('quick scan from a file: pick image, fill form, submit',
      (tester) async {
    // Quick Scan is gated on featureConfig.aiPoweredReceipts, which is off by
    // default on the local backend. Flip it on for this test (restored on
    // teardown) so the menu item appears and the bottom sheet opens. Must run
    // before loginAsAdmin so the app fetches the enabled flag at bootstrap.
    await enableAiPoweredReceiptsForTest();
    await installFileSelectorMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    // Hold the scan slot and pick "Upload from Gallery". Both that entry and
    // the slot's "Scan" label are gated on featureConfig.aiPoweredReceipts
    // plus group.receipts.quick-scan, so their presence verifies the flag --
    // and the sheet re-checks the same gate before opening.
    await pumpUntilFound(tester, scanNavSlot());
    await tester.longPress(scanNavSlot());
    // The menu's items mount on the popup's first frame; a tap computed then
    // misses (deterministic on iOS: "Offset(595.9, 866.0) ... would not hit
    // test"). Wait for hittability, then drain the animation -- same hardening
    // as addManualReceiptViaUI.
    await pumpUntilFound(tester, find.text(uploadFileLabel).hitTestable());
    for (int i = 0; i < 5; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }
    await tester.tap(find.text(uploadFileLabel).hitTestable());

    // The picker mock resolves immediately with one 1x1 PNG, and the sheet
    // opens already seeded with it -- so the QuickScanForm card mounts with its
    // three dropdowns (groupId, paidByUserId, status) without any further
    // interaction.
    await pumpUntilFound(tester, find.text('Group'));

    // Fill the per-image form. e2e-admin's quickScan user prefs are
    // null, so all three fields need to be set explicitly.
    await selectDropdown(tester, 'groupId', 'My Receipts');
    await selectDropdown(tester, 'paidByUserId', adminDisplayName(tester));
    await selectDropdown(tester, 'status', 'Open');

    // Drain the dropdown overlay teardown before tapping Submit.
    await tester.pumpAndSettle(const Duration(seconds: 3));

    await tester.tap(find.byType(BottomSubmitButton));

    // Success message text is from quick_scan.dart:179-182:
    // "Successfully queued $imageWord for processing!" where
    // imageWord is "image" for 1 file.
    await pumpUntilFound(
      tester,
      find.textContaining('Successfully queued'),
      timeout: const Duration(seconds: 15),
    );

    // No cleanup needed: Quick Scan queues an async backend job;
    // the resulting receipt is created by the worker and won't be
    // assigned a deterministic id we can DELETE inline. The unique
    // PNG bytes used by the mock are unlikely to produce a
    // recognizable extracted name -- if these accumulate on demo
    // they're identifiable as low-info OCR results from the test
    // window.
  });
}
