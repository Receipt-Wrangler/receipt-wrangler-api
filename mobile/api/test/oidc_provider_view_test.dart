import 'package:test/test.dart';
import 'package:openapi/openapi.dart';

// tests for OidcProviderView
void main() {
  final instance = OidcProviderViewBuilder();
  // TODO add properties to the builder and call build()

  group(OidcProviderView, () {
    // int id
    test('to test the property `id`', () async {
      // TODO
    });

    // URL slug for this provider; immutable after creation
    // String name
    test('to test the property `name`', () async {
      // TODO
    });

    // String displayName
    test('to test the property `displayName`', () async {
      // TODO
    });

    // OIDC discovery base, e.g. https://accounts.google.com
    // String issuerUrl
    test('to test the property `issuerUrl`', () async {
      // TODO
    });

    // String clientId
    test('to test the property `clientId`', () async {
      // TODO
    });

    // Space-separated OIDC scopes; must include openid
    // String scope
    test('to test the property `scope`', () async {
      // TODO
    });

    // Create a local account for an identity we have never seen
    // bool allowProvisioning
    test('to test the property `allowProvisioning`', () async {
      // TODO
    });

    // On a first login only, attach to an existing local account whose username equals the preferred_username claim. Off by default; that claim is neither stable nor unique, and some providers recycle released usernames.
    // bool linkByUsername
    test('to test the property `linkByUsername`', () async {
      // TODO
    });

    // bool enabled
    test('to test the property `enabled`', () async {
      // TODO
    });

    // Whether a secret is stored. The secret itself is never returned.
    // bool hasClientSecret
    test('to test the property `hasClientSecret`', () async {
      // TODO
    });

    // The exact redirect URI to register with the identity provider
    // String redirectUri
    test('to test the property `redirectUri`', () async {
      // TODO
    });

    // DateTime createdAt
    test('to test the property `createdAt`', () async {
      // TODO
    });

    // DateTime updatedAt
    test('to test the property `updatedAt`', () async {
      // TODO
    });

  });
}
