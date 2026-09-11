import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/shared/functions/quick_scan_field_config.dart';

// resolveQuickScanFieldConfig is the single source of the quick-scan field
// show/require defaults, shared by the per-image form and the submit path.
// These tests pin the backend-mirroring defaults so the two callers can't drift.
api.GroupReceiptSettings _settings({
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
        ..id = 1
        ..createdAt = ''
        ..groupId = 1
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

void main() {
  group('resolveQuickScanFieldConfig', () {
    void expectNoFieldsShown(QuickScanFieldConfig config) {
      expect(config.showPaidBy, isFalse);
      expect(config.requirePaidBy, isFalse);
      expect(config.showStatus, isFalse);
      expect(config.requireStatus, isFalse);
      expect(config.showCategories, isFalse);
      expect(config.requireCategories, isFalse);
      expect(config.showTags, isFalse);
      expect(config.requireTags, isFalse);
      expect(config.showComment, isFalse);
      expect(config.requireComment, isFalse);
    }

    test('no group selected hides every field', () {
      // Nothing but the Group dropdown renders until a group is picked: there is
      // no config to honour, and guessing one means flipping the field set the
      // moment the user chooses.
      final config = resolveQuickScanFieldConfig(
        null,
        hasGroup: false,
        canCreateComments: true,
      );

      expectNoFieldsShown(config);
    });

    test('no group wins over a fully-enabled config', () {
      // Pins that the early return is unconditional - settings can never sneak a
      // field back in while the group is unpicked.
      final config = resolveQuickScanFieldConfig(
        _settings(
          categoriesEnabled: true,
          tagsEnabled: true,
          commentEnabled: true,
        ),
        hasGroup: false,
        canCreateComments: true,
      );

      expectNoFieldsShown(config);
    });

    test('a selected group with no settings still gets the backend defaults', () {
      // The stale-quickScanDefaultGroupId / AppData-not-loaded case: a group id we
      // cannot resolve is still a choice, so it keeps the backend defaults rather
      // than collapsing to the Group field. See quick_scan_initial_values_test's
      // 'prefers the quick-scan default group over the only group', which pins
      // that such an id reaches the form.
      final config = resolveQuickScanFieldConfig(
        null,
        hasGroup: true,
        canCreateComments: true,
      );

      // Paid-by/status shown+required; categories/tags hidden.
      expect(config.showPaidBy, isTrue);
      expect(config.requirePaidBy, isTrue);
      expect(config.showStatus, isTrue);
      expect(config.requireStatus, isTrue);
      expect(config.showCategories, isFalse);
      expect(config.requireCategories, isFalse);
      expect(config.showTags, isFalse);
      expect(config.requireTags, isFalse);
      expect(config.showComment, isFalse);
      expect(config.requireComment, isFalse);
    });

    test('mirrors each required flag when the fields are shown', () {
      // All fields shown so require reflects the persisted required flag directly.
      final config = resolveQuickScanFieldConfig(
        _settings(
          paidByEnabled: true,
          paidByRequired: false,
          statusEnabled: true,
          statusRequired: true,
          categoriesEnabled: true,
          categoriesRequired: true,
          tagsEnabled: true,
          tagsRequired: false,
          commentEnabled: true,
          commentRequired: true,
        ),
        hasGroup: true,
        canCreateComments: true,
      );

      expect(config.showPaidBy, isTrue);
      expect(config.requirePaidBy, isFalse);
      expect(config.showStatus, isTrue);
      expect(config.requireStatus, isTrue);
      expect(config.showCategories, isTrue);
      expect(config.requireCategories, isTrue);
      expect(config.showTags, isTrue);
      expect(config.requireTags, isFalse);
      expect(config.showComment, isTrue);
      expect(config.requireComment, isTrue);
    });

    test('require is false when the field is hidden even if required flag set', () {
      // A hidden field can never be "required" — require is gated on show.
      final config = resolveQuickScanFieldConfig(
        _settings(
          paidByEnabled: false,
          paidByRequired: true,
          statusEnabled: false,
          statusRequired: true,
          categoriesEnabled: false,
          categoriesRequired: true,
          tagsEnabled: false,
          tagsRequired: true,
          commentEnabled: false,
          commentRequired: true,
        ),
        hasGroup: true,
        canCreateComments: true,
      );

      expect(config.showPaidBy, isFalse);
      expect(config.requirePaidBy, isFalse);
      expect(config.showStatus, isFalse);
      expect(config.requireStatus, isFalse);
      expect(config.showCategories, isFalse);
      expect(config.requireCategories, isFalse);
      expect(config.showTags, isFalse);
      expect(config.requireTags, isFalse);
      expect(config.showComment, isFalse);
      expect(config.requireComment, isFalse);
    });

    test('comment is hidden and not required without group.comments.create', () {
      // A member who cannot comment must not be blocked from quick scanning by a
      // required comment they can never fill; the server drops one sent anyway.
      final config = resolveQuickScanFieldConfig(
        _settings(commentEnabled: true, commentRequired: true),
        hasGroup: true,
        canCreateComments: false,
      );

      expect(config.showComment, isFalse);
      expect(config.requireComment, isFalse);
    });

    test('comment is hidden when the group hides comments', () {
      // hideComments overrides the quick-scan toggle without changing it.
      final config = resolveQuickScanFieldConfig(
        _settings(
          commentEnabled: true,
          commentRequired: true,
          hideComments: true,
        ),
        hasGroup: true,
        canCreateComments: true,
      );

      expect(config.showComment, isFalse);
      expect(config.requireComment, isFalse);
    });

    test('a shown comment is optional unless required', () {
      final config = resolveQuickScanFieldConfig(
        _settings(commentEnabled: true, commentRequired: false),
        hasGroup: true,
        canCreateComments: true,
      );

      expect(config.showComment, isTrue);
      expect(config.requireComment, isFalse);
    });
  });
}
