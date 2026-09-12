// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_system_task_command.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$GetSystemTaskCommand extends GetSystemTaskCommand {
  @override
  final SystemTaskPagedRequestFilter? filter;
  @override
  final int? associatedEntityId;
  @override
  final AssociatedEntityType? associatedEntityType;
  @override
  final int page;
  @override
  final int pageSize;
  @override
  final String? orderBy;
  @override
  final SortDirection? sortDirection;

  factory _$GetSystemTaskCommand(
          [void Function(GetSystemTaskCommandBuilder)? updates]) =>
      (GetSystemTaskCommandBuilder()..update(updates))._build();

  _$GetSystemTaskCommand._(
      {this.filter,
      this.associatedEntityId,
      this.associatedEntityType,
      required this.page,
      required this.pageSize,
      this.orderBy,
      this.sortDirection})
      : super._();
  @override
  GetSystemTaskCommand rebuild(
          void Function(GetSystemTaskCommandBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  GetSystemTaskCommandBuilder toBuilder() =>
      GetSystemTaskCommandBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is GetSystemTaskCommand &&
        filter == other.filter &&
        associatedEntityId == other.associatedEntityId &&
        associatedEntityType == other.associatedEntityType &&
        page == other.page &&
        pageSize == other.pageSize &&
        orderBy == other.orderBy &&
        sortDirection == other.sortDirection;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, filter.hashCode);
    _$hash = $jc(_$hash, associatedEntityId.hashCode);
    _$hash = $jc(_$hash, associatedEntityType.hashCode);
    _$hash = $jc(_$hash, page.hashCode);
    _$hash = $jc(_$hash, pageSize.hashCode);
    _$hash = $jc(_$hash, orderBy.hashCode);
    _$hash = $jc(_$hash, sortDirection.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'GetSystemTaskCommand')
          ..add('filter', filter)
          ..add('associatedEntityId', associatedEntityId)
          ..add('associatedEntityType', associatedEntityType)
          ..add('page', page)
          ..add('pageSize', pageSize)
          ..add('orderBy', orderBy)
          ..add('sortDirection', sortDirection))
        .toString();
  }
}

class GetSystemTaskCommandBuilder
    implements
        Builder<GetSystemTaskCommand, GetSystemTaskCommandBuilder>,
        PagedRequestCommandBuilder {
  _$GetSystemTaskCommand? _$v;

  SystemTaskPagedRequestFilterBuilder? _filter;
  SystemTaskPagedRequestFilterBuilder get filter =>
      _$this._filter ??= SystemTaskPagedRequestFilterBuilder();
  set filter(covariant SystemTaskPagedRequestFilterBuilder? filter) =>
      _$this._filter = filter;

  int? _associatedEntityId;
  int? get associatedEntityId => _$this._associatedEntityId;
  set associatedEntityId(covariant int? associatedEntityId) =>
      _$this._associatedEntityId = associatedEntityId;

  AssociatedEntityType? _associatedEntityType;
  AssociatedEntityType? get associatedEntityType =>
      _$this._associatedEntityType;
  set associatedEntityType(
          covariant AssociatedEntityType? associatedEntityType) =>
      _$this._associatedEntityType = associatedEntityType;

  int? _page;
  int? get page => _$this._page;
  set page(covariant int? page) => _$this._page = page;

  int? _pageSize;
  int? get pageSize => _$this._pageSize;
  set pageSize(covariant int? pageSize) => _$this._pageSize = pageSize;

  String? _orderBy;
  String? get orderBy => _$this._orderBy;
  set orderBy(covariant String? orderBy) => _$this._orderBy = orderBy;

  SortDirection? _sortDirection;
  SortDirection? get sortDirection => _$this._sortDirection;
  set sortDirection(covariant SortDirection? sortDirection) =>
      _$this._sortDirection = sortDirection;

  GetSystemTaskCommandBuilder() {
    GetSystemTaskCommand._defaults(this);
  }

  GetSystemTaskCommandBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _filter = $v.filter?.toBuilder();
      _associatedEntityId = $v.associatedEntityId;
      _associatedEntityType = $v.associatedEntityType;
      _page = $v.page;
      _pageSize = $v.pageSize;
      _orderBy = $v.orderBy;
      _sortDirection = $v.sortDirection;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(covariant GetSystemTaskCommand other) {
    _$v = other as _$GetSystemTaskCommand;
  }

  @override
  void update(void Function(GetSystemTaskCommandBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  GetSystemTaskCommand build() => _build();

  _$GetSystemTaskCommand _build() {
    _$GetSystemTaskCommand _$result;
    try {
      _$result = _$v ??
          _$GetSystemTaskCommand._(
            filter: _filter?.build(),
            associatedEntityId: associatedEntityId,
            associatedEntityType: associatedEntityType,
            page: BuiltValueNullFieldError.checkNotNull(
                page, r'GetSystemTaskCommand', 'page'),
            pageSize: BuiltValueNullFieldError.checkNotNull(
                pageSize, r'GetSystemTaskCommand', 'pageSize'),
            orderBy: orderBy,
            sortDirection: sortDirection,
          );
    } catch (_) {
      late String _$failedField;
      try {
        _$failedField = 'filter';
        _filter?.build();
      } catch (e) {
        throw BuiltValueNestedFieldError(
            r'GetSystemTaskCommand', _$failedField, e.toString());
      }
      rethrow;
    }
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
