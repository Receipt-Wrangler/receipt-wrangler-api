import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/unrenderable_file_placeholder.dart';

/// The first bytes of a PDF. Undecodable as an image, which is the whole point:
/// the file source legitimately accepts PDFs and the backend converts them
/// server-side, so these bytes reach `Image.memory` on a *successful* upload.
final _pdfBytes = Uint8List.fromList('%PDF-1.4\n%âãÏÓ\n'.codeUnits);

void main() {
  Future<void> pumpPlaceholder(WidgetTester tester, {String? filename}) async {
    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: UnrenderableFilePlaceholder(filename: filename),
      ),
    ));
    await tester.pump();
  }

  group('UnrenderableFilePlaceholder', () {
    testWidgets('names the file and uses the PDF icon for a .pdf',
        (tester) async {
      await pumpPlaceholder(tester, filename: 'dinner-receipt.pdf');

      expect(find.text('dinner-receipt.pdf'), findsOneWidget);
      expect(find.byIcon(Icons.picture_as_pdf), findsOneWidget);
      expect(find.byIcon(Icons.insert_drive_file), findsNothing);
    });

    testWidgets('matches the extension case-insensitively', (tester) async {
      await pumpPlaceholder(tester, filename: 'SCAN.PDF');

      expect(find.byIcon(Icons.picture_as_pdf), findsOneWidget);
    });

    testWidgets('falls back to the generic icon for anything else',
        (tester) async {
      await pumpPlaceholder(tester, filename: 'notes.txt');

      expect(find.byIcon(Icons.insert_drive_file), findsOneWidget);
      expect(find.byIcon(Icons.picture_as_pdf), findsNothing);
    });

    testWidgets('says something useful when the filename is unknown',
        (tester) async {
      await pumpPlaceholder(tester);

      expect(find.text('Preview unavailable'), findsOneWidget);
    });
  });

  group('unrenderableFileErrorBuilder', () {
    testWidgets('stands in for bytes Image.memory cannot decode',
        (tester) async {
      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: Image.memory(
            _pdfBytes,
            errorBuilder:
                unrenderableFileErrorBuilder(filename: 'receipt.pdf'),
          ),
        ),
      ));
      // The decode failure arrives asynchronously, so the placeholder is not up
      // on the first frame.
      await tester.pumpAndSettle();

      expect(find.byType(UnrenderableFilePlaceholder), findsOneWidget);
      expect(find.text('receipt.pdf'), findsOneWidget);
      expect(tester.takeException(), isNull,
          reason: 'the builder handles the decode failure rather than letting '
              'it reach the framework');
    });
  });
}
