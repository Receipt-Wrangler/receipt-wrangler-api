import 'dart:async';

import 'package:built_collection/built_collection.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:infinite_carousel/infinite_carousel.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/models/user_preferences_model.dart';
import 'package:receipt_wrangler_mobile/receipts/widgets/quick_scan.dart';
import 'package:receipt_wrangler_mobile/shared/classes/quick_scan_image.dart';
import 'package:receipt_wrangler_mobile/shared/functions/permissions.dart';
import 'package:receipt_wrangler_mobile/shared/functions/quick_scan_field_config.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/delete_button.dart';
import 'package:receipt_wrangler_mobile/utils/scan.dart';
import 'package:receipt_wrangler_mobile/utils/snackbar.dart';
import 'package:rxdart/rxdart.dart';

import '../../client/client.dart';
import '../../constants/colors.dart';
import '../../constants/receipt_entry.dart';
import '../../interfaces/upload_multipart_file_data.dart';
import '../../utils/bottom_sheet.dart';
import 'receipt_entry.dart';
import 'receipt_entry_availability.dart';

Widget _getUploadIcon(
    context,
    BehaviorSubject<List<QuickScanImage>> imageSubject,
    BehaviorSubject<bool> isCompletedSubject) {
  return StreamBuilder<bool>(
    stream: isCompletedSubject.stream,
    builder: (context, snapshot) {
      final isCompleted = snapshot.hasData && snapshot.data == true;

      return IconButton(
        icon: const Icon(Icons.add_a_photo),
        onPressed: isCompleted
            ? null
            : () async {
                var uploadedImages = await scanImagesMultiPart(100);
                if (uploadedImages.isNotEmpty && context.mounted) {
                  imageSubject.add(imageSubject.value +
                      buildQuickScanImages(context, uploadedImages));
                }
              },
      );
    },
  );
}

Widget _getGalleryUploadImage(
    context,
    BehaviorSubject<List<QuickScanImage>> imageSubject,
    BehaviorSubject<bool> isCompletedSubject) {
  return StreamBuilder<bool>(
    stream: isCompletedSubject.stream,
    builder: (context, snapshot) {
      final isCompleted = snapshot.hasData && snapshot.data == true;

      return IconButton(
        icon: const Icon(Icons.upload_file_rounded),
        onPressed: isCompleted
            ? null
            : () async {
                var uploadedImages = await pickGalleryImages(context);
                if (uploadedImages.isNotEmpty && context.mounted) {
                  imageSubject.add(imageSubject.value +
                      buildQuickScanImages(context, uploadedImages));
                }
              },
      );
    },
  );
}

/// Wraps freshly picked files as [QuickScanImage]s, seeding each one from the
/// caller's quick-scan user preferences.
///
/// Shared by the sheet's own camera/gallery icons and by the nav entry points
/// that open the sheet already seeded, so a scan started from the bottom nav and
/// one started from inside the sheet prefill identically.
List<QuickScanImage> buildQuickScanImages(
    BuildContext context, List<UploadMultipartFileData> uploadedImages) {
  final initialQuickScanValues = _getInitialQuickScanValues(context);
  return uploadedImages
      .map((image) => QuickScanImage.fromUploadMultipartFileData(
            image,
            initialQuickScanValues.groupId,
            initialQuickScanValues.paidByUserId,
            initialQuickScanValues.status,
          ))
      .toList();
}

({int? groupId, int? paidByUserId, api.ReceiptStatus? status})
    _getInitialQuickScanValues(BuildContext context) {
  late final userPreferenceModel =
      Provider.of<UserPreferencesModel>(context, listen: false);
  final userPreferences = userPreferenceModel.userPreferences;
  // Unset reads as 0, not null -- the generated model defaults it (see
  // mobile/CLAUDE.md -> "Regenerating API Client Models").
  final preferredGroupId = userPreferences.quickScanDefaultGroupId ?? 0;
  return (
    // The user's own quick-scan default wins. Without one, a member of exactly
    // one group has no choice to make, so seed it rather than handing them a
    // picker with a single option.
    groupId: preferredGroupId > 0
        ? preferredGroupId
        : Provider.of<GroupModel>(context, listen: false).soleGroupId,
    paidByUserId: userPreferences.quickScanDefaultPaidById,
    status: userPreferences.quickScanDefaultStatus,
  );
}

Widget _getSubmitButton(
    BuildContext context,
    BehaviorSubject<List<QuickScanImage>> imageSubject,
    BehaviorSubject<bool> isCompletedSubject) {
  return StreamBuilder<bool>(
    stream: isCompletedSubject.stream,
    builder: (context, snapshot) {
      if (snapshot.hasData && snapshot.data == true) {
        return const SizedBox.shrink();
      }

      return BottomSubmitButton(
        onPressed: () async {
          final loadingModel = Provider.of<LoadingModel>(context, listen: false);
          await _submitQuickScan(context, imageSubject.value, loadingModel, isCompletedSubject);
        },
      );
    },
  );
}

