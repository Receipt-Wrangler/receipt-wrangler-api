import 'package:openapi/openapi.dart';

/// Resolved show/require flags for the five quick-scan fields (paid-by, status,
/// categories, tags, comment), derived from a group's [GroupReceiptSettings].
/// Mirrors the backend's `resolveQuickScanFields` defaults so the per-image form
/// (visibility + validators) and the submit path (which fields to send / require)
/// stay aligned.
///
/// With no group selected the form shows only the Group field - see [noGroupQuickScanFieldConfig].
/// Null [settings] with a group that IS selected still falls back to the backend
/// defaults (paid-by/status shown, categories/tags/comment hidden).
class QuickScanFieldConfig {
  final bool showPaidBy;
  final bool requirePaidBy;
  final bool showStatus;
  final bool requireStatus;
  final bool showCategories;
  final bool requireCategories;
  final bool showTags;
  final bool requireTags;
  final bool showComment;
  final bool requireComment;

  const QuickScanFieldConfig({
    required this.showPaidBy,
    required this.requirePaidBy,
    required this.showStatus,
    required this.requireStatus,
    required this.showCategories,
    required this.requireCategories,
    required this.showTags,
    required this.requireTags,
    required this.showComment,
    required this.requireComment,
  });
}

/// Every field hidden and none required. Until a group is picked there is no
/// configuration to honour, so the form renders only the Group dropdown rather
/// than guessing a field set it would have to flip the moment a group is chosen.
const noGroupQuickScanFieldConfig = QuickScanFieldConfig(
  showPaidBy: false,
  requirePaidBy: false,
  showStatus: false,
  requireStatus: false,
  showCategories: false,
  requireCategories: false,
  showTags: false,
  requireTags: false,
  showComment: false,
  requireComment: false,
);

/// [canCreateComments] is whether the caller holds `group.comments.create` in the
/// target group. It acts as an extra AND on the comment field's "enabled": without
/// it the field is hidden, is never required (so a member who cannot comment is
/// never locked out of quick scan), and a comment sent anyway is dropped by the
/// server. It is passed in rather than read from a provider so this helper stays
/// pure, and is a required named argument so both call sites - the form and the
/// submit - are forced to supply it and cannot drift apart.
/// [hasGroup] is whether the user has actually picked a group, and is deliberately
/// NOT derived from `settings == null`: a group id we cannot resolve settings for
/// (a stale `quickScanDefaultGroupId`, or AppData not carrying that group) is still
/// a choice, so it keeps the backend-mirroring defaults below rather than collapsing
/// the form to the Group field alone. Required, like [canCreateComments], so both
/// call sites - the form and the submit - must supply it and cannot drift.
QuickScanFieldConfig resolveQuickScanFieldConfig(
  GroupReceiptSettings? settings, {
  required bool hasGroup,
  required bool canCreateComments,
}) {
  if (!hasGroup) {
    return noGroupQuickScanFieldConfig;
  }

  final showPaidBy = settings?.quickScanPaidByEnabled ?? true;
  final showStatus = settings?.quickScanStatusEnabled ?? true;
  final showCategories = settings?.quickScanCategoriesEnabled ?? false;
  final showTags = settings?.quickScanTagsEnabled ?? false;
  // hideComments hides comments for the whole group, so it hides the quick-scan
  // comment too - without changing the stored toggle (mirrors the backend's
  // GroupReceiptSettings.IsQuickScanCommentShown).
  final showComment = (settings?.quickScanCommentEnabled ?? false) &&
      !(settings?.hideComments ?? false) &&
      canCreateComments;

  return QuickScanFieldConfig(
    showPaidBy: showPaidBy,
    requirePaidBy: showPaidBy && (settings?.quickScanPaidByRequired ?? true),
    showStatus: showStatus,
    requireStatus: showStatus && (settings?.quickScanStatusRequired ?? true),
    showCategories: showCategories,
    requireCategories:
        showCategories && (settings?.quickScanCategoriesRequired ?? false),
    showTags: showTags,
    requireTags: showTags && (settings?.quickScanTagsRequired ?? false),
    showComment: showComment,
    requireComment:
        showComment && (settings?.quickScanCommentRequired ?? false),
  );
}
