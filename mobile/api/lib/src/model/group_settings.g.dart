// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'group_settings.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$GroupSettings extends GroupSettings {
  @override
  final int id;
  @override
  final int groupId;
  @override
  final bool? emailIntegrationEnabled;
  @override
  final bool? emailBodyProcessingEnabled;
  @override
  final bool? emailRequireSignedSender;
  @override
  final int? systemEmailId;
  @override
  final SystemEmail? systemEmail;
  @override
  final String? emailToRead;
  @override
  final BuiltList<SubjectLineRegex>? subjectLineRegexes;
  @override
  final BuiltList<GroupSettingsWhiteListEmail>? emailWhiteList;
  @override
  final ReceiptStatus? emailDefaultReceiptStatus;
  @override
  final int? emailDefaultReceiptPaidById;
  @override
  final Prompt? prompt;
  @override
  final int? promptId;
  @override
  final Prompt? fallbackPrompt;
  @override
  final int? fallbackPromptId;
  @override
  final String? createdAt;
  @override
  final int? createdBy;
  @override
  final String? updatedAt;

  factory _$GroupSettings([void Function(GroupSettingsBuilder)? updates]) =>
      (GroupSettingsBuilder()..update(updates))._build();

  _$GroupSettings._(
      {required this.id,
      required this.groupId,
      this.emailIntegrationEnabled,
      this.emailBodyProcessingEnabled,
      this.emailRequireSignedSender,
      this.systemEmailId,
      this.systemEmail,
      this.emailToRead,
      this.subjectLineRegexes,
      this.emailWhiteList,
      this.emailDefaultReceiptStatus,
      this.emailDefaultReceiptPaidById,
      this.prompt,
      this.promptId,
      this.fallbackPrompt,
      this.fallbackPromptId,
      this.createdAt,
      this.createdBy,
      this.updatedAt})
      : super._();
  @override
  GroupSettings rebuild(void Function(GroupSettingsBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  GroupSettingsBuilder toBuilder() => GroupSettingsBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is GroupSettings &&
        id == other.id &&
        groupId == other.groupId &&
        emailIntegrationEnabled == other.emailIntegrationEnabled &&
        emailBodyProcessingEnabled == other.emailBodyProcessingEnabled &&
        emailRequireSignedSender == other.emailRequireSignedSender &&
        systemEmailId == other.systemEmailId &&
        systemEmail == other.systemEmail &&
        emailToRead == other.emailToRead &&
        subjectLineRegexes == other.subjectLineRegexes &&
        emailWhiteList == other.emailWhiteList &&
        emailDefaultReceiptStatus == other.emailDefaultReceiptStatus &&
        emailDefaultReceiptPaidById == other.emailDefaultReceiptPaidById &&
        prompt == other.prompt &&
        promptId == other.promptId &&
        fallbackPrompt == other.fallbackPrompt &&
        fallbackPromptId == other.fallbackPromptId &&
        createdAt == other.createdAt &&
        createdBy == other.createdBy &&
        updatedAt == other.updatedAt;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, id.hashCode);
    _$hash = $jc(_$hash, groupId.hashCode);
    _$hash = $jc(_$hash, emailIntegrationEnabled.hashCode);
    _$hash = $jc(_$hash, emailBodyProcessingEnabled.hashCode);
    _$hash = $jc(_$hash, emailRequireSignedSender.hashCode);
    _$hash = $jc(_$hash, systemEmailId.hashCode);
    _$hash = $jc(_$hash, systemEmail.hashCode);
    _$hash = $jc(_$hash, emailToRead.hashCode);
    _$hash = $jc(_$hash, subjectLineRegexes.hashCode);
    _$hash = $jc(_$hash, emailWhiteList.hashCode);
    _$hash = $jc(_$hash, emailDefaultReceiptStatus.hashCode);
    _$hash = $jc(_$hash, emailDefaultReceiptPaidById.hashCode);
    _$hash = $jc(_$hash, prompt.hashCode);
    _$hash = $jc(_$hash, promptId.hashCode);
    _$hash = $jc(_$hash, fallbackPrompt.hashCode);
    _$hash = $jc(_$hash, fallbackPromptId.hashCode);
    _$hash = $jc(_$hash, createdAt.hashCode);
    _$hash = $jc(_$hash, createdBy.hashCode);
    _$hash = $jc(_$hash, updatedAt.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'GroupSettings')
          ..add('id', id)
          ..add('groupId', groupId)
          ..add('emailIntegrationEnabled', emailIntegrationEnabled)
          ..add('emailBodyProcessingEnabled', emailBodyProcessingEnabled)
          ..add('emailRequireSignedSender', emailRequireSignedSender)
          ..add('systemEmailId', systemEmailId)
          ..add('systemEmail', systemEmail)
          ..add('emailToRead', emailToRead)
          ..add('subjectLineRegexes', subjectLineRegexes)
          ..add('emailWhiteList', emailWhiteList)
          ..add('emailDefaultReceiptStatus', emailDefaultReceiptStatus)
          ..add('emailDefaultReceiptPaidById', emailDefaultReceiptPaidById)
          ..add('prompt', prompt)
          ..add('promptId', promptId)
          ..add('fallbackPrompt', fallbackPrompt)
          ..add('fallbackPromptId', fallbackPromptId)
          ..add('createdAt', createdAt)
          ..add('createdBy', createdBy)
          ..add('updatedAt', updatedAt))
        .toString();
  }
}

