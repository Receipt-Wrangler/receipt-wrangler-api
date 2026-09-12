import 'package:built_collection/built_collection.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_svg/svg.dart';
import 'package:form_builder_validators/form_builder_validators.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/constants/spacing.dart';
import 'package:receipt_wrangler_mobile/enums/form_state.dart';
import 'package:receipt_wrangler_mobile/receipts/widgets/quick_actions.dart';
import 'package:receipt_wrangler_mobile/receipts/widgets/quick_actions_submit_button.dart';
import 'package:receipt_wrangler_mobile/receipts/widgets/receipt_item_list.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/amount_field.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/tag_select_field.dart';
import 'package:receipt_wrangler_mobile/utils/bottom_sheet.dart';
import 'package:receipt_wrangler_mobile/utils/date.dart';
import 'package:receipt_wrangler_mobile/utils/forms.dart';
import 'package:receipt_wrangler_mobile/utils/group.dart';

import '../../interfaces/form_item.dart';
import '../../models/category_model.dart';
import '../../models/context_model.dart';
import '../../models/custom_field_model.dart';
import '../../models/group_model.dart';
import '../../models/receipt_model.dart';
import '../../models/tag_model.dart';
import '../../shared/functions/custom_field_values.dart';
import '../../shared/functions/forms.dart';
import '../../shared/functions/status_field.dart';
import '../../shared/widgets/audit_detail_section.dart';
import '../../shared/widgets/category_select_field.dart';
import '../../shared/widgets/custom_field_widget.dart';
import '../screens/receipt_image_screen.dart';
import '../screens/receipt_comment_screen.dart';

class ReceiptForm extends StatefulWidget {
  const ReceiptForm({super.key});

  @override
  State<ReceiptForm> createState() => _ReceiptForm();
}

class _ReceiptForm extends State<ReceiptForm> {
  late final receipt =
      Provider.of<ReceiptModel>(context, listen: false).receipt;
  late final receiptModel = Provider.of<ReceiptModel>(context, listen: false);
  late final formKey =
      Provider.of<ReceiptModel>(context, listen: false).receiptFormKey;
  late final groupModel = Provider.of<GroupModel>(context, listen: false);
  late final formState = getFormStateFromContext(context);
  late final categoryModel = Provider.of<CategoryModel>(context, listen: false);
  late final tagModel = Provider.of<TagModel>(context, listen: false);
  late final customFieldModel = Provider.of<CustomFieldModel>(context, listen: false);
  final addSharesFormKey = GlobalKey<FormBuilderState>();
  int groupId = 0;
  bool isAddingShare = false;

  /// The custom fields this form attached on the user's behalf because the
  /// selected group declares them as defaults. Only these are candidates for
  /// removal when the group changes -- anything the user added by hand, or
  /// typed a value into, is theirs (see [_applyGroupDefaultCustomFields]).
  final Set<int> _autoAppliedCustomFieldIds = {};

  @override
  void initState() {
    super.initState();
    Provider.of<ReceiptModel>(context, listen: false);

    groupId = modifiedReceipt.groupId;

    // Always scheduled, not just when the receipt already knows its group: on an
    // add form _resolveInitialGroupId can produce one the receipt does not carry
    // (the group being browsed, or the user's only group). Deferred to after the
    // first frame because applying the group's defaults mutates ReceiptModel, and
    // notifying its listeners while the tree is still building throws -- and
    // because `formState`/`getGroupId` read the route, which needs a mounted
    // element.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || formState != WranglerFormState.add) {
        return;
      }

      // Mirrors what buildGroupField seeded the dropdown with.
      final resolvedGroupId = _resolveInitialGroupId();
      if (resolvedGroupId == 0) {
        return;
      }

      // Before the setState below: it writes to ReceiptModel, and
      // notifyListeners() must not fire from inside a setState callback.
      _applyGroupDefaultCustomFields(resolvedGroupId);

