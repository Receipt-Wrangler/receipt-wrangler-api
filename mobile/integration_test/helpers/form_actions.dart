import 'package:flutter/cupertino.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';

import 'pump.dart';

/// Locates a `FormBuilderTextField` (or `FormBuilderField` subclass) by its
/// `name`. The receipt-add form uses `name`-based identity rather than
/// `Key`, so `find.byKey` doesn't apply.
Finder formField(String name) => find.byWidgetPredicate(
      (w) => w is FormBuilderTextField && w.name == name,
    );

/// Locates a `CupertinoButton` (filled or otherwise) by its visible label.
/// Used for the homeserver/login screens where the action buttons are
/// `CupertinoButton.filled` with a `Text` child.
Finder filledButton(String text) =>
    find.widgetWithText(CupertinoButton, text);

/// Drives a `FormBuilderDropdown<T>` by tapping it open and then tapping
/// the menu item whose visible text matches [optionText].
///
/// Note: when the menu is open the option text appears in TWO places --
/// the closed-state child (still in the tree behind the menu) and the
/// menu item itself. `find.text(optionText).last` picks the menu's copy.
/// If two copies aren't enough disambiguation (e.g. the same text
/// renders elsewhere on screen), scope the call site with a `find.descendant`
/// rather than expanding this helper.
Future<void> selectDropdown(
  WidgetTester tester,
  String name,
  String optionText,
) async {
  await tester.tap(find.byWidgetPredicate(
    (w) => w is FormBuilderDropdown && w.name == name,
  ));
  // Pass the bare text finder, NOT `.last`. _LastFinderMixin.filter does
  // `yield input.last`, which throws StateError("No element") if the parent
  // is empty -- which it briefly is on slow targets (Android emulator) while
  // the dropdown menu is still opening. Wait for any match, then tap .last.
  await pumpUntilFound(tester, find.text(optionText));
  // That wait is only a lower bound now that fields can arrive pre-selected:
  // a dropdown already showing this option matches in its CLOSED state, so the
  // wait can return before the menu is built and `.last` would resolve to the
  // closed-state child. Draining the open animation closes that race.
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
  // Long menus (e.g. a groupId dropdown with several fixture groups) can lay
  // the target item out below the menu's viewport. ensureVisible jumps the
  // menu's scroll position but does NOT relayout -- pump a frame so the tap
  // computes the post-scroll center instead of the stale off-screen one.
  await tester.ensureVisible(find.text(optionText).last);
  await tester.pump(const Duration(milliseconds: 100));
  await tester.tap(find.text(optionText).last);
  // Drain frames in a loop -- a single pump only advances one frame, but
  // a dropdown selection needs at least three: the tap dispatch +
  // FormBuilderField.didChange (which fires the user-supplied onChanged),
  // the setState-triggered rebuild, and the menu's exit animation. A
  // single pump(800ms) advances time but only renders one frame, leaving
  // downstream widgets that depend on the user-onChanged setState (e.g.,
  // receipt_form.dart's `_ReceiptForm.groupId` State field driving the
  // Add Share button's disabled flag) unfrozen at their pre-selection
  // values on slower targets. We avoid pumpAndSettle outright because
  // the receipt form's CircularLoadingProgress can spin indefinitely
  // when customFieldModel reloads in the background.
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 200));
  }
}
