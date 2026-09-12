import 'package:flutter_test/flutter_test.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';

import '../helpers/receipt_form_test_helpers.dart';

// GroupModel.soleGroupId backs the "a picker with one option is not a choice"
// rule in the receipt form and Quick Scan. The load-bearing part is that it
// counts SELECTABLE groups: every user also carries the synthetic "All" group,
// which is never a receipt target and must not make the count two.
void main() {
  GroupModel modelWith(List<int> ids, {int? allGroupId}) => GroupModel()
    ..setGroups([
      if (allGroupId != null)
        buildGroup(id: allGroupId, name: 'All Groups', isAllGroup: true),
      for (final id in ids) buildGroup(id: id, name: 'Group $id'),
    ]);

  test('returns the id when the user belongs to exactly one group', () {
    expect(modelWith([7]).soleGroupId, 7);
  });

  test('ignores the synthetic All group when counting', () {
    expect(modelWith([7], allGroupId: 1).soleGroupId, 7);
  });

  test('returns null when the user belongs to more than one group', () {
    expect(modelWith([7, 8], allGroupId: 1).soleGroupId, isNull);
  });

  test('returns null when only the All group is present', () {
    expect(modelWith(const [], allGroupId: 1).soleGroupId, isNull);
  });

  test('returns null when there are no groups at all', () {
    expect(modelWith(const []).soleGroupId, isNull);
  });
}
