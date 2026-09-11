import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:form_builder_validators/form_builder_validators.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/constants/spacing.dart';
import 'package:receipt_wrangler_mobile/enums/form_state.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/shared/classes/quick_scan_image.dart';
import 'package:receipt_wrangler_mobile/shared/functions/permissions.dart';
import 'package:receipt_wrangler_mobile/shared/functions/quick_scan_field_config.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/category_select_field.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/tag_select_field.dart';
import 'package:receipt_wrangler_mobile/utils/forms.dart';

import '../../models/user_preferences_model.dart';

/// The per-image values the quick-scan form reports on every change. A record
/// rather than positional arguments: the list already opened with two `int?`s and
/// is only growing, so naming each value keeps callers from transposing them.
typedef QuickScanFormValues = ({
  int? groupId,
  int? paidByUserId,
  api.ReceiptStatus? status,
  List<api.Category> categories,
  List<api.Tag> tags,
  String? comment,
});

class QuickScanForm extends StatefulWidget {
  const QuickScanForm(
      {super.key,
      required this.formKey,
      required this.image,
      required this.index,
      required this.onFormChangeCallback,
      this.enabled = true});

  final GlobalKey<FormBuilderState> formKey;
  final QuickScanImage image;
  final int index;
  final void Function(QuickScanFormValues values) onFormChangeCallback;
  final bool enabled;

  @override
  State<QuickScanForm> createState() => _QuickScanForm();
}

class _QuickScanForm extends State<QuickScanForm> {
  late final userPreferences =
      Provider.of<UserPreferencesModel>(context, listen: false).userPreferences;
  int groupId = 0;

  @override
  initState() {
    super.initState();
    groupId = widget.image.groupId ?? 0;
  }

  void onValueChange() {
    widget.formKey.currentState!.save();
    final formValue = widget.formKey.currentState!.value;

    // A field the config hides is never BUILT (Visibility defaults to
    // maintainState: false), so it never registers with FormBuilder and its key
    // is absent from `value`. FormBuilder runs with the default
    // clearValueOnUnregister: false, so a key that HAS been registered survives
    // the field being hidden later - an absent key therefore means "never
    // shown", not "cleared", and the image's own value is the right answer.
    //
    // Reporting null for an absent key would erase the user's quickScanDefault*
    // prefill on the very first group selection, because the consumer
    // (quick_scan.dart) writes every record member onto the image
    // unconditionally and this runs BEFORE the setState that reveals the fields.
    // A field that IS mounted always reports its own value, so the explicit
    // clears in the group dropdown's onChanged still take effect.
    widget.onFormChangeCallback((
      // Always built, so no fallback is needed.
      groupId: formValue["groupId"],
      paidByUserId: formValue.containsKey("paidByUserId")
          ? formValue["paidByUserId"] as int?
          : widget.image.paidByUserId,
      status: formValue.containsKey("status")
          ? formValue["status"] as api.ReceiptStatus?
          : widget.image.status,
      categories: formValue.containsKey("categories")
          ? (formValue["categories"] as List?)?.cast<api.Category>() ?? const []
          : widget.image.categories,
      tags: formValue.containsKey("tags")
          ? (formValue["tags"] as List?)?.cast<api.Tag>() ?? const []
          : widget.image.tags,
      comment: formValue.containsKey("comment")
          ? formValue["comment"] as String?
          : widget.image.comment,
    ));
  }

