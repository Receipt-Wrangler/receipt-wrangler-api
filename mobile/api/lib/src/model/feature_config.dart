//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/oidc_provider_summary.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'feature_config.g.dart';

/// FeatureConfig
///
/// Properties:
/// * [aiPoweredReceipts] - Whether AI powered receipts are enabled
/// * [enableLocalSignUp] - Whether local sign up is enabled
/// * [loginQrUrl] - Composed deep link the desktop login page encodes as a QR; empty unless the login QR is enabled with a mobile server URL
/// * [oidcProviders] - Enabled OIDC identity providers, so a login screen can render one button per provider. Always present as an array (never null) and carries only the slug and display name -- never the issuer, client id, or any secret.
@BuiltValue()
abstract class FeatureConfig implements Built<FeatureConfig, FeatureConfigBuilder> {
  /// Whether AI powered receipts are enabled
  @BuiltValueField(wireName: r'aiPoweredReceipts')
  bool get aiPoweredReceipts;

  /// Whether local sign up is enabled
  @BuiltValueField(wireName: r'enableLocalSignUp')
  bool get enableLocalSignUp;

  /// Composed deep link the desktop login page encodes as a QR; empty unless the login QR is enabled with a mobile server URL
  @BuiltValueField(wireName: r'loginQrUrl')
  String? get loginQrUrl;

  /// Enabled OIDC identity providers, so a login screen can render one button per provider. Always present as an array (never null) and carries only the slug and display name -- never the issuer, client id, or any secret.
  @BuiltValueField(wireName: r'oidcProviders')
  BuiltList<OidcProviderSummary>? get oidcProviders;

  FeatureConfig._();

  factory FeatureConfig([void updates(FeatureConfigBuilder b)]) = _$FeatureConfig;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(FeatureConfigBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<FeatureConfig> get serializer => _$FeatureConfigSerializer();
}

class _$FeatureConfigSerializer implements PrimitiveSerializer<FeatureConfig> {
  @override
  final Iterable<Type> types = const [FeatureConfig, _$FeatureConfig];

  @override
  final String wireName = r'FeatureConfig';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    FeatureConfig object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'aiPoweredReceipts';
    yield serializers.serialize(
      object.aiPoweredReceipts,
      specifiedType: const FullType(bool),
    );
    yield r'enableLocalSignUp';
    yield serializers.serialize(
      object.enableLocalSignUp,
      specifiedType: const FullType(bool),
    );
    if (object.loginQrUrl != null) {
      yield r'loginQrUrl';
      yield serializers.serialize(
        object.loginQrUrl,
        specifiedType: const FullType(String),
      );
    }
    if (object.oidcProviders != null) {
      yield r'oidcProviders';
      yield serializers.serialize(
        object.oidcProviders,
        specifiedType: const FullType(BuiltList, [FullType(OidcProviderSummary)]),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    FeatureConfig object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required FeatureConfigBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'aiPoweredReceipts':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.aiPoweredReceipts = valueDes;
          break;
        case r'enableLocalSignUp':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.enableLocalSignUp = valueDes;
          break;
        case r'loginQrUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.loginQrUrl = valueDes;
          break;
        case r'oidcProviders':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(OidcProviderSummary)]),
          ) as BuiltList<OidcProviderSummary>;
          result.oidcProviders.replace(valueDes);
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  FeatureConfig deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = FeatureConfigBuilder();
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

