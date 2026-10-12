//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'oidc_connection_view.g.dart';

/// OidcConnectionView
///
/// Properties:
/// * [providerName] 
/// * [providerDisplayName] 
/// * [preferredUsername] 
/// * [email] 
/// * [provisionedUser] - Whether this connection created the local account. Such an account has no usable password, so its last connection cannot be removed.
/// * [linkedAt] 
/// * [lastLoginAt] 
@BuiltValue()
abstract class OidcConnectionView implements Built<OidcConnectionView, OidcConnectionViewBuilder> {
  @BuiltValueField(wireName: r'providerName')
  String get providerName;

  @BuiltValueField(wireName: r'providerDisplayName')
  String get providerDisplayName;

  @BuiltValueField(wireName: r'preferredUsername')
  String? get preferredUsername;

  @BuiltValueField(wireName: r'email')
  String? get email;

  /// Whether this connection created the local account. Such an account has no usable password, so its last connection cannot be removed.
  @BuiltValueField(wireName: r'provisionedUser')
  bool get provisionedUser;

  @BuiltValueField(wireName: r'linkedAt')
  DateTime get linkedAt;

  @BuiltValueField(wireName: r'lastLoginAt')
  DateTime? get lastLoginAt;

  OidcConnectionView._();

  factory OidcConnectionView([void updates(OidcConnectionViewBuilder b)]) = _$OidcConnectionView;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(OidcConnectionViewBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<OidcConnectionView> get serializer => _$OidcConnectionViewSerializer();
}

class _$OidcConnectionViewSerializer implements PrimitiveSerializer<OidcConnectionView> {
  @override
  final Iterable<Type> types = const [OidcConnectionView, _$OidcConnectionView];

  @override
  final String wireName = r'OidcConnectionView';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    OidcConnectionView object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'providerName';
    yield serializers.serialize(
      object.providerName,
      specifiedType: const FullType(String),
    );
    yield r'providerDisplayName';
    yield serializers.serialize(
      object.providerDisplayName,
      specifiedType: const FullType(String),
    );
    if (object.preferredUsername != null) {
      yield r'preferredUsername';
      yield serializers.serialize(
        object.preferredUsername,
        specifiedType: const FullType(String),
      );
    }
    if (object.email != null) {
      yield r'email';
      yield serializers.serialize(
        object.email,
        specifiedType: const FullType(String),
      );
    }
    yield r'provisionedUser';
    yield serializers.serialize(
      object.provisionedUser,
      specifiedType: const FullType(bool),
    );
    yield r'linkedAt';
    yield serializers.serialize(
      object.linkedAt,
      specifiedType: const FullType(DateTime),
    );
    if (object.lastLoginAt != null) {
      yield r'lastLoginAt';
      yield serializers.serialize(
        object.lastLoginAt,
        specifiedType: const FullType(DateTime),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    OidcConnectionView object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required OidcConnectionViewBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'providerName':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.providerName = valueDes;
          break;
        case r'providerDisplayName':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.providerDisplayName = valueDes;
          break;
        case r'preferredUsername':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.preferredUsername = valueDes;
          break;
        case r'email':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.email = valueDes;
          break;
        case r'provisionedUser':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.provisionedUser = valueDes;
          break;
        case r'linkedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(DateTime),
          ) as DateTime;
          result.linkedAt = valueDes;
          break;
        case r'lastLoginAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(DateTime),
          ) as DateTime;
          result.lastLoginAt = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  OidcConnectionView deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = OidcConnectionViewBuilder();
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

