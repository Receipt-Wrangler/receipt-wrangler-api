import 'package:flutter_test/flutter_test.dart';
import 'package:receipt_wrangler_mobile/services/oidc_service.dart';

void main() {
  group('buildOidcLoginUrl', () {
    test('composes the login URL from the server base path', () {
      final url = buildOidcLoginUrl(
        'https://receipts.example.com/api',
        'google',
        'a-challenge',
      );

      expect(
        url,
        'https://receipts.example.com/api/oidc/google/login'
        '?client=mobile&codeChallenge=a-challenge',
      );
    });

    test('does not double the slash when the base path has a trailing one', () {
      final url = buildOidcLoginUrl(
        'https://receipts.example.com/api/',
        'google',
        'a-challenge',
      );

      expect(url.contains('//oidc'), isFalse);
      expect(url.startsWith('https://receipts.example.com/api/oidc/google/login'), isTrue);
    });

    test('escapes the provider name and the challenge', () {
      final url = buildOidcLoginUrl(
        'https://receipts.example.com/api',
        'my provider/../evil',
        'a+challenge/with=chars',
      );

      expect(url.contains('my%20provider%2F..%2Fevil'), isTrue);
      expect(url.contains('/../'), isFalse);
      // A raw '+' would decode to a space server-side and fail the PKCE check.
      expect(url.contains('a+challenge'), isFalse);
    });

    // The mobile leg has no browser cookie to bind to, so the app's own PKCE
    // challenge is the only thing tying the exchange back to this app.
    test('always marks itself as the mobile client and carries a challenge', () {
      final url = buildOidcLoginUrl('https://x/api', 'p', 'c');

      expect(url.contains('client=mobile'), isTrue);
      expect(url.contains('codeChallenge=c'), isTrue);
    });
  });

  group('extractCodeFromCallback', () {
    test('returns the one-time code', () {
      expect(
        extractCodeFromCallback('io.receiptwrangler://oidc?code=abc123'),
        'abc123',
      );
    });

    test('throws with the mapped message when the backend sent an error', () {
      expect(
        () => extractCodeFromCallback('io.receiptwrangler://oidc?error=no_account'),
        throwsA(
          isA<OidcSignInException>().having(
            (e) => e.message,
            'message',
            contains('No Receipt Wrangler account is linked'),
          ),
        ),
      );
    });

    test('throws when the callback carries neither a code nor an error', () {
      expect(
        () => extractCodeFromCallback('io.receiptwrangler://oidc'),
        throwsA(isA<OidcSignInException>()),
      );
    });
  });

  group('extractLinkedProviderFromCallback', () {
    test('returns the provider that was connected', () {
      expect(
        extractLinkedProviderFromCallback('io.receiptwrangler://oidc?linked=google'),
        'google',
      );
    });

    test('throws with the mapped message when the backend sent an error', () {
      expect(
        () => extractLinkedProviderFromCallback(
          'io.receiptwrangler://oidc?error=already_linked',
        ),
        throwsA(
          isA<OidcSignInException>().having(
            (e) => e.message,
            'message',
            contains('already connected'),
          ),
        ),
      );
    });

    test('throws when the callback carries neither a provider nor an error', () {
      expect(
        () => extractLinkedProviderFromCallback('io.receiptwrangler://oidc'),
        throwsA(isA<OidcSignInException>()),
      );
    });

    // A link mints no session, so a callback carrying a redeemable code is the
    // login flow's shape and must not be read as a successful link.
    test('a code alone is not a successful link', () {
      expect(
        () => extractLinkedProviderFromCallback(
          'io.receiptwrangler://oidc?code=an-exchange-code',
        ),
        throwsA(isA<OidcSignInException>()),
      );
    });

    test('an error wins over a linked parameter', () {
      expect(
        () => extractLinkedProviderFromCallback(
          'io.receiptwrangler://oidc?linked=google&error=no_account',
        ),
        throwsA(isA<OidcSignInException>()),
      );
    });
  });

  group('oidcErrorMessage', () {
    test('maps every code the backend can send', () {
      for (final code in [
        'unknown_provider',
        'no_account',
        'account_exists',
        'already_linked',
        'invalid_state',
        'nonce_mismatch',
        'no_id_token',
        'provider_error',
      ]) {
        expect(oidcErrorMessage(code), isNotEmpty);
        // Never echo the raw code at the user.
        expect(oidcErrorMessage(code), isNot(code));
      }
    });

    test('falls back for an unrecognized code', () {
      expect(oidcErrorMessage('something_new'), 'Sign in failed. Please try again.');
    });
  });

  test('the callback scheme is reverse-DNS of a domain the project owns', () {
    // RFC 8252 section 7.1, and it must match the Android applicationId and the
    // manifest's intent-filter.
    expect(oidcCallbackScheme, 'io.receiptwrangler');
    expect(oidcCallbackUrl, 'io.receiptwrangler://oidc');
  });
}
