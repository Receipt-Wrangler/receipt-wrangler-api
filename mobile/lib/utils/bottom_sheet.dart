import 'package:flutter/material.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/screen_wrapper.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/top_app_bar.dart';

/// [bodyFillsSheet] is for a [child] that manages its own scrolling and needs to
/// size itself to the space available (a carousel, a `ListView`). It changes two
/// things, both of which that kind of child needs:
///
///  * the body is **not** wrapped in a `SingleChildScrollView`. The wrapper makes
///    the body's height unbounded, leaving such a child nothing to measure
///    against -- and guessing the screen height overflows the sheet, whose body
///    is shorter than the screen by the app bar, drag handle and safe area.
///  * [bottomSheetWidget] is pinned as the scaffold's bottom bar rather than as
///    its `bottomSheet`. A `Scaffold.bottomSheet` **floats over** the body, so
///    the tail of the content sits underneath it and cannot be scrolled clear;
///    a bottom bar reserves its space, so the body shrinks to fit instead.
///
/// Both defaults keep the existing behaviour for every other sheet.
showFullscreenBottomSheet(BuildContext context, Widget child, String label,
    {List<Widget>? actions,
    Widget? bottomSheetWidget,
    EdgeInsets? bodyPadding,
    bool bodyFillsSheet = false}) {
  return showModalBottomSheet(
    context: context,
    enableDrag: true,
    isDismissible: true,
    useSafeArea: true,
    isScrollControlled: true,
    showDragHandle: true,
    constraints: BoxConstraints(),
    builder: (BuildContext context) {
      return StatefulBuilder(builder:
          (BuildContext context, void Function(void Function()) setState) {
        return ScreenWrapper(
            bodyPadding: bodyPadding,
            appBarWidget: TopAppBar(
              titleText: label,
              actions: actions,
              hideAvatar: true,
              surfaceTintColor: Colors.white,
            ),
            bottomSheetWidget: bodyFillsSheet ? null : bottomSheetWidget,
            bottomNavigationBarWidget:
                bodyFillsSheet ? bottomSheetWidget : null,
            child:
                bodyFillsSheet ? child : SingleChildScrollView(child: child));
      });
    },
  );
}
