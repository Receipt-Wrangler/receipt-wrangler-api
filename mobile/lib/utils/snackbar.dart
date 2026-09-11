import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:sentry_flutter/sentry_flutter.dart';

void showSuccessSnackbar(BuildContext context, String message,
    {SnackBarAction? action}) {
  ScaffoldMessenger.of(context).showSnackBar(SnackBar(
      content: Text(message), backgroundColor: Colors.green, action: action));
}

/// A neutral, non-error notice — used where the app quietly changes course and
/// wants to say why (e.g. falling back to the gallery when the camera is off).
/// Left on the theme's default background so it doesn't read as a failure.
void showInfoSnackbar(BuildContext context, String message,
    {SnackBarAction? action}) {
  ScaffoldMessenger.of(context)
      .showSnackBar(SnackBar(content: Text(message), action: action));
}

void showErrorSnackbar(BuildContext context, String message,
    {SnackBarAction? action}) {
  ScaffoldMessenger.of(context).showSnackBar(SnackBar(
    content: Text(message),
    backgroundColor: Colors.red,
    action: action,
  ));
}

/// Shows the server's `errorMsg` when [error] is a [DioException] that carries
/// one, and a generic notice otherwise.
///
/// Takes `Object` rather than `DioException` on purpose. Every call site sits
/// in a `catch (e)`, where `e` is an `Object` — a narrower parameter made each
/// one an implicit downcast (or an explicit `e as DioException` / `e as
/// dynamic`), which **throws a `TypeError` from inside the catch block** the
/// moment the failure is anything other than a Dio error. That replaced the
/// real error with an unhandled one at the exact point the code was trying to
/// report it, so the original was lost and the user saw nothing.
///
/// A non-Dio error is reported to Sentry, since it is by definition unexpected
/// on a path that was written to handle API failures.
void showApiErrorSnackbar(BuildContext context, Object error,
    [StackTrace? stackTrace]) {
  String? message;

  if (error is DioException) {
    final data = error.response?.data;
    if (data is Map) {
      final raw = data['errorMsg'];
      if (raw is String && raw.isNotEmpty) {
        message = raw;
      }
    }
  } else {
    unawaited(
        Sentry.captureException(error, stackTrace: stackTrace).catchError((_) {
      // Reporting must never displace the message the user is owed.
      return SentryId.empty();
    }));
  }

  ScaffoldMessenger.of(context).showSnackBar(SnackBar(
    content: Text(message ?? 'An error occurred'),
    backgroundColor: Colors.red,
  ));
}
