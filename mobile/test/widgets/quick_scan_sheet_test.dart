import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:infinite_carousel/infinite_carousel.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/constants/receipt_entry.dart';
import 'package:receipt_wrangler_mobile/receipts/widgets/quick_scan.dart';
import 'package:receipt_wrangler_mobile/shared/functions/quick_scan.dart';
import 'package:rxdart/rxdart.dart';

import 'package:file_selector_platform_interface/file_selector_platform_interface.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:plugin_platform_interface/plugin_platform_interface.dart';
import 'package:receipt_wrangler_mobile/enums/upload_method.dart';

import '../helpers/channel_mocks.dart';
import '../helpers/permission_test_helpers.dart';
import '../helpers/receipt_entry_test_helpers.dart';
import '../helpers/receipt_form_test_helpers.dart';

/// The Quick Scan sheet is now reachable from the nav tap, the long-press menu
/// and the overflow menu, so it re-checks its own gates rather than trusting the
/// caller -- and it offers a way out to manual entry only for users who could
/// actually save one.
class _RecordingImagePicker extends ImagePickerPlatform
    with MockPlatformInterfaceMixin {
  int pickCalls = 0;

  @override
  Future<List<XFile>> getMultiImageWithOptions({
    MultiImagePickerOptions options = const MultiImagePickerOptions(),
  }) async {
    pickCalls++;
    return <XFile>[];
  }
}

class _RecordingFileSelector extends FileSelectorPlatform
    with MockPlatformInterfaceMixin {
  int openFilesCalls = 0;

  @override
  Future<List<XFile>> openFiles({
    List<XTypeGroup>? acceptedTypeGroups,
    String? initialDirectory,
    String? confirmButtonText,
  }) async {
    openFilesCalls++;
    return <XFile>[];
  }
}

