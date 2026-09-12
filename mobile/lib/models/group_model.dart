import 'package:flutter/material.dart';
import 'package:openapi/openapi.dart';

class GroupModel extends ChangeNotifier {
  List<Group> _groups = [];

  List<Group> get groups => _groups;

  List<Group> get groupsWithoutAllGroup =>
      _groups.where((group) => !group.isAllGroup).toList();

  /// The user's only real group, when they belong to exactly one -- a picker with
  /// a single option is not a choice, so the receipt form and quick scan seed it
  /// rather than making the user pick it. Null when they have several (or none):
  /// built off [groupsWithoutAllGroup] so the auto-selected set and the set the
  /// pickers offer cannot drift apart.
  int? get soleGroupId {
    final selectable = groupsWithoutAllGroup;
    return selectable.length == 1 ? selectable.first.id : null;
  }

  void setGroups(List<Group> groups) {
    _groups = groups;

    notifyListeners();
  }

  Group? getGroupById(String id) {
    try {
      return _groups.firstWhere((group) => group.id == int.tryParse(id));
    } catch (e) {
      return null;
    }
  }

  GroupReceiptSettings? getGroupReceiptSettings(int groupId) {
    if (groupId == 0) {
      return null;
    }

    return getGroupById(groupId.toString())?.groupReceiptSettings;
  }
}
