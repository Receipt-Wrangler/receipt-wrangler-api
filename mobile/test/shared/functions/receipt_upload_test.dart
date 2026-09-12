import 'package:file_selector_platform_interface/file_selector_platform_interface.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:plugin_platform_interface/plugin_platform_interface.dart';
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/enums/upload_method.dart';
import 'package:receipt_wrangler_mobile/shared/functions/receipt_upload.dart';

import '../../helpers/channel_mocks.dart';
import '../../helpers/receipt_entry_test_helpers.dart';
import '../../helpers/image_picker_mock.dart';

/// Records which source was asked for, so a test can prove the dispatcher
/// routed to the right picker rather than merely returning something.
class _RecordingFileSelector extends FileSelectorPlatform
    with MockPlatformInterfaceMixin {
  int openFilesCalls = 0;
  List<XTypeGroup>? lastTypeGroups;

  @override
  Future<List<XFile>> openFiles({
    List<XTypeGroup>? acceptedTypeGroups,
    String? initialDirectory,
    String? confirmButtonText,
  }) async {
    openFilesCalls++;
    lastTypeGroups = acceptedTypeGroups;
    return <XFile>[];
  }
}

class _RecordingImagePicker extends ImagePickerPlatform
    with MockPlatformInterfaceMixin {
  int pickCalls = 0;

  @override
  Future<List<XFile>> getMultiImageWithOptions({
    MultiImagePickerOptions options = const MultiImagePickerOptions(),
  }) async {
    pickCalls++;
    return <XFile>[];
  }
}

class _FailingFileSelector extends FileSelectorPlatform
    with MockPlatformInterfaceMixin {
  @override
  Future<List<XFile>> openFiles({
    List<XTypeGroup>? acceptedTypeGroups,
    String? initialDirectory,
    String? confirmButtonText,
  }) async {
    throw Exception('file picker unavailable');
  }
}

void main() {
  late BuildContext capturedContext;

  /// A bare Scaffold is enough: the dispatcher only needs a context with a
  /// ScaffoldMessenger above it.
  Future<void> pumpHost(WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: Builder(builder: (context) {
          capturedContext = context;
          return const SizedBox.shrink();
        }),
      ),
    ));
    await tester.pump();
  }

  group('acquireReceiptFiles routes to the right source', () {
    testWidgets('photos goes to the photo picker, not the file browser',
        (tester) async {
      final picker = _RecordingImagePicker();
      final selector = _RecordingFileSelector();
      ImagePickerPlatform.instance = picker;
      FileSelectorPlatform.instance = selector;
      await pumpHost(tester);

      await acquireReceiptFiles(capturedContext, UploadMethod.photos);

      expect(picker.pickCalls, 1);
      expect(selector.openFilesCalls, 0,
          reason: 'the whole point of the photo source is that it is not the '
              'document picker');
    });

    testWidgets('files goes to the file browser, not the photo picker',
        (tester) async {
      final picker = _RecordingImagePicker();
      final selector = _RecordingFileSelector();
      ImagePickerPlatform.instance = picker;
      FileSelectorPlatform.instance = selector;
      await pumpHost(tester);

      await acquireReceiptFiles(capturedContext, UploadMethod.files);

      expect(selector.openFilesCalls, 1);
      expect(picker.pickCalls, 0);
    });

    testWidgets('the file source accepts PDFs as well as images',
        (tester) async {
      final selector = _RecordingFileSelector();
      FileSelectorPlatform.instance = selector;
      await pumpHost(tester);

      await acquireReceiptFiles(capturedContext, UploadMethod.files);

      final group = selector.lastTypeGroups!.single;
      expect(group.mimeTypes, contains('application/pdf'),
          reason: 'PDF receipts are a supported input; the backend converts '
              'them server-side');
      expect(group.uniformTypeIdentifiers, contains('com.adobe.pdf'),
          reason: 'one group carries every platform spelling, which is what '
              'removed the Platform.operatingSystem switch');
      expect(group.mimeTypes, contains('image/*'));
    });
  });

  group('acquireReceiptFiles failure handling', () {
    setUp(() {
      installCameraGalleryPermissionMocks();
      addTearDown(clearPermissionMocks);
    });

    testWidgets('a failing photo picker explains itself and yields nothing',
        (tester) async {
      installFailingImagePickerMock();
      await pumpHost(tester);

      final result =
          await acquireReceiptFiles(capturedContext, UploadMethod.photos);
      await tester.pump();

      expect(result, isEmpty);
      expect(find.text(photoPickerUnavailableMessage), findsOneWidget);
    });

    testWidgets('a failing file picker gets its own message', (tester) async {
      FileSelectorPlatform.instance = _FailingFileSelector();
      await pumpHost(tester);

      final result =
          await acquireReceiptFiles(capturedContext, UploadMethod.files);
      await tester.pump();

      expect(result, isEmpty);
      expect(find.text(filePickerUnavailableMessage), findsOneWidget,
          reason: 'the two sources fail for different reasons, so one shared '
              '"gallery" message would misdescribe half the cases');
      expect(find.text(photoPickerUnavailableMessage), findsNothing);
    });
  });

  group('acquireReceiptFiles success', () {
    testWidgets('a picked photo carries a non-empty filename', (tester) async {
      await pumpHost(tester);

      // `runAsync` is mandatory here, not stylistic: the mock writes a real
      // temp file and the picker reads it back, and `testWidgets` otherwise
      // runs in fake-async where dart:io futures never complete -- the test
      // hangs until the harness gives up rather than failing.
      //
      // Explicit bytes rather than the asset default for the same class of
      // reason: `rootBundle` is not serviced in the widget suite the way it is
      // under the integration binding.
      final result = await tester.runAsync(() async {
        await installImagePickerMock(
            bytes: quickScanTestPngBytes, name: 'receipt.png');
        return acquireReceiptFiles(capturedContext, UploadMethod.photos);
      });

      expect(result, hasLength(1));
      expect(result!.single.filename, 'receipt.png',
          reason: 'an empty multipart filename makes the Go parser close the '
              'connection, saving a receipt with zero images');
      expect(result.single.bytes, isNotEmpty);
    });
  });
}
