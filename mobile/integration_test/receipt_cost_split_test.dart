import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/receipt_edit_popup_menu.dart';

import 'helpers/api.dart';
import 'helpers/env.dart';
import 'helpers/login.dart';
import 'helpers/permission_fixtures.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/receipt_test_helpers.dart';
import 'helpers/users.dart';

/// Exercises `ReceiptQuickActions` from the edit form: open the Quick
/// Actions bottom sheet, select two users, drive the "Split Evenly"
/// mode, save the receipt, then verify the API has two items each
/// charged to one of the users at half the receipt total. Same shape
/// for the "By Percentage" mode in the second testWidgets case (75/25).
///
/// We use two separate `testWidgets` (and two separate receipts) so the
/// modes don't share item state -- `splitEvenly` appends to existing
/// items rather than replacing them (quick_actions_submit_button.dart:44),
/// so reusing one receipt for both modes would conflate the two cases'
/// outputs.
///
/// Both cases assert against `Receipt.receiptItems` returned by the API
/// -- the form-local `FormItem` list is only the source of truth until
/// save, after which the server takes over.
///
/// The receipt must live in a group BOTH users belong to: the Quick
/// Actions user picker lists only the receipt's group members
/// (`getUsersInGroup`, lib/utils/users.dart), and `e2e-user` is not a
/// member of admin's personal "My Receipts" group. Each case provisions
/// a fresh shared group (admin auto-owner + e2e-user as Legacy Editor)
/// and creates the receipt there.
void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  // Quick Actions opens via `openQuickActionsBottomSheet` (receipt_form.dart).
  // That used to crash because it passed a cached, since-deactivated
  // `shellContext` to `showModalBottomSheet` (null-check on a defunct
  // element). The form now re-reads the shell context at tap time and falls
  // back to its own mounted context, so the sheet opens reliably.
  testWidgets('Split Evenly creates one item per selected user',
      (tester) async {
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    final groupName = await _provisionSharedSplitGroup();
    await loginAsAdmin(tester);

    final receiptName = 'e2e-split-even-${DateTime.now().millisecondsSinceEpoch}';
    final receiptId = await addManualReceiptViaUI(
      tester,
      receiptName,
      amount: '100.00',
      groupName: groupName,
    );
    scheduleReceiptCleanup(receiptId);

    await _navigateToEdit(tester);
    await _openQuickActionsSheet(tester);

    // "Split Evenly" is index 0 of the ToggleButtons and is preselected
    // (quick_actions.dart:43 `quickActionsSelection = [true, false, false]`).
    // No mode toggle needed -- just pick the users and submit.
    await _selectUsers(tester, [
      adminDisplayName(tester),
      userDisplayName(tester),
    ]);

    // The total widget updates with "2 users × $50.00 each" once both
    // users are in the form. Wait for it so a slow recompute can't make
    // the Split tap fire against stale fields.
    await pumpUntilFound(tester, find.textContaining('2 users'));

    await _tapSplitAndSave(tester);

    final jwt = await apiLogin();
    final receipt = await getReceipt(receiptId, jwt: jwt);
    final items =
        ((receipt['receiptItems'] as List?) ?? const []).cast<Map>();
    expect(items.length, 2,
        reason: 'Split Evenly with 2 users should produce 2 receipt items; '
            'fewer suggests buildEvenSplitFormItems ran with the wrong '
            'getSelectedUsers() snapshot');
    final amounts = items.map((i) => _toDouble(i['amount'])).toList();
    for (final a in amounts) {
      expect(a, closeTo(50.0, 0.01),
          reason: 'Each Split Evenly item should be receiptTotal / 2 = 50.00. '
              'Off-by-cents would indicate Money2 rounding regressed.');
    }
    expect(
        items.map((i) => i['chargedToUserId']).toSet().length,
        2,
        reason: 'Each item should be charged to a distinct user');
  });

  testWidgets('By Percentage creates items proportional to picked %',
      (tester) async {
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    final groupName = await _provisionSharedSplitGroup();
    await loginAsAdmin(tester);

    final receiptName = 'e2e-split-pct-${DateTime.now().millisecondsSinceEpoch}';
    final receiptId = await addManualReceiptViaUI(
      tester,
      receiptName,
      amount: '100.00',
      groupName: groupName,
    );
    scheduleReceiptCleanup(receiptId);

    await _navigateToEdit(tester);
    await _openQuickActionsSheet(tester);

    // Switch to "By Percentage" (toggle index 2). The labels live in
    // `quickActions` in quick_actions.dart:38-42.
    await tester.tap(find.text('By Percentage').hitTestable());
    await tester.pumpAndSettle();

    await _selectUsers(tester, [
      adminDisplayName(tester),
      userDisplayName(tester),
    ]);

    // Wait for the per-user FilterChip rows to appear -- buildPercentageFields()
    // only renders them once `users` is non-empty.
    await pumpUntilFound(tester, find.widgetWithText(FilterChip, '75%'));

    // Pick 75% for admin, 25% for the e2e user. FilterChip labels for
    // each user are rendered identically ("25%", "50%", "75%", "100%",
    // "Custom"), so we tap the .first instance for admin and .last for
    // the second user. The rows are rendered in users-array order which
    // matches our selection order.
    await tester.tap(find.widgetWithText(FilterChip, '75%').first);
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilterChip, '25%').last);
    await tester.pumpAndSettle();

    await _tapSplitAndSave(tester);

    final jwt = await apiLogin();
    final receipt = await getReceipt(receiptId, jwt: jwt);
    final items =
        ((receipt['receiptItems'] as List?) ?? const []).cast<Map>();
    expect(items.length, 2,
        reason: 'By Percentage with 2 users should produce 2 items');
    final amounts = items.map((i) => _toDouble(i['amount'])).toList()
      ..sort();
    expect(amounts[0], closeTo(25.0, 0.01),
        reason: r'Lower portion should be 25% of $100 = $25.00');
    expect(amounts[1], closeTo(75.0, 0.01),
        reason: r'Higher portion should be 75% of $100 = $75.00');
  });
}

