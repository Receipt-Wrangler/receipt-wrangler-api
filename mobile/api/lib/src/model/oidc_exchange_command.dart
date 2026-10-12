//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'oidc_exchange_command.g.dart';

/// OidcExchangeCommand
///
/// Properties:
/// * [code] - The single-use code delivered to the app's callback URL
/// * [codeVerifier] - The PKCE verifier the app generated before starting the flow
@BuiltValue()
abstract class OidcExchangeCommand implements Built<OidcExchangeCommand, OidcExchangeCommandBuilder> {
  /// The single-use code delivered to the app's callback URL
  @BuiltValueField(wireName: r'code')
  String get code;

  /// The PKCE verifier the app generated before starting the flow
  @BuiltValueField(wireName: r'codeVerifier')
  String get codeVerifier;

  OidcExchangeCommand._();

  factory OidcExchangeCommand([void updates(OidcExchangeCommandBuilder b)]) = _$OidcExchangeCommand;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(OidcExchangeCommandBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<OidcExchangeCommand> get serializer => _$OidcExchangeCommandSerializer();
}

class _$OidcExchangeCommandSerializer implements PrimitiveSerializer<OidcExchangeCommand> {
  @override
  final Iterable<Type> types = const [OidcExchangeCommand, _$OidcExchangeCommand];

  @override
  final String wireName = r'OidcExchangeCommand';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    OidcExchangeCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'code';
    yield serializers.serialize(
      object.code,
      specifiedType: const FullType(String),
    );
    yield r'codeVerifier';
    yield serializers.serialize(
      object.codeVerifier,
      specifiedType: const FullType(String),
    );
  }

  @override
  Object serialize(
    Serializers serializers,
    OidcExchangeCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required OidcExchangeCommandBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'code':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.code = valueDes;
          break;
        case r'codeVerifier':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.codeVerifier = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  OidcExchangeCommand deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = OidcExchangeCommandBuilder();
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

