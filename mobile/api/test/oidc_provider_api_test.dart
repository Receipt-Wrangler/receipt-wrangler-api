import 'package:test/test.dart';
import 'package:openapi/openapi.dart';


/// tests for OidcProviderApi
void main() {
  final instance = Openapi().getOidcProviderApi();

  group(OidcProviderApi, () {
    // Create an OIDC provider
    //
    //Future<OidcProviderView> createOidcProvider(UpsertOidcProviderCommand upsertOidcProviderCommand) async
    test('test createOidcProvider', () async {
      // TODO
    });

    // Delete an OIDC provider
    //
    // Accounts linked to it lose the ability to sign in with it.
    //
    //Future deleteOidcProvider(String oidcProviderId) async
    test('test deleteOidcProvider', () async {
      // TODO
    });

    // Get an OIDC provider
    //
    // The client secret is never returned; hasClientSecret reports whether one is stored.
    //
    //Future<OidcProviderView> getOidcProviderById(String oidcProviderId) async
    test('test getOidcProviderById', () async {
      // TODO
    });

    // Get paged OIDC providers
    //
    //Future<PagedData> getPagedOidcProviders(PagedRequestCommand pagedRequestCommand) async
    test('test getPagedOidcProviders', () async {
      // TODO
    });

    // Update an OIDC provider
    //
    // Omit clientSecret to keep the stored one. The name cannot be changed -- it is part of the redirect URI already registered with the identity provider.
    //
    //Future<OidcProviderView> updateOidcProvider(String oidcProviderId, UpsertOidcProviderCommand upsertOidcProviderCommand) async
    test('test updateOidcProvider', () async {
      // TODO
    });

  });
}
