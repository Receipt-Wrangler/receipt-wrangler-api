// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'upsert_system_settings_command.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$UpsertSystemSettingsCommand extends UpsertSystemSettingsCommand {
  @override
  final bool? enableLocalSignUp;
  @override
  final String? currencyDisplay;
  @override
  final CurrencySeparator currencyThousandthsSeparator;
  @override
  final CurrencySeparator currencyDecimalSeparator;
  @override
  final CurrencySymbolPosition currencySymbolPosition;
  @override
  final bool currencyHideDecimalPlaces;
  @override
  final bool? debugOcr;
  @override
  final int? numWorkers;
  @override
  final int? emailPollingInterval;
  @override
  final int? receiptProcessingSettingsId;
  @override
  final int? fallbackReceiptProcessingSettingsId;
  @override
  final int taskConcurrency;
  @override
  final int? pdfDpi;
  @override
  final BuiltList<UpsertTaskQueueConfiguration>? taskQueueConfigurations;
  @override
  final bool? mcpEnabled;
  @override
  final String? mcpPublicUrl;
  @override
  final String? serverPublicUrl;
  @override
  final bool? showLoginQr;
  @override
  final String? mobileServerUrl;
  @override
  final int? refreshTokenValidForHours;
  @override
  final int? mcpRefreshTokenValidForHours;
  @override
  final int? tempFileRetentionHours;
  @override
  final String? timeZone;
  @override
  final bool? secureCookies;

  factory _$UpsertSystemSettingsCommand(
          [void Function(UpsertSystemSettingsCommandBuilder)? updates]) =>
      (UpsertSystemSettingsCommandBuilder()..update(updates))._build();

  _$UpsertSystemSettingsCommand._(
      {this.enableLocalSignUp,
      this.currencyDisplay,
      required this.currencyThousandthsSeparator,
      required this.currencyDecimalSeparator,
      required this.currencySymbolPosition,
      required this.currencyHideDecimalPlaces,
      this.debugOcr,
      this.numWorkers,
      this.emailPollingInterval,
      this.receiptProcessingSettingsId,
      this.fallbackReceiptProcessingSettingsId,
      required this.taskConcurrency,
      this.pdfDpi,
      this.taskQueueConfigurations,
      this.mcpEnabled,
      this.mcpPublicUrl,
      this.serverPublicUrl,
      this.showLoginQr,
      this.mobileServerUrl,
      this.refreshTokenValidForHours,
      this.mcpRefreshTokenValidForHours,
      this.tempFileRetentionHours,
      this.timeZone,
      this.secureCookies})
      : super._();
  @override
  UpsertSystemSettingsCommand rebuild(
          void Function(UpsertSystemSettingsCommandBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  UpsertSystemSettingsCommandBuilder toBuilder() =>
      UpsertSystemSettingsCommandBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is UpsertSystemSettingsCommand &&
        enableLocalSignUp == other.enableLocalSignUp &&
        currencyDisplay == other.currencyDisplay &&
        currencyThousandthsSeparator == other.currencyThousandthsSeparator &&
        currencyDecimalSeparator == other.currencyDecimalSeparator &&
        currencySymbolPosition == other.currencySymbolPosition &&
        currencyHideDecimalPlaces == other.currencyHideDecimalPlaces &&
        debugOcr == other.debugOcr &&
        numWorkers == other.numWorkers &&
        emailPollingInterval == other.emailPollingInterval &&
        receiptProcessingSettingsId == other.receiptProcessingSettingsId &&
        fallbackReceiptProcessingSettingsId ==
            other.fallbackReceiptProcessingSettingsId &&
        taskConcurrency == other.taskConcurrency &&
        pdfDpi == other.pdfDpi &&
        taskQueueConfigurations == other.taskQueueConfigurations &&
        mcpEnabled == other.mcpEnabled &&
        mcpPublicUrl == other.mcpPublicUrl &&
        serverPublicUrl == other.serverPublicUrl &&
        showLoginQr == other.showLoginQr &&
        mobileServerUrl == other.mobileServerUrl &&
        refreshTokenValidForHours == other.refreshTokenValidForHours &&
        mcpRefreshTokenValidForHours == other.mcpRefreshTokenValidForHours &&
        tempFileRetentionHours == other.tempFileRetentionHours &&
        timeZone == other.timeZone &&
        secureCookies == other.secureCookies;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, enableLocalSignUp.hashCode);
    _$hash = $jc(_$hash, currencyDisplay.hashCode);
    _$hash = $jc(_$hash, currencyThousandthsSeparator.hashCode);
    _$hash = $jc(_$hash, currencyDecimalSeparator.hashCode);
    _$hash = $jc(_$hash, currencySymbolPosition.hashCode);
    _$hash = $jc(_$hash, currencyHideDecimalPlaces.hashCode);
    _$hash = $jc(_$hash, debugOcr.hashCode);
    _$hash = $jc(_$hash, numWorkers.hashCode);
    _$hash = $jc(_$hash, emailPollingInterval.hashCode);
    _$hash = $jc(_$hash, receiptProcessingSettingsId.hashCode);
    _$hash = $jc(_$hash, fallbackReceiptProcessingSettingsId.hashCode);
    _$hash = $jc(_$hash, taskConcurrency.hashCode);
    _$hash = $jc(_$hash, pdfDpi.hashCode);
    _$hash = $jc(_$hash, taskQueueConfigurations.hashCode);
    _$hash = $jc(_$hash, mcpEnabled.hashCode);
    _$hash = $jc(_$hash, mcpPublicUrl.hashCode);
    _$hash = $jc(_$hash, serverPublicUrl.hashCode);
    _$hash = $jc(_$hash, showLoginQr.hashCode);
    _$hash = $jc(_$hash, mobileServerUrl.hashCode);
    _$hash = $jc(_$hash, refreshTokenValidForHours.hashCode);
    _$hash = $jc(_$hash, mcpRefreshTokenValidForHours.hashCode);
    _$hash = $jc(_$hash, tempFileRetentionHours.hashCode);
    _$hash = $jc(_$hash, timeZone.hashCode);
    _$hash = $jc(_$hash, secureCookies.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'UpsertSystemSettingsCommand')
          ..add('enableLocalSignUp', enableLocalSignUp)
          ..add('currencyDisplay', currencyDisplay)
          ..add('currencyThousandthsSeparator', currencyThousandthsSeparator)
          ..add('currencyDecimalSeparator', currencyDecimalSeparator)
          ..add('currencySymbolPosition', currencySymbolPosition)
          ..add('currencyHideDecimalPlaces', currencyHideDecimalPlaces)
          ..add('debugOcr', debugOcr)
          ..add('numWorkers', numWorkers)
          ..add('emailPollingInterval', emailPollingInterval)
          ..add('receiptProcessingSettingsId', receiptProcessingSettingsId)
          ..add('fallbackReceiptProcessingSettingsId',
              fallbackReceiptProcessingSettingsId)
          ..add('taskConcurrency', taskConcurrency)
          ..add('pdfDpi', pdfDpi)
          ..add('taskQueueConfigurations', taskQueueConfigurations)
          ..add('mcpEnabled', mcpEnabled)
          ..add('mcpPublicUrl', mcpPublicUrl)
          ..add('serverPublicUrl', serverPublicUrl)
          ..add('showLoginQr', showLoginQr)
          ..add('mobileServerUrl', mobileServerUrl)
          ..add('refreshTokenValidForHours', refreshTokenValidForHours)
          ..add('mcpRefreshTokenValidForHours', mcpRefreshTokenValidForHours)
          ..add('tempFileRetentionHours', tempFileRetentionHours)
          ..add('timeZone', timeZone)
          ..add('secureCookies', secureCookies))
        .toString();
  }
}

