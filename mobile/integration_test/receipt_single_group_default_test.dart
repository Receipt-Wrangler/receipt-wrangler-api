// The manual receipt form's Group field, end to end against a real backend.
//
// Two seeding rules, isolated from each other by the account each case uses:
//
//   1. SOLE GROUP -- a freshly provisioned user (no group role, so no fixture
//      group is created) owns exactly their personal "My Receipts" plus the
//      synthetic "All" group. A picker with one option is not a choice, so the
//      form seeds it. The widget tests inject a GroupModel and so prove nothing
//      about the wire; here the group list arrives on AppData at login.
//   2. ROUTE GROUP -- a user with TWO groups, so rule 1 cannot fire. Opening
//      the form from INSIDE a group must still seed that group (the id
//      `openManualReceipt` puts on the route's `extra`, which the form used to
//      discard).
//
// Both assert what the closed dropdown renders, with no `selectDropdown` call:
// the whole point is that no pick was needed.

import 'dart:io' show Platform;

import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'helpers/api.dart';
import 'helpers/form_actions.dart';
import 'helpers/login.dart';
import 'helpers/nav.dart';
import 'helpers/permission_fixtures.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/receipt_test_helpers.dart';

Finder _groupDropdown() => find.byWidgetPredicate(
      (w) => w is FormBuilderDropdown && w.name == 'groupId',
    );

/// Asserts the closed dropdown is showing [groupName] -- i.e. what the user
/// sees on arriving at the form, before touching anything.
void _expectGroupShown(String groupName) {
  expect(
    find.descendant(of: _groupDropdown(), matching: find.text(groupName)),
    findsOneWidget,
  );
}

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  testWidgets(
      'a user with one group opens the manual form on it, without picking',
      (tester) async {
    // No roleName: provisionPermUser only creates a fixture group when a group
    // role resolves, so this account has exactly one selectable group.
    final fixture = await provisionPermUser();

    await loginAs(
      tester,
      username: fixture.username,
      password: fixture.password,
    );

    // Opened from the group-select screen, so no route group is in play -- the
    // seed can only have come from the sole-group rule.
    await openManualReceiptForm(tester);

    _expectGroupShown('My Receipts');

    // And it is a real selection, not just a label: the receipt saves without
    // the Group field ever being touched, and lands in that group.
    await tester.enterText(formField('name'), 'e2e-sole-group');
    await tester.enterText(formField('amount'), '9.99');
    await selectDropdown(tester, 'paidByUserId', fixture.displayName);
    await tester.pumpAndSettle(const Duration(seconds: 3));

    final receiptId = await submitManualReceiptForm(tester);

    // The user's own token, not the admin's: the admin is not a member of
    // someone else's personal group, so both the read and the cleanup 403.
    // Registered after provisionPermUser's user-delete so LIFO runs it first.
    final jwt = await apiLoginAs(fixture.username, fixture.password);
    addTearDown(() async => deleteReceipt(receiptId, jwt: jwt));

    final soleGroup = await firstNonAllGroup(jwt);
    expect(soleGroup.name, 'My Receipts');
    final receipt = await getReceipt(receiptId, jwt: jwt);
    expect(receipt['groupId'], soleGroup.id);
  });

  testWidgets('opening the form from inside a group seeds that group',
      (tester) async {
    // Two groups (personal + fixture), so the sole-group rule cannot account
    // for the seed -- only the route can.
    final fixture = await provisionPermUser(roleName: 'Legacy Owner');

    await loginAs(
      tester,
      username: fixture.username,
      password: fixture.password,
    );
    await enterGroup(tester, fixture.groupName!);

    await openManualReceiptForm(tester);

    _expectGroupShown(fixture.groupName!);
  });
}
