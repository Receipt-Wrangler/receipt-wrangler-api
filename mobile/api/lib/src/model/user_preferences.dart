//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/base_model.dart';
import 'package:openapi/src/model/receipt_status.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/user_shortcut.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'user_preferences.g.dart';

/// UserPreferences
///
/// Properties:
/// * [id] - User preferences id
/// * [createdAt] 
/// * [createdBy] 
/// * [createdByString] - Created by entity's name
/// * [updatedAt] 
/// * [userId] - User foreign key
/// * [quickScanDefaultGroupId] - Group foreign key
/// * [quickScanDefaultPaidById] - User foreign key
/// * [quickScanDefaultStatus] - Default quick scan status
/// * [showLargeImagePreviews] - Whether to show large image previews
/// * [userShortcuts] 
@BuiltValue()
abstract class UserPreferences implements BaseModel, Built<UserPreferences, UserPreferencesBuilder> {
  /// Default quick scan status
  @BuiltValueField(wireName: r'quickScanDefaultStatus')
  ReceiptStatus? get quickScanDefaultStatus;
  // enum quickScanDefaultStatusEnum {  OPEN,  NEEDS_ATTENTION,  RESOLVED,  DRAFT,  DECLINED,  ,  };

  @BuiltValueField(wireName: r'userShortcuts')
  BuiltList<UserShortcut>? get userShortcuts;

  /// Whether to show large image previews
  @BuiltValueField(wireName: r'showLargeImagePreviews')
  bool? get showLargeImagePreviews;

  /// User foreign key
  @BuiltValueField(wireName: r'userId')
  int get userId;

  /// Group foreign key
  @BuiltValueField(wireName: r'quickScanDefaultGroupId')
  int? get quickScanDefaultGroupId;

  /// User foreign key
  @BuiltValueField(wireName: r'quickScanDefaultPaidById')
  int? get quickScanDefaultPaidById;

  UserPreferences._();

  factory UserPreferences([void updates(UserPreferencesBuilder b)]) = _$UserPreferences;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(UserPreferencesBuilder b) => b
      ..quickScanDefaultStatus = ReceiptStatus.OPEN
      ..createdBy = 0
      ..quickScanDefaultGroupId = 0
      ..quickScanDefaultPaidById = 0
      ..createdByString = ''
      ..updatedAt = '';

  @BuiltValueSerializer(custom: true)
  static Serializer<UserPreferences> get serializer => _$UserPreferencesSerializer();
}

class _$UserPreferencesSerializer implements PrimitiveSerializer<UserPreferences> {
  @override
  final Iterable<Type> types = const [UserPreferences, _$UserPreferences];

  @override
  final String wireName = r'UserPreferences';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    UserPreferences object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'createdAt';
    yield serializers.serialize(
      object.createdAt,
      specifiedType: const FullType(String),
    );
    if (object.quickScanDefaultStatus != null) {
      yield r'quickScanDefaultStatus';
      yield serializers.serialize(
        object.quickScanDefaultStatus,
        specifiedType: const FullType(ReceiptStatus),
      );
    }
    if (object.createdBy != null) {
      yield r'createdBy';
      yield serializers.serialize(
        object.createdBy,
        specifiedType: const FullType(int),
      );
    }
    if (object.userShortcuts != null) {
      yield r'userShortcuts';
      yield serializers.serialize(
        object.userShortcuts,
        specifiedType: const FullType(BuiltList, [FullType(UserShortcut)]),
      );
    }
    if (object.showLargeImagePreviews != null) {
      yield r'showLargeImagePreviews';
      yield serializers.serialize(
        object.showLargeImagePreviews,
        specifiedType: const FullType(bool),
      );
    }
    yield r'id';
    yield serializers.serialize(
      object.id,
      specifiedType: const FullType(int),
    );
    yield r'userId';
    yield serializers.serialize(
      object.userId,
      specifiedType: const FullType(int),
    );
    if (object.quickScanDefaultGroupId != null) {
      yield r'quickScanDefaultGroupId';
      yield serializers.serialize(
        object.quickScanDefaultGroupId,
        specifiedType: const FullType(int),
      );
    }
    if (object.quickScanDefaultPaidById != null) {
      yield r'quickScanDefaultPaidById';
      yield serializers.serialize(
        object.quickScanDefaultPaidById,
        specifiedType: const FullType(int),
      );
    }
    if (object.createdByString != null) {
      yield r'createdByString';
      yield serializers.serialize(
        object.createdByString,
        specifiedType: const FullType(String),
      );
    }
    if (object.updatedAt != null) {
      yield r'updatedAt';
      yield serializers.serialize(
        object.updatedAt,
        specifiedType: const FullType(String),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    UserPreferences object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required UserPreferencesBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'createdAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.createdAt = valueDes;
          break;
        case r'quickScanDefaultStatus':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(ReceiptStatus),
          ) as ReceiptStatus;
          result.quickScanDefaultStatus = valueDes;
          break;
        case r'createdBy':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.createdBy = valueDes;
          break;
        case r'userShortcuts':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(UserShortcut)]),
          ) as BuiltList<UserShortcut>;
          result.userShortcuts.replace(valueDes);
          break;
        case r'showLargeImagePreviews':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.showLargeImagePreviews = valueDes;
          break;
        case r'id':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.id = valueDes;
          break;
        case r'userId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.userId = valueDes;
          break;
        case r'quickScanDefaultGroupId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.quickScanDefaultGroupId = valueDes;
          break;
        case r'quickScanDefaultPaidById':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.quickScanDefaultPaidById = valueDes;
          break;
        case r'createdByString':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.createdByString = valueDes;
          break;
        case r'updatedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.updatedAt = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  UserPreferences deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = UserPreferencesBuilder();
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

