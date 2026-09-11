import 'dart:typed_data';
import 'package:dio/dio.dart';

class UploadMultipartFileData {
  MultipartFile multipartFile;
  Uint8List bytes;

  UploadMultipartFileData({required this.multipartFile, required this.bytes});

  /// The name the file will be uploaded under.
  ///
  /// Reaches through to the multipart file so callers (notably the
  /// unrenderable-file placeholder) don't each have to.
  String? get filename => multipartFile.filename;
}
