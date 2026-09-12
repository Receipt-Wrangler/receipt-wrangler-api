//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/receipt_status.dart';
import 'package:openapi/src/model/system_email.dart';
import 'package:built_collection/built_collection.dart';
import 'package:openapi/src/model/prompt.dart';
import 'package:openapi/src/model/subject_line_regex.dart';
import 'package:openapi/src/model/group_settings_white_list_email.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'group_settings.g.dart';

/// GroupSettings
///
/// Properties:
/// * [id] - Group settings id
/// * [groupId] - Group foreign key
/// * [emailIntegrationEnabled] - Whether email integration is enabled
/// * [emailBodyProcessingEnabled] - Whether email body text processing is enabled (opt-in, default false)
/// * [systemEmailId] - System email foreign key
/// * [systemEmail] 
/// * [emailToRead] - Email to read
/// * [subjectLineRegexes] - Subject line regexes
/// * [emailWhiteList] - Email white list
/// * [emailDefaultReceiptStatus] - Default receipt status
/// * [emailDefaultReceiptPaidById] - User foreign key
/// * [prompt] 
/// * [promptId] - Prompt foreign key
/// * [fallbackPrompt] 
/// * [fallbackPromptId] - Fallback prompt foreign key
/// * [createdAt] 
/// * [createdBy] 
/// * [updatedAt] 
@BuiltValue()
abstract class GroupSettings implements Built<GroupSettings, GroupSettingsBuilder> {
  /// Group settings id
  @BuiltValueField(wireName: r'id')
  int get id;

  /// Group foreign key
  @BuiltValueField(wireName: r'groupId')
  int get groupId;

  /// Whether email integration is enabled
  @BuiltValueField(wireName: r'emailIntegrationEnabled')
  bool? get emailIntegrationEnabled;

  /// Whether email body text processing is enabled (opt-in, default false)
  @BuiltValueField(wireName: r'emailBodyProcessingEnabled')
  bool? get emailBodyProcessingEnabled;

  /// System email foreign key
  @BuiltValueField(wireName: r'systemEmailId')
  int? get systemEmailId;

  @BuiltValueField(wireName: r'systemEmail')
  SystemEmail? get systemEmail;

  /// Email to read
  @BuiltValueField(wireName: r'emailToRead')
  String? get emailToRead;

  /// Subject line regexes
  @BuiltValueField(wireName: r'subjectLineRegexes')
  BuiltList<SubjectLineRegex>? get subjectLineRegexes;

  /// Email white list
  @BuiltValueField(wireName: r'emailWhiteList')
  BuiltList<GroupSettingsWhiteListEmail>? get emailWhiteList;

  /// Default receipt status
  @BuiltValueField(wireName: r'emailDefaultReceiptStatus')
  ReceiptStatus? get emailDefaultReceiptStatus;
  // enum emailDefaultReceiptStatusEnum {  OPEN,  NEEDS_ATTENTION,  RESOLVED,  DRAFT,  DECLINED,  ,  };

  /// User foreign key
  @BuiltValueField(wireName: r'emailDefaultReceiptPaidById')
  int? get emailDefaultReceiptPaidById;

  @BuiltValueField(wireName: r'prompt')
  Prompt? get prompt;

  /// Prompt foreign key
  @BuiltValueField(wireName: r'promptId')
  int? get promptId;

  @BuiltValueField(wireName: r'fallbackPrompt')
  Prompt? get fallbackPrompt;

  /// Fallback prompt foreign key
  @BuiltValueField(wireName: r'fallbackPromptId')
  int? get fallbackPromptId;

  @BuiltValueField(wireName: r'createdAt')
  String? get createdAt;

  @BuiltValueField(wireName: r'createdBy')
  int? get createdBy;

  @BuiltValueField(wireName: r'updatedAt')
  String? get updatedAt;

  GroupSettings._();

