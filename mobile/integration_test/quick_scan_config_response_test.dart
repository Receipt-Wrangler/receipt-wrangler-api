// Verifies the Quick Scan per-image FORM responds to the selected group's
// quick-scan configuration (GroupReceiptSettings.quickScan*) -- no scan is sent
// (each test either asserts visibility or that a required+empty field blocks
// submit). The submit-succeeds paths live in quick_scan_submit_test.dart.
//
// Runs headless on Linux desktop. Two facts shape the approach:
//   * Quick Scan is gated on featureConfig.aiPoweredReceipts (off locally), so
//     we flip it on server-side before login (enableAiPoweredReceiptsForTest).
//   * The sheet's gallery-upload icon throws "Unsupported platform" on desktop,
//     so we feed an image through the document-scanner icon via a mocked
//     `cunning_document_scanner` channel (installDocumentScannerMock), which
//     also grants the camera permission getPictures requests.
//
// The group config is injected the same way quick_scan_disabled_test injects the
// feature flag: a Provider mutation on the live GroupModel. This is deterministic
// and independent of whether the local API persists the new fields.

import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/shared/functions/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/category_select_field.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/tag_select_field.dart';

import 'helpers/document_scanner_mock.dart';
import 'helpers/feature_flags.dart';
import 'helpers/form_actions.dart';
import 'helpers/login.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/quick_scan_actions.dart';

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  testWidgets('quick scan form shows/hides + requires fields per group config',
      (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    // Paid By hidden, Status shown+optional, Categories shown+required, Tags hidden.
    final group = await configureFirstGroup(tester, (b) => b
      ..groupReceiptSettings.quickScanPaidByEnabled = false
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = true
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = true
      ..groupReceiptSettings.quickScanCategoriesRequired = true
      ..groupReceiptSettings.quickScanTagsEnabled = false
      ..groupReceiptSettings.quickScanTagsRequired = false);

    await openQuickScanImageForm(tester);

    // Select the configured group so the form reads its config.
    await selectDropdown(tester, 'groupId', group.name);

    // Fields now reflect the injected config.
    expect(quickScanDropdown('paidByUserId'), findsNothing,
        reason: 'paid-by hidden');
    expect(quickScanDropdown('status'), findsOneWidget, reason: 'status shown');
    expect(find.byType(CategorySelectField), findsOneWidget,
        reason: 'categories shown');
    expect(find.byType(TagSelectField), findsNothing, reason: 'tags hidden');

    // Categories is required and empty -> submit is blocked with the fix-error
    // snackbar and queues nothing.
    await expectQuickScanSubmitBlocked(tester);
  });

  testWidgets('quick scan form re-reads config when the group changes',
      (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    // Two groups with opposite configs. Group 1 is the admin's real group
    // rebuilt with config A; group 2 is a clone with a fresh id + name and the
    // inverse config B. Injected via the live GroupModel (deterministic, no
    // reliance on the API persisting the new fields).
    // This spec fabricates a second group that does not exist on the server, so
    // it cannot persist its fixture the way the others do -- and starting a scan
    // now re-fetches AppData, which would wipe both the synthetic group and the
    // trimmed list below. Skip that refresh: what is under test here is the form
    // re-reading config when the dropdown changes, not the round trip.
    debugSkipQuickScanAppDataRefresh = true;
    addTearDown(() => debugSkipQuickScanAppDataRefresh = false);

    final ctx = tester.element(find.byType(Scaffold).first);
    final groupModel = Provider.of<GroupModel>(ctx, listen: false);
    final real = groupModel.groups.firstWhere((g) => !g.isAllGroup);
    final group2Id =
        groupModel.groups.fold<int>(0, (m, g) => g.id > m ? g.id : m) + 1;

    // Config A: paid-by shown+optional, status hidden, categories shown+required,
    // tags hidden.
    final group1 = real.rebuild((b) => b
      ..groupReceiptSettings.quickScanPaidByEnabled = true
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = false
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = true
      ..groupReceiptSettings.quickScanCategoriesRequired = true
      ..groupReceiptSettings.quickScanTagsEnabled = false
      ..groupReceiptSettings.quickScanTagsRequired = false);

    // Config B: paid-by hidden, status shown+required, categories hidden,
    // tags shown+optional. Distinct id/name so it's a separate dropdown entry.
    final group2 = real.rebuild((b) => b
      ..id = group2Id
      ..name = 'E2E Switch Group B'
      ..groupReceiptSettings.id = group2Id
      ..groupReceiptSettings.groupId = group2Id
      ..groupReceiptSettings.quickScanPaidByEnabled = false
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = true
      ..groupReceiptSettings.quickScanStatusRequired = true
      ..groupReceiptSettings.quickScanCategoriesEnabled = false
      ..groupReceiptSettings.quickScanCategoriesRequired = false
      ..groupReceiptSettings.quickScanTagsEnabled = true
      ..groupReceiptSettings.quickScanTagsRequired = false);

    // Keep only the all-group placeholder(s) plus our two configured groups so
    // the group dropdown is short and both entries are on-screen. The seeded
    // admin can accumulate many groups; an appended entry would land below the
    // dropdown menu's viewport, where `find.text` can't reach it.
    final allGroups = groupModel.groups.where((g) => g.isAllGroup).toList();
    groupModel.setGroups([...allGroups, group1, group2]);
    await tester.pump();

    await openQuickScanImageForm(tester);

    // Select group 1 → config-A field set.
    await selectDropdown(tester, 'groupId', real.name);
    expect(quickScanDropdown('paidByUserId'), findsOneWidget,
        reason: 'A: paid-by shown');
    expect(quickScanDropdown('status'), findsNothing, reason: 'A: status hidden');
    expect(find.byType(CategorySelectField), findsOneWidget,
        reason: 'A: categories shown');
    expect(find.byType(TagSelectField), findsNothing, reason: 'A: tags hidden');

    // Switch to group 2 → the field set flips to config B.
    await selectDropdown(tester, 'groupId', group2.name);
    expect(quickScanDropdown('paidByUserId'), findsNothing,
        reason: 'B: paid-by hidden');
    expect(quickScanDropdown('status'), findsOneWidget, reason: 'B: status shown');
    expect(find.byType(CategorySelectField), findsNothing,
        reason: 'B: categories hidden');
    expect(find.byType(TagSelectField), findsOneWidget, reason: 'B: tags shown');

    // Group 2's status is shown+required and empty → submit is blocked with the
    // fix-error snackbar (a required field other than the categories case above).
    await expectQuickScanSubmitBlocked(tester);
  });

  testWidgets('paid-by required + empty blocks submit', (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    // Only paid-by shown+required; everything else off so paid-by is the sole
    // blocker.
    final group = await configureFirstGroup(tester, (b) => b
      ..groupReceiptSettings.quickScanPaidByEnabled = true
      ..groupReceiptSettings.quickScanPaidByRequired = true
      ..groupReceiptSettings.quickScanStatusEnabled = false
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = false
      ..groupReceiptSettings.quickScanCategoriesRequired = false
      ..groupReceiptSettings.quickScanTagsEnabled = false
      ..groupReceiptSettings.quickScanTagsRequired = false);

    await openQuickScanImageForm(tester);
    await selectDropdown(tester, 'groupId', group.name);
    expect(quickScanDropdown('paidByUserId'), findsOneWidget,
        reason: 'paid-by shown');

    // Paid-by required and left empty → submit blocked.
    await expectQuickScanSubmitBlocked(tester);
  });

  testWidgets('tags required + empty blocks submit', (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    // Only tags shown+required. Tags-required is enforced at submit (no field
    // validator), so this exercises the submit-side check specifically.
    final group = await configureFirstGroup(tester, (b) => b
      ..groupReceiptSettings.quickScanPaidByEnabled = false
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = false
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = false
      ..groupReceiptSettings.quickScanCategoriesRequired = false
      ..groupReceiptSettings.quickScanTagsEnabled = true
      ..groupReceiptSettings.quickScanTagsRequired = true);

    await openQuickScanImageForm(tester);
    await selectDropdown(tester, 'groupId', group.name);
    expect(find.byType(TagSelectField), findsOneWidget, reason: 'tags shown');

    // Tags required and left empty → submit blocked.
    await expectQuickScanSubmitBlocked(tester);
  });

  testWidgets('comment shown+required blocks submit until filled',
      (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    // Only the comment shown+required, so it is the sole blocker. The admin holds
    // group.comments.create, which the field additionally requires.
    final group = await configureFirstGroup(tester, (b) => b
      ..groupReceiptSettings.hideComments = false
      ..groupReceiptSettings.quickScanPaidByEnabled = false
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = false
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = false
      ..groupReceiptSettings.quickScanCategoriesRequired = false
      ..groupReceiptSettings.quickScanTagsEnabled = false
      ..groupReceiptSettings.quickScanTagsRequired = false
      ..groupReceiptSettings.quickScanCommentEnabled = true
      ..groupReceiptSettings.quickScanCommentRequired = true);

    await openQuickScanImageForm(tester);
    await selectDropdown(tester, 'groupId', group.name);

    expect(quickScanCommentField(), findsOneWidget, reason: 'comment shown');

    // Required and empty -> submit blocked with the fix-error snackbar.
    await expectQuickScanSubmitBlocked(tester);
  });

  testWidgets('comment field is hidden when the group hides comments',
      (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    await loginAsAdmin(tester);

    // hideComments overrides an enabled+required quick-scan comment: the field is
    // absent AND does not block submit (which would be unfixable).
    final group = await configureFirstGroup(tester, (b) => b
      ..groupReceiptSettings.hideComments = true
      ..groupReceiptSettings.quickScanPaidByEnabled = false
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = false
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = false
      ..groupReceiptSettings.quickScanCategoriesRequired = false
      ..groupReceiptSettings.quickScanTagsEnabled = false
      ..groupReceiptSettings.quickScanTagsRequired = false
      ..groupReceiptSettings.quickScanCommentEnabled = true
      ..groupReceiptSettings.quickScanCommentRequired = true);

    await openQuickScanImageForm(tester);
    await selectDropdown(tester, 'groupId', group.name);

    expect(quickScanCommentField(), findsNothing, reason: 'comment hidden');
  });

  // The two cases below assert the fields are genuinely ON SCREEN, not merely in
  // the widget tree -- see expectQuickScanFieldOnScreen for why the obvious
  // assertions all pass on a broken sheet. The sheet used to render its last
  // field into a clipped, unscrollable dead zone, which shipped a Quick Scan
  // whose configured comment box no one could see: the widget tests pump
  // QuickScanForm directly and never exercise the sheet layout, the cases above
  // only ever ask `findsOneWidget`, and quick_scan_submit_test reaches its
  // comment via `tester.ensureVisible`. They deliberately do not pin a surface
  // size -- `setSurfaceSize` is a no-op on the Linux desktop runner, whose
  // 1280x720 window is already short enough to expose this.
  testWidgets('a configured comment field is actually on screen',
      (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await loginAsAdmin(tester);

    // Exactly the reported configuration: comment the only optional field on.
    final group = await configureFirstGroup(tester, (b) => b
      ..groupReceiptSettings.hideComments = false
      ..groupReceiptSettings.quickScanPaidByEnabled = true
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = true
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = false
      ..groupReceiptSettings.quickScanCategoriesRequired = false
      ..groupReceiptSettings.quickScanTagsEnabled = false
      ..groupReceiptSettings.quickScanTagsRequired = false
      ..groupReceiptSettings.quickScanCommentEnabled = true
      ..groupReceiptSettings.quickScanCommentRequired = false);

    await openQuickScanImageForm(tester);
    await selectDropdown(tester, 'groupId', group.name);

    await expectQuickScanFieldOnScreen(tester, quickScanCommentField(),
        label: 'comment');

    // Visible is not enough -- it has to be usable.
    await tester.enterText(quickScanCommentField(), 'a typed quick scan note');
    await tester.pump();
    expect(find.text('a typed quick scan note'), findsOneWidget);
  });

  testWidgets('every optional field is actually on screen',
      (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();
    await loginAsAdmin(tester);

    // The tallest form the config can produce: every field shown at once.
    // Comment is last in the column, so it is the first to fall off the bottom --
    // but Categories and Tags are asserted too, so a partial regression cannot
    // hide behind it.
    final group = await configureFirstGroup(tester, (b) => b
      ..groupReceiptSettings.hideComments = false
      ..groupReceiptSettings.quickScanPaidByEnabled = true
      ..groupReceiptSettings.quickScanPaidByRequired = false
      ..groupReceiptSettings.quickScanStatusEnabled = true
      ..groupReceiptSettings.quickScanStatusRequired = false
      ..groupReceiptSettings.quickScanCategoriesEnabled = true
      ..groupReceiptSettings.quickScanCategoriesRequired = false
      ..groupReceiptSettings.quickScanTagsEnabled = true
      ..groupReceiptSettings.quickScanTagsRequired = false
      ..groupReceiptSettings.quickScanCommentEnabled = true
      ..groupReceiptSettings.quickScanCommentRequired = false);

    await openQuickScanImageForm(tester);
    await selectDropdown(tester, 'groupId', group.name);

    await expectQuickScanFieldOnScreen(
        tester, find.byType(CategorySelectField),
        label: 'categories');
    await expectQuickScanFieldOnScreen(tester, find.byType(TagSelectField),
        label: 'tags');
    await expectQuickScanFieldOnScreen(tester, quickScanCommentField(),
        label: 'comment');
  });
}
