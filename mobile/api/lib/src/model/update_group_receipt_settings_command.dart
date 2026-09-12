//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/receipt_status.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/quick_scan_default_paid_by_type.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'update_group_receipt_settings_command.g.dart';

/// UpdateGroupReceiptSettingsCommand
///
/// Properties:
/// * [hideImages] - Hide receipt images
/// * [hideReceiptCategories] - Hide receipt categories
/// * [hideReceiptTags] - Hide receipt tags
/// * [hideItemCategories] - Hide receipt item categories
/// * [hideItemTags] - Hide receipt item tags
/// * [hideComments] - Hide receipt comments
/// * [hideShareCategories] - Hide share categories
/// * [hideShareTags] - Hide share tags
/// * [quickScanPaidByEnabled] - Show the paid by field in quick scan
/// * [quickScanPaidByRequired] - Require the paid by field in quick scan
/// * [quickScanDefaultPaidByType] 
/// * [quickScanDefaultPaidById] - Default paid by user id when paid by is optional and type is USER
/// * [quickScanStatusEnabled] - Show the status field in quick scan
/// * [quickScanStatusRequired] - Require the status field in quick scan
/// * [quickScanDefaultStatus] 
/// * [quickScanCategoriesEnabled] - Show the categories field in quick scan
/// * [quickScanCategoriesRequired] - Require the categories field in quick scan
/// * [quickScanTagsEnabled] - Show the tags field in quick scan
/// * [quickScanTagsRequired] - Require the tags field in quick scan
/// * [quickScanCommentEnabled] - Show the comment field in quick scan
/// * [quickScanCommentRequired] - Require the comment field in quick scan
/// * [defaultCustomFieldIds] - Custom field ids to pre-add to every receipt created for this group. OMIT the key to leave the configured set unchanged (clients that hide this section, e.g. for a user without app.custom-fields.read, must omit it); send an empty array to clear it. Requires app.custom-fields.read - a caller without it gets a 403.
/// * [applyDefaultCustomFieldsOnIngest] - Also attach the group's default custom fields to receipts the SERVER creates (quick scan, email integration). OMIT the key to leave the stored value unchanged.
@BuiltValue()
abstract class UpdateGroupReceiptSettingsCommand implements Built<UpdateGroupReceiptSettingsCommand, UpdateGroupReceiptSettingsCommandBuilder> {
  /// Hide receipt images
  @BuiltValueField(wireName: r'hideImages')
  bool? get hideImages;

  /// Hide receipt categories
  @BuiltValueField(wireName: r'hideReceiptCategories')
  bool? get hideReceiptCategories;

  /// Hide receipt tags
  @BuiltValueField(wireName: r'hideReceiptTags')
  bool? get hideReceiptTags;

  /// Hide receipt item categories
  @BuiltValueField(wireName: r'hideItemCategories')
  bool? get hideItemCategories;

  /// Hide receipt item tags
  @BuiltValueField(wireName: r'hideItemTags')
  bool? get hideItemTags;

  /// Hide receipt comments
  @BuiltValueField(wireName: r'hideComments')
  bool? get hideComments;

  /// Hide share categories
  @BuiltValueField(wireName: r'hideShareCategories')
  bool? get hideShareCategories;

  /// Hide share tags
  @BuiltValueField(wireName: r'hideShareTags')
  bool? get hideShareTags;

  /// Show the paid by field in quick scan
  @BuiltValueField(wireName: r'quickScanPaidByEnabled')
  bool? get quickScanPaidByEnabled;

  /// Require the paid by field in quick scan
  @BuiltValueField(wireName: r'quickScanPaidByRequired')
  bool? get quickScanPaidByRequired;

  @BuiltValueField(wireName: r'quickScanDefaultPaidByType')
  QuickScanDefaultPaidByType? get quickScanDefaultPaidByType;
  // enum quickScanDefaultPaidByTypeEnum {  UPLOADER,  USER,  ,  };

  /// Default paid by user id when paid by is optional and type is USER
  @BuiltValueField(wireName: r'quickScanDefaultPaidById')
  int? get quickScanDefaultPaidById;

  /// Show the status field in quick scan
  @BuiltValueField(wireName: r'quickScanStatusEnabled')
  bool? get quickScanStatusEnabled;

  /// Require the status field in quick scan
  @BuiltValueField(wireName: r'quickScanStatusRequired')
  bool? get quickScanStatusRequired;

  @BuiltValueField(wireName: r'quickScanDefaultStatus')
  ReceiptStatus? get quickScanDefaultStatus;
  // enum quickScanDefaultStatusEnum {  OPEN,  NEEDS_ATTENTION,  RESOLVED,  DRAFT,  DECLINED,  ,  };

