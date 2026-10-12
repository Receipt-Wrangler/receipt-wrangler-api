// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'oidc_flow_error_view.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$OidcFlowErrorView extends OidcFlowErrorView {
  @override
  final String errorCode;

  factory _$OidcFlowErrorView(
          [void Function(OidcFlowErrorViewBuilder)? updates]) =>
      (OidcFlowErrorViewBuilder()..update(updates))._build();

  _$OidcFlowErrorView._({required this.errorCode}) : super._();
  @override
  OidcFlowErrorView rebuild(void Function(OidcFlowErrorViewBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  OidcFlowErrorViewBuilder toBuilder() =>
      OidcFlowErrorViewBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is OidcFlowErrorView && errorCode == other.errorCode;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, errorCode.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'OidcFlowErrorView')
          ..add('errorCode', errorCode))
        .toString();
  }
}

class OidcFlowErrorViewBuilder
    implements Builder<OidcFlowErrorView, OidcFlowErrorViewBuilder> {
  _$OidcFlowErrorView? _$v;

  String? _errorCode;
  String? get errorCode => _$this._errorCode;
  set errorCode(String? errorCode) => _$this._errorCode = errorCode;

  OidcFlowErrorViewBuilder() {
    OidcFlowErrorView._defaults(this);
  }

  OidcFlowErrorViewBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _errorCode = $v.errorCode;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(OidcFlowErrorView other) {
    _$v = other as _$OidcFlowErrorView;
  }

  @override
  void update(void Function(OidcFlowErrorViewBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  OidcFlowErrorView build() => _build();

  _$OidcFlowErrorView _build() {
    final _$result = _$v ??
        _$OidcFlowErrorView._(
          errorCode: BuiltValueNullFieldError.checkNotNull(
              errorCode, r'OidcFlowErrorView', 'errorCode'),
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