class UpsertSystemSettingsCommandBuilder
    implements
        Builder<UpsertSystemSettingsCommand,
            UpsertSystemSettingsCommandBuilder> {
  _$UpsertSystemSettingsCommand? _$v;

  bool? _enableLocalSignUp;
  bool? get enableLocalSignUp => _$this._enableLocalSignUp;
  set enableLocalSignUp(bool? enableLocalSignUp) =>
      _$this._enableLocalSignUp = enableLocalSignUp;

  String? _currencyDisplay;
  String? get currencyDisplay => _$this._currencyDisplay;
  set currencyDisplay(String? currencyDisplay) =>
      _$this._currencyDisplay = currencyDisplay;

  CurrencySeparator? _currencyThousandthsSeparator;
  CurrencySeparator? get currencyThousandthsSeparator =>
      _$this._currencyThousandthsSeparator;
  set currencyThousandthsSeparator(
          CurrencySeparator? currencyThousandthsSeparator) =>
      _$this._currencyThousandthsSeparator = currencyThousandthsSeparator;

  CurrencySeparator? _currencyDecimalSeparator;
  CurrencySeparator? get currencyDecimalSeparator =>
      _$this._currencyDecimalSeparator;
  set currencyDecimalSeparator(CurrencySeparator? currencyDecimalSeparator) =>
      _$this._currencyDecimalSeparator = currencyDecimalSeparator;

  CurrencySymbolPosition? _currencySymbolPosition;
  CurrencySymbolPosition? get currencySymbolPosition =>
      _$this._currencySymbolPosition;
  set currencySymbolPosition(CurrencySymbolPosition? currencySymbolPosition) =>
      _$this._currencySymbolPosition = currencySymbolPosition;

  bool? _currencyHideDecimalPlaces;
  bool? get currencyHideDecimalPlaces => _$this._currencyHideDecimalPlaces;
  set currencyHideDecimalPlaces(bool? currencyHideDecimalPlaces) =>
      _$this._currencyHideDecimalPlaces = currencyHideDecimalPlaces;

  bool? _debugOcr;
  bool? get debugOcr => _$this._debugOcr;
  set debugOcr(bool? debugOcr) => _$this._debugOcr = debugOcr;

  int? _numWorkers;
  int? get numWorkers => _$this._numWorkers;
  set numWorkers(int? numWorkers) => _$this._numWorkers = numWorkers;

  int? _emailPollingInterval;
  int? get emailPollingInterval => _$this._emailPollingInterval;
  set emailPollingInterval(int? emailPollingInterval) =>
      _$this._emailPollingInterval = emailPollingInterval;

  int? _receiptProcessingSettingsId;
  int? get receiptProcessingSettingsId => _$this._receiptProcessingSettingsId;
  set receiptProcessingSettingsId(int? receiptProcessingSettingsId) =>
      _$this._receiptProcessingSettingsId = receiptProcessingSettingsId;

  int? _fallbackReceiptProcessingSettingsId;
  int? get fallbackReceiptProcessingSettingsId =>
      _$this._fallbackReceiptProcessingSettingsId;
  set fallbackReceiptProcessingSettingsId(
          int? fallbackReceiptProcessingSettingsId) =>
      _$this._fallbackReceiptProcessingSettingsId =
          fallbackReceiptProcessingSettingsId;

  int? _taskConcurrency;
  int? get taskConcurrency => _$this._taskConcurrency;
  set taskConcurrency(int? taskConcurrency) =>
      _$this._taskConcurrency = taskConcurrency;

  int? _pdfDpi;
  int? get pdfDpi => _$this._pdfDpi;
  set pdfDpi(int? pdfDpi) => _$this._pdfDpi = pdfDpi;

  ListBuilder<UpsertTaskQueueConfiguration>? _taskQueueConfigurations;
  ListBuilder<UpsertTaskQueueConfiguration> get taskQueueConfigurations =>
      _$this._taskQueueConfigurations ??=
          ListBuilder<UpsertTaskQueueConfiguration>();
  set taskQueueConfigurations(
          ListBuilder<UpsertTaskQueueConfiguration>? taskQueueConfigurations) =>
      _$this._taskQueueConfigurations = taskQueueConfigurations;

  bool? _mcpEnabled;
  bool? get mcpEnabled => _$this._mcpEnabled;
  set mcpEnabled(bool? mcpEnabled) => _$this._mcpEnabled = mcpEnabled;

  String? _mcpPublicUrl;
  String? get mcpPublicUrl => _$this._mcpPublicUrl;
  set mcpPublicUrl(String? mcpPublicUrl) => _$this._mcpPublicUrl = mcpPublicUrl;

  String? _serverPublicUrl;
  String? get serverPublicUrl => _$this._serverPublicUrl;
  set serverPublicUrl(String? serverPublicUrl) =>
      _$this._serverPublicUrl = serverPublicUrl;

  bool? _showLoginQr;
  bool? get showLoginQr => _$this._showLoginQr;
  set showLoginQr(bool? showLoginQr) => _$this._showLoginQr = showLoginQr;

  String? _mobileServerUrl;
  String? get mobileServerUrl => _$this._mobileServerUrl;
  set mobileServerUrl(String? mobileServerUrl) =>
      _$this._mobileServerUrl = mobileServerUrl;

  int? _refreshTokenValidForHours;
  int? get refreshTokenValidForHours => _$this._refreshTokenValidForHours;
  set refreshTokenValidForHours(int? refreshTokenValidForHours) =>
      _$this._refreshTokenValidForHours = refreshTokenValidForHours;

  int? _mcpRefreshTokenValidForHours;
  int? get mcpRefreshTokenValidForHours => _$this._mcpRefreshTokenValidForHours;
  set mcpRefreshTokenValidForHours(int? mcpRefreshTokenValidForHours) =>
      _$this._mcpRefreshTokenValidForHours = mcpRefreshTokenValidForHours;

  int? _tempFileRetentionHours;
  int? get tempFileRetentionHours => _$this._tempFileRetentionHours;
  set tempFileRetentionHours(int? tempFileRetentionHours) =>
      _$this._tempFileRetentionHours = tempFileRetentionHours;

  String? _timeZone;
  String? get timeZone => _$this._timeZone;
  set timeZone(String? timeZone) => _$this._timeZone = timeZone;

  bool? _secureCookies;
  bool? get secureCookies => _$this._secureCookies;
  set secureCookies(bool? secureCookies) =>
      _$this._secureCookies = secureCookies;

  UpsertSystemSettingsCommandBuilder() {
    UpsertSystemSettingsCommand._defaults(this);
  }

  UpsertSystemSettingsCommandBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _enableLocalSignUp = $v.enableLocalSignUp;
      _currencyDisplay = $v.currencyDisplay;
      _currencyThousandthsSeparator = $v.currencyThousandthsSeparator;
      _currencyDecimalSeparator = $v.currencyDecimalSeparator;
      _currencySymbolPosition = $v.currencySymbolPosition;
      _currencyHideDecimalPlaces = $v.currencyHideDecimalPlaces;
      _debugOcr = $v.debugOcr;
      _numWorkers = $v.numWorkers;
      _emailPollingInterval = $v.emailPollingInterval;
      _receiptProcessingSettingsId = $v.receiptProcessingSettingsId;
      _fallbackReceiptProcessingSettingsId =
          $v.fallbackReceiptProcessingSettingsId;
      _taskConcurrency = $v.taskConcurrency;
      _pdfDpi = $v.pdfDpi;
      _taskQueueConfigurations = $v.taskQueueConfigurations?.toBuilder();
      _mcpEnabled = $v.mcpEnabled;
      _mcpPublicUrl = $v.mcpPublicUrl;
      _serverPublicUrl = $v.serverPublicUrl;
      _showLoginQr = $v.showLoginQr;
      _mobileServerUrl = $v.mobileServerUrl;
      _refreshTokenValidForHours = $v.refreshTokenValidForHours;
      _mcpRefreshTokenValidForHours = $v.mcpRefreshTokenValidForHours;
      _tempFileRetentionHours = $v.tempFileRetentionHours;
      _timeZone = $v.timeZone;
      _secureCookies = $v.secureCookies;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(UpsertSystemSettingsCommand other) {
    _$v = other as _$UpsertSystemSettingsCommand;
  }

  @override
  void update(void Function(UpsertSystemSettingsCommandBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  UpsertSystemSettingsCommand build() => _build();

  _$UpsertSystemSettingsCommand _build() {
    _$UpsertSystemSettingsCommand _$result;
    try {
      _$result = _$v ??
          _$UpsertSystemSettingsCommand._(
            enableLocalSignUp: enableLocalSignUp,
            currencyDisplay: currencyDisplay,
            currencyThousandthsSeparator: BuiltValueNullFieldError.checkNotNull(
                currencyThousandthsSeparator,
                r'UpsertSystemSettingsCommand',
                'currencyThousandthsSeparator'),
            currencyDecimalSeparator: BuiltValueNullFieldError.checkNotNull(
                currencyDecimalSeparator,
                r'UpsertSystemSettingsCommand',
                'currencyDecimalSeparator'),
            currencySymbolPosition: BuiltValueNullFieldError.checkNotNull(
                currencySymbolPosition,
                r'UpsertSystemSettingsCommand',
                'currencySymbolPosition'),
            currencyHideDecimalPlaces: BuiltValueNullFieldError.checkNotNull(
                currencyHideDecimalPlaces,
                r'UpsertSystemSettingsCommand',
                'currencyHideDecimalPlaces'),
            debugOcr: debugOcr,
            numWorkers: numWorkers,
            emailPollingInterval: emailPollingInterval,
            receiptProcessingSettingsId: receiptProcessingSettingsId,
            fallbackReceiptProcessingSettingsId:
                fallbackReceiptProcessingSettingsId,
            taskConcurrency: BuiltValueNullFieldError.checkNotNull(
                taskConcurrency,
                r'UpsertSystemSettingsCommand',
                'taskConcurrency'),
            pdfDpi: pdfDpi,
            taskQueueConfigurations: _taskQueueConfigurations?.build(),
            mcpEnabled: mcpEnabled,
            mcpPublicUrl: mcpPublicUrl,
            serverPublicUrl: serverPublicUrl,
            showLoginQr: showLoginQr,
            mobileServerUrl: mobileServerUrl,
            refreshTokenValidForHours: refreshTokenValidForHours,
            mcpRefreshTokenValidForHours: mcpRefreshTokenValidForHours,
            tempFileRetentionHours: tempFileRetentionHours,
            timeZone: timeZone,
            secureCookies: secureCookies,
          );
    } catch (_) {
      late String _$failedField;
      try {
        _$failedField = 'taskQueueConfigurations';
        _taskQueueConfigurations?.build();
      } catch (e) {
        throw BuiltValueNestedFieldError(
            r'UpsertSystemSettingsCommand', _$failedField, e.toString());
      }
      rethrow;
    }
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
