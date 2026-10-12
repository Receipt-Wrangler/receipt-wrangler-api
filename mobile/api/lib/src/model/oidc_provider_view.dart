//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'oidc_provider_view.g.dart';

/// OidcProviderView
///
/// Properties:
/// * [id] 
/// * [name] - URL slug for this provider; immutable after creation
/// * [displayName] 
/// * [issuerUrl] - OIDC discovery base, e.g. https://accounts.google.com
/// * [clientId] 
/// * [scope] - Space-separated OIDC scopes; must include openid
/// * [allowProvisioning] - Create a local account for an identity we have never seen
/// * [linkByUsername] - On a first login only, attach to an existing local account whose username equals the preferred_username claim. Off by default; that claim is neither stable nor unique, and some providers recycle released usernames.
/// * [enabled] 
/// * [hasClientSecret] - Whether a secret is stored. The secret itself is never returned.
/// * [redirectUri] - The exact redirect URI to register with the identity provider
/// * [createdAt] 
/// * [updatedAt] 
@BuiltValue()
abstract class OidcProviderView implements Built<OidcProviderView, OidcProviderViewBuilder> {
  @BuiltValueField(wireName: r'id')
  int get id;

  /// URL slug for this provider; immutable after creation
  @BuiltValueField(wireName: r'name')
  String get name;

  @BuiltValueField(wireName: r'displayName')
  String get displayName;

  /// OIDC discovery base, e.g. https://accounts.google.com
  @BuiltValueField(wireName: r'issuerUrl')
  String get issuerUrl;

  @BuiltValueField(wireName: r'clientId')
  String get clientId;

  /// Space-separated OIDC scopes; must include openid
  @BuiltValueField(wireName: r'scope')
  String get scope;

  /// Create a local account for an identity we have never seen
  @BuiltValueField(wireName: r'allowProvisioning')
  bool get allowProvisioning;

  /// On a first login only, attach to an existing local account whose username equals the preferred_username claim. Off by default; that claim is neither stable nor unique, and some providers recycle released usernames.
  @BuiltValueField(wireName: r'linkByUsername')
  bool get linkByUsername;

  @BuiltValueField(wireName: r'enabled')
  bool get enabled;

  /// Whether a secret is stored. The secret itself is never returned.
  @BuiltValueField(wireName: r'hasClientSecret')
  bool get hasClientSecret;

  /// The exact redirect URI to register with the identity provider
  @BuiltValueField(wireName: r'redirectUri')
  String get redirectUri;

  @BuiltValueField(wireName: r'createdAt')
  DateTime? get createdAt;

  @BuiltValueField(wireName: r'updatedAt')
  DateTime? get updatedAt;

  OidcProviderView._();

  factory OidcProviderView([void updates(OidcProviderViewBuilder b)]) = _$OidcProviderView;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(OidcProviderViewBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<OidcProviderView> get serializer => _$OidcProviderViewSerializer();
}

class _$OidcProviderViewSerializer implements PrimitiveSerializer<OidcProviderView> {
  @override
  final Iterable<Type> types = const [OidcProviderView, _$OidcProviderView];

  @override
  final String wireName = r'OidcProviderView';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    OidcProviderView object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'id';
    yield serializers.serialize(
      object.id,
      specifiedType: const FullType(int),
    );
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
    yield r'scope';
    yield serializers.serialize(
      object.scope,
      specifiedType: const FullType(String),
    );
    yield r'allowProvisioning';
    yield serializers.serialize(
      object.allowProvisioning,
      specifiedType: const FullType(bool),
    );
    yield r'linkByUsername';
    yield serializers.serialize(
      object.linkByUsername,
      specifiedType: const FullType(bool),
    );
    yield r'enabled';
    yield serializers.serialize(
      object.enabled,
      specifiedType: const FullType(bool),
    );
    yield r'hasClientSecret';
    yield serializers.serialize(
      object.hasClientSecret,
      specifiedType: const FullType(bool),
    );
    yield r'redirectUri';
    yield serializers.serialize(
      object.redirectUri,
      specifiedType: const FullType(String),
    );
    if (object.createdAt != null) {
      yield r'createdAt';
      yield serializers.serialize(
        object.createdAt,
        specifiedType: const FullType(DateTime),
      );
    }
    if (object.updatedAt != null) {
      yield r'updatedAt';
      yield serializers.serialize(
        object.updatedAt,
        specifiedType: const FullType(DateTime),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    OidcProviderView object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required OidcProviderViewBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'id':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.id = valueDes;
          break;
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
        case r'hasClientSecret':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.hasClientSecret = valueDes;
          break;
        case r'redirectUri':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.redirectUri = valueDes;
          break;
        case r'createdAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(DateTime),
          ) as DateTime;
          result.createdAt = valueDes;
          break;
        case r'updatedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(DateTime),
          ) as DateTime;
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
  OidcProviderView deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = OidcProviderViewBuilder();
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

