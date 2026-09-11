import 'package:dio/dio.dart' as dio;
import 'package:file_selector/file_selector.dart';
import 'package:image_picker/image_picker.dart';
import 'package:receipt_wrangler_mobile/interfaces/upload_multipart_file_data.dart';

/// The type filter the document source uses.
///
/// One group carries the Android/Linux spelling (`mimeTypes`), the Apple one
/// (`uniformTypeIdentifiers`) and the Windows one (`extensions`) at the same
/// time — each `file_selector` platform implementation reads only the field it
/// understands, which is why this needs no `Platform.operatingSystem` switch.
///
/// `extensions` is not redundant: `file_selector_windows` throws
/// `ArgumentError` on a non-`allowsAny` group without it, and
/// `file_selector_linux` (the e2e host) reads `mimeTypes` + `extensions`
/// because GTK's `add_mime_type("image/*")` wildcard handling is unreliable.
const XTypeGroup receiptFileTypeGroup = XTypeGroup(
  label: "Receipts",
  mimeTypes: ["image/*", "application/pdf"],
  uniformTypeIdentifiers: ["public.image", "com.adobe.pdf"],
  extensions: ["png", "jpg", "jpeg", "heic", "heif", "webp", "pdf"],
);

/// Picks from the OS photo library — the Android Photo Picker on 13+ (the Play
/// Services backport below that) and `PHPickerViewController` on iOS 14+.
///
/// This exists as a source in its own right because [pickDocuments] cannot
/// stand in for it: on iOS that opens the Files app, and the photo library is
/// not a Files provider, so the camera roll is unreachable from it entirely.
///
/// Needs **no runtime permission** on either platform — that is the point of
/// the system pickers. Do not route it through `ensureCameraAccess` or
/// `Gal.requestAccess` (that one is for *saving* to the library), and never
/// declare `READ_MEDIA_IMAGES`: Play Console requires a restricted-permission
/// declaration for it when the system picker would do.
///
/// No `imageQuality` / `maxWidth` / `maxHeight`. Those make `image_picker`
/// re-encode the file on device, and the backend already normalises HEIC
/// (`FileRepository.GetBytesFromImageBytes`) before OCR — so re-encoding here
/// would cost fidelity on the exact images we need read accurately.
Future<List<UploadMultipartFileData>> pickPhotos() async {
  return _toUploadData(await ImagePicker().pickMultiImage());
}

/// Picks from the OS document browser — `ACTION_OPEN_DOCUMENT` (SAF) on
/// Android, the Files app on iOS.
///
/// The only source that can reach a PDF, which the backend converts to JPG
/// server-side. The photo pickers are media-only by design.
Future<List<UploadMultipartFileData>> pickDocuments() async {
  return _toUploadData(
      await openFiles(acceptedTypeGroups: const [receiptFileTypeGroup]));
}

/// Reads each picked file into the shape the upload paths expect.
///
/// The filename must be non-empty: the Go multipart parser closes the
/// connection on a blank one, which surfaces as an opaque
/// `DioException [unknown]` and a receipt saved with no images. `XFile.name`
/// off a real path is always populated — which is also why the test fakes back
/// their `XFile`s with real temp files rather than `XFile.fromData`.
Future<List<UploadMultipartFileData>> _toUploadData(List<XFile> files) async {
  final uploads = <UploadMultipartFileData>[];

  for (final file in files) {
    final bytes = await file.readAsBytes();
    uploads.add(UploadMultipartFileData(
      multipartFile: dio.MultipartFile.fromBytes(bytes, filename: file.name),
      bytes: bytes,
    ));
  }

  return uploads;
}
