import { GroupReceiptSettings } from "../../open-api";

// Resolved show/require flags for the five quick-scan fields (paid-by, status,
// categories, tags, comment), derived from a group's GroupReceiptSettings.
// Mirrors the backend's resolveQuickScanFields defaults -- and mobile's
// resolveQuickScanFieldConfig (mobile/lib/shared/functions/quick_scan_field_config.dart),
// which this is a line-for-line port of -- so field visibility, the required
// validators and the submitted payload can't drift from each other.
export interface QuickScanFieldConfig {
  showPaidBy: boolean;
  requirePaidBy: boolean;
  showStatus: boolean;
  requireStatus: boolean;
  showCategories: boolean;
  requireCategories: boolean;
  showTags: boolean;
  requireTags: boolean;
  showComment: boolean;
  requireComment: boolean;
}

// Every field hidden and none required. Until a group is picked there is no
// configuration to honour, so the form renders only the Group field rather than
// guessing a field set it would have to flip the moment a group is chosen.
export const NO_GROUP_QUICK_SCAN_FIELDS: QuickScanFieldConfig = Object.freeze({
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
});

export interface QuickScanFieldConfigOptions {
  // Whether the user has actually picked a group. Deliberately NOT derived from
  // `settings === undefined`: a group id we cannot resolve settings for -- a stale
  // quickScanDefaultGroupId, or AppData not carrying that group -- is still a
  // choice, so it keeps the backend-mirroring defaults below rather than
  // collapsing the form to the Group field alone.
  hasGroup: boolean;
  // Whether the caller holds group.comments.create in the target group. An extra
  // AND on the comment field's "enabled": without it the field is hidden and never
  // required, so a member who cannot comment is never locked out of quick scan, and
  // a comment sent anyway is dropped by the server.
  canCreateComments: boolean;
}

export function resolveQuickScanFieldConfig(
  settings: GroupReceiptSettings | undefined,
  { hasGroup, canCreateComments }: QuickScanFieldConfigOptions
): QuickScanFieldConfig {
  if (!hasGroup) {
    return NO_GROUP_QUICK_SCAN_FIELDS;
  }

  const showPaidBy = settings?.quickScanPaidByEnabled ?? true;
  const showStatus = settings?.quickScanStatusEnabled ?? true;
  const showCategories = settings?.quickScanCategoriesEnabled ?? false;
  const showTags = settings?.quickScanTagsEnabled ?? false;
  // hideComments hides comments for the whole group, so it hides the quick-scan
  // comment too -- without changing the stored toggle (mirrors the backend's
  // GroupReceiptSettings.IsQuickScanCommentShown).
  const showComment =
    (settings?.quickScanCommentEnabled ?? false) &&
    !(settings?.hideComments ?? false) &&
    canCreateComments;

  return {
    showPaidBy,
    requirePaidBy: showPaidBy && (settings?.quickScanPaidByRequired ?? true),
    showStatus,
    requireStatus: showStatus && (settings?.quickScanStatusRequired ?? true),
    showCategories,
    requireCategories: showCategories && (settings?.quickScanCategoriesRequired ?? false),
    showTags,
    requireTags: showTags && (settings?.quickScanTagsRequired ?? false),
    showComment,
    requireComment: showComment && (settings?.quickScanCommentRequired ?? false),
  };
}
