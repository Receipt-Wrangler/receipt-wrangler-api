//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:openapi/src/model/permission_scope.dart';
import 'package:openapi/src/model/permission.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'permission_descriptor.g.dart';

/// PermissionDescriptor
///
/// Properties:
/// * [key] 
/// * [label] - Short human-readable label for the role editor.
/// * [description] - Full description used as tooltip/help text in the role editor.
/// * [category] - Logical grouping for the role editor UI (e.g., \"User Management\", \"Receipts\").
/// * [scope] 
@BuiltValue()
abstract class PermissionDescriptor implements Built<PermissionDescriptor, PermissionDescriptorBuilder> {
  @BuiltValueField(wireName: r'key')
  Permission get key;
  // enum keyEnum {  app.users.create,  app.users.read,  app.users.update,  app.users.delete,  app.prompts.create,  app.prompts.read,  app.prompts.update,  app.prompts.delete,  app.categories.create,  app.categories.read,  app.categories.update,  app.categories.delete,  app.tags.create,  app.tags.read,  app.tags.update,  app.tags.delete,  app.custom-fields.create,  app.custom-fields.read,  app.custom-fields.update,  app.custom-fields.delete,  app.system-settings.read,  app.system-settings.update,  app.system-settings.restart-task-server,  app.receipt-processing-settings.create,  app.receipt-processing-settings.read,  app.receipt-processing-settings.update,  app.receipt-processing-settings.delete,  app.system-emails.create,  app.system-emails.read,  app.system-emails.update,  app.system-emails.delete,  app.system-tasks.read,  app.imports.run,  app.groups.create,  app.groups.read,  app.groups.update-settings,  app.groups.delete,  app.api-keys.create,  app.api-keys.read,  app.api-keys.read-any,  app.api-keys.update,  app.api-keys.delete,  app.api-keys.delete-any,  app.oidc-providers.create,  app.oidc-providers.read,  app.oidc-providers.update,  app.oidc-providers.delete,  app.roles.create,  app.roles.read,  app.roles.update,  app.roles.delete,  app.notifications.read,  app.notifications.delete,  app.user-preferences.read,  app.user-preferences.update,  app.account.read,  app.account.update,  app.account.delete,  app.receipts.search,  app.reports.create,  app.reports.delete,  app.reports.duplicate,  app.reports.generate,  app.reports.read,  app.reports.update,  app.reports.readAll,  app.reports.createAll,  app.reports.updateAll,  app.reports.deleteAll,  app.reports.duplicateAll,  app.reports.generateAll,  group.view,  group.update,  group.delete,  group.members.create,  group.members.update,  group.members.delete,  group.members.grants.update,  group.receipts.create,  group.receipts.read,  group.receipts.update,  group.receipts.delete,  group.receipts.duplicate,  group.receipts.magic-fill,  group.receipts.quick-scan,  group.comments.create,  group.comments.delete,  group.dashboards.create,  group.dashboards.read,  group.dashboards.update,  group.dashboards.delete,  group.widgets.read,  group.reports.read,  group.activities.read,  group.activities.rerun,  group.email.poll,  };

  /// Short human-readable label for the role editor.
  @BuiltValueField(wireName: r'label')
  String get label;

  /// Full description used as tooltip/help text in the role editor.
  @BuiltValueField(wireName: r'description')
  String get description;

  /// Logical grouping for the role editor UI (e.g., \"User Management\", \"Receipts\").
  @BuiltValueField(wireName: r'category')
  String get category;

  @BuiltValueField(wireName: r'scope')
  PermissionScope get scope;
  // enum scopeEnum {  APP,  GROUP,  };

  PermissionDescriptor._();

  factory PermissionDescriptor([void updates(PermissionDescriptorBuilder b)]) = _$PermissionDescriptor;

  @BuiltValueHook(initializeBuilder: true)
  static void _defaults(PermissionDescriptorBuilder b) => b;

  @BuiltValueSerializer(custom: true)
  static Serializer<PermissionDescriptor> get serializer => _$PermissionDescriptorSerializer();
}

class _$PermissionDescriptorSerializer implements PrimitiveSerializer<PermissionDescriptor> {
  @override
  final Iterable<Type> types = const [PermissionDescriptor, _$PermissionDescriptor];

  @override
  final String wireName = r'PermissionDescriptor';

  Iterable<Object?> _serializeProperties(
    Serializers serializers,
    PermissionDescriptor object, {
    FullType specifiedType = FullType.unspecified,
  }) sync* {
    yield r'key';
    yield serializers.serialize(
      object.key,
      specifiedType: const FullType(Permission),
    );
    yield r'label';
    yield serializers.serialize(
      object.label,
      specifiedType: const FullType(String),
    );
    yield r'description';
    yield serializers.serialize(
      object.description,
      specifiedType: const FullType(String),
    );
    yield r'category';
    yield serializers.serialize(
      object.category,
      specifiedType: const FullType(String),
    );
    yield r'scope';
    yield serializers.serialize(
      object.scope,
      specifiedType: const FullType(PermissionScope),
    );
  }

  @override
  Object serialize(
    Serializers serializers,
    PermissionDescriptor object, {
    FullType specifiedType = FullType.unspecified,
  }) {
    return _serializeProperties(serializers, object, specifiedType: specifiedType).toList();
  }

  void _deserializeProperties(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
    required List<Object?> serializedList,
    required PermissionDescriptorBuilder result,
    required List<Object?> unhandled,
  }) {
    for (var i = 0; i < serializedList.length; i += 2) {
      final key = serializedList[i] as String;
      final value = serializedList[i + 1];
      switch (key) {
        case r'key':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(Permission),
          ) as Permission;
          result.key = valueDes;
          break;
        case r'label':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.label = valueDes;
          break;
        case r'description':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.description = valueDes;
          break;
        case r'category':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(String),
          ) as String;
          result.category = valueDes;
          break;
        case r'scope':
          final valueDes = serializers.deserialize(
            value,
            specifiedType: const FullType(PermissionScope),
          ) as PermissionScope;
          result.scope = valueDes;
          break;
        default:
          unhandled.add(key);
          unhandled.add(value);
          break;
      }
    }
  }

  @override
  PermissionDescriptor deserialize(
    Serializers serializers,
    Object serialized, {
    FullType specifiedType = FullType.unspecified,
  }) {
    final result = PermissionDescriptorBuilder();
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

