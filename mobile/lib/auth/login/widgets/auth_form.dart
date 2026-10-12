import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:form_builder_validators/form_builder_validators.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/constants/spacing.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/category_model.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/models/tag_model.dart';
import 'package:receipt_wrangler_mobile/models/user_model.dart';
import 'package:receipt_wrangler_mobile/models/user_preferences_model.dart';
import 'package:receipt_wrangler_mobile/services/oidc_service.dart';
import 'package:receipt_wrangler_mobile/utils/auth.dart';
import 'package:receipt_wrangler_mobile/utils/snackbar.dart';

import '../../../models/system_settings_model.dart';
import '../../../utils/currency.dart';

class AuthForm extends StatefulWidget {
  const AuthForm({super.key});

  @override
  State<AuthForm> createState() => _Login();
}

class _Login extends State<AuthForm> {
  final _formKey = GlobalKey<FormBuilderState>();
  late final screenSize = MediaQuery.of(context).size;
  bool isSignUp = false;
  bool isLoading = false;

  /// Whether the password field is masked. Mirrors the desktop auth screen's
  /// visibility eye (`app-input`'s `showVisibilityEye`), which starts masked and
  /// stays however the user left it across a submit or a failed validation.
  bool obscurePassword = true;

  Future<void> _submit() async {
    _formKey.currentState!.save();

    if (_formKey.currentState!.validate()) {
      if (isSignUp) {
        await signUp();
      } else {
        await login();
      }
    }
  }

  Future<void> login() async {
    var form = _formKey.currentState!.value;
    var command = (api.LoginCommandBuilder()
          ..username = form["username"]
          ..password = form["password"])
        .build();
    try {
      toggleIsLoading();
      var appDataResponse =
          await OpenApiClient.client.getAuthApi().login(loginCommand: command);
      await _onLoginSuccess(appDataResponse.data as api.AppData);
    } catch (e) {
      toggleIsLoading();
      print(e);
      showApiErrorSnackbar(context, e as dynamic);
    }
  }

  Future<void> signUp() async {
    var form = _formKey.currentState!.value;
    toggleIsLoading();
    OpenApiClient.client
        .getAuthApi()
        .signUp(
            signUpCommand: (api.SignUpCommandBuilder()
                  ..username = form["username"]
                  ..password = form["password"]
                  ..displayName = form["displayName"])
                .build())
        .then((data) async => {await login()})
        .catchError((err) => {
              toggleIsLoading(),
              print(err),
              showApiErrorSnackbar(context, err),
            });
  }

  /// Signs in through an external identity provider.
  ///
  /// The whole OIDC exchange happens on the server; this only opens the system
  /// browser and redeems the one-time code that comes back. The response has the
  /// same shape as a password login with tokensInBody, so it goes through the
  /// same _onLoginSuccess path.
  Future<void> loginWithOidc(api.OidcProviderSummary provider) async {
    final authModel = Provider.of<AuthModel>(context, listen: false);
    final basePath = authModel.basePath;

    if (basePath.isEmpty) {
      showErrorSnackbar(context, "Connect to a server first.");
      return;
    }

    try {
      toggleIsLoading();
      final appData = await signInWithOidc(
        basePath: basePath,
        providerName: provider.name,
      );

      if (!mounted) {
        return;
      }

      await _onLoginSuccess(appData);
    } on OidcSignInCancelled {
      // The user dismissed the browser. Nothing to report.
      if (mounted) {
        toggleIsLoading();
      }
    } on OidcSignInException catch (e) {
      if (mounted) {
        toggleIsLoading();
        showErrorSnackbar(context, e.message);
      }
    } catch (e) {
      if (mounted) {
        toggleIsLoading();
        showErrorSnackbar(context, "Sign in failed. Please try again.");
      }
    }
  }

