import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart' as api;
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/constants/font.dart';
import 'package:receipt_wrangler_mobile/enums/form_state.dart';
import 'package:receipt_wrangler_mobile/models/user_model.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/list_item_lead.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/slidable_edit_button.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/slidable_widget.dart';
import 'package:receipt_wrangler_mobile/utils/currency.dart';
import 'package:receipt_wrangler_mobile/utils/date.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';

import '../../models/group_model.dart';
import '../../models/permissions_model.dart';
import '../../shared/functions/permissions.dart';
import '../../shared/widgets/list_item_trailing_status.dart';

class ReceiptListItem extends StatefulWidget {
  const ReceiptListItem(
      {super.key, required this.receipt, this.displayGroup = false});

  final api.Receipt receipt;

  final bool displayGroup;

  @override
  State<ReceiptListItem> createState() => _ReceiptListItem();
}

class _ReceiptListItem extends State<ReceiptListItem> {
  late final permissionsModel =
      Provider.of<PermissionsModel>(context, listen: false);
  late final groupModel = Provider.of<GroupModel>(context, listen: false);

  Widget getStatusText() {
    return ListItemTrailingStatus(
        color: getStatusColor(),
        text: receiptStatusLabel(widget.receipt.status));
  }

  Widget getLeadingWidget() {
    return ListItemLead(
        date: widget.receipt.createdAt ?? "", color: getStatusColor());
  }

  Color getStatusColor() {
    return receiptStatusColor(widget.receipt.status);
  }

  Widget getSubtitleText() {
    var userNotFoundText = "User not found!";
    var userModel = Provider.of<UserModel>(context, listen: false);

    var user = userModel.getUserById(widget.receipt.paidByUserId.toString());
    var formattedAmount = formatCurrency(context, widget.receipt.amount);
    var formattedDate =
        formatDate(defaultDateFormat, DateTime.parse(widget.receipt.date));

    return Text(
        "${formattedAmount} paid by ${user?.displayName ?? userNotFoundText} on ${formattedDate}");
  }

  Widget buildListTile() {
    var titleText = widget.receipt.name;

    if (widget.displayGroup) {
      var group = groupModel.getGroupById(widget.receipt.groupId.toString());
      titleText = "${widget.receipt.name}\n(${group?.name})";
    }

    return ListTile(
        title: Text(
          titleText,
          style: boldText,
        ),
        subtitle: getSubtitleText(),
        leading: getLeadingWidget(),
        trailing: getStatusText(),
        onTap: () => navigateToReceipt(WranglerFormState.view));
  }

  Widget buildEditButton() {
    return SlidableEditButton(
        onPressed: () => navigateToReceipt(WranglerFormState.edit));
  }

  void navigateToReceipt(WranglerFormState formState) {
    context.go("/receipts/${widget.receipt.id}/${formState.name}");
  }

  @override
  Widget build(BuildContext context) {
    var canEdit = canEditReceipt(permissionsModel, widget.receipt.groupId);

    return SlidableWidget(
        slideEnabled: canEdit,
        endActionPaneChildren: [buildEditButton()],
        slidableChild: buildListTile());
  }
}
