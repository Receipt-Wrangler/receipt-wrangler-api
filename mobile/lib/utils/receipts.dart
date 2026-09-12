import 'dart:convert';
import 'dart:typed_data';

import 'package:built_collection/built_collection.dart';
import 'package:built_value/json_object.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:openapi/openapi.dart';
import 'package:receipt_wrangler_mobile/constants/colors.dart';
import 'package:receipt_wrangler_mobile/enums/form_state.dart';
import 'package:receipt_wrangler_mobile/utils/forms.dart';

String? getReceiptId(BuildContext context) {
  return GoRouterState.of(context).pathParameters["receiptId"];
}

Receipt getDefaultReceipt() {
  return (ReceiptBuilder()
        ..id = 0
        ..date = DateTime.now().toIso8601String()
        ..groupId = 0
        ..paidByUserId = 0
        ..amount = "0"
        ..name = ""
        ..status = ReceiptStatus.OPEN
        ..categories = ListBuilder<Category>([])
        ..tags = ListBuilder<Tag>([]))
      .build();
}

/// Human label for a receipt status.
///
/// A status this build has no case for falls through to the desktop
/// `formatStatus` rule (SNAKE_CASE -> Title Case) so the two clients degrade
/// identically, and [ReceiptStatus.empty] — the "unset" sentinel Quick Scan
/// submits — renders as no label rather than "Empty". Together with
/// [receiptStatusColor] this is the single owner of receipt-status
/// presentation on mobile; keep the two exhaustive over the same values.
String receiptStatusLabel(ReceiptStatus status) {
  switch (status) {
    case ReceiptStatus.OPEN:
      return "Open";
    case ReceiptStatus.NEEDS_ATTENTION:
      return "Needs Attention";
    case ReceiptStatus.RESOLVED:
      return "Resolved";
    case ReceiptStatus.DRAFT:
      return "Draft";
    case ReceiptStatus.DECLINED:
      return "Declined";
    case ReceiptStatus.empty:
      return "";
    default:
      // Degrade instead of throwing, which is what both switches used to do.
      // That arm was already reachable via ReceiptStatus.empty (the search
      // screen builds a Receipt from a SearchResult whose receiptStatus is
      // nullable), and it is what keeps a status added to the API after this
      // build ships from taking the whole receipt list down.
      return titleCaseStatusName(status.name);
  }
}

/// SNAKE_CASE -> Title Case, the fallback [receiptStatusLabel] uses for a status
/// this build predates. Split out because a `ReceiptStatus` value the generated
/// enum does not declare cannot be constructed from a test, so this is the only
/// way to pin the rule.
String titleCaseStatusName(String name) {
  return name
      .split("_")
      .where((word) => word.isNotEmpty)
      .map((word) => word[0].toUpperCase() + word.substring(1).toLowerCase())
      .join(" ");
}

/// Tint for a receipt status. Every value is a pale tint, because
/// `ListItemTrailingStatus` paints its label in the default dark
/// `onBackground` and `ListItemColorBlock` sits beside dark text too.
Color receiptStatusColor(ReceiptStatus status) {
  switch (status) {
    case ReceiptStatus.OPEN:
      return const Color.fromRGBO(255, 250, 205, 1);
    case ReceiptStatus.NEEDS_ATTENTION:
      return warningAmber;
    case ReceiptStatus.RESOLVED:
      return successGreen;
    case ReceiptStatus.DRAFT:
      return neutralStatusGrey;
    case ReceiptStatus.DECLINED:
      // The red NEEDS_ATTENTION gave up.
      return errorRed;
    default:
      return neutralStatusGrey;
  }
}

String getTitleText(WranglerFormState formState, String receiptName) {
  return "${getFormStateHeader(formState)} $receiptName Receipt";
}

String getLeadingArrowRedirect(String groupId) {
  return "/groups/$groupId/receipts";
}

double getImagePreviewWidth(BuildContext context) {
  return MediaQuery.of(context).size.width;
}

double getImagePreviewHeight(BuildContext context) {
  return MediaQuery.of(context).size.height * .5;
}

EdgeInsets getImageDataPadding() {
  return const EdgeInsets.all(26);
}

Uint8List getBytesFromEncodedImage(String encodedImage) {
  var base64Image = encodedImage.split(",").last;
  if (base64Image == "") {
    return Uint8List(0);
  }

  var bytes = base64Decode(base64Image);
  return bytes;
}

ReceiptPagedRequestFilterBuilder dashboardConfigurationToFilter(
    BuiltMap<String, JsonObject?>? configuration) {
  var filter = (ReceiptPagedRequestFilterBuilder());
  if (configuration == null) {
    return filter;
  }

  configuration.forEach((key, value) {
    if (value.toString().length == 0) {
      return;
    }

    var valueMap =
        (value?.asMap as Map<String, dynamic>) ?? Map<String, dynamic>();
    var filterObject = JsonObject({
      "operation": valueMap["operation"] ?? "",
      "value": valueMap["value"] ?? null,
    });

    switch (key) {
      case 'date':
        filter.date = filterObject;
        break;
      case 'amount':
        filter.amount = filterObject;
        break;
      case 'name':
        filter.name = filterObject;
        break;
      case 'paidBy':
        filter.paidBy = filterObject;
        break;
      case 'categories':
        filter.categories = filterObject;
        break;
      case 'tags':
        filter.tags = filterObject;
        break;
      case 'status':
        filter.status = filterObject;
        break;
      case 'resolvedDate':
        filter.resolvedDate = filterObject;
        break;
      case 'createdAt':
        filter.createdAt = filterObject;
        break;
    }
  });

  return filter;
}