void main() {
  const quickScan = api.Permission.groupPeriodReceiptsPeriodQuickScan;
  const create = api.Permission.groupPeriodReceiptsPeriodCreate;
  const manualLinkKey = ValueKey('quick-scan-manual-entry-link');
  const pageIndicatorKey = ValueKey('quick-scan-page-indicator');
  const previousPageKey = ValueKey('quick-scan-previous-page');
  const nextPageKey = ValueKey('quick-scan-next-page');

  late BuildContext capturedContext;

  Future<void> pumpSheet(
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
          builder: (context, state) => const Scaffold(body: Text('Add screen')),
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

    showQuickScanBottomSheet(capturedContext);
    await tester.pumpAndSettle();
  }

  /// Mounts the carousel with [pageCount] pages already loaded.
  ///
  /// The sheet builds its own `imageSubject` internally and seeds it empty, so
  /// [pumpSheet] can never reach a loaded state -- it can only prove the counter
  /// is absent. Mounting [QuickScan] directly is what lets the counter be
  /// asserted present, which is the whole point of keying it.
  Future<InfiniteScrollController> pumpCarousel(
    WidgetTester tester, {
    required int pageCount,
  }) async {
    final images = List.generate(pageCount, (_) => buildQuickScanImage(groupId: 5));
    final controller = InfiniteScrollController();

    final router = GoRouter(
      initialLocation: '/',
      routes: [
        GoRoute(
          path: '/',
          builder: (context, state) => Scaffold(
            body: QuickScan(
              imageSubject: BehaviorSubject.seeded(images),
              infiniteScrollController: controller,
              isCompletedSubject: BehaviorSubject.seeded(false),
            ),
          ),
        ),
      ],
    );

    await tester.pumpWidget(pumpReceiptEntryApp(
      router: router,
      aiEnabled: true,
      permissions: seededPermissions(group: {5: [quickScan, create]}),
      groups: [buildGroup(id: 5, name: 'Household')],
    ));
    await tester.pump();

    return controller;
  }

  group('the upload source menu', () {
    const menuKey = ValueKey('quick-scan-upload-source-menu');

    testWidgets('offers both picker sources behind one icon', (tester) async {
      await pumpSheet(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      // Closed, it is one icon -- the sheet's app bar already carries the
      // scanner and delete beside the title.
      expect(find.byKey(menuKey), findsOneWidget);
      expect(find.text(uploadPhotoLabel), findsNothing);
      expect(find.text(uploadFileLabel), findsNothing);

      await tester.tap(find.byKey(menuKey));
      await tester.pumpAndSettle();

      expect(find.text(uploadPhotoLabel), findsOneWidget);
      expect(find.text(uploadFileLabel), findsOneWidget);
    });

    testWidgets('a denied camera explains itself instead of escaping',
        (tester) async {
      // acquireReceiptFiles RETHROWS CunningDocumentScannerException rather than
      // swallowing it, because the right answer differs per call site. The sheet
      // has to catch it: uncaught, it escaped an async onPressed with no user
      // feedback at all, so the scan icon just appeared to do nothing.
      //
      // Deliberately the message and not fallBackToPhotos -- that opens a whole
      // new Quick Scan flow, and this sheet is already open with a photo source
      // one tap away in its own app bar.
      final permCalls = installPermissionMocks(status: PermissionStatusWire.denied);
      addTearDown(clearPermissionMocks);

      await pumpSheet(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      await tester.tap(find.byIcon(Icons.add_a_photo));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull,
          reason: 'the permission failure must not escape the callback');
      // findsWidgets, not findsOneWidget: ScaffoldMessenger renders the snack
      // bar into every registered Scaffold, and while the sheet is open that is
      // both the sheet's and the route's underneath. `requests` is the honest
      // proof it was REPORTED once -- the scanner asks for camera permission
      // exactly once per invocation.
      expect(find.text(cameraDeniedFallbackMessage), findsWidgets);
      expect(permCalls.requests, 1, reason: 'one tap, one attempt');
    });

    testWidgets('routes each entry to its own picker', (tester) async {
      final picker = _RecordingImagePicker();
      final selector = _RecordingFileSelector();
      ImagePickerPlatform.instance = picker;
      FileSelectorPlatform.instance = selector;

      await pumpSheet(tester,
          aiEnabled: true, permissions: [quickScan, create]);

      await tester.tap(find.byKey(menuKey));
      await tester.pumpAndSettle();
      await tester.tap(find.text(uploadPhotoLabel));
      await tester.pumpAndSettle();

      expect(picker.pickCalls, 1);
      expect(selector.openFilesCalls, 0);

      await tester.tap(find.byKey(menuKey));
      await tester.pumpAndSettle();
      await tester.tap(find.text(uploadFileLabel));
      await tester.pumpAndSettle();

      expect(selector.openFilesCalls, 1,
          reason: 'the file entry must reach the document picker, which is '
              'the only source that can produce a PDF');
      expect(picker.pickCalls, 1);
    });

    // The submitted state disables this menu the same way it disables the
    // scanner icon (`enabled: !isCompleted`), but the sheet owns
    // `isCompletedSubject` internally and only a real submit flips it, so that
    // branch is not reachable from a widget test. It is covered on-device by
    // the queued-confirmation e2e.
  });

  testWidgets('opens for a user who holds the quick-scan permission',
      (tester) async {
    await pumpSheet(tester, aiEnabled: true, permissions: [quickScan, create]);

    expect(find.text(quickScanLabel), findsOneWidget);
    expect(find.text('Scan or upload an image to get started'), findsOneWidget);
  });

  testWidgets('refuses to open without the permission, naming the group',
      (tester) async {
    await pumpSheet(tester, aiEnabled: true, permissions: [create]);

    expect(find.text('Scan or upload an image to get started'), findsNothing);
    expect(find.text(quickScanNoPermissionMessageForGroup('Household')),
        findsOneWidget);
  });

  testWidgets('refuses to open with the ai flag off, naming that reason',
      (tester) async {
    await pumpSheet(tester,
        aiEnabled: false, permissions: [quickScan, create]);

    expect(find.text('Scan or upload an image to get started'), findsNothing);
    expect(find.text(quickScanAiDisabledMessage), findsOneWidget);
  });

  testWidgets('offers manual entry to a user who can create receipts',
      (tester) async {
    await pumpSheet(tester, aiEnabled: true, permissions: [quickScan, create]);

    expect(find.byKey(manualLinkKey), findsOneWidget);
    expect(find.text(enterDetailsManuallyLabel), findsOneWidget);
  });

  testWidgets('hides the manual link without the create permission',
      (tester) async {
    await pumpSheet(tester, aiEnabled: true, permissions: [quickScan]);

    expect(find.byKey(manualLinkKey), findsNothing,
        reason: 'the manual form would only reject their save');
  });

  testWidgets('nothing confirms a queued scan until one is submitted',
      (tester) async {
    await pumpSheet(tester, aiEnabled: true, permissions: [quickScan, create]);

    expect(find.byKey(const ValueKey('quick-scan-queued-confirmation')),
        findsNothing);
    expect(find.text(quickScanQueuedMessage), findsNothing);
  });

  testWidgets('a single-page scan gets no page counter and no arrows',
      (tester) async {
    await pumpCarousel(tester, pageCount: 1);

    expect(find.byKey(pageIndicatorKey), findsNothing,
        reason: 'nothing to page to');
    expect(find.byKey(previousPageKey), findsNothing);
    expect(find.byKey(nextPageKey), findsNothing);
  });

  testWidgets('a multi-page scan counts its pages', (tester) async {
    await pumpCarousel(tester, pageCount: 3);

    expect(find.byKey(pageIndicatorKey), findsOneWidget);
    expect(find.text('1 of 3'), findsOneWidget,
        reason: 'a scan carries up to 100 pages and the carousel gives no '
            'other hint the later ones exist');
  });

  testWidgets('the arrows only ever point at a page that exists',
      (tester) async {
    final controller = await pumpCarousel(tester, pageCount: 3);

    expect(find.byKey(previousPageKey), findsNothing,
        reason: 'the first page has nothing behind it');
    expect(find.byKey(nextPageKey), findsOneWidget);

    await tester.tap(find.byKey(nextPageKey));
    await tester.pumpAndSettle();

    expect(controller.selectedItem, 1);
    expect(find.text('2 of 3'), findsOneWidget);
    expect(find.byKey(previousPageKey), findsOneWidget,
        reason: 'a middle page can go either way');
    expect(find.byKey(nextPageKey), findsOneWidget);

    await tester.tap(find.byKey(nextPageKey));
    await tester.pumpAndSettle();

    expect(controller.selectedItem, 2);
    expect(find.text('3 of 3'), findsOneWidget);
    expect(find.byKey(nextPageKey), findsNothing,
        reason: 'the last page has nothing ahead of it');

    await tester.tap(find.byKey(previousPageKey));
    await tester.pumpAndSettle();

    expect(controller.selectedItem, 1,
        reason: 'the arrows are the swipe, not a separate notion of where the '
            'user is');
    expect(find.text('2 of 3'), findsOneWidget);
  });

  testWidgets('the manual link closes the sheet and opens the form',
      (tester) async {
    await pumpSheet(tester, aiEnabled: true, permissions: [quickScan, create]);

    await tester.tap(find.byKey(manualLinkKey));
    await tester.pumpAndSettle();

    expect(find.text('Add screen'), findsOneWidget);
    expect(find.text(quickScanLabel), findsNothing,
        reason: 'leaving the modal above the form would hide the screen the '
            'user just asked for');
  });
}
