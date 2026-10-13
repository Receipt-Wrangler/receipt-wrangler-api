//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/base_model.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/currency_symbol_position.dart';
import 'package:openapi/src/model/task_queue_configuration.dart';
import 'package:openapi/src/model/currency_separator.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'system_settings.g.dart';

/// SystemSettings
///
/// Properties:
/// * [id] 
/// * [createdAt] 
/// * [createdBy] 
/// * [createdByString] - Created by entity's name
/// * [updatedAt] 
/// * [enableLocalSignUp] - Whether local sign up is enabled
/// * [currencyDisplay] - Currency display
/// * [currencyThousandthsSeparator] 
/// * [currencyDecimalSeparator] 
/// * [currencySymbolPosition] 
/// * [currencyHideDecimalPlaces] - Whether to hide decimal places
/// * [debugOcr] - Debug OCR
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
/// * [refreshTokenValidForHours] - How long a refresh token stays valid, in hours. Refresh tokens rotate on every use, so this is how long a user can be away and still return signed in, not an absolute session cap. 1-720 (30 days); 0 means unset and falls back to the default.
/// * [mcpRefreshTokenValidForHours] - The same for MCP/OAuth connector refresh tokens, kept separate so a long window chosen for human convenience does not extend third-party client tokens. 1-720 (30 days); 0 means unset and falls back to the default.
/// * [tempFileRetentionHours] - How long a file in temp/ is kept once nothing can still act on it, i.e. how long a user has to rerun, preview or download the source image of a failed upload. 24-8760 (1 year); 0 means unset and falls back to the default.
/// * [timeZone] - The app time zone, an IANA name such as America/New_York. Every calendar boundary (which day an Added At or Resolved Date falls on, \"today\", \"this month\", report periods) is worked out in it. A receipt's Date is a calendar day and never shifts with it.
/// * [secureCookies] - Marks every cookie the server sets as Secure, so browsers only send them over HTTPS. Recommended for any install served over HTTPS. Off by default, because a browser drops a Secure cookie set over plain HTTP, which breaks login.
@BuiltValue()
abstract class SystemSettings implements BaseModel, Built<SystemSettings, SystemSettingsBuilder> {
  /// Whether the OAuth 2.1-protected MCP server is enabled
  @BuiltValueField(wireName: r'mcpEnabled')
  bool? get mcpEnabled;

  @BuiltValueField(wireName: r'currencyThousandthsSeparator')
  CurrencySeparator? get currencyThousandthsSeparator;
  // enum currencyThousandthsSeparatorEnum {  ,,  .,  };

  /// DPI used when rasterizing PDFs for OCR/vision processing
  @BuiltValueField(wireName: r'pdfDpi')
  int? get pdfDpi;

  /// How long a file in temp/ is kept once nothing can still act on it, i.e. how long a user has to rerun, preview or download the source image of a failed upload. 24-8760 (1 year); 0 means unset and falls back to the default.
  @BuiltValueField(wireName: r'tempFileRetentionHours')
  int? get tempFileRetentionHours;

  /// Marks every cookie the server sets as Secure, so browsers only send them over HTTPS. Recommended for any install served over HTTPS. Off by default, because a browser drops a Secure cookie set over plain HTTP, which breaks login.
  @BuiltValueField(wireName: r'secureCookies')
  bool? get secureCookies;

  /// Currency display
  @BuiltValueField(wireName: r'currencyDisplay')
  String? get currencyDisplay;

  /// Whether to show the mobile-setup QR code on the desktop login page
  @BuiltValueField(wireName: r'showLoginQr')
  bool? get showLoginQr;

  /// Externally reachable origin of this API, used to build each OIDC provider's redirect URI. Separate from mcpPublicUrl, which is bound into the MCP token audience.
  @BuiltValueField(wireName: r'serverPublicUrl')
  String? get serverPublicUrl;

