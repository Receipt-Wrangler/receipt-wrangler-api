// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'upsert_oidc_provider_command.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$UpsertOidcProviderCommand extends UpsertOidcProviderCommand {
  @override
  final String name;
  @override
  final String displayName;
  @override
  final String issuerUrl;
  @override
  final String clientId;
  @override
  final String? clientSecret;
  @override
  final String scope;
  @override
  final bool? allowProvisioning;
  @override
  final bool? linkByUsername;
  @override
  final bool? enabled;

  factory _$UpsertOidcProviderCommand(
          [void Function(UpsertOidcProviderCommandBuilder)? updates]) =>
      (UpsertOidcProviderCommandBuilder()..update(updates))._build();

  _$UpsertOidcProviderCommand._(
      {required this.name,
      required this.displayName,
      required this.issuerUrl,
      required this.clientId,
      this.clientSecret,
      required this.scope,
      this.allowProvisioning,
      this.linkByUsername,
      this.enabled})
      : super._();
  @override
  UpsertOidcProviderCommand rebuild(
          void Function(UpsertOidcProviderCommandBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  UpsertOidcProviderCommandBuilder toBuilder() =>
      UpsertOidcProviderCommandBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is UpsertOidcProviderCommand &&
        name == other.name &&
        displayName == other.displayName &&
        issuerUrl == other.issuerUrl &&
        clientId == other.clientId &&
        clientSecret == other.clientSecret &&
        scope == other.scope &&
        allowProvisioning == other.allowProvisioning &&
        linkByUsername == other.linkByUsername &&
        enabled == other.enabled;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, name.hashCode);
    _$hash = $jc(_$hash, displayName.hashCode);
    _$hash = $jc(_$hash, issuerUrl.hashCode);
    _$hash = $jc(_$hash, clientId.hashCode);
    _$hash = $jc(_$hash, clientSecret.hashCode);
    _$hash = $jc(_$hash, scope.hashCode);
    _$hash = $jc(_$hash, allowProvisioning.hashCode);
    _$hash = $jc(_$hash, linkByUsername.hashCode);
    _$hash = $jc(_$hash, enabled.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'UpsertOidcProviderCommand')
          ..add('name', name)
          ..add('displayName', displayName)
          ..add('issuerUrl', issuerUrl)
          ..add('clientId', clientId)
          ..add('clientSecret', clientSecret)
          ..add('scope', scope)
          ..add('allowProvisioning', allowProvisioning)
          ..add('linkByUsername', linkByUsername)
          ..add('enabled', enabled))
        .toString();
  }
}

class UpsertOidcProviderCommandBuilder
    implements
        Builder<UpsertOidcProviderCommand, UpsertOidcProviderCommandBuilder> {
  _$UpsertOidcProviderCommand? _$v;

  String? _name;
  String? get name => _$this._name;
  set name(String? name) => _$this._name = name;

  String? _displayName;
  String? get displayName => _$this._displayName;
  set displayName(String? displayName) => _$this._displayName = displayName;

  String? _issuerUrl;
  String? get issuerUrl => _$this._issuerUrl;
  set issuerUrl(String? issuerUrl) => _$this._issuerUrl = issuerUrl;

  String? _clientId;
  String? get clientId => _$this._clientId;
  set clientId(String? clientId) => _$this._clientId = clientId;

  String? _clientSecret;
  String? get clientSecret => _$this._clientSecret;
  set clientSecret(String? clientSecret) => _$this._clientSecret = clientSecret;

  String? _scope;
  String? get scope => _$this._scope;
  set scope(String? scope) => _$this._scope = scope;

  bool? _allowProvisioning;
  bool? get allowProvisioning => _$this._allowProvisioning;
  set allowProvisioning(bool? allowProvisioning) =>
      _$this._allowProvisioning = allowProvisioning;

  bool? _linkByUsername;
  bool? get linkByUsername => _$this._linkByUsername;
  set linkByUsername(bool? linkByUsername) =>
      _$this._linkByUsername = linkByUsername;

  bool? _enabled;
  bool? get enabled => _$this._enabled;
  set enabled(bool? enabled) => _$this._enabled = enabled;

  UpsertOidcProviderCommandBuilder() {
    UpsertOidcProviderCommand._defaults(this);
  }

  UpsertOidcProviderCommandBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _name = $v.name;
      _displayName = $v.displayName;
      _issuerUrl = $v.issuerUrl;
      _clientId = $v.clientId;
      _clientSecret = $v.clientSecret;
      _scope = $v.scope;
      _allowProvisioning = $v.allowProvisioning;
      _linkByUsername = $v.linkByUsername;
      _enabled = $v.enabled;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(UpsertOidcProviderCommand other) {
    _$v = other as _$UpsertOidcProviderCommand;
  }

  @override
  void update(void Function(UpsertOidcProviderCommandBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  UpsertOidcProviderCommand build() => _build();

  _$UpsertOidcProviderCommand _build() {
    final _$result = _$v ??
        _$UpsertOidcProviderCommand._(
          name: BuiltValueNullFieldError.checkNotNull(
              name, r'UpsertOidcProviderCommand', 'name'),
          displayName: BuiltValueNullFieldError.checkNotNull(
              displayName, r'UpsertOidcProviderCommand', 'displayName'),
          issuerUrl: BuiltValueNullFieldError.checkNotNull(
              issuerUrl, r'UpsertOidcProviderCommand', 'issuerUrl'),
          clientId: BuiltValueNullFieldError.checkNotNull(
              clientId, r'UpsertOidcProviderCommand', 'clientId'),
          clientSecret: clientSecret,
          scope: BuiltValueNullFieldError.checkNotNull(
              scope, r'UpsertOidcProviderCommand', 'scope'),
          allowProvisioning: allowProvisioning,
          linkByUsername: linkByUsername,
          enabled: enabled,
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