  // TODO: refactor to a common Widget to use in receipt form
  Widget _buildGroupField() {
    // Get the list of groups for dropdown
    final dropdownItems = buildGroupDropDownMenuItems(context);
    int? initialValue = widget.image.groupId;

    // Check if initialValue exists in the dropdown items
    bool valueExists = dropdownItems.any((item) => item.value == initialValue);

    return FormBuilderDropdown(
      name: "groupId",
      decoration: const InputDecoration(labelText: "Group"),
      items: dropdownItems,
      validator: FormBuilderValidators.required(),
      initialValue: valueExists ? initialValue : null,
      enabled: widget.enabled,
      // Set to null if value doesn't exist
      onChanged: (value) {
        // The paid-by members and category/tag catalogs are group-scoped, so
        // clear them when the group changes. Fields may be hidden for either the
        // old or new group's config, so guard each access.
        widget.formKey.currentState?.fields["paidByUserId"]?.setValue(null);
        widget.formKey.currentState?.fields["categories"]
            ?.setValue(<api.Category>[]);
        widget.formKey.currentState?.fields["tags"]?.setValue(<api.Tag>[]);
        onValueChange();
        setState(() {
          groupId = value as int;
        });
        // The new group's field set mounts on the NEXT frame, each field seeding
        // itself from the image. Report again once it exists, so the image
        // matches what the user can actually see: a prefilled paid-by who is not
        // a member of the group just picked seeds the dropdown BLANK
        // (`valueExists` in _buildUserDropDown), and without this the image would
        // keep the invisible id and _submitQuickScan would send it. Safe to
        // re-enter: onValueChange does not setState.
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) {
            onValueChange();
          }
        });
      },
    );
  }

  Widget _buildUserDropDown(bool required) {
    List<DropdownMenuItem> items = [];
    int? initialValue = widget.image.paidByUserId;

    if (groupId > 0) {
      items = buildGroupMemberDropDownMenuItems(context, groupId.toString());
    }

    // Check if initialValue exists in the dropdown items
    bool valueExists = items.any((item) => item.value == initialValue);

    return FormBuilderDropdown(
      name: "paidByUserId",
      decoration: const InputDecoration(labelText: "Paid By"),
      items: items,
      validator: required ? FormBuilderValidators.required() : null,
      initialValue: valueExists ? initialValue : null,
      enabled: widget.enabled,
      // Set to null if value doesn't exist
      onChanged: (value) {
        onValueChange();
      },
    );
  }

  Widget _buildStatusDropdown(bool required) {
    api.ReceiptStatus? initialValue = widget.image.status;
    final items = buildReceiptStatusDropDownMenuItems();

    // Check if initialValue exists in the dropdown items
    bool valueExists = items.any((item) => item.value == initialValue);

    return FormBuilderDropdown(
      name: "status",
      decoration: const InputDecoration(labelText: "Status"),
      items: items,
      validator: required ? FormBuilderValidators.required() : null,
      initialValue: valueExists ? initialValue : null,
      enabled: widget.enabled,
      // Set to null if value doesn't exist
      onChanged: (value) {
        onValueChange();
      },
    );
  }

  Widget _buildCategoryField() {
    return CategorySelectField(
      fieldName: "categories",
      label: "Categories",
      groupId: groupId,
      initialCategories:
          (widget.formKey.currentState?.fields["categories"]?.value
                  as List<api.Category>?) ??
              widget.image.categories,
      formState: WranglerFormState.add,
      onCategoriesChanged: (categories) {
        setState(() {
          widget.formKey.currentState!.fields["categories"]!
              .setValue(categories);
        });
        onValueChange();
      },
    );
  }

  Widget _buildTagField() {
    return TagSelectField(
      fieldName: "tags",
      label: "Tags",
      groupId: groupId,
      initialTags: (widget.formKey.currentState?.fields["tags"]?.value
              as List<api.Tag>?) ??
          widget.image.tags,
      formState: WranglerFormState.add,
      onTagsChanged: (tags) {
        setState(() {
          widget.formKey.currentState!.fields["tags"]!.setValue(tags);
        });
        onValueChange();
      },
    );
  }

  Widget _buildCommentField(bool required) {
    return FormBuilderTextField(
      name: "comment",
      decoration: const InputDecoration(labelText: "Comment"),
      // A receipt note rather than a one-liner; the length cap matches the
      // backend's models.MaxCommentLength, which rejects anything longer.
      maxLines: 3,
      maxLength: 500,
      initialValue: widget.image.comment,
      validator: required ? FormBuilderValidators.required() : null,
      enabled: widget.enabled,
      onChanged: (value) {
        onValueChange();
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    // Field visibility/requirement follows the selected group's quick-scan
    // config. When paid-by/status is not shown+required the server backfills a
    // configured default, so the field can be omitted here. With no group picked
    // yet only the Group dropdown renders - there is no config to honour, and
    // guessing one means flipping the field set the moment a group is chosen.
    final settings = Provider.of<GroupModel>(context, listen: false)
        .getGroupReceiptSettings(groupId);
    final permissionsModel =
        Provider.of<PermissionsModel>(context, listen: false);

    final config = resolveQuickScanFieldConfig(
      settings,
      hasGroup: groupId > 0,
      canCreateComments: canCommentCreate(permissionsModel, groupId),
    );
    final showPaidBy = config.showPaidBy;
    final requirePaidBy = config.requirePaidBy;
    final showStatus = config.showStatus;
    final requireStatus = config.requireStatus;
    final showCategories = config.showCategories;
    final showTags = config.showTags;
    final showComment = config.showComment;
    final requireComment = config.requireComment;

    return FormBuilder(
        key: widget.formKey,
        child: Column(
          children: [
            textFieldSpacing,
            _buildGroupField(),
            Visibility(
              visible: showPaidBy,
              child: Column(children: [
                textFieldSpacing,
                _buildUserDropDown(requirePaidBy),
              ]),
            ),
            Visibility(
              visible: showStatus,
              child: Column(children: [
                textFieldSpacing,
                _buildStatusDropdown(requireStatus),
              ]),
            ),
            Visibility(
              visible: showCategories,
              child: Column(children: [
                textFieldSpacing,
                _buildCategoryField(),
              ]),
            ),
            Visibility(
              visible: showTags,
              child: Column(children: [
                textFieldSpacing,
                _buildTagField(),
              ]),
            ),
            Visibility(
              visible: showComment,
              child: Column(children: [
                textFieldSpacing,
                _buildCommentField(requireComment),
              ]),
            ),
            submitButtonSpacing
          ],
        ));
  }
}