double _toDouble(dynamic v) {
  if (v is num) return v.toDouble();
  return double.parse(v.toString());
}

/// Creates a fresh group with admin as auto-Owner and `e2e-user` as Legacy
/// Editor, so the Quick Actions user picker (group-members-only) offers both
/// display names. Registers an `addTearDown` that deletes the group (which
/// cascades its receipts). Returns the group's name for the groupId dropdown.
Future<String> _provisionSharedSplitGroup() async {
  final jwt = await apiLogin(); // admin
  final userId = await userIdByUsername(E2eEnv.userUsername, jwt: jwt);
  final editorRoleId = await groupRoleIdByName('Legacy Editor', jwt: jwt);
  final groupName =
      'e2e-split-grp-${DateTime.now().millisecondsSinceEpoch}';
  final groupId = await createGroupWithMember(
    name: groupName,
    memberUserId: userId,
    groupRoleId: editorRoleId,
    jwt: jwt,
  );
  addTearDown(() async => deleteGroup(groupId, jwt: await apiLogin()));
  return groupName;
}

Future<void> _navigateToEdit(WidgetTester tester) async {
  // The ReceiptEditPopupMenu is gated on canEditReceipt(); on cold-boot
  // after the /view navigation, PermissionsModel may not yet have the user's
  // group.receipts.update permission, so the button isn't mounted immediately
  // (see receipt_edit_test.dart:50 for the same pattern).
  final menuButton = find.byType(PopupMenuButton<dynamic>);
  await pumpUntilFound(tester, menuButton);
  await tester.tap(menuButton);
  // The popup scales in; the "Edit" Text mounts on the animation's first
  // frame at a transform where a tap computed from its center misses
  // (observed flake on the emulator). Wait until it's actually hittable,
  // then drain the open animation so the center is its settled position.
  await pumpUntilFound(tester, find.text('Edit').hitTestable());
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
  await tester.tap(find.text('Edit').hitTestable());
  // /edit's destination-mounted marker: the bottom save button, which only
  // renders on edit/add paths (receipt_bottom_sheet_builder.dart,
  // isEditingBasedOnFullPath). find.text('Name') is NOT a marker -- the
  // /view form renders the same label, so it matches before navigation.
  await pumpUntilFound(tester, find.byType(BottomSubmitButton));
}

