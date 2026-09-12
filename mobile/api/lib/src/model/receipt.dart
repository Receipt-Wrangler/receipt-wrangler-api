//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/receipt_status.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/category.dart';
import 'package:openapi/src/model/custom_field_value.dart';
import 'package:openapi/src/model/tag.dart';
import 'package:openapi/src/model/comment.dart';
import 'package:openapi/src/model/file_data.dart';
import 'package:openapi/src/model/item.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'receipt.g.dart';

/// Receipt
///
/// Properties:
/// * [amount] - Receipt total amount
/// * [categories] - Categories associated to receipt
/// * [comments] - Comments associated to receipt
/// * [customFields] - Custom fields associated to receipt
/// * [createdAt] 
/// * [createdBy] 
/// * [date] - Receipt date
/// * [groupId] - Group foreign key
/// * [id] 
/// * [imageFiles] - Files associated to receipt
/// * [name] - Receipt name
/// * [paidByUserId] - User paid foreign key
/// * [receiptItems] - Items associated to receipt
/// * [resolvedDate] - Date resolved
/// * [status] 
/// * [tags] - Tags associated to receipt
/// * [updatedAt] 
/// * [createdByString] - Created by string, which is anything that is not a user
@BuiltValue()
abstract class Receipt implements Built<Receipt, ReceiptBuilder> {
  /// Receipt total amount
  @BuiltValueField(wireName: r'amount')
  String get amount;

  /// Categories associated to receipt
  @BuiltValueField(wireName: r'categories')
  BuiltList<Category> get categories;

  /// Comments associated to receipt
  @BuiltValueField(wireName: r'comments')
  BuiltList<Comment> get comments;

  /// Custom fields associated to receipt
  @BuiltValueField(wireName: r'customFields')
  BuiltList<CustomFieldValue> get customFields;

  @BuiltValueField(wireName: r'createdAt')
  String? get createdAt;

  @BuiltValueField(wireName: r'createdBy')
  int? get createdBy;

  /// Receipt date
  @BuiltValueField(wireName: r'date')
  String get date;

  /// Group foreign key
  @BuiltValueField(wireName: r'groupId')
  int get groupId;

  @BuiltValueField(wireName: r'id')
  int get id;

  /// Files associated to receipt
  @BuiltValueField(wireName: r'imageFiles')
  BuiltList<FileData>? get imageFiles;

  /// Receipt name
  @BuiltValueField(wireName: r'name')
  String get name;

  /// User paid foreign key
  @BuiltValueField(wireName: r'paidByUserId')
  int get paidByUserId;

  /// Items associated to receipt
  @BuiltValueField(wireName: r'receiptItems')
  BuiltList<Item> get receiptItems;

  /// Date resolved
  @BuiltValueField(wireName: r'resolvedDate')
  String? get resolvedDate;

  @BuiltValueField(wireName: r'status')
  ReceiptStatus get status;
  // enum statusEnum {  OPEN,  NEEDS_ATTENTION,  RESOLVED,  DRAFT,  DECLINED,  ,  };

  /// Tags associated to receipt
  @BuiltValueField(wireName: r'tags')
  BuiltList<Tag> get tags;

  @BuiltValueField(wireName: r'updatedAt')
  String? get updatedAt;

  /// Created by string, which is anything that is not a user
  @BuiltValueField(wireName: r'createdByString')
  String? get createdByString;

  Receipt._();

  factory Receipt([void updates(ReceiptBuilder b)]) = _$Receipt;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(ReceiptBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<Receipt> get serializer => _$ReceiptSerializer();
}

class _$ReceiptSerializer implements PrimitiveSerializer<Receipt> {
  @override
  final Iterable<Type> types = const [Receipt, _$Receipt];

