//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/currency_symbol_position.dart';
import 'package:openapi/src/model/upsert_task_queue_configuration.dart';
import 'package:openapi/src/model/currency_separator.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'upsert_system_settings_command.g.dart';

/// UpsertSystemSettingsCommand
///
/// Properties:
/// * [enableLocalSignUp] - Whether local sign up is enabled
/// * [currencyDisplay] - Currency display
/// * [currencyThousandthsSeparator] 
/// * [currencyDecimalSeparator] 
/// * [currencySymbolPosition] 
/// * [currencyHideDecimalPlaces] - Whether to hide decimal places
/// * [debugOcr] 
/// * [numWorkers] - Number of workers to use
/// * [emailPollingInterval] - Email polling interval
/// * [receiptProcessingSettingsId] - Receipt processing settings foreign key
/// * [fallbackReceiptProcessingSettingsId] - Fallback receipt processing settings foreign key
/// * [taskConcurrency] - Concurrency for task worker
/// * [pdfDpi] - DPI used when rasterizing PDFs for OCR/vision processing
/// * [taskQueueConfigurations] 
/// * [mcpEnabled] - Whether the OAuth 2.1-protected MCP server is enabled
/// * [mcpPublicUrl] - Externally reachable origin used for MCP OAuth/metadata/redirect URLs and token audience
/// * [serverPublicUrl] - Externally reachable origin of this API, used to build each OIDC provider's redirect URI. Separate from mcpPublicUrl, which is bound into the MCP token audience.
/// * [showLoginQr] - Whether to show the mobile-setup QR code on the desktop login page
/// * [mobileServerUrl] - Server/API URL mobile clients connect to; encoded into the login QR's deep link
/// * [refreshTokenValidForHours] - How long a refresh token stays valid, in hours. Accepted values are 1-720 (30 days), or 0 meaning unset, which falls back to the default of 24. Omit the key entirely to leave the currently configured value unchanged.
/// * [mcpRefreshTokenValidForHours] - How long an MCP/OAuth connector refresh token stays valid, in hours. Accepted values are 1-720 (30 days), or 0 meaning unset, which falls back to the default of 24. Omit the key entirely to leave the currently configured value unchanged.
/// * [tempFileRetentionHours] - How long a temp file is kept once nothing can still act on it, in hours. Accepted values are 24-8760 (1 year), or 0 meaning unset, which falls back to the default of 720 (30 days). Omit the key entirely to leave the currently configured value unchanged.
/// * [timeZone] - The app time zone, an IANA name such as America/New_York that the server can load (an unknown name is a 400 under the timeZone key). Omit the key entirely to leave the currently configured value unchanged.
@BuiltValue()
abstract class UpsertSystemSettingsCommand implements Built<UpsertSystemSettingsCommand, UpsertSystemSettingsCommandBuilder> {
  /// Whether local sign up is enabled
  @BuiltValueField(wireName: r'enableLocalSignUp')
  bool? get enableLocalSignUp;

  /// Currency display
  @BuiltValueField(wireName: r'currencyDisplay')
  String? get currencyDisplay;

  @BuiltValueField(wireName: r'currencyThousandthsSeparator')
  CurrencySeparator get currencyThousandthsSeparator;
  // enum currencyThousandthsSeparatorEnum {  ,,  .,  };

  @BuiltValueField(wireName: r'currencyDecimalSeparator')
  CurrencySeparator get currencyDecimalSeparator;
  // enum currencyDecimalSeparatorEnum {  ,,  .,  };

  @BuiltValueField(wireName: r'currencySymbolPosition')
  CurrencySymbolPosition get currencySymbolPosition;
  // enum currencySymbolPositionEnum {  START,  END,  };

  /// Whether to hide decimal places
  @BuiltValueField(wireName: r'currencyHideDecimalPlaces')
  bool get currencyHideDecimalPlaces;

  @BuiltValueField(wireName: r'debugOcr')
  bool? get debugOcr;

  /// Number of workers to use
  @BuiltValueField(wireName: r'numWorkers')
  int? get numWorkers;

  /// Email polling interval
  @BuiltValueField(wireName: r'emailPollingInterval')
  int? get emailPollingInterval;

  /// Receipt processing settings foreign key
  @BuiltValueField(wireName: r'receiptProcessingSettingsId')
  int? get receiptProcessingSettingsId;

  /// Fallback receipt processing settings foreign key
  @BuiltValueField(wireName: r'fallbackReceiptProcessingSettingsId')
  int? get fallbackReceiptProcessingSettingsId;

  /// Concurrency for task worker
  @BuiltValueField(wireName: r'taskConcurrency')
  int get taskConcurrency;

  /// DPI used when rasterizing PDFs for OCR/vision processing
  @BuiltValueField(wireName: r'pdfDpi')
  int? get pdfDpi;

  @BuiltValueField(wireName: r'taskQueueConfigurations')
  BuiltList<UpsertTaskQueueConfiguration>? get taskQueueConfigurations;