  /// Show the categories field in quick scan
  @BuiltValueField(wireName: r'quickScanCategoriesEnabled')
  bool? get quickScanCategoriesEnabled;

  /// Require the categories field in quick scan
  @BuiltValueField(wireName: r'quickScanCategoriesRequired')
  bool? get quickScanCategoriesRequired;

  /// Show the tags field in quick scan
  @BuiltValueField(wireName: r'quickScanTagsEnabled')
  bool? get quickScanTagsEnabled;

  /// Require the tags field in quick scan
  @BuiltValueField(wireName: r'quickScanTagsRequired')
  bool? get quickScanTagsRequired;

  /// Show the comment field in quick scan
  @BuiltValueField(wireName: r'quickScanCommentEnabled')
  bool? get quickScanCommentEnabled;

  /// Require the comment field in quick scan
  @BuiltValueField(wireName: r'quickScanCommentRequired')
  bool? get quickScanCommentRequired;

  /// Custom field ids to pre-add to every receipt created for this group. OMIT the key to leave the configured set unchanged (clients that hide this section, e.g. for a user without app.custom-fields.read, must omit it); send an empty array to clear it. Requires app.custom-fields.read - a caller without it gets a 403.
  @BuiltValueField(wireName: r'defaultCustomFieldIds')
  BuiltList<int>? get defaultCustomFieldIds;

  /// Also attach the group's default custom fields to receipts the SERVER creates (quick scan, email integration). OMIT the key to leave the stored value unchanged.
  @BuiltValueField(wireName: r'applyDefaultCustomFieldsOnIngest')
  bool? get applyDefaultCustomFieldsOnIngest;

  UpdateGroupReceiptSettingsCommand._();

  factory UpdateGroupReceiptSettingsCommand([void updates(UpdateGroupReceiptSettingsCommandBuilder b)]) = _$UpdateGroupReceiptSettingsCommand;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(UpdateGroupReceiptSettingsCommandBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<UpdateGroupReceiptSettingsCommand> get serializer => _$UpdateGroupReceiptSettingsCommandSerializer();
}

class _$UpdateGroupReceiptSettingsCommandSerializer implements PrimitiveSerializer<UpdateGroupReceiptSettingsCommand> {
  @override
  final Iterable<Type> types = const [UpdateGroupReceiptSettingsCommand, _$UpdateGroupReceiptSettingsCommand];

