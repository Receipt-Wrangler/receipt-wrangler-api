//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/associated_entity_type.dart';
import 'package:openapi/src/model/sort_direction.dart';
import 'package:openapi/src/model/system_task_paged_request_filter.dart';
import 'package:openapi/src/model/paged_request_command.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'get_system_task_command.g.dart';

/// GetSystemTaskCommand
///
/// Properties:
/// * [page] - Page number
/// * [pageSize] - Number of records per page
/// * [orderBy] - field to order on
/// * [sortDirection] 
/// * [associatedEntityId] - Associated entity id
/// * [associatedEntityType] 
/// * [filter] 
@BuiltValue()
abstract class GetSystemTaskCommand implements PagedRequestCommand, Built<GetSystemTaskCommand, GetSystemTaskCommandBuilder> {
  @BuiltValueField(wireName: r'filter')
  SystemTaskPagedRequestFilter? get filter;

  /// Associated entity id
  @BuiltValueField(wireName: r'associatedEntityId')
  int? get associatedEntityId;

  @BuiltValueField(wireName: r'associatedEntityType')
  AssociatedEntityType? get associatedEntityType;
  // enum associatedEntityTypeEnum {  NOOP_ENTITY_TYPE,  RECEIPT,  SYSTEM_EMAIL,  RECEIPT_PROCESSING_SETTINGS,  PROMPT,  API_KEY,  };

  GetSystemTaskCommand._();

  factory GetSystemTaskCommand([void updates(GetSystemTaskCommandBuilder b)]) = _$GetSystemTaskCommand;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(GetSystemTaskCommandBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<GetSystemTaskCommand> get serializer => _$GetSystemTaskCommandSerializer();
}

class _$GetSystemTaskCommandSerializer implements PrimitiveSerializer<GetSystemTaskCommand> {
  @override
  final Iterable<Type> types = const [GetSystemTaskCommand, _$GetSystemTaskCommand];

  @override
  final String wireName = r'GetSystemTaskCommand';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    GetSystemTaskCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    if (object.filter != null) {
      yield r'filter';
      yield serializers.serialize(
        object.filter,
        specifiedType: const FullType(SystemTaskPagedRequestFilter),
      );
    }
    if (object.associatedEntityId != null) {
      yield r'associatedEntityId';
      yield serializers.serialize(
        object.associatedEntityId,
        specifiedType: const FullType(int),
      );
    }
    if (object.sortDirection != null) {
      yield r'sortDirection';
      yield serializers.serialize(
        object.sortDirection,
        specifiedType: const FullType(SortDirection),
      );
    }
    if (object.associatedEntityType != null) {
      yield r'associatedEntityType';
      yield serializers.serialize(
        object.associatedEntityType,
        specifiedType: const FullType(AssociatedEntityType),
      );
    }
    yield r'pageSize';
    yield serializers.serialize(
      object.pageSize,
      specifiedType: const FullType(int),
    );
    if (object.orderBy != null) {
      yield r'orderBy';
      yield serializers.serialize(
        object.orderBy,
        specifiedType: const FullType(String),
      );
    }
    yield r'page';
    yield serializers.serialize(
      object.page,
      specifiedType: const FullType(int),
    );
  }

  @override
  Object serialize(
    Serializers serializers,
    GetSystemTaskCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required GetSystemTaskCommandBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'filter':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(SystemTaskPagedRequestFilter),
          ) as SystemTaskPagedRequestFilter;
          result.filter.replace(valueDes);
          break;
        case r'associatedEntityId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.associatedEntityId = valueDes;
          break;
        case r'sortDirection':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(SortDirection),
          ) as SortDirection;
          result.sortDirection = valueDes;
          break;
        case r'associatedEntityType':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(AssociatedEntityType),
          ) as AssociatedEntityType;
          result.associatedEntityType = valueDes;
          break;
        case r'pageSize':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.pageSize = valueDes;
          break;
        case r'orderBy':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.orderBy = valueDes;
          break;
        case r'page':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.page = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  GetSystemTaskCommand deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = GetSystemTaskCommandBuilder();
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

