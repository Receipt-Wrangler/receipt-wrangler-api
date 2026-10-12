//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'oidc_provider_summary.g.dart';

/// OidcProviderSummary
///
/// Properties:
/// * [name] - The provider's slug, used to build its login URL
/// * [displayName] - Rendered to users as \"Log in with {displayName}\"
@BuiltValue()
abstract class OidcProviderSummary implements Built<OidcProviderSummary, OidcProviderSummaryBuilder> {
  /// The provider's slug, used to build its login URL
  @BuiltValueField(wireName: r'name')
  String get name;

  /// Rendered to users as \"Log in with {displayName}\"
  @BuiltValueField(wireName: r'displayName')
  String get displayName;

  OidcProviderSummary._();

  factory OidcProviderSummary([void updates(OidcProviderSummaryBuilder b)]) = _$OidcProviderSummary;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(OidcProviderSummaryBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<OidcProviderSummary> get serializer => _$OidcProviderSummarySerializer();
}

class _$OidcProviderSummarySerializer implements PrimitiveSerializer<OidcProviderSummary> {
  @override
  final Iterable<Type> types = const [OidcProviderSummary, _$OidcProviderSummary];

  @override
  final String wireName = r'OidcProviderSummary';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    OidcProviderSummary object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'name';
    yield serializers.serialize(
      object.name,
      specifiedType: const FullType(String),
    );
    yield r'displayName';
    yield serializers.serialize(
      object.displayName,
      specifiedType: const FullType(String),
    );
  }

  @override
  Object serialize(
    Serializers serializers,
    OidcProviderSummary object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required OidcProviderSummaryBuilder result,
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
        case r'displayName':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.displayName = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  OidcProviderSummary deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = OidcProviderSummaryBuilder();
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

