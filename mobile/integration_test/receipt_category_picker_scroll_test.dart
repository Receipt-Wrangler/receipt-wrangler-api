import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_staggered_grid_view/flutter_staggered_grid_view.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/bottom_submit_button.dart';

import 'helpers/api.dart';
import 'helpers/form_actions.dart';
import 'helpers/login.dart';
import 'helpers/nav.dart';
import 'helpers/permission_fixtures.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/receipt_test_helpers.dart';

/// The category/tag picker sheet has to be scrollable with a real catalog.
///
/// This is an e2e rather than a widget test because the failure is about the
/// **sheet's geometry on a real window** — how much of the catalog fits, and
/// what the confirm button covers. `test/widgets/filter_multiselect_test.dart`
/// pins the same contract in isolation; this proves it against a catalog that
/// actually arrived over the wire.
///
/// Before the fix the grid was `shrinkWrap: true` inside the sheet's
/// `SingleChildScrollView`: it had no scroll extent of its own but still claimed
/// every vertical drag, so a user could not reach past the first screenful of
/// categories — and the confirm button, pinned as `Scaffold.bottomSheet`,
/// floated over the last row.
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  // The Linux runner's window is 1280x720 and `binding.setSurfaceSize` is a
  // no-op there, so the catalog has to be big enough to overflow that: 60
  // categories is 20 rows at three chips per row, well past what the sheet's
  // body can show.
  const categoryCount = 60;

  /// The grid's **own** scroll position — the one a drag on the chips has to
  /// move. An ancestor scrollable moving would not mean the chips did.
  ScrollableState gridScrollable(WidgetTester tester) =>
      tester.state<ScrollableState>(find.descendant(
        of: find.byType(MasonryGridView),
        matching: find.byType(Scrollable),
      ));

  testWidgets('the category picker scrolls through a large catalog',
      (tester) async {
    final adminJwt = await apiLogin();
    final stamp = DateTime.now().microsecondsSinceEpoch;

    // Names sort and read predictably in the recording, and the zero-padded
    // index makes "the last one" unambiguous on screen.
    final names = List.generate(
      categoryCount,
      (i) => 'e2e-scroll-$stamp-${i.toString().padLeft(2, '0')}',
    );
    final ids = <int>[];
    for (final name in names) {
      ids.add(await createCategory(name: name, jwt: adminJwt));
    }
    addTearDown(() async {
      final jwt = await apiLogin();
      for (final id in ids) {
        await deleteCategory(id, jwt: jwt);
      }
    });

    final fixture = await provisionPermUser(roleName: 'Legacy Editor');
    await loginAs(
      tester,
      username: fixture.username,
      password: fixture.password,
    );
    await enterGroup(tester, fixture.groupName!);

    await openManualReceiptForm(tester);
    await pumpUntilFound(tester, find.text('Name'));
    await selectDropdown(tester, 'groupId', fixture.groupName!);

    final field = find.text('No Categories selected');
    await pumpUntilFound(tester, field);
    await tester.ensureVisible(field);
    await tester.pump(const Duration(milliseconds: 200));
    await tester.tap(field);
    await pumpUntilFound(tester, find.byType(MasonryGridView));

    // The list is longer than the sheet, and a drag on the chips scrolls it.
    // Both are zero on the pre-fix sheet.
    final position = gridScrollable(tester).position;
    expect(position.maxScrollExtent, greaterThan(0));

    await tester.drag(find.byType(MasonryGridView), const Offset(0, -250));
    await tester.pumpAndSettle();
    expect(position.pixels, greaterThan(0));

    // Now reach the very end. `jumpTo` after the drags for determinism -- a
    // lazily-built list only refines maxScrollExtent as it lays out.
    for (var i = 0; i < 5 && position.pixels < position.maxScrollExtent; i++) {
      position.jumpTo(position.maxScrollExtent);
      await tester.pumpAndSettle();
    }

    final lastChip = find.widgetWithText(ChoiceChip, names.last);
    expect(lastChip, findsOneWidget);

    // Geometry, not finders: a clipped chip still satisfies `findsOneWidget`,
    // and `ensureVisible` would drag it into view even where a user could not.
    final chipRect = tester.getRect(lastChip);
    final windowRect =
        Offset.zero & tester.view.physicalSize / tester.view.devicePixelRatio;
    expect(chipRect.top, greaterThanOrEqualTo(windowRect.top - 0.5));
    expect(chipRect.bottom, lessThanOrEqualTo(windowRect.bottom + 0.5));

    // Scoped by its label: the receipt form underneath the sheet has a
    // BottomSubmitButton of its own, so `find.byType` is ambiguous here.
    final selectButton = find.widgetWithText(BottomSubmitButton, 'Select');
    expect(selectButton, findsOneWidget);
    expect(chipRect.overlaps(tester.getRect(selectButton)), isFalse);

    // And it can actually be picked from there — the point of reaching it.
    await tester.tap(lastChip);
    await tester.pumpAndSettle();
    await pumpUntilFound(tester, selectButton.hitTestable());
    await tester.tap(selectButton);
    await tester.pumpAndSettle();

    await pumpUntilFound(tester, find.text(names.last));
    expect(find.text(names.last), findsWidgets);
  });
}
