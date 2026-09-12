//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/receipt_status.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/subject_line_regex.dart';
import 'package:openapi/src/model/group_settings_white_list_email.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'update_group_settings_command.g.dart';

/// UpdateGroupSettingsCommand
///
/// Properties:
/// * [systemEmailId] - System email foreign key
/// * [emailIntegrationEnabled] - Whether email integration is enabled
/// * [emailBodyProcessingEnabled] - Whether email body text processing is enabled (opt-in, default false)
/// * [subjectLineRegexes] - Subject line regexes
/// * [emailWhiteList] - Email white list
/// * [emailDefaultReceiptStatus] - Default receipt status
/// * [emailDefaultReceiptPaidById] - User foreign key
/// * [promptId] - Prompt foreign key
/// * [fallbackPromptId] - Fallback prompt foreign key
@BuiltValue()
abstract class UpdateGroupSettingsCommand implements Built<UpdateGroupSettingsCommand, UpdateGroupSettingsCommandBuilder> {
  /// System email foreign key
  @BuiltValueField(wireName: r'systemEmailId')
  int get systemEmailId;

  /// Whether email integration is enabled
  @BuiltValueField(wireName: r'emailIntegrationEnabled')
  bool? get emailIntegrationEnabled;

  /// Whether email body text processing is enabled (opt-in, default false)
  @BuiltValueField(wireName: r'emailBodyProcessingEnabled')
  bool? get emailBodyProcessingEnabled;

  /// Subject line regexes
  @BuiltValueField(wireName: r'subjectLineRegexes')
  BuiltList<SubjectLineRegex> get subjectLineRegexes;

  /// Email white list
  @BuiltValueField(wireName: r'emailWhiteList')
  BuiltList<GroupSettingsWhiteListEmail> get emailWhiteList;

  /// Default receipt status
  @BuiltValueField(wireName: r'emailDefaultReceiptStatus')
  ReceiptStatus? get emailDefaultReceiptStatus;
  // enum emailDefaultReceiptStatusEnum {  OPEN,  NEEDS_ATTENTION,  RESOLVED,  DRAFT,  DECLINED,  ,  };

  /// User foreign key
  @BuiltValueField(wireName: r'emailDefaultReceiptPaidById')
  int? get emailDefaultReceiptPaidById;

  /// Prompt foreign key
  @BuiltValueField(wireName: r'promptId')
  int? get promptId;

  /// Fallback prompt foreign key
  @BuiltValueField(wireName: r'fallbackPromptId')
  int? get fallbackPromptId;

  UpdateGroupSettingsCommand._();

  factory UpdateGroupSettingsCommand([void updates(UpdateGroupSettingsCommandBuilder b)]) = _$UpdateGroupSettingsCommand;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(UpdateGroupSettingsCommandBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<UpdateGroupSettingsCommand> get serializer => _$UpdateGroupSettingsCommandSerializer();
}

class _$UpdateGroupSettingsCommandSerializer implements PrimitiveSerializer<UpdateGroupSettingsCommand> {
  @override
  final Iterable<Type> types = const [UpdateGroupSettingsCommand, _$UpdateGroupSettingsCommand];

  @override
  final String wireName = r'UpdateGroupSettingsCommand';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    UpdateGroupSettingsCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'systemEmailId';
    yield serializers.serialize(
      object.systemEmailId,
      specifiedType: const FullType(int),
    );
    if (object.emailIntegrationEnabled != null) {
      yield r'emailIntegrationEnabled';
      yield serializers.serialize(
        object.emailIntegrationEnabled,
        specifiedType: const FullType(bool),
      );
    }
    if (object.emailBodyProcessingEnabled != null) {
      yield r'emailBodyProcessingEnabled';
      yield serializers.serialize(
        object.emailBodyProcessingEnabled,
        specifiedType: const FullType(bool),
      );
    }
    yield r'subjectLineRegexes';
    yield serializers.serialize(
      object.subjectLineRegexes,
      specifiedType: const FullType(BuiltList, [FullType(SubjectLineRegex)]),
    );
    yield r'emailWhiteList';
    yield serializers.serialize(
      object.emailWhiteList,
      specifiedType: const FullType(BuiltList, [FullType(GroupSettingsWhiteListEmail)]),
    );
    if (object.emailDefaultReceiptStatus != null) {
      yield r'emailDefaultReceiptStatus';
      yield serializers.serialize(
        object.emailDefaultReceiptStatus,
        specifiedType: const FullType(ReceiptStatus),
      );
    }
    if (object.emailDefaultReceiptPaidById != null) {
      yield r'emailDefaultReceiptPaidById';
      yield serializers.serialize(
        object.emailDefaultReceiptPaidById,
        specifiedType: const FullType(int),
      );
    }
    if (object.promptId != null) {
      yield r'promptId';
      yield serializers.serialize(
        object.promptId,
        specifiedType: const FullType(int),
      );
    }
    if (object.fallbackPromptId != null) {
      yield r'fallbackPromptId';
      yield serializers.serialize(
        object.fallbackPromptId,
        specifiedType: const FullType(int),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    UpdateGroupSettingsCommand object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required UpdateGroupSettingsCommandBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'systemEmailId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.systemEmailId = valueDes;
          break;
        case r'emailIntegrationEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.emailIntegrationEnabled = valueDes;
          break;
        case r'emailBodyProcessingEnabled':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(bool),
          ) as bool;
          result.emailBodyProcessingEnabled = valueDes;
          break;
        case r'subjectLineRegexes':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(SubjectLineRegex)]),
          ) as BuiltList<SubjectLineRegex>;
          result.subjectLineRegexes.replace(valueDes);
          break;
        case r'emailWhiteList':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(BuiltList, [FullType(GroupSettingsWhiteListEmail)]),
          ) as BuiltList<GroupSettingsWhiteListEmail>;
          result.emailWhiteList.replace(valueDes);
          break;
        case r'emailDefaultReceiptStatus':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(ReceiptStatus),
          ) as ReceiptStatus;
          result.emailDefaultReceiptStatus = valueDes;
          break;
        case r'emailDefaultReceiptPaidById':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.emailDefaultReceiptPaidById = valueDes;
          break;
        case r'promptId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.promptId = valueDes;
          break;
        case r'fallbackPromptId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.fallbackPromptId = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  UpdateGroupSettingsCommand deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = UpdateGroupSettingsCommandBuilder();
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

