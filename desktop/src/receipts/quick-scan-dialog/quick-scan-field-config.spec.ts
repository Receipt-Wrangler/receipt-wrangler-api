import { GroupReceiptSettings } from "../../open-api";
import { QuickScanFieldConfig, resolveQuickScanFieldConfig } from "./quick-scan-field-config";

// resolveQuickScanFieldConfig is the single source of the quick-scan field show/require defaults,
// shared by the dialog's show*(i) getters and configureImages(). These cases mirror mobile's
// test/shared/functions/quick_scan_field_config_test.dart one-for-one, so the two clients can be
// diffed against each other.
function settings(overrides: Partial<GroupReceiptSettings> = {}): GroupReceiptSettings {
  return {
    quickScanPaidByEnabled: true,
    quickScanPaidByRequired: true,
    quickScanStatusEnabled: true,
    quickScanStatusRequired: true,
    quickScanCategoriesEnabled: false,
    quickScanCategoriesRequired: false,
    quickScanTagsEnabled: false,
    quickScanTagsRequired: false,
    quickScanCommentEnabled: false,
    quickScanCommentRequired: false,
    hideComments: false,
    ...overrides,
  } as GroupReceiptSettings;
}

function expectNoFieldsShown(config: QuickScanFieldConfig): void {
  expect(config).toEqual({
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
}

describe("resolveQuickScanFieldConfig", () => {
  it("hides every field when no group is selected", () => {
    // Nothing but the Group field renders until a group is picked: there is no config to honour,
    // and guessing one means flipping the field set the moment the user chooses.
    expectNoFieldsShown(
      resolveQuickScanFieldConfig(undefined, { hasGroup: false, canCreateComments: true })
    );
  });

  it("lets no-group win over a fully-enabled config", () => {
    // Pins that the early return is unconditional -- settings can never sneak a field back in
    // while the group is unpicked.
    expectNoFieldsShown(
      resolveQuickScanFieldConfig(
        settings({
          quickScanCategoriesEnabled: true,
          quickScanTagsEnabled: true,
          quickScanCommentEnabled: true,
        }),
        { hasGroup: false, canCreateComments: true }
      )
    );
  });

  it("gives a selected group with no settings the backend defaults", () => {
    // The stale-quickScanDefaultGroupId / AppData-not-loaded case: a group id the store cannot
    // resolve is still a choice, so it keeps the backend defaults rather than collapsing to the
    // Group field. Pinned end-to-end by the dialog spec's "should push new image when there user
    // preferences", which seeds a group id that is absent from the store.
    const config = resolveQuickScanFieldConfig(undefined, {
      hasGroup: true,
      canCreateComments: true,
    });

    expect(config.showPaidBy).toBe(true);
    expect(config.requirePaidBy).toBe(true);
    expect(config.showStatus).toBe(true);
    expect(config.requireStatus).toBe(true);
    expect(config.showCategories).toBe(false);
    expect(config.showTags).toBe(false);
    expect(config.showComment).toBe(false);
  });

  it("mirrors each required flag when the fields are shown", () => {
    const config = resolveQuickScanFieldConfig(
      settings({
        quickScanPaidByRequired: false,
        quickScanCategoriesEnabled: true,
        quickScanCategoriesRequired: true,
        quickScanTagsEnabled: true,
        quickScanTagsRequired: false,
        quickScanCommentEnabled: true,
        quickScanCommentRequired: true,
      }),
      { hasGroup: true, canCreateComments: true }
    );

    expect(config).toEqual({
      showPaidBy: true,
      requirePaidBy: false,
      showStatus: true,
      requireStatus: true,
      showCategories: true,
      requireCategories: true,
      showTags: true,
      requireTags: false,
      showComment: true,
      requireComment: true,
    });
  });

  it("never requires a hidden field, even with the required flag set", () => {
    expectNoFieldsShown(
      resolveQuickScanFieldConfig(
        settings({
          quickScanPaidByEnabled: false,
          quickScanStatusEnabled: false,
          quickScanCategoriesRequired: true,
          quickScanTagsRequired: true,
          quickScanCommentRequired: true,
        }),
        { hasGroup: true, canCreateComments: true }
      )
    );
  });

  it("hides the comment without group.comments.create", () => {
    // A member who cannot comment must not be blocked from quick scanning by a required comment
    // they can never fill; the server drops one sent anyway.
    const config = resolveQuickScanFieldConfig(
      settings({ quickScanCommentEnabled: true, quickScanCommentRequired: true }),
      { hasGroup: true, canCreateComments: false }
    );

    expect(config.showComment).toBe(false);
    expect(config.requireComment).toBe(false);
  });

  it("hides the comment when the group hides comments", () => {
    // hideComments overrides the quick-scan toggle without changing it.
    const config = resolveQuickScanFieldConfig(
      settings({
        quickScanCommentEnabled: true,
        quickScanCommentRequired: true,
        hideComments: true,
      }),
      { hasGroup: true, canCreateComments: true }
    );

    expect(config.showComment).toBe(false);
    expect(config.requireComment).toBe(false);
  });

  it("leaves a shown comment optional unless required", () => {
    const config = resolveQuickScanFieldConfig(
      settings({ quickScanCommentEnabled: true, quickScanCommentRequired: false }),
      { hasGroup: true, canCreateComments: true }
    );

    expect(config.showComment).toBe(true);
    expect(config.requireComment).toBe(false);
  });
});
