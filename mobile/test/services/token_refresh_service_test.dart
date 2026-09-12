import 'package:built_collection/built_collection.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:openapi/openapi.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/services/token_refresh_service.dart';

import '../helpers/auth_test_helpers.dart';

// --- Test-specific mocks (not shared) ---

class MockAppData extends Mock implements AppData {}

class MockClaims extends Mock implements Claims {}

class MockFeatureConfig extends Mock implements FeatureConfig {}

class MockUserPreferences extends Mock implements UserPreferences {}

/// A fully stubbed `AppData` for the paths where `getAppData` succeeds.
/// `storeAppData` reads every one of these fields, so a mock missing any would
/// throw mid-store rather than fail the assertion under test. [jwt] and
/// [refreshToken] are parameters because whether AppData carries a token pair is
/// itself under test.
MockAppData buildMockAppData({String? jwt = '', String? refreshToken = ''}) {
  final appData = MockAppData();
  when(() => appData.jwt).thenReturn(jwt);
  when(() => appData.refreshToken).thenReturn(refreshToken);
  when(() => appData.claims).thenReturn(MockClaims());
  when(() => appData.featureConfig).thenReturn(MockFeatureConfig());
  when(() => appData.groups).thenReturn(BuiltList<Group>());
  when(() => appData.users).thenReturn(BuiltList<UserView>());
  when(() => appData.userPreferences).thenReturn(MockUserPreferences());
  when(() => appData.categories).thenReturn(BuiltList<Category>());
  when(() => appData.tags).thenReturn(BuiltList<Tag>());
  when(() => appData.currencyDisplay).thenReturn('');
  when(() => appData.currencyDecimalSeparator)
      .thenReturn(CurrencySeparator.period);
  when(() => appData.currencyThousandthsSeparator)
      .thenReturn(CurrencySeparator.comma);
  when(() => appData.currencySymbolPosition)
      .thenReturn(CurrencySymbolPosition.END);
  when(() => appData.currencyHideDecimalPlaces).thenReturn(false);
  when(() => appData.appPermissions).thenReturn(BuiltList<String>());
  when(() => appData.groupPermissions)
      .thenReturn(BuiltMap<String, BuiltList<String>>());
  return appData;
}

