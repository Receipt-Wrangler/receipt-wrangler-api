// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'update_group_settings_command.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$UpdateGroupSettingsCommand extends UpdateGroupSettingsCommand {
  @override
  final int systemEmailId;
  @override
  final bool? emailIntegrationEnabled;
  @override
  final bool? emailBodyProcessingEnabled;
  @override
  final bool? emailRequireSignedSender;
  @override
  final BuiltList<SubjectLineRegex> subjectLineRegexes;
  @override
  final BuiltList<GroupSettingsWhiteListEmail> emailWhiteList;
  @override
  final ReceiptStatus? emailDefaultReceiptStatus;
  @override
  final int? emailDefaultReceiptPaidById;
  @override
  final int? promptId;
  @override
  final int? fallbackPromptId;

  factory _$UpdateGroupSettingsCommand(
          [void Function(UpdateGroupSettingsCommandBuilder)? updates]) =>
      (UpdateGroupSettingsCommandBuilder()..update(updates))._build();

  _$UpdateGroupSettingsCommand._(
      {required this.systemEmailId,
      this.emailIntegrationEnabled,
      this.emailBodyProcessingEnabled,
      this.emailRequireSignedSender,
      required this.subjectLineRegexes,
      required this.emailWhiteList,
      this.emailDefaultReceiptStatus,
      this.emailDefaultReceiptPaidById,
      this.promptId,
      this.fallbackPromptId})
      : super._();
  @override
  UpdateGroupSettingsCommand rebuild(
          void Function(UpdateGroupSettingsCommandBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  UpdateGroupSettingsCommandBuilder toBuilder() =>
      UpdateGroupSettingsCommandBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is UpdateGroupSettingsCommand &&
        systemEmailId == other.systemEmailId &&
        emailIntegrationEnabled == other.emailIntegrationEnabled &&
        emailBodyProcessingEnabled == other.emailBodyProcessingEnabled &&
        emailRequireSignedSender == other.emailRequireSignedSender &&
        subjectLineRegexes == other.subjectLineRegexes &&
        emailWhiteList == other.emailWhiteList &&
        emailDefaultReceiptStatus == other.emailDefaultReceiptStatus &&
        emailDefaultReceiptPaidById == other.emailDefaultReceiptPaidById &&
        promptId == other.promptId &&
        fallbackPromptId == other.fallbackPromptId;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, systemEmailId.hashCode);
    _$hash = $jc(_$hash, emailIntegrationEnabled.hashCode);
    _$hash = $jc(_$hash, emailBodyProcessingEnabled.hashCode);
    _$hash = $jc(_$hash, emailRequireSignedSender.hashCode);
    _$hash = $jc(_$hash, subjectLineRegexes.hashCode);
    _$hash = $jc(_$hash, emailWhiteList.hashCode);
    _$hash = $jc(_$hash, emailDefaultReceiptStatus.hashCode);
    _$hash = $jc(_$hash, emailDefaultReceiptPaidById.hashCode);
    _$hash = $jc(_$hash, promptId.hashCode);
    _$hash = $jc(_$hash, fallbackPromptId.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'UpdateGroupSettingsCommand')
          ..add('systemEmailId', systemEmailId)
          ..add('emailIntegrationEnabled', emailIntegrationEnabled)
          ..add('emailBodyProcessingEnabled', emailBodyProcessingEnabled)
          ..add('emailRequireSignedSender', emailRequireSignedSender)
          ..add('subjectLineRegexes', subjectLineRegexes)
          ..add('emailWhiteList', emailWhiteList)
          ..add('emailDefaultReceiptStatus', emailDefaultReceiptStatus)
          ..add('emailDefaultReceiptPaidById', emailDefaultReceiptPaidById)
          ..add('promptId', promptId)
          ..add('fallbackPromptId', fallbackPromptId))
        .toString();
  }
}

class UpdateGroupSettingsCommandBuilder
    implements
        Builder<UpdateGroupSettingsCommand, UpdateGroupSettingsCommandBuilder> {
  _$UpdateGroupSettingsCommand? _$v;

  int? _systemEmailId;
  int? get systemEmailId => _$this._systemEmailId;
  set systemEmailId(int? systemEmailId) =>
      _$this._systemEmailId = systemEmailId;

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

  int? _promptId;
  int? get promptId => _$this._promptId;
  set promptId(int? promptId) => _$this._promptId = promptId;

  int? _fallbackPromptId;
  int? get fallbackPromptId => _$this._fallbackPromptId;
  set fallbackPromptId(int? fallbackPromptId) =>
      _$this._fallbackPromptId = fallbackPromptId;

  UpdateGroupSettingsCommandBuilder() {
    UpdateGroupSettingsCommand._defaults(this);
  }

  UpdateGroupSettingsCommandBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _systemEmailId = $v.systemEmailId;
      _emailIntegrationEnabled = $v.emailIntegrationEnabled;
      _emailBodyProcessingEnabled = $v.emailBodyProcessingEnabled;
      _emailRequireSignedSender = $v.emailRequireSignedSender;
      _subjectLineRegexes = $v.subjectLineRegexes.toBuilder();
      _emailWhiteList = $v.emailWhiteList.toBuilder();
      _emailDefaultReceiptStatus = $v.emailDefaultReceiptStatus;
      _emailDefaultReceiptPaidById = $v.emailDefaultReceiptPaidById;
      _promptId = $v.promptId;
      _fallbackPromptId = $v.fallbackPromptId;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(UpdateGroupSettingsCommand other) {
    _$v = other as _$UpdateGroupSettingsCommand;
  }

  @override
  void update(void Function(UpdateGroupSettingsCommandBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  UpdateGroupSettingsCommand build() => _build();

  _$UpdateGroupSettingsCommand _build() {
    _$UpdateGroupSettingsCommand _$result;
    try {
      _$result = _$v ??
          _$UpdateGroupSettingsCommand._(
            systemEmailId: BuiltValueNullFieldError.checkNotNull(
                systemEmailId, r'UpdateGroupSettingsCommand', 'systemEmailId'),
            emailIntegrationEnabled: emailIntegrationEnabled,
            emailBodyProcessingEnabled: emailBodyProcessingEnabled,
            emailRequireSignedSender: emailRequireSignedSender,
            subjectLineRegexes: subjectLineRegexes.build(),
            emailWhiteList: emailWhiteList.build(),
            emailDefaultReceiptStatus: emailDefaultReceiptStatus,
            emailDefaultReceiptPaidById: emailDefaultReceiptPaidById,
            promptId: promptId,
            fallbackPromptId: fallbackPromptId,
          );
    } catch (_) {
      late String _$failedField;
      try {
        _$failedField = 'subjectLineRegexes';
        subjectLineRegexes.build();
        _$failedField = 'emailWhiteList';
        emailWhiteList.build();
      } catch (e) {
        throw BuiltValueNestedFieldError(
            r'UpdateGroupSettingsCommand', _$failedField, e.toString());
      }
      rethrow;
    }
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
