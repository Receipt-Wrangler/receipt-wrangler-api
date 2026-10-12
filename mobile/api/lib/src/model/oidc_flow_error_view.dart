//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'oidc_flow_error_view.g.dart';

/// A failure from a flow start that answers with JSON rather than a redirect. The code comes from the same small fixed vocabulary the redirect form puts in its query string, so one client-side table maps both. An identity provider's own error text is never echoed through.
///
/// Properties:
/// * [errorCode] 
@BuiltValue()
abstract class OidcFlowErrorView implements Built<OidcFlowErrorView, OidcFlowErrorViewBuilder> {
  @BuiltValueField(wireName: r'errorCode')
  String get errorCode;

  OidcFlowErrorView._();

  factory OidcFlowErrorView([void updates(OidcFlowErrorViewBuilder b)]) = _$OidcFlowErrorView;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(OidcFlowErrorViewBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<OidcFlowErrorView> get serializer => _$OidcFlowErrorViewSerializer();
}

class _$OidcFlowErrorViewSerializer implements PrimitiveSerializer<OidcFlowErrorView> {
  @override
  final Iterable<Type> types = const [OidcFlowErrorView, _$OidcFlowErrorView];

  @override
  final String wireName = r'OidcFlowErrorView';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    OidcFlowErrorView object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'errorCode';
    yield serializers.serialize(
      object.errorCode,
      specifiedType: const FullType(String),
    );
  }

  @override
  Object serialize(
    Serializers serializers,
    OidcFlowErrorView object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required OidcFlowErrorViewBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'errorCode':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.errorCode = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  OidcFlowErrorView deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = OidcFlowErrorViewBuilder();
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

