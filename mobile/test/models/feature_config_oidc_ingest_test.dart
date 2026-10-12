import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart';

/// Guards the mobile side of the `featureConfig.oidcProviders` contract.
///
/// `GET /featureConfig` is fetched UNAUTHENTICATED by the Connect-to-Server
/// screen, before login, so a payload this client cannot deserialize does not
/// fail one screen -- it reports itself as "Failed to connect to server" and the
/// app is unusable. That is the same blast radius as the two documented login
/// outages, which is why the field is exercised in both directions here.
Map<String, Object?> _featureConfigJson({Object? oidcProviders = _absent}) => {
      'aiPoweredReceipts': false,
      'enableLocalSignUp': false,
      if (!identical(oidcProviders, _absent)) 'oidcProviders': oidcProviders,
    };

const Object _absent = Object();

void main() {
  group('FeatureConfig oidcProviders ingest', () {
    // An already-released build talking to a server that predates the field,
    // and a build talking to a server that has it but has none configured.
    test('deserializes a payload that omits oidcProviders entirely', () {
      final config = standardSerializers.deserializeWith(
        FeatureConfig.serializer,
        _featureConfigJson(),
      )!;

      expect(config.oidcProviders, isNull);
      expect(config.enableLocalSignUp, false);
    });

    test('deserializes an empty oidcProviders array', () {
      final config = standardSerializers.deserializeWith(
        FeatureConfig.serializer,
        _featureConfigJson(oidcProviders: <Object?>[]),
      )!;

      expect(config.oidcProviders, isNotNull);
      expect(config.oidcProviders!.length, 0);
    });

    test('deserializes configured providers', () {
      final config = standardSerializers.deserializeWith(
        FeatureConfig.serializer,
        _featureConfigJson(oidcProviders: [
          {'name': 'google', 'displayName': 'Google'},
          {'name': 'keycloak', 'displayName': 'Keycloak'},
        ]),
      )!;

      expect(config.oidcProviders!.length, 2);
      expect(config.oidcProviders!.first.name, 'google');
      expect(config.oidcProviders!.first.displayName, 'Google');
    });

    // The whole reason the summary carries only these two fields: this payload
    // is public, so it must not leak the issuer, client id, or anything else.
    test('the summary type exposes only a name and a display name', () {
      final summary = (OidcProviderSummaryBuilder()
            ..name = 'google'
            ..displayName = 'Google')
          .build();

      final serialized = standardSerializers.serializeWith(
        OidcProviderSummary.serializer,
        summary,
      ) as Map<Object?, Object?>;

      expect(serialized.keys.toSet(), {'name', 'displayName'});
    });
  });
}
