import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/enums/form_state.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/category_select_field.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';

import '../helpers/receipt_form_test_helpers.dart';

// The add form's Group field is a default, not a lock: a picker with one option
// is not a choice, so the form seeds it (GroupModel.soleGroupId).
//
// The subtle half is that seeding the DROPDOWN is not enough -- the rest of the
// form reads the `groupId` State field, so paid-by, the category/tag pickers and
// the add-share button stay dead until the post-frame callback mirrors the seed
// into it. Those are asserted here rather than the form value alone.

const _memberId = 7;
const _groupOneId = 1;
const _groupTwoId = 2;
const _allGroupId = 99;
const _fieldId = 31;

final _users = [buildUserView(id: _memberId, displayName: 'Only Member')];

final _customField = buildCustomField(
  id: _fieldId,
  name: 'PO Number',
  type: api.CustomFieldType.TEXT,
);

api.Group _group(int id, {String? name, List<int>? defaultCustomFieldIds}) =>
    buildGroup(
      id: id,
      name: name ?? 'Group $id',
      members: [buildGroupMember(userId: _memberId, groupId: id)],
      defaultCustomFieldIds: defaultCustomFieldIds,
    );

Finder _dropdown(String name) =>
    find.byWidgetPredicate((w) => w is FormBuilderDropdown && w.name == name);

dynamic _groupValue(ReceiptFormHarness harness) =>
    harness.formKey.currentState?.fields['groupId']?.value;

/// The add-share "+" -- disabled while `groupId` is 0, so its enabled-ness is
/// the observable proof that the State field (not just the form value) was set.
Finder _addShareButton() => find.ancestor(
      of: find.byIcon(Icons.add),
      matching: find.byType(IconButton),
    );

void _expectNoUiErrors(WidgetTester tester) {
  expect(tester.takeException(), isNull);
  expect(find.byType(ErrorWidget), findsNothing);
}

void main() {
  testWidgets('seeds the group when the user belongs to exactly one',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: [_group(_groupOneId, name: 'My Receipts')],
      users: _users,
    );

    expect(_groupValue(harness), _groupOneId);
    _expectNoUiErrors(tester);
  });

  testWidgets('ignores the synthetic All group when counting', (tester) async {
    // Every user carries the All group alongside their real ones, so a
    // single-group user still has two entries in GroupModel.
    final harness = await pumpReceiptForm(
      tester,
      groups: [
        buildGroup(id: _allGroupId, name: 'All Groups', isAllGroup: true),
        _group(_groupOneId, name: 'My Receipts'),
      ],
      users: _users,
    );

    expect(_groupValue(harness), _groupOneId);
    _expectNoUiErrors(tester);
  });

  testWidgets('unlocks the group-derived fields, not just the form value',
      (tester) async {
    await pumpReceiptForm(
      tester,
      groups: [_group(_groupOneId)],
      users: _users,
    );

    // All three read the `groupId` State field, which only the post-frame
    // mirror sets -- they would stay hidden/disabled if the dropdown alone
    // had been seeded.
    expect(find.byType(CategorySelectField), findsOneWidget);
    expect(tester.widget<IconButton>(_addShareButton().first).onPressed,
        isNotNull);
    expect(
      tester
          .widget<FormBuilderDropdown>(_dropdown('paidByUserId'))
          .items
          .isNotEmpty,
      isTrue,
      reason: 'paid-by is group-scoped and empty at groupId 0',
    );
    _expectNoUiErrors(tester);
  });

  testWidgets('applies the sole group\'s default custom fields on load',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: [_group(_groupOneId, defaultCustomFieldIds: const [_fieldId])],
      users: _users,
      customFields: [_customField],
    );

    expect(
      harness.receiptModel.modifiedReceipt.customFields
          .map((value) => value.customFieldId),
      [_fieldId],
      reason: 'a seeded group is a picked group -- its defaults apply',
    );
    _expectNoUiErrors(tester);
  });

  testWidgets('leaves the group blank when the user belongs to more than one',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: [_group(_groupOneId), _group(_groupTwoId)],
      users: _users,
    );

    expect(_groupValue(harness), isNull);
    expect(find.byType(CategorySelectField), findsNothing);
    expect(
        tester.widget<IconButton>(_addShareButton().first).onPressed, isNull);
    _expectNoUiErrors(tester);
  });

  testWidgets('keeps an existing receipt\'s own group in edit mode',
      (tester) async {
    final harness = await pumpReceiptForm(
      tester,
      groups: [
        _group(_groupOneId),
        _group(_groupTwoId),
      ],
      receipt: getDefaultReceipt().rebuild((b) => b
        ..id = 5
        ..groupId = _groupTwoId
        ..paidByUserId = _memberId),
      users: _users,
      formState: WranglerFormState.edit,
    );

    expect(_groupValue(harness), _groupTwoId);
    _expectNoUiErrors(tester);
  });

  testWidgets('does not seed a group in view mode', (tester) async {
    // View mode always has a saved receipt, so the sole-group rule must not
    // reach it -- a "view" that invented a group would be a lie.
    final harness = await pumpReceiptForm(
      tester,
      groups: [_group(_groupOneId)],
      receipt: getDefaultReceipt().rebuild((b) => b
        ..id = 5
        ..groupId = _groupOneId
        ..paidByUserId = _memberId),
      users: _users,
      formState: WranglerFormState.view,
    );

    expect(_groupValue(harness), _groupOneId);
    expect(
      harness.receiptModel.modifiedReceipt.customFields,
      isEmpty,
      reason: 'view mode never applies group defaults',
    );
    _expectNoUiErrors(tester);
  });
}
