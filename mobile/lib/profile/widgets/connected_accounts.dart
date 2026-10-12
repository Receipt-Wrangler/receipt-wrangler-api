import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/services/oidc_service.dart';
import 'package:receipt_wrangler_mobile/utils/snackbar.dart';

/// Lets a signed-in user connect and disconnect identity providers.
///
/// This is what makes the server's "link by username" toggle safe to leave off:
/// connecting from here proves identity with the current session, so the backend
/// never has to guess which account an unseen identity belongs to.
class ConnectedAccounts extends StatefulWidget {
  const ConnectedAccounts({super.key});

  @override
  State<ConnectedAccounts> createState() => _ConnectedAccountsState();
}

class _ConnectedAccountsState extends State<ConnectedAccounts> {
  List<api.OidcConnectionView> _connections = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadConnections();
  }

  Future<void> _loadConnections() async {
    try {
      final response =
          await OpenApiClient.client.getOidcApi().getOidcConnections();

      if (!mounted) return;

      setState(() {
        _connections = response.data?.toList() ?? [];
        _isLoading = false;
      });
    } catch (_) {
      // A server that predates this feature, or a caller without the
      // account-read permission, simply gets no section rather than an error.
      if (!mounted) return;

      setState(() {
        _connections = [];
        _isLoading = false;
      });
    }
  }

  /// Providers this account has not connected yet, taken from the public feature
  /// config the Connect screen already loaded.
  List<api.OidcProviderSummary> _availableProviders(AuthModel auth) {
    final connected = _connections.map((c) => c.providerName).toSet();

    return (auth.featureConfig.oidcProviders?.toList() ?? const [])
        .where((p) => !connected.contains(p.name))
        .toList();
  }

  /// An account created by a provider has only an unusable password, so removing
  /// its last connection would lock it out. The server refuses either way; this
  /// keeps the button from dangling.
  bool _canDisconnect(api.OidcConnectionView connection) {
    return !connection.provisionedUser || _connections.length > 1;
  }

  Future<void> _connect(String providerName) async {
    final authModel = Provider.of<AuthModel>(context, listen: false);
    final basePath = authModel.basePath;

    if (basePath.isEmpty) {
      showErrorSnackbar(context, 'Connect to a server first.');
      return;
    }

    try {
      // linkWithOidc, NOT signInWithOidc. A login flow asks the server "who is
      // this?", so from here it would provision a second account rather than
      // connect this one -- and this button is only ever shown for providers
      // this account has not connected, which is exactly when that happens.
      await linkWithOidc(providerName: providerName);

      if (!mounted) return;

      showSuccessSnackbar(context, 'Account connected');
      await _loadConnections();
    } on OidcSignInCancelled {
      // The user dismissed the browser.
    } on OidcSignInException catch (e) {
      if (mounted) showErrorSnackbar(context, e.message);
    } catch (_) {
      if (mounted) showErrorSnackbar(context, 'Could not connect that account.');
    }
  }

  Future<void> _disconnect(api.OidcConnectionView connection) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Disconnect ${connection.providerDisplayName}'),
        content: Text(
          'Are you sure you want to disconnect ${connection.providerDisplayName}? '
          'You will no longer be able to sign in with it.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Disconnect'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    try {
      await OpenApiClient.client
          .getOidcApi()
          .deleteOidcConnection(name: connection.providerName);

      if (!mounted) return;

      showSuccessSnackbar(
        context,
        '${connection.providerDisplayName} disconnected',
      );
      await _loadConnections();
    } on DioException catch (e) {
      if (mounted) showApiErrorSnackbar(context, e);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthModel>(
      builder: (context, auth, child) {
        final available = _availableProviders(auth);

        if (_isLoading || (_connections.isEmpty && available.isEmpty)) {
          return const SizedBox.shrink();
        }

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 24),
            const Text(
              'Connected accounts',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
            ),
            const SizedBox(height: 8),
            const Text(
              'Sign in with an identity provider instead of your password.',
            ),
            for (final connection in _connections)
              ListTile(
                key: ValueKey('oidc-connection-${connection.providerName}'),
                contentPadding: EdgeInsets.zero,
                title: Text(connection.providerDisplayName),
                subtitle: Text(
                  connection.email?.isNotEmpty == true
                      ? connection.email!
                      : (connection.preferredUsername ?? 'Connected'),
                ),
                trailing: _canDisconnect(connection)
                    ? TextButton(
                        key: ValueKey('oidc-disconnect-${connection.providerName}'),
                        onPressed: () => _disconnect(connection),
                        child: const Text('Disconnect'),
                      )
                    : const Text('Required'),
              ),
            for (final provider in available)
              ListTile(
                key: ValueKey('oidc-available-${provider.name}'),
                contentPadding: EdgeInsets.zero,
                title: Text(provider.displayName),
                subtitle: const Text('Not connected'),
                trailing: TextButton(
                  key: ValueKey('oidc-connect-${provider.name}'),
                  onPressed: () => _connect(provider.name),
                  child: const Text('Connect'),
                ),
              ),
          ],
        );
      },
    );
  }
}
