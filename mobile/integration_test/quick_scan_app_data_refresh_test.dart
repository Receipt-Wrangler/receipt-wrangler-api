// Starting a Quick Scan re-fetches AppData first, so a group config changed
// elsewhere (an admin on desktop, say) reaches a RUNNING app.
//
// This is the one spec that proves the round trip. Every other quick-scan spec
// arranges its config before login, so it cannot tell a client that re-fetches
// from one that merely read the config once at login. Here the config is flipped
// *after* login and the app is never restarted: without the pre-scan reload the
// comment field is simply absent, because GroupModel still holds the copy AppData
// delivered at login.
//
// Deliberately no `configureFirstGroup` / GroupModel mutation anywhere — a
// client-side injection would prove nothing about the refresh.

import 'dart:io' show Platform;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';

import 'helpers/api.dart';
import 'helpers/document_scanner_mock.dart';
import 'helpers/feature_flags.dart';
import 'helpers/form_actions.dart';
import 'helpers/login.dart';
import 'helpers/permission_fixtures.dart';
import 'helpers/platform_mocks.dart';
import 'helpers/pump.dart';
import 'helpers/quick_scan_actions.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    if (Platform.isLinux) {
      installLinuxDesktopMocks();
    }
  });

  testWidgets('a comment field enabled after login reaches a running app',
      (tester) async {
    await enableAiPoweredReceiptsForTest();
    await installDocumentScannerMock();

    final jwt = await apiLogin();
    final group = await firstNonAllGroup(jwt);

    // Paid-by and status stay shown+required throughout: the backend rejects a
    // config that lets either be skipped without a default to backfill it.
    Future<void> setComment(bool enabled) async =>
        setGroupQuickScanConfig(
          groupId: group.id,
          jwt: await apiLogin(),
          overrides: {
            'hideComments': false,
            'quickScanPaidByEnabled': true,
            'quickScanPaidByRequired': true,
            'quickScanStatusEnabled': true,
            'quickScanStatusRequired': true,
            'quickScanCategoriesEnabled': false,
            'quickScanTagsEnabled': false,
            'quickScanCommentEnabled': enabled,
            'quickScanCommentRequired': false,
          },
        );

    // Comment OFF, then log in -- so the app's AppData snapshot says "no comment".
    await setComment(false);
    await loginAsAdmin(tester);

    final groupModel = Provider.of<GroupModel>(
        tester.element(find.byType(Scaffold).first),
        listen: false);
    expect(
        groupModel
            .getGroupReceiptSettings(group.id)
            ?.quickScanCommentEnabled ??
            false,
        isFalse,
        reason: 'precondition: the app started with the comment field disabled');

    // Someone turns it on elsewhere. The app is not told and is not restarted.
    await setComment(true);

    // Opening Quick Scan must re-fetch before the scanner, so the form built
    // moments later renders the field the server now says is enabled.
    await openQuickScanImageForm(tester);
    await selectDropdown(tester, 'groupId', group.name);

    await pumpUntilFound(tester, quickScanCommentField());
    expect(quickScanCommentField(), findsOneWidget,
        reason: 'the comment field enabled after login is present');
  });
}