  factory GroupSettings([void updates(GroupSettingsBuilder b)]) = _$GroupSettings;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(GroupSettingsBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<GroupSettings> get serializer => _$GroupSettingsSerializer();
}

class _$GroupSettingsSerializer implements PrimitiveSerializer<GroupSettings> {
  @override
  final Iterable<Type> types = const [GroupSettings, _$GroupSettings];

  @override
  final String wireName = r'GroupSettings';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    GroupSettings object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'id';
    yield serializers.serialize(
      object.id,
      specifiedType: const FullType(int),
    );
    yield r'groupId';
    yield serializers.serialize(
      object.groupId,
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
    if (object.systemEmailId != null) {
      yield r'systemEmailId';
      yield serializers.serialize(
        object.systemEmailId,
        specifiedType: const FullType(int),
      );
    }
    if (object.systemEmail != null) {
      yield r'systemEmail';
      yield serializers.serialize(
        object.systemEmail,
        specifiedType: const FullType(SystemEmail),
      );
    }
    if (object.emailToRead != null) {
      yield r'emailToRead';
      yield serializers.serialize(
        object.emailToRead,
        specifiedType: const FullType(String),
      );
    }
    if (object.subjectLineRegexes != null) {
      yield r'subjectLineRegexes';
      yield serializers.serialize(
        object.subjectLineRegexes,
        specifiedType: const FullType(BuiltList, [FullType(SubjectLineRegex)]),
      );
    }
    if (object.emailWhiteList != null) {
      yield r'emailWhiteList';
      yield serializers.serialize(
        object.emailWhiteList,
        specifiedType: const FullType(BuiltList, [FullType(GroupSettingsWhiteListEmail)]),
      );
    }
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
    if (object.prompt != null) {
      yield r'prompt';
      yield serializers.serialize(
        object.prompt,
        specifiedType: const FullType(Prompt),
      );
    }
    if (object.promptId != null) {
      yield r'promptId';
      yield serializers.serialize(
        object.promptId,
        specifiedType: const FullType(int),
      );
    }
    if (object.fallbackPrompt != null) {
      yield r'fallbackPrompt';
      yield serializers.serialize(
        object.fallbackPrompt,
        specifiedType: const FullType(Prompt),
      );
    }
    if (object.fallbackPromptId != null) {
      yield r'fallbackPromptId';
      yield serializers.serialize(
        object.fallbackPromptId,
        specifiedType: const FullType(int),
      );
    }
    if (object.createdAt != null) {
      yield r'createdAt';
      yield serializers.serialize(
        object.createdAt,
        specifiedType: const FullType(String),
      );
    }
    if (object.createdBy != null) {
      yield r'createdBy';
      yield serializers.serialize(
        object.createdBy,
        specifiedType: const FullType(int),
      );
    }
    if (object.updatedAt != null) {
      yield r'updatedAt';
      yield serializers.serialize(
        object.updatedAt,
        specifiedType: const FullType(String),
      );
    }
  }

  @override
  Object serialize(
    Serializers serializers,
    GroupSettings object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required GroupSettingsBuilder result,
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
        case r'groupId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.groupId = valueDes;
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
        case r'systemEmailId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.systemEmailId = valueDes;
          break;
        case r'systemEmail':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(SystemEmail),
          ) as SystemEmail;
          result.systemEmail.replace(valueDes);
          break;
        case r'emailToRead':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.emailToRead = valueDes;
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
        case r'prompt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(Prompt),
          ) as Prompt;
          result.prompt.replace(valueDes);
          break;
        case r'promptId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.promptId = valueDes;
          break;
        case r'fallbackPrompt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(Prompt),
          ) as Prompt;
          result.fallbackPrompt.replace(valueDes);
          break;
        case r'fallbackPromptId':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.fallbackPromptId = valueDes;
          break;
        case r'createdAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.createdAt = valueDes;
          break;
        case r'createdBy':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(int),
          ) as int;
          result.createdBy = valueDes;
          break;
        case r'updatedAt':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
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
  GroupSettings deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = GroupSettingsBuilder();
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

