// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'oidc_connection_view.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$OidcConnectionView extends OidcConnectionView {
  @override
  final String providerName;
  @override
  final String providerDisplayName;
  @override
  final String? preferredUsername;
  @override
  final String? email;
  @override
  final bool provisionedUser;
  @override
  final DateTime linkedAt;
  @override
  final DateTime? lastLoginAt;

  factory _$OidcConnectionView(
          [void Function(OidcConnectionViewBuilder)? updates]) =>
      (OidcConnectionViewBuilder()..update(updates))._build();

  _$OidcConnectionView._(
      {required this.providerName,
      required this.providerDisplayName,
      this.preferredUsername,
      this.email,
      required this.provisionedUser,
      required this.linkedAt,
      this.lastLoginAt})
      : super._();
  @override
  OidcConnectionView rebuild(
          void Function(OidcConnectionViewBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  OidcConnectionViewBuilder toBuilder() =>
      OidcConnectionViewBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is OidcConnectionView &&
        providerName == other.providerName &&
        providerDisplayName == other.providerDisplayName &&
        preferredUsername == other.preferredUsername &&
        email == other.email &&
        provisionedUser == other.provisionedUser &&
        linkedAt == other.linkedAt &&
        lastLoginAt == other.lastLoginAt;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, providerName.hashCode);
    _$hash = $jc(_$hash, providerDisplayName.hashCode);
    _$hash = $jc(_$hash, preferredUsername.hashCode);
    _$hash = $jc(_$hash, email.hashCode);
    _$hash = $jc(_$hash, provisionedUser.hashCode);
    _$hash = $jc(_$hash, linkedAt.hashCode);
    _$hash = $jc(_$hash, lastLoginAt.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'OidcConnectionView')
          ..add('providerName', providerName)
          ..add('providerDisplayName', providerDisplayName)
          ..add('preferredUsername', preferredUsername)
          ..add('email', email)
          ..add('provisionedUser', provisionedUser)
          ..add('linkedAt', linkedAt)
          ..add('lastLoginAt', lastLoginAt))
        .toString();
  }
}

class OidcConnectionViewBuilder
    implements Builder<OidcConnectionView, OidcConnectionViewBuilder> {
  _$OidcConnectionView? _$v;

  String? _providerName;
  String? get providerName => _$this._providerName;
  set providerName(String? providerName) => _$this._providerName = providerName;

  String? _providerDisplayName;
  String? get providerDisplayName => _$this._providerDisplayName;
  set providerDisplayName(String? providerDisplayName) =>
      _$this._providerDisplayName = providerDisplayName;

  String? _preferredUsername;
  String? get preferredUsername => _$this._preferredUsername;
  set preferredUsername(String? preferredUsername) =>
      _$this._preferredUsername = preferredUsername;

  String? _email;
  String? get email => _$this._email;
  set email(String? email) => _$this._email = email;

  bool? _provisionedUser;
  bool? get provisionedUser => _$this._provisionedUser;
  set provisionedUser(bool? provisionedUser) =>
      _$this._provisionedUser = provisionedUser;

  DateTime? _linkedAt;
  DateTime? get linkedAt => _$this._linkedAt;
  set linkedAt(DateTime? linkedAt) => _$this._linkedAt = linkedAt;

  DateTime? _lastLoginAt;
  DateTime? get lastLoginAt => _$this._lastLoginAt;
  set lastLoginAt(DateTime? lastLoginAt) => _$this._lastLoginAt = lastLoginAt;

  OidcConnectionViewBuilder() {
    OidcConnectionView._defaults(this);
  }

  OidcConnectionViewBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _providerName = $v.providerName;
      _providerDisplayName = $v.providerDisplayName;
      _preferredUsername = $v.preferredUsername;
      _email = $v.email;
      _provisionedUser = $v.provisionedUser;
      _linkedAt = $v.linkedAt;
      _lastLoginAt = $v.lastLoginAt;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(OidcConnectionView other) {
    _$v = other as _$OidcConnectionView;
  }

  @override
  void update(void Function(OidcConnectionViewBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  OidcConnectionView build() => _build();

  _$OidcConnectionView _build() {
    final _$result = _$v ??
        _$OidcConnectionView._(
          providerName: BuiltValueNullFieldError.checkNotNull(
              providerName, r'OidcConnectionView', 'providerName'),
          providerDisplayName: BuiltValueNullFieldError.checkNotNull(
              providerDisplayName,
              r'OidcConnectionView',
              'providerDisplayName'),
          preferredUsername: preferredUsername,
          email: email,
          provisionedUser: BuiltValueNullFieldError.checkNotNull(
              provisionedUser, r'OidcConnectionView', 'provisionedUser'),
          linkedAt: BuiltValueNullFieldError.checkNotNull(
              linkedAt, r'OidcConnectionView', 'linkedAt'),
          lastLoginAt: lastLoginAt,
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
