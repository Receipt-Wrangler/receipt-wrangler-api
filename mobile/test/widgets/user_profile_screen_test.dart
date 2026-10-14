import 'dart:async';

import 'package:built_collection/built_collection.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:mocktail/mocktail.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/persistence/global_shared_preferences.dart';
import 'package:receipt_wrangler_mobile/profile/screens/user_profile_screen.dart';
import 'package:receipt_wrangler_mobile/services/oidc_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../helpers/auth_test_helpers.dart';

class MockOidcApi extends Mock implements api.OidcApi {}

typedef ConfirmDelete = Future<String> Function({required String providerName});

// An account an identity provider created has no password of its own, so the
// password dialog is a dead end for it: it confirms a self-delete by signing in
// at one of its providers again, and sends the code that comes back instead.
// Every other account keeps the password dialog exactly as before.
void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late MockOpenapi mockClient;
  late MockOidcApi mockOidcApi;
  late MockUserApi mockUserApi;
  late MockAuthModel authModel;

  setUpAll(() {
    registerFallbackValue(api.DeleteAccountCommand());
  });

  setUp(() async {
    // CrashReportingToggle reads GlobalSharedPreferences during build.
    SharedPreferences.setMockInitialValues({});
    await GlobalSharedPreferences.initialize();

    mockClient = MockOpenapi();
    mockOidcApi = MockOidcApi();
    mockUserApi = MockUserApi();
    when(() => mockClient.getOidcApi()).thenReturn(mockOidcApi);
    when(() => mockClient.getUserApi()).thenReturn(mockUserApi);
    // The Connected accounts section loads its own list; empty keeps it hidden.
    when(() => mockOidcApi.getOidcConnections()).thenAnswer(
      (_) async => Response<BuiltList<api.OidcConnectionView>>(
        requestOptions: RequestOptions(path: '/oidc/connections'),
        data: BuiltList<api.OidcConnectionView>(),
      ),
    );
    when(() => mockUserApi.deleteAccount(
          deleteAccountCommand: any(named: 'deleteAccountCommand'),
        )).thenAnswer(
      (_) async => Response<void>(
        requestOptions: RequestOptions(path: '/user/deleteAccount'),
        statusCode: 200,
      ),
    );
    OpenApiClient.client = mockClient;

    // A mock rather than the real model: purgeTokens would reach the secure
    // storage plugin and rebuild OpenApiClient.client.
    authModel = MockAuthModel();
    when(() => authModel.featureConfig).thenReturn(
      (api.FeatureConfigBuilder()
            ..aiPoweredReceipts = false
            ..enableLocalSignUp = false)
          .build(),
    );
    when(() => authModel.purgeTokens()).thenAnswer((_) async {});
  });

  api.OidcConnectionView connection({required bool provisionedUser}) =>
      (api.OidcConnectionViewBuilder()
            ..providerName = 'google'
            ..providerDisplayName = 'Google'
            ..provisionedUser = provisionedUser
            ..linkedAt = DateTime.utc(2026, 1, 1))
          .build();

  Future<void> pumpScreen(
    WidgetTester tester, {
    required Future<List<api.OidcConnectionView>> Function() loadConnections,
    ConfirmDelete? confirmDelete,
  }) async {
    final router = GoRouter(
      routes: [
        GoRoute(
          path: '/',
          builder: (context, state) => UserProfileScreen(
            loadConnections: loadConnections,
            confirmDelete: confirmDelete ??
                ({required providerName}) =>
                    throw StateError('confirmDelete was not expected'),
          ),
        ),
        GoRoute(
          path: '/login',
          builder: (context, state) =>
              const Scaffold(body: Text('login screen')),
        ),
      ],
    );
    addTearDown(router.dispose);

    await tester.pumpWidget(
      MultiProvider(
        providers: [
          ChangeNotifierProvider<AuthModel>.value(value: authModel),
          ChangeNotifierProvider<LoadingModel>(create: (_) => LoadingModel()),
        ],
        child: MaterialApp.router(routerConfig: router),
      ),
    );
    await tester.pumpAndSettle();
  }

  Finder deleteButton() =>
      find.widgetWithText(ElevatedButton, 'Delete Account').first;

  Future<void> tapDelete(WidgetTester tester) async {
    await tester.ensureVisible(deleteButton());
    await tester.tap(deleteButton());
    await tester.pumpAndSettle();
  }

  api.DeleteAccountCommand sentCommand() {
    return verify(() => mockUserApi.deleteAccount(
          deleteAccountCommand: captureAny(named: 'deleteAccountCommand'),
        )).captured.single as api.DeleteAccountCommand;
  }

  testWidgets(
      'a passwordless account confirms at its provider and sends the code',
      (tester) async {
    final confirmedWith = <String>[];

    await pumpScreen(
      tester,
      loadConnections: () async => [connection(provisionedUser: true)],
      confirmDelete: ({required providerName}) async {
        confirmedWith.add(providerName);
        return 'the-confirmation-code';
      },
    );

    await tapDelete(tester);

    // The provider dialog, not the password one.
    expect(find.byKey(const ValueKey('password-field')), findsNothing);
    await tester.tap(find.byKey(const ValueKey('delete-confirm-google')));
    await tester.pumpAndSettle();

    expect(confirmedWith, ['google']);
    final command = sentCommand();
    expect(command.confirmationCode, 'the-confirmation-code');
    expect(command.password, isNull);
    verify(() => authModel.purgeTokens()).called(1);
    expect(find.text('login screen'), findsOneWidget);
  });

  testWidgets('an account with a password still confirms with it',
      (tester) async {
    await pumpScreen(
      tester,
      loadConnections: () async => [connection(provisionedUser: false)],
    );

    await tapDelete(tester);

    expect(find.byKey(const ValueKey('delete-confirm-google')), findsNothing);
    await tester.enterText(
        find.byKey(const ValueKey('password-field')), 'MySecretPass123');
    await tester.pump();
    await tester.tap(find.widgetWithText(ElevatedButton, 'Delete Account').last);
    await tester.pumpAndSettle();

    final command = sentCommand();
    expect(command.password, 'MySecretPass123');
    expect(command.confirmationCode, isNull);
    expect(find.text('login screen'), findsOneWidget);
  });

  testWidgets('a failed connections load falls back to the password dialog',
      (tester) async {
    await pumpScreen(
      tester,
      loadConnections: () async => throw Exception('older server'),
    );

    await tapDelete(tester);

    expect(find.byKey(const ValueKey('password-field')), findsOneWidget);
  });

  testWidgets('cancelling the provider dialog makes no API call',
      (tester) async {
    var confirmations = 0;

    await pumpScreen(
      tester,
      loadConnections: () async => [connection(provisionedUser: true)],
      confirmDelete: ({required providerName}) async {
        confirmations += 1;
        return 'unused';
      },
    );

    await tapDelete(tester);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();

    expect(confirmations, 0);
    verifyNever(() => mockUserApi.deleteAccount(
          deleteAccountCommand: any(named: 'deleteAccountCommand'),
        ));
    expect(find.text('login screen'), findsNothing);
  });

  testWidgets('a refused confirmation shows its message and deletes nothing',
      (tester) async {
    await pumpScreen(
      tester,
      loadConnections: () async => [connection(provisionedUser: true)],
      confirmDelete: ({required providerName}) async =>
          throw OidcSignInException(oidcErrorMessage('identity_mismatch')),
    );

    await tapDelete(tester);
    await tester.tap(find.byKey(const ValueKey('delete-confirm-google')));
    await tester.pumpAndSettle();

    expect(find.text(oidcErrorMessage('identity_mismatch')), findsOneWidget);
    verifyNever(() => mockUserApi.deleteAccount(
          deleteAccountCommand: any(named: 'deleteAccountCommand'),
        ));
    expect(find.text('login screen'), findsNothing);
  });

  testWidgets('a second tap does not start a second delete flow',
      (tester) async {
    var started = 0;
    final neverCompletes = Completer<List<api.OidcConnectionView>>();

    await pumpScreen(
      tester,
      loadConnections: () {
        started += 1;
        return neverCompletes.future;
      },
    );

    await tester.ensureVisible(deleteButton());
    // Both taps land before any frame is pumped -- the window a disabled
    // button alone cannot close, since onPressed only goes null on the frame
    // after setState.
    await tester.tap(deleteButton(), warnIfMissed: false);
    await tester.tap(deleteButton(), warnIfMissed: false);
    await tester.pump();

    expect(started, 1);
    expect(tester.widget<ElevatedButton>(deleteButton()).onPressed, isNull);
  });
}
