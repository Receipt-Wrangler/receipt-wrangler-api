// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'oidc_link_start_view.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$OidcLinkStartView extends OidcLinkStartView {
  @override
  final String launchUrl;

  factory _$OidcLinkStartView(
          [void Function(OidcLinkStartViewBuilder)? updates]) =>
      (OidcLinkStartViewBuilder()..update(updates))._build();

  _$OidcLinkStartView._({required this.launchUrl}) : super._();
  @override
  OidcLinkStartView rebuild(void Function(OidcLinkStartViewBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  OidcLinkStartViewBuilder toBuilder() =>
      OidcLinkStartViewBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is OidcLinkStartView && launchUrl == other.launchUrl;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, launchUrl.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'OidcLinkStartView')
          ..add('launchUrl', launchUrl))
        .toString();
  }
}

class OidcLinkStartViewBuilder
    implements Builder<OidcLinkStartView, OidcLinkStartViewBuilder> {
  _$OidcLinkStartView? _$v;

  String? _launchUrl;
  String? get launchUrl => _$this._launchUrl;
  set launchUrl(String? launchUrl) => _$this._launchUrl = launchUrl;

  OidcLinkStartViewBuilder() {
    OidcLinkStartView._defaults(this);
  }

  OidcLinkStartViewBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _launchUrl = $v.launchUrl;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(OidcLinkStartView other) {
    _$v = other as _$OidcLinkStartView;
  }

  @override
  void update(void Function(OidcLinkStartViewBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  OidcLinkStartView build() => _build();

  _$OidcLinkStartView _build() {
    final _$result = _$v ??
        _$OidcLinkStartView._(
          launchUrl: BuiltValueNullFieldError.checkNotNull(
              launchUrl, r'OidcLinkStartView', 'launchUrl'),
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
