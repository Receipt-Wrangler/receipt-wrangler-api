import 'package:flutter/material.dart';

/// Stands in for a file Flutter cannot decode into an image.
///
/// The file source legitimately accepts PDFs — the backend converts them to JPG
/// on upload (`FileRepository.GetBytesFromImageBytes`), and quick scan OCRs them
/// end to end — so a PDF *upload succeeds* while `Image.memory` on its bytes
/// fails. Without this the user is shown Flutter's grey broken-image glyph and
/// reasonably concludes the upload broke.
///
/// Also covers the degenerate decode cases on the server-bytes side: an absent
/// `encodedImage` funnels through as an empty `Uint8List`, which throws exactly
/// the same way.
class UnrenderableFilePlaceholder extends StatelessWidget {
  const UnrenderableFilePlaceholder({super.key, this.filename});

  /// Shown under the icon when known. Picked photos on iOS carry a synthetic
  /// name (`image_picker_XXXX.jpg`), so this is a hint rather than a promise.
  final String? filename;

  bool get _isPdf => filename?.toLowerCase().endsWith(".pdf") ?? false;

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.onSurfaceVariant;

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(_isPdf ? Icons.picture_as_pdf : Icons.insert_drive_file,
                size: 48, color: color),
            const SizedBox(height: 8),
            Text(
              filename ?? "Preview unavailable",
              key: const ValueKey("unrenderable-file-name"),
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: color),
            ),
          ],
        ),
      ),
    );
  }
}

/// [UnrenderableFilePlaceholder] as an [Image.errorBuilder], so the call sites
/// stay one-liners.
ImageErrorWidgetBuilder unrenderableFileErrorBuilder({String? filename}) {
  return (_, __, ___) => UnrenderableFilePlaceholder(filename: filename);
}
