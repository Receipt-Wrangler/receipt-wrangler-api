// Verifies that the receipt-entry affordances hide Quick Scan when
// authModel.featureConfig.aiPoweredReceipts is false, and that the slot itself
// changes what it promises.
//
// Regression for the gating: the Quick Scan entry used to render regardless of
// the AI feature flag, with only the in-action sheet enforcing it (by firing an
// error snackbar instead of opening). The gate now lives in
// resolveReceiptEntryAvailability, which every affordance reads -- so the slot's
// own label and icon change too, rather than promising a scanner the install
// cannot run.
//
// Strategy: log in normally, then mutate AuthModel.featureConfig directly via
// Provider. That is closer to the real failure mode than mocking /featureConfig,
// because the gate reads featureConfig synchronously at build/tap time -- which
// IS the decision point. After login nothing else writes featureConfig until the
// next storeAppData call, so the mutation sticks.

import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';

import 'helpers/login.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/receipt_test_helpers.dart';

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  testWidgets('the scan entry hides Quick Scan when aiPoweredReceipts is false',
      (tester) async {
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    final ctx = tester.element(find.byType(Scaffold).first);
    final authModel = Provider.of<AuthModel>(ctx, listen: false);
    final originalEnableLocalSignUp =
        authModel.featureConfig.enableLocalSignUp;

    void setAi(bool enabled) {
      authModel.setFeatureConfig(
        (api.FeatureConfigBuilder()
              ..aiPoweredReceipts = enabled
              ..enableLocalSignUp = originalEnableLocalSignUp)
            .build(),
      );
    }

    Future<void> openEntryMenu() async {
      await pumpUntilFound(tester, scanNavSlot());
      await tester.longPress(scanNavSlot());
      for (int i = 0; i < 5; i++) {
        await tester.pump(const Duration(milliseconds: 100));
      }
      await pumpUntilFound(tester, find.text(addManualReceiptLabel));
    }

    setAi(false);
    await tester.pump();

    // Only the MENU is asserted here, not the slot's label. The gates are read
    // with `listen: false` (the house pattern -- permissions and featureConfig
    // hydrate before first paint and don't change during a session), so the
    // label is fixed at nav build time while the menu is built fresh on each
    // open. Flipping the flag live therefore moves the menu but not the label.
    // The label is covered where the flag is real: permission_add_menu_test and
    // quick_scan_entry_gated_test set it server-side before login.
    await openEntryMenu();
    expect(find.text(quickScanLabel), findsNothing,
        reason: 'Quick Scan must not appear when aiPoweredReceipts=false');
    expect(find.text(uploadPhotoLabel), findsNothing,
        reason: 'both picker entries feed Quick Scan, so they go with it');
    expect(find.text(uploadFileLabel), findsNothing,
        reason: 'both picker entries feed Quick Scan, so they go with it');

    // Dismiss the popup before reopening with the flag flipped on. Waiting for
    // the items to actually leave -- rather than pumping a fixed duration -- is
    // what makes the second long-press land on the nav rather than on the
    // dismissing route's barrier.
    await tester.tapAt(const Offset(10, 10));
    await pumpUntilGone(tester, find.text(addManualReceiptLabel));
    await tester.pumpAndSettle();

    setAi(true);
    await tester.pump();

    await openEntryMenu();
    expect(find.text(quickScanLabel), findsOneWidget,
        reason: 'Quick Scan must reappear when aiPoweredReceipts=true');
    expect(find.text(uploadPhotoLabel), findsOneWidget,
        reason: 'and both picker entries come back with it');
    expect(find.text(uploadFileLabel), findsOneWidget,
        reason: 'and both picker entries come back with it');
  });
}
