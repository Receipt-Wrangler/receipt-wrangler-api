import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/openapi.dart'
    show
        Comment,
        FileDataView,
        FileDataViewBuilder,
        Item,
        Openapi,
        Receipt,
        ReceiptImageApi;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/interfaces/upload_multipart_file_data.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/models/receipt_model.dart';
import 'package:receipt_wrangler_mobile/shared/functions/receipt_upload.dart';

// uploadImagesToReceipt uploads one API call per image. A failure partway
// through leaves the earlier images already persisted server-side, so they have
// to reach imageBehaviorSubject even though the batch failed -- otherwise the
// user sees only an error, and repeating the action uploads them a SECOND time
// and the receipt ends up with duplicates.

class MockOpenapi extends Mock implements Openapi {}

class MockReceiptImageApi extends Mock implements ReceiptImageApi {}

/// Receipt has a dozen non-nullable fields and the code under test reads only
/// `id`; setReceipt additionally touches `comments` and `receiptItems`.
class MockReceipt extends Mock implements Receipt {}

FileDataView _fileData(int id) => (FileDataViewBuilder()
      ..id = id
      ..createdAt = ''
      ..encodedImage = ''
      ..name = 'image-$id.png')
    .build();

UploadMultipartFileData _image(String name) => UploadMultipartFileData(
      multipartFile: MultipartFile.fromBytes(const <int>[1], filename: name),
      bytes: Uint8List.fromList(const <int>[1]),
    );

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late MockOpenapi mockClient;
  late MockReceiptImageApi mockReceiptImageApi;
  late ReceiptModel receiptModel;

  setUpAll(() {
    // The upload stubs match on any(named: 'file'), and MultipartFile is not a
    // primitive, so mocktail needs a fallback instance to build the matcher.
    registerFallbackValue(MultipartFile.fromBytes(const <int>[]));
  });

  setUp(() {
    mockClient = MockOpenapi();
    mockReceiptImageApi = MockReceiptImageApi();
    when(() => mockClient.getReceiptImageApi()).thenReturn(mockReceiptImageApi);
    // The default client builds a real Openapi with a relative baseUrl that dio
    // rejects off-Web, so install the mock for the test's lifetime.
    OpenApiClient.client = mockClient;

    final receipt = MockReceipt();
    when(() => receipt.id).thenReturn(1);
    when(() => receipt.comments).thenReturn(BuiltList<Comment>());
    when(() => receipt.receiptItems).thenReturn(BuiltList<Item>());

    receiptModel = ReceiptModel();
    receiptModel.setReceipt(receipt, false);
  });

  Future<BuildContext> pumpHost(WidgetTester tester) async {
    late BuildContext captured;
    await tester.pumpWidget(
      ChangeNotifierProvider<LoadingModel>(
        create: (_) => LoadingModel(),
        child: MaterialApp(
          home: Scaffold(
            body: Builder(builder: (context) {
              captured = context;
              return const SizedBox.shrink();
            }),
          ),
        ),
      ),
    );
    await tester.pump();
    return captured;
  }

  Response<FileDataView> ok(int id) => Response<FileDataView>(
        requestOptions: RequestOptions(path: ''),
        statusCode: 200,
        data: _fileData(id),
      );

  testWidgets('publishes the images that uploaded before a mid-batch failure',
      (tester) async {
    var call = 0;
    when(() => mockReceiptImageApi.uploadReceiptImage(
        file: any(named: 'file'),
        receiptId: any(named: 'receiptId'))).thenAnswer((_) async {
      call++;
      if (call == 3) {
        throw DioException(requestOptions: RequestOptions(path: ''));
      }
      return ok(call);
    });

    final context = await pumpHost(tester);
    await uploadImagesToReceipt(context, receiptModel,
        [_image('a.png'), _image('b.png'), _image('c.png')]);
    await tester.pump();

    // The first two are on the server; dropping them here is what would cause a
    // retry to duplicate them.
    expect(
      receiptModel.imageBehaviorSubject.value.map((f) => f?.id).toList(),
      [1, 2],
      reason: 'images that succeeded must survive the failed batch',
    );
  });

  testWidgets('publishes nothing when the very first upload fails',
      (tester) async {
    when(() => mockReceiptImageApi.uploadReceiptImage(
            file: any(named: 'file'), receiptId: any(named: 'receiptId')))
        .thenThrow(DioException(requestOptions: RequestOptions(path: '')));

    final context = await pumpHost(tester);
    await uploadImagesToReceipt(context, receiptModel, [_image('a.png')]);
    await tester.pump();

    // Nothing uploaded, so nothing to publish -- the guard must not emit an
    // empty batch and rebuild every carousel listener for nothing.
    expect(receiptModel.imageBehaviorSubject.value, isEmpty);
  });

  testWidgets('publishes the whole batch exactly once on success',
      (tester) async {
    var call = 0;
    when(() => mockReceiptImageApi.uploadReceiptImage(
        file: any(named: 'file'),
        receiptId: any(named: 'receiptId'))).thenAnswer((_) async {
      call++;
      return ok(call);
    });

    var emissions = 0;
    final sub = receiptModel.imageBehaviorSubject.skip(1).listen((_) {
      emissions++;
    });
    addTearDown(sub.cancel);

    final context = await pumpHost(tester);
    await uploadImagesToReceipt(
        context, receiptModel, [_image('a.png'), _image('b.png')]);
    await tester.pump();

    expect(receiptModel.imageBehaviorSubject.value.map((f) => f?.id).toList(),
        [1, 2]);
    expect(emissions, 1,
        reason: 'moving the publish into finally must not double-emit');
  });
}
