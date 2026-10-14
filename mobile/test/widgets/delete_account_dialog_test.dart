import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:receipt_wrangler_mobile/profile/widgets/delete_account_dialog.dart';

// The delete-account dialog masks its password the same way the login form
// does, and offers the same eye to reveal it -- the icon names the ACTION (an
// open eye while masked, crossed-out while revealed), matching the desktop
// visibility eye. Both of the app's password fields must behave identically,
// so what auth_form_test asserts about the login field is asserted here too.
void main() {
  const toggleKey = ValueKey('password-visibility-toggle');
  const passwordFieldKey = ValueKey('password-field');

  /// Opens the dialog through its public entry point -- the widget itself is
  /// private, so `showDeleteAccountDialog` is the only way in, and it is also
  /// how the profile screen reaches it.
  Future<void> pumpDialog(WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (context) => Scaffold(
            body: Center(
              child: ElevatedButton(
                onPressed: () => showDeleteAccountDialog(context),
                child: const Text('open'),
              ),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();
  }

  bool isObscured(WidgetTester tester) =>
      tester.widget<TextField>(find.byKey(passwordFieldKey)).obscureText;

  Future<void> tapToggle(WidgetTester tester) async {
    await tester.tap(find.byKey(toggleKey));
    await tester.pump();
  }

  testWidgets('the password starts masked, offering to show it',
      (tester) async {
    await pumpDialog(tester);

    expect(isObscured(tester), isTrue);
    expect(find.byKey(toggleKey), findsOneWidget);
    expect(
      tester.widget<IconButton>(find.byKey(toggleKey)).tooltip,
      'Show Password',
    );
    expect(
      find.descendant(
          of: find.byKey(toggleKey), matching: find.byIcon(Icons.visibility)),
      findsOneWidget,
    );
  });

  testWidgets('tapping the eye reveals the password and offers to hide it',
      (tester) async {
    await pumpDialog(tester);

    await tapToggle(tester);

    expect(isObscured(tester), isFalse);
    expect(
      tester.widget<IconButton>(find.byKey(toggleKey)).tooltip,
      'Hide Password',
    );
    expect(
      find.descendant(
          of: find.byKey(toggleKey),
          matching: find.byIcon(Icons.visibility_off)),
      findsOneWidget,
    );
  });

  testWidgets('tapping it again re-masks', (tester) async {
    await pumpDialog(tester);

    await tapToggle(tester);
    await tapToggle(tester);

    expect(isObscured(tester), isTrue);
    expect(
      tester.widget<IconButton>(find.byKey(toggleKey)).tooltip,
      'Show Password',
    );
  });

  testWidgets('revealing does not submit the dialog', (tester) async {
    await pumpDialog(tester);

    await tester.enterText(find.byKey(passwordFieldKey), 'MySecretPass123');
    await tester.pump();

    await tapToggle(tester);

    // Still open, still holding what was typed -- the eye is not a submit.
    expect(find.text('Delete Account'), findsWidgets);
    expect(find.text('MySecretPass123'), findsOneWidget);
    expect(isObscured(tester), isFalse);
  });

  // An account an identity provider created has no password to type, so it
  // confirms by signing in at one of its providers again instead.
  group('showDeleteAccountProviderDialog', () {
    api.OidcConnectionView connection(String name, String displayName) =>
        (api.OidcConnectionViewBuilder()
              ..providerName = name
              ..providerDisplayName = displayName
              ..provisionedUser = true
              ..linkedAt = DateTime.utc(2026, 1, 1))
            .build();

    final connections = [
      connection('google', 'Google'),
      connection('keycloak', 'Keycloak'),
    ];

    /// Opens the dialog and records what it returns once it closes.
    Future<List<String?>> pumpProviderDialog(WidgetTester tester) async {
      final results = <String?>[];

      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => Scaffold(
              body: Center(
                child: ElevatedButton(
                  onPressed: () async => results.add(
                    await showDeleteAccountProviderDialog(context, connections),
                  ),
                  child: const Text('open'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      return results;
    }

    testWidgets('offers one button per connection, and no password field',
        (tester) async {
      await pumpProviderDialog(tester);

      expect(find.byKey(const ValueKey('delete-confirm-google')), findsOneWidget);
      expect(
          find.byKey(const ValueKey('delete-confirm-keycloak')), findsOneWidget);
      expect(find.text('Confirm with Google'), findsOneWidget);
      expect(find.text('Confirm with Keycloak'), findsOneWidget);
      expect(find.byKey(passwordFieldKey), findsNothing);
    });

    testWidgets('returns the chosen provider name', (tester) async {
      final results = await pumpProviderDialog(tester);

      await tester.tap(find.byKey(const ValueKey('delete-confirm-keycloak')));
      await tester.pumpAndSettle();

      expect(results, ['keycloak']);
      expect(find.text('Confirm with Keycloak'), findsNothing);
    });

    testWidgets('cancel returns null', (tester) async {
      final results = await pumpProviderDialog(tester);

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(results, [null]);
    });
  });
}
