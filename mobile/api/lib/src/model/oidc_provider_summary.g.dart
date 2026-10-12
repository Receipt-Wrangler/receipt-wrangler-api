// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'oidc_provider_summary.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$OidcProviderSummary extends OidcProviderSummary {
  @override
  final String name;
  @override
  final String displayName;

  factory _$OidcProviderSummary(
          [void Function(OidcProviderSummaryBuilder)? updates]) =>
      (OidcProviderSummaryBuilder()..update(updates))._build();

  _$OidcProviderSummary._({required this.name, required this.displayName})
      : super._();
  @override
  OidcProviderSummary rebuild(
          void Function(OidcProviderSummaryBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  OidcProviderSummaryBuilder toBuilder() =>
      OidcProviderSummaryBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is OidcProviderSummary &&
        name == other.name &&
        displayName == other.displayName;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, name.hashCode);
    _$hash = $jc(_$hash, displayName.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'OidcProviderSummary')
          ..add('name', name)
          ..add('displayName', displayName))
        .toString();
  }
}

class OidcProviderSummaryBuilder
    implements Builder<OidcProviderSummary, OidcProviderSummaryBuilder> {
  _$OidcProviderSummary? _$v;

  String? _name;
  String? get name => _$this._name;
  set name(String? name) => _$this._name = name;

  String? _displayName;
  String? get displayName => _$this._displayName;
  set displayName(String? displayName) => _$this._displayName = displayName;

  OidcProviderSummaryBuilder() {
    OidcProviderSummary._defaults(this);
  }

  OidcProviderSummaryBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _name = $v.name;
      _displayName = $v.displayName;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(OidcProviderSummary other) {
    _$v = other as _$OidcProviderSummary;
  }

  @override
  void update(void Function(OidcProviderSummaryBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  OidcProviderSummary build() => _build();

  _$OidcProviderSummary _build() {
    final _$result = _$v ??
        _$OidcProviderSummary._(
          name: BuiltValueNullFieldError.checkNotNull(
              name, r'OidcProviderSummary', 'name'),
          displayName: BuiltValueNullFieldError.checkNotNull(
              displayName, r'OidcProviderSummary', 'displayName'),
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
