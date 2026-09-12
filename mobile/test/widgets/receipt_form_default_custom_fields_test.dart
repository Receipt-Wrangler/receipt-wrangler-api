import 'package:built_collection/built_collection.dart';
import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/enums/form_state.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/custom_field_widget.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';

import '../helpers/receipt_form_test_helpers.dart';

// A group can declare default custom fields (`GroupReceiptSettings
// .defaultCustomFieldIds`), which the receipt form pre-adds for that group --
// each group is effectively its own receipt template. They are applied on load
// in EVERY form state, so a saved receipt that predates the configuration shows
// them too. Selecting a different group "smart swaps": defaults this form added
// and the user never filled in are dropped, anything they typed into or added
// by hand is kept, and the new group's missing defaults are added.
//
// Every case asserts the swap left NO UI error behind. The path mutates the
// receipt's custom field list while FormBuilder fields mount and unmount under
// it, so a mistake surfaces as a thrown exception or a red ErrorWidget -- a
// red screen is perfectly findable, so a naive `find.text` assertion would
// happily pass around one.

const _groupOneId = 1;
const _groupTwoId = 2;
const _groupThreeId = 3;

const _fieldAId = 11;
const _fieldBId = 12;
const _fieldCId = 13;
const _fieldDId = 14;
const _missingFieldId = 99;

final _fieldA = buildCustomField(id: _fieldAId, name: 'Cost Centre');
final _fieldB = buildCustomField(id: _fieldBId, name: 'PO Number');
final _fieldC = buildCustomField(id: _fieldCId, name: 'Notes');
final _fieldD = buildCustomField(
  id: _fieldDId,
  name: 'Reimbursed',
  type: api.CustomFieldType.BOOLEAN,
);

final _catalog = [_fieldA, _fieldB, _fieldC, _fieldD];

const _memberId = 7;

/// Every group holds the same single member, so the paid-by dropdown always has
/// an item -- FormBuilderDropdown asserts its initial value is one of them.
final _users = [buildUserView(id: _memberId)];

/// Group One defaults to field A, Group Two to field B, Group Three to none.
List<api.Group> _groups({
  List<int>? groupOneDefaults = const [_fieldAId],
  List<int>? groupTwoDefaults = const [_fieldBId],
}) =>
    [
      buildGroup(
        id: _groupOneId,
        name: 'Group One',
        members: [buildGroupMember(userId: _memberId, groupId: _groupOneId)],
        defaultCustomFieldIds: groupOneDefaults,
      ),
      buildGroup(
        id: _groupTwoId,
        name: 'Group Two',
        members: [buildGroupMember(userId: _memberId, groupId: _groupTwoId)],
        defaultCustomFieldIds: groupTwoDefaults,
      ),
      // A group that has configured nothing. Its ids are absent rather than
      // `[]`, which is what a client predating the field would also see.
      buildGroup(
        id: _groupThreeId,
        name: 'Group Three',
        members: [buildGroupMember(userId: _memberId, groupId: _groupThreeId)],
      ),
    ];

api.Receipt _receiptInGroup(
  int groupId, {
  int id = 0,
  List<api.CustomFieldValue> customFields = const [],
}) =>
    getDefaultReceipt().rebuild((b) => b
      ..id = id
      ..groupId = groupId
      ..paidByUserId = _memberId
      ..customFields = ListBuilder<api.CustomFieldValue>(customFields));

Finder _dropdown(String name) =>
    find.byWidgetPredicate((w) => w is FormBuilderDropdown && w.name == name);

/// Locates a custom field row by the ValueKey the receipt form stamps on it,
/// rather than by widget type -- per mobile/CLAUDE.md, prefer find.byKey where a
/// key exists. (The FormBuilder field finders below stay name-predicates: those
/// fields carry no Key, which is the documented exception.)
Finder _customFieldWidget(int customFieldId) =>
    find.byKey(ValueKey('customFieldValue_$customFieldId'));

Finder _textInput(int customFieldId) => find.byWidgetPredicate(
      (w) => w is FormBuilderTextField && w.name == 'customField_$customFieldId',
    );

Finder _addCustomFieldButton() => find.text('Add Custom Field');

