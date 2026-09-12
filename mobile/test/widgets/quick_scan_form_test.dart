import 'dart:typed_data';

import 'package:built_collection/built_collection.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/category_model.dart';
import 'package:receipt_wrangler_mobile/models/context_model.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/models/tag_model.dart';
import 'package:receipt_wrangler_mobile/models/user_model.dart';
import 'package:receipt_wrangler_mobile/models/user_preferences_model.dart';
import 'package:receipt_wrangler_mobile/receipts/widgets/quick_scan_form.dart';
import 'package:receipt_wrangler_mobile/shared/classes/quick_scan_image.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/category_select_field.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/tag_select_field.dart';

// QuickScanForm shows/hides + conditionally-requires each field per the selected
// group's quick-scan config (GroupReceiptSettings.quickScan*). When a field is
// hidden/optional the server backfills a default, so the form omits it. Field
// visibility/requirement is reactive to the group dropdown: changing the group
// re-reads the new group's config and clears the group-scoped fields.
//
// Until a group is picked there is no config to honour, so ONLY the Group
// dropdown renders. Hiding a field for that reason must not destroy its value —
// the user's quickScanDefault* prefills have to survive the first group pick.

const _groupId = 1;
const _group2Id = 2;

api.GroupReceiptSettings _settings({
  int id = _groupId,
  bool paidByEnabled = true,
  bool paidByRequired = true,
  bool statusEnabled = true,
  bool statusRequired = true,
  bool categoriesEnabled = false,
  bool categoriesRequired = false,
  bool tagsEnabled = false,
  bool tagsRequired = false,
  bool commentEnabled = false,
  bool commentRequired = false,
  bool hideComments = false,
}) {
  return (api.GroupReceiptSettingsBuilder()
        ..id = id
        ..createdAt = ''
        ..groupId = id
        ..quickScanPaidByEnabled = paidByEnabled
        ..quickScanPaidByRequired = paidByRequired
        ..quickScanStatusEnabled = statusEnabled
        ..quickScanStatusRequired = statusRequired
        ..quickScanCategoriesEnabled = categoriesEnabled
        ..quickScanCategoriesRequired = categoriesRequired
        ..quickScanTagsEnabled = tagsEnabled
        ..quickScanTagsRequired = tagsRequired
        ..quickScanCommentEnabled = commentEnabled
        ..quickScanCommentRequired = commentRequired
        ..hideComments = hideComments)
      .build();
}

api.Group _group(
  api.GroupReceiptSettings settings, {
  String name = 'Test Group',
  List<api.GroupMember> members = const [],
}) =>
    (api.GroupBuilder()
          ..id = settings.groupId
          ..createdAt = ''
          ..name = name
          ..isAllGroup = false
          ..status = api.GroupStatus.ACTIVE
          ..groupMembers = ListBuilder<api.GroupMember>(members)
          ..groupReceiptSettings.replace(settings))
        .build();

api.GroupMember _member(int userId, int groupId) =>
    (api.GroupMemberBuilder()
          ..groupId = groupId
          ..userId = userId)
        .build();

api.UserView _user(int id, String displayName) =>
    (api.UserViewBuilder()
          ..id = id
          ..username = 'u$id'
          ..displayName = displayName
          ..isDummyUser = false)
        .build();

api.Category _category() =>
    (api.CategoryBuilder()
          ..id = 1
          ..name = 'Food')
        .build();

api.Tag _tag() =>
    (api.TagBuilder()
          ..id = 1
          ..name = 'Trip')
        .build();

QuickScanImage _image(
  int? groupId, {
  int? paidByUserId,
  api.ReceiptStatus? status,
}) => QuickScanImage(
  multipartFile: MultipartFile.fromBytes(const <int>[]),
  bytes: Uint8List(0),
  formKey: GlobalKey<FormBuilderState>(),
  groupId: groupId,
  paidByUserId: paidByUserId,
  status: status,
);

