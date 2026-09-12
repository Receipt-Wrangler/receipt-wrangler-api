import 'package:flutter/material.dart';
import 'package:flutter_staggered_grid_view/flutter_staggered_grid_view.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/shared/functions/multi_select_bottom_sheet.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';

/// Layout contract for the shared multi-select sheet — the Categories, Tags and
/// Users pickers all open it through `showMultiselectBottomSheet`.
///
/// With a catalog bigger than one screenful the sheet used to be a dead end: its
/// `MasonryGridView` was `shrinkWrap: true`, so the grid sized itself to its
/// content and had **no scroll extent of its own** — yet it still claimed every
/// vertical drag, which therefore never reached the `SingleChildScrollView` the
/// sheet wrapped the body in. The chips cover the whole sheet, so no drag a user
/// could make scrolled anything. On top of that the confirm button was pinned as
/// `Scaffold.bottomSheet`, which *floats over* the body and buried the last row.
///
/// These assert **geometry, not finders**: `findsOneWidget` is true for a chip
/// clipped outside its viewport, and `tester.ensureVisible` walks every ancestor
/// scrollable, so both pass on the broken sheet. Same lesson as the Quick Scan
/// sheet — see `mobile/CLAUDE.md`.
void main() {
  // Comfortably more than the ~13 rows a phone-sized sheet can show at three
  // chips per row, so the grid must overflow whatever else changes.
  const optionCount = 90;
  final options = List.generate(
    optionCount,
    (i) => 'Option ${i.toString().padLeft(2, '0')}',
  );

  /// Opens the real sheet the pickers use, on a phone-sized surface.
  Future<void> pumpSheet(WidgetTester tester) async {
    tester.view.physicalSize = const Size(390, 844);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    late BuildContext sheetContext;

    await tester.pumpWidget(
      MultiProvider(
        providers: [
          // The sheet's TopAppBar reads AuthModel; BottomSubmitButton reads
          // LoadingModel.
          ChangeNotifierProvider<AuthModel>(create: (_) => AuthModel()),
          ChangeNotifierProvider<LoadingModel>(create: (_) => LoadingModel()),
        ],
        child: MaterialApp(
          home: Builder(
            builder: (context) {
              sheetContext = context;
              return const Scaffold(body: SizedBox.shrink());
            },
          ),
        ),
      ),
    );
    await tester.pump();

    showMultiselectBottomSheet(
      sheetContext,
      'Select Options',
      'Select',
      options,
      <String>[],
      (option) => option as String,
    );
    await tester.pumpAndSettle();
  }

  /// The grid's **own** scroll position — the one a drag on the chips has to
  /// move. Reading it (rather than any ancestor scrollable) is what makes these
  /// tests fail on the pre-fix sheet.
  ScrollableState gridScrollable(WidgetTester tester) => tester.state<
      ScrollableState>(find.descendant(
    of: find.byType(MasonryGridView),
    matching: find.byType(Scrollable),
  ));

  /// Scrolls the grid as far as it goes. Repeated because a lazily-built list
  /// only refines `maxScrollExtent` as more of it is laid out.
  Future<void> scrollGridToEnd(WidgetTester tester) async {
    for (var i = 0; i < 5; i++) {
      final position = gridScrollable(tester).position;
      if (position.pixels >= position.maxScrollExtent) {
        break;
      }
      position.jumpTo(position.maxScrollExtent);
      await tester.pumpAndSettle();
    }
  }

  Finder filterField() => find.byWidgetPredicate(
        (widget) => widget is TextField && widget.decoration?.labelText == 'Filter',
      );

  testWidgets('the option list scrolls when dragged on the chips',
      (tester) async {
    await pumpSheet(tester);

    expect(gridScrollable(tester).position.pixels, 0);

    await tester.drag(find.byType(MasonryGridView), const Offset(0, -300));
    await tester.pumpAndSettle();

    expect(gridScrollable(tester).position.pixels, greaterThan(0));
  });

  testWidgets('the last option is fully visible and clear of the confirm button',
      (tester) async {
    await pumpSheet(tester);
    await scrollGridToEnd(tester);

    final lastChip = find.widgetWithText(ChoiceChip, options.last);
    expect(lastChip, findsOneWidget);

    final chipRect = tester.getRect(lastChip);

    // On screen. This is the assertion the pre-fix sheet fails: its grid was
    // shrink-wrapped inside a SingleChildScrollView, so it laid out its full
    // content height and the last chip sat hundreds of pixels below the window
    // with no reachable scrollable to bring it up.
    final windowRect = Offset.zero & tester.view.physicalSize / tester.view.devicePixelRatio;
    expect(chipRect.top, greaterThanOrEqualTo(windowRect.top - 0.5));
    expect(chipRect.bottom, lessThanOrEqualTo(windowRect.bottom + 0.5));

    // And not buried under the confirm button, which now reserves its space as
    // the scaffold's bottom bar instead of floating over the body.
    expect(
      chipRect.overlaps(tester.getRect(find.byType(BottomSubmitButton))),
      isFalse,
    );
  });

  testWidgets('the filter bar stays pinned while the list scrolls',
      (tester) async {
    await pumpSheet(tester);

    final before = tester.getRect(filterField());
    await scrollGridToEnd(tester);

    expect(gridScrollable(tester).position.pixels, greaterThan(0));
    expect(tester.getRect(filterField()), before);
  });

  testWidgets('a selection made after scrolling is returned to the caller',
      (tester) async {
    await pumpSheet(tester);
    await scrollGridToEnd(tester);

    await tester.tap(find.widgetWithText(ChoiceChip, options.last));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(BottomSubmitButton, 'Select'));
    await tester.pumpAndSettle();

    // The sheet is gone and nothing threw on the way out.
    expect(find.byType(MasonryGridView), findsNothing);
    expect(tester.takeException(), isNull);
  });
}