  /// The app time zone, an IANA name such as America/New_York. Every calendar boundary (which day an Added At or Resolved Date falls on, \"today\", \"this month\", report periods) is worked out in it. A receipt's Date is a calendar day and never shifts with it.
  @BuiltValueField(wireName: r'timeZone')
  String? get timeZone;

  /// Whether to hide decimal places
  @BuiltValueField(wireName: r'currencyHideDecimalPlaces')
  bool? get currencyHideDecimalPlaces;

  /// Externally reachable origin used for MCP OAuth/metadata/redirect URLs and token audience
  @BuiltValueField(wireName: r'mcpPublicUrl')
  String? get mcpPublicUrl;

  @BuiltValueField(wireName: r'currencyDecimalSeparator')
  CurrencySeparator? get currencyDecimalSeparator;
  // enum currencyDecimalSeparatorEnum {  ,,  .,  };

  /// Debug OCR
  @BuiltValueField(wireName: r'debugOcr')
  bool? get debugOcr;

  /// Fallback receipt processing settings foreign key
  @BuiltValueField(wireName: r'fallbackReceiptProcessingSettingsId')
  int? get fallbackReceiptProcessingSettingsId;

  /// Receipt processing settings foreign key
  @BuiltValueField(wireName: r'receiptProcessingSettingsId')
  int? get receiptProcessingSettingsId;

  /// Server/API URL mobile clients connect to; encoded into the login QR's deep link
  @BuiltValueField(wireName: r'mobileServerUrl')
  String? get mobileServerUrl;

  /// How long a refresh token stays valid, in hours. Refresh tokens rotate on every use, so this is how long a user can be away and still return signed in, not an absolute session cap. 1-720 (30 days); 0 means unset and falls back to the default.
  @BuiltValueField(wireName: r'refreshTokenValidForHours')
  int? get refreshTokenValidForHours;

  @BuiltValueField(wireName: r'currencySymbolPosition')
  CurrencySymbolPosition? get currencySymbolPosition;
  // enum currencySymbolPositionEnum {  START,  END,  };

  /// Concurrency for task worker
  @BuiltValueField(wireName: r'taskConcurrency')
  int? get taskConcurrency;

  /// Email polling interval
  @BuiltValueField(wireName: r'emailPollingInterval')
  int? get emailPollingInterval;

  /// Number of workers to use
  @BuiltValueField(wireName: r'numWorkers')
  int? get numWorkers;

  /// Whether local sign up is enabled
  @BuiltValueField(wireName: r'enableLocalSignUp')
  bool? get enableLocalSignUp;

  /// The same for MCP/OAuth connector refresh tokens, kept separate so a long window chosen for human convenience does not extend third-party client tokens. 1-720 (30 days); 0 means unset and falls back to the default.
  @BuiltValueField(wireName: r'mcpRefreshTokenValidForHours')
  int? get mcpRefreshTokenValidForHours;

  @BuiltValueField(wireName: r'taskQueueConfigurations')
  BuiltList<TaskQueueConfiguration> get taskQueueConfigurations;

  SystemSettings._();

  factory SystemSettings([void updates(SystemSettingsBuilder b)]) = _$SystemSettings;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(SystemSettingsBuilder b) => b
      ..mcpEnabled = false
      ..pdfDpi = 300
      ..tempFileRetentionHours = 720
      ..currencyDisplay = r'$'
      ..currencyHideDecimalPlaces = false
      ..numWorkers = 1
      ..enableLocalSignUp = false
      ..mcpRefreshTokenValidForHours = 24
      ..createdByString = ''
      ..updatedAt = ''
      ..secureCookies = false
      ..showLoginQr = false
      ..timeZone = 'UTC'
      ..debugOcr = false
      ..refreshTokenValidForHours = 24
      ..createdBy = 0
      ..taskConcurrency = 10
      ..emailPollingInterval = 1800;

  @BuiltValueSerializer(custom: true)
  static Serializer<SystemSettings> get serializer => _$SystemSettingsSerializer();
}

class _$SystemSettingsSerializer implements PrimitiveSerializer<SystemSettings> {
  @override
  final Iterable<Type> types = const [SystemSettings, _$SystemSettings];