/// The custom field ids currently attached to the receipt being edited.
List<int> _attachedIds(ReceiptFormHarness harness) => harness
    .receiptModel.modifiedReceipt.customFields
    .map((cfv) => cfv.customFieldId)
    .toList();

/// Neither a thrown exception nor a red `ErrorWidget` survived the last pump.
void _expectNoUiErrors(WidgetTester tester) {
  expect(tester.takeException(), isNull);
  expect(find.byType(ErrorWidget), findsNothing);
}

/// Opens the group dropdown and picks [name]. `.last` picks the open menu's
/// copy of the label -- the closed field renders the selected one too.
Future<void> _selectGroup(WidgetTester tester, String name) async {
  await tester.ensureVisible(_dropdown('groupId'));
  await tester.pumpAndSettle();
  await tester.tap(_dropdown('groupId'));
  await tester.pumpAndSettle();
  await tester.tap(find.text(name).last);
  await tester.pumpAndSettle();
}

/// Opens the "Add Custom Field" sheet and picks [name]. The button sits below
/// the fold once a few fields are mounted, so scroll it into view first.
Future<void> _addCustomFieldNamed(WidgetTester tester, String name) async {
  await tester.ensureVisible(_addCustomFieldButton());
  await tester.pumpAndSettle();
  await tester.tap(_addCustomFieldButton());
  await tester.pumpAndSettle();

  await tester.tap(find.descendant(
    of: find.byType(ListTile),
    matching: find.text(name),
  ));
  await tester.pumpAndSettle();
}

/// Taps the remove button of the custom field [customFieldId] renders.
Future<void> _removeCustomField(WidgetTester tester, int customFieldId) async {
  final removeButton = find.descendant(
    of: _customFieldWidget(customFieldId),
    matching: find.byIcon(Icons.remove_circle_outline),
  );

  await tester.ensureVisible(removeButton);
  await tester.pumpAndSettle();
  await tester.tap(removeButton);
  await tester.pumpAndSettle();
}