class GroupSettingsBuilder
    implements Builder<GroupSettings, GroupSettingsBuilder> {
  _$GroupSettings? _$v;

  int? _id;
  int? get id => _$this._id;
  set id(int? id) => _$this._id = id;

  int? _groupId;
  int? get groupId => _$this._groupId;
  set groupId(int? groupId) => _$this._groupId = groupId;

  bool? _emailIntegrationEnabled;
  bool? get emailIntegrationEnabled => _$this._emailIntegrationEnabled;
  set emailIntegrationEnabled(bool? emailIntegrationEnabled) =>
      _$this._emailIntegrationEnabled = emailIntegrationEnabled;

  bool? _emailBodyProcessingEnabled;
  bool? get emailBodyProcessingEnabled => _$this._emailBodyProcessingEnabled;
  set emailBodyProcessingEnabled(bool? emailBodyProcessingEnabled) =>
      _$this._emailBodyProcessingEnabled = emailBodyProcessingEnabled;

  bool? _emailRequireSignedSender;
  bool? get emailRequireSignedSender => _$this._emailRequireSignedSender;
  set emailRequireSignedSender(bool? emailRequireSignedSender) =>
      _$this._emailRequireSignedSender = emailRequireSignedSender;

  int? _systemEmailId;
  int? get systemEmailId => _$this._systemEmailId;
  set systemEmailId(int? systemEmailId) =>
      _$this._systemEmailId = systemEmailId;

  SystemEmailBuilder? _systemEmail;
  SystemEmailBuilder get systemEmail =>
      _$this._systemEmail ??= SystemEmailBuilder();
  set systemEmail(SystemEmailBuilder? systemEmail) =>
      _$this._systemEmail = systemEmail;

  String? _emailToRead;
  String? get emailToRead => _$this._emailToRead;
  set emailToRead(String? emailToRead) => _$this._emailToRead = emailToRead;

  ListBuilder<SubjectLineRegex>? _subjectLineRegexes;
  ListBuilder<SubjectLineRegex> get subjectLineRegexes =>
      _$this._subjectLineRegexes ??= ListBuilder<SubjectLineRegex>();
  set subjectLineRegexes(ListBuilder<SubjectLineRegex>? subjectLineRegexes) =>
      _$this._subjectLineRegexes = subjectLineRegexes;

  ListBuilder<GroupSettingsWhiteListEmail>? _emailWhiteList;
  ListBuilder<GroupSettingsWhiteListEmail> get emailWhiteList =>
      _$this._emailWhiteList ??= ListBuilder<GroupSettingsWhiteListEmail>();
  set emailWhiteList(
          ListBuilder<GroupSettingsWhiteListEmail>? emailWhiteList) =>
      _$this._emailWhiteList = emailWhiteList;

  ReceiptStatus? _emailDefaultReceiptStatus;
  ReceiptStatus? get emailDefaultReceiptStatus =>
      _$this._emailDefaultReceiptStatus;
  set emailDefaultReceiptStatus(ReceiptStatus? emailDefaultReceiptStatus) =>
      _$this._emailDefaultReceiptStatus = emailDefaultReceiptStatus;

  int? _emailDefaultReceiptPaidById;
  int? get emailDefaultReceiptPaidById => _$this._emailDefaultReceiptPaidById;
  set emailDefaultReceiptPaidById(int? emailDefaultReceiptPaidById) =>
      _$this._emailDefaultReceiptPaidById = emailDefaultReceiptPaidById;

  PromptBuilder? _prompt;
  PromptBuilder get prompt => _$this._prompt ??= PromptBuilder();
  set prompt(PromptBuilder? prompt) => _$this._prompt = prompt;

  int? _promptId;
  int? get promptId => _$this._promptId;
  set promptId(int? promptId) => _$this._promptId = promptId;

  PromptBuilder? _fallbackPrompt;
  PromptBuilder get fallbackPrompt =>
      _$this._fallbackPrompt ??= PromptBuilder();
  set fallbackPrompt(PromptBuilder? fallbackPrompt) =>
      _$this._fallbackPrompt = fallbackPrompt;

  int? _fallbackPromptId;
  int? get fallbackPromptId => _$this._fallbackPromptId;
  set fallbackPromptId(int? fallbackPromptId) =>
      _$this._fallbackPromptId = fallbackPromptId;

  String? _createdAt;
  String? get createdAt => _$this._createdAt;
  set createdAt(String? createdAt) => _$this._createdAt = createdAt;

  int? _createdBy;
  int? get createdBy => _$this._createdBy;
  set createdBy(int? createdBy) => _$this._createdBy = createdBy;

  String? _updatedAt;
  String? get updatedAt => _$this._updatedAt;
  set updatedAt(String? updatedAt) => _$this._updatedAt = updatedAt;

  GroupSettingsBuilder() {
    GroupSettings._defaults(this);
  }

  GroupSettingsBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _id = $v.id;
      _groupId = $v.groupId;
      _emailIntegrationEnabled = $v.emailIntegrationEnabled;
      _emailBodyProcessingEnabled = $v.emailBodyProcessingEnabled;
      _emailRequireSignedSender = $v.emailRequireSignedSender;
      _systemEmailId = $v.systemEmailId;
      _systemEmail = $v.systemEmail?.toBuilder();
      _emailToRead = $v.emailToRead;
      _subjectLineRegexes = $v.subjectLineRegexes?.toBuilder();
      _emailWhiteList = $v.emailWhiteList?.toBuilder();
      _emailDefaultReceiptStatus = $v.emailDefaultReceiptStatus;
      _emailDefaultReceiptPaidById = $v.emailDefaultReceiptPaidById;
      _prompt = $v.prompt?.toBuilder();
      _promptId = $v.promptId;
      _fallbackPrompt = $v.fallbackPrompt?.toBuilder();
      _fallbackPromptId = $v.fallbackPromptId;
      _createdAt = $v.createdAt;
      _createdBy = $v.createdBy;
      _updatedAt = $v.updatedAt;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(GroupSettings other) {
    _$v = other as _$GroupSettings;
  }

  @override
  void update(void Function(GroupSettingsBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  GroupSettings build() => _build();

  _$GroupSettings _build() {
    _$GroupSettings _$result;
    try {
      _$result = _$v ??
          _$GroupSettings._(
            id: BuiltValueNullFieldError.checkNotNull(
                id, r'GroupSettings', 'id'),
            groupId: BuiltValueNullFieldError.checkNotNull(
                groupId, r'GroupSettings', 'groupId'),
            emailIntegrationEnabled: emailIntegrationEnabled,
            emailBodyProcessingEnabled: emailBodyProcessingEnabled,
            emailRequireSignedSender: emailRequireSignedSender,
            systemEmailId: systemEmailId,
            systemEmail: _systemEmail?.build(),
            emailToRead: emailToRead,
            subjectLineRegexes: _subjectLineRegexes?.build(),
            emailWhiteList: _emailWhiteList?.build(),
            emailDefaultReceiptStatus: emailDefaultReceiptStatus,
            emailDefaultReceiptPaidById: emailDefaultReceiptPaidById,
            prompt: _prompt?.build(),
            promptId: promptId,
            fallbackPrompt: _fallbackPrompt?.build(),
            fallbackPromptId: fallbackPromptId,
            createdAt: createdAt,
            createdBy: createdBy,
            updatedAt: updatedAt,
          );
    } catch (_) {
      late String _$failedField;
      try {
        _$failedField = 'systemEmail';
        _systemEmail?.build();

        _$failedField = 'subjectLineRegexes';
        _subjectLineRegexes?.build();
        _$failedField = 'emailWhiteList';
        _emailWhiteList?.build();

        _$failedField = 'prompt';
        _prompt?.build();

        _$failedField = 'fallbackPrompt';
        _fallbackPrompt?.build();
      } catch (e) {
        throw BuiltValueNestedFieldError(
            r'GroupSettings', _$failedField, e.toString());
      }
      rethrow;
    }
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
