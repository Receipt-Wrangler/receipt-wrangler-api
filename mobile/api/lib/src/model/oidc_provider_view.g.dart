// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'oidc_provider_view.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$OidcProviderView extends OidcProviderView {
  @override
  final int id;
  @override
  final String name;
  @override
  final String displayName;
  @override
  final String issuerUrl;
  @override
  final String clientId;
  @override
  final String scope;
  @override
  final bool allowProvisioning;
  @override
  final bool linkByUsername;
  @override
  final bool enabled;
  @override
  final bool hasClientSecret;
  @override
  final String redirectUri;
  @override
  final DateTime? createdAt;
  @override
  final DateTime? updatedAt;

  factory _$OidcProviderView(
          [void Function(OidcProviderViewBuilder)? updates]) =>
      (OidcProviderViewBuilder()..update(updates))._build();

  _$OidcProviderView._(
      {required this.id,
      required this.name,
      required this.displayName,
      required this.issuerUrl,
      required this.clientId,
      required this.scope,
      required this.allowProvisioning,
      required this.linkByUsername,
      required this.enabled,
      required this.hasClientSecret,
      required this.redirectUri,
      this.createdAt,
      this.updatedAt})
      : super._();
  @override
  OidcProviderView rebuild(void Function(OidcProviderViewBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  OidcProviderViewBuilder toBuilder() =>
      OidcProviderViewBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is OidcProviderView &&
        id == other.id &&
        name == other.name &&
        displayName == other.displayName &&
        issuerUrl == other.issuerUrl &&
        clientId == other.clientId &&
        scope == other.scope &&
        allowProvisioning == other.allowProvisioning &&
        linkByUsername == other.linkByUsername &&
        enabled == other.enabled &&
        hasClientSecret == other.hasClientSecret &&
        redirectUri == other.redirectUri &&
        createdAt == other.createdAt &&
        updatedAt == other.updatedAt;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, id.hashCode);
    _$hash = $jc(_$hash, name.hashCode);
    _$hash = $jc(_$hash, displayName.hashCode);
    _$hash = $jc(_$hash, issuerUrl.hashCode);
    _$hash = $jc(_$hash, clientId.hashCode);
    _$hash = $jc(_$hash, scope.hashCode);
    _$hash = $jc(_$hash, allowProvisioning.hashCode);
    _$hash = $jc(_$hash, linkByUsername.hashCode);
    _$hash = $jc(_$hash, enabled.hashCode);
    _$hash = $jc(_$hash, hasClientSecret.hashCode);
    _$hash = $jc(_$hash, redirectUri.hashCode);
    _$hash = $jc(_$hash, createdAt.hashCode);
    _$hash = $jc(_$hash, updatedAt.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'OidcProviderView')
          ..add('id', id)
          ..add('name', name)
          ..add('displayName', displayName)
          ..add('issuerUrl', issuerUrl)
          ..add('clientId', clientId)
          ..add('scope', scope)
          ..add('allowProvisioning', allowProvisioning)
          ..add('linkByUsername', linkByUsername)
          ..add('enabled', enabled)
          ..add('hasClientSecret', hasClientSecret)
          ..add('redirectUri', redirectUri)
          ..add('createdAt', createdAt)
          ..add('updatedAt', updatedAt))
        .toString();
  }
}

class OidcProviderViewBuilder
    implements Builder<OidcProviderView, OidcProviderViewBuilder> {
  _$OidcProviderView? _$v;

  int? _id;
  int? get id => _$this._id;
  set id(int? id) => _$this._id = id;

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

  bool? _hasClientSecret;
  bool? get hasClientSecret => _$this._hasClientSecret;
  set hasClientSecret(bool? hasClientSecret) =>
      _$this._hasClientSecret = hasClientSecret;

  String? _redirectUri;
  String? get redirectUri => _$this._redirectUri;
  set redirectUri(String? redirectUri) => _$this._redirectUri = redirectUri;

  DateTime? _createdAt;
  DateTime? get createdAt => _$this._createdAt;
  set createdAt(DateTime? createdAt) => _$this._createdAt = createdAt;

  DateTime? _updatedAt;
  DateTime? get updatedAt => _$this._updatedAt;
  set updatedAt(DateTime? updatedAt) => _$this._updatedAt = updatedAt;

  OidcProviderViewBuilder() {
    OidcProviderView._defaults(this);
  }

  OidcProviderViewBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _id = $v.id;
      _name = $v.name;
      _displayName = $v.displayName;
      _issuerUrl = $v.issuerUrl;
      _clientId = $v.clientId;
      _scope = $v.scope;
      _allowProvisioning = $v.allowProvisioning;
      _linkByUsername = $v.linkByUsername;
      _enabled = $v.enabled;
      _hasClientSecret = $v.hasClientSecret;
      _redirectUri = $v.redirectUri;
      _createdAt = $v.createdAt;
      _updatedAt = $v.updatedAt;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(OidcProviderView other) {
    _$v = other as _$OidcProviderView;
  }

  @override
  void update(void Function(OidcProviderViewBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  OidcProviderView build() => _build();

  _$OidcProviderView _build() {
    final _$result = _$v ??
        _$OidcProviderView._(
          id: BuiltValueNullFieldError.checkNotNull(
              id, r'OidcProviderView', 'id'),
          name: BuiltValueNullFieldError.checkNotNull(
              name, r'OidcProviderView', 'name'),
          displayName: BuiltValueNullFieldError.checkNotNull(
              displayName, r'OidcProviderView', 'displayName'),
          issuerUrl: BuiltValueNullFieldError.checkNotNull(
              issuerUrl, r'OidcProviderView', 'issuerUrl'),
          clientId: BuiltValueNullFieldError.checkNotNull(
              clientId, r'OidcProviderView', 'clientId'),
          scope: BuiltValueNullFieldError.checkNotNull(
              scope, r'OidcProviderView', 'scope'),
          allowProvisioning: BuiltValueNullFieldError.checkNotNull(
              allowProvisioning, r'OidcProviderView', 'allowProvisioning'),
          linkByUsername: BuiltValueNullFieldError.checkNotNull(
              linkByUsername, r'OidcProviderView', 'linkByUsername'),
          enabled: BuiltValueNullFieldError.checkNotNull(
              enabled, r'OidcProviderView', 'enabled'),
          hasClientSecret: BuiltValueNullFieldError.checkNotNull(
              hasClientSecret, r'OidcProviderView', 'hasClientSecret'),
          redirectUri: BuiltValueNullFieldError.checkNotNull(
              redirectUri, r'OidcProviderView', 'redirectUri'),
          createdAt: createdAt,
          updatedAt: updatedAt,
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
