import 'package:flutter_test/flutter_test.dart';
import 'package:receipt_wrangler_mobile/utils/pkce.dart';

void main() {
  group('generateCodeVerifier', () {
    test('produces 64 characters from the RFC 7636 alphabet', () {
      final verifier = generateCodeVerifier();

      expect(verifier.length, 64);
      expect(RegExp(r'^[A-Za-z0-9\-._~]+$').hasMatch(verifier), isTrue);
    });

    test('produces a different verifier each time', () {
      expect(generateCodeVerifier(), isNot(generateCodeVerifier()));
    });
  });

  group('codeChallengeS256', () {
    // RFC 7636 Appendix B. This is the cross-language contract with the Go
    // side's utils.VerifyPkceS256 -- both must agree byte for byte.
    test('matches the RFC 7636 Appendix B vector', () {
      const verifier = 'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk';

      expect(
        codeChallengeS256(verifier),
        'E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM',
      );
    });

    test('emits unpadded base64url', () {
      final challenge = codeChallengeS256(generateCodeVerifier());

      // Dart's base64UrlEncode pads; Go's RawURLEncoding does not, and the
      // server compares the strings directly.
      expect(challenge.contains('='), isFalse);
      expect(challenge.contains('+'), isFalse);
      expect(challenge.contains('/'), isFalse);
      expect(challenge.length, 43);
    });

    test('is deterministic for a given verifier', () {
      const verifier = 'a-fixed-verifier-value-for-this-test-1234567890';

      expect(codeChallengeS256(verifier), codeChallengeS256(verifier));
    });
  });
}
