import 'dart:convert';
import 'dart:math';

import 'package:crypto/crypto.dart';

/// The character set RFC 7636 allows in a code verifier.
const _verifierAlphabet =
    'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';

/// Generates a PKCE code verifier: 64 characters from the RFC 7636 alphabet,
/// drawn from a cryptographically secure source.
///
/// The verifier never leaves the app. It is what proves, at the exchange
/// endpoint, that this is the same app that started the flow -- so another app
/// that registered the same URL scheme and intercepted the redirect still
/// cannot redeem the code.
String generateCodeVerifier({Random? random}) {
  final source = random ?? Random.secure();

  return List.generate(
    64,
    (_) => _verifierAlphabet[source.nextInt(_verifierAlphabet.length)],
  ).join();
}

/// Derives the S256 code challenge for a verifier.
///
/// The padding strip is load-bearing and is the whole reason this is a named,
/// tested function: Dart's [base64UrlEncode] pads with '=', Go's
/// `base64.RawURLEncoding` does not, and the server compares the two strings
/// directly. A padded challenge would pass every unit test on each side
/// independently and fail only on a real device.
String codeChallengeS256(String verifier) {
  final digest = sha256.convert(utf8.encode(verifier));

  return base64UrlEncode(digest.bytes).replaceAll('=', '');
}
