import 'package:flutter/material.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/enums/form_state.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/full_screen_image_viewer.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';
import 'package:rxdart/rxdart.dart';

import '../../models/loading_model.dart';
import '../../client/client.dart';
import '../../interfaces/upload_multipart_file_data.dart';
import '../../models/receipt_model.dart';
import '../../shared/functions/receipt_upload.dart';
import '../../shared/functions/receipt_image_gallery.dart';
import '../../shared/widgets/receipt_edit_popup_menu.dart';
import '../../shared/widgets/unrenderable_file_placeholder.dart';
import '../../utils/snackbar.dart';

class ReceiptImageAppBar extends StatelessWidget implements PreferredSizeWidget {
  const ReceiptImageAppBar({
    super.key,
    required this.formState,
    required this.title,
    required this.onBack,
    this.onEditMode,
  });

  final WranglerFormState formState;
  final String title;
  final VoidCallback onBack;
  final VoidCallback? onEditMode;

  @override
  Size get preferredSize => const Size.fromHeight(kToolbarHeight);

  @override
  Widget build(BuildContext context) {
    final receiptModel = Provider.of<ReceiptModel>(context, listen: false);
    
    return AppBar(
      title: Text(title),
      leading: IconButton(
        icon: const Icon(Icons.arrow_back),
        onPressed: onBack,
      ),
      actions: [_buildImageActions(context, receiptModel)],
    );
  }

  Widget _buildImageActions(BuildContext context, ReceiptModel receiptModel) {
    List<PopupMenuEntry> options = [];
    
    if (_isEditMode()) {
      options = _buildEditModeActions(context, receiptModel);
    } else {
      options = _buildViewModeActions(context, receiptModel);
    }

    options.add(_buildImageDownloadButton(context, receiptModel));
    options.add(_buildViewInFullScreenButton(context, receiptModel));

    var combinedStream = Rx.merge([
      receiptModel.imagesToUploadBehaviorSubject.stream,
      receiptModel.imageBehaviorSubject.stream
    ]).asBroadcastStream();

    return StreamBuilder(
        stream: combinedStream,
        builder: (context, snapshot) {
          return ReceiptEditPopupMenu(
              groupId: receiptModel.receipt.groupId,
              popupMenuChildren: options,
              formState: formState);
        });
  }

  bool _isEditMode() {
    return formState == WranglerFormState.edit || formState == WranglerFormState.add;
  }

  List<PopupMenuEntry> _buildViewModeActions(BuildContext context, ReceiptModel receiptModel) {
    return [
      PopupMenuItem(
          value: "edit",
          child: const Text("Edit"),
          onTap: () {
            if (onEditMode != null) {
              onEditMode!();
            }
          }),
    ];
  }

  List<PopupMenuEntry> _buildEditModeActions(BuildContext context, ReceiptModel receiptModel) {
    final popupMenuEntries = buildImageSourceMenuItems(context,
        receiptModel: receiptModel, formState: formState);

    if (receiptModel.imageBehaviorSubject.value.isNotEmpty ||
        receiptModel.imagesToUploadBehaviorSubject.value.isNotEmpty) {
      popupMenuEntries.add(
        _buildDeleteButton(context, receiptModel),
      );
    }

    return popupMenuEntries;
  }

  PopupMenuItem _buildViewInFullScreenButton(BuildContext context, ReceiptModel receiptModel) {
    return PopupMenuItem(
        child: Text("View in full screen"),
        enabled: formState == WranglerFormState.add ? _areImagesToUpload(receiptModel) : true,
        onTap: () {
          var selectedIndex = receiptModel.infiniteScrollController.selectedItem;
          var image = receiptModel
                  .imageBehaviorSubject.value[selectedIndex]?.encodedImage ??
              "";

          var bytes = getBytesFromEncodedImage(image);
          Navigator.push(
            context,
            MaterialPageRoute(
                builder: (context) =>
                    FullScreenImageViewer(
                        image: Image.memory(bytes,
                            errorBuilder: unrenderableFileErrorBuilder()))),
          );
        });
  }

  PopupMenuItem _buildImageDownloadButton(BuildContext context, ReceiptModel receiptModel) {
    return PopupMenuItem(
      child: const Text("Download"),
      enabled: formState == WranglerFormState.add ? _areImagesToUpload(receiptModel) : true,
      onTap: () async => await _downloadImage(context, receiptModel),
    );
  }

  Future _downloadImage(BuildContext context, ReceiptModel receiptModel) async {
    final loadingModel = Provider.of<LoadingModel>(context, listen: false);
    var receiptImage = _getCurrentlySelectedImage(receiptModel);
    if (receiptImage == null) {
      return;
    }
    await saveReceiptImageToGallery(context, loadingModel, receiptImage.id);
  }

  PopupMenuEntry _buildDeleteButton(BuildContext context, ReceiptModel receiptModel) {
    return PopupMenuItem(
      child: const Text("Delete Image"),
      value: "delete",
      onTap: () async => await _deleteImage(context, receiptModel),
    );
  }

  Future<void> _deleteImage(BuildContext context, ReceiptModel receiptModel) async {
    if (formState == WranglerFormState.add) {
      _deleteImageFromModel(receiptModel);
      return;
    }

    if (formState == WranglerFormState.edit) {
      await _deleteImageViaApi(context, receiptModel);
      return;
    }
  }

  void _deleteImageFromModel(ReceiptModel receiptModel) {
    var index = receiptModel.infiniteScrollController.selectedItem;
    var currentImages = List<UploadMultipartFileData>.from(
        receiptModel.imagesToUploadBehaviorSubject.value);
    currentImages.removeAt(index);
    receiptModel.imagesToUploadBehaviorSubject.add(currentImages);
  }

  Future<void> _deleteImageViaApi(BuildContext context, ReceiptModel receiptModel) async {
    try {
      var index = receiptModel.infiniteScrollController.selectedItem;
      await OpenApiClient.client.getReceiptImageApi().deleteReceiptImageById(
          receiptImageId: receiptModel.imageBehaviorSubject.value[index]!.id);
      var currentImages =
          List<api.FileDataView?>.from(receiptModel.imageBehaviorSubject.value);
      currentImages.removeAt(index);
      receiptModel.imageBehaviorSubject.add(currentImages);

      showSuccessSnackbar(context, "Successfully deleted image");
    } catch (e) {
      showApiErrorSnackbar(context, e);
    }
  }

  bool _areImagesToUpload(ReceiptModel receiptModel) {
    return !receiptModel.imagesToUploadBehaviorSubject.value.isEmpty;
  }

  api.FileDataView? _getCurrentlySelectedImage(ReceiptModel receiptModel) {
    var index = receiptModel.infiniteScrollController.selectedItem;
    return receiptModel.imageBehaviorSubject.value[index];
  }
}