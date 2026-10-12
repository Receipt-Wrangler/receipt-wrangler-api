import 'package:test/test.dart';
import 'package:openapi/openapi.dart';

// tests for UpsertOidcProviderCommand
void main() {
  final instance = UpsertOidcProviderCommandBuilder();
  // TODO add properties to the builder and call build()

  group(UpsertOidcProviderCommand, () {
    // URL slug. Lowercase letters, numbers and dashes. Cannot be changed after creation, and cannot be one of the reserved words login, callback, link, exchange or connections.
    // String name
    test('to test the property `name`', () async {
      // TODO
    });

    // String displayName
    test('to test the property `displayName`', () async {
      // TODO
    });

    // Must use https unless the host is localhost
    // String issuerUrl
    test('to test the property `issuerUrl`', () async {
      // TODO
    });

    // String clientId
    test('to test the property `clientId`', () async {
      // TODO
    });

    // Omit on update to keep the stored secret. Required on create.
    // String clientSecret
    test('to test the property `clientSecret`', () async {
      // TODO
    });

    // Space-separated OIDC scopes; must include openid
    // String scope
    test('to test the property `scope`', () async {
      // TODO
    });

    // bool allowProvisioning
    test('to test the property `allowProvisioning`', () async {
      // TODO
    });

    // bool linkByUsername
    test('to test the property `linkByUsername`', () async {
      // TODO
    });

    // bool enabled
    test('to test the property `enabled`', () async {
      // TODO
    });

  });
}
