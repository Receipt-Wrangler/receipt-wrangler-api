import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/receipt_edit_popup_menu.dart';

import 'helpers/api.dart';
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

  testWidgets(
      'admin can add a receipt with an image picked from files',
      (tester) async {
    await installFileSelectorMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    final receiptName =
        'e2e-image-${DateTime.now().millisecondsSinceEpoch}';

    await openManualReceiptForm(tester);
    await pumpUntilFound(tester, find.text('Name'));

    // Open the Images sub-screen via the "Images" compact action button
    // in the receipt form's Details header. Target the Tooltip wrapper
    // (mobile/lib/receipts/widgets/receipt_form.dart:411) by its message
    // for explicit semantics (the Text "Images" is just the label inside
    // the InkWell -- tapping the Tooltip taps the same widget but says
    // out loud which button we mean).
    final imagesButton = find.byTooltip('View Images');
    await tester.ensureVisible(imagesButton);
    await tester.pump();
    await tester.tap(imagesButton);
    // Drain the iOS Cupertino page-transition slide-in (~400ms). Without
    // this, pumpUntilFound returns the moment the new route mounts, but
    // the PopupMenuButton is still translated off-screen (x ~= 1836 on a
    // 1280-wide surface) and tap() derives an offset outside the render
    // tree -- the tap silently misses and the popup never opens. Same
    // shape as the dropdown-overlay drain in form_actions.dart:54-56;
    // finite, deterministic, avoids pumpAndSettle (which would hang on
    // the new screen's image-fetch StreamBuilder).
    for (int i = 0; i < 6; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }
    // Find the popup menu by widget type, not by icon. PopupMenuButton's
    // default icon is `Icons.adaptive.more` (Flutter material/popup_menu.dart),
    // which is `Icons.more_vert` on Android/desktop and `Icons.more_horiz`
    // on iOS -- byIcon(more_vert) never matches on iOS.
    await pumpUntilFound(tester, find.byType(PopupMenuButton));

    // Open the image-screen popup menu and pick the file source.
    await tester.tap(find.byType(PopupMenuButton));
    await pumpUntilFound(tester, find.text(uploadFileLabel));
    await tester.tap(find.text(uploadFileLabel));
    // Mocked openFiles() resolves immediately; the model's
    // imagesToUploadBehaviorSubject emits, the carousel updates.
    await tester.pumpAndSettle(const Duration(seconds: 2));

    // Back to the form.
    await tester.tap(find.byIcon(Icons.arrow_back));
    await pumpUntilFound(tester, find.text('Name'));

    await tester.enterText(formField('name'), receiptName);
    await tester.enterText(formField('amount'), '12.34');
    await selectDropdown(tester, 'groupId', 'My Receipts');
    await selectDropdown(tester, 'paidByUserId', adminDisplayName(tester));

    await tester.pumpAndSettle(const Duration(seconds: 3));
    await tester.tap(find.byType(BottomSubmitButton));
    // /view shell mounted -> ReceiptEditPopupMenu is in the tree.
    await pumpUntilFound(tester, find.byType(ReceiptEditPopupMenu));
    final receiptId = receiptIdFromUrl(currentUrl(tester));

    // Verify server-side that the image upload sequence completed.
    // Bug-fix #2 means a partial-upload failure would NOT navigate to
    // /view, so reaching this assertion at all is half the verification;
    // the imageFiles count is the other half.
    final jwt = await apiLogin();
    final receipt = await getReceipt(receiptId, jwt: jwt);
    final imageFiles = receipt['imageFiles'] as List?;
    expect(imageFiles?.length ?? 0, 1,
        reason: 'Receipt should have exactly 1 imageFile after save');

    scheduleReceiptCleanup(receiptId);
  });
}
