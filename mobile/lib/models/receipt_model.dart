import 'package:flutter/material.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_form_builder/flutter_form_builder.dart';
import 'package:infinite_carousel/infinite_carousel.dart';
import 'package:openapi/openapi.dart';
import 'package:receipt_wrangler_mobile/interfaces/form_item.dart';
import 'package:receipt_wrangler_mobile/interfaces/upload_multipart_file_data.dart';
import 'package:receipt_wrangler_mobile/utils/receipts.dart';
import 'package:rxdart/rxdart.dart';

class ReceiptModel extends ChangeNotifier {
  Receipt _receipt = getDefaultReceipt();

  Receipt get receipt => _receipt;

  Receipt _modifiedReceipt = getDefaultReceipt();

  Receipt get modifiedReceipt => _modifiedReceipt;

  /// Custom field ids the receipt form attached on the user's behalf from the
  /// selected group's default set, as opposed to ones the user added by hand.
  /// Only the former are candidates for removal when the group changes.
  ///
  /// Provenance lives here rather than in the form's State because this model
  /// outlives the screen: the app bar menu's Edit entry is a plain `go`
  /// (`receipt_app_bar_action_builder.dart`), so a view -> edit navigation
  /// remounts the form against this same [modifiedReceipt]. A fresh State could
  /// only guess, and would reclaim -- then silently drop on the next group
  /// change -- a default the user had removed and re-added themselves.
  ///
  /// Cleared wherever the working copy is replaced ([setReceipt], [resetModel])
  /// so the two can never disagree. Deliberately does not notify: nothing
  /// renders from it.
  final Set<int> _autoAppliedCustomFieldIds = {};

  Set<int> get autoAppliedCustomFieldIds => _autoAppliedCustomFieldIds;

  List<Comment> _comments = [];

  List<Comment> get comments => _comments;

  List<FormItem> _items = [];

  List<FormItem> get items => _items;

  BehaviorSubject<List<FileDataView?>> _imageBehaviorSubject =
      BehaviorSubject<List<FileDataView?>>.seeded([]);

  BehaviorSubject<List<FileDataView?>> get imageBehaviorSubject =>
      _imageBehaviorSubject;

  BehaviorSubject<List<UploadMultipartFileData>>
      _imagesToUploadBehaviorSubject =
      BehaviorSubject<List<UploadMultipartFileData>>.seeded([]);

  BehaviorSubject<List<UploadMultipartFileData>>
      get imagesToUploadBehaviorSubject => _imagesToUploadBehaviorSubject;

  InfiniteScrollController _infiniteScrollController =
      InfiniteScrollController();

  InfiniteScrollController get infiniteScrollController =>
      _infiniteScrollController;

  var _receiptFormKey = GlobalKey<FormBuilderState>();

  GlobalKey<FormBuilderState> get receiptFormKey => _receiptFormKey;

  var _quickActionsFormKey = GlobalKey<FormBuilderState>();

  GlobalKey<FormBuilderState> get quickActionsFormKey => _quickActionsFormKey;

  void setReceipt(Receipt receipt, bool notify) {
    // Only regenerate the form key when we're loading a *different* receipt
    // (or moving from the default-id-0 placeholder to a real one).
    // `ReceiptFormScreen.build()` (receipt_form_screen.dart:88) calls
    // setReceipt unconditionally on every screen rebuild after the
    // FutureBuilder resolves. If we regen the GlobalKey every time, the
    // FormBuilder gets a new key on each parent rebuild and Flutter tears
    // down + remounts the entire form -- which makes the submit handler's
    // `currentState` go null between taps ("Null check operator used on a
    // null value" at receipt_bottom_sheet_builder.dart:389) and breaks
    // dropdown interactions whose state is wiped mid-tap. Regenerating
    // the key is only needed when the receipt identity actually changes,
    // so the FormBuilderField initialValues get re-read for the new data.
    final isNewReceipt = _receipt.id != receipt.id;

    _receipt = receipt;

    _modifiedReceipt = receipt;

    _autoAppliedCustomFieldIds.clear();

    _comments = (receipt.comments)?.toList() ?? [];

    _items = FormItem.fromItems((receipt.receiptItems)?.toList() ?? []);

    _imageBehaviorSubject = BehaviorSubject<List<FileDataView?>>.seeded([]);

    _imagesToUploadBehaviorSubject =
        BehaviorSubject<List<UploadMultipartFileData>>.seeded([]);

    if (isNewReceipt) {
      _receiptFormKey = GlobalKey<FormBuilderState>();
    }

    if (notify) {
      notifyListeners();
    }
  }

  void setComments(List<Comment> comments) {
    _comments = comments;
    notifyListeners();
  }

  void setItems(List<FormItem> items) {
    _items = items;
    notifyListeners();
  }

  void setModifiedReceipt(Receipt receipt) {
    _modifiedReceipt = receipt;
    notifyListeners();
  }

  void resetQuickActionsFormKey() {
    _quickActionsFormKey = GlobalKey<FormBuilderState>();
  }

  void resetModel() {
    _receipt = getDefaultReceipt();
    _modifiedReceipt = getDefaultReceipt();
    _autoAppliedCustomFieldIds.clear();
    _comments = [];
    _items = [];
    _imageBehaviorSubject = BehaviorSubject<List<FileDataView?>>.seeded([]);
    _imagesToUploadBehaviorSubject =
        BehaviorSubject<List<UploadMultipartFileData>>.seeded([]);
    _infiniteScrollController = InfiniteScrollController();
    _receiptFormKey = GlobalKey<FormBuilderState>();
  }
}
