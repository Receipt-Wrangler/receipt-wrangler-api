// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'system_task_paged_request_filter.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$SystemTaskPagedRequestFilter extends SystemTaskPagedRequestFilter {
  @override
  final JsonObject? type;
  @override
  final JsonObject? ranBy;
  @override
  final JsonObject? startedAt;
  @override
  final JsonObject? endedAt;

  factory _$SystemTaskPagedRequestFilter(
          [void Function(SystemTaskPagedRequestFilterBuilder)? updates]) =>
      (SystemTaskPagedRequestFilterBuilder()..update(updates))._build();

  _$SystemTaskPagedRequestFilter._(
      {this.type, this.ranBy, this.startedAt, this.endedAt})
      : super._();
  @override
  SystemTaskPagedRequestFilter rebuild(
          void Function(SystemTaskPagedRequestFilterBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  SystemTaskPagedRequestFilterBuilder toBuilder() =>
      SystemTaskPagedRequestFilterBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is SystemTaskPagedRequestFilter &&
        type == other.type &&
        ranBy == other.ranBy &&
        startedAt == other.startedAt &&
        endedAt == other.endedAt;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, type.hashCode);
    _$hash = $jc(_$hash, ranBy.hashCode);
    _$hash = $jc(_$hash, startedAt.hashCode);
    _$hash = $jc(_$hash, endedAt.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'SystemTaskPagedRequestFilter')
          ..add('type', type)
          ..add('ranBy', ranBy)
          ..add('startedAt', startedAt)
          ..add('endedAt', endedAt))
        .toString();
  }
}

class SystemTaskPagedRequestFilterBuilder
    implements
        Builder<SystemTaskPagedRequestFilter,
            SystemTaskPagedRequestFilterBuilder> {
  _$SystemTaskPagedRequestFilter? _$v;

  JsonObject? _type;
  JsonObject? get type => _$this._type;
  set type(JsonObject? type) => _$this._type = type;

  JsonObject? _ranBy;
  JsonObject? get ranBy => _$this._ranBy;
  set ranBy(JsonObject? ranBy) => _$this._ranBy = ranBy;

  JsonObject? _startedAt;
  JsonObject? get startedAt => _$this._startedAt;
  set startedAt(JsonObject? startedAt) => _$this._startedAt = startedAt;

  JsonObject? _endedAt;
  JsonObject? get endedAt => _$this._endedAt;
  set endedAt(JsonObject? endedAt) => _$this._endedAt = endedAt;

  SystemTaskPagedRequestFilterBuilder() {
    SystemTaskPagedRequestFilter._defaults(this);
  }

  SystemTaskPagedRequestFilterBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _type = $v.type;
      _ranBy = $v.ranBy;
      _startedAt = $v.startedAt;
      _endedAt = $v.endedAt;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(SystemTaskPagedRequestFilter other) {
    _$v = other as _$SystemTaskPagedRequestFilter;
  }

  @override
  void update(void Function(SystemTaskPagedRequestFilterBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  SystemTaskPagedRequestFilter build() => _build();

  _$SystemTaskPagedRequestFilter _build() {
    final _$result = _$v ??
        _$SystemTaskPagedRequestFilter._(
          type: type,
          ranBy: ranBy,
          startedAt: startedAt,
          endedAt: endedAt,
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
