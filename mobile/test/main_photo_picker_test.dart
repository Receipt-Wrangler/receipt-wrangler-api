import 'package:flutter_test/flutter_test.dart';
import 'package:image_picker_android/image_picker_android.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:receipt_wrangler_mobile/main.dart';

// configureAndroidPhotoPicker opts into the Android Photo Picker / PHPicker on
// API 33-35 (a no-op on 36+, backported below 33 via Play Services). It runs in
// main(), which no test pumps -- the e2e suite pumps buildApp() -- so the branch
// is only reachable through the injectable parameter.
//
// What is pinned here is the BRANCH, not the plugin registration: that
// ImagePickerAndroid gets the opt-in and that any other platform
// implementation is left exactly as it was. The real registration still has to
// be verified on a physical Android device.

/// A stand-in for a non-Android implementation (iOS, Linux, a test fake). The
/// interface is `platform_interface`-style, so implementing it requires the
/// mixin rather than plain `implements`.
class _NonAndroidImagePicker extends ImagePickerPlatform {}

void main() {
  group('configureAndroidPhotoPicker', () {
    test('enables the photo picker on the Android implementation', () {
      final picker = ImagePickerAndroid();
      expect(picker.useAndroidPhotoPicker, isFalse,
          reason: 'precondition: the plugin default is off');

      configureAndroidPhotoPicker(picker);

      expect(picker.useAndroidPhotoPicker, isTrue);
    });

    test('leaves a non-Android implementation untouched', () {
      final picker = _NonAndroidImagePicker();

      // The type check is the whole guard: without it this would throw, since
      // useAndroidPhotoPicker exists only on ImagePickerAndroid.
      expect(() => configureAndroidPhotoPicker(picker), returnsNormally);
      expect(ImagePickerPlatform.instance, isNot(same(picker)),
          reason: 'configuring must not install the picker it was handed');
    });

    test('is idempotent', () {
      // main() runs once per launch, but a hot restart re-enters it against the
      // same already-configured plugin instance.
      final picker = ImagePickerAndroid();

      configureAndroidPhotoPicker(picker);
      configureAndroidPhotoPicker(picker);

      expect(picker.useAndroidPhotoPicker, isTrue);
    });
  });
}
