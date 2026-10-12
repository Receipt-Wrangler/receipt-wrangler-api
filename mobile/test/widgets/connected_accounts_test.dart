import 'dart:async';

import 'package:built_collection/built_collection.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/persistence/global_shared_preferences.dart';
import 'package:receipt_wrangler_mobile/profile/widgets/connected_accounts.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../helpers/auth_test_helpers.dart';

class MockOidcApi extends Mock implements api.OidcApi {}

// Connecting a provider opens an EXTERNAL browser session, and this widget --
// unlike the login form, which swaps itself for a spinner -- keeps its buttons
// mounted and enabled for the whole of it. `_isLoading` here tracks the initial
// connection-list fetch and nothing else, so without a dedicated guard every
// extra tap starts another flow and burns another single-use launch handle on
// the server.
void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late MockOpenapi mockClient;
  late MockOidcApi mockOidcApi;

  setUp(() async {
    // AuthModel.basePath reads GlobalSharedPreferences during build.
    SharedPreferences.setMockInitialValues({});
    await GlobalSharedPreferences.initialize();

    mockClient = MockOpenapi();
    mockOidcApi = MockOidcApi();
    when(() => mockClient.getOidcApi()).thenReturn(mockOidcApi);
    when(() => mockOidcApi.getOidcConnections()).thenAnswer(
      (_) async => Response<BuiltList<api.OidcConnectionView>>(
        requestOptions: RequestOptions(path: '/oidc/connections'),
        data: BuiltList<api.OidcConnectionView>(),
      ),
    );
    OpenApiClient.client = mockClient;
  });

  /// Pumps the section with one unconnected provider, so exactly one Connect
  /// button renders.
  Future<void> pumpSection(
    WidgetTester tester, {
    required Future<String> Function({required String providerName}) link,
  }) async {
    // Written through the live instance: AuthModel.setBasePath would also
    // rebuild OpenApiClient.client and throw away the mock set up above, and
    // re-running GlobalSharedPreferences.initialize() no-ops once its static
    // instance is set.
    await GlobalSharedPreferences.instance
        .setString('basePath', 'https://receipts.example.com/api');

    final authModel = AuthModel();
    authModel.setFeatureConfig(
      (api.FeatureConfigBuilder()
            ..aiPoweredReceipts = false
            ..enableLocalSignUp = false
            ..oidcProviders.add((api.OidcProviderSummaryBuilder()
                  ..name = 'google'
                  ..displayName = 'Google')
                .build()))
          .build(),
    );

    await tester.pumpWidget(
      ChangeNotifierProvider<AuthModel>.value(
        value: authModel,
        child: MaterialApp(
          home: Scaffold(body: ConnectedAccounts(link: link)),
        ),
      ),
    );
    // Let the initial getOidcConnections settle so the section renders.
    await tester.pump();
    await tester.pump();
  }

  Finder connectButton() => find.byKey(const ValueKey('oidc-connect-google'));

  testWidgets('a second tap does not start a second link flow',
      (tester) async {
    var started = 0;
    final neverCompletes = Completer<String>();

    await pumpSection(tester, link: ({required providerName}) {
      started += 1;
      return neverCompletes.future;
    });

    expect(connectButton(), findsOneWidget);

    // Both taps land before any frame is pumped -- the window a disabled
    // button alone cannot close, since onPressed only goes null on the frame
    // after setState.
    await tester.tap(connectButton(), warnIfMissed: false);
    await tester.tap(connectButton(), warnIfMissed: false);
    await tester.pump();

    expect(started, 1);
  });

  testWidgets('the Connect button is disabled while a link is open',
      (tester) async {
    final neverCompletes = Completer<String>();

    await pumpSection(
      tester,
      link: ({required providerName}) => neverCompletes.future,
    );

    expect(tester.widget<TextButton>(connectButton()).onPressed, isNotNull);

    await tester.tap(connectButton());
    await tester.pump();

    expect(tester.widget<TextButton>(connectButton()).onPressed, isNull);
  });

  testWidgets('a failed link re-enables the button', (tester) async {
    var started = 0;

    await pumpSection(tester, link: ({required providerName}) async {
      started += 1;
      throw Exception('nope');
    });

    await tester.tap(connectButton());
    await tester.pump();
    await tester.pump();

    // The finally clears the flag on every exit, not just the happy one --
    // otherwise one transient failure would leave the section permanently
    // unable to connect anything.
    expect(tester.widget<TextButton>(connectButton()).onPressed, isNotNull);

    await tester.tap(connectButton());
    await tester.pump();

    expect(started, 2);
  });
}