  /// Whether the OAuth 2.1-protected MCP server is enabled
  @BuiltValueField(wireName: r'mcpEnabled')
  bool? get mcpEnabled;

  /// Externally reachable origin used for MCP OAuth/metadata/redirect URLs and token audience
  @BuiltValueField(wireName: r'mcpPublicUrl')
  String? get mcpPublicUrl;

  /// Externally reachable origin of this API, used to build each OIDC provider's redirect URI. Separate from mcpPublicUrl, which is bound into the MCP token audience.
  @BuiltValueField(wireName: r'serverPublicUrl')
  String? get serverPublicUrl;

  /// Whether to show the mobile-setup QR code on the desktop login page
  @BuiltValueField(wireName: r'showLoginQr')
  bool? get showLoginQr;

  /// Server/API URL mobile clients connect to; encoded into the login QR's deep link
  @BuiltValueField(wireName: r'mobileServerUrl')
  String? get mobileServerUrl;

  /// How long a refresh token stays valid, in hours. Accepted values are 1-720 (30 days), or 0 meaning unset, which falls back to the default of 24. Omit the key entirely to leave the currently configured value unchanged.
  @BuiltValueField(wireName: r'refreshTokenValidForHours')
  int? get refreshTokenValidForHours;

  /// How long an MCP/OAuth connector refresh token stays valid, in hours. Accepted values are 1-720 (30 days), or 0 meaning unset, which falls back to the default of 24. Omit the key entirely to leave the currently configured value unchanged.
  @BuiltValueField(wireName: r'mcpRefreshTokenValidForHours')
  int? get mcpRefreshTokenValidForHours;

  /// How long a temp file is kept once nothing can still act on it, in hours. Accepted values are 24-8760 (1 year), or 0 meaning unset, which falls back to the default of 720 (30 days). Omit the key entirely to leave the currently configured value unchanged.
  @BuiltValueField(wireName: r'tempFileRetentionHours')
  int? get tempFileRetentionHours;

  /// The app time zone, an IANA name such as America/New_York that the server can load (an unknown name is a 400 under the timeZone key). Omit the key entirely to leave the currently configured value unchanged.
  @BuiltValueField(wireName: r'timeZone')
  String? get timeZone;

  UpsertSystemSettingsCommand._();

  factory UpsertSystemSettingsCommand([void updates(UpsertSystemSettingsCommandBuilder b)]) = _$UpsertSystemSettingsCommand;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(UpsertSystemSettingsCommandBuilder b) => b
      ..numWorkers = 1;

  @BuiltValueSerializer(custom: true)
  static Serializer<UpsertSystemSettingsCommand> get serializer => _$UpsertSystemSettingsCommandSerializer();
}

class _$UpsertSystemSettingsCommandSerializer implements PrimitiveSerializer<UpsertSystemSettingsCommand> {
  @override
  final Iterable<Type> types = const [UpsertSystemSettingsCommand, _$UpsertSystemSettingsCommand];