Future<void> _submitQuickScan(
    BuildContext context,
    List<QuickScanImage> images,
    LoadingModel loadingModel,
    BehaviorSubject<bool> isCompletedSubject) async {
  List<int> groupIds = [];
  List<int> paidByUserIds = [];
  List<api.ReceiptStatus> statuses = [];
  List<String> categoryIds = [];
  List<String> tagIds = [];
  List<String> comments = [];
  List<MultipartFile> files = [];

  if (images.isEmpty) {
    showErrorSnackbar(context, "Please upload at least one image");
    return;
  }

  final groupModel = Provider.of<GroupModel>(context, listen: false);
  final permissionsModel = Provider.of<PermissionsModel>(context, listen: false);

  var errored = false;
  for (var (index, image) in images.indexed) {
    final groupId = image.groupId ?? 0;
    if (groupId <= 0) {
      errored = true;
      showErrorSnackbar(
          context, "Please fix error on quick scan ${index + 1} to continue");
      break;
    }

    // Resolve each field against the group's quick-scan config, mirroring the
    // backend's resolveQuickScanFields. A hidden/optional field is sent as the
    // "unset" sentinel (0 / empty) so the server backfills the configured
    // default; a shown+required field must have a value.
    final settings = groupModel.getGroupReceiptSettings(groupId);

    final config = resolveQuickScanFieldConfig(
      settings,
      canCreateComments: canCommentCreate(permissionsModel, groupId),
    );
    final showPaidBy = config.showPaidBy;
    final requirePaidBy = config.requirePaidBy;
    final showStatus = config.showStatus;
    final requireStatus = config.requireStatus;
    final showCategories = config.showCategories;
    final requireCategories = config.requireCategories;
    final showTags = config.showTags;
    final requireTags = config.requireTags;
    final showComment = config.showComment;
    final requireComment = config.requireComment;

    int paidBy = 0;
    if (showPaidBy) {
      paidBy = image.paidByUserId ?? 0;
      if (requirePaidBy && paidBy <= 0) {
        errored = true;
      }
    }

    api.ReceiptStatus status = api.ReceiptStatus.empty;
    if (showStatus) {
      status = image.status ?? api.ReceiptStatus.empty;
      if (requireStatus && status == api.ReceiptStatus.empty) {
        errored = true;
      }
    }

    List<int> catIds = [];
    if (showCategories) {
      catIds =
          image.categories.map((c) => c.id ?? 0).where((id) => id > 0).toList();
      if (requireCategories && catIds.isEmpty) {
        errored = true;
      }
    }

    List<int> tgIds = [];
    if (showTags) {
      tgIds = image.tags.map((t) => t.id ?? 0).where((id) => id > 0).toList();
      if (requireTags && tgIds.isEmpty) {
        errored = true;
      }
    }

    String comment = "";
    if (showComment) {
      comment = image.comment?.trim() ?? "";
      if (requireComment && comment.isEmpty) {
        errored = true;
      }
    }

    if (errored) {
      showErrorSnackbar(
          context, "Please fix error on quick scan ${index + 1} to continue");
      break;
    }

    files.add(image.multipartFile);
    groupIds.add(groupId);
    paidByUserIds.add(paidBy);
    statuses.add(status);
    categoryIds.add(catIds.join(","));
    tagIds.add(tgIds.join(","));
    comments.add(comment);
  }

  if (errored) {
    return;
  }

  loadingModel.setIsLoading(true);

  try {
    await OpenApiClient.client.getReceiptApi().quickScanReceipt(
        files: files.toBuiltList(),
        groupIds: groupIds.toBuiltList(),
        paidByUserIds: paidByUserIds.toBuiltList(),
        statuses: statuses.toBuiltList(),
        categoryIds: categoryIds.toBuiltList(),
        tagIds: tagIds.toBuiltList(),
        comments: comments.toBuiltList());

    var imageWord = images.length > 1 ? "images" : "image";

    showSuccessSnackbar(
      context,
      "Successfully queued $imageWord for processing!",
    );

    isCompletedSubject.add(true);
  } catch (e) {
    print(e);
    showApiErrorSnackbar(context, e as dynamic);
  } finally {
    loadingModel.setIsLoading(false);
  }

  return;
}

Widget _getDeleteIcon(
    InfiniteScrollController infiniteScrollController,
    BehaviorSubject<List<QuickScanImage>> imageSubject,
    BehaviorSubject<bool> isCompletedSubject) {
  return StreamBuilder<List<QuickScanImage>>(
    stream: imageSubject.stream.asBroadcastStream(),
    builder: (context, imageSnapshot) {
      if (imageSnapshot.hasData && imageSnapshot.data!.isNotEmpty) {
        return StreamBuilder<bool>(
          stream: isCompletedSubject.stream,
          builder: (context, completedSnapshot) {
            final isCompleted = completedSnapshot.hasData && completedSnapshot.data == true;

            return DeleteButton(
              onPressed: isCompleted
                  ? null
                  : () {
                      var images = imageSubject.value;
                      images.removeAt(infiniteScrollController.selectedItem);
                      imageSubject.add(images);
                    },
            );
          },
        );
      } else {
        return const SizedBox();
      }
    },
  );
}