/// Locates the split-action IconButton on the edit form. It's the
/// IconButton whose `icon` is an `SvgPicture` for `assets/icons/split.svg`
/// (receipt_form.dart:482-489). The neighboring "Add Share" IconButton
/// uses a Material `Icon`, not `SvgPicture`, so the predicate is unique.
Future<void> _openQuickActionsSheet(WidgetTester tester) async {
  final splitButton = find.byWidgetPredicate(
    (w) => w is IconButton && w.icon is SvgPicture,
  );
  await pumpUntilFound(tester, splitButton);
  // The split-action button sits below the Shares row on the form,
  // off-screen on the 1280x900 test surface. Scroll it into view so
  // the tap lands -- otherwise the bottom sheet never opens.
  await tester.ensureVisible(splitButton);
  await tester.pumpAndSettle();
  await tester.tap(splitButton);
  // The fullscreen bottom sheet header is "Quick Actions"; wait for the
  // ToggleButtons row to render before driving the form. The text mounts on
  // the slide-in's first frame while the sheet is still rising -- a tap
  // computed then lands near the bottom edge and misses (observed as
  // "Offset(408.4, 852.0) ... would not hit test" on the By Percentage
  // toggle). Wait for hittability, then drain the slide-in.
  await pumpUntilFound(tester, find.text('Split Evenly').hitTestable());
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
}

/// Opens the "Users" MultiSelectField, taps each ChoiceChip whose label
/// matches a display name in [displayNames], then taps the "Select"
/// confirm button. The MultiSelectField wraps its InputDecorator in an
/// opaque GestureDetector(onTap:) that fires `showUserMultiSelect`, so any
/// point inside the labeled "Users" field hits it.
Future<void> _selectUsers(
  WidgetTester tester,
  List<String> displayNames,
) async {
  // The whole field is one tap target, so the decorator's center would work
  // too. The placeholder text is kept as the tap site because it is the
  // unambiguous locator for this field on a form full of decorators.
  await tester.tap(find.text('No Users selected'));
  await pumpUntilFound(tester, find.text('Select Users'));

  for (final name in displayNames) {
    // Wait for each chip -- the sheet's slide-in means the chips can be
    // mounted but mid-animation on the first loop iteration.
    final chip = find.widgetWithText(ChoiceChip, name);
    await pumpUntilFound(tester, chip);
    await tester.tap(chip);
    await tester.pump(const Duration(milliseconds: 200));
  }

  // The receipt-save success snackbar (root ScaffoldMessenger) renders ABOVE
  // the modal sheet and covers the bottom row where the Select button sits,
  // absorbing taps for its ~4s lifetime -- observed as "Offset(640.0, 875.0)
  // ... would not hit test on the specified widget". hitTestable() polls
  // until the snackbar departs so the tap actually lands.
  final selectButton = find.widgetWithText(BottomSubmitButton, 'Select');
  await pumpUntilFound(tester, selectButton.hitTestable());
  await tester.tap(selectButton.hitTestable());
  await tester.pumpAndSettle();
}

/// Submits the "Split" form, returns to the receipt edit screen, then
/// submits the receipt itself. Lands on `/view` so the caller can poll
/// the API.
Future<void> _tapSplitAndSave(WidgetTester tester) async {
  // Same snackbar-over-bottom-button hazard as _selectUsers' Select tap.
  final splitButton = find.widgetWithText(BottomSubmitButton, 'Split');
  await pumpUntilFound(tester, splitButton.hitTestable());
  await tester.tap(splitButton.hitTestable());
  // The bottom sheet pops; the edit form's Name field is the
  // destination-mounted marker for /edit being the visible route again.
  await pumpUntilFound(tester, find.text('Name'));

  // Drain frames so the outer BottomSubmitButton has the new items list
  // committed to the form before we tap save.
  await tester.pumpAndSettle(const Duration(seconds: 2));

  final saveButton = find.byType(BottomSubmitButton);
  await pumpUntilFound(tester, saveButton.hitTestable());
  await tester.tap(saveButton.hitTestable());
  // /view shell mounted -> ReceiptEditPopupMenu is in the tree.
  await pumpUntilFound(tester, find.byType(ReceiptEditPopupMenu));
}
