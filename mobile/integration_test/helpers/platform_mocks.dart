import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../test/helpers/channel_mocks.dart';

/// Re-exported so e2e specs keep importing it from here while the single
/// implementation lives in `test/helpers/channel_mocks.dart` — the widget suite
/// is the gating one and must not reach into `integration_test/`.
export '../../test/helpers/channel_mocks.dart'
    show
        installCameraGalleryPermissionMocks,
        installPermissionMocks,
        installDocumentScannerChannelMock,
        clearPermissionMocks,
        PermissionMockCalls,
        PermissionStatusWire;

/// Re-exported for the same reason: the photo-library fake lives beside the
/// channel mocks in `test/helpers/` so the widget suite can use it too.
export '../../test/helpers/image_picker_mock.dart'
    show installImagePickerMock, installFailingImagePickerMock;

/// Installs mock [MethodChannel] handlers for plugins that don't have a working
/// Linux desktop implementation in this project's runtime environment:
///
/// * **permission_handler** / **gal** — see [installCameraGalleryPermissionMocks].
///   On Linux these method channels have no backing, so the fire-and-forget call
///   from `main.dart` surfaces as an unhandled async exception and fails the test.
/// * **flutter_secure_storage** — has a Linux backend (libsecret) but it
///   requires an unlocked keyring + dbus session. Headless CI/containers
///   don't have that, so reads/writes throw `Libsecret error, Failed to
///   unlock the keyring`. The mock below backs the plugin with an in-memory
///   map — good enough for the smoke test, where we only care that the UI
///   writes + reads tokens correctly within one process.
///
/// Call this from `setUpAll` (or at the top of `main()`) before any test
/// pumps the app. On Android/iOS targets these plugins work natively, so
/// only install the mocks when running on desktop — gate on `Platform.isLinux`.
void installLinuxDesktopMocks() {
  installCameraGalleryPermissionMocks();

  final messenger = TestDefaultBinaryMessengerBinding
      .instance.defaultBinaryMessenger;

  final storage = <String, String>{};
  const secureStorage =
      MethodChannel('plugins.it_nomads.com/flutter_secure_storage');
  messenger.setMockMethodCallHandler(secureStorage, (call) async {
    final args = (call.arguments as Map?)?.cast<String, dynamic>() ?? {};
    final key = args['key'] as String?;
    switch (call.method) {
      case 'write':
        if (key != null) storage[key] = args['value'] as String? ?? '';
        return null;
      case 'read':
        return key == null ? null : storage[key];
      case 'readAll':
        return Map<String, String>.from(storage);
      case 'delete':
        if (key != null) storage.remove(key);
        return null;
      case 'deleteAll':
        storage.clear();
        return null;
      case 'containsKey':
        return key != null && storage.containsKey(key);
      default:
        return null;
    }
  });
}