  @override
  final String wireName = r'Receipt';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    Receipt object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'amount';
    yield serializers.serialize(
      object.amount,
      specifiedType: const FullType(String),
    );
    yield r'categories';
    yield serializers.serialize(
      object.categories,
      specifiedType: const FullType(BuiltList, [FullType(Category)]),
    );
    yield r'comments';
    yield serializers.serialize(
      object.comments,
      specifiedType: const FullType(BuiltList, [FullType(Comment)]),
    );
    yield r'customFields';
    yield serializers.serialize(
      object.customFields,
      specifiedType: const FullType(BuiltList, [FullType(CustomFieldValue)]),
    );
    if (object.createdAt != null) {
      yield r'createdAt';
      yield serializers.serialize(
        object.createdAt,
        specifiedType: const FullType(String),
      );
    }
    if (object.createdBy != null) {
      yield r'createdBy';
      yield serializers.serialize(
        object.createdBy,
        specifiedType: const FullType(int),
      );
    }
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
    yield r'id';
    yield serializers.serialize(
      object.id,
      specifiedType: const FullType(int),
    );
    if (object.imageFiles != null) {
      yield r'imageFiles';
      yield serializers.serialize(
        object.imageFiles,
        specifiedType: const FullType(BuiltList, [FullType(FileData)]),
      );
    }
    yield r'name';
    yield serializers.serialize(
      object.name,
      specifiedType: const FullType(String),
    );
    yield r'paidByUserId';
    yield serializers.serialize(
      object.paidByUserId,
      specifiedType: const FullType(int),
    );
    yield r'receiptItems';
    yield serializers.serialize(
      object.receiptItems,
      specifiedType: const FullType(BuiltList, [FullType(Item)]),
    );
    if (object.resolvedDate != null) {
      yield r'resolvedDate';
      yield serializers.serialize(
        object.resolvedDate,
        specifiedType: const FullType(String),
      );
    }
    yield r'status';
    yield serializers.serialize(
      object.status,
      specifiedType: const FullType(ReceiptStatus),
    );
    yield r'tags';
    yield serializers.serialize(
      object.tags,
      specifiedType: const FullType(BuiltList, [FullType(Tag)]),
    );
    if (object.updatedAt != null) {
      yield r'updatedAt';
      yield serializers.serialize(
        object.updatedAt,
        specifiedType: const FullType(String),
      );
    }
    if (object.createdByString != null) {
      yield r'createdByString';
      yield serializers.serialize(
        object.createdByString,
        specifiedType: const FullType(String),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    Receipt object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required ReceiptBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'amount':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.amount = valueDes;
          break;
        case r'categories':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(Category)]),
          ) as BuiltList<Category>;
          result.categories.replace(valueDes);
          break;
        case r'comments':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(Comment)]),
          ) as BuiltList<Comment>;
          result.comments.replace(valueDes);
          break;
        case r'customFields':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(CustomFieldValue)]),
          ) as BuiltList<CustomFieldValue>;
          result.customFields.replace(valueDes);
          break;
        case r'createdAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.createdAt = valueDes;
          break;
        case r'createdBy':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.createdBy = valueDes;
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
        case r'id':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.id = valueDes;
          break;
        case r'imageFiles':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(FileData)]),
          ) as BuiltList<FileData>;
          result.imageFiles.replace(valueDes);
          break;
        case r'name':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.name = valueDes;
          break;
        case r'paidByUserId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.paidByUserId = valueDes;
          break;
        case r'receiptItems':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(Item)]),
          ) as BuiltList<Item>;
          result.receiptItems.replace(valueDes);
          break;
        case r'resolvedDate':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.resolvedDate = valueDes;
          break;
        case r'status':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(ReceiptStatus),
          ) as ReceiptStatus;
          result.status = valueDes;
          break;
        case r'tags':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(Tag)]),
          ) as BuiltList<Tag>;
          result.tags.replace(valueDes);
          break;
        case r'updatedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.updatedAt = valueDes;
          break;
        case r'createdByString':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.createdByString = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  Receipt deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = ReceiptBuilder();
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

