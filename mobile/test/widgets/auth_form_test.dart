import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/auth/login/widgets/auth_form.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/persistence/global_shared_preferences.dart';
import 'package:shared_preferences/shared_preferences.dart';

// The login form's password field can be unmasked with an eye in its trailing
// slot, mirroring the desktop auth screen (`app-input`'s showVisibilityEye).
// The icon names the ACTION rather than the state -- an open eye while masked,
// a crossed-out one while revealed -- which is the polarity desktop uses and
// the easiest thing to get backwards.
void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  const toggleKey = ValueKey('password-visibility-toggle');

  const passwordFieldKey = ValueKey('password-field');

  /// Locates the password field the way the E2E SUITE does -- by type and
  /// `name`, not by key.
  ///
  /// Used only by the contract test at the bottom, which exists to prove that
  /// locator still resolves; asserting it any other way would defeat the point.
  /// Deliberately a local copy of `integration_test/helpers/form_actions.dart`'s
  /// `formField` rather than an import: the widget suite is the gating one and
  /// must not depend on `integration_test/`.
  Finder e2eFormFieldLocator(String name) => find.byWidgetPredicate(
        (w) => w is FormBuilderTextField && w.name == name,
      );

  bool isObscured(WidgetTester tester) =>
      tester.widget<FormBuilderTextField>(find.byKey(passwordFieldKey))
          .obscureText;

  setUp(() async {
    // AuthModel.basePath reads GlobalSharedPreferences during build.
    SharedPreferences.setMockInitialValues({});
    await GlobalSharedPreferences.initialize();
  });

  Future<void> pumpForm(
    WidgetTester tester, {
    bool enableLocalSignUp = false,
    List<({String name, String displayName})> oidcProviders = const [],
  }) async {
    final authModel = AuthModel();
    final builder = api.FeatureConfigBuilder()
      ..aiPoweredReceipts = false
      ..enableLocalSignUp = enableLocalSignUp;

    if (oidcProviders.isNotEmpty) {
      builder.oidcProviders.addAll(oidcProviders.map(
        (p) => (api.OidcProviderSummaryBuilder()
              ..name = p.name
              ..displayName = p.displayName)
            .build(),
      ));
    }

    authModel.setFeatureConfig(builder.build());

    await tester.pumpWidget(
      ChangeNotifierProvider<AuthModel>.value(
        value: authModel,
        child: const MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(child: AuthForm()),
          ),
        ),
      ),
    );
    await tester.pump();
  }

  Future<void> tapToggle(WidgetTester tester) async {
    await tester.ensureVisible(find.byKey(toggleKey));
    await tester.pump();
    await tester.tap(find.byKey(toggleKey));
    await tester.pump();
  }

  testWidgets('the password starts masked, offering to show it',
      (tester) async {
    await pumpForm(tester);

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
    await pumpForm(tester);

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
    await pumpForm(tester);

    await tapToggle(tester);
    await tapToggle(tester);

    expect(isObscured(tester), isTrue);
    expect(
      tester.widget<IconButton>(find.byKey(toggleKey)).tooltip,
      'Show Password',
    );
  });

  testWidgets('revealing does not submit the form', (tester) async {
    await pumpForm(tester);

    await tapToggle(tester);

    // A submit would have run the required-validator on the empty fields.
    expect(find.text('This field cannot be empty.'), findsNothing);
    expect(isObscured(tester), isFalse);
  });

  testWidgets('switching to sign up re-masks a revealed password',
      (tester) async {
    await pumpForm(tester, enableLocalSignUp: true);

    await tapToggle(tester);
    expect(isObscured(tester), isFalse);

    // Desktop rebuilds the whole form here (separate routes), so it re-masks.
    // This form only flips a flag and keeps what was typed.
    await tester.ensureVisible(find.text('Create an Account').last);
    await tester.pump();
    await tester.tap(find.text('Create an Account').last);
    await tester.pump();

    expect(isObscured(tester), isTrue);
  });

  testWidgets('the field stays a single FormBuilderTextField named password',
      (tester) async {
    await pumpForm(tester);

    // ~30 e2e specs log in via helpers/login.dart, which fills this field with
    // find.byWidgetPredicate(w is FormBuilderTextField && w.name == 'password').
    // Renaming it, or wrapping it in a way that mounts a second one, takes out
    // the whole integration suite.
    //
    // This is the one assertion that must NOT go through the key: it is pinning
    // the e2e locator itself, so it has to use the same shape the e2e uses.
    expect(e2eFormFieldLocator('password'), findsOneWidget);

    // ...and the keyed field is that same one, so the rest of this file's
    // key-based assertions are talking about the field e2e drives.
    expect(
      tester.widget<FormBuilderTextField>(find.byKey(passwordFieldKey)).name,
      'password',
    );
  });

  // The provider list rides the public feature config, which the
  // Connect-to-Server screen already fetches before this screen is reachable.
  testWidgets('renders no OIDC buttons when none are configured',
      (tester) async {
    await pumpForm(tester);

    expect(find.textContaining('Log in with'), findsNothing);
  });

  testWidgets('renders one login button per configured provider',
      (tester) async {
    await pumpForm(tester, oidcProviders: const [
      (name: 'google', displayName: 'Google'),
      (name: 'keycloak', displayName: 'Keycloak'),
    ]);

    expect(find.text('Log in with Google'), findsOneWidget);
    expect(find.text('Log in with Keycloak'), findsOneWidget);
    expect(find.byKey(const ValueKey('oidc-login-google')), findsOneWidget);
    expect(find.byKey(const ValueKey('oidc-login-keycloak')), findsOneWidget);
  });

  testWidgets('renders no OIDC buttons when the payload omits the field',
      (tester) async {
    // A server that predates the feature sends nothing, and the generated model
    // leaves the field null. Iterating it must not throw.
    await pumpForm(tester, oidcProviders: const []);

    expect(find.textContaining('Log in with'), findsNothing);
    expect(tester.takeException(), isNull);
  });
}
