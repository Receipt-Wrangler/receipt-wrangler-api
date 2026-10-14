import 'package:flutter/material.dart';
import 'package:openapi/openapi.dart' as api;

const _permanentWarning = Text.rich(
  TextSpan(
    children: [
      TextSpan(text: 'This action is '),
      TextSpan(
        text: 'permanent',
        style: TextStyle(fontWeight: FontWeight.bold),
      ),
      TextSpan(
        text:
            ' and cannot be undone. All of your data, including receipts, group memberships, and preferences will be deleted.',
      ),
    ],
  ),
);

ButtonStyle _destructiveButtonStyle(BuildContext context) {
  return ElevatedButton.styleFrom(
    backgroundColor: Theme.of(context).colorScheme.error,
    foregroundColor: Theme.of(context).colorScheme.onError,
  );
}

Future<String?> showDeleteAccountDialog(BuildContext context) {
  return showDialog<String?>(
    context: context,
    builder: (context) => const _DeleteAccountDialog(),
  );
}

/// The confirmation for an account with no password of its own -- one an
/// identity provider created. Offers one button per connected provider and
/// returns the chosen provider's name, or null when the user cancels.
Future<String?> showDeleteAccountProviderDialog(
  BuildContext context,
  List<api.OidcConnectionView> connections,
) {
  return showDialog<String?>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('Delete Account'),
      content: const Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _permanentWarning,
          SizedBox(height: 12),
          Text(
            'Your account does not have a password. To confirm account '
            'deletion, sign in again with a connected account.',
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, null),
          child: const Text('Cancel'),
        ),
        for (final connection in connections)
          ElevatedButton(
            key: ValueKey('delete-confirm-${connection.providerName}'),
            onPressed: () => Navigator.pop(context, connection.providerName),
            style: _destructiveButtonStyle(context),
            child: Text('Confirm with ${connection.providerDisplayName}'),
          ),
      ],
    ),
  );
}

class _DeleteAccountDialog extends StatefulWidget {
  const _DeleteAccountDialog();

  @override
  State<_DeleteAccountDialog> createState() => _DeleteAccountDialogState();
}

class _DeleteAccountDialogState extends State<_DeleteAccountDialog> {
  final _passwordController = TextEditingController();
  bool _obscurePassword = true;

  @override
  void dispose() {
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Delete Account'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _permanentWarning,
          const SizedBox(height: 12),
          const Text('Please enter your password to confirm account deletion.'),
          const SizedBox(height: 16),
          TextField(
            // Same key the login form's password field uses -- the two never
            // share a tree, and matching them keeps both password fields
            // addressable the same way, as their eye toggles already are.
            key: const ValueKey('password-field'),
            controller: _passwordController,
            obscureText: _obscurePassword,
            decoration: InputDecoration(
              labelText: 'Password',
              suffixIcon: IconButton(
                key: const ValueKey('password-visibility-toggle'),
                icon: Icon(
                  _obscurePassword ? Icons.visibility : Icons.visibility_off,
                ),
                tooltip: _obscurePassword ? 'Show Password' : 'Hide Password',
                onPressed: () {
                  setState(() {
                    _obscurePassword = !_obscurePassword;
                  });
                },
              ),
            ),
            onChanged: (_) => setState(() {}),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, null),
          child: const Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: _passwordController.text.isEmpty
              ? null
              : () => Navigator.pop(context, _passwordController.text),
          style: _destructiveButtonStyle(context),
          child: const Text('Delete Account'),
        ),
      ],
    );
  }
}
