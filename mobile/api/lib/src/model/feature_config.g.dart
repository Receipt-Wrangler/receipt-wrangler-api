// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'feature_config.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$FeatureConfig extends FeatureConfig {
  @override
  final bool aiPoweredReceipts;
  @override
  final bool enableLocalSignUp;
  @override
  final String? loginQrUrl;
  @override
  final BuiltList<OidcProviderSummary>? oidcProviders;

  factory _$FeatureConfig([void Function(FeatureConfigBuilder)? updates]) =>
      (FeatureConfigBuilder()..update(updates))._build();

  _$FeatureConfig._(
      {required this.aiPoweredReceipts,
      required this.enableLocalSignUp,
      this.loginQrUrl,
      this.oidcProviders})
      : super._();
  @override
  FeatureConfig rebuild(void Function(FeatureConfigBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  FeatureConfigBuilder toBuilder() => FeatureConfigBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is FeatureConfig &&
        aiPoweredReceipts == other.aiPoweredReceipts &&
        enableLocalSignUp == other.enableLocalSignUp &&
        loginQrUrl == other.loginQrUrl &&
        oidcProviders == other.oidcProviders;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, aiPoweredReceipts.hashCode);
    _$hash = $jc(_$hash, enableLocalSignUp.hashCode);
    _$hash = $jc(_$hash, loginQrUrl.hashCode);
    _$hash = $jc(_$hash, oidcProviders.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'FeatureConfig')
          ..add('aiPoweredReceipts', aiPoweredReceipts)
          ..add('enableLocalSignUp', enableLocalSignUp)
          ..add('loginQrUrl', loginQrUrl)
          ..add('oidcProviders', oidcProviders))
        .toString();
  }
}

class FeatureConfigBuilder
    implements Builder<FeatureConfig, FeatureConfigBuilder> {
  _$FeatureConfig? _$v;

  bool? _aiPoweredReceipts;
  bool? get aiPoweredReceipts => _$this._aiPoweredReceipts;
  set aiPoweredReceipts(bool? aiPoweredReceipts) =>
      _$this._aiPoweredReceipts = aiPoweredReceipts;

  bool? _enableLocalSignUp;
  bool? get enableLocalSignUp => _$this._enableLocalSignUp;
  set enableLocalSignUp(bool? enableLocalSignUp) =>
      _$this._enableLocalSignUp = enableLocalSignUp;

  String? _loginQrUrl;
  String? get loginQrUrl => _$this._loginQrUrl;
  set loginQrUrl(String? loginQrUrl) => _$this._loginQrUrl = loginQrUrl;

  ListBuilder<OidcProviderSummary>? _oidcProviders;
  ListBuilder<OidcProviderSummary> get oidcProviders =>
      _$this._oidcProviders ??= ListBuilder<OidcProviderSummary>();
  set oidcProviders(ListBuilder<OidcProviderSummary>? oidcProviders) =>
      _$this._oidcProviders = oidcProviders;

  FeatureConfigBuilder() {
    FeatureConfig._defaults(this);
  }

  FeatureConfigBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _aiPoweredReceipts = $v.aiPoweredReceipts;
      _enableLocalSignUp = $v.enableLocalSignUp;
      _loginQrUrl = $v.loginQrUrl;
      _oidcProviders = $v.oidcProviders?.toBuilder();
      _$v = null;
    }
    return this;
  }

  @override
  void replace(FeatureConfig other) {
    _$v = other as _$FeatureConfig;
  }

  @override
  void update(void Function(FeatureConfigBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  FeatureConfig build() => _build();

  _$FeatureConfig _build() {
    _$FeatureConfig _$result;
    try {
      _$result = _$v ??
          _$FeatureConfig._(
            aiPoweredReceipts: BuiltValueNullFieldError.checkNotNull(
                aiPoweredReceipts, r'FeatureConfig', 'aiPoweredReceipts'),
            enableLocalSignUp: BuiltValueNullFieldError.checkNotNull(
                enableLocalSignUp, r'FeatureConfig', 'enableLocalSignUp'),
            loginQrUrl: loginQrUrl,
            oidcProviders: _oidcProviders?.build(),
          );
    } catch (_) {
      late String _$failedField;
      try {
        _$failedField = 'oidcProviders';
        _oidcProviders?.build();
      } catch (e) {
        throw BuiltValueNestedFieldError(
            r'FeatureConfig', _$failedField, e.toString());
      }
      rethrow;
    }
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
