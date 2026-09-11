/// Where a receipt image comes from.
///
/// [camera] is the document scanner (`scanImagesMultiPart`), not a picker.
/// [photos] is the OS photo library — the Android Photo Picker /
/// PHPickerViewController. [files] is the OS document browser, and the only
/// source that can reach a PDF.
///
/// The photo library is a source in its own right because the document browser
/// cannot stand in for it: on iOS `UIDocumentPickerViewController` shows the
/// Files app, and the camera roll is not a Files provider, so photos are
/// unreachable from it entirely.
enum UploadMethod { camera, photos, files }
