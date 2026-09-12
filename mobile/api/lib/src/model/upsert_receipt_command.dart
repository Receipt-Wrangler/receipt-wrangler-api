//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/upsert_item_command.dart';
import 'package:openapi/src/model/receipt_status.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/upsert_category_command.dart';
import 'package:openapi/src/model/upsert_comment_command.dart';
import 'package:openapi/src/model/upsert_custom_field_value_command.dart';
import 'package:openapi/src/model/upsert_tag_command.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'upsert_receipt_command.g.dart';

/// UpsertReceiptCommand
///
/// Properties:
/// * [name] - Receipt name
/// * [amount] - Receipt total amount
/// * [date] - Receipt date
/// * [groupId] - Group foreign key
/// * [paidByUserId] - User paid foreign key
/// * [status] 
/// * [categories] - Categories associated to receipt
/// * [tags] - Tags associated to receipt
/// * [receiptItems] - Items associated to receipt
/// * [comments] - Comments associated to receipt
/// * [customFields] - Custom fields associated to receipt
@BuiltValue()
abstract class UpsertReceiptCommand implements Built<UpsertReceiptCommand, UpsertReceiptCommandBuilder> {
  /// Receipt name
  @BuiltValueField(wireName: r'name')
  String get name;

  /// Receipt total amount
  @BuiltValueField(wireName: r'amount')
  String get amount;

  /// Receipt date
  @BuiltValueField(wireName: r'date')
  String get date;

  /// Group foreign key
  @BuiltValueField(wireName: r'groupId')
  int get groupId;

  /// User paid foreign key
  @BuiltValueField(wireName: r'paidByUserId')
  int get paidByUserId;

  @BuiltValueField(wireName: r'status')
  ReceiptStatus get status;
  // enum statusEnum {  OPEN,  NEEDS_ATTENTION,  RESOLVED,  DRAFT,  DECLINED,  ,  };

  /// Categories associated to receipt
  @BuiltValueField(wireName: r'categories')
  BuiltList<UpsertCategoryCommand>? get categories;

  /// Tags associated to receipt
  @BuiltValueField(wireName: r'tags')
  BuiltList<UpsertTagCommand>? get tags;

  /// Items associated to receipt
  @BuiltValueField(wireName: r'receiptItems')
  BuiltList<UpsertItemCommand>? get receiptItems;

  /// Comments associated to receipt
  @BuiltValueField(wireName: r'comments')
  BuiltList<UpsertCommentCommand>? get comments;

  /// Custom fields associated to receipt
  @BuiltValueField(wireName: r'customFields')
  BuiltList<UpsertCustomFieldValueCommand>? get customFields;

  UpsertReceiptCommand._();

  factory UpsertReceiptCommand([void updates(UpsertReceiptCommandBuilder b)]) = _$UpsertReceiptCommand;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(UpsertReceiptCommandBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<UpsertReceiptCommand> get serializer => _$UpsertReceiptCommandSerializer();
}

class _$UpsertReceiptCommandSerializer implements PrimitiveSerializer<UpsertReceiptCommand> {
  @override
  final Iterable<Type> types = const [UpsertReceiptCommand, _$UpsertReceiptCommand];

  @override
  final String wireName = r'UpsertReceiptCommand';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    UpsertReceiptCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'name';
    yield serializers.serialize(
      object.name,
      specifiedType: const FullType(String),
    );
    yield r'amount';
    yield serializers.serialize(
      object.amount,
      specifiedType: const FullType(String),
    );
    yield r'date';
    yield serializers.serialize(
      object.date,
      specifiedType: const FullType(String),
    );
    yield r'groupId';
    yield serializers.serialize(
      object.groupId,
      specifiedType: const FullType(int),
    );
    yield r'paidByUserId';
    yield serializers.serialize(
      object.paidByUserId,
      specifiedType: const FullType(int),
    );
    yield r'status';
    yield serializers.serialize(
      object.status,
      specifiedType: const FullType(ReceiptStatus),
    );
    if (object.categories != null) {
      yield r'categories';
      yield serializers.serialize(
        object.categories,
        specifiedType: const FullType(BuiltList, [FullType(UpsertCategoryCommand)]),
      );
    }
    if (object.tags != null) {
      yield r'tags';
      yield serializers.serialize(
        object.tags,
        specifiedType: const FullType(BuiltList, [FullType(UpsertTagCommand)]),
      );
    }
    if (object.receiptItems != null) {
      yield r'receiptItems';
      yield serializers.serialize(
        object.receiptItems,
        specifiedType: const FullType(BuiltList, [FullType(UpsertItemCommand)]),
      );
    }
    if (object.comments != null) {
      yield r'comments';
      yield serializers.serialize(
        object.comments,
        specifiedType: const FullType(BuiltList, [FullType(UpsertCommentCommand)]),
      );
    }
    if (object.customFields != null) {
      yield r'customFields';
      yield serializers.serialize(
        object.customFields,
        specifiedType: const FullType(BuiltList, [FullType(UpsertCustomFieldValueCommand)]),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    UpsertReceiptCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required UpsertReceiptCommandBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'name':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.name = valueDes;
          break;
        case r'amount':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.amount = valueDes;
          break;
        case r'date':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.date = valueDes;
          break;
        case r'groupId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.groupId = valueDes;
          break;
        case r'paidByUserId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.paidByUserId = valueDes;
          break;
        case r'status':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(ReceiptStatus),
          ) as ReceiptStatus;
          result.status = valueDes;
          break;
        case r'categories':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(UpsertCategoryCommand)]),
          ) as BuiltList<UpsertCategoryCommand>;
          result.categories.replace(valueDes);
          break;
        case r'tags':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(UpsertTagCommand)]),
          ) as BuiltList<UpsertTagCommand>;
          result.tags.replace(valueDes);
          break;
        case r'receiptItems':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(UpsertItemCommand)]),
          ) as BuiltList<UpsertItemCommand>;
          result.receiptItems.replace(valueDes);
          break;
        case r'comments':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(UpsertCommentCommand)]),
          ) as BuiltList<UpsertCommentCommand>;
          result.comments.replace(valueDes);
          break;
        case r'customFields':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(UpsertCustomFieldValueCommand)]),
          ) as BuiltList<UpsertCustomFieldValueCommand>;
          result.customFields.replace(valueDes);
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  UpsertReceiptCommand deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = UpsertReceiptCommandBuilder();
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