/// Pumps [QuickScanForm] with a real [GroupModel] holding [groups] and a
/// [UserModel] holding [users] (the latter backs the paid-by dropdown items).
/// [imageGroupId] drives which group's config the form initially reads
/// (0 = no group selected → only the Group dropdown renders). Returns the form
/// key.
///
/// The form change callback MIRRORS the real consumer (`quick_scan.dart`), which
/// writes every record member onto the image unconditionally. That write-back is
/// what makes the prefill-preservation cases meaningful — with an inert `(_) {}`
/// callback the image is never mutated, so those tests would pass even with the
/// bug present. [onFormChange] additionally exposes the last emitted record.
Future<GlobalKey<FormBuilderState>> _pumpFormGroups(
  WidgetTester tester, {
  required List<api.Group> groups,
  required int imageGroupId,
  List<api.UserView> users = const [],
  int? imagePaidByUserId,
  api.ReceiptStatus? imageStatus,
  void Function(QuickScanFormValues values)? onFormChange,
  // The comment field is additionally gated on group.comments.create, so every
  // group the test uses is granted it unless a test opts out.
  bool canCreateComments = true,
}) async {
  final image = _image(
    imageGroupId,
    paidByUserId: imagePaidByUserId,
    status: imageStatus,
  );
  final groupModel = GroupModel()..setGroups(groups);
  final userModel = UserModel()..setUsers(users);
  final permissionsModel =
      PermissionsModel()..setPermissions(const [], {
        for (final group in groups)
          '${group.id}':
              canCreateComments
                  ? const ['group.comments.create']
                  : const <String>[],
      });

  await tester.pumpWidget(
    MultiProvider(
      providers: [
        ChangeNotifierProvider<GroupModel>.value(value: groupModel),
        ChangeNotifierProvider<UserModel>.value(value: userModel),
        ChangeNotifierProvider<PermissionsModel>.value(value: permissionsModel),
        ChangeNotifierProvider<UserPreferencesModel>(
          create: (_) => UserPreferencesModel(),
        ),
        ChangeNotifierProvider<CategoryModel>(create: (_) => CategoryModel()),
        ChangeNotifierProvider<TagModel>(create: (_) => TagModel()),
        ChangeNotifierProvider<ContextModel>(create: (_) => ContextModel()),
        // The category/tag pickers open a fullscreen sheet whose TopAppBar reads
        // these two; needed once a test taps a picker open.
        ChangeNotifierProvider<AuthModel>(create: (_) => AuthModel()),
        ChangeNotifierProvider<LoadingModel>(create: (_) => LoadingModel()),
      ],
      child: MaterialApp(
        home: Scaffold(
          body: SingleChildScrollView(
            child: QuickScanForm(
              formKey: image.formKey,
              image: image,
              index: 0,
              onFormChangeCallback: (values) {
                // Mirrors quick_scan.dart's carousel consumer exactly.
                image.groupId = values.groupId;
                image.paidByUserId = values.paidByUserId;
                image.status = values.status;
                image.categories = values.categories;
                image.tags = values.tags;
                image.comment = values.comment;
                onFormChange?.call(values);
              },
            ),
          ),
        ),
      ),
    ),
  );
  await tester.pump();
  return image.formKey;
}

/// Convenience wrapper for the single-group tests.
Future<GlobalKey<FormBuilderState>> _pumpForm(
  WidgetTester tester, {
  required api.GroupReceiptSettings settings,
  int imageGroupId = _groupId,
  bool canCreateComments = true,
}) => _pumpFormGroups(
  tester,
  groups: [_group(settings)],
  imageGroupId: imageGroupId,
  canCreateComments: canCreateComments,
);

Finder _commentField() => find.byWidgetPredicate(
  (w) => w is FormBuilderTextField && w.name == 'comment',
);

Finder _dropdown(String name) =>
    find.byWidgetPredicate((w) => w is FormBuilderDropdown && w.name == name);

