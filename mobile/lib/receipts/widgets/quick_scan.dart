import 'package:flutter/material.dart';
import 'package:infinite_carousel/infinite_carousel.dart';
import 'package:receipt_wrangler_mobile/receipts/widgets/quick_scan_form.dart';
import 'package:receipt_wrangler_mobile/shared/classes/quick_scan_image.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/image_viewer.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';
import 'package:rxdart/rxdart.dart';

class QuickScan extends StatelessWidget {
  const QuickScan(
      {super.key,
      required this.imageSubject,
      required this.infiniteScrollController,
      required this.isCompletedSubject});

  final BehaviorSubject<List<QuickScanImage>> imageSubject;

  final InfiniteScrollController infiniteScrollController;

  final BehaviorSubject<bool> isCompletedSubject;

  Widget _buildImagePreview(BuildContext context, int index) {
    var image = Image.memory(imageSubject.value[index].bytes);
    return SizedBox(
        height: getImagePreviewHeight(context),
        width: getImagePreviewWidth(context),
        child: ImageViewer(image: image));
  }

  /// An arrow, or the space one would take.
  ///
  /// The ends of the scan drop their arrow rather than disabling it, so the
  /// arrows that remain always lead somewhere. The placeholder keeps the
  /// counter between them centered instead of letting it slide as the arrows
  /// come and go -- [kMinInteractiveDimension] is the `IconButton` default
  /// constraint, so the swap is exactly the same size.
  Widget _buildNavArrow(
      {required bool show,
      required Key key,
      required IconData icon,
      required String tooltip,
      required VoidCallback onPressed}) {
    if (!show) {
      return const SizedBox.square(dimension: kMinInteractiveDimension);
    }

    return IconButton(
      key: key,
      icon: Icon(icon),
      tooltip: tooltip,
      onPressed: onPressed,
    );
  }

  /// Tells the user there is more than one page, and gives them a way to reach
  /// it that isn't a swipe.
  ///
  /// A scan can carry up to 100 pages, and the carousel gives no other hint
  /// that the pages after the first exist -- so without this a multi-page scan
  /// looks like a single-page one, and the forms behind it go unfilled. The
  /// swipe alone left that hint to a line of text, which is easy to skip past.
  ///
  /// [index] is the page this row belongs to, not the page in view: the
  /// carousel builds only the item on screen, and taking the index from the
  /// item rather than tracking the selection means a page can never render an
  /// arrow that points off the end of a scan an image was just deleted from.
  Widget _buildPageNav(BuildContext context, int index, int pageCount) {
    if (pageCount < 2) {
      return const SizedBox.shrink();
    }

    return Padding(
      padding: const EdgeInsets.only(top: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          _buildNavArrow(
            show: index > 0,
            key: const ValueKey('quick-scan-previous-page'),
            icon: Icons.chevron_left,
            tooltip: 'Previous page',
            onPressed: () => infiniteScrollController.previousItem(),
          ),
          Text(
            key: const ValueKey('quick-scan-page-indicator'),
            '${index + 1} of $pageCount',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          _buildNavArrow(
            show: index < pageCount - 1,
            key: const ValueKey('quick-scan-next-page'),
            icon: Icons.chevron_right,
            tooltip: 'Next page',
            onPressed: () => infiniteScrollController.nextItem(),
          ),
        ],
      ),
    );
  }

  Widget _buildCarousel(BuildContext context, bool isCompleted) {
    return InfiniteCarousel.builder(
      itemCount: imageSubject.value.length,
      itemExtent: MediaQuery.of(context).size.width,
      center: false,
      velocityFactor: 0.2,
      controller: infiniteScrollController,
      axisDirection: Axis.horizontal,
      loop: false,
      itemBuilder: (BuildContext context, int itemIndex, int realIndex) {
        return SingleChildScrollView(
          child: Column(
            children: [
              _buildImagePreview(context, realIndex),
              _buildPageNav(context, realIndex, imageSubject.value.length),
              Padding(
                padding: getImageDataPadding(),
                child: QuickScanForm(
                  // Keyed by the image, not by position. Deleting a page shifts
                  // every later image down an index, and without a key Flutter
                  // would match by position and hand the surviving form the old
                  // page's State -- whose `groupId` (seeded once in initState)
                  // would then disagree with the group the dropdown shows, so the
                  // form would resolve its field set against the wrong group.
                  key: ObjectKey(imageSubject.value[realIndex]),
                  formKey: imageSubject.value[realIndex].formKey,
                  image: imageSubject.value[realIndex],
                  index: realIndex,
                  enabled: !isCompleted,
                  onFormChangeCallback: (values) {
                    var newImage = imageSubject.value[realIndex];
                    newImage.groupId = values.groupId;
                    newImage.paidByUserId = values.paidByUserId;
                    newImage.status = values.status;
                    newImage.categories = values.categories;
                    newImage.tags = values.tags;
                    newImage.comment = values.comment;

                    var newImages = imageSubject.value;

                    newImages[realIndex] = newImage;
                    imageSubject.add(newImages);
                  },
                ),
              )
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<bool>(
        stream: isCompletedSubject.stream,
        builder: (context, completedSnapshot) {
          final isCompleted = completedSnapshot.hasData && completedSnapshot.data == true;

          return StreamBuilder<List<QuickScanImage>>(
              stream: imageSubject.stream,
              builder: (context, imageSnapshot) {
                if (imageSnapshot.hasData && imageSnapshot.data!.isEmpty) {
                  return const Center(
                    child: Text("Scan or upload an image to get started"),
                  );
                }

                if (imageSnapshot.hasData && imageSnapshot.data!.isNotEmpty) {
                  // Fill the sheet's body, not the screen. The carousel is
                  // horizontal, so it needs a bounded height from somewhere --
                  // but the sheet's body is shorter than the screen by the app
                  // bar, drag handle and safe area. Sizing to the screen
                  // overflowed the body, and because each slide's own
                  // SingleChildScrollView then believed it had a full-screen
                  // viewport it never scrolled, so the overflow was clipped and
                  // unreachable: a configured comment field (last in the form)
                  // simply could not be seen. Bounded constraints reach us
                  // because the sheet no longer wraps the body in a scroll view
                  // (`bodyFillsSheet: true` in showQuickScanBottomSheet).
                  return SizedBox.expand(
                      child: _buildCarousel(context, isCompleted));
                }

                return const SizedBox.shrink();
              });
        });
  }
}
