import 'package:flutter_test/flutter_test.dart';
import 'package:openapi/openapi.dart';
import 'package:receipt_wrangler_mobile/constants/colors.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';

/// Guards the API boundary for `ReceiptStatus`, the sibling of the two
/// production login outages that `app_data_permission_ingest_test.dart` covers.
///
/// built_value renders an OpenAPI enum as a CLOSED set: `_$valueOf` used to end
/// in `default: throw ArgumentError(name)`, so a single unrecognized wire value
/// failed the *entire* enclosing payload -- the whole paged receipts response,
/// and login itself when the value rode on `AppData` via a group's
/// `emailDefaultReceiptStatus` or either `quickScanDefaultStatus`. That is the
/// mechanism by which adding DECLINED server-side would have broken every
/// already-released build.
///
/// `ReceiptStatus.empty` is annotated `fallback: true` (a hand-patch on the
/// generated `receipt_status.dart` -- see mobile/CLAUDE.md), which makes the
/// generated `_$valueOf` return it rather than throw. This suite exercises the
/// REAL generated deserializer, which is where the throw happened and which sits
/// upstream of the presentation fallback in `receiptStatusLabel`.
Map<String, Object?> _receiptJson(String status) => {
      'id': 1,
      'name': 'Test Receipt',
      'amount': '12.34',
      'date': '2026-01-01T00:00:00Z',
      'paidByUserId': 1,
      'groupId': 1,
      'status': status,
    };

void main() {
  group('unknown wire status', () {
    // The headline: a status this build predates no longer throws.
    test('deserializes to the empty fallback instead of throwing', () {
      expect(
        standardSerializers.deserializeWith(
            ReceiptStatus.serializer, 'AWAITING_REVIEW'),
        ReceiptStatus.empty,
      );
    });

    // The blast radius that actually mattered: one bad value used to take the
    // whole enclosing object down, not just the field.
    test('does not fail the enclosing Receipt payload', () {
      final receipt = standardSerializers.deserializeWith(
          Receipt.serializer, _receiptJson('AWAITING_REVIEW'));

      expect(receipt, isNotNull);
      expect(receipt!.name, 'Test Receipt');
      expect(receipt.status, ReceiptStatus.empty);
    });

    // Deserialization and presentation have to compose: the fallback member
    // must render as something harmless rather than crash the list item.
    test('renders through the presentation layer without throwing', () {
      final receipt = standardSerializers.deserializeWith(
          Receipt.serializer, _receiptJson('AWAITING_REVIEW'))!;

      expect(receiptStatusLabel(receipt.status), '');
      expect(receiptStatusColor(receipt.status), neutralStatusGrey);
    });
  });

  group('known wire statuses still round-trip exactly', () {
    // The fallback must not swallow a value the build DOES know -- that would
    // silently blank every status in the app.
    const wireValues = <String, ReceiptStatus>{
      'OPEN': ReceiptStatus.OPEN,
      'NEEDS_ATTENTION': ReceiptStatus.NEEDS_ATTENTION,
      'RESOLVED': ReceiptStatus.RESOLVED,
      'DRAFT': ReceiptStatus.DRAFT,
      'DECLINED': ReceiptStatus.DECLINED,
      '': ReceiptStatus.empty,
    };

    wireValues.forEach((wire, expected) {
      test('"$wire" deserializes to ${expected.name}', () {
        expect(
          standardSerializers.deserializeWith(ReceiptStatus.serializer, wire),
          expected,
        );
      });

      test('${expected.name} serializes back to "$wire"', () {
        expect(
          standardSerializers.serializeWith(ReceiptStatus.serializer, expected),
          wire,
        );
      });
    });
  });
}
