import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/user_avatar.dart';
import 'package:receipt_wrangler_mobile/utils/snackbar.dart';

import '../../client/client.dart';

/// Height of the progress bar under the toolbar while a request is in flight.
const double _loadingIndicatorHeight = 4;

class TopAppBar extends StatefulWidget implements PreferredSizeWidget {
  const TopAppBar(
      {super.key,
      required this.titleText,
      this.leadingArrowRedirect,
      this.leadingArrowExtra,
      this.onLeadingArrowPressed,
      this.leadingArrowPop,
      this.actions,
      this.hideAvatar,
      this.surfaceTintColor});

  final String titleText;

  final String? leadingArrowRedirect;

  final dynamic leadingArrowExtra;

  final bool? leadingArrowPop;

  final Function? onLeadingArrowPressed;

  final List<Widget>? actions;

  final bool? hideAvatar;

  final Color? surfaceTintColor;

  @override
  State<TopAppBar> createState() => _TopAppBar();

  @override
  Size get preferredSize => AppBar().preferredSize;
}

class _TopAppBar extends State<TopAppBar> {
  late final AuthModel authModel =
      Provider.of<AuthModel>(context, listen: false);
  late final LoadingModel loadingModel =
      Provider.of<LoadingModel>(context, listen: true);

  Future<void> _logout() async {
    AuthModel authModel = Provider.of<AuthModel>(context, listen: false);
    try {
      var refreshToken = await authModel.getRefreshToken();
      await OpenApiClient.client.getAuthApi().logout(
            logoutCommand: (api.LogoutCommandBuilder()
                  ..refreshToken = refreshToken ?? "")
                .build(),
          );
      await authModel.purgeTokens();
      context.go("/login");
    } catch (e) {
      showErrorSnackbar(context, e as dynamic);
    }
  }

  Widget? getIconButton() {
    if (widget.leadingArrowRedirect != null || widget.leadingArrowPop == true) {
      return IconButton(
        icon: const Icon(Icons.arrow_back),
        onPressed: () {
          if (widget.onLeadingArrowPressed != null) {
            widget.onLeadingArrowPressed!();
          }

          if (widget.leadingArrowPop == true) {
            context.pop();
            return;
          }

          context.go(widget.leadingArrowRedirect ?? "/",
              extra: widget.leadingArrowExtra);
        },
      );
    } else {
      return null;
    }
  }

  Widget getUserAvatar() {
    if (widget.hideAvatar == true) {
      return const SizedBox.shrink();
    }

    // Each item pops the menu itself and then navigates from `this.context` --
    // the State's context, which is the *parent* of the AppBar rather than
    // something inside its toolbar. Keep it that way: a context taken from in
    // here would be at the mercy of the toolbar's lifetime (see
    // [_buildLoadingIndicator]).
    return PopupMenuButton(
        key: const ValueKey('user-avatar-menu'),
        child: const UserAvatar(),
        itemBuilder: (BuildContext context) {
          final permissionsModel =
              Provider.of<PermissionsModel>(context, listen: false);
          final canViewReports = permissionsModel.hasAnyAppPermission([
            api.Permission.appPeriodReportsPeriodRead,
            api.Permission.appPeriodReportsPeriodReadAll,
          ]);

          // These items are TextButtons rather than the plain Text every other
          // menu in the app uses, and a TextButton takes its foreground from
          // `colorScheme.primary` -- so they rendered in the brand blue while
          // every other menu rendered in ordinary text colour. Pin them to the
          // colour a bare PopupMenuItem would have used.
          final itemStyle = TextButton.styleFrom(
            foregroundColor: Theme.of(context).colorScheme.onSurface,
          );

          return [
            PopupMenuItem(
              child: TextButton(
                style: itemStyle,
                onPressed: () {
                  Navigator.pop(context);
                  this.context.push('/profile');
                },
                child: const Text('User Profile'),
              ),
            ),
            if (canViewReports)
              PopupMenuItem(
                child: TextButton(
                  style: itemStyle,
                  onPressed: () {
                    Navigator.pop(context);
                    this.context.push('/reports');
                  },
                  child: const Text('Reports'),
                ),
              ),
            PopupMenuItem(
              child: TextButton(
                style: itemStyle,
                onPressed: () => _logout(),
                child: const Text('Logout'),
              ),
            ),
          ];
        });
  }

  /// The progress bar, **always mounted** — zero-height and empty when idle.
  ///
  /// It must never be `null`, however tempting the conditional looks. `AppBar`
  /// re-parents its whole toolbar into a `Column` *only* when `bottom != null`
  /// (`material/app_bar.dart`), so a null <-> non-null swap changes that slot's
  /// widget from a `ClipRect` to a `Column`, fails `Widget.canUpdate`, and
  /// **unmounts `leading`, `title` and every `actions` child**. Anything in
  /// `actions` that captured its own `BuildContext` and uses it after an
  /// `await` then fails its `mounted` guard and silently gives up — which is
  /// how the receipts-screen overflow menu's Quick Scan and Upload from Gallery
  /// items broke: the AppData refresh they await is the very thing that raises
  /// this bar.
  ///
  /// Keeping a constant height instead would push every screen's body down, so
  /// the height goes to zero rather than the widget going away. The child is
  /// still swapped out when idle: a permanently mounted `LinearProgressIndicator`
  /// animates forever, and `pumpAndSettle` would never return.
  ///
  /// Pinned by `test/widgets/top_app_bar_loading_indicator_test.dart`.
  PreferredSizeWidget _buildLoadingIndicator() {
    final isLoading = loadingModel.isLoading;

    return PreferredSize(
      preferredSize: Size.fromHeight(isLoading ? _loadingIndicatorHeight : 0),
      child: isLoading
          ? const LinearProgressIndicator()
          : const SizedBox.shrink(),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AppBar(
      automaticallyImplyLeading: false,
      leading: getIconButton(),
      title: Text(widget.titleText),
      surfaceTintColor: widget.surfaceTintColor,
      bottom: _buildLoadingIndicator(),
      actions: [
        Padding(
          padding: const EdgeInsets.only(right: 16.0),
          child: getUserAvatar(),
        ),
        ...widget.actions ?? [],
      ],
      centerTitle: false,
    );
  }
}
