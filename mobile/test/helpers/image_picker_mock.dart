import 'dart:io';
import 'dart:typed_data';

import 'package:flutter/services.dart' show rootBundle;
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:plugin_platform_interface/plugin_platform_interface.dart';

/// Replaces [ImagePickerPlatform.instance] with a stub, so tests can exercise
/// the **photo-library** source without a native picker.
///
/// This lives in `test/helpers/` rather than `integration_test/helpers/` for the
/// reason `channel_mocks.dart` does: the widget suite is the gating one and must
/// not import from `integration_test/`. `platform_mocks.dart` re-exports it.
///
/// Same platform-interface swap as `installFileSelectorMock`, and the same
/// on-disk-file rationale — see that file's comment. An `XFile.fromData` has an
/// empty `.name`, which reaches the Go multipart parser as an empty filename;
/// the parser closes the connection, dio reports a generic
/// `DioException [unknown]`, and the receipt saves with zero images.
///
/// **On Linux this fake is redundant and cannot be told apart from the file
/// one**: `image_picker_linux` is implemented on top of `file_selector_linux`,
/// so `installFileSelectorMock` already intercepts the photo path there. A spec
/// that must prove the two sources are distinct has to be a widget test using
/// this fake explicitly, or mobile-only.
class _FakeImagePicker extends ImagePickerPlatform
    with MockPlatformInterfaceMixin {
  _FakeImagePicker(this._path, this._name);

  final String _path;
  final String _name;

  @override
  Future<List<XFile>> getMultiImageWithOptions({
    MultiImagePickerOptions options = const MultiImagePickerOptions(),
  }) async {
    return <XFile>[XFile(_path, name: _name, mimeType: 'image/png')];
  }
}

/// An [ImagePickerPlatform] whose every pick throws.
///
/// Drives the "couldn't open your photos" branch explicitly. Before this, the
/// only test of that branch relied on `getGalleryImages` throwing
/// `"Unsupported platform"` on a desktop host — an accident of where the suite
/// ran rather than an asserted contract.
class _FailingImagePicker extends ImagePickerPlatform
    with MockPlatformInterfaceMixin {
  @override
  Future<List<XFile>> getMultiImageWithOptions({
    MultiImagePickerOptions options = const MultiImagePickerOptions(),
  }) async {
    throw Exception('photo picker unavailable');
  }
}

/// Writes [bytes] to a fresh tempdir and installs a fake photo picker returning
/// an [XFile] backed by that path.
///
/// [bytes] defaults to the bundled `assets/test/sample.png`, which only works
/// under the **integration** binding — the widget suite does not service
/// `rootBundle` the same way and the load never completes there, so widget
/// tests must pass bytes explicitly (`quickScanTestPngBytes`).
Future<void> installImagePickerMock({
  Uint8List? bytes,
  String name = 'sample.png',
}) async {
  final pngBytes = bytes ??
      (await rootBundle.load('assets/test/sample.png')).buffer.asUint8List();
  final tempDir = await Directory.systemTemp.createTemp('image_picker_mock_');
  final tempFile = File('${tempDir.path}/$name');
  await tempFile.writeAsBytes(pngBytes, flush: true);
  ImagePickerPlatform.instance = _FakeImagePicker(tempFile.path, name);
}

/// Installs a photo picker that always fails, for specs asserting the failure
/// message.
void installFailingImagePickerMock() {
  ImagePickerPlatform.instance = _FailingImagePicker();
}