  @override
  final String wireName = r'SystemSettings';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    SystemSettings object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    if (object.mcpEnabled != null) {
      yield r'mcpEnabled';
      yield serializers.serialize(
        object.mcpEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.pdfDpi != null) {
      yield r'pdfDpi';
      yield serializers.serialize(
        object.pdfDpi,
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
    if (object.currencyDisplay != null) {
      yield r'currencyDisplay';
      yield serializers.serialize(
        object.currencyDisplay,
        specifiedType: const FullType(String),
      );
    }
    if (object.currencyHideDecimalPlaces != null) {
      yield r'currencyHideDecimalPlaces';
      yield serializers.serialize(
        object.currencyHideDecimalPlaces,
        specifiedType: const FullType(bool),
      );
    }
    if (object.currencyDecimalSeparator != null) {
      yield r'currencyDecimalSeparator';
      yield serializers.serialize(
        object.currencyDecimalSeparator,
        specifiedType: const FullType(CurrencySeparator),
      );
    }
    yield r'createdAt';
    yield serializers.serialize(
      object.createdAt,
      specifiedType: const FullType(String),
    );
    if (object.mobileServerUrl != null) {
      yield r'mobileServerUrl';
      yield serializers.serialize(
        object.mobileServerUrl,
        specifiedType: const FullType(String),
      );
    }
    if (object.numWorkers != null) {
      yield r'numWorkers';
      yield serializers.serialize(
        object.numWorkers,
        specifiedType: const FullType(int),
      );
    }
    if (object.enableLocalSignUp != null) {
      yield r'enableLocalSignUp';
      yield serializers.serialize(
        object.enableLocalSignUp,
        specifiedType: const FullType(bool),
      );
    }
    yield r'id';
    yield serializers.serialize(
      object.id,
      specifiedType: const FullType(int),
    );
    if (object.mcpRefreshTokenValidForHours != null) {
      yield r'mcpRefreshTokenValidForHours';
      yield serializers.serialize(
        object.mcpRefreshTokenValidForHours,
        specifiedType: const FullType(int),
      );
    }
    if (object.createdByString != null) {
      yield r'createdByString';
      yield serializers.serialize(
        object.createdByString,
        specifiedType: const FullType(String),
      );
    }
    if (object.updatedAt != null) {
      yield r'updatedAt';
      yield serializers.serialize(
        object.updatedAt,
        specifiedType: const FullType(String),
      );
    }
    if (object.currencyThousandthsSeparator != null) {
      yield r'currencyThousandthsSeparator';
      yield serializers.serialize(
        object.currencyThousandthsSeparator,
        specifiedType: const FullType(CurrencySeparator),
      );
    }
    if (object.secureCookies != null) {
      yield r'secureCookies';
      yield serializers.serialize(
        object.secureCookies,
        specifiedType: const FullType(bool),
      );
    }
    if (object.showLoginQr != null) {
      yield r'showLoginQr';
      yield serializers.serialize(
        object.showLoginQr,
        specifiedType: const FullType(bool),
      );
    }
    if (object.serverPublicUrl != null) {
      yield r'serverPublicUrl';
      yield serializers.serialize(
        object.serverPublicUrl,
        specifiedType: const FullType(String),
      );
    }
    if (object.timeZone != null) {
      yield r'timeZone';
      yield serializers.serialize(
        object.timeZone,
        specifiedType: const FullType(String),
      );
    }
    if (object.mcpPublicUrl != null) {
      yield r'mcpPublicUrl';
      yield serializers.serialize(
        object.mcpPublicUrl,
        specifiedType: const FullType(String),
      );
    }
    if (object.debugOcr != null) {
      yield r'debugOcr';
      yield serializers.serialize(
        object.debugOcr,
        specifiedType: const FullType(bool),
      );
    }
    if (object.fallbackReceiptProcessingSettingsId != null) {
      yield r'fallbackReceiptProcessingSettingsId';
      yield serializers.serialize(
        object.fallbackReceiptProcessingSettingsId,
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
    if (object.refreshTokenValidForHours != null) {
      yield r'refreshTokenValidForHours';
      yield serializers.serialize(
        object.refreshTokenValidForHours,
        specifiedType: const FullType(int),
      );
    }
    if (object.createdBy != null) {
      yield r'createdBy';
      yield serializers.serialize(
        object.createdBy,
        specifiedType: const FullType(int),
      );
    }
    if (object.currencySymbolPosition != null) {
      yield r'currencySymbolPosition';
      yield serializers.serialize(
        object.currencySymbolPosition,
        specifiedType: const FullType(CurrencySymbolPosition),
      );
    }
    if (object.taskConcurrency != null) {
      yield r'taskConcurrency';
      yield serializers.serialize(
        object.taskConcurrency,
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
    yield r'taskQueueConfigurations';
    yield serializers.serialize(
      object.taskQueueConfigurations,
      specifiedType: const FullType(BuiltList, [FullType(TaskQueueConfiguration)]),
    );
  }

  @override
  Object serialize(
    Serializers serializers,
    SystemSettings object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required SystemSettingsBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'mcpEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.mcpEnabled = valueDes;
          break;
        case r'pdfDpi':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.pdfDpi = valueDes;
          break;
        case r'tempFileRetentionHours':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.tempFileRetentionHours = valueDes;
          break;
        case r'currencyDisplay':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.currencyDisplay = valueDes;
          break;
        case r'currencyHideDecimalPlaces':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.currencyHideDecimalPlaces = valueDes;
          break;
        case r'currencyDecimalSeparator':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(CurrencySeparator),
          ) as CurrencySeparator;
          result.currencyDecimalSeparator = valueDes;
          break;
        case r'createdAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.createdAt = valueDes;
          break;
        case r'mobileServerUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.mobileServerUrl = valueDes;
          break;
        case r'numWorkers':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.numWorkers = valueDes;
          break;
        case r'enableLocalSignUp':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.enableLocalSignUp = valueDes;
          break;
        case r'id':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.id = valueDes;
          break;
        case r'mcpRefreshTokenValidForHours':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.mcpRefreshTokenValidForHours = valueDes;
          break;
        case r'createdByString':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.createdByString = valueDes;
          break;
        case r'updatedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.updatedAt = valueDes;
          break;
        case r'currencyThousandthsSeparator':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(CurrencySeparator),
          ) as CurrencySeparator;
          result.currencyThousandthsSeparator = valueDes;
          break;
        case r'secureCookies':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.secureCookies = valueDes;
          break;
        case r'showLoginQr':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.showLoginQr = valueDes;
          break;
        case r'serverPublicUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.serverPublicUrl = valueDes;
          break;
        case r'timeZone':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.timeZone = valueDes;
          break;
        case r'mcpPublicUrl':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.mcpPublicUrl = valueDes;
          break;
        case r'debugOcr':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.debugOcr = valueDes;
          break;
        case r'fallbackReceiptProcessingSettingsId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.fallbackReceiptProcessingSettingsId = valueDes;
          break;
        case r'receiptProcessingSettingsId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.receiptProcessingSettingsId = valueDes;
          break;
        case r'refreshTokenValidForHours':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.refreshTokenValidForHours = valueDes;
          break;
        case r'createdBy':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.createdBy = valueDes;
          break;
        case r'currencySymbolPosition':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(CurrencySymbolPosition),
          ) as CurrencySymbolPosition;
          result.currencySymbolPosition = valueDes;
          break;
        case r'taskConcurrency':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.taskConcurrency = valueDes;
          break;
        case r'emailPollingInterval':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.emailPollingInterval = valueDes;
          break;
        case r'taskQueueConfigurations':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(TaskQueueConfiguration)]),
          ) as BuiltList<TaskQueueConfiguration>;
          result.taskQueueConfigurations.replace(valueDes);
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  SystemSettings deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = SystemSettingsBuilder();
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

