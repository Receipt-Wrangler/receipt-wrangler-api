import 'dart:ui';

import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart';
import 'package:receipt_wrangler_mobile/constants/colors.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';

/// `receiptStatusLabel` / `receiptStatusColor` (lib/utils/receipts.dart) are the
/// single owner of receipt-status presentation on mobile, and had no coverage
/// before DECLINED was added. They are tested here rather than through
/// ReceiptListItem because that widget's trailing pill is a fixed 100px wide, so
/// a longer label ("Needs Attention") reports a render overflow that has nothing
/// to do with the mapping under test.
void main() {
  group('receiptStatusLabel', () {
    const cases = <ReceiptStatus, String>{
      ReceiptStatus.OPEN: 'Open',
      ReceiptStatus.NEEDS_ATTENTION: 'Needs Attention',
      ReceiptStatus.RESOLVED: 'Resolved',
      ReceiptStatus.DRAFT: 'Draft',
      ReceiptStatus.DECLINED: 'Declined',
    };

    cases.forEach((status, label) {
      test('labels ${status.name} as "$label"', () {
        expect(receiptStatusLabel(status), label);
      });
    });

    // The "unset" sentinel Quick Scan submits. It reaches the list through the
    // search screen, which builds a Receipt from a SearchResult whose
    // receiptStatus is nullable — and used to throw here.
    test('renders the empty sentinel as no label', () {
      expect(receiptStatusLabel(ReceiptStatus.empty), '');
    });

    test('covers every value the generated enum declares', () {
      for (final status in ReceiptStatus.values) {
        expect(
          () => receiptStatusLabel(status),
          returnsNormally,
          reason: '${status.name} has no label',
        );
      }
    });
  });

  // The fallback for a status added to the API after this build ships. A
  // ReceiptStatus the generated enum does not declare cannot be constructed,
  // so the rule is pinned through the helper the default arm delegates to.
  group('titleCaseStatusName', () {
    test('title-cases a SNAKE_CASE status name', () {
      expect(titleCaseStatusName('AWAITING_REVIEW'), 'Awaiting Review');
    });

    test('handles a single word', () {
      expect(titleCaseStatusName('DECLINED'), 'Declined');
    });

    test('does not choke on repeated or trailing separators', () {
      expect(titleCaseStatusName('A__B_'), 'A B');
    });
  });

  group('receiptStatusColor', () {
    // DECLINED owns the red NEEDS_ATTENTION gave up.
    const cases = <ReceiptStatus, Color>{
      ReceiptStatus.OPEN: Color.fromRGBO(255, 250, 205, 1),
      ReceiptStatus.NEEDS_ATTENTION: warningAmber,
      ReceiptStatus.RESOLVED: successGreen,
      ReceiptStatus.DRAFT: neutralStatusGrey,
      ReceiptStatus.DECLINED: errorRed,
    };

    cases.forEach((status, color) {
      test('tints ${status.name}', () {
        expect(receiptStatusColor(status), color);
      });
    });

    test('NEEDS_ATTENTION and DECLINED are visually distinct', () {
      expect(
        receiptStatusColor(ReceiptStatus.NEEDS_ATTENTION),
        isNot(receiptStatusColor(ReceiptStatus.DECLINED)),
      );
    });

    test('covers every value the generated enum declares', () {
      for (final status in ReceiptStatus.values) {
        expect(
          () => receiptStatusColor(status),
          returnsNormally,
          reason: '${status.name} has no tint',
        );
      }
    });
  });
}