void main() {
  testWidgets("applies the selected group's default custom fields",
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
    );

    expect(find.byType(CustomFieldWidget), findsNothing);

    await _selectGroup(tester, 'Group One');

    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('applies the defaults on load when the add form knows its group',
      (tester) async {
    // The add form is normally opened without a group, but it can mount with
    // one already set (prefilled from the group the user came from).
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(_groupOneId),
    );

    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId]);
    _expectNoUiErrors(tester);
  });

  testWidgets("drops an empty auto-applied default for the new group's",
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group One');
    expect(_attachedIds(harness), [_fieldAId]);
    _expectNoUiErrors(tester);

    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldAId), findsNothing);
    expect(_customFieldWidget(_fieldBId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldBId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('keeps an auto-applied default the user typed into',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group One');
    await tester.enterText(_textInput(_fieldAId), 'CC-42');
    await tester.pumpAndSettle();

    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(
      harness.customFieldValue(_fieldAId),
      'CC-42',
      reason: 'a filled default is the user\'s data now, not the swap\'s',
    );
    expect(_customFieldWidget(_fieldBId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId, _fieldBId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('keeps a manually added custom field, with its value',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group One');
    await _addCustomFieldNamed(tester, 'Notes');
    await tester.enterText(_textInput(_fieldCId), 'hand written');
    await tester.pumpAndSettle();
    expect(_attachedIds(harness), [_fieldAId, _fieldCId]);

    // Field A is dropped from the middle of the list here, so the manually
    // added field shifts up a slot -- its value and its rendered text have to
    // travel with it.
    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldCId), findsOneWidget);
    expect(harness.customFieldValue(_fieldCId), 'hand written');
    // Rendered under its OWN label, and nowhere else: the elements have to
    // follow custom field identity rather than list position, or the text stays
    // behind in the slot and shows up under the field that moved into it.
    expect(
      find.descendant(
        of: _customFieldWidget(_fieldCId),
        matching: find.text('hand written'),
      ),
      findsOneWidget,
    );
    expect(find.text('hand written'), findsOneWidget);
    expect(harness.customFieldValue(_fieldBId), isNull);
    expect(_customFieldWidget(_fieldAId), findsNothing);
    expect(_attachedIds(harness), [_fieldCId, _fieldBId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('A -> B -> A leaves no residue and the default comes back blank',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group One');
    // Typed and then cleared: the field is empty again, so it is still the
    // swap's to take -- but FormBuilder keeps an unregistered field's value
    // unless it is explicitly cleared, so this is what would come back.
    await tester.enterText(_textInput(_fieldAId), 'CC-42');
    await tester.pumpAndSettle();
    await tester.enterText(_textInput(_fieldAId), '');
    await tester.pumpAndSettle();

    await _selectGroup(tester, 'Group Two');
    expect(_attachedIds(harness), [_fieldBId]);
    _expectNoUiErrors(tester);

    await _selectGroup(tester, 'Group One');

    expect(_customFieldWidget(_fieldBId), findsNothing);
    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId],
        reason: 'no duplicate and nothing left over from Group Two');
    expect(harness.customFieldValue(_fieldAId), isNull);
    expect(find.text('CC-42'), findsNothing);
    _expectNoUiErrors(tester);
  });

  testWidgets('does not fight a default the user removed by hand',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group One');
    await _removeCustomField(tester, _fieldAId);

    expect(_customFieldWidget(_fieldAId), findsNothing);
    expect(_attachedIds(harness), isEmpty);
    _expectNoUiErrors(tester);

    // Adding it back by hand makes it the user's field, so the swap must leave
    // it alone even though it is empty and Group Two does not want it.
    await _addCustomFieldNamed(tester, 'Cost Centre');
    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId, _fieldBId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('treats an unchecked BOOLEAN default as empty, a checked one not',
      (tester) async {
    // CustomFieldWidget seeds checkboxes with `false`, so the emptiness rule
    // has to be type-aware: unchecked is empty and swappable, deliberately
    // checked is the user's answer and stays.
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(groupOneDefaults: const [_fieldDId]),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group One');
    expect(_attachedIds(harness), [_fieldDId]);

    await _selectGroup(tester, 'Group Two');
    expect(_customFieldWidget(_fieldDId), findsNothing,
        reason: 'left unchecked, so still the swap\'s to take back');
    _expectNoUiErrors(tester);

    await _selectGroup(tester, 'Group One');
    await tester.ensureVisible(find.byType(Checkbox));
    await tester.pumpAndSettle();
    await tester.tap(find.byType(Checkbox));
    await tester.pumpAndSettle();
    expect(harness.customFieldValue(_fieldDId), isTrue);

    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldDId), findsOneWidget);
    expect(harness.customFieldValue(_fieldDId), isTrue);
    expect(_attachedIds(harness), [_fieldDId, _fieldBId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('skips default ids that are not in the custom field catalog',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(groupOneDefaults: const [_fieldAId, _missingFieldId]),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group One');

    expect(_attachedIds(harness), [_fieldAId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('applies nothing when the custom field catalog is empty',
      (tester) async {
    // No `app.custom-fields.read`: CustomFieldModel swallows the 403 into an
    // empty catalog. Attaching a field the caller cannot read would make the
    // backend 403 their save (enforceReceiptCustomFieldSelection).
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: const [],
      users: _users,
    );

    await _selectGroup(tester, 'Group One');

    expect(find.byType(CustomFieldWidget), findsNothing);
    expect(_attachedIds(harness), isEmpty);
    _expectNoUiErrors(tester);
  });

  testWidgets('a group with no configured defaults is a no-op', (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
    );

    await _selectGroup(tester, 'Group Three');

    expect(find.byType(CustomFieldWidget), findsNothing);
    expect(_attachedIds(harness), isEmpty);
    _expectNoUiErrors(tester);
  });

  // A group's defaults are meant to read as its built-in receipt fields, so a
  // saved receipt that predates the configuration picks them up on load -- in
  // view as well as edit.
  testWidgets("applies the group's missing defaults in view mode",
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(_groupOneId, id: 5),
      formState: WranglerFormState.view,
    );

    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId]);
    expect(harness.customFieldValue(_fieldAId), isNull);
    // Display only: view mode offers no way to change the set.
    expect(
      find.descendant(
        of: _customFieldWidget(_fieldAId),
        matching: find.byIcon(Icons.remove_circle_outline),
      ),
      findsNothing,
    );
    expect(_addCustomFieldButton(), findsNothing);
    _expectNoUiErrors(tester);
  });

  testWidgets("applies the group's missing defaults on an edit form's load",
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(_groupOneId, id: 5),
      formState: WranglerFormState.edit,
    );

    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('does not duplicate a default the receipt already carries',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(
        _groupOneId,
        id: 5,
        customFields: [
          buildCustomFieldValue(
            customFieldId: _fieldAId,
            receiptId: 5,
            stringValue: 'CC-7',
          ),
        ],
      ),
      formState: WranglerFormState.edit,
    );

    expect(_customFieldWidget(_fieldAId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldAId]);
    expect(harness.customFieldValue(_fieldAId), 'CC-7');
    _expectNoUiErrors(tester);
  });

  testWidgets('edit mode applies the new group\'s defaults on a group change',
      (tester) async {
    // Group Three configures none, so nothing lands until the group changes.
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(_groupThreeId, id: 5),
      formState: WranglerFormState.edit,
    );

    expect(find.byType(CustomFieldWidget), findsNothing);
    expect(_attachedIds(harness), isEmpty);
    _expectNoUiErrors(tester);

    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldBId), findsOneWidget);
    expect(_attachedIds(harness), [_fieldBId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('a default applied on load is still the swap\'s to take back',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(_groupOneId, id: 5),
      formState: WranglerFormState.edit,
    );

    expect(_attachedIds(harness), [_fieldAId]);

    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldAId), findsNothing);
    expect(_attachedIds(harness), [_fieldBId]);
    _expectNoUiErrors(tester);
  });

  // The view -> edit navigation, for real: the app bar menu's Edit entry is a
  // plain `go`, so the form is remounted (fresh State) against the same
  // ReceiptModel. Provenance lives on the model precisely so these next two
  // cases can disagree -- the form's field is still the form's, the user's is
  // still the user's.
  testWidgets('takes back a default a previous mount of the form attached',
      (tester) async {
    final viewHarness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(_groupOneId, id: 5),
      formState: WranglerFormState.view,
    );
    expect(_attachedIds(viewHarness), [_fieldAId]);

    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receiptModel: viewHarness.receiptModel,
      formState: WranglerFormState.edit,
    );
    expect(_attachedIds(harness), [_fieldAId]);
    _expectNoUiErrors(tester);

    await _selectGroup(tester, 'Group Two');

    expect(_customFieldWidget(_fieldAId), findsNothing);
    expect(_attachedIds(harness), [_fieldBId]);
    _expectNoUiErrors(tester);
  });

  testWidgets('a default re-added by hand stays the user\'s across a remount',
      (tester) async {
    // Removing an auto-applied default and adding it straight back makes it the
    // user's, and leaves it empty -- so only provenance separates it from one
    // the form just attached. A remount must not forget that and swap it out.
    final first = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receipt: _receiptInGroup(_groupOneId, id: 5),
      formState: WranglerFormState.edit,
    );
    expect(_attachedIds(first), [_fieldAId]);

    await _removeCustomField(tester, _fieldAId);
    expect(_attachedIds(first), isEmpty);
    await _addCustomFieldNamed(tester, 'Cost Centre');
    expect(_attachedIds(first), [_fieldAId]);
    _expectNoUiErrors(tester);

    final harness = await pumpReceiptForm(
      tester,
      groups: _groups(),
      customFields: _catalog,
      users: _users,
      receiptModel: first.receiptModel,
      formState: WranglerFormState.edit,
    );
    expect(_attachedIds(harness), [_fieldAId]);

    await _selectGroup(tester, 'Group Two');

    expect(
      _customFieldWidget(_fieldAId),
      findsOneWidget,
      reason: 'the user put this field back by hand; the swap does not own it',
    );
    expect(_attachedIds(harness), [_fieldAId, _fieldBId]);
    _expectNoUiErrors(tester);
  });
}
