// Group default custom fields, on device.
//
// A group can declare custom fields that are pre-added to its receipts
// (`GroupReceiptSettings.defaultCustomFieldIds`), so each group is effectively
// its own receipt template. They are applied on load in every form state, so an
// existing receipt shows them too. Selecting a different group "smart swaps": a
// default this form added and the user never filled in is dropped, anything
// they typed into or added by hand is kept, and the new group's missing
// defaults are added.
//
// `test/widgets/receipt_form_default_custom_fields_test.dart` covers those
// rules exhaustively against injected models. What only this spec can prove is
// that the ids survive the wire: persisted through
// `PUT /group/{id}/groupReceiptSettings`, hydrated back onto `GroupModel` via
// AppData at login, applied by the real form, and accepted by the backend's
// `enforceReceiptCustomFieldSelection` on save.
//
// Every step also asserts NO UI error surfaced. The swap mutates the receipt's
// custom field list while FormBuilder fields mount and unmount under it, and a
// mistake there shows up as a thrown exception or a red ErrorWidget -- both of
// which a naive `find.text` assertion would happily pass around.
//
// Own file for the GoRouter-persistence reason (see other test files).

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
import 'helpers/nav.dart';
import 'helpers/permission_fixtures.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/receipt_test_helpers.dart';

/// The three seeded TEXT fields plus the two groups the specs switch between.
class _Fixture {
  _Fixture({
    required this.user,
    required this.alphaId,
    required this.alphaName,
    required this.betaId,
    required this.betaName,
    required this.fieldIds,
    required this.fieldNames,
  });

  final PermFixture user;
  final int alphaId;
  final String alphaName;
  final int betaId;
  final String betaName;

  /// Keyed 'a' / 'b' / 'c'.
  final Map<String, int> fieldIds;
  final Map<String, String> fieldNames;

  int id(String key) => fieldIds[key]!;

  String name(String key) => fieldNames[key]!;
}

/// Seeds three custom fields and a user belonging to exactly two groups.
///
/// Exactly two groups matters: the group dropdown is a Material menu whose
/// scrollable does not build off-screen items into the element tree, so a user
/// with many groups (the shared admin accumulates them across runs) can have
/// its target sitting below the menu's viewport where `find.text` can't reach.
/// A freshly provisioned user is a member of nothing until a fixture adds it.
///
/// The names are `zz-` prefixed because the "Add Custom Field" sheet lists the
/// catalog name-DESC in a lazy `ListView.builder`: a `zz-` name sorts to the
/// top, so its row is built without scrolling the sheet.
Future<_Fixture> _seedFixture() async {
  final jwt = await apiLogin(); // admin
  final suffix = DateTime.now().microsecondsSinceEpoch.toString();

  final fieldIds = <String, int>{};
  final fieldNames = <String, String>{};
  for (final key in const ['a', 'b', 'c']) {
    final name = 'zz-dcf-$suffix-$key';
    final field = await createCustomField(jwt: jwt, name: name, type: 'TEXT');
    fieldIds[key] = field['id'] as int;
    fieldNames[key] = name;
  }

  // Registered BEFORE the user/group teardowns so LIFO runs it LAST: deleting a
  // custom field destroys every value stored against it, so the receipts
  // holding those values (cascaded with their group) have to go first.
  addTearDown(() async {
    final j = await apiLogin();
    for (final id in fieldIds.values) {
      await deleteCustomField(id, jwt: j);
    }
  });

  // Legacy Editor holds group.receipts.create (needed to submit); the default
  // Legacy User app role holds app.custom-fields.read, which the whole feature
  // gates on -- without it the catalog fetch 403s into an empty list and the
  // form deliberately applies nothing.
  final editorRoleId = await groupRoleIdByName('Legacy Editor', jwt: jwt);
  final user = await provisionPermUser(groupRoleId: editorRoleId);

  final betaName = 'e2e-dcf-beta-$suffix';
  final betaId = await createGroupWithMember(
    name: betaName,
    memberUserId: user.userId,
    groupRoleId: editorRoleId,
    jwt: jwt,
  );
  addTearDown(() async => deleteGroup(betaId, jwt: await apiLogin()));

  return _Fixture(
    user: user,
    alphaId: user.groupId!,
    alphaName: user.groupName!,
    betaId: betaId,
    betaName: betaName,
    fieldIds: fieldIds,
    fieldNames: fieldNames,
  );
}

