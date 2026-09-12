import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/receipt_edit_popup_menu.dart';

import 'api.dart';
import 'form_actions.dart';
import 'pump.dart';
import 'users.dart';

/// Reads the current GoRouter URL by grabbing a context from inside the
/// routed tree (`MaterialApp` itself sits above the GoRouter scope, so
/// its element fails the inherited-widget lookup with "No GoRouter
/// found in context").
String currentUrl(WidgetTester tester) {
  final scaffold = find.byType(Scaffold).evaluate();
  if (scaffold.isEmpty) return '';
  return GoRouter.of(scaffold.first)
      .routerDelegate
      .currentConfiguration
      .uri
      .toString();
}

/// Pumps until the current GoRouter URL matches [pattern], or [timeout]
/// elapses.
///
/// PREFER `pumpUntilFound(tester, find.byType(SomeDestinationWidget))`
/// for asserting that a route has actually mounted -- a widget is the
/// stronger signal that the shell finished building (per `mobile/CLAUDE.md`
/// "Assert navigation by widget presence, not URL"). Use this URL helper
/// only when you also need to *extract* something from the URL after
/// arrival (e.g. the receipt id from `/receipts/<id>/view`); pair it with
/// a widget-presence wait first so the URL read isn't a race.
Future<String> pumpUntilUrl(
  WidgetTester tester,
  RegExp pattern, {
  Duration timeout = const Duration(seconds: 15),
}) async {
  final deadline = DateTime.now().add(timeout);
  while (DateTime.now().isBefore(deadline)) {
    await tester.pump(const Duration(milliseconds: 100));
    final url = currentUrl(tester);
    if (pattern.hasMatch(url)) return url;
  }
  throw StateError(
    'Timed out after ${timeout.inSeconds}s waiting for URL matching '
    '$pattern. Last seen: ${currentUrl(tester)}',
  );
}

/// Extracts the receipt id from a `/receipts/<id>/view` URL produced by
/// the production save handler.
int receiptIdFromUrl(String url) {
  final match = RegExp(r'/receipts/(\d+)/view').firstMatch(url);
  if (match == null) {
    throw StateError('Expected /receipts/<id>/view URL, got: $url');
  }
  return int.parse(match.group(1)!);
}

/// Best-effort cleanup: log in via the API and DELETE the receipt at
/// end-of-test. Wrapped via [addTearDown] so it runs even if the test
/// body throws after the receipt was created.
void scheduleReceiptCleanup(int receiptId) {
  addTearDown(() async {
    final jwt = await apiLogin();
    await deleteReceipt(receiptId, jwt: jwt);
  });
}

/// The bottom-nav scan/add slot, whichever label it is currently wearing.
///
/// It reads "Scan" when Quick Scan can run for the caller and "Add" when it
/// can't, so a spec that only cares about *reaching* the entry point must not
/// pin the label. Inside a group two bottom navs are mounted (the group-select
/// shell sits under the group shell), so two slots match; `.hitTestable()`
/// targets the visible one.
Finder scanNavSlot() => find
    .byWidgetPredicate((w) =>
        w is NavigationDestination && (w.label == 'Scan' || w.label == 'Add'))
    .hitTestable();

/// Opens the manual receipt form from the bottom-nav scan/add slot.
///
/// A **tap** on that slot is a direct action now -- it opens the document
/// scanner when Quick Scan can run, and the manual form when it can't -- so
/// manual entry is reached by **holding** it. The hold works on every screen and
/// in every feature-flag state, unlike the receipts-screen overflow menu, which
/// is why it is the path used here.
///
/// Pre-condition: the caller holds `group.receipts.create`, so the slot is
/// present and the menu carries the entry.
Future<void> openManualReceiptForm(WidgetTester tester) async {
  await pumpUntilFound(tester, scanNavSlot());
  await tester.longPress(scanNavSlot());

  // The menu's items mount on the popup's first frame while it is still
  // growing, so a tap computed then lands short and misses (observed as
  // "Offset(411.6, 852.0) ... would not hit test"). Wait for hittability, then
  // drain the animation.
  await pumpUntilFound(tester, find.text('Add Manual Receipt').hitTestable());
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
  await tester.tap(find.text('Add Manual Receipt').hitTestable());
  await pumpUntilFound(tester, find.text('Name'));
}

/// Drives the receipt-add UI from `/groups`: opens the bottom-nav Add
/// menu, fills the required fields, taps Submit, waits for navigation
/// to `/receipts/<id>/view`. Returns the new receipt's id.
///
/// Used by tests that need a baseline receipt to operate on (Flow #4
/// edits it, Flow #5 chains two of these). The same field-fill sequence
/// as Flow #1's smoke happy path -- if Flow #1 is green, this is too.
///
/// Pre-conditions: caller is logged in and currently on `/groups`.
Future<int> addManualReceiptViaUI(
  WidgetTester tester,
  String name, {
  String amount = '12.34',
  String groupName = 'My Receipts',
}) async {
  await openManualReceiptForm(tester);

  await tester.enterText(formField('name'), name);
  await tester.enterText(formField('amount'), amount);
  await selectDropdown(tester, 'groupId', groupName);
  await selectDropdown(tester, 'paidByUserId', adminDisplayName(tester));

  // Drain the dropdown overlay teardown -- the popup-route's overlay
  // entry can otherwise leave the Scaffold's bottom-sheet area in an
  // Offstage state and the BottomSubmitButton tap silently misses.
  await tester.pumpAndSettle(const Duration(seconds: 3));

  return submitManualReceiptForm(tester);
}

/// Submits a filled receipt form and returns the new receipt's id.
///
/// Split out of [addManualReceiptViaUI] for specs that fill the form
/// differently -- notably the ones asserting a field was pre-seeded, which must
/// not touch it.
Future<int> submitManualReceiptForm(WidgetTester tester) async {
  await tester.tap(find.byType(BottomSubmitButton));
  // Assert /view shell has mounted via the ReceiptEditPopupMenu, which
  // only renders on /view (gated on canEditReceipt -- see
  // mobile/lib/shared/widgets/receipt_edit_popup_menu.dart:31). Then
  // extract the id from the URL.
  await pumpUntilFound(tester, find.byType(ReceiptEditPopupMenu));
  return receiptIdFromUrl(currentUrl(tester));
}