/// Opens the group dropdown and picks [name]. `.last` picks the open menu's
/// copy of the label (the closed-state selected value also renders it).
Future<void> _selectGroup(WidgetTester tester, String name) async {
  await tester.tap(_dropdown('groupId'));
  await tester.pumpAndSettle();
  await tester.tap(find.text(name).last);
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('shows/hides each field per the group config', (tester) async {
    // Paid By hidden, Status shown, Categories shown, Tags hidden.
    await _pumpForm(
      tester,
      settings: _settings(
        paidByEnabled: false,
        statusEnabled: true,
        categoriesEnabled: true,
        tagsEnabled: false,
      ),
    );

    expect(_dropdown('groupId'), findsOneWidget); // always present
    expect(_dropdown('paidByUserId'), findsNothing);
    expect(_dropdown('status'), findsOneWidget);
    expect(find.byType(CategorySelectField), findsOneWidget);
    expect(find.byType(TagSelectField), findsNothing);
  });

  testWidgets('shows every field when all are enabled', (tester) async {
    await _pumpForm(
      tester,
      settings: _settings(
        paidByEnabled: true,
        statusEnabled: true,
        categoriesEnabled: true,
        tagsEnabled: true,
      ),
    );

    expect(_dropdown('paidByUserId'), findsOneWidget);
    expect(_dropdown('status'), findsOneWidget);
    expect(find.byType(CategorySelectField), findsOneWidget);
    expect(find.byType(TagSelectField), findsOneWidget);
  });

  testWidgets('renders only the Group field when no group is selected', (
    tester,
  ) async {
    // imageGroupId 0 → no group picked → nothing but Group renders. The config
    // that WOULD apply enables every field, so this also pins that the no-group
    // rule wins over the settings.
    await _pumpForm(
      tester,
      settings: _settings(
        categoriesEnabled: true,
        tagsEnabled: true,
        commentEnabled: true,
      ),
      imageGroupId: 0,
    );

    expect(_dropdown('groupId'), findsOneWidget);
    expect(_dropdown('paidByUserId'), findsNothing);
    expect(_dropdown('status'), findsNothing);
    expect(find.byType(CategorySelectField), findsNothing);
    expect(find.byType(TagSelectField), findsNothing);
    expect(_commentField(), findsNothing);
  });

  testWidgets('keeps the paid-by and status prefills through the first group '
      'selection', (tester) async {
    // The regression guard for the prefill wipe: paid-by/status are unmounted
    // while no group is picked, so they are absent from the form value. Reporting
    // them as null would have the consumer erase the user's quickScanDefault*
    // prefill on the very interaction that reveals the fields.
    final group = _group(_settings(), members: [_member(42, _groupId)]);
    final key = await _pumpFormGroups(
      tester,
      groups: [group],
      imageGroupId: 0,
      users: [_user(42, 'Payer')],
      imagePaidByUserId: 42,
      imageStatus: api.ReceiptStatus.OPEN,
    );

    // Precondition: neither field is mounted yet.
    expect(_dropdown('paidByUserId'), findsNothing);
    expect(_dropdown('status'), findsNothing);

    await _selectGroup(tester, 'Test Group');

    expect(key.currentState!.fields['paidByUserId']!.value, 42);
    expect(key.currentState!.fields['status']!.value, api.ReceiptStatus.OPEN);
  });

  testWidgets('drops a prefilled paid-by who is not a member of the group '
      'just picked', (tester) async {
    // The other half of the prefill fix. The dropdown seeds BLANK for a payer who
    // is not a member (`valueExists` in _buildUserDropDown), so without the
    // post-frame re-sync the image would keep the invisible id and the submit
    // would send a user the caller can neither see nor have meant.
    final group = _group(_settings()); // no members
    QuickScanFormValues? last;
    final key = await _pumpFormGroups(
      tester,
      groups: [group],
      imageGroupId: 0,
      imagePaidByUserId: 42,
      onFormChange: (values) => last = values,
    );

    await _selectGroup(tester, 'Test Group');

    expect(key.currentState!.fields['paidByUserId']!.value, isNull);
    expect(
      last?.paidByUserId,
      isNull,
      reason: 'the image must not keep a payer the field cannot show',
    );
  });

  testWidgets('a group with default settings shows paid-by + status, '
      'hides categories + tags', (tester) async {
    // Non-null settings carrying the backend defaults. This is the load-bearing
    // contrast to the no-group case above: a PICKED group with these settings
    // still shows paid-by + status.
    await _pumpForm(tester, settings: _settings());

    expect(_dropdown('paidByUserId'), findsOneWidget);
    expect(_dropdown('status'), findsOneWidget);
    expect(find.byType(CategorySelectField), findsNothing);
    expect(find.byType(TagSelectField), findsNothing);
  });

  testWidgets('a shown+required field carries a required validator', (
    tester,
  ) async {
    await _pumpForm(
      tester,
      settings: _settings(paidByEnabled: true, paidByRequired: true),
    );

    final paidBy =
        tester.widget(_dropdown('paidByUserId')) as FormBuilderDropdown;
    expect(paidBy.validator, isNotNull);
    expect(
      paidBy.validator!(null),
      isNotNull,
      reason: 'empty fails validation',
    );
  });

  testWidgets('a shown+optional field has no required validator', (
    tester,
  ) async {
    await _pumpForm(
      tester,
      settings: _settings(paidByEnabled: true, paidByRequired: false),
    );

    final paidBy =
        tester.widget(_dropdown('paidByUserId')) as FormBuilderDropdown;
    expect(paidBy.validator, isNull);
  });

  testWidgets('status shown+required carries a required validator', (
    tester,
  ) async {
    await _pumpForm(
      tester,
      settings: _settings(statusEnabled: true, statusRequired: true),
    );

    final status = tester.widget(_dropdown('status')) as FormBuilderDropdown;
    expect(status.validator, isNotNull);
    expect(
      status.validator!(null),
      isNotNull,
      reason: 'empty fails validation',
    );
  });

  testWidgets('status shown+optional has no required validator', (
    tester,
  ) async {
    await _pumpForm(
      tester,
      settings: _settings(statusEnabled: true, statusRequired: false),
    );

    final status = tester.widget(_dropdown('status')) as FormBuilderDropdown;
    expect(status.validator, isNull);
  });

  testWidgets('a selected paid-by does not come back after a group that hides '
      'it', (tester) async {
    // A -> B -> C, where B HIDES paid-by and A and C both show it. The concern
    // is that B never registers the field, so the group dropdown's setValue(null)
    // cannot reach it and the image keeps the member picked under A -- which C
    // would then re-display.
    //
    // It does not happen, and the reason is worth recording: the clear runs
    // while the OLD field set is still mounted (onChanged fires before the
    // setState that re-resolves the config), so A's live field is cleared on the
    // way out. FormBuilder also runs with clearValueOnUnregister: false, so the
    // key survives B unmounting the field -- `containsKey` stays true and
    // onValueChange's fallback to the image is never taken.
    final groupA = _group(_settings(id: _groupId, paidByEnabled: true),
        name: 'Group A', members: [_member(42, _groupId)]);
    final groupB = _group(_settings(id: _group2Id, paidByEnabled: false),
        name: 'Group B', members: [_member(42, _group2Id)]);
    const group3Id = 3;
    final groupC = _group(_settings(id: group3Id, paidByEnabled: true),
        name: 'Group C', members: [_member(42, group3Id)]);

    final key = await _pumpFormGroups(
      tester,
      groups: [groupA, groupB, groupC],
      imageGroupId: 0,
      users: [_user(42, 'Payer')],
    );

    // Pick A and choose a member, so the value is a deliberate selection rather
    // than a prefill.
    await _selectGroup(tester, 'Group A');
    await tester.tap(_dropdown('paidByUserId'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Payer').last);
    await tester.pumpAndSettle();
    expect(key.currentState!.fields['paidByUserId']!.value, 42);

    // B hides it.
    await _selectGroup(tester, 'Group B');
    expect(_dropdown('paidByUserId'), findsNothing);

    // C shows it again -- and it must be empty, not Payer.
    await _selectGroup(tester, 'Group C');
    expect(_dropdown('paidByUserId'), findsOneWidget);
    expect(
      key.currentState!.fields['paidByUserId']!.value,
      isNull,
      reason: 'a member chosen under A must not survive into C',
    );
  });

  testWidgets('a prefilled paid-by DOES survive a group that hides it', (
    tester,
  ) async {
    // The deliberate counterpart to the case above, pinning the two apart. An
    // untouched quickScanDefaultPaidById is the user's standing preference, not
    // a stale selection, so a group that shows the field is meant to offer it --
    // exactly as it would have on the very first group picked.
    final groupA = _group(_settings(id: _groupId, paidByEnabled: false),
        name: 'Group A', members: [_member(42, _groupId)]);
    final groupB = _group(_settings(id: _group2Id, paidByEnabled: true),
        name: 'Group B', members: [_member(42, _group2Id)]);

    final key = await _pumpFormGroups(
      tester,
      groups: [groupA, groupB],
      imageGroupId: 0,
      users: [_user(42, 'Payer')],
      imagePaidByUserId: 42,
    );

    await _selectGroup(tester, 'Group A'); // hides paid-by; never mounted
    expect(_dropdown('paidByUserId'), findsNothing);

    await _selectGroup(tester, 'Group B'); // shows it
    expect(key.currentState!.fields['paidByUserId']!.value, 42);
  });

  testWidgets('re-renders fields per config when the group changes', (
    tester,
  ) async {
    // Group A: paid-by/status shown, categories/tags hidden.
    // Group B: the exact inverse — proves visibility is reactive to the group.
    final groupA = _group(
      _settings(
        id: _groupId,
        paidByEnabled: true,
        statusEnabled: true,
        categoriesEnabled: false,
        tagsEnabled: false,
      ),
      name: 'Group A',
    );
    final groupB = _group(
      _settings(
        id: _group2Id,
        paidByEnabled: false,
        statusEnabled: false,
        categoriesEnabled: true,
        tagsEnabled: true,
      ),
      name: 'Group B',
    );
    await _pumpFormGroups(
      tester,
      groups: [groupA, groupB],
      imageGroupId: _groupId,
    );

    // Group A's field set.
    expect(_dropdown('paidByUserId'), findsOneWidget);
    expect(_dropdown('status'), findsOneWidget);
    expect(find.byType(CategorySelectField), findsNothing);
    expect(find.byType(TagSelectField), findsNothing);

    await _selectGroup(tester, 'Group B');

    // Group B's field set — flipped.
    expect(_dropdown('paidByUserId'), findsNothing);
    expect(_dropdown('status'), findsNothing);
    expect(find.byType(CategorySelectField), findsOneWidget);
    expect(find.byType(TagSelectField), findsOneWidget);
  });

  testWidgets('clears paid-by / categories / tags when the group changes', (
    tester,
  ) async {
    // Both groups show all three group-scoped fields; group A carries the
    // member that backs a valid paid-by selection.
    final groupA = _group(
      _settings(id: _groupId, categoriesEnabled: true, tagsEnabled: true),
      name: 'Group A',
      members: [_member(42, _groupId)],
    );
    final groupB = _group(
      _settings(id: _group2Id, categoriesEnabled: true, tagsEnabled: true),
      name: 'Group B',
    );
    final formKey = await _pumpFormGroups(
      tester,
      groups: [groupA, groupB],
      imageGroupId: _groupId,
      users: [_user(42, 'Payer')],
    );

    formKey.currentState!.fields['paidByUserId']!.didChange(42);
    formKey.currentState!.fields['categories']!.didChange(<api.Category>[
      _category(),
    ]);
    formKey.currentState!.fields['tags']!.didChange(<api.Tag>[_tag()]);
    await tester.pump();
    expect(formKey.currentState!.fields['paidByUserId']!.value, 42);
    expect(formKey.currentState!.fields['categories']!.value, isNotEmpty);
    expect(formKey.currentState!.fields['tags']!.value, isNotEmpty);

    await _selectGroup(tester, 'Group B');

    expect(
      formKey.currentState!.fields['paidByUserId']!.value,
      isNull,
      reason: 'paid-by cleared on group change',
    );
    expect(
      formKey.currentState!.fields['categories']!.value,
      isEmpty,
      reason: 'categories cleared on group change',
    );
    expect(
      formKey.currentState!.fields['tags']!.value,
      isEmpty,
      reason: 'tags cleared on group change',
    );
  });

  testWidgets('re-evaluates the required validator after a group switch', (
    tester,
  ) async {
    // Paid-by optional in group A, required in group B.
    final groupA = _group(
      _settings(id: _groupId, paidByEnabled: true, paidByRequired: false),
      name: 'Group A',
    );
    final groupB = _group(
      _settings(id: _group2Id, paidByEnabled: true, paidByRequired: true),
      name: 'Group B',
    );
    await _pumpFormGroups(
      tester,
      groups: [groupA, groupB],
      imageGroupId: _groupId,
    );

    var paidBy =
        tester.widget(_dropdown('paidByUserId')) as FormBuilderDropdown;
    expect(paidBy.validator, isNull, reason: 'optional in group A');

    await _selectGroup(tester, 'Group B');

    paidBy = tester.widget(_dropdown('paidByUserId')) as FormBuilderDropdown;
    expect(paidBy.validator, isNotNull, reason: 'required in group B');
    expect(
      paidBy.validator!(null),
      isNotNull,
      reason: 'empty fails after switch',
    );
  });

  testWidgets('tapping Categories opens the picker with no shell context', (
    tester,
  ) async {
    // ContextModel.shellContext is null in this harness — the same condition as
    // the real Quick Scan flow, which never mounts the receipt-form screen that
    // sets it. Pre-fix this tap crashed via Navigator.of(null); the fix falls
    // back to the field's own (mounted) context so the picker opens.
    await _pumpForm(tester, settings: _settings(categoriesEnabled: true));

    await tester.tap(find.text('No Categories selected'));
    await tester.pumpAndSettle();

    expect(find.text('Select Categories'), findsOneWidget);
  });

  testWidgets('tapping Tags opens the picker with no shell context', (
    tester,
  ) async {
    await _pumpForm(tester, settings: _settings(tagsEnabled: true));

    await tester.tap(find.text('No Tags selected'));
    await tester.pumpAndSettle();

    expect(find.text('Select Tags'), findsOneWidget);
  });

  // Prefill (from userPreferences.quickScanDefault*) vs the group's config: the
  // per-image form seeds paid-by/status from the image, but a field the group
  // hides must not appear -- the preset "falls off" (and _submitQuickScan sends
  // the sentinel for it).
  testWidgets('a prefilled paid-by shows when the group shows paid-by', (
    tester,
  ) async {
    final group = _group(
      _settings(id: _groupId, paidByEnabled: true),
      members: [_member(42, _groupId)],
    );
    await _pumpFormGroups(
      tester,
      groups: [group],
      imageGroupId: _groupId,
      users: [_user(42, 'Payer')],
      imagePaidByUserId: 42,
    );

    final paidBy =
        tester.widget(_dropdown('paidByUserId')) as FormBuilderDropdown;
    expect(paidBy.initialValue, 42, reason: 'prefill honored on a shown field');
  });

  testWidgets('a prefilled paid-by falls off when the group hides paid-by', (
    tester,
  ) async {
    final group = _group(_settings(id: _groupId, paidByEnabled: false));
    await _pumpFormGroups(
      tester,
      groups: [group],
      imageGroupId: _groupId,
      imagePaidByUserId: 42,
    );

    expect(
      _dropdown('paidByUserId'),
      findsNothing,
      reason: 'preset paid-by fell off the hidden field',
    );
  });

  testWidgets('a prefilled status shows when the group shows status', (
    tester,
  ) async {
    final group = _group(_settings(id: _groupId, statusEnabled: true));
    await _pumpFormGroups(
      tester,
      groups: [group],
      imageGroupId: _groupId,
      imageStatus: api.ReceiptStatus.OPEN,
    );

    final status = tester.widget(_dropdown('status')) as FormBuilderDropdown;
    expect(
      status.initialValue,
      api.ReceiptStatus.OPEN,
      reason: 'prefill honored on a shown field',
    );
  });

  testWidgets('a prefilled status falls off when the group hides status', (
    tester,
  ) async {
    final group = _group(_settings(id: _groupId, statusEnabled: false));
    await _pumpFormGroups(
      tester,
      groups: [group],
      imageGroupId: _groupId,
      imageStatus: api.ReceiptStatus.OPEN,
    );

    expect(
      _dropdown('status'),
      findsNothing,
      reason: 'preset status fell off the hidden field',
    );
  });

  testWidgets('renders an optional comment field with no required validator', (
    tester,
  ) async {
    await _pumpForm(
      tester,
      settings: _settings(commentEnabled: true, commentRequired: false),
    );

    final comment = tester.widget(_commentField()) as FormBuilderTextField;
    expect(comment.validator, isNull);
  });

  testWidgets('hides the comment field by default', (tester) async {
    await _pumpForm(tester, settings: _settings());

    expect(_commentField(), findsNothing);
  });

  testWidgets('a shown+required comment carries a required validator', (
    tester,
  ) async {
    await _pumpForm(
      tester,
      settings: _settings(commentEnabled: true, commentRequired: true),
    );

    final comment = tester.widget(_commentField()) as FormBuilderTextField;
    expect(comment.validator, isNotNull);
    expect(
      comment.validator!(null),
      isNotNull,
      reason: 'empty fails validation',
    );
    expect(comment.validator!('A note'), isNull);
  });

  testWidgets('hides the comment field without group.comments.create', (
    tester,
  ) async {
    // Required in the group config, but the member cannot comment - the field
    // must be absent, so a required comment can never lock them out of quick
    // scan (the server skips the check for them too).
    await _pumpForm(
      tester,
      settings: _settings(commentEnabled: true, commentRequired: true),
      canCreateComments: false,
    );

    expect(_commentField(), findsNothing);
  });

  testWidgets('hides the comment field when the group hides comments', (
    tester,
  ) async {
    await _pumpForm(
      tester,
      settings: _settings(
        commentEnabled: true,
        commentRequired: true,
        hideComments: true,
      ),
    );

    expect(_commentField(), findsNothing);
  });

  testWidgets('keeps a typed comment when the group changes', (tester) async {
    // Unlike paid-by/categories/tags, a comment is not group-scoped data, so
    // switching groups must not silently discard what the user typed.
    await _pumpFormGroups(
      tester,
      groups: [
        _group(_settings(commentEnabled: true), name: 'Group One'),
        _group(
          _settings(id: _group2Id, commentEnabled: true),
          name: 'Group Two',
        ),
      ],
      imageGroupId: _groupId,
    );

    await tester.enterText(_commentField(), 'Kept across groups');
    await tester.pump();

    await _selectGroup(tester, 'Group Two');

    expect(_commentField(), findsOneWidget);
    expect(find.text('Kept across groups'), findsOneWidget);
  });
}