  @override
  final String wireName = r'UpdateGroupReceiptSettingsCommand';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    UpdateGroupReceiptSettingsCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    if (object.hideImages != null) {
      yield r'hideImages';
      yield serializers.serialize(
        object.hideImages,
        specifiedType: const FullType(bool),
      );
    }
    if (object.hideReceiptCategories != null) {
      yield r'hideReceiptCategories';
      yield serializers.serialize(
        object.hideReceiptCategories,
        specifiedType: const FullType(bool),
      );
    }
    if (object.hideReceiptTags != null) {
      yield r'hideReceiptTags';
      yield serializers.serialize(
        object.hideReceiptTags,
        specifiedType: const FullType(bool),
      );
    }
    if (object.hideItemCategories != null) {
      yield r'hideItemCategories';
      yield serializers.serialize(
        object.hideItemCategories,
        specifiedType: const FullType(bool),
      );
    }
    if (object.hideItemTags != null) {
      yield r'hideItemTags';
      yield serializers.serialize(
        object.hideItemTags,
        specifiedType: const FullType(bool),
      );
    }
    if (object.hideComments != null) {
      yield r'hideComments';
      yield serializers.serialize(
        object.hideComments,
        specifiedType: const FullType(bool),
      );
    }
    if (object.hideShareCategories != null) {
      yield r'hideShareCategories';
      yield serializers.serialize(
        object.hideShareCategories,
        specifiedType: const FullType(bool),
      );
    }
    if (object.hideShareTags != null) {
      yield r'hideShareTags';
      yield serializers.serialize(
        object.hideShareTags,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanPaidByEnabled != null) {
      yield r'quickScanPaidByEnabled';
      yield serializers.serialize(
        object.quickScanPaidByEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanPaidByRequired != null) {
      yield r'quickScanPaidByRequired';
      yield serializers.serialize(
        object.quickScanPaidByRequired,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanDefaultPaidByType != null) {
      yield r'quickScanDefaultPaidByType';
      yield serializers.serialize(
        object.quickScanDefaultPaidByType,
        specifiedType: const FullType(QuickScanDefaultPaidByType),
      );
    }
    if (object.quickScanDefaultPaidById != null) {
      yield r'quickScanDefaultPaidById';
      yield serializers.serialize(
        object.quickScanDefaultPaidById,
        specifiedType: const FullType(int),
      );
    }
    if (object.quickScanStatusEnabled != null) {
      yield r'quickScanStatusEnabled';
      yield serializers.serialize(
        object.quickScanStatusEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanStatusRequired != null) {
      yield r'quickScanStatusRequired';
      yield serializers.serialize(
        object.quickScanStatusRequired,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanDefaultStatus != null) {
      yield r'quickScanDefaultStatus';
      yield serializers.serialize(
        object.quickScanDefaultStatus,
        specifiedType: const FullType(ReceiptStatus),
      );
    }
    if (object.quickScanCategoriesEnabled != null) {
      yield r'quickScanCategoriesEnabled';
      yield serializers.serialize(
        object.quickScanCategoriesEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanCategoriesRequired != null) {
      yield r'quickScanCategoriesRequired';
      yield serializers.serialize(
        object.quickScanCategoriesRequired,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanTagsEnabled != null) {
      yield r'quickScanTagsEnabled';
      yield serializers.serialize(
        object.quickScanTagsEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanTagsRequired != null) {
      yield r'quickScanTagsRequired';
      yield serializers.serialize(
        object.quickScanTagsRequired,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanCommentEnabled != null) {
      yield r'quickScanCommentEnabled';
      yield serializers.serialize(
        object.quickScanCommentEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.quickScanCommentRequired != null) {
      yield r'quickScanCommentRequired';
      yield serializers.serialize(
        object.quickScanCommentRequired,
        specifiedType: const FullType(bool),
      );
    }
    if (object.defaultCustomFieldIds != null) {
      yield r'defaultCustomFieldIds';
      yield serializers.serialize(
        object.defaultCustomFieldIds,
        specifiedType: const FullType(BuiltList, [FullType(int)]),
      );
    }
    if (object.applyDefaultCustomFieldsOnIngest != null) {
      yield r'applyDefaultCustomFieldsOnIngest';
      yield serializers.serialize(
        object.applyDefaultCustomFieldsOnIngest,
        specifiedType: const FullType(bool),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    UpdateGroupReceiptSettingsCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required UpdateGroupReceiptSettingsCommandBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'hideImages':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideImages = valueDes;
          break;
        case r'hideReceiptCategories':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideReceiptCategories = valueDes;
          break;
        case r'hideReceiptTags':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideReceiptTags = valueDes;
          break;
        case r'hideItemCategories':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideItemCategories = valueDes;
          break;
        case r'hideItemTags':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideItemTags = valueDes;
          break;
        case r'hideComments':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideComments = valueDes;
          break;
        case r'hideShareCategories':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideShareCategories = valueDes;
          break;
        case r'hideShareTags':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hideShareTags = valueDes;
          break;
        case r'quickScanPaidByEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanPaidByEnabled = valueDes;
          break;
        case r'quickScanPaidByRequired':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanPaidByRequired = valueDes;
          break;
        case r'quickScanDefaultPaidByType':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(QuickScanDefaultPaidByType),
          ) as QuickScanDefaultPaidByType;
          result.quickScanDefaultPaidByType = valueDes;
          break;
        case r'quickScanDefaultPaidById':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.quickScanDefaultPaidById = valueDes;
          break;
        case r'quickScanStatusEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanStatusEnabled = valueDes;
          break;
        case r'quickScanStatusRequired':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanStatusRequired = valueDes;
          break;
        case r'quickScanDefaultStatus':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(ReceiptStatus),
          ) as ReceiptStatus;
          result.quickScanDefaultStatus = valueDes;
          break;
        case r'quickScanCategoriesEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanCategoriesEnabled = valueDes;
          break;
        case r'quickScanCategoriesRequired':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanCategoriesRequired = valueDes;
          break;
        case r'quickScanTagsEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanTagsEnabled = valueDes;
          break;
        case r'quickScanTagsRequired':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanTagsRequired = valueDes;
          break;
        case r'quickScanCommentEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanCommentEnabled = valueDes;
          break;
        case r'quickScanCommentRequired':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.quickScanCommentRequired = valueDes;
          break;
        case r'defaultCustomFieldIds':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(int)]),
          ) as BuiltList<int>;
          result.defaultCustomFieldIds.replace(valueDes);
          break;
        case r'applyDefaultCustomFieldsOnIngest':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.applyDefaultCustomFieldsOnIngest = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  UpdateGroupReceiptSettingsCommand deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = UpdateGroupReceiptSettingsCommandBuilder();
    final serializedList = (serialized as Iterable<Object?>).toList();
    final unhandled = <Object?>[];
    _deserializeProperties(
      serializers,
      serialized,
      specifiedType: specifiedType,
      serializedList: serializedList,
      unhandled: unhandled,
      result: result,
    );
    return result.build();
  }
}