/// Asserts no UI error survived the step just performed, naming [step] so a red
/// run says which transition broke rather than only which frame.
///
/// The binding already hard-fails the test on an uncaught framework error (a
/// throw from the group dropdown's onChanged is reported with its own stack the
/// moment it happens), so this is the belt to that braces: it names the step and
/// catches a rendered `ErrorWidget` — a red screen is perfectly findable, so a
/// naive `find.text` assertion would happily pass around one.
///
/// Deliberately NOT a `FlutterError.onError` collector. That looks like the
/// obvious way to capture framework errors, but it takes the handler away from
/// `TestWidgetsFlutterBinding`, whose `_runTest` then trips its own
/// `'_pendingExceptionDetails != null'` assertion the moment anything fails — so
/// every real failure in this file, a plain `expect` mismatch included, is
/// reported as an unreadable framework assertion instead of its own message.
/// (Verified: it turned a genuine stale-value failure into exactly that.)
void _expectNoUiErrors(WidgetTester tester, String step) {
  expect(tester.takeException(), isNull,
      reason: 'A framework error surfaced during: $step');
  expect(find.byType(ErrorWidget), findsNothing,
      reason: 'A red ErrorWidget rendered during: $step');
}

Finder _customField(int customFieldId) =>
    find.byKey(ValueKey('customFieldValue_$customFieldId'));

Finder _textInput(int customFieldId) => find.byWidgetPredicate(
      (w) => w is FormBuilderTextField && w.name == 'customField_$customFieldId',
    );

/// The text the field for [customFieldId] currently renders, or null when it is
/// not mounted. Reads the live `TextField` controller rather than the
/// FormBuilder value map, because a resurrected value is exactly what the
/// controller would show and the map would not.
String? _renderedText(WidgetTester tester, int customFieldId) {
  final input = find.descendant(
    of: _customField(customFieldId),
    matching: find.byType(TextField),
  );
  if (input.evaluate().isEmpty) {
    return null;
  }
  return tester.widget<TextField>(input.first).controller?.text;
}

/// Asserts every id in [expected] is mounted exactly once and every id in
/// [unexpected] is not mounted at all. Exact on purpose: a field mounted twice
/// fails as loudly as one missing, which is the shape a duplicated default takes.
void _expectFields(List<int> expected, List<int> unexpected) {
  for (final id in expected) {
    expect(_customField(id), findsOneWidget,
        reason: 'custom field $id should be mounted exactly once');
  }
  for (final id in unexpected) {
    expect(_customField(id), findsNothing,
        reason: 'custom field $id should not be mounted');
  }
}

