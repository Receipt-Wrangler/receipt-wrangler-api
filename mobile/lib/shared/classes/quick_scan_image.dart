import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/cupertino.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:openapi/openapi.dart';
import 'package:receipt_wrangler_mobile/interfaces/upload_multipart_file_data.dart';

class QuickScanImage extends UploadMultipartFileData {
  QuickScanImage(
      {required MultipartFile multipartFile,
      required Uint8List bytes,
      required this.formKey,
      this.groupId,
      this.paidByUserId,
      this.status,
      this.categories = const [],
      this.tags = const [],
      this.comment})
      : super(multipartFile: multipartFile, bytes: bytes);

  // `multipartFile` and `bytes` deliberately are NOT redeclared here. They used
  // to be, shadowing the superclass fields -- both copies held the same object
  // (the constructor forwards to `super`), so it worked, but any getter added to
  // the base class reads the base field while subclass code reads the shadow.
  // `UploadMultipartFileData.filename` is exactly such a getter.

  final GlobalKey<FormBuilderState> formKey;

  int? groupId;

  int? paidByUserId;

  ReceiptStatus? status;

  List<Category> categories;

  List<Tag> tags;

  String? comment;

  static QuickScanImage fromUploadMultipartFileData(
      UploadMultipartFileData data,
      int? initialGroupId,
      int? initialPaidByUserId,
      ReceiptStatus? initialStatus) {
    return QuickScanImage(
        multipartFile: data.multipartFile,
        bytes: data.bytes,
        formKey: GlobalKey<FormBuilderState>(),
        groupId: initialGroupId,
        paidByUserId: initialPaidByUserId,
        status: initialStatus);
  }
}
