import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../models/receipt_model.dart';
import '../../shared/widgets/receipt_edit_popup_menu.dart';

/// The receipt form screen's app-bar menu.
///
/// This used to also carry image and comment menus, selected by testing
/// `state.fullPath` for "images" / "comments". No such route exists -- the full
/// set is `/receipts/add`, `/receipts/:id/view` and `/receipts/:id/edit` (plus
/// the group, profile, reports and search routes) -- and the image screen is
/// pushed as a plain `MaterialPageRoute` from `receipt_form.dart`, so the path
/// never changes when it opens. Both branches were unreachable, and the image
/// one was a stale copy of `ReceiptImageAppBar` that navigated to a URL
/// (`/receipts/:id/images/edit`) the router does not define. They were deleted
/// rather than merged; `ReceiptImageAppBar` is the live implementation.
class ReceiptAppBarActionBuilder {
  late final ReceiptModel receiptModel;

  late final BuildContext context;

  ReceiptAppBarActionBuilder(BuildContext context, ReceiptModel receiptModel) {
    this.context = context;
    this.receiptModel = receiptModel;
  }

  List<Widget> buildAppBarMenu(GoRouterState state) {
    return [_buildReceiptAppBarMenu(state.fullPath!)];
  }

  Widget _buildReceiptAppBarMenu(String fullPath) {
    if (fullPath.contains("view")) {
      return ReceiptEditPopupMenu(
          groupId: receiptModel.receipt.groupId,
          popupMenuChildren: [
            PopupMenuItem(
              child: Text("Edit"),
              onTap: () => _goAfterMenu(
                  "/receipts/${receiptModel.receipt.id}/edit"),
            )
          ]);
    }

    return SizedBox.shrink();
  }

  // A PopupMenuItem invokes onTap *after* it pops the menu route. Navigating
  // straight from the (now-departing) menu/page context is lost under the
  // receipt routes' NoTransitionPage swap, which removes the old page in the
  // same frame. Capture the stable GoRouter up front and navigate once the
  // menu has finished dismissing so the go isn't clobbered by the pop.
  void _goAfterMenu(String location) {
    final router = GoRouter.of(context);
    Future.delayed(
        const Duration(milliseconds: 50), () => router.go(location));
  }
}
