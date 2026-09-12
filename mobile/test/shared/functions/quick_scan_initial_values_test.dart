import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/interfaces/upload_multipart_file_data.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/models/user_preferences_model.dart';
import 'package:receipt_wrangler_mobile/shared/functions/quick_scan.dart';

import '../../helpers/receipt_entry_test_helpers.dart';
import '../../helpers/receipt_form_test_helpers.dart';

// Every Quick Scan entry point funnels through buildQuickScanImages, so this is
// where a freshly picked image gets its group. Two rules stack: the user's own
// quick-scan default group wins, and failing that a member of exactly one group
// has no choice to make.
//
// The 0-vs-null distinction is load-bearing: the generated Dart model DEFAULTS
// quickScanDefaultGroupId to 0, so "unset" never arrives as null.

const _groupOneId = 1;
const _groupTwoId = 2;
const _preferredGroupId = 42;

api.UserPreferences _preferences({int? quickScanDefaultGroupId}) =>
    (api.UserPreferencesBuilder()
          ..id = 0
          ..userId = 0
          ..createdAt = ''
          ..quickScanDefaultGroupId = quickScanDefaultGroupId ?? 0)
        .build();

UploadMultipartFileData _uploadedImage() => UploadMultipartFileData(
      multipartFile: MultipartFile.fromBytes(quickScanTestPngBytes),
      bytes: quickScanTestPngBytes,
    );

/// Seeds the two models [buildQuickScanImages] reads and returns the group id
/// it stamps on a freshly picked image.
Future<int?> _seededGroupId(
  WidgetTester tester, {
  required List<api.Group> groups,
  int? quickScanDefaultGroupId,
}) async {
  late BuildContext capturedContext;

  await tester.pumpWidget(MultiProvider(
    providers: [
      ChangeNotifierProvider<GroupModel>.value(value: groupModelWith(groups)),
      ChangeNotifierProvider<UserPreferencesModel>.value(
        value: UserPreferencesModel()
          ..setUserPreferences(_preferences(
              quickScanDefaultGroupId: quickScanDefaultGroupId)),
      ),
    ],
    child: MaterialApp(
      home: Builder(builder: (context) {
        capturedContext = context;
        return const SizedBox.shrink();
      }),
    ),
  ));

  return buildQuickScanImages(capturedContext, [_uploadedImage()])
      .first
      .groupId;
}

void main() {
  testWidgets('seeds the only group when there is no default preference',
      (tester) async {
    expect(
      await _seededGroupId(tester,
          groups: [buildGroup(id: _groupOneId, name: 'My Receipts')]),
      _groupOneId,
    );
  });

  testWidgets('ignores the synthetic All group when counting', (tester) async {
    expect(
      await _seededGroupId(tester, groups: [
        buildGroup(id: 99, name: 'All Groups', isAllGroup: true),
        buildGroup(id: _groupOneId, name: 'My Receipts'),
      ]),
      _groupOneId,
    );
  });

  testWidgets('prefers the quick-scan default group over the only group',
      (tester) async {
    expect(
      await _seededGroupId(
        tester,
        groups: [buildGroup(id: _groupOneId, name: 'My Receipts')],
        quickScanDefaultGroupId: _preferredGroupId,
      ),
      _preferredGroupId,
    );
  });

  testWidgets('leaves the group unset when the user belongs to more than one',
      (tester) async {
    expect(
      await _seededGroupId(tester, groups: [
        buildGroup(id: _groupOneId),
        buildGroup(id: _groupTwoId),
      ]),
      isNull,
    );
  });

  testWidgets('leaves the group unset when the user has no selectable group',
      (tester) async {
    expect(
      await _seededGroupId(tester, groups: [
        buildGroup(id: 99, name: 'All Groups', isAllGroup: true),
      ]),
      isNull,
    );
  });
}
