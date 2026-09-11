import 'dart:io';
import 'dart:typed_data';

import 'package:flutter/services.dart' show rootBundle;
import 'platform_mocks.dart';

/// Stubs the `cunning_document_scanner` method channel so Quick Scan's
/// "add a photo" action (`scanImagesMultiPart` → `CunningDocumentScanner.getPictures`)
/// returns a fixed on-disk image instead of driving the native camera scanner.
///
/// This used to be the *only* way to feed an image into the Quick Scan sheet on
/// Linux desktop, because the sheet's upload icon hard-threw "Unsupported
/// platform" before it reached `file_selector`. That switch is gone (see
/// `mobile/CLAUDE.md` → "Picking receipt files"), so `installFileSelectorMock`
/// now works on desktop too. Reach for this one when a spec wants the scanner
/// specifically, or a source neither picker mock covers.
///
/// `getPictures` requests `Permission.camera` **itself** (Dart-side) before
/// invoking its native channel, so we also grant camera/gallery permission here
/// via [installCameraGalleryPermissionMocks] on **every** platform. On Linux
/// `installLinuxDesktopMocks` already grants it; on iOS/Android permission_handler
/// is real, and without this grant the fire-and-forget `requestPermissions()` from
/// `main.dart` leaves an app-init camera request pending (its native dialog is
/// never dismissed in a headless test) so `getPictures`' own camera request
/// collides with it and throws
/// `PlatformException(ERROR_ALREADY_REQUESTING_PERMISSIONS)`. Granting up front
/// makes both requests resolve instantly with no native dialog. (This is the one
/// scan path that needs the grant on mobile; other specs hit permission_handler
/// natively and never trigger a second request.)
///
/// A real temp file is written to disk (not `XFile.fromData`) so `MultipartFile.fromPath`
/// / `File.readAsBytes` behave exactly like the production picker. The bundled
/// `assets/test/sample.png` (16x16 RGBA PNG) ships with the app, so `rootBundle.load`
/// works on every target; the process-scoped tempdir is cleaned up by the OS.
Future<void> installDocumentScannerMock({
  Uint8List? bytes,
  String name = 'sample.png',
}) async {
  installCameraGalleryPermissionMocks();

  final pngBytes = bytes ??
      (await rootBundle.load('assets/test/sample.png')).buffer.asUint8List();
  final tempDir = await Directory.systemTemp.createTemp('doc_scanner_mock_');
  final tempFile = File('${tempDir.path}/$name');
  await tempFile.writeAsBytes(pngBytes, flush: true);

  installDocumentScannerChannelMock(<String>[tempFile.path]);
}