/// Drains a few frames so a sheet/menu transition settles before the next tap.
Future<void> _drain(WidgetTester tester) async {
  for (int i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
}

/// Opens an empty Add Manual Receipt form from the bottom-nav scan/add slot and
/// returns once the custom field catalog has loaded.
Future<void> _openAddReceiptForm(WidgetTester tester) async {
  await openManualReceiptForm(tester);
  // The form screen's FutureBuilder awaits loadCustomFields before it builds,
  // so this button's presence means the catalog is in -- and the swap skips ids
  // it can't find in the catalog, which would otherwise be a silent race.
  await pumpUntilFound(tester, find.text('Add Custom Field'));
}

Future<void> _selectGroup(WidgetTester tester, String name) async {
  await tester.ensureVisible(
      find.byWidgetPredicate((w) => w is FormBuilderDropdown && w.name == 'groupId'));
  await tester.pump(const Duration(milliseconds: 100));
  await selectDropdown(tester, 'groupId', name);
}

Future<void> _enterFieldText(
    WidgetTester tester, int customFieldId, String text) async {
  await tester.ensureVisible(_textInput(customFieldId));
  await tester.pump(const Duration(milliseconds: 100));
  await tester.enterText(_textInput(customFieldId), text);
  await _drain(tester);
}

/// Adds [name] by hand through the "Add Custom Field" sheet.
Future<void> _addCustomFieldByHand(WidgetTester tester, String name) async {
  final button = find.text('Add Custom Field');
  await tester.ensureVisible(button);
  // ensureVisible jumps the scroll position WITHOUT relayout, so the button's
  // global offset is stale until a frame is pumped; tapping now would compute
  // the pre-scroll centre and miss (tap-flake pattern #1 in mobile/CLAUDE.md).
  await tester.pump(const Duration(milliseconds: 100));
  await tester.tap(button);
  await pumpUntilFound(tester, find.text('Select Custom Field'));
  await _drain(tester);

  final row = find.descendant(of: find.byType(ListTile), matching: find.text(name));
  await pumpUntilFound(tester, row.hitTestable());
  await tester.tap(row);
  await _drain(tester);
}

/// Taps the remove button on the mounted field for [customFieldId].
Future<void> _removeCustomFieldByHand(
    WidgetTester tester, int customFieldId) async {
  final removeButton = find.descendant(
    of: _customField(customFieldId),
    matching: find.byIcon(Icons.remove_circle_outline),
  );
  await tester.ensureVisible(removeButton);
  await tester.pump(const Duration(milliseconds: 100));
  await tester.tap(removeButton);
  await _drain(tester);
}

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  testWidgets('a group change swaps the defaults it owns and keeps the rest',
      (tester) async {
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    final fixture = await _seedFixture();
    final adminJwt = await apiLogin();
    await setGroupDefaultCustomFields(
      groupId: fixture.alphaId,
      jwt: adminJwt,
      customFieldIds: [fixture.id('a'), fixture.id('b')],
    );
    await setGroupDefaultCustomFields(
      groupId: fixture.betaId,
      jwt: adminJwt,
      customFieldIds: [fixture.id('c')],
    );

    await loginAs(tester,
        username: fixture.user.username, password: fixture.user.password);
    _expectNoUiErrors(tester, 'login');

    final receiptName = 'e2e-dcf-${DateTime.now().millisecondsSinceEpoch}';
    await _openAddReceiptForm(tester);
    // No group picked yet, so nothing is pre-added.
    _expectFields(const [], [fixture.id('a'), fixture.id('b'), fixture.id('c')]);
    _expectNoUiErrors(tester, 'opening the add form with no group selected');

    await tester.enterText(formField('name'), receiptName);
    await tester.enterText(formField('amount'), '12.34');

    // Alpha declares A + B.
    await _selectGroup(tester, fixture.alphaName);
    _expectFields([fixture.id('a'), fixture.id('b')], [fixture.id('c')]);
    _expectNoUiErrors(tester, 'selecting Alpha');

    // Typing into A makes it the user's data; C is added by hand, so the swap
    // never owned it in the first place.
    await _enterFieldText(tester, fixture.id('a'), 'kept-A');
    await _addCustomFieldByHand(tester, fixture.name('c'));
    _expectFields(
        [fixture.id('a'), fixture.id('b'), fixture.id('c')], const []);
    _expectNoUiErrors(tester, 'typing into A and adding C by hand');

    // Beta declares only C: B is the one field the swap owns AND is still
    // empty, so it is the only one dropped. C is already attached and must not
    // be added a second time.
    await _selectGroup(tester, fixture.betaName);
    _expectFields([fixture.id('a'), fixture.id('c')], [fixture.id('b')]);
    expect(_renderedText(tester, fixture.id('a')), 'kept-A',
        reason: 'a default the user typed into must survive the swap');
    _expectNoUiErrors(tester, 'switching Alpha -> Beta');

    // Back to Alpha: B returns as a default and must come back BLANK, C stays
    // because the user added it and A is still theirs.
    await _selectGroup(tester, fixture.alphaName);
    _expectFields(
        [fixture.id('a'), fixture.id('b'), fixture.id('c')], const []);
    expect(_renderedText(tester, fixture.id('a')), 'kept-A');
    expect(_renderedText(tester, fixture.id('b')), '',
        reason: 're-added defaults must not resurrect a previous value');
    _expectNoUiErrors(tester, 'switching Beta -> Alpha');

    await _enterFieldText(tester, fixture.id('b'), 'typed-B');
    await selectDropdown(tester, 'paidByUserId', fixture.user.displayName);

    // Deliberately NOT pumpAndSettle: the receipt form hosts a
    // CircularLoadingProgress that spins while customFieldModel reloads in the
    // background, so settling is not guaranteed -- and pumpAndSettle's first
    // positional argument is the frame interval, not a timeout, so a tree that
    // never settles blocks on the 10-minute default. Drain the dropdown
    // overlay's teardown, then wait for the button to be genuinely tappable.
    await _drain(tester);
    await pumpUntilFound(tester, find.byType(BottomSubmitButton).hitTestable());
    await tester.tap(find.byType(BottomSubmitButton).hitTestable());
    await pumpUntilFound(tester, find.byType(ReceiptEditPopupMenu),
        timeout: const Duration(seconds: 20));
    _expectNoUiErrors(tester, 'saving the receipt');

    // Reaching /view already proves the save cleared the backend's
    // enforceReceiptCustomFieldSelection, which 403s a submitted id set that
    // does not match the stored one -- so every attached field rode the
    // payload, the empty one included. Read it back to pin the values too.
    final receiptId = receiptIdFromUrl(currentUrl(tester));
    final receipt = await getReceipt(receiptId, jwt: await apiLogin());
    final values = ((receipt['customFields'] as List?) ?? const [])
        .cast<Map<String, dynamic>>();
    String? valueFor(int id) => values
        .firstWhere((cf) => cf['customFieldId'] == id,
            orElse: () => <String, dynamic>{})['stringValue'] as String?;

    expect(values.map((cf) => cf['customFieldId']).toSet(),
        {fixture.id('a'), fixture.id('b'), fixture.id('c')},
        reason: 'the saved receipt should carry exactly the attached fields');
    expect(valueFor(fixture.id('a')), 'kept-A');
    expect(valueFor(fixture.id('b')), 'typed-B');
    expect(valueFor(fixture.id('c')), isNull,
        reason: 'an attached-but-empty field is stored with null columns');
  });

  testWidgets('a hand-removed default comes back blank, not with its old value',
      (tester) async {
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    final fixture = await _seedFixture();
    final adminJwt = await apiLogin();
    await setGroupDefaultCustomFields(
      groupId: fixture.alphaId,
      jwt: adminJwt,
      customFieldIds: [fixture.id('a'), fixture.id('b')],
    );
    await setGroupDefaultCustomFields(
      groupId: fixture.betaId,
      jwt: adminJwt,
      customFieldIds: [fixture.id('c')],
    );

    await loginAs(tester,
        username: fixture.user.username, password: fixture.user.password);
    await _openAddReceiptForm(tester);
    await _selectGroup(tester, fixture.alphaName);
    _expectFields([fixture.id('a'), fixture.id('b')], [fixture.id('c')]);

    // Type into B, then take it off by hand. Removing by hand also un-owns it,
    // so the swap must not fight the decision on the way out.
    await _enterFieldText(tester, fixture.id('b'), 'ghost');
    await _removeCustomFieldByHand(tester, fixture.id('b'));
    _expectFields([fixture.id('a')], [fixture.id('b'), fixture.id('c')]);
    _expectNoUiErrors(tester, 'removing B by hand');

    // Beta has no opinion about B, so it stays gone. A is empty and owned, so
    // it is dropped; C arrives.
    await _selectGroup(tester, fixture.betaName);
    _expectFields([fixture.id('c')], [fixture.id('a'), fixture.id('b')]);
    _expectNoUiErrors(tester, 'switching Alpha -> Beta after a hand removal');

    // Back to Alpha: B is re-added as a default. FormBuilder's
    // clearValueOnUnregister defaults to FALSE, so without the form's explicit
    // clear-on-remove the re-registered field would be handed 'ghost' straight
    // back out of the value map.
    await _selectGroup(tester, fixture.alphaName);
    _expectFields([fixture.id('a'), fixture.id('b')], [fixture.id('c')]);
    expect(_renderedText(tester, fixture.id('b')), '',
        reason: 'B must come back blank, not carrying "ghost"');
    _expectNoUiErrors(tester, 'switching Beta -> Alpha');
  });

  testWidgets("an existing receipt shows its group's defaults in view and edit",
      (tester) async {
    await binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => binding.setSurfaceSize(null));

    final fixture = await _seedFixture();
    final adminJwt = await apiLogin();
    await setGroupDefaultCustomFields(
      groupId: fixture.alphaId,
      jwt: adminJwt,
      customFieldIds: [fixture.id('a'), fixture.id('b')],
    );
    await setGroupDefaultCustomFields(
      groupId: fixture.betaId,
      jwt: adminJwt,
      customFieldIds: [fixture.id('c')],
    );

    // Seeded server-side with no custom fields at all, even though its group
    // declares two -- the form is what retro-fits them, on load.
    final receiptName = 'e2e-dcf-existing-${DateTime.now().millisecondsSinceEpoch}';
    await createReceipt(
      groupId: fixture.alphaId,
      paidByUserId: fixture.user.userId,
      jwt: adminJwt,
      name: receiptName,
    );

    await loginAs(tester,
        username: fixture.user.username, password: fixture.user.password);
    await openGroupReceipts(tester, fixture.alphaName, receiptName);

    // View mode renders Alpha's two defaults, blank and read-only -- they are
    // meant to read as that group's built-in receipt fields.
    await tester.tap(find.text(receiptName).hitTestable());
    await pumpUntilFound(tester, find.byType(ReceiptEditPopupMenu));
    await _drain(tester);
    await pumpUntilFound(tester, _customField(fixture.id('a')));
    _expectFields([fixture.id('a'), fixture.id('b')], [fixture.id('c')]);
    expect(_renderedText(tester, fixture.id('a')), '');
    _expectNoUiErrors(tester, 'opening the receipt in view mode');

    await tester.tap(find.byType(ReceiptEditPopupMenu));
    await pumpUntilFound(tester, find.text('Edit').hitTestable());
    await _drain(tester);
    await tester.tap(find.text('Edit').hitTestable());
    await pumpUntilFound(tester, find.byType(BottomSubmitButton));
    await pumpUntilFound(tester, find.text('Add Custom Field'));
    _expectFields([fixture.id('a'), fixture.id('b')], [fixture.id('c')]);
    _expectNoUiErrors(tester, 'opening the receipt in edit mode');

    // The load-applied defaults are still the swap's: both are empty, so a
    // group change takes them back and leaves only Beta's.
    await _selectGroup(tester, fixture.betaName);
    _expectFields([fixture.id('c')], [fixture.id('a'), fixture.id('b')]);
    _expectNoUiErrors(tester, 'changing the group in edit mode');
  });
}
