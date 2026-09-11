import 'package:cunning_document_scanner/cunning_document_scanner.dart'
    show CunningDocumentScannerException;
import 'package:flutter/material.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';

import '../../client/client.dart';
import '../../constants/receipt_entry.dart';
import '../../enums/form_state.dart';
import '../../enums/upload_method.dart';
import '../../interfaces/upload_multipart_file_data.dart';
import '../../models/loading_model.dart';
import '../../models/receipt_model.dart';
import '../../utils/media_picker.dart';
import '../../utils/scan.dart';
import '../../utils/snackbar.dart';

/// Acquires receipt files from [method].
///
/// The single entry point for every image source in the app, so the call sites
/// — the Quick Scan sheet's two app-bar actions, the scan slot's long-press
/// menu, the receipts overflow menu and the receipt-image app bar — cannot
/// drift on which picker a source opens, how a failure is reported, or whether
/// the result is guarded before it touches the tree.
///
/// Returns an empty list for both "the user cancelled" and "the picker failed".
/// Cancelling is an ordinary flow at every call site, and a failure has already
/// been explained to the user by the time this returns, so neither is worth an
/// exception the callers would each have to handle identically.
///
/// [cameraPages] is the document scanner's page ceiling and is ignored by the
/// other sources: the Quick Scan sheet accepts a whole multi-page scan, while
/// the receipt-image app bar attaches one page at a time.
Future<List<UploadMultipartFileData>> acquireReceiptFiles(
  BuildContext context,
  UploadMethod method, {
  int cameraPages = 100,
}) async {
  try {
    switch (method) {
      case UploadMethod.camera:
        return await scanImagesMultiPart(cameraPages);
      case UploadMethod.photos:
        return await pickPhotos();
      case UploadMethod.files:
        return await pickDocuments();
    }
  } on CunningDocumentScannerException {
    // The scanner re-checks camera permission itself and throws when it is
    // missing. That is not a picker failure but a permission one, and the scan
    // entry points answer it with the camera-denied fallback — turning it into
    // a snackbar here would silently delete that behaviour.
    rethrow;
  } catch (_) {
    if (context.mounted) {
      showErrorSnackbar(context, _unavailableMessage(method));
    }
    return [];
  }
}

/// Each source fails for its own reason, so each says so. A single "gallery"
/// message would misdescribe two of the three.
String _unavailableMessage(UploadMethod method) {
  switch (method) {
    case UploadMethod.camera:
      return scannerUnavailableMessage;
    case UploadMethod.photos:
      return photoPickerUnavailableMessage;
    case UploadMethod.files:
      return filePickerUnavailableMessage;
  }
}

/// The source entries every receipt-image menu offers.
///
/// Built here rather than in the app bar so the label set, the ordering and the
/// action behind each entry are defined once.
List<PopupMenuEntry> buildImageSourceMenuItems(
  BuildContext context, {
  required ReceiptModel receiptModel,
  required WranglerFormState formState,
}) {
  PopupMenuEntry entry(String value, String label, UploadMethod method) {
    return PopupMenuItem(
      value: value,
      onTap: () async => await addImagesToReceipt(
        context,
        receiptModel: receiptModel,
        formState: formState,
        method: method,
      ),
      child: Text(label),
    );
  }

  return [
    entry("camera", uploadFromCameraLabel, UploadMethod.camera),
    entry("photos", uploadPhotoLabel, UploadMethod.photos),
    entry("files", uploadFileLabel, UploadMethod.files),
  ];
}

/// Acquires files from [method] and routes them into [receiptModel] the way
/// [formState] demands: an unsaved receipt stashes them for the form's own
/// submit, a saved one uploads each immediately.
///
/// [formState] is a parameter rather than resolved here because the callers
/// know it differently — one reads it off the route via
/// `getFormStateFromContext`, while `ReceiptImageAppBar` is handed it as a
/// constructor field by a screen that pushes itself into edit mode through a
/// `MaterialPageRoute`, with no route change to read.
Future<void> addImagesToReceipt(
  BuildContext context, {
  required ReceiptModel receiptModel,
  required WranglerFormState formState,
  required UploadMethod method,
  int cameraPages = 1,
}) async {
  final images =
      await acquireReceiptFiles(context, method, cameraPages: cameraPages);
  if (images.isEmpty || !context.mounted) {
    return;
  }

  switch (formState) {
    case WranglerFormState.add:
      stageImagesForUpload(receiptModel, images);
    case WranglerFormState.edit:
      await uploadImagesToReceipt(context, receiptModel, images);
    case WranglerFormState.view:
      // No upload affordance is offered here; nothing to do if one is reached.
      break;
  }
}

/// Stashes [images] on the model for a receipt that has not been saved yet. The
/// receipt form uploads them as part of its own submit.
///
/// Emits once for the whole batch rather than once per image, so a multi-image
/// pick rebuilds every carousel listener a single time.
void stageImagesForUpload(
    ReceiptModel receiptModel, List<UploadMultipartFileData> images) {
  receiptModel.imagesToUploadBehaviorSubject
      .add([...receiptModel.imagesToUploadBehaviorSubject.value, ...images]);
}

/// Uploads [images] to an already-saved receipt, one API call each.
Future<void> uploadImagesToReceipt(BuildContext context,
    ReceiptModel receiptModel, List<UploadMultipartFileData> images) async {
  if (images.isEmpty) {
    return;
  }

  // Captured before the first await: the `finally` must be able to lower the
  // spinner even once this context is gone.
  final loadingModel = Provider.of<LoadingModel>(context, listen: false);
  loadingModel.setIsLoading(true);

  try {
    final uploaded = <api.FileDataView?>[];
    for (final image in images) {
      final response = await OpenApiClient.client
          .getReceiptImageApi()
          .uploadReceiptImage(
              file: image.multipartFile, receiptId: receiptModel.receipt.id);
      uploaded.add(response.data);
    }
    receiptModel.imageBehaviorSubject
        .add([...receiptModel.imageBehaviorSubject.value, ...uploaded]);

    if (!context.mounted) {
      return;
    }
    showSuccessSnackbar(
        context,
        images.length > 1
            ? "Successfully uploaded ${images.length} images"
            : "Successfully uploaded image");
  } catch (error, stackTrace) {
    if (!context.mounted) {
      return;
    }
    showApiErrorSnackbar(context, error, stackTrace);
  } finally {
    // Exactly one place, on every path. Raising it only for a non-empty list
    // but lowering it unconditionally used to let a cancelled pick clear a
    // spinner some other in-flight request had raised.
    loadingModel.setIsLoading(false);
  }
}
