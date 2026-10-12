import 'package:test/test.dart';
import 'package:openapi/openapi.dart';


/// tests for OidcApi
void main() {
  final instance = Openapi().getOidcApi();

  group(OidcApi, () {
    // Disconnect a provider from the caller's account
    //
    // Refused when it is the caller's last connection and the account was created by that provider, because such an account has no password to fall back on.
    //
    //Future deleteOidcConnection(String name) async
    test('test deleteOidcConnection', () async {
      // TODO
    });

    // List the caller's connected accounts
    //
    //Future<BuiltList<OidcConnectionView>> getOidcConnections() async
    test('test getOidcConnections', () async {
      // TODO
    });

    // OIDC redirect URI
    //
    // Where the identity provider returns the user. This is the exact URL that must be registered with the provider. Not called by clients directly.
    //
    //Future oidcCallback(String name, { String code, String state }) async
    test('test oidcCallback', () async {
      // TODO
    });

    // Redeem a mobile sign-in code
    //
    // Trades the single-use code the app received on its private-use URL scheme for a session. The PKCE verifier is what proves this is the app that started the flow, so an app that intercepted the redirect cannot redeem the code. Returns the same payload as login with tokensInBody, and never sets a cookie.
    //
    //Future<AppData> oidcExchange(OidcExchangeCommand oidcExchangeCommand) async
    test('test oidcExchange', () async {
      // TODO
    });

    // Connect a provider to the signed-in account
    //
    // Starts the same flow as a login, but the session proves who the caller is, so the callback links the identity directly instead of matching or provisioning. Navigate to this URL; do not fetch it.
    //
    //Future oidcLinkStart(String name) async
    test('test oidcLinkStart', () async {
      // TODO
    });

    // Start an OIDC login
    //
    // Redirects the user agent to the configured identity provider. The API acts as the relying party, so the whole exchange (PKCE, state, nonce, code exchange, ID token verification) happens server-side and no client ever handles an identity provider token. Navigate to this URL; do not fetch it.
    //
    //Future oidcLogin(String name, { String client, String codeChallenge }) async
    test('test oidcLogin', () async {
      // TODO
    });

  });
}