  /// One button per enabled provider, or nothing when none is configured.
  ///
  /// The list rides the public feature config, which the Connect-to-Server
  /// screen already fetches before this screen is reachable, so there is no
  /// extra request here.
  Widget _getOidcButtons(AuthModel auth) {
    final providers = auth.featureConfig.oidcProviders?.toList() ?? const [];

    if (isSignUp || providers.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      children: [
        const SizedBox(height: 16),
        Row(children: [
          const Expanded(child: Divider()),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            child: Text("or", style: Theme.of(context).textTheme.bodySmall),
          ),
          const Expanded(child: Divider()),
        ]),
        const SizedBox(height: 8),
        for (final provider in providers) ...[
          Row(children: [
            Expanded(
              child: CupertinoButton(
                key: ValueKey("oidc-login-${provider.name}"),
                onPressed: () => loginWithOidc(provider),
                child: Text("Log in with ${provider.displayName}"),
              ),
            )
          ]),
        ],
      ],
    );
  }

  void toggleIsLoading() {
    setState(() {
      isLoading = !isLoading;
    });
  }

  Future<void> _onLoginSuccess(api.AppData appData) async {
    var authModel = Provider.of<AuthModel>(context, listen: false);
    var groupModel = Provider.of<GroupModel>(context, listen: false);
    var userModel = Provider.of<UserModel>(context, listen: false);
    var userPreferencesModel =
        Provider.of<UserPreferencesModel>(context, listen: false);
    var categoryModel = Provider.of<CategoryModel>(context, listen: false);
    var tagModel = Provider.of<TagModel>(context, listen: false);
    var systemSettingsModel =
        Provider.of<SystemSettingsModel>(context, listen: false);
    var permissionsModel =
        Provider.of<PermissionsModel>(context, listen: false);

    await storeAppData(authModel, groupModel, userModel, userPreferencesModel,
        categoryModel, tagModel, systemSettingsModel, permissionsModel, appData);
    registerCustomCurrency(context);
    context.go("/groups");
  }

  Widget _getDisplaynameField() {
    if (isSignUp) {
      return FormBuilderTextField(
          name: "displayName",
          decoration: const InputDecoration(
              labelText: "Displayname", border: OutlineInputBorder()),
          validator: FormBuilderValidators.compose([
            FormBuilderValidators.required(),
          ]));
    } else {
      return const SizedBox.shrink();
    }
  }

  Widget _getSignUpButton() {
    var buttonText = "";

    if (isSignUp) {
      buttonText = "Return to Login";
    } else {
      buttonText = "Create an Account";
    }

    return CupertinoButton(
        onPressed: () {
          setState(() {
            isSignUp = !isSignUp;
            // Desktop re-masks here because Login and Sign Up are separate
            // routes, so the form is destroyed and rebuilt. This form only
            // flips a flag and keeps what was typed, so without this a
            // revealed password would stay on screen across a deliberate
            // switch -- worse than desktop rather than equal to it.
            obscurePassword = true;
          });
        },
        child: Text(buttonText));
  }

  Widget _getSubmitButtonText() {
    if (isSignUp) {
      return const Text("Create an Account");
    } else {
      return const Text("Log In");
    }
  }

  Widget _getServerInfoText(AuthModel server) {
    if (isSignUp) {
      return Text('Signing up on: ${server.basePath}');
    }
    return Text('Logging into: ${server.basePath}');
  }

  Widget _getChangeServerButton() {
    return CupertinoButton(
        onPressed: () {
          context.go("/");
        },
        child: const Text("Change Server"));
  }

  @override
  Widget build(BuildContext context) {
    return AutofillGroup(
        child: FormBuilder(
      key: _formKey,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          if (isLoading) ...[
            const CircularProgressIndicator(),
          ] else ...[
            SvgPicture.asset(
              "assets/branding/logo-large.svg",
              width: screenSize.width * 0.25,
              height: screenSize.width * 0.25,
            ),
            const SizedBox(
              height: 10,
            ),
            Consumer<AuthModel>(
              builder: (context, auth, child) {
                return _getServerInfoText(auth);
              },
            ),
            headerSpacing,
            _getDisplaynameField(),
            isSignUp ? textFieldSpacing : const SizedBox.shrink(),
            FormBuilderTextField(
                name: "username",
                autofillHints: const [AutofillHints.username],
                decoration: const InputDecoration(
                    labelText: "Username", border: OutlineInputBorder()),
                validator: FormBuilderValidators.compose([
                  FormBuilderValidators.required(),
                ])),
            textFieldSpacing,
            FormBuilderTextField(
                key: const ValueKey("password-field"),
                name: "password",
                autofillHints: const [AutofillHints.password],
                obscureText: obscurePassword,
                decoration: InputDecoration(
                    labelText: "Password",
                    border: const OutlineInputBorder(),
                    // The icon names the ACTION, not the state: an open eye
                    // while masked ("show me"), a crossed-out one while
                    // revealed ("hide it"). Same polarity as the desktop eye
                    // and the delete-account dialog.
                    suffixIcon: IconButton(
                      key: const ValueKey("password-visibility-toggle"),
                      icon: Icon(obscurePassword
                          ? Icons.visibility
                          : Icons.visibility_off),
                      tooltip:
                          obscurePassword ? "Show Password" : "Hide Password",
                      onPressed: () {
                        setState(() {
                          obscurePassword = !obscurePassword;
                        });
                      },
                    )),
                validator: FormBuilderValidators.compose([
                  FormBuilderValidators.required(),
                ])),
            lastFieldSpacing,
            Row(
              children: [
                Expanded(
                  child: CupertinoButton.filled(
                      onPressed: () {
                        _submit();
                      },
                      child: _getSubmitButtonText()),
                )
              ],
            ),
            Consumer<AuthModel>(
              builder: (context, auth, child) {
                if (auth.featureConfig.enableLocalSignUp) {
                  return _getSignUpButton();
                } else {
                  return const SizedBox.shrink();
                }
              },
            ),
            Consumer<AuthModel>(
              builder: (context, auth, child) => _getOidcButtons(auth),
            ),
            _getChangeServerButton(),
          ],
        ],
      ),
    ));
  }
}
