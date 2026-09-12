//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_value/json_object.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'system_task_paged_request_filter.g.dart';

/// SystemTaskPagedRequestFilter
///
/// Properties:
/// * [type] - Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
/// * [ranBy] - Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
/// * [startedAt] - Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
/// * [endedAt] - Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
@BuiltValue()
abstract class SystemTaskPagedRequestFilter implements Built<SystemTaskPagedRequestFilter, SystemTaskPagedRequestFilterBuilder> {
  /// Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
  @BuiltValueField(wireName: r'type')
  JsonObject? get type;

  /// Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
  @BuiltValueField(wireName: r'ranBy')
  JsonObject? get ranBy;

  /// Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
  @BuiltValueField(wireName: r'startedAt')
  JsonObject? get startedAt;

  /// Contains two keys: operation of type FilterOperation and value which can a different type depending on the field.
  @BuiltValueField(wireName: r'endedAt')
  JsonObject? get endedAt;

  SystemTaskPagedRequestFilter._();

  factory SystemTaskPagedRequestFilter([void updates(SystemTaskPagedRequestFilterBuilder b)]) = _$SystemTaskPagedRequestFilter;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(SystemTaskPagedRequestFilterBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<SystemTaskPagedRequestFilter> get serializer => _$SystemTaskPagedRequestFilterSerializer();
}

class _$SystemTaskPagedRequestFilterSerializer implements PrimitiveSerializer<SystemTaskPagedRequestFilter> {
  @override
  final Iterable<Type> types = const [SystemTaskPagedRequestFilter, _$SystemTaskPagedRequestFilter];

  @override
  final String wireName = r'SystemTaskPagedRequestFilter';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    SystemTaskPagedRequestFilter object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    if (object.type != null) {
      yield r'type';
      yield serializers.serialize(
        object.type,
        specifiedType: const FullType(JsonObject),
      );
    }
    if (object.ranBy != null) {
      yield r'ranBy';
      yield serializers.serialize(
        object.ranBy,
        specifiedType: const FullType(JsonObject),
      );
    }
    if (object.startedAt != null) {
      yield r'startedAt';
      yield serializers.serialize(
        object.startedAt,
        specifiedType: const FullType(JsonObject),
      );
    }
    if (object.endedAt != null) {
      yield r'endedAt';
      yield serializers.serialize(
        object.endedAt,
        specifiedType: const FullType(JsonObject),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    SystemTaskPagedRequestFilter object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required SystemTaskPagedRequestFilterBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'type':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(JsonObject),
          ) as JsonObject;
          result.type = valueDes;
          break;
        case r'ranBy':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(JsonObject),
          ) as JsonObject;
          result.ranBy = valueDes;
          break;
        case r'startedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(JsonObject),
          ) as JsonObject;
          result.startedAt = valueDes;
          break;
        case r'endedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(JsonObject),
          ) as JsonObject;
          result.endedAt = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  SystemTaskPagedRequestFilter deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = SystemTaskPagedRequestFilterBuilder();
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

