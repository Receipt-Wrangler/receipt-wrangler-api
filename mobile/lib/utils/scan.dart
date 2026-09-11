import 'dart:io';

import 'package:cunning_document_scanner/cunning_document_scanner.dart';
import 'package:dio/dio.dart' as dio;
import 'package:http/http.dart';
import 'package:receipt_wrangler_mobile/interfaces/upload_multipart_file_data.dart';

/// The multipart field name handed to `MultipartFile.fromPath` below.
///
/// It never reaches the wire — the generated client hardcodes its own field
/// names (`files` on quick scan, `file` on receipt-image upload) — so this only
/// exists to satisfy the constructor whose basename derivation we actually
/// want.
const _uploadFieldName = "file";

/// Captures pages with the document scanner.
///
/// The gallery and file sources live in `lib/utils/media_picker.dart`; the three
/// are selected between by `acquireReceiptFiles`.
Future<List<UploadMultipartFileData>> scanImagesMultiPart(
    int numberOfPages) async {
  var files = <UploadMultipartFileData>[];
  // Null is what the scanner returns when the user backs out of it, which the
  // Scan entry point makes an ordinary flow rather than an edge case -- so it
  // must read as "no pages", not as a null dereference.
  var filePaths =
      await CunningDocumentScanner.getPictures(noOfPages: numberOfPages) ?? [];

  if (filePaths.isEmpty) {
    return files;
  }

  for (var filePath in filePaths) {
    // Built only to derive the basename from the path; the field name is
    // discarded and the filename is what carries through.
    var multipartFile =
        await MultipartFile.fromPath(_uploadFieldName, filePath);
    var bytes = await File(filePath).readAsBytes();

    var dioMultipartFile =
        dio.MultipartFile.fromBytes(bytes, filename: multipartFile.filename);

    files.add(
        UploadMultipartFileData(multipartFile: dioMultipartFile, bytes: bytes));
  }

  return files;
}
