import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';

import 'api.dart';
import 'permission_fixtures.dart';
import 'pump.dart';
import 'receipt_test_helpers.dart';

/// Shared Quick Scan e2e actions, used by the config + submit specs.

/// Asserts [field] is genuinely on screen in the Quick Scan sheet.
///
/// None of the obvious assertions work here:
///  * `findsOneWidget` passes for a widget clipped outside its viewport;
///  * `Finder.hitTestable` reports 0 under `IntegrationTestWidgetsFlutterBinding`
///    on desktop even for plainly visible widgets;
///  * `tester.ensureVisible` and `Scrollable.ensureVisible` walk *every* ancestor
///    scrollable, so they drag the field into view even when the sheet is broken
///    -- precisely how this shipped (`quick_scan_submit_test` reaches its comment
///    that way).
///
/// So this scrolls [field]'s **own** scroll view by the least amount that would
/// bring it fully into the viewport, clamped to what that view can actually
/// scroll, and then checks geometry. One rule covers the whole column:
///
///  * a mid-column field (Categories, Tags) scrolls just far enough and stays
///    inside the viewport -- jumping to `maxScrollExtent` for those risked
///    pushing them off the *top*, since Comment sits below them;
///  * the last field needs more than `maxScrollExtent`, so the clamp leaves it
///    as far down as a user could ever get it. If it is still not fully inside
///    the window there, nothing the user does will reveal it -- which is the
///    original bug.
///
/// The viewport check additionally catches a field left buried under the pinned
/// submit button.
///
/// No surface size is forced: this runs at the target's natural viewport
/// (1280x720 for the Linux desktop runner, a real handset for Android/iOS).
/// `setSurfaceSize` is a no-op on the desktop target, so pinning a "phone" size
/// here would only be misleading.
Future<void> expectQuickScanFieldOnScreen(
  WidgetTester tester,
  Finder field, {
  required String label,
}) async {
  expect(field, findsOneWidget, reason: '$label is in the tree');

  final scrollable = Scrollable.of(tester.element(field));
  final position = scrollable.position;
  final viewportFinder = find.byWidget(scrollable.widget);

  final startRect = tester.getRect(field);
  final startViewport = tester.getRect(viewportFinder);
  final below = startRect.bottom - startViewport.bottom;
  final above = startViewport.top - startRect.top;
  final delta = below > 0 ? below : (above > 0 ? -above : 0.0);

  if (delta != 0.0) {
    position.jumpTo((position.pixels + delta)
        .clamp(position.minScrollExtent, position.maxScrollExtent)
        .toDouble());
    await tester.pump(const Duration(milliseconds: 200));
  }

  final fieldRect = tester.getRect(field);
  final window =
      Offset.zero & tester.view.physicalSize / tester.view.devicePixelRatio;
  final viewport = tester.getRect(viewportFinder);

  expect(window.contains(fieldRect.topLeft), isTrue,
      reason: '$label starts inside the window ($fieldRect vs $window)');
  expect(window.contains(fieldRect.bottomRight), isTrue,
      reason: '$label ends inside the window ($fieldRect vs $window)');
  expect(fieldRect.bottom <= viewport.bottom, isTrue,
      reason: '$label is not buried under the pinned submit button '
          '($fieldRect vs viewport $viewport)');
  expect(fieldRect.top >= viewport.top, isTrue,
      reason: '$label is not scrolled off the top of the sheet '
          '($fieldRect vs viewport $viewport)');
}

/// Finds a `FormBuilderDropdown` in the Quick Scan form by its `name`.
Finder quickScanDropdown(String name) =>
    find.byWidgetPredicate((w) => w is FormBuilderDropdown && w.name == name);

/// Finds the Quick Scan form's comment text field.
Finder quickScanCommentField() => find.byWidgetPredicate(
    (w) => w is FormBuilderTextField && w.name == 'comment');