  @override
  final String wireName = r'UpsertSystemSettingsCommand';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    UpsertSystemSettingsCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    if (object.enableLocalSignUp != null) {
      yield r'enableLocalSignUp';
      yield serializers.serialize(
        object.enableLocalSignUp,
        specifiedType: const FullType(bool),
      );
    }
    if (object.currencyDisplay != null) {
      yield r'currencyDisplay';
      yield serializers.serialize(
        object.currencyDisplay,
        specifiedType: const FullType(String),
      );
    }
    yield r'currencyThousandthsSeparator';
    yield serializers.serialize(
      object.currencyThousandthsSeparator,
      specifiedType: const FullType(CurrencySeparator),
    );
    yield r'currencyDecimalSeparator';
    yield serializers.serialize(
      object.currencyDecimalSeparator,
      specifiedType: const FullType(CurrencySeparator),
    );
    yield r'currencySymbolPosition';
    yield serializers.serialize(
      object.currencySymbolPosition,
      specifiedType: const FullType(CurrencySymbolPosition),
    );
    yield r'currencyHideDecimalPlaces';
    yield serializers.serialize(
      object.currencyHideDecimalPlaces,
      specifiedType: const FullType(bool),
    );
    if (object.debugOcr != null) {
      yield r'debugOcr';
      yield serializers.serialize(
        object.debugOcr,
        specifiedType: const FullType(bool),
      );
    }
    if (object.numWorkers != null) {
      yield r'numWorkers';
      yield serializers.serialize(
        object.numWorkers,
        specifiedType: const FullType(int),
      );
    }
    if (object.emailPollingInterval != null) {
      yield r'emailPollingInterval';
      yield serializers.serialize(
        object.emailPollingInterval,
        specifiedType: const FullType(int),
      );
    }
    if (object.receiptProcessingSettingsId != null) {
      yield r'receiptProcessingSettingsId';
      yield serializers.serialize(
        object.receiptProcessingSettingsId,
        specifiedType: const FullType(int),
      );
    }
    if (object.fallbackReceiptProcessingSettingsId != null) {
      yield r'fallbackReceiptProcessingSettingsId';
      yield serializers.serialize(
        object.fallbackReceiptProcessingSettingsId,
        specifiedType: const FullType(int),
      );
    }
    yield r'taskConcurrency';
    yield serializers.serialize(
      object.taskConcurrency,
      specifiedType: const FullType(int),
    );
    if (object.pdfDpi != null) {
      yield r'pdfDpi';
      yield serializers.serialize(
        object.pdfDpi,
        specifiedType: const FullType(int),
      );
    }
    if (object.taskQueueConfigurations != null) {
      yield r'taskQueueConfigurations';
      yield serializers.serialize(
        object.taskQueueConfigurations,
        specifiedType: const FullType(BuiltList, [FullType(UpsertTaskQueueConfiguration)]),
      );
    }
    if (object.mcpEnabled != null) {
      yield r'mcpEnabled';
      yield serializers.serialize(
        object.mcpEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.mcpPublicUrl != null) {
      yield r'mcpPublicUrl';
      yield serializers.serialize(
        object.mcpPublicUrl,
        specifiedType: const FullType(String),
      );
    }
    if (object.serverPublicUrl != null) {
      yield r'serverPublicUrl';
      yield serializers.serialize(
        object.serverPublicUrl,
        specifiedType: const FullType(String),
      );
    }
    if (object.showLoginQr != null) {
      yield r'showLoginQr';
      yield serializers.serialize(
        object.showLoginQr,
        specifiedType: const FullType(bool),
      );
    }
    if (object.mobileServerUrl != null) {
      yield r'mobileServerUrl';
      yield serializers.serialize(
        object.mobileServerUrl,
        specifiedType: const FullType(String),
      );
    }
    if (object.refreshTokenValidForHours != null) {
      yield r'refreshTokenValidForHours';
      yield serializers.serialize(
        object.refreshTokenValidForHours,
        specifiedType: const FullType(int),
      );
    }
    if (object.mcpRefreshTokenValidForHours != null) {
      yield r'mcpRefreshTokenValidForHours';
      yield serializers.serialize(
        object.mcpRefreshTokenValidForHours,
        specifiedType: const FullType(int),
      );
    }
    if (object.tempFileRetentionHours != null) {
      yield r'tempFileRetentionHours';
      yield serializers.serialize(
        object.tempFileRetentionHours,
        specifiedType: const FullType(int),
      );
    }
    if (object.timeZone != null) {
      yield r'timeZone';
      yield serializers.serialize(
        object.timeZone,
        specifiedType: const FullType(String),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    UpsertSystemSettingsCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required UpsertSystemSettingsCommandBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'enableLocalSignUp':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.enableLocalSignUp = valueDes;
          break;
        case r'currencyDisplay':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.currencyDisplay = valueDes;
          break;
        case r'currencyThousandthsSeparator':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(CurrencySeparator),
          ) as CurrencySeparator;
          result.currencyThousandthsSeparator = valueDes;
          break;
        case r'currencyDecimalSeparator':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(CurrencySeparator),
          ) as CurrencySeparator;
          result.currencyDecimalSeparator = valueDes;
          break;
        case r'currencySymbolPosition':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(CurrencySymbolPosition),
          ) as CurrencySymbolPosition;
          result.currencySymbolPosition = valueDes;
          break;
        case r'currencyHideDecimalPlaces':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.currencyHideDecimalPlaces = valueDes;
          break;
        case r'debugOcr':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.debugOcr = valueDes;
          break;
        case r'numWorkers':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.numWorkers = valueDes;
          break;
        case r'emailPollingInterval':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.emailPollingInterval = valueDes;
          break;
        case r'receiptProcessingSettingsId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.receiptProcessingSettingsId = valueDes;
          break;
        case r'fallbackReceiptProcessingSettingsId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.fallbackReceiptProcessingSettingsId = valueDes;
          break;
        case r'taskConcurrency':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.taskConcurrency = valueDes;
          break;
        case r'pdfDpi':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.pdfDpi = valueDes;
          break;
        case r'taskQueueConfigurations':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(UpsertTaskQueueConfiguration)]),
          ) as BuiltList<UpsertTaskQueueConfiguration>;
          result.taskQueueConfigurations.replace(valueDes);
          break;
        case r'mcpEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.mcpEnabled = valueDes;
          break;
        case r'mcpPublicUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.mcpPublicUrl = valueDes;
          break;
        case r'serverPublicUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.serverPublicUrl = valueDes;
          break;
        case r'showLoginQr':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.showLoginQr = valueDes;
          break;
        case r'mobileServerUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.mobileServerUrl = valueDes;
          break;
        case r'refreshTokenValidForHours':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.refreshTokenValidForHours = valueDes;
          break;
        case r'mcpRefreshTokenValidForHours':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.mcpRefreshTokenValidForHours = valueDes;
          break;
        case r'tempFileRetentionHours':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.tempFileRetentionHours = valueDes;
          break;
        case r'timeZone':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.timeZone = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  UpsertSystemSettingsCommand deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = UpsertSystemSettingsCommandBuilder();
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

