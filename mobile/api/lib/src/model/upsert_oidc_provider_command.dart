//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'upsert_oidc_provider_command.g.dart';

/// UpsertOidcProviderCommand
///
/// Properties:
/// * [name] - URL slug. Lowercase letters, numbers and dashes. Cannot be changed after creation, and cannot be one of the reserved words login, callback, link, exchange or connections.
/// * [displayName] 
/// * [issuerUrl] - Must use https unless the host is localhost
/// * [clientId] 
/// * [clientSecret] - Omit on update to keep the stored secret. Required on create.
/// * [scope] - Space-separated OIDC scopes; must include openid
/// * [allowProvisioning] 
/// * [linkByUsername] 
/// * [enabled] 
@BuiltValue()
abstract class UpsertOidcProviderCommand implements Built<UpsertOidcProviderCommand, UpsertOidcProviderCommandBuilder> {
  /// URL slug. Lowercase letters, numbers and dashes. Cannot be changed after creation, and cannot be one of the reserved words login, callback, link, exchange or connections.
  @BuiltValueField(wireName: r'name')
  String get name;

  @BuiltValueField(wireName: r'displayName')
  String get displayName;

  /// Must use https unless the host is localhost
  @BuiltValueField(wireName: r'issuerUrl')
  String get issuerUrl;

  @BuiltValueField(wireName: r'clientId')
  String get clientId;

  /// Omit on update to keep the stored secret. Required on create.
  @BuiltValueField(wireName: r'clientSecret')
  String? get clientSecret;

  /// Space-separated OIDC scopes; must include openid
  @BuiltValueField(wireName: r'scope')
  String get scope;

  @BuiltValueField(wireName: r'allowProvisioning')
  bool? get allowProvisioning;

  @BuiltValueField(wireName: r'linkByUsername')
  bool? get linkByUsername;

  @BuiltValueField(wireName: r'enabled')
  bool? get enabled;

  UpsertOidcProviderCommand._();

  factory UpsertOidcProviderCommand([void updates(UpsertOidcProviderCommandBuilder b)]) = _$UpsertOidcProviderCommand;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(UpsertOidcProviderCommandBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<UpsertOidcProviderCommand> get serializer => _$UpsertOidcProviderCommandSerializer();
}

class _$UpsertOidcProviderCommandSerializer implements PrimitiveSerializer<UpsertOidcProviderCommand> {
  @override
  final Iterable<Type> types = const [UpsertOidcProviderCommand, _$UpsertOidcProviderCommand];

  @override
  final String wireName = r'UpsertOidcProviderCommand';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    UpsertOidcProviderCommand object, {
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
    yield r'issuerUrl';
    yield serializers.serialize(
      object.issuerUrl,
      specifiedType: const FullType(String),
    );
    yield r'clientId';
    yield serializers.serialize(
      object.clientId,
      specifiedType: const FullType(String),
    );
    if (object.clientSecret != null) {
      yield r'clientSecret';
      yield serializers.serialize(
        object.clientSecret,
        specifiedType: const FullType(String),
      );
    }
    yield r'scope';
    yield serializers.serialize(
      object.scope,
      specifiedType: const FullType(String),
    );
    if (object.allowProvisioning != null) {
      yield r'allowProvisioning';
      yield serializers.serialize(
        object.allowProvisioning,
        specifiedType: const FullType(bool),
      );
    }
    if (object.linkByUsername != null) {
      yield r'linkByUsername';
      yield serializers.serialize(
        object.linkByUsername,
        specifiedType: const FullType(bool),
      );
    }
    if (object.enabled != null) {
      yield r'enabled';
      yield serializers.serialize(
        object.enabled,
        specifiedType: const FullType(bool),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    UpsertOidcProviderCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required UpsertOidcProviderCommandBuilder result,
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
        case r'issuerUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.issuerUrl = valueDes;
          break;
        case r'clientId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.clientId = valueDes;
          break;
        case r'clientSecret':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.clientSecret = valueDes;
          break;
        case r'scope':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.scope = valueDes;
          break;
        case r'allowProvisioning':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.allowProvisioning = valueDes;
          break;
        case r'linkByUsername':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.linkByUsername = valueDes;
          break;
        case r'enabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.enabled = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  UpsertOidcProviderCommand deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = UpsertOidcProviderCommandBuilder();
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

