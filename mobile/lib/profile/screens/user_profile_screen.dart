import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/profile/widgets/connected_accounts.dart';
import 'package:receipt_wrangler_mobile/profile/widgets/crash_reporting_toggle.dart';
import 'package:receipt_wrangler_mobile/profile/widgets/delete_account_dialog.dart';
import 'package:receipt_wrangler_mobile/services/oidc_service.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/screen_wrapper.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/top_app_bar.dart';
import 'package:receipt_wrangler_mobile/utils/snackbar.dart';

/// Fetches this account's identity-provider connections. The default for
/// [UserProfileScreen.loadConnections].
Future<List<api.OidcConnectionView>> _fetchOidcConnections() async {
  final response = await OpenApiClient.client.getOidcApi().getOidcConnections();
  return response.data?.toList() ?? [];
}

class UserProfileScreen extends StatefulWidget {
  const UserProfileScreen({super.key, this.loadConnections, this.confirmDelete});

  /// Injectable for tests: fetches this account's identity-provider
  /// connections, which say whether it has a password to confirm with.
  /// Defaults to a `getOidcConnections` call.
  final Future<List<api.OidcConnectionView>> Function()? loadConnections;

  /// Injectable for tests: runs the provider-confirmed delete flow and returns
  /// the single-use confirmation code. Defaults to [confirmDeleteWithOidc].
  final Future<String> Function({required String providerName})? confirmDelete;

  @override
  State<UserProfileScreen> createState() => _UserProfileScreenState();
}

class _UserProfileScreenState extends State<UserProfileScreen> {
  /// A delete is being confirmed or sent. The passwordless path opens an
  /// external browser session, during which this screen's button stays mounted,
  /// so without this every extra tap starts another flow -- the same reasoning
  /// as `ConnectedAccounts._isLinking`.
  bool _isDeleting = false;

  Future<void> _handleDeleteAccount() async {
    if (_isDeleting) {
      return;
    }

    setState(() => _isDeleting = true);

    // Read before any await: once the server has deleted the account its
    // tokens are dead, so they are purged whether or not this screen is still
    // mounted by then.
    final authModel = Provider.of<AuthModel>(context, listen: false);

    try {
      final command = await _confirmDeletion();
      if (command == null) return;

      await OpenApiClient.client
          .getUserApi()
          .deleteAccount(deleteAccountCommand: command);
      await authModel.purgeTokens();

      if (!mounted) return;

      showSuccessSnackbar(context, 'Account deleted successfully.');
      context.go('/login');
    } on OidcSignInCancelled {
      // The user dismissed the browser.
    } on OidcSignInException catch (e) {
      if (mounted) showErrorSnackbar(context, e.message);
    } on DioException catch (e) {
      if (mounted) showApiErrorSnackbar(context, e);
    } catch (_) {
      if (mounted) {
        showErrorSnackbar(context, 'An error occurred while deleting account.');
      }
    } finally {
      if (mounted) setState(() => _isDeleting = false);
    }
  }

  /// Asks the user to confirm, and returns the command carrying their proof --
  /// a password, or for an account with none a code from signing in at one of
  /// its providers again. Null when they back out.
  Future<api.DeleteAccountCommand?> _confirmDeletion() async {
    List<api.OidcConnectionView> connections;
    try {
      connections = await (widget.loadConnections ?? _fetchOidcConnections)();
    } catch (_) {
      // A server that predates OIDC, or a caller without the account-read
      // permission, falls back to the password dialog, as before.
      connections = const [];
    }

    if (!mounted) return null;

    // provisionedUser is an account fact replicated onto every connection, so
    // any one of them answers for the account.
    if (connections.any((c) => c.provisionedUser)) {
      final providerName =
          await showDeleteAccountProviderDialog(context, connections);
      if (providerName == null || !mounted) return null;

      final code = await (widget.confirmDelete ?? confirmDeleteWithOidc)(
        providerName: providerName,
      );

      return (api.DeleteAccountCommandBuilder()..confirmationCode = code)
          .build();
    }

    final password = await showDeleteAccountDialog(context);
    if (password == null) return null;

    return (api.DeleteAccountCommandBuilder()..password = password).build();
  }

  @override
  Widget build(BuildContext context) {
    return ScreenWrapper(
      appBarWidget: const TopAppBar(
        titleText: 'User Profile',
        leadingArrowPop: true,
        hideAvatar: true,
      ),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 24),
            const Text(
              'Privacy',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
            ),
            const CrashReportingToggle(),
            const ConnectedAccounts(),
            const SizedBox(height: 24),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                border: Border.all(
                  color: Theme.of(context).colorScheme.error,
                ),
                borderRadius: BorderRadius.circular(8),
                color: Theme.of(context).colorScheme.error.withValues(alpha: 0.05),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Danger Zone',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Account deletion is irreversible. This will permanently remove all associated data including receipts, group memberships, and preferences.',
                  ),
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: _isDeleting ? null : _handleDeleteAccount,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Theme.of(context).colorScheme.error,
                        foregroundColor: Theme.of(context).colorScheme.onError,
                      ),
                      child: const Text('Delete Account'),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