      if (resolvedGroupId != groupId) {
        // Carry the seed into the State field the group-derived parts of the
        // form read -- the paid-by members, the category/tag pickers and the
        // add-share button all stay dead at 0.
        setState(() {
          groupId = resolvedGroupId;
        });
      }
    });
  }

  /// The group the form starts on: the receipt's own group, else -- on an add
  /// form -- the group being browsed, else the user's only group. Zero when they
  /// genuinely have to pick one.
  ///
  /// Pure and stable across rebuilds, so the dropdown's `initialValue` and the
  /// [groupId] State field cannot disagree. Reads the route, so it must not be
  /// called from `initState`.
  int _resolveInitialGroupId() {
    if (modifiedReceipt.groupId > 0) {
      return modifiedReceipt.groupId;
    }
    if (formState != WranglerFormState.add) {
      return 0;
    }

    // The group the user came from -- `openManualReceipt` puts it on the route's
    // `extra`. The synthetic "All" group is not a receipt target, so it does not
    // count (mirrors desktop's initForm).
    final routeGroup = groupModel.getGroupById(getGroupId(context));
    if (routeGroup != null && !routeGroup.isAllGroup) {
      return routeGroup.id;
    }

    return groupModel.soleGroupId ?? 0;
  }

  Widget buildAuditDetailSection() {
    if (formState == WranglerFormState.add) {
      return SizedBox.shrink();
    }

    return AuditDetailSection(entity: receipt);
  }

  Widget buildNameField() {
    return FormBuilderTextField(
      name: "name",
      decoration: const InputDecoration(labelText: "Name"),
      initialValue: modifiedReceipt.name,
      validator: FormBuilderValidators.required(),
      readOnly: isFieldReadOnly(formState),
    );
  }

  Widget buildAmountField() {
    return AmountField(
        label: "Amount",
        fieldName: "amount",
        initialAmount: modifiedReceipt.amount.toString(),
        formState: formState);
  }

  Widget buildDateField() {
    if (formState == WranglerFormState.view) {
      var formattedDate =
          formatDate(defaultDateFormat, DateTime.parse(modifiedReceipt.date));
      // NOTE: distinct field name from the edit-mode picker below.
      // Both used to be named "date", which made FormBuilder's
      // internal value map collide across mode transitions: the
      // TextField stored a String there, then the DateTimePicker
      // would re-register and FormBuilder's `setValue(_instantValue[name])`
      // would feed that String into a `setValue(DateTime?)` call,
      // throwing `type 'String' is not a subtype of type 'DateTime?'
      // of 'value'`. The view-mode field is read-only and not part of
      // save, so it's safe to use a separate name.
      return FormBuilderTextField(
          name: "dateDisplay",
          decoration: const InputDecoration(labelText: "Date"),
          initialValue: formattedDate,
          readOnly: true);
    } else {
      return FormBuilderDateTimePicker(
        name: "date",
        decoration: const InputDecoration(labelText: "Date"),
        validator: FormBuilderValidators.required(),
        initialValue: DateTime.parse(modifiedReceipt.date),
        inputType: InputType.date,
      );
    }
  }

  Widget buildGroupField() {
    // Zero means "the user really does have to pick"; the dropdown wants null.
    final resolvedGroupId = _resolveInitialGroupId();
    final int? initialValue = resolvedGroupId == 0 ? null : resolvedGroupId;

    return FormBuilderDropdown(
      name: "groupId",
      decoration: const InputDecoration(labelText: "Group"),
      items: buildGroupDropDownMenuItems(context),
      initialValue: initialValue,
      enabled: !isFieldReadOnly(formState),
      validator: FormBuilderValidators.required(),
      onChanged: (value) {
        var newGroupId = value as int;

        // Each group is effectively its own receipt template, so re-apply the
        // target group's default custom fields. Done BEFORE the setState below
        // because it writes to ReceiptModel, and notifyListeners() must not run
        // from inside a setState callback.
        _applyGroupDefaultCustomFields(newGroupId);

        setState(() {
          // The paid-by members are group-scoped, so clear the selection when
          // the group changes. Guard the access: the field may not be mounted
          // (matching quick_scan_form.dart's group dropdown).
          formKey.currentState?.fields["paidByUserId"]?.setValue(null);
          groupId = newGroupId;
        });
      },
    );
  }

  Widget buildPaidByField() {
    List<DropdownMenuItem> items = [];
    var initialValue = null;
    if (groupId == modifiedReceipt.groupId) {
      initialValue = modifiedReceipt.paidByUserId;
    }

    if (groupId > 0) {
      items = buildGroupMemberDropDownMenuItems(context, groupId.toString());
    }

    if (formState == WranglerFormState.add && initialValue == 0) {
      initialValue = null;
    }

    return FormBuilderDropdown(
      name: "paidByUserId",
      decoration: const InputDecoration(labelText: "Paid By"),
      items: items.toList(),
      initialValue: initialValue,
      validator: FormBuilderValidators.required(),
      enabled: !isFieldReadOnly(formState),
    );
  }

  Widget buildStatusField() {
    return receiptStatusField(
        "Status", "status", modifiedReceipt.status, formState);
  }

  Widget buildCategoryField() {
    return CategorySelectField(
      fieldName: "categories",
      label: "Categories",
      groupId: groupId,
      initialCategories: formKey.currentState?.fields["categories"]?.value ??
          modifiedReceipt.categories!.toList(),
      formState: formState,
      onCategoriesChanged: (categories) => {
        setState(() {
          formKey.currentState!.fields["categories"]!.setValue(categories);
        }),
      },
    );
  }

  Widget buildTagField() {
    return TagSelectField(
        label: "Tags",
        fieldName: "tags",
        groupId: groupId,
        initialTags: formKey.currentState?.fields["tags"]?.value ??
            modifiedReceipt.tags!.toList(),
        formState: formState,
        onTagsChanged: (tags) => {
              setState(() {
                formKey.currentState!.fields["tags"]!.setValue(tags);
              })
            });
  }

  Widget buildCustomFieldsSection() {
    return Column(
      children: [
        // Render existing custom fields
        ...modifiedReceipt.customFields.map((customFieldValue) {
          final customField = customFieldModel.customFields
              .where((cf) => cf.id == customFieldValue.customFieldId)
              .firstOrNull;
          
          // Show loading placeholder if custom field template is not found but still loading
          if (customField == null) {
            if (customFieldModel.isLoading) {
              return Container(
                margin: const EdgeInsets.only(bottom: 16),
                child: const Row(
                  children: [
                    SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                    SizedBox(width: 8),
                    Text('Loading custom field...'),
                  ],
                ),
              );
            }
            return const SizedBox.shrink();
          }

          return Column(
            // Keyed by custom field id so the element tree follows identity,
            // not list position. Without it, removing a field re-uses its
            // element for whatever field shifted into its slot: FormBuilder
            // re-registers the field under the new name, but a
            // FormBuilderTextField has no didUpdateWidget and keeps the old
            // TextEditingController text -- so a group swap would show one
            // field's text under another field's label.
            key: ValueKey("customFieldValue_${customFieldValue.customFieldId}"),
            children: [
              CustomFieldWidget(
                customField: customField,
                existingValue: customFieldValue,
                formState: formState,
                onRemove: formState != WranglerFormState.view
                    ? () {
                        _removeCustomField(customFieldValue.customFieldId);
                      }
                    : null,
              ),
              textFieldSpacing,
            ],
          );
        }).toList(),
        
        // Add Custom Field button
        if (formState != WranglerFormState.view)
          buildAddCustomFieldButton(),
      ],
    );
  }

  Widget buildAddCustomFieldButton() {
    // Get custom field IDs that are already added
    final addedCustomFieldIds = modifiedReceipt.customFields
        .map((cfv) => cfv.customFieldId)
        .toSet();
    
    // Get available custom fields that haven't been added yet
    final availableCustomFields = customFieldModel.customFields
        .where((cf) => !addedCustomFieldIds.contains(cf.id))
        .toList();

    if (availableCustomFields.isEmpty) {
      return SizedBox.shrink();
    }

    return Align(
      alignment: Alignment.centerLeft,
      child: TextButton.icon(
        onPressed: () {
          showModalBottomSheet(
            context: context,
            builder: (context) => buildCustomFieldSelectionSheet(availableCustomFields),
          );
        },
        icon: Icon(Icons.add, color: Theme.of(context).primaryColor),
        label: Text(
          'Add Custom Field',
          style: TextStyle(color: Theme.of(context).primaryColor),
        ),
      ),
    );
  }

  Widget buildCustomFieldSelectionSheet(List<api.CustomField> availableCustomFields) {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Select Custom Field',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 16),
          Flexible(
            child: ListView.builder(
              shrinkWrap: true,
              itemCount: availableCustomFields.length,
              itemBuilder: (context, index) {
                final customField = availableCustomFields[index];
                return ListTile(
                  title: Text(customField.name),
                  subtitle: customField.description != null
                      ? Text(customField.description!)
                      : null,
                  trailing: Text(customField.type.name),
                  onTap: () {
                    _addCustomField(customField.id);
                    Navigator.of(context).pop();
                  },
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  /// A brand-new, empty value row for [customFieldId]. Empty is a meaningful
  /// state, not a placeholder -- see
  /// `lib/shared/functions/custom_field_values.dart`.
  api.CustomFieldValue _buildEmptyCustomFieldValue(int customFieldId) {
    return (api.CustomFieldValueBuilder()
          ..id = 0  // Use 0 for new custom field values
          ..customFieldId = customFieldId
          ..receiptId = modifiedReceipt.id
          ..createdAt = DateTime.now().toIso8601String()  // Set current timestamp
          ..createdBy = 0  // Placeholder for user ID
          ..createdByString = ''  // Empty string placeholder
          ..updatedAt = '')  // Empty string placeholder
        .build();
  }

  /// The receipt's custom field values with [remove] dropped and [add]
  /// appended, keeping the order of the values already attached.
  ///
  /// Pure on purpose: every caller hands the whole result to a single
  /// [_setCustomFieldValues], so a multi-field change (the group-default swap)
  /// costs one model write instead of one per field.
  List<api.CustomFieldValue> _customFieldValuesWith({
    Set<int> add = const {},
    Set<int> remove = const {},
  }) {
    return [
      ...modifiedReceipt.customFields
          .where((cfv) => !remove.contains(cfv.customFieldId)),
      ...add.map(_buildEmptyCustomFieldValue),
    ];
  }

  /// Clears the form value backing [customFieldId] BEFORE its field unmounts.
  /// FormBuilder's `clearValueOnUnregister` defaults to false, so an
  /// unregistered field's value stays in the form's value map and is handed
  /// straight back to any field that later re-registers under the same name --
  /// a re-added custom field would silently resurrect what was typed into it.
  void _clearCustomFieldFormValue(int customFieldId) {
    formKey.currentState?.fields[customFieldFormFieldName(customFieldId)]
        ?.didChange(null);
  }

  /// Writes [values] back to the receipt being edited. The form reads the model
  /// with listen:false, so rebuild it explicitly -- otherwise a newly added
  /// custom-field widget only appears after some unrelated rebuild.
  void _setCustomFieldValues(List<api.CustomFieldValue> values) {
    receiptModel.setModifiedReceipt(
        modifiedReceipt.rebuild((b) => b..customFields = ListBuilder(values)));
    setState(() {});
  }

  void _addCustomField(int customFieldId) {
    // Adding a field by hand makes it the user's, so the group-default swap
    // must never take it back out from under them.
    _autoAppliedCustomFieldIds.remove(customFieldId);

    _setCustomFieldValues(_customFieldValuesWith(add: {customFieldId}));
  }

  void _removeCustomField(int customFieldId) {
    _clearCustomFieldFormValue(customFieldId);

    // Removing a field by hand is a decision the swap must not fight: forget
    // that this form ever auto-applied it.
    _autoAppliedCustomFieldIds.remove(customFieldId);

    _setCustomFieldValues(_customFieldValuesWith(remove: {customFieldId}));
  }

  /// Applies [newGroupId]'s default custom fields to the form, swapping out the
  /// ones the previously selected group put there. Each group is effectively
  /// its own receipt template, so this runs on every group change (and once on
  /// load for an add form that already knows its group).
  ///
  /// The swap is deliberately conservative. It only removes a field this form
  /// added itself ([_autoAppliedCustomFieldIds]) that is still **empty**; a
  /// default the user typed into is kept and handed over to them (dropped from
  /// the auto set), as is anything they added by hand.
  ///
  /// Ids missing from the [CustomFieldModel] catalog are skipped. That catalog
  /// is empty for a caller without `app.custom-fields.read` (the 403 is
  /// swallowed into an empty list), which is exactly the gate we want here: the
  /// backend's `enforceReceiptCustomFieldSelection` would 403 their save.
  void _applyGroupDefaultCustomFields(int newGroupId) {
    if (formState == WranglerFormState.view) {
      return;
    }

    var knownCustomFieldIds =
        customFieldModel.customFields.map((cf) => cf.id).toSet();
    var defaultIds = <int>[
      ...?groupModel.getGroupReceiptSettings(newGroupId)?.defaultCustomFieldIds
    ].where(knownCustomFieldIds.contains).toSet();

    var toRemove = <int>{};
    for (var customFieldId in _autoAppliedCustomFieldIds.toList()) {
      if (defaultIds.contains(customFieldId)) {
        continue;
      }

      // Ours to take back only while it is untouched; the moment it holds a
      // value it is the user's data and stays, unowned by the swap.
      if (_isCustomFieldEmpty(customFieldId)) {
        toRemove.add(customFieldId);
      }
      _autoAppliedCustomFieldIds.remove(customFieldId);
    }

    var attachedIds =
        modifiedReceipt.customFields.map((cfv) => cfv.customFieldId).toSet();
    var toAdd = defaultIds.difference(attachedIds);
    _autoAppliedCustomFieldIds.addAll(toAdd);

    if (toRemove.isEmpty && toAdd.isEmpty) {
      return;
    }

    for (var customFieldId in toRemove) {
      _clearCustomFieldFormValue(customFieldId);
    }

    // ONE model write for the whole swap: setModifiedReceipt notifies its
    // listeners and _setCustomFieldValues rebuilds the form, so adding and
    // removing field by field would rebuild the form repeatedly with its
    // fields half-mounted.
    _setCustomFieldValues(_customFieldValuesWith(add: toAdd, remove: toRemove));
  }

  /// Whether the mounted form field for [customFieldId] currently holds no
  /// meaningful value.
  ///
  /// Type-aware: a BOOLEAN left unchecked counts as empty. `CustomFieldWidget`
  /// seeds checkboxes with `false`, so a plain null check would call every
  /// checkbox non-empty. Mirrors the desktop emptiness rule (empty <=> every
  /// value column null-or-"" and `booleanValue` falsy).
  bool _isCustomFieldEmpty(int customFieldId) {
    final value = formKey
        .currentState?.fields[customFieldFormFieldName(customFieldId)]?.value;

    final customField = customFieldModel.customFields
        .where((cf) => cf.id == customFieldId)
        .firstOrNull;

    if (value is bool || customField?.type == api.CustomFieldType.BOOLEAN) {
      return value != true;
    }

    return value == null || (value is String && value.isEmpty);
  }

  Widget buildReceiptItemList() {
    return ReceiptItemField(
      groupId: groupId,
    );
  }

  Widget buildDetailsHeader() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        buildHeaderText("Details"),
        Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            _buildCompactActionButton(
              icon: Icons.image_outlined,
              label: "Images",
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => ReceiptImageScreen(formState: formState),
                  ),
                );
              },
            ),
            const SizedBox(width: 12), // Increased spacing to prevent mis-taps
            _buildCompactActionButton(
              icon: Icons.chat_bubble_outline,
              label: "Comments", 
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => ReceiptCommentScreen(formState: formState),
                  ),
                );
              },
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildCompactActionButton({
    required IconData icon,
    required String label,
    required VoidCallback onTap,
  }) {
    return Tooltip(
      message: 'View $label',
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(24),
          child: Container(
            // Minimum 48dp touch target for mobile accessibility
            constraints: const BoxConstraints(minHeight: 48, minWidth: 48),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surface,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(
                color: Theme.of(context).colorScheme.outline.withValues(alpha: 0.3),
                width: 1,
              ),
              // Add subtle shadow for better depth perception
              boxShadow: [
                BoxShadow(
                  color: Theme.of(context).colorScheme.shadow.withValues(alpha: 0.1),
                  blurRadius: 2,
                  offset: const Offset(0, 1),
                ),
              ],
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  icon,
                  size: 20, // Increased from 16 for better visibility
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
                const SizedBox(width: 6), // Increased spacing
                Text(
                  label,
                  style: TextStyle(
                    fontSize: 14, // Increased from 12 for better readability
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }


  Widget buildSharesHeader() {
    var rowChildren = [
      buildHeaderText("Shares"),
    ];

    if (formState != WranglerFormState.view) {
      rowChildren.add(Row(
        children: [
          IconButton(
            icon: Icon(Icons.add, color: Theme.of(context).primaryColor),
            onPressed: (isAddingShare || groupId == 0)
                ? null
                : () {
                    setState(() {
                      isAddingShare = true;
                    });
                  },
          ),
          IconButton(
              onPressed: openQuickActionsBottomSheet,
              icon: SvgPicture.asset(
                "assets/icons/split.svg",
                width: 24,
                height: 24,
              ))
        ],
      ));
    }

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        ...rowChildren,
      ],
    );
  }

  void openQuickActionsBottomSheet() {
    receiptModel.resetQuickActionsFormKey();
    // Resolve a mounted context for the modal sheet at tap time. The shell
    // context cached in ContextModel by the form screen can refer to a
    // deactivated element after navigating between /view and /edit, and
    // showModalBottomSheet dereferences context.widget -- which throws a
    // null-check on a defunct element. resolveSheetContext re-reads it (no
    // stale snapshot) and falls back to this form's own always-mounted context.
    final sheetContext = Provider.of<ContextModel>(context, listen: false)
        .resolveSheetContext(context);
    showFullscreenBottomSheet(
        sheetContext,
        ReceiptQuickActions(
          groupId: groupId,
        ),
        "Quick Actions",
        bottomSheetWidget: const ReceiptQuickActionsSubmitButton());
  }

  Widget buildAddSharesCard() {
    if (isAddingShare) {
      return Card(
        color: Colors.white,
        surfaceTintColor: Colors.white,
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Column(
            children: [
              Row(children: [
                Text(
                  "Add Share",
                  style: TextStyle(fontSize: 15, fontWeight: FontWeight.bold),
                ),
              ]),
              textFieldSpacing,
              FormBuilder(
                key: addSharesFormKey,
                child: Column(
                  children: [
                    FormBuilderDropdown(
                      name: "chargedToUserId",
                      decoration:
                          const InputDecoration(labelText: "Shared With"),
                      items: buildGroupMemberDropDownMenuItems(
                          context, groupId.toString()),
                      initialValue: "",
                      validator: FormBuilderValidators.required(),
                      enabled: !isFieldReadOnly(formState),
                    ),
                    textFieldSpacing,
                    FormBuilderTextField(
                      name: "name",
                      decoration: InputDecoration(labelText: "Name"),
                      validator: FormBuilderValidators.required(),
                    ),
                    textFieldSpacing,
                    AmountField(
                        label: "Amount",
                        fieldName: "amount",
                        initialAmount: "0.00",
                        formState: formState),
                    Row(
                      mainAxisSize: MainAxisSize.max,
                      children: [
                        Expanded(
                          child: ElevatedButton(
                            child: Text("Add Share"),
                            onPressed: () {
                              if (!addSharesFormKey.currentState!
                                  .saveAndValidate()) {
                                return;
                              }
                              var form = addSharesFormKey.currentState!.value;

                              var items = [...receiptModel.items];
                              var newItem = (api.ItemBuilder()
                                    ..name = form["name"]
                                    ..amount = form["amount"]
                                    ..chargedToUserId = form["chargedToUserId"]
                                    ..receiptId = receipt?.id ?? 0
                                    ..status = api.ItemStatus.OPEN)
                                  .build();

                              items.add(FormItem.fromItem(newItem));

                              receiptModel.setItems(items);
                              setState(() {
                                isAddingShare = false;
                              });
                            },
                          ),
                        )
                      ],
                    ),
                    Row(
                      children: [
                        Expanded(
                          child: TextButton(
                            child: Text("Cancel"),
                            onPressed: () {
                              setState(() {
                                isAddingShare = false;
                              });
                            },
                          ),
                        )
                      ],
                    ),
                  ],
                ),
              )
              // user field,
              // item name
              // amount field,
            ],
          ),
        ),
      );
    } else {
      return SizedBox.shrink();
    }
  }

  // Get the current modified receipt from the model
  api.Receipt get modifiedReceipt => receiptModel.modifiedReceipt;

  @override
  Widget build(BuildContext context) {
    return Consumer<ReceiptModel>(
      builder: (context, receiptModel, child) {
        return FormBuilder(
          key: formKey,
          child: Column(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          buildAuditDetailSection(),
          textFieldSpacing,
          buildDetailsHeader(),
          textFieldSpacing,
          buildNameField(),
          textFieldSpacing,
          buildAmountField(),
          textFieldSpacing,
          buildDateField(),
          textFieldSpacing,
          buildGroupField(),
          textFieldSpacing,
          buildPaidByField(),
          textFieldSpacing,
          buildStatusField(),
          textFieldSpacing,
          buildCustomFieldsSection(),
          textFieldSpacing,
          Visibility(
              visible: groupModel
                      .getGroupReceiptSettings(groupId)
                      ?.hideReceiptCategories ==
                  false,
              child: Column(children: [
                buildCategoryField(),
                textFieldSpacing,
              ])),
          Visibility(
              visible: groupModel
                      .getGroupReceiptSettings(groupId)
                      ?.hideReceiptTags ==
                  false,
              child: Column(children: [
                buildTagField(),
                textFieldSpacing,
              ])),
          buildSharesHeader(),
          textFieldSpacing,
          buildAddSharesCard(),
          textFieldSpacing,
          buildReceiptItemList(),
          textFieldSpacing,
          kDebugMode
              ? ElevatedButton(
                  onPressed: () => {
                        if (formKey.currentState!.saveAndValidate())
                          {print(formKey.currentState!.value)}
                      },
                  child: Text("Check form value"))
              : SizedBox.shrink(),
        ],
          ),
        );
      },
    );
  }
}