void main() {
  late MockAuthModel mockAuthModel;
  late MockGroupModel mockGroupModel;
  late MockUserModel mockUserModel;
  late MockUserPreferencesModel mockUserPreferencesModel;
  late MockCategoryModel mockCategoryModel;
  late MockTagModel mockTagModel;
  late MockSystemSettingsModel mockSystemSettingsModel;
  // PermissionsModel has no I/O, so use a real instance (per mobile testing guide).
  late PermissionsModel permissionsModel;
  late MockOpenapi mockClient;
  late MockAuthApi mockAuthApi;
  late MockUserApi mockUserApi;
  late TokenRefreshService service;

  setUpAll(() {
    registerFallbackValue(FakeLogoutCommand());
    registerFallbackValue(MockClaims());
    registerFallbackValue(MockFeatureConfig());
    registerFallbackValue(MockUserPreferences());
    registerFallbackValue(<Group>[]);
    registerFallbackValue(<UserView>[]);
    registerFallbackValue(<Category>[]);
    registerFallbackValue(<Tag>[]);
    registerFallbackValue('');
    registerFallbackValue(CurrencySeparator.period);
    registerFallbackValue(CurrencySymbolPosition.END);
    registerFallbackValue(false);
  });

  setUp(() {
    mockAuthModel = MockAuthModel();
    mockGroupModel = MockGroupModel();
    mockUserModel = MockUserModel();
    mockUserPreferencesModel = MockUserPreferencesModel();
    mockCategoryModel = MockCategoryModel();
    mockTagModel = MockTagModel();
    mockSystemSettingsModel = MockSystemSettingsModel();
    permissionsModel = PermissionsModel();
    mockClient = MockOpenapi();
    mockAuthApi = MockAuthApi();
    mockUserApi = MockUserApi();

    when(() => mockClient.getAuthApi()).thenReturn(mockAuthApi);
    when(() => mockClient.getUserApi()).thenReturn(mockUserApi);

    OpenApiClient.client = mockClient;

    // Reset and re-initialize the singleton for each test.
    service = TokenRefreshService();
    service.resetForTesting();
    service.initialize(
      authModel: mockAuthModel,
      groupModel: mockGroupModel,
      userModel: mockUserModel,
      userPreferencesModel: mockUserPreferencesModel,
      categoryModel: mockCategoryModel,
      tagModel: mockTagModel,
      systemSettingsModel: mockSystemSettingsModel,
      permissionsModel: permissionsModel,
    );
  });

  group('TokenRefreshService', () {
    group('refreshTokens with force=false', () {
      test('returns true when JWT is still valid', () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);

        final result = await service.refreshTokens();

        expect(result, true);
        verifyNever(() => mockAuthApi.getNewRefreshToken(
            logoutCommand: any(named: 'logoutCommand')));
      });

      test('refreshes token when JWT is expired but refresh token is valid',
          () async {
        final newJwt = validJwt;
        final newRefresh = validJwt;

        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);

        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenAnswer(
                (_) async => createTokenRefreshResponse(newJwt, newRefresh));
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});

        final result = await service.refreshTokens();

        expect(result, true);
        verify(() => mockAuthModel.setTokens(newJwt, newRefresh)).called(1);
      });

      test('purges tokens when both JWT and refresh token are expired',
          () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});

        final result = await service.refreshTokens();

        expect(result, false);
        verify(() => mockAuthModel.purgeTokens()).called(1);
      });

      test('treats corrupted JWT as invalid and attempts refresh', () async {
        final newJwt = validJwt;
        final newRefresh = validJwt;

        when(() => mockAuthModel.getJwt())
            .thenAnswer((_) async => 'not-a-valid-jwt');
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);

        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenAnswer(
                (_) async => createTokenRefreshResponse(newJwt, newRefresh));
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});

        final result = await service.refreshTokens();

        expect(result, true);
        verify(() => mockAuthApi.getNewRefreshToken(
            logoutCommand: any(named: 'logoutCommand'))).called(1);
      });

      test('purges tokens when both JWT and refresh token are corrupted',
          () async {
        when(() => mockAuthModel.getJwt())
            .thenAnswer((_) async => 'corrupted-jwt');
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => 'corrupted-refresh');
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});

        final result = await service.refreshTokens();

        expect(result, false);
        verify(() => mockAuthModel.purgeTokens()).called(1);
      });

      test('purges tokens when JWT is null', () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => null);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => null);
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});

        final result = await service.refreshTokens();

        expect(result, false);
        verify(() => mockAuthModel.purgeTokens()).called(1);
      });

      test('purges tokens when refresh endpoint throws', () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});
        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenThrow(DioException(
          requestOptions: RequestOptions(path: '/token/'),
          response: Response(
            statusCode: 500,
            requestOptions: RequestOptions(path: '/token/'),
          ),
        ));

        final result = await service.refreshTokens();

        expect(result, false);
        verify(() => mockAuthModel.purgeTokens()).called(1);
      });
    });

    group('refreshTokens with force=true', () {
      test('refreshes even when JWT is still valid', () async {
        final newJwt = validJwt;
        final newRefresh = validJwt;

        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);

        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenAnswer(
                (_) async => createTokenRefreshResponse(newJwt, newRefresh));
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});

        final result = await service.refreshTokens(force: true);

        expect(result, true);
        verify(() => mockAuthApi.getNewRefreshToken(
            logoutCommand: any(named: 'logoutCommand'))).called(1);
      });

      test('returns false when force=true but refresh token is expired',
          () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});

        final result = await service.refreshTokens(force: true);

        expect(result, false);
        verify(() => mockAuthModel.purgeTokens()).called(1);
        verifyNever(() => mockAuthApi.getNewRefreshToken(
            logoutCommand: any(named: 'logoutCommand')));
      });
    });

    group('serialization - concurrent calls share one Future', () {
      test('concurrent calls return the same result without duplicate requests',
          () async {
        var callCount = 0;
        final newJwt = validJwt;
        final newRefresh = validJwt;

        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);

        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenAnswer((_) async {
          callCount++;
          await Future.delayed(const Duration(milliseconds: 50));
          return createTokenRefreshResponse(newJwt, newRefresh);
        });
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});

        final results = await Future.wait([
          service.refreshTokens(),
          service.refreshTokens(),
          service.refreshTokens(),
          service.refreshTokens(),
          service.refreshTokens(),
        ]);

        expect(results, everyElement(true));
        expect(callCount, 1);
      });

      test('after completion, a new call makes a new request', () async {
        var callCount = 0;
        final newJwt = validJwt;
        final newRefresh = validJwt;

        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);

        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenAnswer((_) async {
          callCount++;
          return createTokenRefreshResponse(newJwt, newRefresh);
        });
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});

        await service.refreshTokens();
        expect(callCount, 1);

        await service.refreshTokens();
        expect(callCount, 2);
      });

      test('concurrent calls all get false when refresh fails', () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});
        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenAnswer((_) async {
          await Future.delayed(const Duration(milliseconds: 50));
          throw DioException(
            requestOptions: RequestOptions(path: '/token/'),
            response: Response(
              statusCode: 500,
              requestOptions: RequestOptions(path: '/token/'),
            ),
          );
        });

        final results = await Future.wait([
          service.refreshTokens(),
          service.refreshTokens(),
          service.refreshTokens(),
        ]);

        expect(results, everyElement(false));
      });
    });

    group('app data loading', () {
      test('loads app data when groups are empty and user is authenticated',
          () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([]);

        final mockAppData = buildMockAppData();

        when(() => mockUserApi.getAppData()).thenAnswer((_) async => Response(
              data: mockAppData,
              requestOptions: RequestOptions(path: '/user/appData'),
              statusCode: 200,
            ));
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});
        when(() => mockAuthModel.setClaims(any())).thenReturn(null);
        when(() => mockAuthModel.setFeatureConfig(any())).thenReturn(null);
        when(() => mockGroupModel.setGroups(any())).thenReturn(null);
        when(() => mockUserModel.setUsers(any())).thenReturn(null);
        when(() => mockUserPreferencesModel.setUserPreferences(any()))
            .thenReturn(null);
        when(() => mockCategoryModel.setCategories(any())).thenReturn(null);
        when(() => mockTagModel.setTags(any())).thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyDisplay(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyDecimalSeparator(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyThousandSeparator(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencySymbolPosition(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyHideDecimalPlaces(any()))
            .thenReturn(null);

        final result = await service.refreshTokens();

        expect(result, true);
        verify(() => mockUserApi.getAppData()).called(1);
      });

      test(
          'stores all app data fields when jwt and refreshToken are null in response',
          () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([]);

        final mockAppData = buildMockAppData(jwt: null, refreshToken: null);

        when(() => mockUserApi.getAppData()).thenAnswer((_) async => Response(
              data: mockAppData,
              requestOptions: RequestOptions(path: '/user/appData'),
              statusCode: 200,
            ));
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});
        when(() => mockAuthModel.setClaims(any())).thenReturn(null);
        when(() => mockAuthModel.setFeatureConfig(any())).thenReturn(null);
        when(() => mockGroupModel.setGroups(any())).thenReturn(null);
        when(() => mockUserModel.setUsers(any())).thenReturn(null);
        when(() => mockUserPreferencesModel.setUserPreferences(any()))
            .thenReturn(null);
        when(() => mockCategoryModel.setCategories(any())).thenReturn(null);
        when(() => mockTagModel.setTags(any())).thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyDisplay(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyDecimalSeparator(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyThousandSeparator(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencySymbolPosition(any()))
            .thenReturn(null);
        when(() => mockSystemSettingsModel.setCurrencyHideDecimalPlaces(any()))
            .thenReturn(null);

        final result = await service.refreshTokens();

        expect(result, true);
        // setTokens should NOT be called when both tokens are null
        verifyNever(() => mockAuthModel.setTokens(any(), any()));
        // But all other app data fields should still be stored
        verify(() => mockAuthModel.setClaims(any())).called(1);
        verify(() => mockGroupModel.setGroups(any())).called(1);
        verify(() => mockUserModel.setUsers(any())).called(1);
      });

      test(
          'returns true when app data loading fails after successful token refresh',
          () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => expiredJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([]);
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});

        final newJwt = validJwt;
        final newRefresh = validJwt;

        when(() => mockAuthApi.getNewRefreshToken(
                logoutCommand: any(named: 'logoutCommand')))
            .thenAnswer(
                (_) async => createTokenRefreshResponse(newJwt, newRefresh));
        when(() => mockAuthModel.setTokens(any(), any()))
            .thenAnswer((_) async {});

        // App data loading throws
        when(() => mockUserApi.getAppData()).thenThrow(DioException(
          requestOptions: RequestOptions(path: '/user/appData'),
          response: Response(
            statusCode: 500,
            requestOptions: RequestOptions(path: '/user/appData'),
          ),
        ));

        final result = await service.refreshTokens();

        // Should still return true because tokens were refreshed successfully
        expect(result, true);
        // Should NOT purge the freshly-obtained tokens
        verifyNever(() => mockAuthModel.purgeTokens());
      });

      test('returns true when app data loading fails with valid JWT',
          () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([]);
        when(() => mockAuthModel.purgeTokens()).thenAnswer((_) async {});

        when(() => mockUserApi.getAppData()).thenThrow(DioException(
          requestOptions: RequestOptions(path: '/user/appData'),
          response: Response(
            statusCode: 500,
            requestOptions: RequestOptions(path: '/user/appData'),
          ),
        ));

        final result = await service.refreshTokens();

        expect(result, true);
        verifyNever(() => mockAuthModel.purgeTokens());
      });

      // AppData used to be fetched only when GroupModel was EMPTY, which is
      // false for the whole of a logged-in session -- so the 15-minute timer and
      // the on-resume refresh both became no-ops, and a group setting changed
      // elsewhere could not reach a running app until the user logged out and
      // back in. These two pin the replacement: refresh even when groups are
      // already loaded, but collapse a burst.
      test('reloads app data even when groups already exist', () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);
        when(() => mockUserApi.getAppData()).thenAnswer((_) async => Response(
              data: buildMockAppData(),
              requestOptions: RequestOptions(path: '/user/appData'),
              statusCode: 200,
            ));

        final result = await service.refreshTokens();

        expect(result, true);
        verify(() => mockUserApi.getAppData()).called(1);
      });

      // Quick Scan calls reloadAppData() before opening the scanner and needs a
      // guarantee, not a best effort -- if the debounce silently swallowed it,
      // a config change made seconds earlier would still not reach the form.
      test('reloadAppData bypasses the debounce', () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);
        when(() => mockUserApi.getAppData()).thenAnswer((_) async => Response(
              data: buildMockAppData(),
              requestOptions: RequestOptions(path: '/user/appData'),
              statusCode: 200,
            ));

        // The debounced path would skip this second load; reloadAppData must not.
        await service.refreshTokens();
        await service.reloadAppData();

        verify(() => mockUserApi.getAppData()).called(2);
      });

      // A load is a fetch AND a storeAppData. The scheduled path and the forced
      // path are independent, so unserialized they overlap and whichever
      // response lands last wins -- an older scheduled payload could overwrite
      // the config a Quick Scan just fetched, which is the staleness this whole
      // path exists to prevent.
      test('serializes a forced reload against an in-flight scheduled load',
          () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);

        var inFlight = 0;
        var peakInFlight = 0;
        when(() => mockUserApi.getAppData()).thenAnswer((_) async {
          inFlight++;
          peakInFlight = inFlight > peakInFlight ? inFlight : peakInFlight;
          // Long enough that an unserialized second load would start inside it.
          await Future<void>.delayed(const Duration(milliseconds: 50));
          inFlight--;
          return Response(
            data: buildMockAppData(),
            requestOptions: RequestOptions(path: '/user/appData'),
            statusCode: 200,
          );
        });

        // refreshTokens awaits the token checks before it ever reaches the
        // load, so give it long enough to actually be mid-fetch -- otherwise the
        // forced reload wins the race trivially and nothing is serialized.
        final scheduled = service.refreshTokens();
        await Future<void>.delayed(const Duration(milliseconds: 10));
        final forced = service.reloadAppData();
        await Future.wait([scheduled, forced]);

        expect(peakInFlight, 1, reason: 'the two loads never overlapped');
        // Still two distinct fetches -- serializing must not collapse the forced
        // reload into the scheduled one the way _refreshCompleter would.
        verify(() => mockUserApi.getAppData()).called(2);
      });

      test('reloadAppData swallows a failed fetch', () async {
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);
        when(() => mockUserApi.getAppData()).thenThrow(DioException(
          requestOptions: RequestOptions(path: '/user/appData'),
          response: Response(
            statusCode: 500,
            requestOptions: RequestOptions(path: '/user/appData'),
          ),
        ));

        // It sits in front of the scanner, so a transient failure must fall
        // through to the data already loaded rather than block the scan.
        await expectLater(service.reloadAppData(), completes);
      });

      test('debounces a second refresh that follows immediately', () async {
        when(() => mockAuthModel.getJwt()).thenAnswer((_) async => validJwt);
        when(() => mockAuthModel.getRefreshToken())
            .thenAnswer((_) async => validJwt);
        when(() => mockGroupModel.groups).thenReturn([MockGroup()]);
        when(() => mockUserApi.getAppData()).thenAnswer((_) async => Response(
              data: buildMockAppData(),
              requestOptions: RequestOptions(path: '/user/appData'),
              statusCode: 200,
            ));

        // Resume and the periodic timer can land together; the second call must
        // not repeat the request.
        await service.refreshTokens();
        await service.refreshTokens();

        verify(() => mockUserApi.getAppData()).called(1);
      });
    });
  });
}