/// Confirms the scan was accepted, and stays put once the success snackbar
/// fades.
///
/// Submitting disables every field and hides the submit button; without this the
/// sheet would just sit there greyed out with nothing saying why. Extraction is
/// an async backend job, so the wording promises a result rather than showing
/// one.
Widget _getQueuedConfirmation(BehaviorSubject<bool> isCompletedSubject) {
  return StreamBuilder<bool>(
    stream: isCompletedSubject.stream,
    builder: (context, snapshot) {
      if (!(snapshot.hasData && snapshot.data == true)) {
        return const SizedBox.shrink();
      }

      return Container(
        key: const ValueKey("quick-scan-queued-confirmation"),
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        color: noticeSurface,
        child: Row(
          children: [
            const Icon(Icons.check_circle_outline,
                size: 20, color: onNoticeSurface),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                quickScanQueuedMessage,
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: onNoticeSurface),
              ),
            ),
          ],
        ),
      );
    },
  );
}

/// Offers manual entry as a way out of the sheet, for a user who would rather
/// type the receipt than have it extracted. Gated on `group.receipts.create`:
/// without it the manual form would only reject the save.
///
/// [canCreateManual] is resolved by the caller rather than here: this widget is
/// built inside the modal sheet's route, which sits outside the GoRouter subtree
/// the gate reads the current group from.
Widget _getManualEntryLink(
  BuildContext callerContext,
  bool canCreateManual,
  BehaviorSubject<bool> isCompletedSubject,
) {
  if (!canCreateManual) {
    return const SizedBox.shrink();
  }

  return StreamBuilder<bool>(
    stream: isCompletedSubject.stream,
    builder: (sheetContext, snapshot) {
      final isCompleted = snapshot.hasData && snapshot.data == true;
      if (isCompleted) {
        return const SizedBox.shrink();
      }

      return TextButton(
        key: const ValueKey("quick-scan-manual-entry-link"),
        onPressed: () {
          // Close the sheet first: the form is a route, and leaving the modal
          // above it would hide the screen the user just asked for. The pop
          // needs the sheet's own context; the navigation needs the caller's,
          // which is the one under the router.
          Navigator.of(sheetContext).pop();
          openManualReceipt(callerContext);
        },
        child: const Text(enterDetailsManuallyLabel),
      );
    },
  );
}

/// Opens the Quick Scan sheet, optionally already seeded with [initialImages]
/// (the nav's Scan tap captures first, then opens the sheet on the result).
///
/// Re-checks both gates rather than trusting the caller: the sheet is reachable
/// from the nav tap, the long-press menu and the overflow menu, and the server
/// enforces `group.receipts.quick-scan` on submit regardless.
showQuickScanBottomSheet(BuildContext context,
    {List<QuickScanImage> initialImages = const []}) {
  final availability = resolveReceiptEntryAvailability(context);
  if (!availability.canQuickScan) {
    showErrorSnackbar(
        context,
        availability.blockedReason == QuickScanBlockedReason.aiDisabled
            ? quickScanAiDisabledMessage
            : availability.groupName == null
                ? quickScanNoPermissionMessage
                : quickScanNoPermissionMessageForGroup(availability.groupName!));
    return;
  }

  var infiniteScrollController = InfiniteScrollController();
  BehaviorSubject<List<QuickScanImage>> imageSubject =
      BehaviorSubject<List<QuickScanImage>>.seeded(
          List<QuickScanImage>.from(initialImages));
  BehaviorSubject<bool> isCompletedSubject =
      BehaviorSubject<bool>.seeded(false);

  List<Widget> actions = [
    _getUploadIcon(context, imageSubject, isCompletedSubject),
    _getGalleryUploadImage(context, imageSubject, isCompletedSubject),
    _getDeleteIcon(infiniteScrollController, imageSubject, isCompletedSubject),
  ];

  showFullscreenBottomSheet(
      context,
      QuickScan(
        imageSubject: imageSubject,
        infiniteScrollController: infiniteScrollController,
        isCompletedSubject: isCompletedSubject,
      ),
      quickScanLabel,
      actions: actions,
      bodyPadding: EdgeInsets.zero,
      // QuickScan sizes itself to the body and each slide scrolls its own form,
      // so the sheet must neither add a scroll view of its own (that would make
      // the body unbounded again and re-clip the tail of every slide) nor float
      // the submit button over the body (that would bury the form's last field,
      // which is exactly where a configured comment lands).
      bodyFillsSheet: true,
      bottomSheetWidget: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _getQueuedConfirmation(isCompletedSubject),
          _getManualEntryLink(
              context, availability.canCreateManual, isCompletedSubject),
          _getSubmitButton(context, imageSubject, isCompletedSubject),
        ],
      ));
}