/// Applies [configure] to the admin's first non-all group and returns the
/// configured group (its `name` selects it in the form's group dropdown).
///
/// **Persists through the API**, then mirrors the result into the live
/// `GroupModel`. It used to only do the latter, which no longer works: starting
/// a Quick Scan re-fetches AppData before the scanner opens
/// (`TokenRefreshService.reloadAppData`, called from `startScanEntry`), so a
/// purely client-side mutation is overwritten by the server's copy the moment
/// the sheet is opened. Persisting is also what production does -- the client
/// learns the config from AppData and the backend validates a submit against the
/// same rows -- so these specs now prove the round trip rather than the form's
/// reaction to an injected value.
///
/// The local mirror is kept so the sheet is still correct if the refresh fails
/// (it is swallowed by design) and so the config is in place before the first
/// refresh lands. `setGroupQuickScanConfig` restores the group's original
/// settings on teardown.
Future<api.Group> configureFirstGroup(
  WidgetTester tester,
  void Function(api.GroupBuilder) configure,
) async {
  final ctx = tester.element(find.byType(Scaffold).first);
  final groupModel = Provider.of<GroupModel>(ctx, listen: false);
  final target = groupModel.groups.firstWhere((g) => !g.isAllGroup);
  final configured = target.rebuild(configure);
  final settings = configured.groupReceiptSettings;

  // Every key the command carries is sent, read back off the rebuilt group, so
  // the fields [configure] did not touch keep the values they already had rather
  // than being reset to false by the command's defaults.
  await setGroupQuickScanConfig(
    groupId: target.id,
    jwt: await apiLogin(),
    overrides: {
      'hideComments': settings.hideComments ?? false,
      'quickScanPaidByEnabled': settings.quickScanPaidByEnabled ?? false,
      'quickScanPaidByRequired': settings.quickScanPaidByRequired ?? false,
      'quickScanStatusEnabled': settings.quickScanStatusEnabled ?? false,
      'quickScanStatusRequired': settings.quickScanStatusRequired ?? false,
      'quickScanCategoriesEnabled': settings.quickScanCategoriesEnabled ?? false,
      'quickScanCategoriesRequired':
          settings.quickScanCategoriesRequired ?? false,
      'quickScanTagsEnabled': settings.quickScanTagsEnabled ?? false,
      'quickScanTagsRequired': settings.quickScanTagsRequired ?? false,
      'quickScanCommentEnabled': settings.quickScanCommentEnabled ?? false,
      'quickScanCommentRequired': settings.quickScanCommentRequired ?? false,
      // The backend rejects a config where paid-by or status can be skipped
      // without a default to backfill it -- a receipt always has both
      // (UpdateGroupReceiptSettingsCommand.Validate). Most specs here hide or
      // relax one of the two, so send defaults unconditionally: UPLOADER
      // resolves at scan time and needs no user id, and OPEN is a valid status.
      // Harmless when the field is shown+required, since then it is unused.
      'quickScanDefaultPaidByType': 'UPLOADER',
      'quickScanDefaultStatus': 'OPEN',
    },
  );

  groupModel.setGroups(groupModel.groups
      .map((g) => g.id == target.id ? configured : g)
      .toList());
  await tester.pump();
  return configured;
}

/// Trims the live GroupModel to the all-group placeholder(s) plus the group with
/// [groupId] (returned), so the group dropdown is short and the target is on
/// screen regardless of how many groups the admin has accumulated. Presentation
/// only -- it does not touch the group's (persisted) config.
Future<api.Group> keepOnlyGroup(WidgetTester tester, int groupId) async {
  final ctx = tester.element(find.byType(Scaffold).first);
  final groupModel = Provider.of<GroupModel>(ctx, listen: false);
  final target = groupModel.groups.firstWhere((g) => g.id == groupId);
  final allGroups = groupModel.groups.where((g) => g.isAllGroup).toList();
  groupModel.setGroups([...allGroups, target]);
  await tester.pump();
  return target;
}

/// Taps the bottom-nav scan slot to capture one image via the mocked document
/// scanner, leaving the Quick Scan sheet open with its per-image form mounted
/// (the "Group" field is on screen).
///
/// The tap is the direct scan action, so the sheet arrives already seeded --
/// there is no menu step and no separate "add a photo" tap any more. Requires
/// the AI flag already enabled and `installDocumentScannerMock()` already
/// installed. The hitTestable + frame-drain hardening handles the sheet
/// slide-in (a deterministic tap-flake on iOS Cupertino transitions).
Future<void> openQuickScanImageForm(WidgetTester tester) async {
  await pumpUntilFound(tester, scanNavSlot());
  await tester.tap(scanNavSlot());
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }

  await pumpUntilFound(tester, find.text('Group'));
}

/// Taps the form's submit button and waits for the fix-error snackbar (a
/// required+empty field short-circuits the submit before any API call).
Future<void> expectQuickScanSubmitBlocked(WidgetTester tester) async {
  await pumpUntilFound(tester, find.byType(BottomSubmitButton).hitTestable());
  await tester.tap(find.byType(BottomSubmitButton));
  await pumpUntilFound(
    tester,
    find.textContaining('Please fix error on quick scan'),
    timeout: const Duration(seconds: 10),
  );
}

/// Taps the form's submit button and waits for the success snackbar (the scan
/// was accepted and queued server-side), then asserts the sheet confirms it
/// inline.
///
/// The snackbar alone is not enough: submitting disables every field and hides
/// the submit button, so once it fades the sheet would sit greyed out with
/// nothing saying why. This is the only positive assertion on
/// `quick-scan-queued-confirmation` -- the widget suite can only assert its
/// absence, since the sheet builds its `isCompletedSubject` internally and a
/// real submit is what sets it.
Future<void> expectQuickScanQueued(WidgetTester tester) async {
  await pumpUntilFound(tester, find.byType(BottomSubmitButton).hitTestable());
  await tester.tap(find.byType(BottomSubmitButton));
  await pumpUntilFound(
    tester,
    find.textContaining('Successfully queued'),
    timeout: const Duration(seconds: 15),
  );

  await pumpUntilFound(
    tester,
    find.byKey(const ValueKey('quick-scan-queued-confirmation')),
    timeout: const Duration(seconds: 10),
  );
  expect(find.text(quickScanQueuedMessage), findsOneWidget);
}
