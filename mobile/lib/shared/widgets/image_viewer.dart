import 'package:flutter/material.dart';

class ImageViewer extends StatefulWidget {
  const ImageViewer({super.key, required this.image});

  /// A `Widget`, not an `Image`: a source that cannot be decoded renders the
  /// unrenderable-file placeholder instead, and that is not an `Image`.
  final Widget image;

  @override
  State<ImageViewer> createState() => _ImageViewer();
}

class _ImageViewer extends State<ImageViewer> {
  final TransformationController _transformationController =
      TransformationController();

  @override
  Widget build(BuildContext context) {
    return Stack(children: [
      Center(
        child: InteractiveViewer(
          minScale: 0.1,
          maxScale: 10.0,
          transformationController: _transformationController,
          child: widget.image,
        ),
      ),
    ]);
  }
}
