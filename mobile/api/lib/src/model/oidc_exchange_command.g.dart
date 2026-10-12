// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'oidc_exchange_command.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$OidcExchangeCommand extends OidcExchangeCommand {
  @override
  final String code;
  @override
  final String codeVerifier;

  factory _$OidcExchangeCommand(
          [void Function(OidcExchangeCommandBuilder)? updates]) =>
      (OidcExchangeCommandBuilder()..update(updates))._build();

  _$OidcExchangeCommand._({required this.code, required this.codeVerifier})
      : super._();
  @override
  OidcExchangeCommand rebuild(
          void Function(OidcExchangeCommandBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  OidcExchangeCommandBuilder toBuilder() =>
      OidcExchangeCommandBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is OidcExchangeCommand &&
        code == other.code &&
        codeVerifier == other.codeVerifier;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, code.hashCode);
    _$hash = $jc(_$hash, codeVerifier.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'OidcExchangeCommand')
          ..add('code', code)
          ..add('codeVerifier', codeVerifier))
        .toString();
  }
}

class OidcExchangeCommandBuilder
    implements Builder<OidcExchangeCommand, OidcExchangeCommandBuilder> {
  _$OidcExchangeCommand? _$v;

  String? _code;
  String? get code => _$this._code;
  set code(String? code) => _$this._code = code;

  String? _codeVerifier;
  String? get codeVerifier => _$this._codeVerifier;
  set codeVerifier(String? codeVerifier) => _$this._codeVerifier = codeVerifier;

  OidcExchangeCommandBuilder() {
    OidcExchangeCommand._defaults(this);
  }

  OidcExchangeCommandBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _code = $v.code;
      _codeVerifier = $v.codeVerifier;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(OidcExchangeCommand other) {
    _$v = other as _$OidcExchangeCommand;
  }

  @override
  void update(void Function(OidcExchangeCommandBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  OidcExchangeCommand build() => _build();

  _$OidcExchangeCommand _build() {
    final _$result = _$v ??
        _$OidcExchangeCommand._(
          code: BuiltValueNullFieldError.checkNotNull(
              code, r'OidcExchangeCommand', 'code'),
          codeVerifier: BuiltValueNullFieldError.checkNotNull(
              codeVerifier, r'OidcExchangeCommand', 'codeVerifier'),
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
