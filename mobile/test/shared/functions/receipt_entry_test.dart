import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/shared/functions/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/shared/functions/receipt_entry_availability.dart';
import 'package:receipt_wrangler_mobile/utils/permissions.dart';

import '../../helpers/channel_mocks.dart';
import '../../helpers/permission_test_helpers.dart';
import '../../helpers/receipt_entry_test_helpers.dart';
import '../../helpers/receipt_form_test_helpers.dart';

/// What a Scan tap actually *does*, across every gate combination.
///
/// This is the highest-risk surface of the feature: the tap has one entry point
/// and four possible destinations (camera, gallery fallback, manual form,
/// refusal), and picking the wrong one either hides a capability the user has or
/// offers one they don't.
///
/// What these cases **cannot** see is the `context.mounted` guard after
/// `_refreshBeforeQuickScan`. `TokenRefreshService.reloadAppData()` returns
/// immediately while the service is uninitialized -- which it always is outside
/// `main()` -- so the refresh finishes in a microtask, `tester.pump()` drains
/// microtasks before it builds a frame, and the loading bar the guard trips over
/// never gets one. That is how the receipts-screen overflow menu shipped with
/// two dead items despite this file exercising every gate. The frame-level seam
/// lives in `test/widgets/top_app_bar_loading_indicator_test.dart` and the
/// user-visible behaviour in `integration_test/quick_scan_entry_test.dart`.
void main() {
  const quickScan = api.Permission.groupPeriodReceiptsPeriodQuickScan;
  const create = api.Permission.groupPeriodReceiptsPeriodCreate;

  late List<String> visited;
  late List<Object?> extras;
  late BuildContext capturedContext;

  setUp(() {
    visited = [];
    extras = [];
  });

  tearDown(() {
    clearPermissionMocks();
    debugCameraAccessOverride = null;
  });

  /// Pumps a group-receipts screen wired to the real router, and returns once a
  /// context inside it is available to drive the entry functions with.
  Future<void> pumpEntry(
    WidgetTester tester, {
    required bool aiEnabled,
    required List<api.Permission> permissions,
  }) async {
    final router = GoRouter(
      initialLocation: '/groups/5/receipts',
      routes: [
        GoRoute(
          path: '/groups/:groupId/receipts',
          builder: (context, state) {
            capturedContext = context;
            return const Scaffold(body: SizedBox.shrink());
          },
        ),
        GoRoute(
          path: '/receipts/add',
          builder: (context, state) {
            visited.add('/receipts/add');
            extras.add(state.extra);
            return const Scaffold(body: Text('Add Receipt'));
          },
        ),
      ],
    );

    await tester.pumpWidget(pumpReceiptEntryApp(
      router: router,
      aiEnabled: aiEnabled,
      permissions: seededPermissions(group: {5: permissions}),
      groups: [buildGroup(id: 5, name: 'Household')],
    ));
    await tester.pump();
  }

  group('startScanEntry when Quick Scan is blocked', () {
    testWidgets('ai off falls through to the manual form, carrying the reason',
        (tester) async {
      final calls =
          installPermissionMocks(status: PermissionStatusWire.granted);
      await pumpEntry(tester,
          aiEnabled: false, permissions: [quickScan, create]);

      await startScanEntry(capturedContext);
      await tester.pumpAndSettle();

      expect(visited, ['/receipts/add']);
      final extra = extras.single as Map;
      expect(extra[quickScanBlockedReasonExtraKey],
          QuickScanBlockedReason.aiDisabled);
      expect(extra[quickScanBlockedGroupExtraKey], 'Household');
      expect(calls.statusChecks, 0,
          reason: 'the camera is irrelevant when Quick Scan cannot run at all, '
              'so the user must not be prompted for it');
    });

    testWidgets('a missing quick-scan permission carries that reason instead',
        (tester) async {
      await pumpEntry(tester, aiEnabled: true, permissions: [create]);

      await startScanEntry(capturedContext);
      await tester.pumpAndSettle();

      expect((extras.single as Map)[quickScanBlockedReasonExtraKey],
          QuickScanBlockedReason.noPermission);
    });

    testWidgets('the groupId is preserved so the form knows its group',
        (tester) async {
      await pumpEntry(tester, aiEnabled: true, permissions: [create]);

      await startScanEntry(capturedContext);
      await tester.pumpAndSettle();

      expect((extras.single as Map)['groupId'], '5');
    });

    testWidgets('with neither permission it refuses instead of navigating',
        (tester) async {
      await pumpEntry(tester,
          aiEnabled: true,
          permissions: [api.Permission.groupPeriodReceiptsPeriodRead]);

      await startScanEntry(capturedContext);
      await tester.pump();

      expect(visited, isEmpty);
      expect(find.text(noReceiptEntryPermissionMessage), findsOneWidget);
    });
  });

  group('startScanEntry camera branches', () {
    testWidgets('granted goes to the scanner, not the manual form',
        (tester) async {
      debugCameraAccessOverride = () async => CameraAccess.granted;
      // The scanner re-requests camera permission itself before invoking its
      // native channel, so that channel has to answer or the call never
      // resolves. An empty result is what it returns when the user cancels.
      installPermissionMocks(status: PermissionStatusWire.granted);
      installDocumentScannerChannelMock(const <String>[]);
      await pumpEntry(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      await startScanEntry(capturedContext);
      await tester.pumpAndSettle();

      expect(visited, isEmpty, reason: 'the scanner opens, not the form');
      expect(find.byType(SnackBar), findsNothing,
          reason: 'a cancelled scan is "never mind", not an error');
    });

    testWidgets('denied explains itself and falls back to the gallery',
        (tester) async {
      debugCameraAccessOverride = () async => CameraAccess.denied;
      await pumpEntry(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      await startScanEntry(capturedContext);
      await tester.pump();

      expect(find.text(cameraDeniedFallbackMessage), findsOneWidget);
      expect(find.text('Settings'), findsNothing,
          reason: 'the user can still be re-prompted, so no settings detour');

      // ScaffoldMessenger shows one snackbar at a time, so the gallery's own
      // message is queued behind the notice. Dropping the first surfaces it --
      // and the widget suite runs on a desktop host, where getGalleryImages
      // throws "Unsupported platform", so that second message is the proof the
      // gallery was actually attempted rather than the tap ending silently.
      ScaffoldMessenger.of(capturedContext).removeCurrentSnackBar();
      await tester.pump();

      expect(find.text(galleryUnavailableMessage), findsOneWidget);
      expect(visited, isEmpty);
    }, skip: Platform.isAndroid || Platform.isIOS);

    testWidgets('permanently denied offers the settings escape hatch',
        (tester) async {
      debugCameraAccessOverride = () async => CameraAccess.permanentlyDenied;
      await pumpEntry(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      await startScanEntry(capturedContext);
      await tester.pumpAndSettle();

      expect(find.text(cameraDeniedFallbackMessage), findsOneWidget);
      expect(find.text('Settings'), findsOneWidget,
          reason: 'a re-request would resolve with no dialog, so the OS '
              'settings screen is the only way back to the camera');
    });

    testWidgets('a camera check that throws degrades to the gallery fallback',
        (tester) async {
      // permission_handler talks over a platform channel and has thrown here
      // before; an unhandled async error would reach the user as a red screen.
      debugCameraAccessOverride = () async => throw StateError('channel down');
      await pumpEntry(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      await startScanEntry(capturedContext);
      await tester.pump();

      expect(find.text(cameraDeniedFallbackMessage), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('a user with quick scan but no create still reaches the camera',
        (tester) async {
      debugCameraAccessOverride = () async => CameraAccess.granted;
      installPermissionMocks(status: PermissionStatusWire.granted);
      installDocumentScannerChannelMock(const <String>[]);
      await pumpEntry(tester, aiEnabled: true, permissions: [quickScan]);

      await startScanEntry(capturedContext);
      await tester.pumpAndSettle();

      expect(visited, isEmpty,
          reason: 'falling back to a form they cannot submit would be worse '
              'than the scanner they are allowed to use');
    });
  });

  group('buildReceiptEntryMenuItems', () {
    Future<List<String>> menuLabels(
      WidgetTester tester, {
      required bool aiEnabled,
      required List<api.Permission> permissions,
    }) async {
      await pumpEntry(tester, aiEnabled: aiEnabled, permissions: permissions);
      return buildReceiptEntryMenuItems(capturedContext)
          .whereType<PopupMenuItem>()
          .map((item) => (item.child as Text).data!)
          .toList();
    }

    testWidgets('offers everything when both gates pass', (tester) async {
      expect(
        await menuLabels(tester,
            aiEnabled: true, permissions: [quickScan, create]),
        [quickScanLabel, addManualReceiptLabel, uploadFromGalleryLabel],
      );
    });

    testWidgets('drops both scan entries when Quick Scan cannot run',
        (tester) async {
      // Gallery upload feeds Quick Scan, so it is gated on quick-scan rather
      // than create -- offering it here would produce an unsubmittable sheet.
      expect(
        await menuLabels(tester, aiEnabled: false, permissions: [quickScan, create]),
        [addManualReceiptLabel],
      );
    });

    testWidgets('drops manual entry without the create permission',
        (tester) async {
      expect(
        await menuLabels(tester, aiEnabled: true, permissions: [quickScan]),
        [quickScanLabel, uploadFromGalleryLabel],
      );
    });

    testWidgets('is empty when the user can do neither', (tester) async {
      expect(
        await menuLabels(tester,
            aiEnabled: true,
            permissions: [api.Permission.groupPeriodReceiptsPeriodRead]),
        isEmpty,
      );
    });
  });

  group('openManualReceipt', () {
    testWidgets('attaches no reason when chosen deliberately', (tester) async {
      await pumpEntry(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      openManualReceipt(capturedContext);
      await tester.pumpAndSettle();

      final extra = extras.single as Map;
      expect(extra.containsKey(quickScanBlockedReasonExtraKey), isFalse,
          reason: 'a deliberate "Add Manual Receipt" must not be nagged at '
              'about Quick Scan');
    });
  });
}
